package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func cbtAccessTestRequest(claims jwt.MapClaims) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	if claims != nil {
		req = req.WithContext(context.WithValue(req.Context(), api.ClaimsKey, claims))
	}
	return req
}

func TestCbtSessionRequireProctorEventAccessBranches(t *testing.T) {
	sessionID := handlerTestUUID(30)
	roomID := handlerTestUUID(31)
	eventID := handlerTestUUID(32)
	teacherID := handlerTestUUID(33)
	employeeID := handlerTestUUID(34)
	errDB := errors.New("db down")

	t.Run("admin allowed without service checks", func(t *testing.T) {
		fake := &fakeCbtSessionService{}
		rec := httptest.NewRecorder()
		ok := (&CbtSession{svc: fake}).requireProctorEventAccess(rec, cbtAccessTestRequest(jwt.MapClaims{"roles": []any{"admin"}}), db.GetCbtProctorEventScopeRow{ID: eventID, SessionID: sessionID})
		if !ok || rec.Code != http.StatusOK {
			t.Fatalf("admin access ok/status = %v/%d, want true/no error; body=%s", ok, rec.Code, rec.Body.String())
		}
		if fake.checkSessionID.Valid || fake.roomProctorSessionID.Valid {
			t.Fatalf("admin triggered scoped checks: teacher=%v room=%v", fake.checkSessionID, fake.roomProctorSessionID)
		}
	})

	t.Run("room event requires assigned room proctor", func(t *testing.T) {
		fake := &fakeCbtSessionService{roomProctorAllowed: true}
		rec := httptest.NewRecorder()
		req := cbtAccessTestRequest(jwt.MapClaims{"roles": []any{"guru"}, "eid": employeeID.String()})
		ok := (&CbtSession{svc: fake}).requireProctorEventAccess(rec, req, db.GetCbtProctorEventScopeRow{ID: eventID, SessionID: sessionID, RoomID: roomID})
		if !ok || rec.Code != http.StatusOK {
			t.Fatalf("room proctor access ok/status = %v/%d, want true/no error; body=%s", ok, rec.Code, rec.Body.String())
		}
		if fake.roomProctorSessionID != sessionID || fake.roomProctorRoomID != roomID || fake.roomProctorEmployeeID != employeeID {
			t.Fatalf("room proctor check = %v/%v/%v, want session/room/employee", fake.roomProctorSessionID, fake.roomProctorRoomID, fake.roomProctorEmployeeID)
		}
	})

	t.Run("session event falls back to teacher access", func(t *testing.T) {
		fake := &fakeCbtSessionService{checkAllowed: true}
		rec := httptest.NewRecorder()
		req := cbtAccessTestRequest(jwt.MapClaims{"roles": []any{"guru"}, "eid": teacherID.String()})
		ok := (&CbtSession{svc: fake}).requireProctorEventAccess(rec, req, db.GetCbtProctorEventScopeRow{ID: eventID, SessionID: sessionID})
		if !ok || rec.Code != http.StatusOK {
			t.Fatalf("teacher event access ok/status = %v/%d, want true/no error; body=%s", ok, rec.Code, rec.Body.String())
		}
		if fake.checkSessionID != sessionID || fake.checkTeacherID != teacherID {
			t.Fatalf("teacher check = %v/%v, want session/teacher", fake.checkSessionID, fake.checkTeacherID)
		}
	})

	tests := []struct {
		name       string
		fake       *fakeCbtSessionService
		claims     jwt.MapClaims
		scope      db.GetCbtProctorEventScopeRow
		wantStatus int
	}{
		{name: "room proctor missing employee", fake: &fakeCbtSessionService{roomProctorAllowed: true}, claims: jwt.MapClaims{"roles": []any{"guru"}}, scope: db.GetCbtProctorEventScopeRow{ID: eventID, SessionID: sessionID, RoomID: roomID}, wantStatus: http.StatusForbidden},
		{name: "room proctor denied", fake: &fakeCbtSessionService{roomProctorAllowed: false}, claims: jwt.MapClaims{"roles": []any{"guru"}, "eid": employeeID.String()}, scope: db.GetCbtProctorEventScopeRow{ID: eventID, SessionID: sessionID, RoomID: roomID}, wantStatus: http.StatusForbidden},
		{name: "teacher missing eid", fake: &fakeCbtSessionService{checkAllowed: true}, claims: jwt.MapClaims{"roles": []any{"guru"}}, scope: db.GetCbtProctorEventScopeRow{ID: eventID, SessionID: sessionID}, wantStatus: http.StatusForbidden},
		{name: "teacher denied", fake: &fakeCbtSessionService{checkAllowed: false}, claims: jwt.MapClaims{"roles": []any{"guru"}, "eid": teacherID.String()}, scope: db.GetCbtProctorEventScopeRow{ID: eventID, SessionID: sessionID}, wantStatus: http.StatusForbidden},
		{name: "teacher check error", fake: &fakeCbtSessionService{checkErr: errDB}, claims: jwt.MapClaims{"roles": []any{"guru"}, "eid": teacherID.String()}, scope: db.GetCbtProctorEventScopeRow{ID: eventID, SessionID: sessionID}, wantStatus: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			ok := (&CbtSession{svc: tt.fake}).requireProctorEventAccess(rec, cbtAccessTestRequest(tt.claims), tt.scope)
			if ok || rec.Code != tt.wantStatus {
				t.Fatalf("ok/status = %v/%d, want false/%d; body=%s", ok, rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestCbtSessionRequireSessionAnswerForTeacherOrAdminBranches(t *testing.T) {
	sessionID := handlerTestUUID(40)
	answerID := handlerTestUUID(41)
	teacherID := handlerTestUUID(42)
	errDB := errors.New("db down")

	t.Run("admin uses session answer membership", func(t *testing.T) {
		fake := &fakeCbtSessionService{}
		rec := httptest.NewRecorder()
		ok := (&CbtSession{svc: fake}).requireSessionAnswerForTeacherOrAdmin(rec, cbtAccessTestRequest(jwt.MapClaims{"roles": []any{"admin"}}), sessionID, answerID)
		if !ok || rec.Code != http.StatusOK || fake.hasAnswerSessionID != sessionID || fake.hasAnswerID != answerID || fake.hasAnswerTeacherID.Valid {
			t.Fatalf("admin answer access ok/status/args = %v/%d/%v/%v/%v, want direct answer membership", ok, rec.Code, fake.hasAnswerSessionID, fake.hasAnswerID, fake.hasAnswerTeacherID)
		}
	})

	t.Run("teacher uses teacher-scoped answer membership", func(t *testing.T) {
		fake := &fakeCbtSessionService{}
		rec := httptest.NewRecorder()
		req := cbtAccessTestRequest(jwt.MapClaims{"roles": []any{"guru"}, "eid": teacherID.String()})
		ok := (&CbtSession{svc: fake}).requireSessionAnswerForTeacherOrAdmin(rec, req, sessionID, answerID)
		if !ok || rec.Code != http.StatusOK || fake.hasAnswerSessionID != sessionID || fake.hasAnswerID != answerID || fake.hasAnswerTeacherID != teacherID {
			t.Fatalf("teacher answer access ok/status/args = %v/%d/%v/%v/%v, want teacher-scoped answer check", ok, rec.Code, fake.hasAnswerSessionID, fake.hasAnswerID, fake.hasAnswerTeacherID)
		}
	})

	tests := []struct {
		name       string
		fake       *fakeCbtSessionService
		claims     jwt.MapClaims
		wantStatus int
	}{
		{name: "teacher missing eid", fake: &fakeCbtSessionService{}, claims: jwt.MapClaims{"roles": []any{"guru"}}, wantStatus: http.StatusForbidden},
		{name: "teacher answer denied", fake: &fakeCbtSessionService{hasAnswerSet: true, hasAnswer: false}, claims: jwt.MapClaims{"roles": []any{"guru"}, "eid": teacherID.String()}, wantStatus: http.StatusForbidden},
		{name: "teacher answer error", fake: &fakeCbtSessionService{hasAnswerErr: errDB}, claims: jwt.MapClaims{"roles": []any{"guru"}, "eid": teacherID.String()}, wantStatus: http.StatusInternalServerError},
		{name: "admin answer denied", fake: &fakeCbtSessionService{hasAnswerSet: true, hasAnswer: false}, claims: jwt.MapClaims{"roles": []any{"admin"}}, wantStatus: http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			ok := (&CbtSession{svc: tt.fake}).requireSessionAnswerForTeacherOrAdmin(rec, cbtAccessTestRequest(tt.claims), sessionID, answerID)
			if ok || rec.Code != tt.wantStatus {
				t.Fatalf("ok/status = %v/%d, want false/%d; body=%s", ok, rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestCbtSessionRequireParticipantProctorActionAccessBranches(t *testing.T) {
	sessionID := handlerTestUUID(50)
	participantID := handlerTestUUID(51)
	roomID := handlerTestUUID(52)
	employeeID := handlerTestUUID(53)
	errDB := errors.New("db down")

	t.Run("admin can act even when participant has no room", func(t *testing.T) {
		fake := &fakeCbtSessionService{participantScope: db.GetCbtParticipantProctorScopeRow{ParticipantID: participantID, SessionID: sessionID, Nama: "Ahmad"}}
		rec := httptest.NewRecorder()
		scope, ok := (&CbtSession{svc: fake}).requireParticipantProctorActionAccess(rec, cbtAccessTestRequest(jwt.MapClaims{"roles": []any{"admin"}}), sessionID, participantID)
		if !ok || rec.Code != http.StatusOK || scope.ParticipantID != participantID || fake.participantScopeSessionID != sessionID || fake.participantScopeID != participantID {
			t.Fatalf("admin participant action scope/ok/status/args = %+v/%v/%d/%v/%v, want participant scope", scope, ok, rec.Code, fake.participantScopeSessionID, fake.participantScopeID)
		}
	})

	t.Run("assigned room proctor can act on participant in room", func(t *testing.T) {
		fake := &fakeCbtSessionService{roomProctorAllowed: true, participantScope: db.GetCbtParticipantProctorScopeRow{ParticipantID: participantID, SessionID: sessionID, RoomID: roomID, Nama: "Ahmad"}}
		rec := httptest.NewRecorder()
		req := cbtAccessTestRequest(jwt.MapClaims{"roles": []any{"guru"}, "eid": employeeID.String()})
		scope, ok := (&CbtSession{svc: fake}).requireParticipantProctorActionAccess(rec, req, sessionID, participantID)
		if !ok || rec.Code != http.StatusOK || scope.RoomID != roomID {
			t.Fatalf("proctor participant action scope/ok/status = %+v/%v/%d, want room scope allowed", scope, ok, rec.Code)
		}
		if fake.roomProctorSessionID != sessionID || fake.roomProctorRoomID != roomID || fake.roomProctorEmployeeID != employeeID {
			t.Fatalf("room proctor check = %v/%v/%v, want session/room/employee", fake.roomProctorSessionID, fake.roomProctorRoomID, fake.roomProctorEmployeeID)
		}
	})

	tests := []struct {
		name       string
		fake       *fakeCbtSessionService
		claims     jwt.MapClaims
		wantStatus int
	}{
		{name: "scope lookup error", fake: &fakeCbtSessionService{participantScopeErr: errDB}, claims: jwt.MapClaims{"roles": []any{"admin"}}, wantStatus: http.StatusInternalServerError},
		{name: "non-admin participant without room denied", fake: &fakeCbtSessionService{participantScope: db.GetCbtParticipantProctorScopeRow{ParticipantID: participantID, SessionID: sessionID}}, claims: jwt.MapClaims{"roles": []any{"guru"}, "eid": employeeID.String()}, wantStatus: http.StatusForbidden},
		{name: "room proctor missing eid denied", fake: &fakeCbtSessionService{participantScope: db.GetCbtParticipantProctorScopeRow{ParticipantID: participantID, SessionID: sessionID, RoomID: roomID}}, claims: jwt.MapClaims{"roles": []any{"guru"}}, wantStatus: http.StatusForbidden},
		{name: "room proctor unassigned denied", fake: &fakeCbtSessionService{roomProctorAllowed: false, participantScope: db.GetCbtParticipantProctorScopeRow{ParticipantID: participantID, SessionID: sessionID, RoomID: roomID}}, claims: jwt.MapClaims{"roles": []any{"guru"}, "eid": employeeID.String()}, wantStatus: http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			scope, ok := (&CbtSession{svc: tt.fake}).requireParticipantProctorActionAccess(rec, cbtAccessTestRequest(tt.claims), sessionID, participantID)
			if ok || rec.Code != tt.wantStatus || scope.ParticipantID.Valid {
				t.Fatalf("scope/ok/status = %+v/%v/%d, want empty false/%d; body=%s", scope, ok, rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
