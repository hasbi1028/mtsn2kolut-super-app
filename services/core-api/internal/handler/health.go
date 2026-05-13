package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type healthDB interface {
	Ping(ctx context.Context) error
}

type healthClockDB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
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
	serverNow := time.Now()
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
		"clock": h.clockHealth(r.Context(), serverNow),
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

func (h *Health) clockHealth(ctx context.Context, serverNow time.Time) map[string]any {
	clock := map[string]any{
		"status":       "ok",
		"server_local": serverNow.Format(time.RFC3339),
		"server_utc":   serverNow.UTC().Format(time.RFC3339),
	}
	clockDB, ok := h.pool.(healthClockDB)
	if !ok {
		return clock
	}

	var dbNow time.Time
	var dbTimezone string
	if err := clockDB.QueryRow(ctx, "SELECT now(), current_setting('TimeZone')").Scan(&dbNow, &dbTimezone); err != nil {
		clock["status"] = "degraded"
		clock["db_error"] = "clock query failed"
		return clock
	}
	drift := dbNow.Sub(serverNow)
	if drift < 0 {
		drift = -drift
	}
	clock["db_now"] = dbNow.Format(time.RFC3339)
	clock["db_timezone"] = dbTimezone
	clock["db_server_drift_ms"] = drift.Milliseconds()
	if drift > 2*time.Minute {
		clock["status"] = "degraded"
	}
	return clock
}
