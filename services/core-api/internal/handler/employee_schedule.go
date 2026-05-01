package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type EmployeeSchedule struct {
	svc employeeScheduleService
}

type employeeScheduleService interface {
	List(ctx context.Context, employeeID pgtype.UUID) ([]db.ListEmployeeSchedulesRow, error)
	Upsert(ctx context.Context, p db.UpsertEmployeeScheduleParams) (db.EmployeeSchedule, error)
	Delete(ctx context.Context, id, employeeID pgtype.UUID) error
}

func NewEmployeeSchedule(svc *service.EmployeeSchedule) *EmployeeSchedule {
	return &EmployeeSchedule{svc: svc}
}

func (h *EmployeeSchedule) List(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	empID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid employee id")
		return
	}
	rows, err := h.svc.List(r.Context(), empID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *EmployeeSchedule) Upsert(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	empID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid employee id")
		return
	}
	var body struct {
		RunType             string `json:"run_type"`
		RunTime             string `json:"run_time"`
		IsEnabled           bool   `json:"is_enabled"`
		RandomWindowMinutes int16  `json:"random_window_minutes"`
		DayOfWeek           int16  `json:"day_of_week"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	if body.RunType != "checkin" && body.RunType != "checkout" {
		api.BadRequest(w, "run_type must be checkin or checkout")
		return
	}
	if body.RunTime == "" {
		api.BadRequest(w, "run_time is required")
		return
	}
	if body.DayOfWeek < 0 || body.DayOfWeek > 6 {
		api.BadRequest(w, "day_of_week must be 0 (Minggu) through 6 (Sabtu)")
		return
	}
	if body.RandomWindowMinutes < 0 {
		body.RandomWindowMinutes = 0
	}
	sched, err := h.svc.Upsert(r.Context(), db.UpsertEmployeeScheduleParams{
		EmployeeID:          empID,
		RunType:             db.RunTypeEnum(body.RunType),
		RunTime:             body.RunTime,
		IsEnabled:           body.IsEnabled,
		RandomWindowMinutes: body.RandomWindowMinutes,
		DayOfWeek:           body.DayOfWeek,
	})
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, sched)
}

func (h *EmployeeSchedule) Delete(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	empID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid employee id")
		return
	}
	schedID, err := parseUUID(chi.URLParam(r, "scheduleId"))
	if err != nil {
		api.BadRequest(w, "invalid schedule id")
		return
	}
	if err := h.svc.Delete(r.Context(), schedID, empID); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]bool{"deleted": true})
}
