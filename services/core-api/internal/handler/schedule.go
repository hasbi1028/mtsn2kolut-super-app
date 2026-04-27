package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type Schedule struct {
	svc *service.Schedule
}

func NewSchedule(svc *service.Schedule) *Schedule { return &Schedule{svc: svc} }

func (h *Schedule) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.List(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *Schedule) Upsert(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	_ = id // UpsertSchedule conflicts on run_type, not id

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
	sched, err := h.svc.Upsert(r.Context(), db.UpsertScheduleParams{
		Label:     body.Label,
		RunTime:   body.RunTime,
		RunType:   db.RunTypeEnum(body.RunType),
		IsEnabled: body.IsEnabled,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		api.NotFound(w)
		return
	}
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, sched)
}
