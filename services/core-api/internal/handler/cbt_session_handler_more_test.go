package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type cbtSessionMoreService struct {
	*fakeCbtSessionService

	roomDashboardErr      error
	listRoomProctorsErr   error
	roomProctoringRows    []db.GetSessionProctoringStatusRow
	roomProctoringErr     error
	roomEventsErr         error
	roomProctorAllowedErr error
}

func (f *cbtSessionMoreService) GetRoomProctoringDashboard(_ context.Context, roomID pgtype.UUID) (db.GetCbtRoomProctorDashboardRow, error) {
	f.roomDashboardID = roomID
	if f.roomDashboardErr != nil {
		return db.GetCbtRoomProctorDashboardRow{}, f.roomDashboardErr
	}
	return db.GetCbtRoomProctorDashboardRow{ID: roomID, RoomName: "Ruang Edge"}, nil
}

func (f *cbtSessionMoreService) ListRoomProctors(_ context.Context, roomID pgtype.UUID) ([]db.ListCbtRoomProctorsRow, error) {
	f.listRoomProctorsRoomID = roomID
	if f.listRoomProctorsErr != nil {
		return nil, f.listRoomProctorsErr
	}
	return []db.ListCbtRoomProctorsRow{{ExamRoomID: roomID, Nama: "Pengawas Edge"}}, nil
}

func (f *cbtSessionMoreService) HasRoomProctor(_ context.Context, sessionID, roomID, employeeID pgtype.UUID) (bool, error) {
	f.roomProctorSessionID = sessionID
	f.roomProctorRoomID = roomID
	f.roomProctorEmployeeID = employeeID
	if f.roomProctorAllowedErr != nil {
		return false, f.roomProctorAllowedErr
	}
	return f.roomProctorAllowed, nil
}

func (f *cbtSessionMoreService) GetProctoringStatusForRoom(_ context.Context, sessionID, roomID pgtype.UUID) ([]db.GetSessionProctoringStatusRow, error) {
	f.roomProctoringSessionID = sessionID
	f.roomProctoringRoomID = roomID
	if f.roomProctoringErr != nil {
		return nil, f.roomProctoringErr
	}
	return f.roomProctoringRows, nil
}

func (f *cbtSessionMoreService) ListParticipantEventsForRoom(_ context.Context, sessionID, participantID, roomID pgtype.UUID, limit int32) ([]db.ListSessionParticipantEventsRow, error) {
	f.roomEventsSessionID = sessionID
	f.roomEventsParticipantID = participantID
	f.roomEventsRoomID = roomID
	f.roomEventsLimit = limit
	if f.roomEventsErr != nil {
		return nil, f.roomEventsErr
	}
	return []db.ListSessionParticipantEventsRow{}, nil
}

func TestSerializeParticipantAnswerRowsRedactsAnswerKeyUnlessIncluded(t *testing.T) {
	answeredAt := pgtype.Timestamptz{Time: time.Date(2026, 5, 17, 8, 30, 0, 0, time.UTC), Valid: true}
	rows := []db.GetParticipantAnswersRow{{
		ID:            handlerTestUUID(201),
		ParticipantID: handlerTestUUID(202),
		QuestionID:    handlerTestUUID(203),
		QuestionCode:  "Q-EDGE",
		QuestionText:  "Pilih jawaban benar",
		OptionA:       "A",
		OptionB:       "B",
		AnswerKey:     "B",
		Answer:        "A",
		IsCorrect:     pgtype.Bool{Bool: false, Valid: true},
		AnsweredAt:    answeredAt,
	}}

	redacted := serializeParticipantAnswerRows(rows, false)
	if len(redacted) != 1 || redacted[0]["answer_key"] != "" {
		t.Fatalf("redacted answer_key = %#v, want empty", redacted)
	}
	if redacted[0]["id"] != rows[0].ID.String() || redacted[0]["answered_at"] != answeredAt || redacted[0]["is_correct"] != rows[0].IsCorrect {
		t.Fatalf("serialized answer row = %#v, want UUID/time/bool preserved", redacted[0])
	}

	included := serializeParticipantAnswerRows(rows, true)
	if included[0]["answer_key"] != "B" {
		t.Fatalf("included answer_key = %#v, want B", included[0]["answer_key"])
	}
}

func TestCbtSessionForceSubmitParticipantCompatibilityEdges(t *testing.T) {
	sessionID := handlerTestUUID(204)
	participantID := handlerTestUUID(205)
	actorID := handlerTestUUID(206)
	audit := &fakeCbtSessionAuditWriter{}
	fake := &fakeCbtSessionService{hasParticipantSet: true, hasParticipant: true}
	h := &CbtSession{svc: fake, audit: audit}
	pairs := []string{"id", sessionID.String(), "pid", participantID.String()}

	req := withRouteParams(withClaims(httptest.NewRequest(http.MethodPost, "/", strings.NewReader("")), jwt.MapClaims{
		"roles": []any{"admin"}, "uid": actorID.String(), "sub": actorID.String(), "usr": "admin.cbt", "ssid": "force-submit-1",
	}), pairs...)
	rec := httptest.NewRecorder()
	h.ForceSubmitParticipant(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("ForceSubmitParticipant status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.proctorActionInput.ActionType != "force_submit" || fake.proctorActionInput.Reason != "Kirim paksa jawaban" || fake.proctorActionInput.Notes == "" || fake.proctorActionActor.UserID != actorID {
		t.Fatalf("force submit action = input:%+v actor:%+v, want default compatibility action", fake.proctorActionInput, fake.proctorActionActor)
	}
	if len(audit.entries) != 1 || audit.entries[0].Action != "CBT_SESSION_PARTICIPANT_FORCE_SUBMIT" || audit.entries[0].EntityID != pgUUIDString(sessionID) {
		t.Fatalf("force submit audit = %+v, want force-submit audit on session", audit.entries)
	}

	cases := []struct {
		name       string
		svc        *fakeCbtSessionService
		req        *http.Request
		wantStatus int
	}{
		{name: "non admin forbidden before participant lookup", svc: &fakeCbtSessionService{}, req: withRouteParams(httptest.NewRequest(http.MethodPost, "/", nil), pairs...), wantStatus: http.StatusForbidden},
		{name: "invalid participant id", svc: &fakeCbtSessionService{}, req: withRouteParams(adminRequest(http.MethodPost, "/", ""), "id", sessionID.String(), "pid", "bad"), wantStatus: http.StatusBadRequest},
		{name: "participant outside session", svc: &fakeCbtSessionService{hasParticipantSet: true, hasParticipant: false}, req: withRouteParams(adminRequest(http.MethodPost, "/", ""), pairs...), wantStatus: http.StatusForbidden},
		{name: "invalid optional json", svc: &fakeCbtSessionService{hasParticipantSet: true, hasParticipant: true}, req: withRouteParams(adminRequest(http.MethodPost, "/", `{`), pairs...), wantStatus: http.StatusBadRequest},
		{name: "service conflict", svc: &fakeCbtSessionService{hasParticipantSet: true, hasParticipant: true, proctorActionErr: errors.Join(domain.ErrConflict, errors.New("jawaban belum sinkron"))}, req: withRouteParams(adminRequest(http.MethodPost, "/", `{"reason":"manual"}`), pairs...), wantStatus: http.StatusConflict},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			(&CbtSession{svc: tt.svc}).ForceSubmitParticipant(rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestCbtSessionExecuteProctorActionEdges(t *testing.T) {
	sessionID := handlerTestUUID(207)
	roomID := handlerTestUUID(208)
	participantID := handlerTestUUID(209)
	eventID := handlerTestUUID(210)
	employeeID := handlerTestUUID(211)
	actorID := handlerTestUUID(212)
	pairs := []string{"id", sessionID.String(), "pid", participantID.String()}
	proctorReq := func(body string) *http.Request {
		return withRouteParams(withClaims(httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body)), jwt.MapClaims{
			"roles": []any{"guru"}, "eid": employeeID.String(), "uid": actorID.String(), "sub": actorID.String(), "usr": "pengawas.edge",
		}), pairs...)
	}

	fake := &fakeCbtSessionService{roomProctorAllowed: true, participantScope: db.GetCbtParticipantProctorScopeRow{ParticipantID: participantID, SessionID: sessionID, RoomID: roomID}}
	rec := httptest.NewRecorder()
	(&CbtSession{svc: fake}).ExecuteProctorAction(rec, proctorReq(`{"action_type":"mark_incident","event_id":"`+eventID.String()+`","reason":"Dicatat","notes":"Catatan"}`))
	if rec.Code != http.StatusOK {
		t.Fatalf("ExecuteProctorAction status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.proctorActionInput.EventID != eventID || fake.proctorActionInput.ActionType != "mark_incident" || fake.proctorActionActor.EmployeeID != employeeID || fake.roomProctorEmployeeID != employeeID {
		t.Fatalf("proctor action = input:%+v actor:%+v roomEmployee:%v, want scoped proctor action", fake.proctorActionInput, fake.proctorActionActor, fake.roomProctorEmployeeID)
	}

	cases := []struct {
		name       string
		svc        *fakeCbtSessionService
		req        *http.Request
		wantStatus int
	}{
		{name: "invalid participant id", svc: &fakeCbtSessionService{}, req: withRouteParams(adminRequest(http.MethodPost, "/", `{}`), "id", sessionID.String(), "pid", "bad"), wantStatus: http.StatusBadRequest},
		{name: "scope lookup error", svc: &fakeCbtSessionService{participantScopeErr: errors.Join(domain.ErrNotFound, errors.New("peserta tidak ditemukan"))}, req: withRouteParams(adminRequest(http.MethodPost, "/", `{}`), pairs...), wantStatus: http.StatusNotFound},
		{name: "proctor participant has no room", svc: &fakeCbtSessionService{participantScope: db.GetCbtParticipantProctorScopeRow{ParticipantID: participantID, SessionID: sessionID}}, req: proctorReq(`{"action_type":"unlock_access"}`), wantStatus: http.StatusForbidden},
		{name: "invalid json", svc: &fakeCbtSessionService{participantScope: db.GetCbtParticipantProctorScopeRow{ParticipantID: participantID, SessionID: sessionID, RoomID: roomID}, roomProctorAllowed: true}, req: proctorReq(`{`), wantStatus: http.StatusBadRequest},
		{name: "invalid event id", svc: &fakeCbtSessionService{participantScope: db.GetCbtParticipantProctorScopeRow{ParticipantID: participantID, SessionID: sessionID, RoomID: roomID}, roomProctorAllowed: true}, req: proctorReq(`{"action_type":"mark_incident","event_id":"bad"}`), wantStatus: http.StatusBadRequest},
		{name: "service bad request", svc: &fakeCbtSessionService{participantScope: db.GetCbtParticipantProctorScopeRow{ParticipantID: participantID, SessionID: sessionID, RoomID: roomID}, roomProctorAllowed: true, proctorActionErr: errors.Join(domain.ErrBadRequest, errors.New("action_type wajib diisi"))}, req: proctorReq(`{"action_type":""}`), wantStatus: http.StatusBadRequest},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			(&CbtSession{svc: tt.svc}).ExecuteProctorAction(rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestCbtSessionAcknowledgeProctorEventEdges(t *testing.T) {
	sessionID := handlerTestUUID(213)
	roomID := handlerTestUUID(214)
	eventID := handlerTestUUID(215)
	employeeID := handlerTestUUID(216)
	actorID := handlerTestUUID(217)
	proctorReq := func(body string, eventParam string) *http.Request {
		return withRouteParams(withClaims(httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body)), jwt.MapClaims{
			"roles": []any{"guru"}, "eid": employeeID.String(), "uid": actorID.String(), "sub": actorID.String(), "usr": "pengawas.ack",
		}), "event_id", eventParam)
	}

	fake := &fakeCbtSessionService{roomProctorAllowed: true, proctorEventScope: db.GetCbtProctorEventScopeRow{ID: eventID, SessionID: sessionID, RoomID: roomID, Severity: "critical"}}
	rec := httptest.NewRecorder()
	(&CbtSession{svc: fake}).AcknowledgeProctorEvent(rec, proctorReq(`{"note":"  Sudah dicek  ","notes":"ignored"}`, eventID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("AcknowledgeProctorEvent status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.proctorEventScopeID != eventID || fake.ackEventID != eventID || fake.ackNote != "Sudah dicek" || fake.ackActor.Username != "pengawas.ack" || fake.roomProctorEmployeeID != employeeID {
		t.Fatalf("ack args = scope:%v event:%v note:%q actor:%+v roomEmployee:%v", fake.proctorEventScopeID, fake.ackEventID, fake.ackNote, fake.ackActor, fake.roomProctorEmployeeID)
	}

	cases := []struct {
		name       string
		svc        *fakeCbtSessionService
		req        *http.Request
		wantStatus int
	}{
		{name: "invalid event id", svc: &fakeCbtSessionService{}, req: withRouteParams(adminRequest(http.MethodPost, "/", `{}`), "event_id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "scope not found", svc: &fakeCbtSessionService{proctorEventScopeErr: errors.Join(domain.ErrNotFound, errors.New("peringatan tidak ditemukan"))}, req: withRouteParams(adminRequest(http.MethodPost, "/", `{}`), "event_id", eventID.String()), wantStatus: http.StatusNotFound},
		{name: "invalid json", svc: &fakeCbtSessionService{proctorEventScope: db.GetCbtProctorEventScopeRow{ID: eventID, SessionID: sessionID}}, req: withRouteParams(adminRequest(http.MethodPost, "/", `{`), "event_id", eventID.String()), wantStatus: http.StatusBadRequest},
		{name: "teacher access denied for session-scoped event", svc: &fakeCbtSessionService{checkAllowed: false, proctorEventScope: db.GetCbtProctorEventScopeRow{ID: eventID, SessionID: sessionID}}, req: proctorReq(`{}`, eventID.String()), wantStatus: http.StatusForbidden},
		{name: "ack service conflict", svc: &fakeCbtSessionService{proctorEventScope: db.GetCbtProctorEventScopeRow{ID: eventID, SessionID: sessionID}, ackErr: errors.Join(domain.ErrConflict, errors.New("sudah diakui"))}, req: withRouteParams(adminRequest(http.MethodPost, "/", `{"notes":"late"}`), "event_id", eventID.String()), wantStatus: http.StatusConflict},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			(&CbtSession{svc: tt.svc}).AcknowledgeProctorEvent(rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestCbtSessionGetRoomProctoringDashboardEdges(t *testing.T) {
	sessionID := handlerTestUUID(218)
	roomID := handlerTestUUID(219)
	employeeID := handlerTestUUID(220)
	participantID := handlerTestUUID(221)
	pairs := []string{"id", sessionID.String(), "rid", roomID.String()}
	proctorReq := func() *http.Request {
		return withRouteParams(withClaims(httptest.NewRequest(http.MethodGet, "/", nil), jwt.MapClaims{"roles": []any{"guru"}, "eid": employeeID.String()}), pairs...)
	}

	fake := &cbtSessionMoreService{fakeCbtSessionService: &fakeCbtSessionService{roomProctorAllowed: true}, roomProctoringRows: []db.GetSessionProctoringStatusRow{{ParticipantID: participantID, Token: "secret-token", Nama: "Siswa"}}}
	rec := httptest.NewRecorder()
	(&CbtSession{svc: fake}).GetRoomProctoringDashboard(rec, proctorReq())
	if rec.Code != http.StatusOK {
		t.Fatalf("GetRoomProctoringDashboard status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.hasRoomSessionID != sessionID || fake.hasRoomID != roomID || fake.roomDashboardID != roomID || fake.listRoomProctorsRoomID != roomID || fake.roomProctoringSessionID != sessionID || fake.roomEventsLimit != 100 {
		t.Fatalf("dashboard calls = hasRoom:%v/%v dashboard:%v proctors:%v participants:%v/%v eventsLimit:%d", fake.hasRoomSessionID, fake.hasRoomID, fake.roomDashboardID, fake.listRoomProctorsRoomID, fake.roomProctoringSessionID, fake.roomProctoringRoomID, fake.roomEventsLimit)
	}
	if strings.Contains(rec.Body.String(), "secret-token") {
		t.Fatalf("dashboard body leaked token to non-admin proctor: %s", rec.Body.String())
	}

	boom := errors.New("boom")
	cases := []struct {
		name       string
		svc        *cbtSessionMoreService
		req        *http.Request
		wantStatus int
	}{
		{name: "invalid room id", svc: &cbtSessionMoreService{fakeCbtSessionService: &fakeCbtSessionService{}}, req: withRouteParams(adminRequest(http.MethodGet, "/", ""), "id", sessionID.String(), "rid", "bad"), wantStatus: http.StatusBadRequest},
		{name: "room outside session", svc: &cbtSessionMoreService{fakeCbtSessionService: &fakeCbtSessionService{hasRoomSet: true, hasRoom: false}}, req: withRouteParams(adminRequest(http.MethodGet, "/", ""), pairs...), wantStatus: http.StatusForbidden},
		{name: "proctor membership error", svc: &cbtSessionMoreService{fakeCbtSessionService: &fakeCbtSessionService{}, roomProctorAllowedErr: boom}, req: proctorReq(), wantStatus: http.StatusInternalServerError},
		{name: "dashboard query error", svc: &cbtSessionMoreService{fakeCbtSessionService: &fakeCbtSessionService{roomProctorAllowed: true}, roomDashboardErr: boom}, req: proctorReq(), wantStatus: http.StatusInternalServerError},
		{name: "proctors query error", svc: &cbtSessionMoreService{fakeCbtSessionService: &fakeCbtSessionService{roomProctorAllowed: true}, listRoomProctorsErr: boom}, req: proctorReq(), wantStatus: http.StatusInternalServerError},
		{name: "participants query error", svc: &cbtSessionMoreService{fakeCbtSessionService: &fakeCbtSessionService{roomProctorAllowed: true}, roomProctoringErr: boom}, req: proctorReq(), wantStatus: http.StatusInternalServerError},
		{name: "events query error", svc: &cbtSessionMoreService{fakeCbtSessionService: &fakeCbtSessionService{roomProctorAllowed: true}, roomEventsErr: boom}, req: proctorReq(), wantStatus: http.StatusInternalServerError},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			(&CbtSession{svc: tt.svc}).GetRoomProctoringDashboard(rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
