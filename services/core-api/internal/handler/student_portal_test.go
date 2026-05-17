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

func TestStudentPortalPreviewStudentsMapsNullableClassAndStatus(t *testing.T) {
	studentID := handlerTestUUID(237)
	classID := handlerTestUUID(238)
	svc := &fakeStudentPortalSelfService{
		previewRows: []db.ListStudentPortalPreviewStudentsRow{
			{ID: studentID, Nis: "2001", Nisn: "9988", Nama: "Siswa Preview", ClassID: classID, ClassName: pgtype.Text{String: "VIII B", Valid: true}, ClassCode: pgtype.Text{String: "8B", Valid: true}, Status: db.StudentStatusEnumActive, IsActive: true},
			{ID: handlerTestUUID(239), Nis: "2002", Nama: "Tanpa Kelas", Status: db.StudentStatusEnumProspective},
		},
	}
	rec := httptest.NewRecorder()

	(&StudentPortal{svc: svc}).PreviewStudents(rec, httptest.NewRequest(http.MethodGet, "/api/portal/student/preview/students", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("PreviewStudents() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Data struct {
			Students []struct {
				ID        string `json:"id"`
				NIS       string `json:"nis"`
				NISN      string `json:"nisn"`
				Nama      string `json:"nama"`
				ClassID   string `json:"class_id"`
				ClassName string `json:"class_name"`
				ClassCode string `json:"class_code"`
				Status    string `json:"status"`
				IsActive  bool   `json:"is_active"`
			} `json:"students"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("PreviewStudents() json unmarshal failed: %v", err)
	}
	if len(payload.Data.Students) != 2 || payload.Data.Students[0].ID != studentID.String() || payload.Data.Students[0].ClassID != classID.String() || payload.Data.Students[0].ClassName != "VIII B" || payload.Data.Students[0].ClassCode != "8B" || payload.Data.Students[0].Status != "active" || !payload.Data.Students[0].IsActive {
		t.Fatalf("PreviewStudents() first student = %+v, want full mapped DTO", payload.Data.Students)
	}
	if payload.Data.Students[1].ClassID != "" || payload.Data.Students[1].ClassName != "" || payload.Data.Students[1].ClassCode != "" || payload.Data.Students[1].Status != "prospective" {
		t.Fatalf("PreviewStudents() nullable student = %+v, want blank nullable class fields", payload.Data.Students[1])
	}
}

func TestStudentPortalPreviewStudentIDValidationAndProfileScope(t *testing.T) {
	studentID := handlerTestUUID(240)
	svc := &fakeStudentPortalSelfService{}
	h := &StudentPortal{svc: svc}
	req := withRouteParam(httptest.NewRequest(http.MethodGet, "/api/portal/student/preview/students/"+studentID.String()+"/profile", nil), "studentID", "  "+studentID.String()+"  ")
	rec := httptest.NewRecorder()

	h.PreviewProfile(rec, req)

	if rec.Code != http.StatusOK || svc.profileUserID != studentID {
		t.Fatalf("PreviewProfile() status/studentID = %d/%s, want 200/%s; body=%s", rec.Code, svc.profileUserID.String(), studentID.String(), rec.Body.String())
	}
	var payload struct {
		Data struct {
			Preview bool `json:"preview"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("PreviewProfile() json unmarshal failed: %v", err)
	}
	if !payload.Data.Preview {
		t.Fatalf("PreviewProfile() preview flag = false, body=%s", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.PreviewProfile(rec, withRouteParam(httptest.NewRequest(http.MethodGet, "/api/portal/student/preview/students/bad/profile", nil), "studentID", "bad"))
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "ID siswa tidak valid") {
		t.Fatalf("PreviewProfile(invalid id) status/body = %d/%s, want 400/student error", rec.Code, rec.Body.String())
	}
}

func TestStudentPortalRemainingPreviewEndpointsUseStudentRouteID(t *testing.T) {
	studentID := handlerTestUUID(248)
	participantID := handlerTestUUID(249)
	tokenMasked := "MASK-••••"
	svc := &fakeStudentPortalSelfService{
		scheduleRows: []db.ListStudentTimetableRow{{
			ID:          handlerTestUUID(250),
			SubjectName: "Matematika",
			TeacherName: "Guru A",
		}},
		resultsRows: []db.ListStudentExamSessionsRow{{
			SessionID:    handlerTestUUID(251),
			SessionTitle: "Hasil Preview",
			Token:        "secret-token",
		}},
		cbtScheduleRows: []service.StudentPortalCbtScheduleItem{{
			ParticipantID:  participantID.String(),
			SessionTitle:   "CBT Preview",
			CanRevealToken: true,
			TokenMasked:    &tokenMasked,
			PackageTitle:   "Paket CBT",
			Status:         service.StudentPortalCbtTokenWindow,
		}},
	}
	h := &StudentPortal{svc: svc}

	tests := []struct {
		name         string
		handler      func(http.ResponseWriter, *http.Request)
		path         string
		wantFragment string
	}{
		{name: "schedule", handler: h.PreviewSchedule, path: "/api/portal/student/preview/students/" + studentID.String() + "/schedule", wantFragment: "Matematika"},
		{name: "results", handler: h.PreviewResults, path: "/api/portal/student/preview/students/" + studentID.String() + "/results", wantFragment: "Hasil Preview"},
		{name: "cbt schedule", handler: h.PreviewCbtSchedule, path: "/api/portal/student/preview/students/" + studentID.String() + "/cbt", wantFragment: "CBT Preview"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := withRouteParam(httptest.NewRequest(http.MethodGet, tt.path, nil), "studentID", studentID.String())
			tt.handler(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
			}
			body := rec.Body.String()
			if !strings.Contains(body, `"preview":true`) || !strings.Contains(body, tt.wantFragment) {
				t.Fatalf("body = %s, want preview flag and %q", body, tt.wantFragment)
			}
			if strings.Contains(body, "secret-token") || strings.Contains(body, `"token":"`) {
				t.Fatalf("preview endpoint exposed raw token: %s", body)
			}
		})
	}
	if svc.scheduleUserID != studentID || svc.resultsUserID != studentID || svc.cbtUserID != studentID {
		t.Fatalf("preview args schedule/results/cbt = %s/%s/%s, want %s", svc.scheduleUserID.String(), svc.resultsUserID.String(), svc.cbtUserID.String(), studentID.String())
	}

	rec := httptest.NewRecorder()
	h.PreviewSchedule(rec, withRouteParam(httptest.NewRequest(http.MethodGet, "/api/portal/student/preview/students/bad/schedule", nil), "studentID", "bad"))
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "ID siswa tidak valid") {
		t.Fatalf("PreviewSchedule(invalid id) status/body = %d/%s, want 400/student error", rec.Code, rec.Body.String())
	}
}

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

func TestStudentPortalScheduleSuccessForbiddenAndInternalError(t *testing.T) {
	userID := handlerTestUUID(254)
	scheduleID := handlerTestUUID(255)
	svc := &fakeStudentPortalSelfService{
		scheduleRows: []db.ListStudentTimetableRow{{ID: scheduleID, SubjectName: "Bahasa Indonesia", TeacherName: "Guru B"}},
	}
	req := withClaims(httptest.NewRequest(http.MethodGet, "/api/portal/student/schedule", nil), jwt.MapClaims{"roles": []any{"siswa"}, "sub": userID.String()})
	rec := httptest.NewRecorder()

	(&StudentPortal{svc: svc}).Schedule(rec, req)

	if rec.Code != http.StatusOK || svc.scheduleUserID != userID {
		t.Fatalf("Schedule() status/user = %d/%s, want 200/%s; body=%s", rec.Code, svc.scheduleUserID.String(), userID.String(), rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), scheduleID.String()) || !strings.Contains(rec.Body.String(), "Bahasa Indonesia") {
		t.Fatalf("Schedule() body = %s, want schedule row", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	(&StudentPortal{svc: &fakeStudentPortalSelfService{}}).Schedule(rec, withClaims(httptest.NewRequest(http.MethodGet, "/api/portal/student/schedule", nil), jwt.MapClaims{"roles": []any{"ortu"}, "sub": userID.String()}))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("Schedule(forbidden role) status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	(&StudentPortal{svc: &fakeStudentPortalSelfService{scheduleErr: context.DeadlineExceeded}}).Schedule(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("Schedule(internal) status = %d, want 500; body=%s", rec.Code, rec.Body.String())
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
	previewRows []db.ListStudentPortalPreviewStudentsRow
	previewErr  error

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

func (f *fakeStudentPortalSelfService) PreviewStudents(ctx context.Context) ([]db.ListStudentPortalPreviewStudentsRow, error) {
	return f.previewRows, f.previewErr
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

func (f *fakeStudentPortalSelfService) ProfileByStudentID(ctx context.Context, studentID pgtype.UUID) (db.GetStudentByIDRow, error) {
	return f.Profile(ctx, studentID)
}

func (f *fakeStudentPortalSelfService) Schedule(ctx context.Context, userID pgtype.UUID) ([]db.ListStudentTimetableRow, error) {
	f.scheduleUserID = userID
	return f.scheduleRows, f.scheduleErr
}

func (f *fakeStudentPortalSelfService) ScheduleByStudentID(ctx context.Context, studentID pgtype.UUID) ([]db.ListStudentTimetableRow, error) {
	return f.Schedule(ctx, studentID)
}

func (f *fakeStudentPortalSelfService) Results(ctx context.Context, userID pgtype.UUID) ([]db.ListStudentExamSessionsRow, error) {
	f.resultsUserID = userID
	return f.resultsRows, f.resultsErr
}

func (f *fakeStudentPortalSelfService) ResultsByStudentID(ctx context.Context, studentID pgtype.UUID) ([]db.ListStudentExamSessionsRow, error) {
	return f.Results(ctx, studentID)
}

func (f *fakeStudentPortalSelfService) CbtSchedule(ctx context.Context, userID pgtype.UUID) ([]service.StudentPortalCbtScheduleItem, error) {
	f.cbtUserID = userID
	return f.cbtScheduleRows, f.cbtScheduleErr
}

func (f *fakeStudentPortalSelfService) CbtScheduleByStudentID(ctx context.Context, studentID pgtype.UUID) ([]service.StudentPortalCbtScheduleItem, error) {
	return f.CbtSchedule(ctx, studentID)
}

func (f *fakeStudentPortalSelfService) RevealCbtToken(ctx context.Context, userID, participantID pgtype.UUID, roomToken, clientIP string) (service.StudentPortalTokenReveal, error) {
	f.revealUserID = userID
	f.revealParticipantID = participantID
	f.revealRoomToken = roomToken
	f.revealIP = clientIP
	return f.revealResult, f.revealErr
}
