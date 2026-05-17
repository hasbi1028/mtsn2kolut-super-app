package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestCbtSessionMoreExtraEndpointsSuccessAndAudits(t *testing.T) {
	sessionID := handlerTestUUID(180)
	participantID := handlerTestUUID(181)
	roomID := handlerTestUUID(182)
	actorID := handlerTestUUID(183)
	employeeID := handlerTestUUID(184)
	teacherID := handlerTestUUID(185)

	fake := &fakeCbtSessionService{checkAllowed: true, roomProctorAllowed: true}
	audit := &fakeCbtSessionAuditWriter{}
	h := &CbtSession{svc: fake, audit: audit}

	adminRoute := func(method, target, body string, pairs ...string) *http.Request {
		return withRouteParams(withClaims(httptest.NewRequest(method, target, strings.NewReader(body)), jwt.MapClaims{
			"roles": []any{"admin"},
			"uid":   actorID.String(),
			"sub":   actorID.String(),
			"eid":   employeeID.String(),
			"usr":   "admin.extra",
			"ssid":  "session-extra-admin",
		}), pairs...)
	}
	teacherRoute := func(method, target, body string, pairs ...string) *http.Request {
		return withRouteParams(withClaims(httptest.NewRequest(method, target, strings.NewReader(body)), jwt.MapClaims{
			"roles": []any{"guru"},
			"uid":   actorID.String(),
			"sub":   actorID.String(),
			"eid":   teacherID.String(),
			"usr":   "guru.extra",
			"ssid":  "session-extra-teacher",
		}), pairs...)
	}
	roomPairs := []string{"id", sessionID.String(), "rid", roomID.String()}

	rec := httptest.NewRecorder()
	h.ResetParticipantAccess(rec, adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/participants/"+participantID.String()+"/reset-access", `{}`, "id", sessionID.String(), "pid", participantID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("ResetParticipantAccess status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.proctorActionInput.SessionID != sessionID || fake.proctorActionInput.ParticipantID != participantID || fake.proctorActionInput.ActionType != "reset_device_binding" || fake.proctorActionInput.Reason != "Reset akses peserta" {
		t.Fatalf("ResetParticipantAccess action = %+v, want default reset action", fake.proctorActionInput)
	}

	rec = httptest.NewRecorder()
	h.UpdateSchedule(rec, adminRoute(http.MethodPatch, "/api/cbt/sessions/"+sessionID.String()+"/schedule", `{"scheduled_start":"2026-05-03T08:00:00Z","scheduled_end":"2026-05-03T09:00:00Z"}`, "id", sessionID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateSchedule status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.updateScheduleID != sessionID || !fake.updateScheduleStart.Valid || fake.updateScheduleStart.Time.Format(time.RFC3339) != "2026-05-03T08:00:00Z" {
		t.Fatalf("UpdateSchedule args = %v/%+v, want parsed start", fake.updateScheduleID, fake.updateScheduleStart)
	}

	rec = httptest.NewRecorder()
	h.GetRoomReadiness(rec, adminRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/rooms/readiness", "", "id", sessionID.String()))
	if rec.Code != http.StatusOK || fake.roomReadinessSessionID != sessionID {
		t.Fatalf("GetRoomReadiness status/arg = %d/%v, want 200/session; body=%s", rec.Code, fake.roomReadinessSessionID, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.SaveRoomHandover(rec, adminRoute(http.MethodPut, "/api/cbt/sessions/"+sessionID.String()+"/rooms/"+roomID.String()+"/handover", `{"attendance_checked":true,"all_submitted_checked":true,"device_issue_checked":true,"room_clean_checked":true,"token_returned_checked":true,"assets_returned_checked":true,"incident_notes":"incident","operator_notes":"operator","handover_notes":"handover"}`, roomPairs...))
	if rec.Code != http.StatusOK {
		t.Fatalf("SaveRoomHandover status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.saveHandoverRoomID != roomID || fake.saveHandoverUpdatedBy != actorID || !fake.saveHandoverInput.AttendanceChecked || fake.saveHandoverInput.OperatorNotes != "operator" {
		t.Fatalf("SaveRoomHandover args = room:%v by:%v input:%+v, want payload and actor", fake.saveHandoverRoomID, fake.saveHandoverUpdatedBy, fake.saveHandoverInput)
	}

	rec = httptest.NewRecorder()
	h.LockRoomHandover(rec, adminRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/rooms/"+roomID.String()+"/handover/lock", "", roomPairs...))
	if rec.Code != http.StatusOK {
		t.Fatalf("LockRoomHandover status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.lockHandoverRoomID != roomID || fake.lockHandoverLockedBy != actorID {
		t.Fatalf("LockRoomHandover args = room:%v by:%v, want room and actor", fake.lockHandoverRoomID, fake.lockHandoverLockedBy)
	}

	fake.gradeSyncPreflightRow = db.GetCbtSessionGradeSyncPreflightRow{SessionID: sessionID, SessionTitle: "Sesi Extra", ParticipantCount: 3, MissingScoreCount: 0}
	rec = httptest.NewRecorder()
	h.GetGradeSyncPreflight(rec, teacherRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/grade-sync/preflight", "", "id", sessionID.String()))
	if rec.Code != http.StatusOK || fake.gradeSyncPreflightSessionID != sessionID || fake.checkTeacherID != teacherID {
		t.Fatalf("GetGradeSyncPreflight status/args = %d/%v/%v; body=%s", rec.Code, fake.gradeSyncPreflightSessionID, fake.checkTeacherID, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Sesi Extra") || !strings.Contains(rec.Body.String(), `"participant_count":3`) {
		t.Fatalf("GetGradeSyncPreflight body = %s, want serialized row", rec.Body.String())
	}

	wantAudit := []string{"CBT_SESSION_PARTICIPANT_RESET_ACCESS", "CBT_SESSION_SCHEDULE_UPDATE", "CBT_SESSION_ROOM_HANDOVER_SAVE", "CBT_SESSION_ROOM_HANDOVER_LOCK"}
	if len(audit.entries) != len(wantAudit) {
		t.Fatalf("audit entries len = %d, want %d: %+v", len(audit.entries), len(wantAudit), audit.entries)
	}
	for i, action := range wantAudit {
		if audit.entries[i].Action != action {
			t.Fatalf("audit[%d].Action = %q, want %q", i, audit.entries[i].Action, action)
		}
	}
}

func TestCbtSessionMoreExtraEndpointsErrorBranches(t *testing.T) {
	sessionID := handlerTestUUID(186)
	participantID := handlerTestUUID(187)
	roomID := handlerTestUUID(188)

	adminRoute := func(method, target, body string, pairs ...string) *http.Request {
		return withRouteParams(adminRequest(method, target, body), pairs...)
	}
	roomPairs := []string{"id", sessionID.String(), "rid", roomID.String()}

	tests := []struct {
		name       string
		handler    func(*CbtSession, http.ResponseWriter, *http.Request)
		svc        *fakeCbtSessionService
		req        *http.Request
		wantStatus int
	}{
		{name: "reset invalid participant id", handler: (*CbtSession).ResetParticipantAccess, req: adminRoute(http.MethodPost, "/", `{}`, "id", sessionID.String(), "pid", "bad"), wantStatus: http.StatusBadRequest},
		{name: "reset action bad request", handler: (*CbtSession).ResetParticipantAccess, svc: &fakeCbtSessionService{proctorActionErr: errors.Join(domain.ErrBadRequest, errors.New("alasan wajib diisi"))}, req: adminRoute(http.MethodPost, "/", `{"reason":"Reset"}`, "id", sessionID.String(), "pid", participantID.String()), wantStatus: http.StatusBadRequest},
		{name: "update schedule invalid json", handler: (*CbtSession).UpdateSchedule, req: adminRoute(http.MethodPatch, "/", `{`, "id", sessionID.String()), wantStatus: http.StatusBadRequest},
		{name: "update schedule bad start", handler: (*CbtSession).UpdateSchedule, req: adminRoute(http.MethodPatch, "/", `{"scheduled_start":"bad","scheduled_end":"2026-05-03T09:00:00Z"}`, "id", sessionID.String()), wantStatus: http.StatusBadRequest},
		{name: "update schedule end before start", handler: (*CbtSession).UpdateSchedule, req: adminRoute(http.MethodPatch, "/", `{"scheduled_start":"2026-05-03T10:00:00Z","scheduled_end":"2026-05-03T09:00:00Z"}`, "id", sessionID.String()), wantStatus: http.StatusBadRequest},
		{name: "update schedule conflict", handler: (*CbtSession).UpdateSchedule, svc: &fakeCbtSessionService{updateScheduleErr: errors.Join(domain.ErrConflict, errors.New("sesi sudah aktif"))}, req: adminRoute(http.MethodPatch, "/", `{"scheduled_start":"2026-05-03T08:00:00Z","scheduled_end":"2026-05-03T09:00:00Z"}`, "id", sessionID.String()), wantStatus: http.StatusConflict},
		{name: "update schedule not found", handler: (*CbtSession).UpdateSchedule, svc: &fakeCbtSessionService{updateScheduleErr: pgx.ErrNoRows}, req: adminRoute(http.MethodPatch, "/", `{"scheduled_start":"2026-05-03T08:00:00Z","scheduled_end":"2026-05-03T09:00:00Z"}`, "id", sessionID.String()), wantStatus: http.StatusNotFound},
		{name: "save handover invalid json", handler: (*CbtSession).SaveRoomHandover, req: adminRoute(http.MethodPut, "/", `{`, roomPairs...), wantStatus: http.StatusBadRequest},
		{name: "save handover conflict", handler: (*CbtSession).SaveRoomHandover, svc: &fakeCbtSessionService{saveHandoverErr: errors.Join(domain.ErrConflict, errors.New("sudah dikunci"))}, req: adminRoute(http.MethodPut, "/", `{}`, roomPairs...), wantStatus: http.StatusConflict},
		{name: "lock handover bad request", handler: (*CbtSession).LockRoomHandover, svc: &fakeCbtSessionService{lockHandoverErr: errors.Join(domain.ErrBadRequest, errors.New("checklist belum lengkap"))}, req: adminRoute(http.MethodPost, "/", "", roomPairs...), wantStatus: http.StatusBadRequest},
		{name: "grade sync preflight service error", handler: (*CbtSession).GetGradeSyncPreflight, svc: &fakeCbtSessionService{gradeSyncPreflightErr: errors.Join(domain.ErrBadRequest, errors.New("nilai belum lengkap"))}, req: adminRoute(http.MethodGet, "/", "", "id", sessionID.String()), wantStatus: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.svc
			if svc == nil {
				svc = &fakeCbtSessionService{}
			}
			rec := httptest.NewRecorder()
			tt.handler(&CbtSession{svc: svc}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
