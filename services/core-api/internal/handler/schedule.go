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

type pusakaScheduleService interface {
	List(ctx context.Context) ([]db.Schedule, error)
	Create(ctx context.Context, p db.CreateScheduleParams) (db.Schedule, error)
	UpdateByID(ctx context.Context, p db.UpdateScheduleByIDParams) (db.Schedule, error)
	DeleteByID(ctx context.Context, id pgtype.UUID) error
}

type PusakaSchedule struct {
	svc pusakaScheduleService
}

func NewPusakaSchedule(svc *service.PusakaSchedule) *PusakaSchedule { return &PusakaSchedule{svc: svc} }

func (h *PusakaSchedule) List(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	rows, err := h.svc.List(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *PusakaSchedule) Create(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body struct {
		Label     string `json:"label"`
		RunTime   string `json:"run_time"`
		RunType   string `json:"run_type"`
		IsEnabled bool   `json:"is_enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	if body.RunType != "morning" {
		api.BadRequest(w, "only morning schedules can be created via this endpoint")
		return
	}
	if body.RunTime == "" {
		api.BadRequest(w, "run_time is required")
		return
	}
	sched, err := h.svc.Create(r.Context(), db.CreateScheduleParams{
		Label:     body.Label,
		RunTime:   body.RunTime,
		RunType:   db.RunTypeEnum(body.RunType),
		IsEnabled: body.IsEnabled,
	})
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.Created(w, sched)
}

func (h *PusakaSchedule) Update(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body struct {
		Label     string `json:"label"`
		RunTime   string `json:"run_time"`
		IsEnabled bool   `json:"is_enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	sched, err := h.svc.UpdateByID(r.Context(), db.UpdateScheduleByIDParams{
		ID:        id,
		Label:     body.Label,
		RunTime:   body.RunTime,
		IsEnabled: body.IsEnabled,
	})
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, sched)
}

func (h *PusakaSchedule) Delete(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	if err := h.svc.DeleteByID(r.Context(), id); err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]bool{"deleted": true})
}
