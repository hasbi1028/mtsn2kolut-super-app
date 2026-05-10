package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
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

func TestStudentPortalCbtHandlersUseAuthenticatedUserAndRedactScheduleToken(t *testing.T) {
	userID := handlerTestUUID(215)
	tokenMasked := "ABCD-••••"
	svc := &fakeStudentPortalSelfService{
		cbtScheduleRows: []service.StudentPortalCbtScheduleItem{{
			ParticipantID:     handlerTestUUID(216).String(),
			SessionID:         handlerTestUUID(217).String(),
			SessionTitle:      "Ujian IPA",
			PackageTitle:      "Paket IPA",
			Status:            service.StudentPortalCbtTokenWindow,
			CanRevealToken:    true,
			RequiresRoomToken: true,
			TokenMasked:       &tokenMasked,
		}},
	}
	rec := httptest.NewRecorder()
	(&StudentPortal{svc: svc}).CbtSchedule(
		rec,
		withClaims(httptest.NewRequest(http.MethodGet, "/api/portal/student/cbt?student_id=other", nil), jwt.MapClaims{"roles": []any{"siswa"}, "sub": userID.String()}),
	)
	if rec.Code != http.StatusOK || svc.cbtUserID != userID {
		t.Fatalf("CbtSchedule() status/userID = %d/%s, want 200/%s; body=%s", rec.Code, svc.cbtUserID.String(), userID.String(), rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "secret") || strings.Contains(rec.Body.String(), "token\":\"") {
		t.Fatalf("CbtSchedule() exposed raw token-like field: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "token_masked") {
		t.Fatalf("CbtSchedule() body = %s, want token_masked", rec.Body.String())
	}
}

func TestStudentPortalRevealTokenHandlerValidatesRoomTokenAndMapsErrors(t *testing.T) {
	userID := handlerTestUUID(218)
	participantID := handlerTestUUID(219)
	svc := &fakeStudentPortalSelfService{
		revealResult: service.StudentPortalTokenReveal{Token: "student-token", ExpiresAt: "2026-05-11T00:15:00+08:00"},
	}
	req := withClaims(httptest.NewRequest(http.MethodPost, "/api/portal/student/cbt/"+participantID.String()+"/reveal-token", bytes.NewBufferString(`{"room_token":"ROOM-1"}`)), jwt.MapClaims{"permissions": []any{"student_portal.read"}, "sub": userID.String()})
	req = withRouteParam(req, "participantID", participantID.String())
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "203.0.113.10:2222"
	rec := httptest.NewRecorder()

	(&StudentPortal{svc: svc}).RevealCbtToken(rec, req)

	if rec.Code != http.StatusOK || svc.revealUserID != userID || svc.revealParticipantID != participantID || svc.revealRoomToken != "ROOM-1" {
		t.Fatalf("RevealCbtToken() status/args = %d/%s/%s/%q; body=%s", rec.Code, svc.revealUserID.String(), svc.revealParticipantID.String(), svc.revealRoomToken, rec.Body.String())
	}
	var payload struct {
		Data service.StudentPortalTokenReveal `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if payload.Data.Token != "student-token" {
		t.Fatalf("token = %q, want student-token", payload.Data.Token)
	}

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantError  string
	}{
		{name: "room token required", err: service.ErrRoomTokenRequired, wantStatus: http.StatusBadRequest, wantError: "room token required"},
		{name: "room token mismatch", err: service.ErrRoomTokenMismatch, wantStatus: http.StatusForbidden, wantError: "room token mismatch"},
		{name: "forbidden", err: domain.ErrForbidden, wantStatus: http.StatusForbidden, wantError: "forbidden"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			svc := &fakeStudentPortalSelfService{revealErr: tt.err}
			req := withClaims(httptest.NewRequest(http.MethodPost, "/api/portal/student/cbt/"+participantID.String()+"/reveal-token", bytes.NewBufferString(`{"room_token":"ROOM-1"}`)), jwt.MapClaims{"roles": []any{"siswa"}, "sub": userID.String()})
			req = withRouteParam(req, "participantID", participantID.String())
			req.Header.Set("Content-Type", "application/json")

			(&StudentPortal{svc: svc}).RevealCbtToken(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			var payload struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
				t.Fatalf("json unmarshal failed: %v", err)
			}
			if payload.Error != tt.wantError {
				t.Fatalf("error = %q, want %q", payload.Error, tt.wantError)
			}
		})
	}

	rec = httptest.NewRecorder()
	req = withClaims(httptest.NewRequest(http.MethodPost, "/api/portal/student/cbt/bad/reveal-token", bytes.NewBufferString(`{"room_token":"ROOM-1"}`)), jwt.MapClaims{"roles": []any{"siswa"}, "sub": userID.String()})
	req = withRouteParam(req, "participantID", "bad")
	(&StudentPortal{svc: &fakeStudentPortalSelfService{}}).RevealCbtToken(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("RevealCbtToken(invalid id) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = withClaims(httptest.NewRequest(http.MethodPost, "/api/portal/student/cbt/"+participantID.String()+"/reveal-token", bytes.NewBufferString(`{`)), jwt.MapClaims{"roles": []any{"siswa"}, "sub": userID.String()})
	req = withRouteParam(req, "participantID", participantID.String())
	(&StudentPortal{svc: &fakeStudentPortalSelfService{}}).RevealCbtToken(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("RevealCbtToken(invalid json) status = %d, want 400; body=%s", rec.Code, rec.Body.String())
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

	cbtUserID       pgtype.UUID
	cbtScheduleRows []service.StudentPortalCbtScheduleItem
	cbtScheduleErr  error

	revealUserID        pgtype.UUID
	revealParticipantID pgtype.UUID
	revealRoomToken     string
	revealIP            string
	revealResult        service.StudentPortalTokenReveal
	revealErr           error
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

func (f *fakeStudentPortalSelfService) CbtSchedule(ctx context.Context, userID pgtype.UUID) ([]service.StudentPortalCbtScheduleItem, error) {
	f.cbtUserID = userID
	return f.cbtScheduleRows, f.cbtScheduleErr
}

func (f *fakeStudentPortalSelfService) RevealCbtToken(ctx context.Context, userID, participantID pgtype.UUID, roomToken, clientIP string) (service.StudentPortalTokenReveal, error) {
	f.revealUserID = userID
	f.revealParticipantID = participantID
	f.revealRoomToken = roomToken
	f.revealIP = clientIP
	return f.revealResult, f.revealErr
}
