package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

func guruRequest(method, target, body string) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	return withClaims(req, jwt.MapClaims{
		"roles": []any{"guru"},
		"role":  "guru",
		"eid":   "02000000-0000-0000-0000-000000000000",
	})
}

func TestCbtPackageRejectsInvalidRequests(t *testing.T) {
	h := &CbtPackage{}
	tests := []struct {
		name       string
		fn         http.HandlerFunc
		method     string
		target     string
		body       string
		id         string
		wantStatus int
	}{
		{name: "create invalid json", fn: h.Create, method: http.MethodPost, target: "/cbt/packages", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "create invalid subject", fn: h.Create, method: http.MethodPost, target: "/cbt/packages", body: `{"subject_id":"bad"}`, wantStatus: http.StatusBadRequest},
		{name: "create invalid question", fn: h.Create, method: http.MethodPost, target: "/cbt/packages", body: `{"subject_id":"01000000-0000-0000-0000-000000000000","question_ids":["bad"]}`, wantStatus: http.StatusBadRequest},
		{name: "delete invalid id", fn: h.Delete, method: http.MethodDelete, target: "/cbt/packages/bad", id: "bad", wantStatus: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := adminRequest(tt.method, tt.target, tt.body)
			if tt.id != "" {
				req = withRouteParam(req, "id", tt.id)
			}
			tt.fn(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("%s status = %d, want %d; body=%s", tt.name, rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

type fakeCbtPackageHandlerService struct {
	listPackages  []db.ListCbtPackagesRow
	listQuestions []db.ListCbtPackageQuestionsRow
	listEventID   pgtype.UUID
	listErr       error
	createInput   service.CreateCbtPackageInput
	createRow     db.CbtPackage
	createErr     error
	deleteID      pgtype.UUID
	deleteErr     error
}

func (f *fakeCbtPackageHandlerService) List(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtPackagesRow, []db.ListCbtPackageQuestionsRow, error) {
	f.listEventID = eventID
	if f.listErr != nil {
		return nil, nil, f.listErr
	}
	return f.listPackages, f.listQuestions, nil
}

func (f *fakeCbtPackageHandlerService) Create(ctx context.Context, input service.CreateCbtPackageInput) (db.CbtPackage, error) {
	f.createInput = input
	if f.createErr != nil {
		return db.CbtPackage{}, f.createErr
	}
	return f.createRow, nil
}

func (f *fakeCbtPackageHandlerService) Delete(ctx context.Context, id pgtype.UUID) error {
	f.deleteID = id
	return f.deleteErr
}

func TestCbtPackageHandlersForwardValidRequests(t *testing.T) {
	packageID := handlerTestUUID(131)
	eventID := handlerTestUUID(130)
	subjectID := handlerTestUUID(132)
	firstQuestionID := handlerTestUUID(133)
	secondQuestionID := handlerTestUUID(134)
	svc := &fakeCbtPackageHandlerService{
		listPackages: []db.ListCbtPackagesRow{
			{ID: packageID, SubjectID: subjectID, Title: "Paket 1", QuestionCount: 2},
		},
		listQuestions: []db.ListCbtPackageQuestionsRow{
			{PackageID: packageID, QuestionID: firstQuestionID, Position: 1},
		},
		createRow: db.CbtPackage{ID: packageID, SubjectID: subjectID, Title: "Paket 1"},
	}
	audit := &fakeCbtSessionAuditWriter{}
	h := &CbtPackage{svc: svc, audit: audit}

	rec := httptest.NewRecorder()
	h.List(rec, adminRequest(http.MethodGet, "/cbt/packages?event_id="+eventID.String(), ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("List() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if svc.listEventID != eventID {
		t.Fatalf("List() event_id = %v, want %v", svc.listEventID, eventID)
	}

	rec = httptest.NewRecorder()
	body := `{
		"event_id":"` + eventID.String() + `",
		"subject_id":"` + subjectID.String() + `",
		"title":"Paket 1",
		"description":"Latihan",
		"duration_minutes":90,
		"randomize_questions":true,
		"is_active":true,
		"question_ids":["` + firstQuestionID.String() + `","` + secondQuestionID.String() + `"]
	}`
	h.Create(rec, adminRequest(http.MethodPost, "/cbt/packages", body))
	if rec.Code != http.StatusCreated {
		t.Fatalf("Create() status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if svc.createInput.EventID != eventID || svc.createInput.SubjectID != subjectID || svc.createInput.Title != "Paket 1" || svc.createInput.DurationMinutes != 90 || !svc.createInput.RandomizeQuestions || !svc.createInput.IsActive {
		t.Fatalf("Create() input = %+v, want parsed package fields", svc.createInput)
	}
	if len(svc.createInput.QuestionIDs) != 2 || svc.createInput.QuestionIDs[0] != firstQuestionID || svc.createInput.QuestionIDs[1] != secondQuestionID {
		t.Fatalf("Create() question ids = %+v, want parsed question ids", svc.createInput.QuestionIDs)
	}
	if len(audit.entries) != 1 || audit.entries[0].Action != "CBT_PACKAGE_CREATE" {
		t.Fatalf("Create audit = %#v, want one package create audit", audit.entries)
	}

	rec = httptest.NewRecorder()
	req := withRouteParam(adminRequest(http.MethodDelete, "/cbt/packages/"+packageID.String(), ""), "id", packageID.String())
	h.Delete(rec, req)
	if rec.Code != http.StatusNoContent || svc.deleteID != packageID {
		t.Fatalf("Delete() status/id = %d/%v, want 204/%v", rec.Code, svc.deleteID, packageID)
	}
	if len(audit.entries) != 2 || audit.entries[1].Action != "CBT_PACKAGE_DELETE" {
		t.Fatalf("Delete audit = %#v, want package delete audit", audit.entries)
	}
}

func TestCbtPackageHandlersMapServiceErrors(t *testing.T) {
	packageID := handlerTestUUID(135)
	subjectID := handlerTestUUID(136)
	errDB := errors.New("db down")
	tests := []struct {
		name       string
		handler    func(*CbtPackage, http.ResponseWriter, *http.Request)
		svc        *fakeCbtPackageHandlerService
		req        *http.Request
		wantStatus int
	}{
		{
			name:       "list",
			handler:    (*CbtPackage).List,
			svc:        &fakeCbtPackageHandlerService{listErr: errDB},
			req:        adminRequest(http.MethodGet, "/cbt/packages", ""),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "create",
			handler:    (*CbtPackage).Create,
			svc:        &fakeCbtPackageHandlerService{createErr: errDB},
			req:        adminRequest(http.MethodPost, "/cbt/packages", `{"subject_id":"`+subjectID.String()+`","title":"Paket"}`),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "delete",
			handler:    (*CbtPackage).Delete,
			svc:        &fakeCbtPackageHandlerService{deleteErr: errDB},
			req:        withRouteParam(adminRequest(http.MethodDelete, "/cbt/packages/"+packageID.String(), ""), "id", packageID.String()),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.handler(&CbtPackage{svc: tt.svc}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("%s status = %d, want %d; body=%s", tt.name, rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestCbtPackageTeacherMutationsForbidden(t *testing.T) {
	h := &CbtPackage{svc: &fakeCbtPackageHandlerService{}}
	subjectID := handlerTestUUID(137)
	packageID := handlerTestUUID(138)

	rec := httptest.NewRecorder()
	h.Create(rec, guruRequest(http.MethodPost, "/cbt/packages", `{"subject_id":"`+subjectID.String()+`","title":"Paket Guru","question_ids":["`+handlerTestUUID(139).String()+`"]}`))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("Create(guru) status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req := withRouteParam(guruRequest(http.MethodDelete, "/cbt/packages/"+packageID.String(), ""), "id", packageID.String())
	h.Delete(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("Delete(guru) status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}
}

func TestEmployeeScheduleRejectsInvalidAdminRequests(t *testing.T) {
	h := &EmployeeSchedule{}
	validID := handlerTestUUID(111).String()
	tests := []struct {
		name       string
		fn         http.HandlerFunc
		method     string
		target     string
		body       string
		id         string
		scheduleID string
		wantStatus int
	}{
		{name: "list invalid employee", fn: h.List, method: http.MethodGet, target: "/employees/bad/schedules", id: "bad", wantStatus: http.StatusBadRequest},
		{name: "upsert invalid employee", fn: h.Upsert, method: http.MethodPost, target: "/employees/bad/schedules", body: `{}`, id: "bad", wantStatus: http.StatusBadRequest},
		{name: "upsert invalid json", fn: h.Upsert, method: http.MethodPost, target: "/employees/" + validID + "/schedules", body: `{`, id: validID, wantStatus: http.StatusBadRequest},
		{name: "upsert invalid run type", fn: h.Upsert, method: http.MethodPost, target: "/employees/" + validID + "/schedules", body: `{"run_type":"morning","run_time":"07:00","day_of_week":1}`, id: validID, wantStatus: http.StatusBadRequest},
		{name: "upsert missing time", fn: h.Upsert, method: http.MethodPost, target: "/employees/" + validID + "/schedules", body: `{"run_type":"checkin","day_of_week":1}`, id: validID, wantStatus: http.StatusBadRequest},
		{name: "upsert invalid day", fn: h.Upsert, method: http.MethodPost, target: "/employees/" + validID + "/schedules", body: `{"run_type":"checkin","run_time":"07:00","day_of_week":7}`, id: validID, wantStatus: http.StatusBadRequest},
		{name: "delete invalid employee", fn: h.Delete, method: http.MethodDelete, target: "/employees/bad/schedules/" + validID, id: "bad", scheduleID: validID, wantStatus: http.StatusBadRequest},
		{name: "delete invalid schedule", fn: h.Delete, method: http.MethodDelete, target: "/employees/" + validID + "/schedules/bad", id: validID, scheduleID: "bad", wantStatus: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := adminRequest(tt.method, tt.target, tt.body)
			if tt.id != "" || tt.scheduleID != "" {
				req = withRouteParams(req, "id", tt.id, "scheduleId", tt.scheduleID)
			}
			tt.fn(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("%s status = %d, want %d; body=%s", tt.name, rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestClassJournalRejectsInvalidRequests(t *testing.T) {
	h := &ClassJournal{}
	validID := handlerTestUUID(121).String()
	tests := []struct {
		name       string
		fn         http.HandlerFunc
		method     string
		target     string
		body       string
		id         string
		wantStatus int
	}{
		{name: "overview invalid assignment", fn: h.Overview, method: http.MethodGet, target: "/class-journal?assignment_id=bad", wantStatus: http.StatusBadRequest},
		{name: "create invalid json", fn: h.CreateSession, method: http.MethodPost, target: "/class-journal/sessions", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "create invalid assignment", fn: h.CreateSession, method: http.MethodPost, target: "/class-journal/sessions", body: `{"assignment_id":"bad","tanggal":"2026-05-01"}`, wantStatus: http.StatusBadRequest},
		{name: "create missing date", fn: h.CreateSession, method: http.MethodPost, target: "/class-journal/sessions", body: `{"assignment_id":"` + validID + `"}`, wantStatus: http.StatusBadRequest},
		{name: "create invalid date", fn: h.CreateSession, method: http.MethodPost, target: "/class-journal/sessions", body: `{"assignment_id":"` + validID + `","tanggal":"bad"}`, wantStatus: http.StatusBadRequest},
		{name: "get invalid id", fn: h.GetSession, method: http.MethodGet, target: "/class-journal/sessions/bad", id: "bad", wantStatus: http.StatusBadRequest},
		{name: "update invalid id", fn: h.UpdateSession, method: http.MethodPut, target: "/class-journal/sessions/bad", body: `{}`, id: "bad", wantStatus: http.StatusBadRequest},
		{name: "update invalid json", fn: h.UpdateSession, method: http.MethodPut, target: "/class-journal/sessions/" + validID, body: `{`, id: validID, wantStatus: http.StatusBadRequest},
		{name: "delete invalid id", fn: h.DeleteSession, method: http.MethodDelete, target: "/class-journal/sessions/bad", id: "bad", wantStatus: http.StatusBadRequest},
		{name: "bulk invalid id", fn: h.BulkUpsertAttendances, method: http.MethodPost, target: "/class-journal/sessions/bad/attendances", body: `{}`, id: "bad", wantStatus: http.StatusBadRequest},
		{name: "bulk invalid json", fn: h.BulkUpsertAttendances, method: http.MethodPost, target: "/class-journal/sessions/" + validID + "/attendances", body: `{`, id: validID, wantStatus: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := adminRequest(tt.method, tt.target, tt.body)
			if tt.name == "delete invalid id" {
				req = adminRequest(tt.method, tt.target, tt.body)
			}
			if tt.id != "" {
				req = withRouteParam(req, "id", tt.id)
			}
			tt.fn(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("%s status = %d, want %d; body=%s", tt.name, rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}

	req := guruRequest(http.MethodGet, "/class-journal", "")
	got := journalEmployeeID(req)
	if !got.Valid || got.String() != "02000000-0000-0000-0000-000000000000" {
		t.Fatalf("journalEmployeeID(guru) = %v, want guru employee id", got)
	}
	if got := journalEmployeeID(adminRequest(http.MethodGet, "/class-journal", "")); got.Valid {
		t.Fatalf("journalEmployeeID(admin) = %v, want invalid", got)
	}
}
