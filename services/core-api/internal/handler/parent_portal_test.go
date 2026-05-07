package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestParentPortalHandlersUseAuthenticatedUserIDAndPathChildID(t *testing.T) {
	userID := handlerTestUUID(220)
	pathStudentID := handlerTestUUID(221)
	claimParentID := handlerTestUUID(222)
	queryStudentID := handlerTestUUID(223)
	svc := &fakeParentPortalSelfService{}
	h := &ParentPortal{svc: svc}
	req := withClaims(
		withRouteParam(httptest.NewRequest(http.MethodGet, "/api/portal/parent/children/"+pathStudentID.String()+"/profile?student_id="+queryStudentID.String(), nil), "studentID", pathStudentID.String()),
		jwt.MapClaims{
			"roles": []any{"ortu"},
			"sub":   userID.String(),
			"pid":   claimParentID.String(),
		},
	)
	rec := httptest.NewRecorder()

	h.ChildProfile(rec, req)

	if rec.Code != http.StatusOK || svc.profileUserID != userID || svc.profileStudentID != pathStudentID {
		t.Fatalf("ChildProfile() status/user/student = %d/%s/%s, want 200/%s/%s; body=%s", rec.Code, svc.profileUserID.String(), svc.profileStudentID.String(), userID.String(), pathStudentID.String(), rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), queryStudentID.String()) || strings.Contains(rec.Body.String(), claimParentID.String()) {
		t.Fatalf("ChildProfile() leaked query/claim scope IDs: %s", rec.Body.String())
	}
}

func TestParentPortalHandlersForbidUnlinkedChildAndRedactResultSecrets(t *testing.T) {
	userID := handlerTestUUID(224)
	studentID := handlerTestUUID(225)

	rec := httptest.NewRecorder()
	(&ParentPortal{svc: &fakeParentPortalSelfService{scheduleErr: domain.ErrForbidden}}).ChildSchedule(
		rec,
		withClaims(
			withRouteParam(httptest.NewRequest(http.MethodGet, "/api/portal/parent/children/"+studentID.String()+"/schedule", nil), "studentID", studentID.String()),
			jwt.MapClaims{"roles": []any{"ortu"}, "sub": userID.String()},
		),
	)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("ChildSchedule(unlinked) status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}

	resultsSvc := &fakeParentPortalSelfService{
		resultsRows: []db.ListParentPortalChildExamSessionsRow{{
			SessionID:    handlerTestUUID(226),
			SessionTitle: "Ujian IPA",
			PackageTitle: "Paket IPA",
		}},
	}
	rec = httptest.NewRecorder()
	(&ParentPortal{svc: resultsSvc}).ChildResults(
		rec,
		withClaims(
			withRouteParam(httptest.NewRequest(http.MethodGet, "/api/portal/parent/children/"+studentID.String()+"/results", nil), "studentID", studentID.String()),
			jwt.MapClaims{"permissions": []any{"parent_portal.read"}, "sub": userID.String()},
		),
	)
	if rec.Code != http.StatusOK || resultsSvc.resultsUserID != userID || resultsSvc.resultsStudentID != studentID {
		t.Fatalf("ChildResults() status/user/student = %d/%s/%s, want 200/%s/%s; body=%s", rec.Code, resultsSvc.resultsUserID.String(), resultsSvc.resultsStudentID.String(), userID.String(), studentID.String(), rec.Body.String())
	}
	if strings.Contains(strings.ToLower(rec.Body.String()), "token") || strings.Contains(rec.Body.String(), "secret") {
		t.Fatalf("ChildResults() exposed secret result data: %s", rec.Body.String())
	}
}

func TestParentPortalHandlersRequireAuthenticatedClaims(t *testing.T) {
	rec := httptest.NewRecorder()
	(&ParentPortal{svc: &fakeParentPortalSelfService{}}).Children(rec, httptest.NewRequest(http.MethodGet, "/api/portal/parent/children", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Children(no claims) status = %d, want 401; body=%s", rec.Code, rec.Body.String())
	}
}

type fakeParentPortalSelfService struct {
	childrenUserID pgtype.UUID
	childrenRows   []db.ListParentChildrenRow
	childrenErr    error

	profileUserID    pgtype.UUID
	profileStudentID pgtype.UUID
	profileRow       db.GetParentPortalChildProfileRow
	profileErr       error

	scheduleUserID    pgtype.UUID
	scheduleStudentID pgtype.UUID
	scheduleRows      []db.ListParentPortalChildTimetableRow
	scheduleErr       error

	resultsUserID    pgtype.UUID
	resultsStudentID pgtype.UUID
	resultsRows      []db.ListParentPortalChildExamSessionsRow
	resultsErr       error
}

func (f *fakeParentPortalSelfService) Children(ctx context.Context, userID pgtype.UUID) ([]db.ListParentChildrenRow, error) {
	f.childrenUserID = userID
	return f.childrenRows, f.childrenErr
}

func (f *fakeParentPortalSelfService) ChildProfile(ctx context.Context, userID, studentID pgtype.UUID) (db.GetParentPortalChildProfileRow, error) {
	f.profileUserID = userID
	f.profileStudentID = studentID
	if f.profileErr != nil {
		return db.GetParentPortalChildProfileRow{}, f.profileErr
	}
	if f.profileRow.ID.Valid {
		return f.profileRow, nil
	}
	return db.GetParentPortalChildProfileRow{ID: studentID, Nama: "Siswa A"}, nil
}

func (f *fakeParentPortalSelfService) ChildSchedule(ctx context.Context, userID, studentID pgtype.UUID) ([]db.ListParentPortalChildTimetableRow, error) {
	f.scheduleUserID = userID
	f.scheduleStudentID = studentID
	return f.scheduleRows, f.scheduleErr
}

func (f *fakeParentPortalSelfService) ChildResults(ctx context.Context, userID, studentID pgtype.UUID) ([]db.ListParentPortalChildExamSessionsRow, error) {
	f.resultsUserID = userID
	f.resultsStudentID = studentID
	return f.resultsRows, f.resultsErr
}
