package service

import (
	"context"
	"log/slog"
	"time"
)

const (
	defaultInternalAnalyticsRollupInterval = 10 * time.Minute
	minInternalAnalyticsRollupInterval     = 5 * time.Minute
)

type InternalAnalyticsRollupLoopConfig struct {
	Interval       time.Duration
	Lookback       time.Duration
	Limit          int32
	RunImmediately bool
}

func (s *InternalAnalytics) StartRollupLoop(ctx context.Context, cfg InternalAnalyticsRollupLoopConfig) context.CancelFunc {
	loopCtx, cancel := context.WithCancel(ctx)
	normalized := normalizeInternalAnalyticsRollupLoopConfig(cfg)
	go func() {
		ticker := time.NewTicker(normalized.Interval)
		defer ticker.Stop()

		if normalized.RunImmediately {
			s.runRollupLoopOnce(loopCtx, normalized, time.Now())
		}
		for {
			select {
			case <-loopCtx.Done():
				return
			case now := <-ticker.C:
				s.runRollupLoopOnce(loopCtx, normalized, now)
			}
		}
	}()
	return cancel
}

func (s *InternalAnalytics) runRollupLoopOnce(ctx context.Context, cfg InternalAnalyticsRollupLoopConfig, now time.Time) {
	now = now.UTC()
	startAt := internalAnalyticsDateOnly(now.Add(-cfg.Lookback))
	result, err := s.RollupAndCleanup(ctx, InternalAnalyticsRollupCleanupInput{
		StartAt:  startAt,
		EndAt:    now,
		CutoffAt: now,
		Limit:    cfg.Limit,
	})
	if err != nil {
		slog.Error("internal analytics rollup failed", "error", err)
		return
	}
	if result.EventsScanned > 0 || result.ExpiredEventsDeleted > 0 || result.LimitReached {
		slog.Info("internal analytics rollup",
			"events_scanned", result.EventsScanned,
			"aggregates_upserted", result.AggregatesUpserted,
			"expired_events_deleted", result.ExpiredEventsDeleted,
			"limit_reached", result.LimitReached,
		)
	}
}

func normalizeInternalAnalyticsRollupLoopConfig(cfg InternalAnalyticsRollupLoopConfig) InternalAnalyticsRollupLoopConfig {
	if cfg.Interval <= 0 {
		cfg.Interval = defaultInternalAnalyticsRollupInterval
	}
	if cfg.Interval < minInternalAnalyticsRollupInterval {
		cfg.Interval = minInternalAnalyticsRollupInterval
	}
	if cfg.Lookback <= 0 {
		cfg.Lookback = 48 * time.Hour
	}
	if cfg.Lookback < 24*time.Hour {
		cfg.Lookback = 24 * time.Hour
	}
	if cfg.Lookback > 30*24*time.Hour {
		cfg.Lookback = 30 * 24 * time.Hour
	}
	if cfg.Limit <= 0 {
		cfg.Limit = 10000
	}
	if cfg.Limit > 10000 {
		cfg.Limit = 10000
	}
	return cfg
}
