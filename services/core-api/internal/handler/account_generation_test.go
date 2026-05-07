package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/service"
)

func TestUserStudentAccountGenerationHandlersAuthorizeAndForward(t *testing.T) {
	actorID := "01000000-0000-0000-0000-000000000123"
	svc := &fakeStudentAccountGenerationService{
		previewResult: service.StudentAccountGenerationResult{
			Total: 1,
			Ready: 1,
			Candidates: []service.StudentAccountGenerationCandidate{{
				StudentID:         "student-1",
				GeneratedUsername: "1234567890",
				Status:            "ready",
			}},
		},
		generateResult: service.StudentAccountGenerationResult{
			Total:   1,
			Created: 1,
			Candidates: []service.StudentAccountGenerationCandidate{{
				StudentID:         "student-1",
				GeneratedUsername: "1234567890",
				TemporaryPassword: "TempPass123!",
				Status:            "created",
			}},
		},
	}
	h := &User{studentGenerator: svc}

	rec := httptest.NewRecorder()
	h.PreviewStudentAccounts(rec, accountGenerationRequest(http.MethodGet, "/api/users/student-accounts/preview", actorID, "student_accounts.manage"))
	if rec.Code != http.StatusOK || !svc.previewCalled {
		t.Fatalf("PreviewStudentAccounts() status/called = %d/%v; body=%s", rec.Code, svc.previewCalled, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.GenerateStudentAccounts(rec, accountGenerationRequest(http.MethodPost, "/api/users/student-accounts/generate", actorID, "student_accounts.manage"))
	if rec.Code != http.StatusOK || svc.generateActorID.String() != actorID {
		t.Fatalf("GenerateStudentAccounts() status/actor = %d/%s; body=%s", rec.Code, svc.generateActorID.String(), rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "password_hash") {
		t.Fatalf("GenerateStudentAccounts() exposed password_hash: %s", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.PreviewStudentAccounts(rec, httptest.NewRequest(http.MethodGet, "/api/users/student-accounts/preview", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("PreviewStudentAccounts(unauthenticated) status = %d, want 401", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.PreviewStudentAccounts(rec, accountGenerationRequest(http.MethodGet, "/api/users/student-accounts/preview", actorID, "parents.read"))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("PreviewStudentAccounts(missing permission) status = %d, want 403", rec.Code)
	}
}

func TestUserParentAccountGenerationHandlersAuthorizeAndForward(t *testing.T) {
	actorID := "01000000-0000-0000-0000-000000000124"
	svc := &fakeParentAccountGenerationService{
		previewResult: service.ParentAccountGenerationResult{
			Total: 1,
			Ready: 1,
			Candidates: []service.ParentAccountGenerationCandidate{{
				ParentID:          "parent-1",
				GeneratedUsername: "ortu1234567890",
				Status:            "ready",
			}},
		},
		generateResult: service.ParentAccountGenerationResult{
			Total:   1,
			Created: 1,
			Candidates: []service.ParentAccountGenerationCandidate{{
				ParentID:          "parent-1",
				GeneratedUsername: "ortu1234567890",
				TemporaryPassword: "TempPass456!",
				Status:            "created",
			}},
		},
	}
	h := &User{parentGenerator: svc}

	rec := httptest.NewRecorder()
	h.PreviewParentAccounts(rec, accountGenerationRequest(http.MethodGet, "/api/users/parent-accounts/preview", actorID, "parent_accounts.manage"))
	if rec.Code != http.StatusOK || !svc.previewCalled {
		t.Fatalf("PreviewParentAccounts() status/called = %d/%v; body=%s", rec.Code, svc.previewCalled, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.GenerateParentAccounts(rec, accountGenerationRequest(http.MethodPost, "/api/users/parent-accounts/generate", actorID, "parent_accounts.manage"))
	if rec.Code != http.StatusOK || svc.generateActorID.String() != actorID {
		t.Fatalf("GenerateParentAccounts() status/actor = %d/%s; body=%s", rec.Code, svc.generateActorID.String(), rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "password_hash") {
		t.Fatalf("GenerateParentAccounts() exposed password_hash: %s", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.PreviewParentAccounts(rec, accountGenerationRequest(http.MethodGet, "/api/users/parent-accounts/preview", actorID, "student_accounts.manage"))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("PreviewParentAccounts(missing permission) status = %d, want 403", rec.Code)
	}
}

func TestUserForcePasswordChangeAuthorizesAndForwards(t *testing.T) {
	userID := handlerTestUUID(125)
	actorID := "01000000-0000-0000-0000-000000000126"
	lifecycle := &fakeUserLifecycle{}
	h := &User{lifecycle: lifecycle}

	rec := httptest.NewRecorder()
	req := withRouteParam(accountGenerationRequest(http.MethodPost, "/api/users/"+userID.String()+"/force-password-change", actorID, "users.reset_password"), "id", userID.String())
	h.ForcePasswordChange(rec, req)
	if rec.Code != http.StatusOK || lifecycle.forcePasswordID != userID || lifecycle.forcePasswordActorID.String() != actorID {
		t.Fatalf("ForcePasswordChange() status/id/actor = %d/%s/%s", rec.Code, lifecycle.forcePasswordID.String(), lifecycle.forcePasswordActorID.String())
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(httptest.NewRequest(http.MethodPost, "/api/users/"+userID.String()+"/force-password-change", nil), "id", userID.String())
	h.ForcePasswordChange(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("ForcePasswordChange(unauthenticated) status = %d, want 401", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(accountGenerationRequest(http.MethodPost, "/api/users/"+userID.String()+"/force-password-change", actorID, "users.read"), "id", userID.String())
	h.ForcePasswordChange(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("ForcePasswordChange(missing permission) status = %d, want 403", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(accountGenerationRequest(http.MethodPost, "/api/users/bad/force-password-change", actorID, "users.reset_password"), "id", "bad")
	h.ForcePasswordChange(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("ForcePasswordChange(invalid id) status = %d, want 400", rec.Code)
	}

	lifecycle.forcePasswordErr = errors.New("db down")
	rec = httptest.NewRecorder()
	req = withRouteParam(accountGenerationRequest(http.MethodPost, "/api/users/"+userID.String()+"/force-password-change", actorID, "users.reset_password"), "id", userID.String())
	h.ForcePasswordChange(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("ForcePasswordChange(service error) status = %d, want 500", rec.Code)
	}
}

type fakeStudentAccountGenerationService struct {
	previewCalled bool
	previewResult service.StudentAccountGenerationResult
	previewErr    error

	generateActorID pgtype.UUID
	generateResult  service.StudentAccountGenerationResult
	generateErr     error
}

func (f *fakeStudentAccountGenerationService) Preview(ctx context.Context) (service.StudentAccountGenerationResult, error) {
	f.previewCalled = true
	return f.previewResult, f.previewErr
}

func (f *fakeStudentAccountGenerationService) Generate(ctx context.Context, actorID pgtype.UUID) (service.StudentAccountGenerationResult, error) {
	f.generateActorID = actorID
	return f.generateResult, f.generateErr
}

type fakeParentAccountGenerationService struct {
	previewCalled bool
	previewResult service.ParentAccountGenerationResult
	previewErr    error

	generateActorID pgtype.UUID
	generateResult  service.ParentAccountGenerationResult
	generateErr     error
}

func (f *fakeParentAccountGenerationService) Preview(ctx context.Context) (service.ParentAccountGenerationResult, error) {
	f.previewCalled = true
	return f.previewResult, f.previewErr
}

func (f *fakeParentAccountGenerationService) Generate(ctx context.Context, actorID pgtype.UUID) (service.ParentAccountGenerationResult, error) {
	f.generateActorID = actorID
	return f.generateResult, f.generateErr
}

func accountGenerationRequest(method, target, userID, permission string) *http.Request {
	req := httptest.NewRequest(method, target, nil)
	return withClaims(req, jwt.MapClaims{
		"uid":         userID,
		"sub":         userID,
		"permissions": []any{permission},
	})
}
