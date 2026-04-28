package handler

import (
	"net/http"
	"runtime"
	"strings"
	"time"

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

	poolStat := h.pool.Stat()

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
	var schedulerLastTick string
	for _, row := range settings {
		switch {
		case strings.HasPrefix(row.Key, "scheduler_"):
			scheduler[row.Key] = row.Value
			if row.Key == "scheduler_last_tick_at" {
				schedulerLastTick = row.Value
			}
		case strings.HasPrefix(row.Key, "worker_status:"):
			workers[row.Key] = row.Value
		}
	}

	lastTick := any(nil)
	if schedulerLastTick != "" {
		if t, parseErr := time.Parse(time.RFC3339, schedulerLastTick); parseErr == nil {
			lastTick = t.Format(time.RFC3339)
		}
	}

	api.JSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"db":     "connected",
		"db_pool": map[string]int32{
			"total_conns":    poolStat.TotalConns(),
			"idle_conns":     poolStat.IdleConns(),
			"acquired_conns": poolStat.AcquiredConns(),
		},
		"server": map[string]string{
			"version":    "1.0.0-sprint4",
			"go_version": runtime.Version(),
		},
		"queue": map[string]int64{
			"queued":  stats.Queued,
			"running": stats.Running,
			"success": stats.Success,
			"failed":  stats.Failed,
		},
		"scheduler": map[string]any{
			"settings":  scheduler,
			"last_tick": lastTick,
		},
		"workers": workers,
	})
}
