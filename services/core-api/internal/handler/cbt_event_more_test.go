package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"mtsn2kolut-super-app/backend/internal/domain"
	"mtsn2kolut-super-app/backend/internal/service"
)

func TestCbtEventMoreQuestionRequirementsEdges(t *testing.T) {
	eventID := handlerTestUUID(180)
	boom := errors.Join(domain.ErrBadRequest, errors.New("scope_mode tidak valid"))
	tests := []struct {
		name       string
		req        *http.Request
		svc        *fakeCbtEventService
		wantStatus int
	}{
		{
			name:       "forbidden without admin",
			req:        httptest.NewRequest(http.MethodPut, "/api/cbt/events/"+eventID.String()+"/question-requirements", nil),
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "invalid event id",
			req:        withRouteParam(adminRequest(http.MethodPut, "/api/cbt/events/bad/question-requirements", `{}`), "id", "bad"),
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid json",
			req:        withRouteParam(adminRequest(http.MethodPut, "/api/cbt/events/"+eventID.String()+"/question-requirements", `{`), "id", eventID.String()),
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "service validation error",
			req:        withRouteParam(adminRequest(http.MethodPut, "/api/cbt/events/"+eventID.String()+"/question-requirements", `{"scope_mode":"bad"}`), "id", eventID.String()),
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}, upsertRequirementsErr: boom},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			(&CbtEvent{svc: tt.svc}).UpsertQuestionRequirements(rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestCbtEventMoreMemberMutationEdges(t *testing.T) {
	eventID := handlerTestUUID(181)
	memberID := handlerTestUUID(182)
	userID := handlerTestUUID(183)
	body := `{"user_id":"` + userID.String() + `","role":"panitia"}`
	boom := errors.Join(domain.ErrConflict, errors.New("anggota sudah ada"))
	tests := []struct {
		name       string
		handler    func(*CbtEvent, http.ResponseWriter, *http.Request)
		req        *http.Request
		svc        *fakeCbtEventService
		wantStatus int
	}{
		{
			name:       "create forbidden",
			handler:    (*CbtEvent).CreateMember,
			req:        httptest.NewRequest(http.MethodPost, "/api/cbt/events/"+eventID.String()+"/members", nil),
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "create invalid event id",
			handler:    (*CbtEvent).CreateMember,
			req:        withRouteParam(adminRequest(http.MethodPost, "/api/cbt/events/bad/members", body), "id", "bad"),
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "create invalid payload",
			handler:    (*CbtEvent).CreateMember,
			req:        withRouteParam(adminRequest(http.MethodPost, "/api/cbt/events/"+eventID.String()+"/members", `{"user_id":"bad","role":"panitia"}`), "id", eventID.String()),
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "create service conflict",
			handler:    (*CbtEvent).CreateMember,
			req:        withRouteParam(adminRequest(http.MethodPost, "/api/cbt/events/"+eventID.String()+"/members", body), "id", eventID.String()),
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}, createMemberErr: boom},
			wantStatus: http.StatusConflict,
		},
		{
			name:       "update forbidden",
			handler:    (*CbtEvent).UpdateMember,
			req:        httptest.NewRequest(http.MethodPatch, "/api/cbt/events/"+eventID.String()+"/members/"+memberID.String(), nil),
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "update invalid route id",
			handler:    (*CbtEvent).UpdateMember,
			req:        withRouteParams(adminRequest(http.MethodPatch, "/api/cbt/events/"+eventID.String()+"/members/bad", body), "id", eventID.String(), "member_id", "bad"),
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "update invalid payload",
			handler:    (*CbtEvent).UpdateMember,
			req:        withRouteParams(adminRequest(http.MethodPatch, "/api/cbt/events/"+eventID.String()+"/members/"+memberID.String(), `{`), "id", eventID.String(), "member_id", memberID.String()),
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "update service error",
			handler:    (*CbtEvent).UpdateMember,
			req:        withRouteParams(adminRequest(http.MethodPatch, "/api/cbt/events/"+eventID.String()+"/members/"+memberID.String(), body), "id", eventID.String(), "member_id", memberID.String()),
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}, updateMemberErr: errors.New("not found")},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "delete forbidden",
			handler:    (*CbtEvent).DeleteMember,
			req:        httptest.NewRequest(http.MethodDelete, "/api/cbt/events/"+eventID.String()+"/members/"+memberID.String(), nil),
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "delete invalid route id",
			handler:    (*CbtEvent).DeleteMember,
			req:        withRouteParams(adminRequest(http.MethodDelete, "/api/cbt/events/bad/members/"+memberID.String(), ""), "id", "bad", "member_id", memberID.String()),
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "delete service error",
			handler:    (*CbtEvent).DeleteMember,
			req:        withRouteParams(adminRequest(http.MethodDelete, "/api/cbt/events/"+eventID.String()+"/members/"+memberID.String(), ""), "id", eventID.String(), "member_id", memberID.String()),
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}, deleteMemberErr: errors.New("akses ditolak")},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.handler(&CbtEvent{svc: tt.svc}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestCbtEventMoreQuestionTargetMutationEdges(t *testing.T) {
	eventID := handlerTestUUID(184)
	subjectID := handlerTestUUID(185)
	validBody := `{"subject_id":"` + subjectID.String() + `","target_questions":10}`
	tests := []struct {
		name       string
		handler    func(*CbtEvent, http.ResponseWriter, *http.Request)
		req        *http.Request
		svc        *fakeCbtEventService
		wantStatus int
	}{
		{
			name:       "upsert forbidden",
			handler:    (*CbtEvent).UpsertQuestionTarget,
			req:        httptest.NewRequest(http.MethodPut, "/api/cbt/events/"+eventID.String()+"/question-targets", nil),
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "upsert invalid event id",
			handler:    (*CbtEvent).UpsertQuestionTarget,
			req:        withRouteParam(adminRequest(http.MethodPut, "/api/cbt/events/bad/question-targets", validBody), "id", "bad"),
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "upsert invalid json",
			handler:    (*CbtEvent).UpsertQuestionTarget,
			req:        withRouteParam(adminRequest(http.MethodPut, "/api/cbt/events/"+eventID.String()+"/question-targets", `{`), "id", eventID.String()),
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "upsert invalid subject",
			handler:    (*CbtEvent).UpsertQuestionTarget,
			req:        withRouteParam(adminRequest(http.MethodPut, "/api/cbt/events/"+eventID.String()+"/question-targets", `{"subject_id":"bad","target_questions":10}`), "id", eventID.String()),
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "upsert service error",
			handler:    (*CbtEvent).UpsertQuestionTarget,
			req:        withRouteParam(adminRequest(http.MethodPut, "/api/cbt/events/"+eventID.String()+"/question-targets", validBody), "id", eventID.String()),
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}, upsertQuestionTargetErr: errors.New("duplicate target")},
			wantStatus: http.StatusConflict,
		},
		{
			name:       "delete forbidden",
			handler:    (*CbtEvent).DeleteQuestionTarget,
			req:        httptest.NewRequest(http.MethodDelete, "/api/cbt/events/"+eventID.String()+"/question-targets/"+subjectID.String(), nil),
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "delete invalid event id",
			handler:    (*CbtEvent).DeleteQuestionTarget,
			req:        withRouteParams(adminRequest(http.MethodDelete, "/api/cbt/events/bad/question-targets/"+subjectID.String(), ""), "id", "bad", "subject_id", subjectID.String()),
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "delete invalid subject id",
			handler:    (*CbtEvent).DeleteQuestionTarget,
			req:        withRouteParams(adminRequest(http.MethodDelete, "/api/cbt/events/"+eventID.String()+"/question-targets/bad", ""), "id", eventID.String(), "subject_id", "bad"),
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "delete service error",
			handler:    (*CbtEvent).DeleteQuestionTarget,
			req:        withRouteParams(adminRequest(http.MethodDelete, "/api/cbt/events/"+eventID.String()+"/question-targets/"+subjectID.String(), ""), "id", eventID.String(), "subject_id", subjectID.String()),
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}, deleteQuestionTargetErr: errors.New("target tidak ditemukan")},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.handler(&CbtEvent{svc: tt.svc}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
