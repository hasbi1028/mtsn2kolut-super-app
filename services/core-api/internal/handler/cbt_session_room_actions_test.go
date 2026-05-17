package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestCbtSessionRoomParticipantActionEndpointsForwardAssignedProctorRequests(t *testing.T) {
	sessionID := handlerTestUUID(240)
	roomID := handlerTestUUID(241)
	participantID := handlerTestUUID(242)
	eventID := handlerTestUUID(243)
	actorID := handlerTestUUID(244)
	employeeID := handlerTestUUID(245)
	fake := &fakeCbtSessionService{
		roomProctorAllowed:     true,
		roomParticipantAllowed: true,
		proctorEventScope: db.GetCbtProctorEventScopeRow{
			ID:            eventID,
			SessionID:     sessionID,
			RoomID:        roomID,
			ParticipantID: participantID,
			Severity:      "warning",
			Nama:          "Ahmad",
			RoomName:      "Ruang 1",
		},
	}
	audit := &fakeCbtSessionAuditWriter{}
	h := &CbtSession{svc: fake, audit: audit}

	proctorRoute := func(method, target, body string, pairs ...string) *http.Request {
		req := withClaims(httptest.NewRequest(method, target, strings.NewReader(body)), jwt.MapClaims{
			"roles": []any{"guru"},
			"uid":   actorID.String(),
			"sub":   actorID.String(),
			"eid":   employeeID.String(),
			"usr":   "pengawas.ruang",
			"ssid":  "session-proctor-1",
		})
		req.Header.Set("X-Request-ID", "req-room-actions")
		return withRouteParams(req, pairs...)
	}
	roomParticipantPairs := []string{"id", sessionID.String(), "rid", roomID.String(), "pid", participantID.String()}
	run := func(name string, fn http.HandlerFunc, body string) {
		t.Helper()
		rec := httptest.NewRecorder()
		fn(rec, proctorRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/rooms/"+roomID.String()+"/participants/"+participantID.String(), body, roomParticipantPairs...))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want 200; body=%s", name, rec.Code, rec.Body.String())
		}
	}

	run("FlagRoomParticipant", h.FlagRoomParticipant, `{"flag":true}`)
	if fake.hasRoomSessionID != sessionID || fake.hasRoomID != roomID || fake.roomProctorEmployeeID != employeeID || fake.roomParticipantID != participantID || fake.flagParticipantID != participantID || !fake.flagValue {
		t.Fatalf("flag room participant args = hasRoom:%v/%v proctor:%v participant:%v flag:%v/%v, want scoped room participant flag", fake.hasRoomSessionID, fake.hasRoomID, fake.roomProctorEmployeeID, fake.roomParticipantID, fake.flagParticipantID, fake.flagValue)
	}

	run("ResetRoomParticipantAccess", h.ResetRoomParticipantAccess, `{"reason":"Ganti perangkat","notes":"Sinkron aman"}`)
	if fake.proctorActionInput.ActionType != "reset_device_binding" || fake.proctorActionInput.Reason != "Ganti perangkat" || fake.proctorActionInput.Notes != "Sinkron aman" || fake.proctorActionActor.UserID != actorID || fake.proctorActionActor.EmployeeID != employeeID {
		t.Fatalf("reset room action = input:%+v actor:%+v, want reset action by assigned proctor", fake.proctorActionInput, fake.proctorActionActor)
	}

	run("UnlockRoomParticipant", h.UnlockRoomParticipant, `{"reason":"Verifikasi wajah","notes":"Buka kunci"}`)
	if fake.proctorActionInput.ActionType != "unlock_access" || fake.proctorActionInput.Reason != "Verifikasi wajah" || fake.proctorActionInput.Notes != "Buka kunci" {
		t.Fatalf("unlock room action = %+v, want unlock_access with payload reason/notes", fake.proctorActionInput)
	}

	run("AcknowledgeRoomParticipantEvent", h.AcknowledgeRoomParticipantEvent, `{"event_id":"`+eventID.String()+`","notes":"Sudah dicek"}`)
	if fake.proctorEventScopeID != eventID || fake.ackEventID != eventID || fake.ackNote != "Sudah dicek" || fake.ackActor.Username != "pengawas.ruang" {
		t.Fatalf("ack room event args = scope:%v ack:%v note:%q actor:%+v, want participant event acknowledged", fake.proctorEventScopeID, fake.ackEventID, fake.ackNote, fake.ackActor)
	}

	run("RecordRoomParticipantIncidentAction", h.RecordRoomParticipantIncidentAction, `{"event_id":"`+eventID.String()+`","action":"Dicatat di BA","notes":"Perangkat restart"}`)
	if fake.proctorActionInput.ActionType != "mark_incident" || fake.proctorActionInput.EventID != eventID || fake.proctorActionInput.Reason != "Dicatat di BA" || fake.proctorActionInput.Notes != "Perangkat restart" {
		t.Fatalf("incident room action = %+v, want mark_incident event/action/notes", fake.proctorActionInput)
	}

	run("SendRoomParticipantCommand", h.SendRoomParticipantCommand, `{"command_type":"show_message","message":"Tetap di halaman ujian"}`)
	if fake.commandParticipantID != participantID || fake.commandType != "show_message" || fake.commandMessage != "Tetap di halaman ujian" || fake.commandActor != "pengawas.ruang" {
		t.Fatalf("command args = participant:%v type:%q message:%q actor:%q, want queued command by actor", fake.commandParticipantID, fake.commandType, fake.commandMessage, fake.commandActor)
	}

	rec := httptest.NewRecorder()
	h.RecordRoomProctorHeartbeat(rec, proctorRoute(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/rooms/"+roomID.String()+"/heartbeat", "", "id", sessionID.String(), "rid", roomID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("RecordRoomProctorHeartbeat status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"online"`) {
		t.Fatalf("heartbeat body = %s, want online status", rec.Body.String())
	}

	run("ForceSubmitRoomParticipant", h.ForceSubmitRoomParticipant, `{"reason":"Selesai manual","notes":"Jawaban sudah dicek"}`)
	if fake.proctorActionInput.ActionType != "force_submit" || fake.proctorActionInput.Reason != "Selesai manual" || fake.proctorActionInput.Notes != "Jawaban sudah dicek" {
		t.Fatalf("force submit room action = %+v, want force_submit payload", fake.proctorActionInput)
	}

	wantActions := []string{
		"CBT_SESSION_ROOM_PARTICIPANT_FLAG",
		"CBT_SESSION_ROOM_PARTICIPANT_RESET_ACCESS",
		"CBT_SESSION_ROOM_PARTICIPANT_UNLOCK",
		"CBT_SESSION_ROOM_PARTICIPANT_ACKNOWLEDGE",
		"CBT_SESSION_ROOM_PARTICIPANT_INCIDENT_ACTION",
		"CBT_SESSION_ROOM_PARTICIPANT_COMMAND",
		"CBT_PROCTOR_HEARTBEAT",
		"CBT_SESSION_ROOM_PARTICIPANT_FORCE_SUBMIT",
	}
	if len(audit.entries) != len(wantActions) {
		t.Fatalf("audit entries len = %d, want %d: %+v", len(audit.entries), len(wantActions), audit.entries)
	}
	for i, want := range wantActions {
		if audit.entries[i].Action != want || audit.entries[i].EntityID != pgUUIDString(roomID) {
			t.Fatalf("audit[%d] = %+v, want action %s on room", i, audit.entries[i], want)
		}
	}
	meta := mustAuditMetadataMap(t, audit.entries[0].Metadata)
	if meta["session_id"] != pgUUIDString(sessionID) || meta["participant_id"] != pgUUIDString(participantID) || meta["actor_username"] != "pengawas.ruang" {
		t.Fatalf("room participant audit metadata = %+v, want session/participant/actor", meta)
	}
}

func TestCbtSessionRoomParticipantActionEndpointsValidateScopeAndPayloads(t *testing.T) {
	sessionID := handlerTestUUID(246)
	roomID := handlerTestUUID(247)
	participantID := handlerTestUUID(248)
	eventID := handlerTestUUID(249)
	otherParticipantID := handlerTestUUID(250)
	employeeID := handlerTestUUID(251)
	errDB := errors.New("db down")

	guruRoute := func(method, target, body string, pairs ...string) *http.Request {
		actorID := handlerTestUUID(252)
		return withRouteParams(
			withClaims(httptest.NewRequest(method, target, strings.NewReader(body)), jwt.MapClaims{"roles": []any{"guru"}, "eid": employeeID.String(), "uid": actorID.String(), "sub": actorID.String(), "usr": "pengawas.ruang"}),
			pairs...,
		)
	}
	adminRoute := func(method, target, body string, pairs ...string) *http.Request {
		return withRouteParams(adminRequest(method, target, body), pairs...)
	}
	roomParticipantPairs := []string{"id", sessionID.String(), "rid", roomID.String(), "pid", participantID.String()}

	tests := []struct {
		name       string
		handler    func(*CbtSession, http.ResponseWriter, *http.Request)
		svc        *fakeCbtSessionService
		req        *http.Request
		wantStatus int
	}{
		{name: "flag invalid room id", handler: (*CbtSession).FlagRoomParticipant, req: adminRoute(http.MethodPost, "/", `{"flag":true}`, "id", sessionID.String(), "rid", "bad", "pid", participantID.String()), wantStatus: http.StatusBadRequest},
		{name: "flag invalid participant id", handler: (*CbtSession).FlagRoomParticipant, req: adminRoute(http.MethodPost, "/", `{"flag":true}`, "id", sessionID.String(), "rid", roomID.String(), "pid", "bad"), wantStatus: http.StatusBadRequest},
		{name: "flag room outside session", handler: (*CbtSession).FlagRoomParticipant, svc: &fakeCbtSessionService{hasRoomSet: true, hasRoom: false, roomParticipantAllowed: true}, req: adminRoute(http.MethodPost, "/", `{"flag":true}`, roomParticipantPairs...), wantStatus: http.StatusForbidden},
		{name: "flag room membership error", handler: (*CbtSession).FlagRoomParticipant, svc: &fakeCbtSessionService{hasRoomErr: errDB, roomParticipantAllowed: true}, req: adminRoute(http.MethodPost, "/", `{"flag":true}`, roomParticipantPairs...), wantStatus: http.StatusInternalServerError},
		{name: "flag unassigned proctor", handler: (*CbtSession).FlagRoomParticipant, svc: &fakeCbtSessionService{roomProctorAllowed: false, roomParticipantAllowed: true}, req: guruRoute(http.MethodPost, "/", `{"flag":true}`, roomParticipantPairs...), wantStatus: http.StatusForbidden},
		{name: "flag participant outside room", handler: (*CbtSession).FlagRoomParticipant, svc: &fakeCbtSessionService{roomParticipantAllowed: false}, req: adminRoute(http.MethodPost, "/", `{"flag":true}`, roomParticipantPairs...), wantStatus: http.StatusForbidden},
		{name: "flag participant room check error", handler: (*CbtSession).FlagRoomParticipant, svc: &fakeCbtSessionService{roomParticipantErr: errDB}, req: adminRoute(http.MethodPost, "/", `{"flag":true}`, roomParticipantPairs...), wantStatus: http.StatusInternalServerError},
		{name: "flag invalid json", handler: (*CbtSession).FlagRoomParticipant, svc: &fakeCbtSessionService{roomParticipantAllowed: true}, req: adminRoute(http.MethodPost, "/", `{`, roomParticipantPairs...), wantStatus: http.StatusBadRequest},
		{name: "reset invalid json", handler: (*CbtSession).ResetRoomParticipantAccess, svc: &fakeCbtSessionService{roomParticipantAllowed: true}, req: adminRoute(http.MethodPost, "/", `{`, roomParticipantPairs...), wantStatus: http.StatusBadRequest},
		{name: "reset service conflict", handler: (*CbtSession).ResetRoomParticipantAccess, svc: &fakeCbtSessionService{roomParticipantAllowed: true, proctorActionErr: errors.Join(domain.ErrConflict, errors.New("peserta sudah selesai"))}, req: adminRoute(http.MethodPost, "/", `{"reason":"Reset"}`, roomParticipantPairs...), wantStatus: http.StatusConflict},
		{name: "ack invalid json", handler: (*CbtSession).AcknowledgeRoomParticipantEvent, svc: &fakeCbtSessionService{roomParticipantAllowed: true}, req: adminRoute(http.MethodPost, "/", `{`, roomParticipantPairs...), wantStatus: http.StatusBadRequest},
		{name: "ack invalid event id", handler: (*CbtSession).AcknowledgeRoomParticipantEvent, svc: &fakeCbtSessionService{roomParticipantAllowed: true}, req: adminRoute(http.MethodPost, "/", `{"event_id":"bad"}`, roomParticipantPairs...), wantStatus: http.StatusBadRequest},
		{name: "ack event for other participant", handler: (*CbtSession).AcknowledgeRoomParticipantEvent, svc: &fakeCbtSessionService{roomParticipantAllowed: true, proctorEventScope: db.GetCbtProctorEventScopeRow{ID: eventID, SessionID: sessionID, RoomID: roomID, ParticipantID: otherParticipantID}}, req: adminRoute(http.MethodPost, "/", `{"event_id":"`+eventID.String()+`"}`, roomParticipantPairs...), wantStatus: http.StatusForbidden},
		{name: "ack service error", handler: (*CbtSession).AcknowledgeRoomParticipantEvent, svc: &fakeCbtSessionService{roomParticipantAllowed: true, proctorEventScope: db.GetCbtProctorEventScopeRow{ID: eventID, SessionID: sessionID, RoomID: roomID, ParticipantID: participantID}, ackErr: errDB}, req: adminRoute(http.MethodPost, "/", `{"event_id":"`+eventID.String()+`"}`, roomParticipantPairs...), wantStatus: http.StatusBadRequest},
		{name: "incident invalid json", handler: (*CbtSession).RecordRoomParticipantIncidentAction, svc: &fakeCbtSessionService{roomParticipantAllowed: true}, req: adminRoute(http.MethodPost, "/", `{`, roomParticipantPairs...), wantStatus: http.StatusBadRequest},
		{name: "incident invalid event id", handler: (*CbtSession).RecordRoomParticipantIncidentAction, svc: &fakeCbtSessionService{roomParticipantAllowed: true}, req: adminRoute(http.MethodPost, "/", `{"event_id":"bad"}`, roomParticipantPairs...), wantStatus: http.StatusBadRequest},
		{name: "incident service error", handler: (*CbtSession).RecordRoomParticipantIncidentAction, svc: &fakeCbtSessionService{roomParticipantAllowed: true, proctorActionErr: errDB}, req: adminRoute(http.MethodPost, "/", `{"action":"catat"}`, roomParticipantPairs...), wantStatus: http.StatusBadRequest},
		{name: "command invalid json", handler: (*CbtSession).SendRoomParticipantCommand, svc: &fakeCbtSessionService{roomParticipantAllowed: true}, req: adminRoute(http.MethodPost, "/", `{`, roomParticipantPairs...), wantStatus: http.StatusBadRequest},
		{name: "command service bad request", handler: (*CbtSession).SendRoomParticipantCommand, svc: &fakeCbtSessionService{roomParticipantAllowed: true, commandErr: errors.Join(domain.ErrBadRequest, errors.New("command_type wajib diisi"))}, req: adminRoute(http.MethodPost, "/", `{"command_type":"","message":""}`, roomParticipantPairs...), wantStatus: http.StatusBadRequest},
		{name: "heartbeat unassigned proctor", handler: (*CbtSession).RecordRoomProctorHeartbeat, svc: &fakeCbtSessionService{roomProctorAllowed: false}, req: guruRoute(http.MethodPost, "/", "", "id", sessionID.String(), "rid", roomID.String()), wantStatus: http.StatusForbidden},
		{name: "force submit invalid json", handler: (*CbtSession).ForceSubmitRoomParticipant, svc: &fakeCbtSessionService{roomParticipantAllowed: true}, req: adminRoute(http.MethodPost, "/", `{`, roomParticipantPairs...), wantStatus: http.StatusBadRequest},
		{name: "force submit service conflict", handler: (*CbtSession).ForceSubmitRoomParticipant, svc: &fakeCbtSessionService{roomParticipantAllowed: true, proctorActionErr: errors.Join(domain.ErrConflict, errors.New("jawaban belum sinkron"))}, req: adminRoute(http.MethodPost, "/", `{"reason":"manual"}`, roomParticipantPairs...), wantStatus: http.StatusConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.svc
			if svc == nil {
				svc = &fakeCbtSessionService{roomParticipantAllowed: true, roomProctorAllowed: true}
			}
			rec := httptest.NewRecorder()
			tt.handler(&CbtSession{svc: svc}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
