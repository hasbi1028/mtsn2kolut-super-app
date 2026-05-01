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

type fakeEmployeeScheduleService struct {
	*service.EmployeeSchedule

	listEmployeeID pgtype.UUID
	listRows       []db.ListEmployeeSchedulesRow
	listErr        error
	upsertArg      db.UpsertEmployeeScheduleParams
	upsertRow      db.EmployeeSchedule
	upsertErr      error
	deleteID       pgtype.UUID
	deleteEmpID    pgtype.UUID
	deleteErr      error
}

func (f *fakeEmployeeScheduleService) List(_ context.Context, employeeID pgtype.UUID) ([]db.ListEmployeeSchedulesRow, error) {
	f.listEmployeeID = employeeID
	return f.listRows, f.listErr
}

func (f *fakeEmployeeScheduleService) Upsert(_ context.Context, p db.UpsertEmployeeScheduleParams) (db.EmployeeSchedule, error) {
	f.upsertArg = p
	return f.upsertRow, f.upsertErr
}

func (f *fakeEmployeeScheduleService) Delete(_ context.Context, id, employeeID pgtype.UUID) error {
	f.deleteID = id
	f.deleteEmpID = employeeID
	return f.deleteErr
}

func TestEmployeeScheduleSuccessHandlers(t *testing.T) {
	employeeID := handlerTestUUID(246)
	scheduleID := handlerTestUUID(247)
	fake := &fakeEmployeeScheduleService{
		EmployeeSchedule: &service.EmployeeSchedule{},
		listRows: []db.ListEmployeeSchedulesRow{
			{ID: scheduleID, EmployeeID: employeeID, RunType: db.RunTypeEnumCheckin, RunTime: "07:00", IsEnabled: true, DayOfWeek: 1},
		},
		upsertRow: db.EmployeeSchedule{ID: scheduleID, EmployeeID: employeeID, RunType: db.RunTypeEnumCheckout, RunTime: "15:00", IsEnabled: true, RandomWindowMinutes: 5, DayOfWeek: 5},
	}
	h := &EmployeeSchedule{svc: fake}

	rec := httptest.NewRecorder()
	h.List(rec, withRouteParam(adminRequest(http.MethodGet, "/api/employees/"+employeeID.String()+"/schedules", ""), "id", employeeID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("List status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.listEmployeeID != employeeID || !strings.Contains(rec.Body.String(), "checkin") {
		t.Fatalf("List forwarded id=%v body=%s, want employee schedules", fake.listEmployeeID, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.Upsert(rec, withRouteParam(adminRequest(http.MethodPut, "/api/employees/"+employeeID.String()+"/schedules", `{"run_type":"checkout","run_time":"15:00","is_enabled":true,"random_window_minutes":5,"day_of_week":5}`), "id", employeeID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("Upsert status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.upsertArg.EmployeeID != employeeID || fake.upsertArg.RunType != db.RunTypeEnumCheckout || fake.upsertArg.RunTime != "15:00" || !fake.upsertArg.IsEnabled || fake.upsertArg.RandomWindowMinutes != 5 || fake.upsertArg.DayOfWeek != 5 {
		t.Fatalf("Upsert params = %+v, want decoded schedule payload", fake.upsertArg)
	}

	rec = httptest.NewRecorder()
	h.Upsert(rec, withRouteParam(adminRequest(http.MethodPut, "/api/employees/"+employeeID.String()+"/schedules", `{"run_type":"checkin","run_time":"07:00","is_enabled":true,"random_window_minutes":-10,"day_of_week":1}`), "id", employeeID.String()))
	if rec.Code != http.StatusOK {
		t.Fatalf("Upsert(negative random window) status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.upsertArg.RandomWindowMinutes != 0 {
		t.Fatalf("Upsert negative random window = %d, want normalized zero", fake.upsertArg.RandomWindowMinutes)
	}

	rec = httptest.NewRecorder()
	req := withRouteParams(adminRequest(http.MethodDelete, "/api/employees/"+employeeID.String()+"/schedules/"+scheduleID.String(), ""), "id", employeeID.String(), "scheduleId", scheduleID.String())
	h.Delete(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Delete status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.deleteID != scheduleID || fake.deleteEmpID != employeeID {
		t.Fatalf("Delete params = (%v, %v), want (%v, %v)", fake.deleteID, fake.deleteEmpID, scheduleID, employeeID)
	}
}

func TestEmployeeScheduleHandlersMapServiceErrors(t *testing.T) {
	employeeID := handlerTestUUID(248)
	scheduleID := handlerTestUUID(249)
	tests := []struct {
		name       string
		handler    func(*EmployeeSchedule, http.ResponseWriter, *http.Request)
		svc        *fakeEmployeeScheduleService
		req        *http.Request
		wantStatus int
	}{
		{
			name:       "list internal",
			handler:    (*EmployeeSchedule).List,
			svc:        &fakeEmployeeScheduleService{EmployeeSchedule: &service.EmployeeSchedule{}, listErr: errors.New("db down")},
			req:        withRouteParam(adminRequest(http.MethodGet, "/api/employees/"+employeeID.String()+"/schedules", ""), "id", employeeID.String()),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "upsert internal",
			handler:    (*EmployeeSchedule).Upsert,
			svc:        &fakeEmployeeScheduleService{EmployeeSchedule: &service.EmployeeSchedule{}, upsertErr: errors.New("db down")},
			req:        withRouteParam(adminRequest(http.MethodPut, "/api/employees/"+employeeID.String()+"/schedules", `{"run_type":"checkin","run_time":"07:00","day_of_week":1}`), "id", employeeID.String()),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "upsert invalid employee id",
			handler:    (*EmployeeSchedule).Upsert,
			svc:        &fakeEmployeeScheduleService{EmployeeSchedule: &service.EmployeeSchedule{}},
			req:        withRouteParam(adminRequest(http.MethodPut, "/api/employees/bad/schedules", `{}`), "id", "bad"),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "upsert invalid json",
			handler:    (*EmployeeSchedule).Upsert,
			svc:        &fakeEmployeeScheduleService{EmployeeSchedule: &service.EmployeeSchedule{}},
			req:        withRouteParam(adminRequest(http.MethodPut, "/api/employees/"+employeeID.String()+"/schedules", `{`), "id", employeeID.String()),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "upsert invalid run type",
			handler:    (*EmployeeSchedule).Upsert,
			svc:        &fakeEmployeeScheduleService{EmployeeSchedule: &service.EmployeeSchedule{}},
			req:        withRouteParam(adminRequest(http.MethodPut, "/api/employees/"+employeeID.String()+"/schedules", `{"run_type":"lunch","run_time":"07:00","day_of_week":1}`), "id", employeeID.String()),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "upsert missing run time",
			handler:    (*EmployeeSchedule).Upsert,
			svc:        &fakeEmployeeScheduleService{EmployeeSchedule: &service.EmployeeSchedule{}},
			req:        withRouteParam(adminRequest(http.MethodPut, "/api/employees/"+employeeID.String()+"/schedules", `{"run_type":"checkin","day_of_week":1}`), "id", employeeID.String()),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "upsert invalid day",
			handler:    (*EmployeeSchedule).Upsert,
			svc:        &fakeEmployeeScheduleService{EmployeeSchedule: &service.EmployeeSchedule{}},
			req:        withRouteParam(adminRequest(http.MethodPut, "/api/employees/"+employeeID.String()+"/schedules", `{"run_type":"checkin","run_time":"07:00","day_of_week":7}`), "id", employeeID.String()),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "delete internal",
			handler:    (*EmployeeSchedule).Delete,
			svc:        &fakeEmployeeScheduleService{EmployeeSchedule: &service.EmployeeSchedule{}, deleteErr: errors.New("db down")},
			req:        withRouteParams(adminRequest(http.MethodDelete, "/api/employees/"+employeeID.String()+"/schedules/"+scheduleID.String(), ""), "id", employeeID.String(), "scheduleId", scheduleID.String()),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "delete invalid employee id",
			handler:    (*EmployeeSchedule).Delete,
			svc:        &fakeEmployeeScheduleService{EmployeeSchedule: &service.EmployeeSchedule{}},
			req:        withRouteParams(adminRequest(http.MethodDelete, "/api/employees/bad/schedules/"+scheduleID.String(), ""), "id", "bad", "scheduleId", scheduleID.String()),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "delete invalid schedule id",
			handler:    (*EmployeeSchedule).Delete,
			svc:        &fakeEmployeeScheduleService{EmployeeSchedule: &service.EmployeeSchedule{}},
			req:        withRouteParams(adminRequest(http.MethodDelete, "/api/employees/"+employeeID.String()+"/schedules/bad", ""), "id", employeeID.String(), "scheduleId", "bad"),
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.handler(&EmployeeSchedule{svc: tt.svc}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
