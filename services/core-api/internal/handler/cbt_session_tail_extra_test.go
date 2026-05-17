package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"mtsn2kolut-super-app/backend/internal/service"
)

func TestCbtSessionTailParticipantEventsAndRecapHandlers(t *testing.T) {
	sessionID := handlerTestUUID(10)
	participantID := handlerTestUUID(11)
	teacherID := handlerTestUUID(12)
	fake := &fakeCbtSessionService{checkAllowed: true}
	h := &CbtSession{svc: fake}

	teacherRoute := func(method, target string, pairs ...string) *http.Request {
		return withRouteParams(withClaims(httptest.NewRequest(method, target, nil), jwt.MapClaims{
			"roles": []any{"guru"},
			"eid":   teacherID.String(),
			"usr":   "guru.tail",
		}), pairs...)
	}

	rec := httptest.NewRecorder()
	h.ListParticipantEvents(rec, teacherRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/proctoring/events?participant_id="+participantID.String()+"&limit=17", "id", sessionID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("ListParticipantEvents status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.checkSessionID != sessionID || fake.checkTeacherID != teacherID || fake.hasParticipantID != participantID || fake.roomEventsSessionID != sessionID || fake.roomEventsParticipantID != participantID || fake.roomEventsLimit != 17 {
		t.Fatalf("ListParticipantEvents args = check:%v/%v participant:%v events:%v/%v/%d, want scoped teacher participant events", fake.checkSessionID, fake.checkTeacherID, fake.hasParticipantID, fake.roomEventsSessionID, fake.roomEventsParticipantID, fake.roomEventsLimit)
	}

	rec = httptest.NewRecorder()
	h.ListParticipantEvents(rec, teacherRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/proctoring/events?limit=not-number", "id", sessionID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("ListParticipantEvents default limit status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.roomEventsParticipantID.Valid || fake.roomEventsLimit != 100 {
		t.Fatalf("ListParticipantEvents default args = participant:%v limit:%d, want no participant and default 100", fake.roomEventsParticipantID, fake.roomEventsLimit)
	}

	rec = httptest.NewRecorder()
	h.GetSessionOperationalRecap(rec, teacherRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/operational-recap", "id", sessionID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("GetSessionOperationalRecap status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.operationalRecapID != sessionID || !strings.Contains(rec.Body.String(), `"recap"`) || !strings.Contains(rec.Body.String(), `"rooms"`) || !strings.Contains(rec.Body.String(), "Ruang 1") {
		t.Fatalf("GetSessionOperationalRecap args/body = %v/%s, want recap and rooms", fake.operationalRecapID, rec.Body.String())
	}
}

func TestCbtSessionTailProctoringReportHandlers(t *testing.T) {
	sessionID := handlerTestUUID(20)
	roomID := handlerTestUUID(21)
	employeeID := handlerTestUUID(22)
	fake := &fakeCbtSessionService{
		checkAllowed:       true,
		roomProctorAllowed: true,
		liveSummary: service.CbtProctoringLiveSummary{
			SessionID: pgUUIDString(sessionID),
			RoomID:    pgUUIDString(roomID),
			LatestEvents: []service.CbtProctoringEventDTO{
				{ID: "technical-1", Severity: "technical"},
				{ID: "warning-1", Severity: "warning", IsMassTechnicalIssue: true},
			},
			Participants: []service.CbtProctoringParticipantDTO{
				{ParticipantID: "p-submitted", ConnectionStatus: "selesai"},
				{ParticipantID: "p-locked", RiskLevel: "locked", SyncStatus: "tertahan"},
			},
		},
	}
	h := &CbtSession{svc: fake}

	sessionReq := withRouteParams(adminRequest(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/proctoring/report-data", ""), "id", sessionID.String())
	rec := httptest.NewRecorder()
	h.GetProctoringReportData(rec, sessionReq)
	if rec.Code != http.StatusOK {
		t.Fatalf("GetProctoringReportData status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.liveSummarySessionID != sessionID || fake.liveSummaryRoomID.Valid || !strings.Contains(rec.Body.String(), `"technical_incidents"`) || !strings.Contains(rec.Body.String(), `"sync_anomalies"`) || !strings.Contains(rec.Body.String(), `"submitted"`) {
		t.Fatalf("GetProctoringReportData args/body = %v/%v/%s, want session report data", fake.liveSummarySessionID, fake.liveSummaryRoomID, rec.Body.String())
	}

	roomReq := withRouteParams(withClaims(httptest.NewRequest(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/rooms/"+roomID.String()+"/proctoring/report-data", nil), jwt.MapClaims{
		"roles": []any{"guru"},
		"eid":   employeeID.String(),
		"usr":   "pengawas.tail",
	}), "id", sessionID.String(), "rid", roomID.String())
	rec = httptest.NewRecorder()
	h.GetRoomProctoringReportData(rec, roomReq)
	if rec.Code != http.StatusOK {
		t.Fatalf("GetRoomProctoringReportData status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.hasRoomID != roomID || fake.roomProctorEmployeeID != employeeID || fake.liveSummaryRoomID != roomID {
		t.Fatalf("GetRoomProctoringReportData args = room:%v proctor:%v live:%v, want room scoped report", fake.hasRoomID, fake.roomProctorEmployeeID, fake.liveSummaryRoomID)
	}

	fake.liveSummaryErr = errors.New("summary unavailable")
	rec = httptest.NewRecorder()
	h.GetProctoringReportData(rec, sessionReq)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("GetProctoringReportData service error status = %d, want 500; body=%s", rec.Code, rec.Body.String())
	}
}

func TestCbtSessionTailRoomPrintPackAndMyRoomsHandlers(t *testing.T) {
	sessionID := handlerTestUUID(30)
	roomID := handlerTestUUID(31)
	employeeID := handlerTestUUID(32)
	fake := &fakeCbtSessionService{roomProctorAllowed: true}
	h := &CbtSession{svc: fake}

	proctorRoute := func(method, target string, pairs ...string) *http.Request {
		return withRouteParams(withClaims(httptest.NewRequest(method, target, nil), jwt.MapClaims{
			"roles": []any{"guru"},
			"eid":   employeeID.String(),
			"usr":   "pengawas.tail",
		}), pairs...)
	}

	rec := httptest.NewRecorder()
	h.GetRoomProctorPrintPack(rec, proctorRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/rooms/"+roomID.String()+"/print-pack", "id", sessionID.String(), "rid", roomID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("GetRoomProctorPrintPack status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.hasRoomID != roomID || fake.roomProctorEmployeeID != employeeID || fake.roomDashboardID != roomID || fake.listRoomProctorsRoomID != roomID || fake.roomProctoringRoomID != roomID {
		t.Fatalf("GetRoomProctorPrintPack args = has:%v proctor:%v dashboard:%v list:%v participants:%v, want room scoped print pack", fake.hasRoomID, fake.roomProctorEmployeeID, fake.roomDashboardID, fake.listRoomProctorsRoomID, fake.roomProctoringRoomID)
	}
	if !strings.Contains(rec.Body.String(), `"room"`) || !strings.Contains(rec.Body.String(), `"proctors"`) || !strings.Contains(rec.Body.String(), `"participants"`) {
		t.Fatalf("GetRoomProctorPrintPack body = %s, want print pack sections", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.ListMyProctorRooms(rec, proctorRoute(http.MethodGet, "/api/cbt/proctoring/my-rooms"))
	if rec.Code != http.StatusOK {
		t.Fatalf("ListMyProctorRooms status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.proctorRoomsEmployeeID != employeeID || fake.proctorRoomsIncludeAll {
		t.Fatalf("ListMyProctorRooms args = employee:%v includeAll:%v, want non-admin employee scoped rooms", fake.proctorRoomsEmployeeID, fake.proctorRoomsIncludeAll)
	}
	if !strings.Contains(rec.Body.String(), "Ruang 1") {
		t.Fatalf("ListMyProctorRooms body = %s, want room rows", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.ListMyProctorRooms(rec, withClaims(httptest.NewRequest(http.MethodGet, "/api/cbt/proctoring/my-rooms", nil), jwt.MapClaims{"roles": []any{"guru"}}))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("ListMyProctorRooms missing employee status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}
}

func TestCbtSessionTailHandlersValidateBadRouteParams(t *testing.T) {
	sessionID := handlerTestUUID(40)
	roomID := handlerTestUUID(41)
	h := &CbtSession{svc: &fakeCbtSessionService{roomProctorAllowed: true}}

	tests := []struct {
		name    string
		handler http.HandlerFunc
		req     *http.Request
	}{
		{name: "events bad session", handler: h.ListParticipantEvents, req: withRouteParams(adminRequest(http.MethodGet, "/", ""), "id", "bad")},
		{name: "events bad participant", handler: h.ListParticipantEvents, req: withRouteParams(adminRequest(http.MethodGet, "/?participant_id=bad", ""), "id", sessionID.String())},
		{name: "session report bad session", handler: h.GetProctoringReportData, req: withRouteParams(adminRequest(http.MethodGet, "/", ""), "id", "bad")},
		{name: "room report bad room", handler: h.GetRoomProctoringReportData, req: withRouteParams(adminRequest(http.MethodGet, "/", ""), "id", sessionID.String(), "rid", "bad")},
		{name: "print pack bad room", handler: h.GetRoomProctorPrintPack, req: withRouteParams(adminRequest(http.MethodGet, "/", ""), "id", sessionID.String(), "rid", "bad")},
		{name: "operational recap bad session", handler: h.GetSessionOperationalRecap, req: withRouteParams(adminRequest(http.MethodGet, "/", ""), "id", "bad")},
		{name: "room report forbidden room", handler: h.GetRoomProctoringReportData, req: withRouteParams(adminRequest(http.MethodGet, "/", ""), "id", sessionID.String(), "rid", roomID.String())},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if strings.Contains(tt.name, "forbidden room") {
				h.svc = &fakeCbtSessionService{hasRoomSet: true, hasRoom: false, roomProctorAllowed: true}
			} else {
				h.svc = &fakeCbtSessionService{roomProctorAllowed: true}
			}
			rec := httptest.NewRecorder()
			tt.handler(rec, tt.req)
			want := http.StatusBadRequest
			if strings.Contains(tt.name, "forbidden room") {
				want = http.StatusForbidden
			}
			if rec.Code != want {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, want, rec.Body.String())
			}
		})
	}
}
