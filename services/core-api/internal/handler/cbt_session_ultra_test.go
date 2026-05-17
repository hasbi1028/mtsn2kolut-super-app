package handler

import (
	"errors"
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

func TestCbtSessionUltraRoomReadinessProctorsAndHandoverEdges(t *testing.T) {
	sessionID := handlerTestUUID(61)
	roomID := handlerTestUUID(62)
	actorID := handlerTestUUID(63)
	primaryID := handlerTestUUID(64)
	secondaryID := handlerTestUUID(65)
	employeeID := handlerTestUUID(66)
	roomPairs := []string{"id", sessionID.String(), "rid", roomID.String()}

	adminRoute := func(method, target, body string, pairs ...string) *http.Request {
		return withRouteParams(adminRequest(method, target, body), pairs...)
	}
	teacherRoute := func(method, target, body string, pairs ...string) *http.Request {
		req := withClaims(httptest.NewRequest(method, target, strings.NewReader(body)), jwt.MapClaims{"roles": []any{"guru"}, "eid": employeeID.String(), "uid": actorID.String(), "sub": actorID.String(), "usr": "guru.pengawas"})
		return withRouteParams(req, pairs...)
	}

	t.Run("GetRoomReadiness validates session and teacher access before service", func(t *testing.T) {
		h := &CbtSession{svc: &fakeCbtSessionService{checkAllowed: true}}
		rec := httptest.NewRecorder()
		h.GetRoomReadiness(rec, withRouteParam(adminRequest(http.MethodGet, "/api/cbt/sessions/bad/room-readiness", ""), "id", "bad"))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("invalid id status = %d, want 400; body=%s", rec.Code, rec.Body.String())
		}

		fake := &fakeCbtSessionService{checkAllowed: false}
		h = &CbtSession{svc: fake}
		rec = httptest.NewRecorder()
		h.GetRoomReadiness(rec, teacherRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/room-readiness", "", "id", sessionID.String()))
		if rec.Code != http.StatusForbidden || fake.roomReadinessSessionID.Valid {
			t.Fatalf("teacher denied status/roomReadinessID = %d/%v, want 403/no service call", rec.Code, fake.roomReadinessSessionID)
		}

		fake = &fakeCbtSessionService{checkAllowed: true}
		h = &CbtSession{svc: fake}
		rec = httptest.NewRecorder()
		h.GetRoomReadiness(rec, teacherRoute(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/room-readiness", "", "id", sessionID.String()))
		if rec.Code != http.StatusOK || fake.checkSessionID != sessionID || fake.checkTeacherID != employeeID || fake.roomReadinessSessionID != sessionID {
			t.Fatalf("readiness status/args = %d check=%v/%v readiness=%v; body=%s", rec.Code, fake.checkSessionID, fake.checkTeacherID, fake.roomReadinessSessionID, rec.Body.String())
		}
	})

	t.Run("ListRoomProctors validates room scope and forwards room id", func(t *testing.T) {
		fake := &fakeCbtSessionService{hasRoomSet: true, hasRoom: false}
		h := &CbtSession{svc: fake}
		rec := httptest.NewRecorder()
		h.ListRoomProctors(rec, adminRoute(http.MethodGet, "/api/cbt/sessions/rooms/proctors", "", roomPairs...))
		if rec.Code != http.StatusForbidden || fake.listRoomProctorsRoomID.Valid {
			t.Fatalf("room denied status/listRoomID = %d/%v, want 403/no list", rec.Code, fake.listRoomProctorsRoomID)
		}

		fake = &fakeCbtSessionService{}
		h = &CbtSession{svc: fake}
		rec = httptest.NewRecorder()
		h.ListRoomProctors(rec, adminRoute(http.MethodGet, "/api/cbt/sessions/rooms/proctors", "", roomPairs...))
		if rec.Code != http.StatusOK || fake.hasRoomSessionID != sessionID || fake.hasRoomID != roomID || fake.listRoomProctorsRoomID != roomID {
			t.Fatalf("list proctors status/args = %d has=%v/%v list=%v; body=%s", rec.Code, fake.hasRoomSessionID, fake.hasRoomID, fake.listRoomProctorsRoomID, rec.Body.String())
		}
	})

	t.Run("ReplaceRoomProctors validates auth payload and audits forwarded ids", func(t *testing.T) {
		h := &CbtSession{svc: &fakeCbtSessionService{}}
		rec := httptest.NewRecorder()
		h.ReplaceRoomProctors(rec, withRouteParams(httptest.NewRequest(http.MethodPut, "/api/cbt/sessions/rooms/proctors", strings.NewReader(`{}`)), roomPairs...))
		if rec.Code != http.StatusForbidden {
			t.Fatalf("unauthenticated replace status = %d, want 403; body=%s", rec.Code, rec.Body.String())
		}

		rec = httptest.NewRecorder()
		h.ReplaceRoomProctors(rec, adminRoute(http.MethodPut, "/api/cbt/sessions/rooms/proctors", `{"primary_employee_id":"bad"}`, roomPairs...))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("bad primary status = %d, want 400; body=%s", rec.Code, rec.Body.String())
		}

		fake := &fakeCbtSessionService{}
		audit := &fakeCbtSessionAuditWriter{}
		h = &CbtSession{svc: fake, audit: audit}
		body := `{"primary_employee_id":"` + primaryID.String() + `","employee_ids":["","` + primaryID.String() + `","` + secondaryID.String() + `"]}`
		rec = httptest.NewRecorder()
		h.ReplaceRoomProctors(rec, adminRoute(http.MethodPut, "/api/cbt/sessions/rooms/proctors", body, roomPairs...))
		if rec.Code != http.StatusOK || fake.replaceProctorsRoomID != roomID || fake.replaceProctorsPrimaryID != primaryID || len(fake.replaceProctorsEmployeeIDs) != 2 || fake.replaceProctorsEmployeeIDs[1] != secondaryID {
			t.Fatalf("replace status/args = %d room=%v primary=%v employees=%v; body=%s", rec.Code, fake.replaceProctorsRoomID, fake.replaceProctorsPrimaryID, fake.replaceProctorsEmployeeIDs, rec.Body.String())
		}
		if len(audit.entries) != 1 || audit.entries[0].Action != "CBT_SESSION_ROOM_PROCTORS_REPLACE" || audit.entries[0].EntityID != pgUUIDString(roomID) {
			t.Fatalf("replace audit entries = %+v, want one replace audit for room", audit.entries)
		}
	})

	t.Run("GetRoomHandover validates assigned proctor and returns row", func(t *testing.T) {
		fake := &fakeCbtSessionService{roomProctorAllowed: false}
		h := &CbtSession{svc: fake}
		rec := httptest.NewRecorder()
		h.GetRoomHandover(rec, teacherRoute(http.MethodGet, "/api/cbt/sessions/rooms/handover", "", roomPairs...))
		if rec.Code != http.StatusForbidden || fake.handoverRoomID.Valid {
			t.Fatalf("unassigned proctor status/handover = %d/%v, want 403/no handover", rec.Code, fake.handoverRoomID)
		}

		fake = &fakeCbtSessionService{roomProctorAllowed: true}
		h = &CbtSession{svc: fake}
		rec = httptest.NewRecorder()
		h.GetRoomHandover(rec, teacherRoute(http.MethodGet, "/api/cbt/sessions/rooms/handover", "", roomPairs...))
		if rec.Code != http.StatusOK || fake.hasRoomSessionID != sessionID || fake.roomProctorEmployeeID != employeeID || fake.handoverRoomID != roomID {
			t.Fatalf("handover status/args = %d hasRoom=%v proctor=%v handover=%v; body=%s", rec.Code, fake.hasRoomSessionID, fake.roomProctorEmployeeID, fake.handoverRoomID, rec.Body.String())
		}
	})
}

func TestCbtSessionUltraLiveFinalizeAndItemAnalysisEdges(t *testing.T) {
	sessionID := handlerTestUUID(71)
	teacherID := handlerTestUUID(72)
	errDB := errors.New("db down")

	teacherReq := func(target string) *http.Request {
		return withRouteParam(withClaims(httptest.NewRequest(http.MethodGet, target, nil), jwt.MapClaims{"roles": []any{"guru"}, "eid": teacherID.String(), "usr": "guru.ipa"}), "id", sessionID.String())
	}

	tests := []struct {
		name       string
		fn         func(*CbtSession, http.ResponseWriter, *http.Request)
		svc        *fakeCbtSessionService
		req        *http.Request
		wantStatus int
	}{
		{name: "live invalid session id", fn: (*CbtSession).GetProctoringLiveSummary, svc: &fakeCbtSessionService{}, req: withRouteParam(adminRequest(http.MethodGet, "/api/cbt/sessions/bad/proctoring/live", ""), "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "live teacher access error", fn: (*CbtSession).GetProctoringLiveSummary, svc: &fakeCbtSessionService{checkErr: errDB}, req: teacherReq("/api/cbt/sessions/" + sessionID.String() + "/proctoring/live"), wantStatus: http.StatusInternalServerError},
		{name: "live service error", fn: (*CbtSession).GetProctoringLiveSummary, svc: &fakeCbtSessionService{checkAllowed: true, liveSummaryErr: errDB}, req: teacherReq("/api/cbt/sessions/" + sessionID.String() + "/proctoring/live"), wantStatus: http.StatusInternalServerError},
		{name: "finalize forbidden for teacher", fn: (*CbtSession).FinalizeOverdue, svc: &fakeCbtSessionService{checkAllowed: true}, req: teacherReq("/api/cbt/sessions/" + sessionID.String() + "/finalize-overdue"), wantStatus: http.StatusForbidden},
		{name: "finalize conflict maps", fn: (*CbtSession).FinalizeOverdue, svc: &fakeCbtSessionService{finalizeOverdueErr: errors.Join(domain.ErrConflict, errors.New("sesi masih berjalan"))}, req: withRouteParam(adminRequest(http.MethodPost, "/api/cbt/sessions/"+sessionID.String()+"/finalize-overdue", ""), "id", sessionID.String()), wantStatus: http.StatusConflict},
		{name: "item analysis invalid id", fn: (*CbtSession).GetItemAnalysis, svc: &fakeCbtSessionService{}, req: withRouteParam(adminRequest(http.MethodGet, "/api/cbt/sessions/bad/item-analysis", ""), "id", "bad"), wantStatus: http.StatusBadRequest},
		{name: "item analysis service error", fn: (*CbtSession).GetItemAnalysis, svc: &fakeCbtSessionService{itemAnalysisErr: errDB}, req: withRouteParam(adminRequest(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/item-analysis", ""), "id", sessionID.String()), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.fn(&CbtSession{svc: tt.svc}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}

	fake := &fakeCbtSessionService{checkAllowed: true, liveSummary: serviceSummaryForUltraTest(sessionID)}
	rec := httptest.NewRecorder()
	(&CbtSession{svc: fake}).GetProctoringLiveSummary(rec, teacherReq("/api/cbt/sessions/"+sessionID.String()+"/proctoring/live"))
	if rec.Code != http.StatusOK || fake.liveSummarySessionID != sessionID || fake.liveSummaryRoomID.Valid || !strings.Contains(rec.Body.String(), `"participants":7`) {
		t.Fatalf("live success status/args/body = %d/%v/%v/%s, want summary for session", rec.Code, fake.liveSummarySessionID, fake.liveSummaryRoomID, rec.Body.String())
	}

	fake = &fakeCbtSessionService{itemAnalysisRows: []db.GetSessionItemAnalysisRow{
		{QuestionID: handlerTestUUID(73), QuestionType: "essay", SubmittedCount: 4, AnsweredCount: 4, UnscoredCount: 1, DifficultyIndex: 0.5, AnswerDistribution: []byte(`not-json`)},
		{QuestionID: handlerTestUUID(74), QuestionType: "multiple_choice", SubmittedCount: 4, AnsweredCount: 4, DifficultyIndex: 0.1, DiscriminationIndex: -0.1, AnswerKey: "B", AnswerDistribution: []byte(`{"B":2}`)},
	}}
	rec = httptest.NewRecorder()
	(&CbtSession{svc: fake}).GetItemAnalysis(rec, withRouteParam(adminRequest(http.MethodGet, "/api/cbt/sessions/"+sessionID.String()+"/item-analysis", ""), "id", sessionID.String()))
	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(body, "Koreksi uraian belum lengkap") || !strings.Contains(body, "Cek kunci/rubrik") || !strings.Contains(body, `"answer_distribution":[]`) || !strings.Contains(body, `"answer_key":"B"`) {
		t.Fatalf("item analysis edge response status/body = %d/%s, want recommendations, bad-json empty distribution and admin key", rec.Code, body)
	}
}

func serviceSummaryForUltraTest(sessionID pgtype.UUID) service.CbtProctoringLiveSummary {
	return service.CbtProctoringLiveSummary{SessionID: pgUUIDString(sessionID), Counts: map[string]int{"participants": 7}}
}
