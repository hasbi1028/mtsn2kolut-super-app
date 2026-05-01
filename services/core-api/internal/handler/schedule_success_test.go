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
)

type fakePusakaScheduleService struct {
	listRows []db.Schedule
	listErr  error

	createArg db.CreateScheduleParams
	createErr error
	updateArg db.UpdateScheduleByIDParams
	updateErr error
	deleteID  pgtype.UUID
	deleteErr error
}

func (f *fakePusakaScheduleService) List(ctx context.Context) ([]db.Schedule, error) {
	return f.listRows, f.listErr
}

func (f *fakePusakaScheduleService) Create(ctx context.Context, p db.CreateScheduleParams) (db.Schedule, error) {
	f.createArg = p
	if f.createErr != nil {
		return db.Schedule{}, f.createErr
	}
	return db.Schedule{ID: handlerTestUUID(190), Label: p.Label, RunTime: p.RunTime, RunType: p.RunType, IsEnabled: p.IsEnabled}, nil
}

func (f *fakePusakaScheduleService) UpdateByID(ctx context.Context, p db.UpdateScheduleByIDParams) (db.Schedule, error) {
	f.updateArg = p
	if f.updateErr != nil {
		return db.Schedule{}, f.updateErr
	}
	return db.Schedule{ID: p.ID, Label: p.Label, RunTime: p.RunTime, IsEnabled: p.IsEnabled}, nil
}

func (f *fakePusakaScheduleService) DeleteByID(ctx context.Context, id pgtype.UUID) error {
	f.deleteID = id
	return f.deleteErr
}

func TestPusakaScheduleSuccessHandlersForwardPayloads(t *testing.T) {
	id := handlerTestUUID(191)
	svc := &fakePusakaScheduleService{
		listRows: []db.Schedule{{ID: id, Label: "Pagi", RunTime: "07:00", RunType: db.RunTypeEnumMorning, IsEnabled: true}},
	}
	h := &PusakaSchedule{svc: svc}

	rec := httptest.NewRecorder()
	h.List(rec, adminRequest(http.MethodGet, "/api/pusaka/schedules", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("List() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.Create(rec, adminRequest(http.MethodPost, "/api/pusaka/schedules", `{"label":"Pagi","run_time":"07:00","run_type":"morning","is_enabled":true}`))
	if rec.Code != http.StatusCreated || svc.createArg.Label != "Pagi" || svc.createArg.RunTime != "07:00" || svc.createArg.RunType != db.RunTypeEnumMorning || !svc.createArg.IsEnabled {
		t.Fatalf("Create() status/arg = %d/%+v", rec.Code, svc.createArg)
	}

	rec = httptest.NewRecorder()
	req := withRouteParam(adminRequest(http.MethodPut, "/api/pusaka/schedules/"+id.String(), `{"label":"Pagi Baru","run_time":"07:15","is_enabled":false}`), "id", id.String())
	h.Update(rec, req)
	if rec.Code != http.StatusOK || svc.updateArg.ID != id || svc.updateArg.Label != "Pagi Baru" || svc.updateArg.RunTime != "07:15" || svc.updateArg.IsEnabled {
		t.Fatalf("Update() status/arg = %d/%+v", rec.Code, svc.updateArg)
	}

	rec = httptest.NewRecorder()
	req = withRouteParam(adminRequest(http.MethodDelete, "/api/pusaka/schedules/"+id.String(), ""), "id", id.String())
	h.Delete(rec, req)
	if rec.Code != http.StatusOK || svc.deleteID != id {
		t.Fatalf("Delete() status/id = %d/%v", rec.Code, svc.deleteID)
	}
}

func TestPusakaScheduleValidationAndServiceErrors(t *testing.T) {
	id := handlerTestUUID(192)
	tests := []struct {
		name string
		run  func(*PusakaSchedule) *httptest.ResponseRecorder
		want int
	}{
		{
			name: "list forbidden",
			run: func(h *PusakaSchedule) *httptest.ResponseRecorder {
				rec := httptest.NewRecorder()
				h.List(rec, httptest.NewRequest(http.MethodGet, "/api/pusaka/schedules", nil))
				return rec
			},
			want: http.StatusForbidden,
		},
		{
			name: "list service error",
			run: func(h *PusakaSchedule) *httptest.ResponseRecorder {
				h.svc.(*fakePusakaScheduleService).listErr = errors.New("list failed")
				rec := httptest.NewRecorder()
				h.List(rec, adminRequest(http.MethodGet, "/api/pusaka/schedules", ""))
				return rec
			},
			want: http.StatusInternalServerError,
		},
		{
			name: "create invalid json",
			run: func(h *PusakaSchedule) *httptest.ResponseRecorder {
				rec := httptest.NewRecorder()
				h.Create(rec, adminRequest(http.MethodPost, "/api/pusaka/schedules", `{bad`))
				return rec
			},
			want: http.StatusBadRequest,
		},
		{
			name: "create rejects non morning run type",
			run: func(h *PusakaSchedule) *httptest.ResponseRecorder {
				rec := httptest.NewRecorder()
				h.Create(rec, adminRequest(http.MethodPost, "/api/pusaka/schedules", `{"label":"Pulang","run_time":"15:00","run_type":"checkout"}`))
				return rec
			},
			want: http.StatusBadRequest,
		},
		{
			name: "create requires run time",
			run: func(h *PusakaSchedule) *httptest.ResponseRecorder {
				rec := httptest.NewRecorder()
				h.Create(rec, adminRequest(http.MethodPost, "/api/pusaka/schedules", `{"label":"Pagi","run_type":"morning"}`))
				return rec
			},
			want: http.StatusBadRequest,
		},
		{
			name: "create service error",
			run: func(h *PusakaSchedule) *httptest.ResponseRecorder {
				h.svc.(*fakePusakaScheduleService).createErr = errors.New("create failed")
				rec := httptest.NewRecorder()
				h.Create(rec, adminRequest(http.MethodPost, "/api/pusaka/schedules", `{"label":"Pagi","run_time":"07:00","run_type":"morning"}`))
				return rec
			},
			want: http.StatusInternalServerError,
		},
		{
			name: "update invalid id",
			run: func(h *PusakaSchedule) *httptest.ResponseRecorder {
				rec := httptest.NewRecorder()
				req := withRouteParam(adminRequest(http.MethodPut, "/api/pusaka/schedules/bad", `{}`), "id", "bad")
				h.Update(rec, req)
				return rec
			},
			want: http.StatusBadRequest,
		},
		{
			name: "update invalid json",
			run: func(h *PusakaSchedule) *httptest.ResponseRecorder {
				rec := httptest.NewRecorder()
				req := withRouteParam(adminRequest(http.MethodPut, "/api/pusaka/schedules/"+id.String(), `{bad`), "id", id.String())
				h.Update(rec, req)
				return rec
			},
			want: http.StatusBadRequest,
		},
		{
			name: "update service error",
			run: func(h *PusakaSchedule) *httptest.ResponseRecorder {
				h.svc.(*fakePusakaScheduleService).updateErr = errors.New("update failed")
				rec := httptest.NewRecorder()
				req := withRouteParam(adminRequest(http.MethodPut, "/api/pusaka/schedules/"+id.String(), `{"label":"Pagi","run_time":"07:00"}`), "id", id.String())
				h.Update(rec, req)
				return rec
			},
			want: http.StatusInternalServerError,
		},
		{
			name: "delete invalid id",
			run: func(h *PusakaSchedule) *httptest.ResponseRecorder {
				rec := httptest.NewRecorder()
				req := withRouteParam(adminRequest(http.MethodDelete, "/api/pusaka/schedules/bad", ""), "id", "bad")
				h.Delete(rec, req)
				return rec
			},
			want: http.StatusBadRequest,
		},
		{
			name: "delete service error",
			run: func(h *PusakaSchedule) *httptest.ResponseRecorder {
				h.svc.(*fakePusakaScheduleService).deleteErr = errors.New("delete failed")
				rec := httptest.NewRecorder()
				req := withRouteParam(adminRequest(http.MethodDelete, "/api/pusaka/schedules/"+id.String(), ""), "id", id.String())
				h.Delete(rec, req)
				return rec
			},
			want: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &PusakaSchedule{svc: &fakePusakaScheduleService{}}
			rec := tt.run(h)
			if rec.Code != tt.want {
				t.Fatalf("%s status = %d, want %d; body=%s", tt.name, rec.Code, tt.want, strings.TrimSpace(rec.Body.String()))
			}
		})
	}
}
