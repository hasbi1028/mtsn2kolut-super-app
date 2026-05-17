package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

func TestCbtEventQuestionRequirementsHandlersForwardPayloads(t *testing.T) {
	eventID := handlerTestUUID(156)
	fake := &fakeCbtEventService{
		CbtEvent: &service.CbtEvent{},
		requirementsRow: db.GetCbtEventQuestionRequirementsRow{
			ID: eventID, EventID: eventID, ScopeMode: "per_level", TargetPg: 30, TargetEssay: 5, StatusFilter: "all_progress",
		},
		upsertRequirementsRow: db.UpsertCbtEventQuestionRequirementsRow{
			ID: eventID, EventID: eventID, ScopeMode: "per_rombel", TargetPg: 25, TargetEssay: 3, StatusFilter: "published_only",
		},
	}
	h := &CbtEvent{svc: fake}

	rec := httptest.NewRecorder()
	h.GetQuestionRequirements(rec, withRouteParam(adminRequest(http.MethodGet, "/api/cbt/events/"+eventID.String()+"/question-requirements", ""), "id", eventID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("GetQuestionRequirements status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.requirementsID != eventID || !strings.Contains(rec.Body.String(), "per_level") {
		t.Fatalf("GetQuestionRequirements id/body = %v/%s, want %v and row", fake.requirementsID, rec.Body.String(), eventID)
	}

	rec = httptest.NewRecorder()
	h.UpsertQuestionRequirements(rec, withRouteParam(adminRequest(http.MethodPut, "/api/cbt/events/"+eventID.String()+"/question-requirements", `{"scope_mode":"per_rombel","target_pg":25,"target_essay":3,"status_filter":"published_only"}`), "id", eventID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpsertQuestionRequirements status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.upsertRequirementsID != eventID || fake.upsertRequirementsInput.ScopeMode != "per_rombel" || fake.upsertRequirementsInput.TargetPg != 25 || fake.upsertRequirementsInput.TargetEssay != 3 || fake.upsertRequirementsInput.StatusFilter != "published_only" {
		t.Fatalf("UpsertQuestionRequirements input = id %v input %+v, want decoded payload", fake.upsertRequirementsID, fake.upsertRequirementsInput)
	}
}

func TestCbtEventMemberHandlersForwardPayloads(t *testing.T) {
	eventID := handlerTestUUID(157)
	memberID := handlerTestUUID(158)
	userID := handlerTestUUID(159)
	employeeID := handlerTestUUID(160)
	subjectID := handlerTestUUID(161)
	fake := &fakeCbtEventService{
		CbtEvent:        &service.CbtEvent{},
		membersRows:     []db.ListCbtEventMembersRow{{ID: memberID, EventID: eventID, UserID: userID, Username: "guru", Role: db.CbtEventMemberRolePembuatSoal}},
		createMemberRow: db.CbtEventMember{ID: memberID, EventID: eventID, UserID: userID, EmployeeID: employeeID, SubjectID: subjectID, Role: db.CbtEventMemberRolePembuatSoal},
		updateMemberRow: db.CbtEventMember{ID: memberID, EventID: eventID, UserID: userID, EmployeeID: employeeID, SubjectID: subjectID, Role: db.CbtEventMemberRoleReviewer},
	}
	h := &CbtEvent{svc: fake}

	rec := httptest.NewRecorder()
	h.ListMembers(rec, withRouteParam(adminRequest(http.MethodGet, "/api/cbt/events/"+eventID.String()+"/members", ""), "id", eventID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("ListMembers status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.membersID != eventID || !strings.Contains(rec.Body.String(), "guru") {
		t.Fatalf("ListMembers id/body = %v/%s, want %v and member", fake.membersID, rec.Body.String(), eventID)
	}

	createBody := `{"user_id":"` + userID.String() + `","employee_id":"` + employeeID.String() + `","subject_id":"` + subjectID.String() + `","role":"pembuat_soal"}`
	rec = httptest.NewRecorder()
	h.CreateMember(rec, withRouteParam(adminRequest(http.MethodPost, "/api/cbt/events/"+eventID.String()+"/members", createBody), "id", eventID.String()))
	if rec.Code != http.StatusCreated {
		t.Fatalf("CreateMember status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if fake.createMemberEventID != eventID || fake.createMemberInput.UserID != userID || fake.createMemberInput.EmployeeID != employeeID || fake.createMemberInput.SubjectID != subjectID || fake.createMemberInput.Role != db.CbtEventMemberRolePembuatSoal {
		t.Fatalf("CreateMember input = event %v input %+v, want decoded payload", fake.createMemberEventID, fake.createMemberInput)
	}

	updateBody := `{"user_id":"` + userID.String() + `","employee_id":"` + employeeID.String() + `","subject_id":"` + subjectID.String() + `","role":"reviewer"}`
	rec = httptest.NewRecorder()
	h.UpdateMember(rec, withRouteParams(adminRequest(http.MethodPatch, "/api/cbt/events/"+eventID.String()+"/members/"+memberID.String(), updateBody), "id", eventID.String(), "member_id", memberID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateMember status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.updateMemberEventID != eventID || fake.updateMemberID != memberID || fake.updateMemberInput.Role != db.CbtEventMemberRoleReviewer || fake.updateMemberInput.UserID != userID {
		t.Fatalf("UpdateMember input = event %v member %v input %+v, want route ids and decoded payload", fake.updateMemberEventID, fake.updateMemberID, fake.updateMemberInput)
	}

	rec = httptest.NewRecorder()
	h.DeleteMember(rec, withRouteParams(adminRequest(http.MethodDelete, "/api/cbt/events/"+eventID.String()+"/members/"+memberID.String(), ""), "id", eventID.String(), "member_id", memberID.String()))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DeleteMember status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}
	if fake.deleteMemberEventID != eventID || fake.deleteMemberID != memberID {
		t.Fatalf("DeleteMember ids = (%v,%v), want (%v,%v)", fake.deleteMemberEventID, fake.deleteMemberID, eventID, memberID)
	}
}

func TestCbtEventQuestionTargetHandlersForwardPayloads(t *testing.T) {
	eventID := handlerTestUUID(162)
	targetID := handlerTestUUID(163)
	subjectID := handlerTestUUID(164)
	fake := &fakeCbtEventService{
		CbtEvent:                &service.CbtEvent{},
		questionTargetsRows:     []db.ListCbtEventSubjectTargetsRow{{ID: targetID, EventID: eventID, SubjectID: subjectID, SubjectName: "Matematika", TargetQuestions: 40}},
		upsertQuestionTargetRow: db.CbtEventSubjectTarget{ID: targetID, EventID: eventID, SubjectID: subjectID, TargetQuestions: 45},
	}
	h := &CbtEvent{svc: fake}

	rec := httptest.NewRecorder()
	h.ListQuestionTargets(rec, withRouteParam(adminRequest(http.MethodGet, "/api/cbt/events/"+eventID.String()+"/question-targets", ""), "id", eventID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("ListQuestionTargets status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.questionTargetsID != eventID || !strings.Contains(rec.Body.String(), "Matematika") {
		t.Fatalf("ListQuestionTargets id/body = %v/%s, want %v and target", fake.questionTargetsID, rec.Body.String(), eventID)
	}

	rec = httptest.NewRecorder()
	h.UpsertQuestionTarget(rec, withRouteParam(adminRequest(http.MethodPut, "/api/cbt/events/"+eventID.String()+"/question-targets", `{"subject_id":"`+subjectID.String()+`","target_questions":45}`), "id", eventID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpsertQuestionTarget status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.upsertQuestionTargetEventID != eventID || fake.upsertQuestionTargetInput.SubjectID != subjectID || fake.upsertQuestionTargetInput.TargetQuestions != 45 {
		t.Fatalf("UpsertQuestionTarget input = event %v input %+v, want decoded payload", fake.upsertQuestionTargetEventID, fake.upsertQuestionTargetInput)
	}

	rec = httptest.NewRecorder()
	h.DeleteQuestionTarget(rec, withRouteParams(adminRequest(http.MethodDelete, "/api/cbt/events/"+eventID.String()+"/question-targets/"+subjectID.String(), ""), "id", eventID.String(), "subject_id", subjectID.String()))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DeleteQuestionTarget status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}
	if fake.deleteQuestionTargetEventID != eventID || fake.deleteQuestionTargetSubjectID != subjectID {
		t.Fatalf("DeleteQuestionTarget ids = (%v,%v), want (%v,%v)", fake.deleteQuestionTargetEventID, fake.deleteQuestionTargetSubjectID, eventID, subjectID)
	}
}

func TestCbtEventReadinessCoverageHandlersMapServiceErrors(t *testing.T) {
	eventID := handlerTestUUID(172)
	boom := errors.New("db down")
	tests := []struct {
		name       string
		handler    func(*CbtEvent, http.ResponseWriter, *http.Request)
		svc        *fakeCbtEventService
		path       string
		wantStatus int
	}{
		{name: "overview", handler: (*CbtEvent).Overview, svc: &fakeCbtEventService{CbtEvent: &service.CbtEvent{}, overviewErr: boom}, path: "/api/cbt/events/" + eventID.String() + "/overview", wantStatus: http.StatusInternalServerError},
		{name: "sop readiness", handler: (*CbtEvent).SopReadiness, svc: &fakeCbtEventService{CbtEvent: &service.CbtEvent{}, sopErr: boom}, path: "/api/cbt/events/" + eventID.String() + "/sop-readiness", wantStatus: http.StatusInternalServerError},
		{name: "packages", handler: (*CbtEvent).ListPackages, svc: &fakeCbtEventService{CbtEvent: &service.CbtEvent{}, packagesErr: boom}, path: "/api/cbt/events/" + eventID.String() + "/packages", wantStatus: http.StatusInternalServerError},
		{name: "sessions", handler: (*CbtEvent).ListSessions, svc: &fakeCbtEventService{CbtEvent: &service.CbtEvent{}, sessionsErr: boom}, path: "/api/cbt/events/" + eventID.String() + "/sessions", wantStatus: http.StatusInternalServerError},
		{name: "question completeness", handler: (*CbtEvent).QuestionCompleteness, svc: &fakeCbtEventService{CbtEvent: &service.CbtEvent{}, completenessErr: boom}, path: "/api/cbt/events/" + eventID.String() + "/question-completeness", wantStatus: http.StatusInternalServerError},
		{name: "requirements", handler: (*CbtEvent).GetQuestionRequirements, svc: &fakeCbtEventService{CbtEvent: &service.CbtEvent{}, requirementsErr: boom}, path: "/api/cbt/events/" + eventID.String() + "/question-requirements", wantStatus: http.StatusInternalServerError},
		{name: "members", handler: (*CbtEvent).ListMembers, svc: &fakeCbtEventService{CbtEvent: &service.CbtEvent{}, membersErr: boom}, path: "/api/cbt/events/" + eventID.String() + "/members", wantStatus: http.StatusBadRequest},
		{name: "targets", handler: (*CbtEvent).ListQuestionTargets, svc: &fakeCbtEventService{CbtEvent: &service.CbtEvent{}, questionTargetsErr: boom}, path: "/api/cbt/events/" + eventID.String() + "/question-targets", wantStatus: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.handler(&CbtEvent{svc: tt.svc}, rec, withRouteParam(adminRequest(http.MethodGet, tt.path, ""), "id", eventID.String()))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestCbtEventReadinessCoverageHandlersRejectInvalidIDs(t *testing.T) {
	tests := []struct {
		name    string
		handler func(*CbtEvent, http.ResponseWriter, *http.Request)
		path    string
	}{
		{name: "overview", handler: (*CbtEvent).Overview, path: "/api/cbt/events/bad/overview"},
		{name: "sop readiness", handler: (*CbtEvent).SopReadiness, path: "/api/cbt/events/bad/sop-readiness"},
		{name: "packages", handler: (*CbtEvent).ListPackages, path: "/api/cbt/events/bad/packages"},
		{name: "sessions", handler: (*CbtEvent).ListSessions, path: "/api/cbt/events/bad/sessions"},
		{name: "question completeness", handler: (*CbtEvent).QuestionCompleteness, path: "/api/cbt/events/bad/question-completeness"},
		{name: "requirements", handler: (*CbtEvent).GetQuestionRequirements, path: "/api/cbt/events/bad/question-requirements"},
		{name: "members", handler: (*CbtEvent).ListMembers, path: "/api/cbt/events/bad/members"},
		{name: "targets", handler: (*CbtEvent).ListQuestionTargets, path: "/api/cbt/events/bad/question-targets"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.handler(&CbtEvent{svc: &fakeCbtEventService{CbtEvent: &service.CbtEvent{}}}, rec, withRouteParam(adminRequest(http.MethodGet, tt.path, ""), "id", "bad"))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestCbtEventMemberInputRejectsInvalidPayloads(t *testing.T) {
	userID := handlerTestUUID(165)
	employeeID := handlerTestUUID(166)
	subjectID := handlerTestUUID(167)
	tests := []struct {
		name string
		body string
	}{
		{name: "bad json", body: `{"user_id":`},
		{name: "missing user", body: `{"role":"panitia"}`},
		{name: "bad employee", body: `{"user_id":"` + userID.String() + `","employee_id":"bad","role":"panitia"}`},
		{name: "bad subject", body: `{"user_id":"` + userID.String() + `","subject_id":"bad","role":"panitia"}`},
		{name: "bad role", body: `{"user_id":"` + userID.String() + `","employee_id":"` + employeeID.String() + `","subject_id":"` + subjectID.String() + `","role":"admin"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := cbtEventMemberInputFromRequest(httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))); err == nil {
				t.Fatalf("cbtEventMemberInputFromRequest err = nil, want error")
			}
		})
	}
}

func TestCbtEventReadAccessForMembersAndTargetsUsesMembership(t *testing.T) {
	eventID := handlerTestUUID(168)
	userID := handlerTestUUID(169)
	fake := &fakeCbtEventService{CbtEvent: &service.CbtEvent{}, canReadAllowed: true}
	h := &CbtEvent{svc: fake}
	req := withClaims(httptest.NewRequest(http.MethodGet, "/api/cbt/events/"+eventID.String()+"/members", nil), jwt.MapClaims{
		"roles": []any{"guru"}, "role": "guru", "uid": userID.String(),
	})

	rec := httptest.NewRecorder()
	h.ListMembers(rec, withRouteParam(req, "id", eventID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("ListMembers(guru) status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.canReadEventID != eventID || fake.canReadUserID != userID || fake.membersID != eventID {
		t.Fatalf("ListMembers(guru) ids = canRead(%v,%v) members %v, want event/user", fake.canReadEventID, fake.canReadUserID, fake.membersID)
	}

	rec = httptest.NewRecorder()
	fake.canReadAllowed = false
	h.ListQuestionTargets(rec, withRouteParam(req, "id", eventID.String()))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("ListQuestionTargets(guru denied) status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}
}

func TestCbtEventMemberRouteIDsRejectInvalidIDs(t *testing.T) {
	eventID := handlerTestUUID(170)
	memberID := handlerTestUUID(171)
	tests := []struct {
		name     string
		id       string
		memberID string
	}{
		{name: "bad event", id: "bad", memberID: memberID.String()},
		{name: "bad member", id: eventID.String(), memberID: "bad"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			_, _, ok := cbtEventMemberRouteIDs(rec, withRouteParams(adminRequest(http.MethodPatch, "/", ""), "id", tt.id, "member_id", tt.memberID))
			if ok {
				t.Fatalf("cbtEventMemberRouteIDs ok = true, want false")
			}
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}
