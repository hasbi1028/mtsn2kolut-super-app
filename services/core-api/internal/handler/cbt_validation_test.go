package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"mtsn2kolut-super-app/backend/internal/service"
)

func adminClaimsReq(method, target, body string) *http.Request {
	return withClaims(httptest.NewRequest(method, target, strings.NewReader(body)), jwt.MapClaims{"roles": []any{"admin"}})
}

func assertHandlerBadRequest(t *testing.T, name string, fn http.HandlerFunc) {
	t.Helper()
	rec := httptest.NewRecorder()
	req := adminClaimsReq(http.MethodPost, "/bad-request", `{}`)
	fn(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("%s status = %d, want 400; body=%s", name, rec.Code, rec.Body.String())
	}
}

func TestCbtSessionHandlersRejectInvalidIDs(t *testing.T) {
	h := &CbtSession{svc: &service.CbtSession{}}
	tests := []struct {
		name string
		fn   http.HandlerFunc
	}{
		{name: "Get", fn: h.Get},
		{name: "UpdateStatus", fn: h.UpdateStatus},
		{name: "Delete", fn: h.Delete},
		{name: "ListParticipants", fn: h.ListParticipants},
		{name: "Enroll", fn: h.Enroll},
		{name: "EnrollGrade", fn: h.EnrollGrade},
		{name: "EnrollSchool", fn: h.EnrollSchool},
		{name: "GenerateTokens", fn: h.GenerateTokens},
		{name: "RegenerateToken", fn: h.RegenerateToken},
		{name: "AssignSeat", fn: h.AssignSeat},
		{name: "AutoAssignSeats", fn: h.AutoAssignSeats},
		{name: "ListRooms", fn: h.ListRooms},
		{name: "CreateRoom", fn: h.CreateRoom},
		{name: "DeleteRoom", fn: h.DeleteRoom},
		{name: "ShuffleRooms", fn: h.ShuffleRooms},
		{name: "GetProctoringStatus", fn: h.GetProctoringStatus},
		{name: "FlagParticipant", fn: h.FlagParticipant},
		{name: "ListUngradedEssays", fn: h.ListUngradedEssays},
		{name: "GradeEssay", fn: h.GradeEssay},
		{name: "RecordAnswer", fn: h.RecordAnswer},
		{name: "ScoreSession", fn: h.ScoreSession},
		{name: "GuruAwareResults", fn: h.GuruAwareResults},
		{name: "GuruAwareParticipants", fn: h.GuruAwareParticipants},
		{name: "GetResults", fn: h.GetResults},
		{name: "GetMinutes", fn: h.GetMinutes},
		{name: "GetParticipantAnswers", fn: h.GetParticipantAnswers},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/bad-request", strings.NewReader(`{}`))
			switch tt.name {
			case "UpdateStatus", "Delete", "Enroll", "EnrollGrade", "EnrollSchool", "GenerateTokens", "RegenerateToken", "AssignSeat", "AutoAssignSeats", "CreateRoom", "DeleteRoom", "ShuffleRooms", "RecordAnswer", "ScoreSession", "GetResults":
				req = withClaims(req, jwt.MapClaims{"roles": []any{"admin"}})
			}
			tt.fn(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("%s status = %d, want 400; body=%s", tt.name, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestCbtSessionCreateRejectsMalformedInput(t *testing.T) {
	h := &CbtSession{svc: &service.CbtSession{}}
	tests := []struct {
		name string
		body string
	}{
		{name: "invalid json", body: `{`},
		{name: "invalid package", body: `{"package_id":"bad"}`},
		{name: "class scope requires class", body: `{"package_id":"01000000-0000-0000-0000-000000000000","scope_type":"class"}`},
		{name: "cross grade requires special event", body: `{"package_id":"01000000-0000-0000-0000-000000000000","class_id":"02000000-0000-0000-0000-000000000000","allow_cross_grade":true}`},
		{name: "grade scope requires ref", body: `{"package_id":"01000000-0000-0000-0000-000000000000","scope_type":"grade"}`},
		{name: "bad start", body: `{"package_id":"01000000-0000-0000-0000-000000000000","class_id":"02000000-0000-0000-0000-000000000000","scheduled_start":"bad"}`},
		{name: "bad end", body: `{"package_id":"01000000-0000-0000-0000-000000000000","class_id":"02000000-0000-0000-0000-000000000000","scheduled_start":"2026-05-01T08:00:00Z","scheduled_end":"bad"}`},
		{name: "end before start", body: `{"package_id":"01000000-0000-0000-0000-000000000000","class_id":"02000000-0000-0000-0000-000000000000","scheduled_start":"2026-05-01T08:00:00Z","scheduled_end":"2026-05-01T07:00:00Z"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := adminClaimsReq(http.MethodPost, "/cbt/sessions", tt.body)
			h.Create(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("Create(%s) status = %d, want 400; body=%s", tt.name, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestCbtSessionRequiresClaimsForProtectedSession(t *testing.T) {
	h := &CbtSession{svc: &service.CbtSession{}}
	req := httptest.NewRequest(http.MethodGet, "/cbt/sessions/1", nil)
	req = withRouteParam(req, "id", "01000000-0000-0000-0000-000000000000")
	rec := httptest.NewRecorder()

	h.Get(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Get(valid id without claims) status = %d, want 401; body=%s", rec.Code, rec.Body.String())
	}
}

func TestCbtEventHandlersRejectMalformedInput(t *testing.T) {
	h := &CbtEvent{svc: &service.CbtEvent{}}
	invalidIDHandlers := []struct {
		name string
		fn   http.HandlerFunc
	}{
		{name: "Get", fn: h.Get},
		{name: "GetResults", fn: h.GetResults},
		{name: "GetExamCards", fn: h.GetExamCards},
		{name: "Update", fn: h.Update},
		{name: "UpdateStatus", fn: h.UpdateStatus},
		{name: "Delete", fn: h.Delete},
	}
	for _, tt := range invalidIDHandlers {
		t.Run("invalid id "+tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/bad-request", strings.NewReader(`{}`))
			req = withClaims(req, jwt.MapClaims{"roles": []any{"admin"}})
			tt.fn(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("%s status = %d, want 400; body=%s", tt.name, rec.Code, rec.Body.String())
			}
		})
	}

	tests := []struct {
		name string
		fn   http.HandlerFunc
		body string
		id   bool
	}{
		{name: "create invalid json", fn: h.Create, body: `{`},
		{name: "create missing title", fn: h.Create, body: `{}`},
		{name: "create invalid target levels", fn: h.Create, body: `{"title":"PTS","target_levels":["X"]}`},
		{name: "create invalid academic year", fn: h.Create, body: `{"title":"PTS","academic_year_id":"bad"}`},
		{name: "update invalid json", fn: h.Update, body: `{`, id: true},
		{name: "update invalid target levels", fn: h.Update, body: `{"target_levels":["X"]}`, id: true},
		{name: "update invalid academic year", fn: h.Update, body: `{"academic_year_id":"bad"}`, id: true},
		{name: "status invalid json", fn: h.UpdateStatus, body: `{`, id: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/cbt/events", strings.NewReader(tt.body))
			req = withClaims(req, jwt.MapClaims{"roles": []any{"admin"}})
			if tt.id {
				req = withRouteParam(req, "id", "01000000-0000-0000-0000-000000000000")
			}
			tt.fn(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("%s status = %d, want 400; body=%s", tt.name, rec.Code, rec.Body.String())
			}
		})
	}
}
