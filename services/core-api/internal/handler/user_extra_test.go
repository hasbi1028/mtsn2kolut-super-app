package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestUserExtraGenerateAccountHandlersLowBranches(t *testing.T) {
	actorID := "01000000-0000-0000-0000-000000000321"

	t.Run("student generate requires valid actor and maps service error", func(t *testing.T) {
		svc := &fakeStudentAccountGenerationService{generateErr: errors.New("student generator down")}
		h := &User{studentGenerator: svc}
		rec := httptest.NewRecorder()

		h.GenerateStudentAccounts(rec, accountGenerationRequest(http.MethodPost, "/api/users/student-accounts/generate", actorID, "student_accounts.manage"))

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("GenerateStudentAccounts(service error) status = %d, want 500; body=%s", rec.Code, rec.Body.String())
		}
		if svc.generateActorID.String() != actorID {
			t.Fatalf("GenerateStudentAccounts actor = %s, want %s", svc.generateActorID.String(), actorID)
		}

		rec = httptest.NewRecorder()
		h.GenerateStudentAccounts(rec, accountGenerationRequest(http.MethodPost, "/api/users/student-accounts/generate", "not-a-uuid", "student_accounts.manage"))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("GenerateStudentAccounts(bad actor) status = %d, want 401; body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("parent generate missing permission and service error", func(t *testing.T) {
		svc := &fakeParentAccountGenerationService{generateErr: errors.New("parent generator down")}
		h := &User{parentGenerator: svc}
		rec := httptest.NewRecorder()

		h.GenerateParentAccounts(rec, accountGenerationRequest(http.MethodPost, "/api/users/parent-accounts/generate", actorID, "student_accounts.manage"))
		if rec.Code != http.StatusForbidden {
			t.Fatalf("GenerateParentAccounts(missing permission) status = %d, want 403; body=%s", rec.Code, rec.Body.String())
		}

		rec = httptest.NewRecorder()
		h.GenerateParentAccounts(rec, accountGenerationRequest(http.MethodPost, "/api/users/parent-accounts/generate", actorID, "parent_accounts.manage"))
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("GenerateParentAccounts(service error) status = %d, want 500; body=%s", rec.Code, rec.Body.String())
		}
		if svc.generateActorID.String() != actorID {
			t.Fatalf("GenerateParentAccounts actor = %s, want %s", svc.generateActorID.String(), actorID)
		}
	})

	t.Run("employee generate requires authenticated admin actor and maps service error", func(t *testing.T) {
		svc := &fakeEmployeeAccountGenerationService{generateErr: errors.New("employee generator down")}
		h := &User{generator: svc}
		rec := httptest.NewRecorder()

		h.GenerateEmployeeAccounts(rec, withClaims(httptest.NewRequest(http.MethodPost, "/api/users/employee-accounts/generate", nil), jwt.MapClaims{"roles": []any{"admin"}}))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("GenerateEmployeeAccounts(admin without actor) status = %d, want 401; body=%s", rec.Code, rec.Body.String())
		}

		rec = httptest.NewRecorder()
		h.GenerateEmployeeAccounts(rec, accountGenerationAdminRequest(http.MethodPost, "/api/users/employee-accounts/generate", actorID))
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("GenerateEmployeeAccounts(service error) status = %d, want 500; body=%s", rec.Code, rec.Body.String())
		}
		if svc.generateActorID.String() != actorID {
			t.Fatalf("GenerateEmployeeAccounts actor = %s, want %s", svc.generateActorID.String(), actorID)
		}
	})
}

func TestUserExtraProfileLinkFromRequest(t *testing.T) {
	employeeID := "11111111-1111-1111-1111-111111111111"
	studentID := "22222222-2222-2222-2222-222222222222"
	parentID := "33333333-3333-3333-3333-333333333333"

	t.Run("empty clears all profile links", func(t *testing.T) {
		link, err := profileLinkFromRequest("", "", "")
		if err != nil {
			t.Fatalf("profileLinkFromRequest(empty) error = %v", err)
		}
		if link.EmployeeID.Valid || link.StudentID.Valid || link.ParentID.Valid {
			t.Fatalf("empty link = %+v, want all invalid", link)
		}
	})

	t.Run("parses each profile id independently", func(t *testing.T) {
		link, err := profileLinkFromRequest(employeeID, studentID, parentID)
		if err != nil {
			t.Fatalf("profileLinkFromRequest(valid) error = %v", err)
		}
		if pgUUIDString(link.EmployeeID) != employeeID || pgUUIDString(link.StudentID) != studentID || pgUUIDString(link.ParentID) != parentID {
			t.Fatalf("profile link = employee %s student %s parent %s", pgUUIDString(link.EmployeeID), pgUUIDString(link.StudentID), pgUUIDString(link.ParentID))
		}
	})

	for _, tt := range []struct {
		name       string
		employeeID string
		studentID  string
		parentID   string
		want       string
	}{
		{name: "bad employee", employeeID: "bad", want: "employee_id tidak valid"},
		{name: "bad student", studentID: "bad", want: "student_id tidak valid"},
		{name: "bad parent", parentID: "bad", want: "parent_id tidak valid"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := profileLinkFromRequest(tt.employeeID, tt.studentID, tt.parentID)
			if err == nil || err.Error() != tt.want {
				t.Fatalf("profileLinkFromRequest() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestUserExtraUpdateProfileLinkForwardsStudentAndParent(t *testing.T) {
	userID := handlerTestUUID(241)
	actorID := "01000000-0000-0000-0000-000000000332"
	studentID := "22222222-2222-2222-2222-222222222222"
	parentID := "33333333-3333-3333-3333-333333333333"
	lifecycle := &fakeUserLifecycle{}
	h := &User{lifecycle: lifecycle}

	rec := httptest.NewRecorder()
	req := withRouteParam(accountGenerationAdminRequest(http.MethodPatch, "/api/users/"+userID.String()+"/profile-link", actorID, `{"student_id":"`+studentID+`"}`), "id", userID.String())
	h.UpdateProfileLink(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateProfileLink(student) status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if lifecycle.profileID != userID || pgUUIDString(lifecycle.profileActorID) != actorID || pgUUIDString(lifecycle.profileLink.StudentID) != studentID || lifecycle.profileLink.EmployeeID.Valid || lifecycle.profileLink.ParentID.Valid {
		t.Fatalf("student link args = user %v actor %s link %+v", lifecycle.profileID, pgUUIDString(lifecycle.profileActorID), lifecycle.profileLink)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(accountGenerationAdminRequest(http.MethodPatch, "/api/users/"+userID.String()+"/profile-link", actorID, `{"parent_id":"`+parentID+`"}`), "id", userID.String())
	h.UpdateProfileLink(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateProfileLink(parent) status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if pgUUIDString(lifecycle.profileLink.ParentID) != parentID || lifecycle.profileLink.EmployeeID.Valid || lifecycle.profileLink.StudentID.Valid {
		t.Fatalf("parent link = %+v, want only parent", lifecycle.profileLink)
	}
}

func accountGenerationAdminRequest(method, target, userID string, body ...string) *http.Request {
	payload := ""
	if len(body) > 0 {
		payload = body[0]
	}
	req := httptest.NewRequest(method, target, strings.NewReader(payload))
	return withClaims(req, jwt.MapClaims{
		"uid":   userID,
		"sub":   userID,
		"roles": []any{"admin"},
	})
}
