package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type healthDB interface {
	Ping(ctx context.Context) error
}

type healthAnalytics interface {
	ExpiredBacklog(ctx context.Context) (service.InternalAnalyticsExpiredBacklog, error)
}

type Health struct {
	pool      healthDB
	jobs      *service.PusakaJob
	sett      *service.Setting
	analytics healthAnalytics
}

func NewHealth(pool *pgxpool.Pool, jobs *service.PusakaJob, sett *service.Setting, analytics ...healthAnalytics) *Health {
	h := &Health{pool: pool, jobs: jobs, sett: sett}
	if len(analytics) > 0 {
		h.analytics = analytics[0]
	}
	return h
}

func (h *Health) Get(w http.ResponseWriter, r *http.Request) {
	if err := h.pool.Ping(r.Context()); err != nil {
		api.JSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "error",
			"db":     "unreachable",
		})
		return
	}

	payload := map[string]any{
		"status": "ok",
		"db":     "connected",
		"server": map[string]string{
			"version": "1.0.0-sprint4",
		},
	}
	if h.analytics != nil {
		if backlog, err := h.analytics.ExpiredBacklog(r.Context()); err != nil {
			payload["analytics"] = map[string]any{
				"status": "degraded",
			}
		} else {
			analytics := map[string]any{
				"status":                      "ok",
				"expired_event_backlog_count": backlog.ExpiredEventBacklogCount,
			}
			if backlog.OldestExpiredEventAt != nil {
				analytics["oldest_expired_event_at"] = backlog.OldestExpiredEventAt.UTC().Format(time.RFC3339)
			}
			payload["analytics"] = analytics
		}
	}
	api.JSON(w, http.StatusOK, payload)
}
