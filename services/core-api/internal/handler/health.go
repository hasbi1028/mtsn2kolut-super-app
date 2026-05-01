package handler

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type healthDB interface {
	Ping(ctx context.Context) error
}

type Health struct {
	pool healthDB
	jobs *service.PusakaJob
	sett *service.Setting
}

func NewHealth(pool *pgxpool.Pool, jobs *service.PusakaJob, sett *service.Setting) *Health {
	return &Health{pool: pool, jobs: jobs, sett: sett}
}

func (h *Health) Get(w http.ResponseWriter, r *http.Request) {
	if err := h.pool.Ping(r.Context()); err != nil {
		api.JSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "error",
			"db":     "unreachable",
		})
		return
	}

	api.JSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"db":     "connected",
		"server": map[string]string{
			"version": "1.0.0-sprint4",
		},
	})
}
