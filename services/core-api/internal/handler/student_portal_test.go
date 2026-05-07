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

func TestStudentPortalHandlersUseAuthenticatedUserIDAndIgnoreStudentIDQuery(t *testing.T) {
	userID := handlerTestUUID(210)
	otherStudentID := handlerTestUUID(211)
	svc := &fakeStudentPortalSelfService{}
	h := &StudentPortal{svc: svc}
	req := withClaims(httptest.NewRequest(http.MethodGet, "/api/portal/student/profile?student_id="+otherStudentID.String(), nil), jwt.MapClaims{
		"roles": []any{"siswa"},
		"sub":   userID.String(),
		"sid":   otherStudentID.String(),
	})
	rec := httptest.NewRecorder()

	h.Profile(rec, req)

	if rec.Code != http.StatusOK || svc.profileUserID != userID {
		t.Fatalf("Profile() status/userID = %d/%s, want 200/%s; body=%s", rec.Code, svc.profileUserID.String(), userID.String(), rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), otherStudentID.String()) {
		t.Fatalf("Profile() leaked query/claim student_id: %s", rec.Body.String())
	}
}

func TestStudentPortalHandlersRejectUnlinkedUsersAndRedactResultTokens(t *testing.T) {
	userID := handlerTestUUID(212)

	rec := httptest.NewRecorder()
	(&StudentPortal{svc: &fakeStudentPortalSelfService{profileErr: domain.ErrForbidden}}).Profile(
		rec,
		withClaims(httptest.NewRequest(http.MethodGet, "/api/portal/student/profile", nil), jwt.MapClaims{"roles": []any{"siswa"}, "sub": userID.String()}),
	)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("Profile(unlinked) status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}

	resultsSvc := &fakeStudentPortalSelfService{
		resultsRows: []db.ListStudentExamSessionsRow{{
			SessionID:    handlerTestUUID(213),
			SessionTitle: "Ujian IPA",
			Token:        "secret-exam-token",
		}},
	}
	rec = httptest.NewRecorder()
	(&StudentPortal{svc: resultsSvc}).Results(
		rec,
		withClaims(httptest.NewRequest(http.MethodGet, "/api/portal/student/results?student_id=other", nil), jwt.MapClaims{"permissions": []any{"student_portal.read"}, "sub": userID.String()}),
	)
	if rec.Code != http.StatusOK || resultsSvc.resultsUserID != userID {
		t.Fatalf("Results() status/userID = %d/%s, want 200/%s; body=%s", rec.Code, resultsSvc.resultsUserID.String(), userID.String(), rec.Body.String())
	}
	if strings.Contains(strings.ToLower(rec.Body.String()), "token") || strings.Contains(rec.Body.String(), "secret-exam-token") {
		t.Fatalf("Results() exposed token data: %s", rec.Body.String())
	}
}

func TestStudentPortalHandlersRequireAuthenticatedClaims(t *testing.T) {
	rec := httptest.NewRecorder()
	(&StudentPortal{svc: &fakeStudentPortalSelfService{}}).Schedule(rec, httptest.NewRequest(http.MethodGet, "/api/portal/student/schedule", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Schedule(no claims) status = %d, want 401; body=%s", rec.Code, rec.Body.String())
	}
}

type fakeStudentPortalSelfService struct {
	profileUserID pgtype.UUID
	profileRow    db.GetStudentByIDRow
	profileErr    error

	scheduleUserID pgtype.UUID
	scheduleRows   []db.ListStudentTimetableRow
	scheduleErr    error

	resultsUserID pgtype.UUID
	resultsRows   []db.ListStudentExamSessionsRow
	resultsErr    error
}

func (f *fakeStudentPortalSelfService) Profile(ctx context.Context, userID pgtype.UUID) (db.GetStudentByIDRow, error) {
	f.profileUserID = userID
	if f.profileErr != nil {
		return db.GetStudentByIDRow{}, f.profileErr
	}
	if f.profileRow.ID.Valid {
		return f.profileRow, nil
	}
	return db.GetStudentByIDRow{ID: handlerTestUUID(214), Nama: "Siswa A"}, nil
}

func (f *fakeStudentPortalSelfService) Schedule(ctx context.Context, userID pgtype.UUID) ([]db.ListStudentTimetableRow, error) {
	f.scheduleUserID = userID
	return f.scheduleRows, f.scheduleErr
}

func (f *fakeStudentPortalSelfService) Results(ctx context.Context, userID pgtype.UUID) ([]db.ListStudentExamSessionsRow, error) {
	f.resultsUserID = userID
	return f.resultsRows, f.resultsErr
}
