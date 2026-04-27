package handler

import (
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type Health struct {
	pool *pgxpool.Pool
	jobs *service.Job
	sett *service.Setting
}

func NewHealth(pool *pgxpool.Pool, jobs *service.Job, sett *service.Setting) *Health {
	return &Health{pool: pool, jobs: jobs, sett: sett}
}

func (h *Health) Get(w http.ResponseWriter, r *http.Request) {
	if err := h.pool.Ping(r.Context()); err != nil {
		api.JSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "error",
			"db":     "unreachable",
			"error":  err.Error(),
		})
		return
	}

	stats, err := h.jobs.Stats(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	settings, err := h.sett.List(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}

	scheduler := map[string]string{}
	workers := map[string]string{}
	for _, row := range settings {
		switch {
		case strings.HasPrefix(row.Key, "scheduler_"):
			scheduler[row.Key] = row.Value
		case strings.HasPrefix(row.Key, "worker_status:"):
			workers[row.Key] = row.Value
		}
	}

	api.JSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"db":     "connected",
		"queue": map[string]int64{
			"queued":  stats.Queued,
			"running": stats.Running,
			"success": stats.Success,
			"failed":  stats.Failed,
		},
		"scheduler": scheduler,
		"workers":   workers,
	})
}
