package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeCbtEventService struct {
	*service.CbtEvent

	listRows    []db.ListCbtExamEventsRow
	listErr     error
	getID       pgtype.UUID
	getRow      db.GetCbtExamEventRow
	getErr      error
	resultsID   pgtype.UUID
	resultsRows []db.GetEventResultsRow
	resultsErr  error
	cardsID     pgtype.UUID
	cardsRows   []db.GetEventExamCardsRow
	cardsErr    error
	createInput service.CreateCbtEventInput
	createRow   db.CbtExamEvent
	createErr   error
	updateID    pgtype.UUID
	updateInput service.CreateCbtEventInput
	updateRow   db.CbtExamEvent
	updateErr   error
	statusID    pgtype.UUID
	statusValue string
	statusRow   db.CbtExamEvent
	statusErr   error
	deleteID    pgtype.UUID
	deleteErr   error
}

func (f *fakeCbtEventService) List(context.Context) ([]db.ListCbtExamEventsRow, error) {
	return f.listRows, f.listErr
}

func (f *fakeCbtEventService) Get(_ context.Context, id pgtype.UUID) (db.GetCbtExamEventRow, error) {
	f.getID = id
	return f.getRow, f.getErr
}

func (f *fakeCbtEventService) GetResults(_ context.Context, id pgtype.UUID) ([]db.GetEventResultsRow, error) {
	f.resultsID = id
	return f.resultsRows, f.resultsErr
}

func (f *fakeCbtEventService) GetExamCards(_ context.Context, id pgtype.UUID) ([]db.GetEventExamCardsRow, error) {
	f.cardsID = id
	return f.cardsRows, f.cardsErr
}

func (f *fakeCbtEventService) Create(_ context.Context, in service.CreateCbtEventInput) (db.CbtExamEvent, error) {
	f.createInput = in
	return f.createRow, f.createErr
}

func (f *fakeCbtEventService) Update(_ context.Context, id pgtype.UUID, in service.CreateCbtEventInput) (db.CbtExamEvent, error) {
	f.updateID = id
	f.updateInput = in
	return f.updateRow, f.updateErr
}

func (f *fakeCbtEventService) UpdateStatus(_ context.Context, id pgtype.UUID, status string) (db.CbtExamEvent, error) {
	f.statusID = id
	f.statusValue = status
	return f.statusRow, f.statusErr
}

func (f *fakeCbtEventService) Delete(_ context.Context, id pgtype.UUID) error {
	f.deleteID = id
	return f.deleteErr
}

func cbtEventModel(id pgtype.UUID, title string) db.CbtExamEvent {
	return db.CbtExamEvent{ID: id, Title: title, ExamType: db.CbtExamTypeLainnya, Scope: "class", Status: "draft"}
}

func assertCbtTargetLevels(t *testing.T, got []string, wants ...string) {
	t.Helper()
	if len(got) != len(wants) {
		t.Fatalf("target levels = %v, want %v", got, wants)
	}
	set := make(map[string]bool, len(got))
	for _, item := range got {
		set[item] = true
	}
	for _, want := range wants {
		if !set[want] {
			t.Fatalf("target levels = %v, missing %q", got, want)
		}
	}
}

func TestCbtEventReadHandlersForwardIDs(t *testing.T) {
	eventID := handlerTestUUID(250)
	fake := &fakeCbtEventService{
		CbtEvent: &service.CbtEvent{},
		listRows: []db.ListCbtExamEventsRow{
			{ID: eventID, Title: "PAT 2026", ExamType: db.CbtExamTypeUas, Scope: "grade", Status: "active", SessionCount: 2},
		},
		getRow:      db.GetCbtExamEventRow{ID: eventID, Title: "PAT 2026", ExamType: db.CbtExamTypeUas},
		resultsRows: []db.GetEventResultsRow{{ParticipantID: handlerTestUUID(251), StudentNama: "Siswa"}},
		cardsRows:   []db.GetEventExamCardsRow{{EventID: eventID, StudentNama: "Siswa", Token: "ABC123"}},
	}
	h := &CbtEvent{svc: fake}

	rec := httptest.NewRecorder()
	h.List(rec, adminRequest(http.MethodGet, "/api/cbt/events", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("List status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "PAT 2026") {
		t.Fatalf("List body = %s, want event title", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.Get(rec, withRouteParam(adminRequest(http.MethodGet, "/api/cbt/events/"+eventID.String(), ""), "id", eventID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("Get status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.getID != eventID {
		t.Fatalf("Get id = %v, want %v", fake.getID, eventID)
	}

	rec = httptest.NewRecorder()
	h.GetResults(rec, withRouteParam(adminRequest(http.MethodGet, "/api/cbt/events/"+eventID.String()+"/results", ""), "id", eventID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("GetResults status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.resultsID != eventID {
		t.Fatalf("GetResults id = %v, want %v", fake.resultsID, eventID)
	}

	rec = httptest.NewRecorder()
	h.GetExamCards(rec, withRouteParam(adminRequest(http.MethodGet, "/api/cbt/events/"+eventID.String()+"/exam-cards", ""), "id", eventID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("GetExamCards status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.cardsID != eventID {
		t.Fatalf("GetExamCards id = %v, want %v", fake.cardsID, eventID)
	}

	rec = httptest.NewRecorder()
	h.GetResults(rec, withRouteParam(guruRequest(http.MethodGet, "/api/cbt/events/"+eventID.String()+"/results", ""), "id", eventID.String()))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("GetResults(guru) status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.GetExamCards(rec, withRouteParam(guruRequest(http.MethodGet, "/api/cbt/events/"+eventID.String()+"/exam-cards", ""), "id", eventID.String()))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("GetExamCards(guru) status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}
}

func TestCbtEventMutationHandlersForwardPayloads(t *testing.T) {
	eventID := handlerTestUUID(252)
	academicYearID := handlerTestUUID(253)
	fake := &fakeCbtEventService{
		CbtEvent:  &service.CbtEvent{},
		createRow: cbtEventModel(eventID, "PAT 2026"),
		updateRow: cbtEventModel(eventID, "PAT Revisi"),
		statusRow: cbtEventModel(eventID, "PAT Revisi"),
	}
	h := &CbtEvent{svc: fake}

	rec := httptest.NewRecorder()
	h.Create(rec, adminRequest(http.MethodPost, "/api/cbt/events", `{"title":"PAT 2026","target_levels":[" ix ","VII","VII"],"academic_year_id":"`+academicYearID.String()+`"}`))
	if rec.Code != http.StatusCreated {
		t.Fatalf("Create status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if fake.createInput.Title != "PAT 2026" || fake.createInput.ExamType != db.CbtExamTypeLainnya || fake.createInput.Scope != "class" || fake.createInput.AcademicYearID != academicYearID {
		t.Fatalf("Create input = %+v, want defaults and academic year", fake.createInput)
	}
	assertCbtTargetLevels(t, fake.createInput.TargetLevels, "IX", "VII")

	rec = httptest.NewRecorder()
	h.Update(rec, withRouteParam(adminRequest(http.MethodPatch, "/api/cbt/events/"+eventID.String(), `{"title":"PAT Revisi","exam_type":"uts","scope":"grade","target_levels":["VIII"]}`), "id", eventID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("Update status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.updateID != eventID || fake.updateInput.Title != "PAT Revisi" || fake.updateInput.ExamType != db.CbtExamTypeUts || fake.updateInput.Scope != "grade" {
		t.Fatalf("Update input = id %v input %+v, want route id and decoded payload", fake.updateID, fake.updateInput)
	}
	assertCbtTargetLevels(t, fake.updateInput.TargetLevels, "VIII")

	rec = httptest.NewRecorder()
	h.UpdateStatus(rec, withRouteParam(adminRequest(http.MethodPatch, "/api/cbt/events/"+eventID.String()+"/status", `{"status":"active"}`), "id", eventID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateStatus status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.statusID != eventID || fake.statusValue != "active" {
		t.Fatalf("UpdateStatus params = (%v, %q), want (%v, active)", fake.statusID, fake.statusValue, eventID)
	}

	rec = httptest.NewRecorder()
	h.Delete(rec, withRouteParam(adminRequest(http.MethodDelete, "/api/cbt/events/"+eventID.String(), ""), "id", eventID.String()))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("Delete status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}
	if fake.deleteID != eventID {
		t.Fatalf("Delete id = %v, want %v", fake.deleteID, eventID)
	}
}

func TestCbtEventHandlersMapServiceErrors(t *testing.T) {
	eventID := handlerTestUUID(254)
	tests := []struct {
		name       string
		handler    func(*CbtEvent, http.ResponseWriter, *http.Request)
		svc        *fakeCbtEventService
		req        *http.Request
		wantStatus int
	}{
		{
			name:       "list",
			handler:    (*CbtEvent).List,
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}, listErr: errors.New("db down")},
			req:        adminRequest(http.MethodGet, "/api/cbt/events", ""),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "get",
			handler:    (*CbtEvent).Get,
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}, getErr: errors.New("db down")},
			req:        withRouteParam(adminRequest(http.MethodGet, "/api/cbt/events/"+eventID.String(), ""), "id", eventID.String()),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "results",
			handler:    (*CbtEvent).GetResults,
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}, resultsErr: errors.New("db down")},
			req:        withRouteParam(adminRequest(http.MethodGet, "/api/cbt/events/"+eventID.String()+"/results", ""), "id", eventID.String()),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "cards",
			handler:    (*CbtEvent).GetExamCards,
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}, cardsErr: errors.New("db down")},
			req:        withRouteParam(adminRequest(http.MethodGet, "/api/cbt/events/"+eventID.String()+"/exam-cards", ""), "id", eventID.String()),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "create",
			handler:    (*CbtEvent).Create,
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}, createErr: errors.New("db down")},
			req:        adminRequest(http.MethodPost, "/api/cbt/events", `{"title":"PAT"}`),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "update",
			handler:    (*CbtEvent).Update,
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}, updateErr: errors.New("db down")},
			req:        withRouteParam(adminRequest(http.MethodPatch, "/api/cbt/events/"+eventID.String(), `{"title":"PAT"}`), "id", eventID.String()),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "status",
			handler:    (*CbtEvent).UpdateStatus,
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}, statusErr: errors.New("db down")},
			req:        withRouteParam(adminRequest(http.MethodPatch, "/api/cbt/events/"+eventID.String()+"/status", `{"status":"active"}`), "id", eventID.String()),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "delete",
			handler:    (*CbtEvent).Delete,
			svc:        &fakeCbtEventService{CbtEvent: &service.CbtEvent{}, deleteErr: errors.New("db down")},
			req:        withRouteParam(adminRequest(http.MethodDelete, "/api/cbt/events/"+eventID.String(), ""), "id", eventID.String()),
			wantStatus: http.StatusInternalServerError,
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
