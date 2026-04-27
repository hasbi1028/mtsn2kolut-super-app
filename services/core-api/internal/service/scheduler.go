package service

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

const (
	defaultSchedulerInterval = 30 * time.Second
	defaultMaxAttempts       = 3
)

type schedulerStore interface {
	ClaimDueSchedules(ctx context.Context, arg db.ClaimDueSchedulesParams) ([]db.ClaimDueSchedulesRow, error)
	ResetScheduleEnqueueState(ctx context.Context, arg db.ResetScheduleEnqueueStateParams) error
	GetSetting(ctx context.Context, key string) (db.AppSetting, error)
}

type schedulerJobRunner interface {
	RunAll(ctx context.Context, runType string, maxAttempts int32) (inserted, skipped int, err error)
}

type Scheduler struct {
	store schedulerStore
	jobs  schedulerJobRunner
	sett  *Setting
	loc   *time.Location
}

type SchedulerResult struct {
	CheckedAt string `json:"checked_at"`
	Processed int    `json:"processed"`
	Enqueued  int    `json:"enqueued"`
	Skipped   int    `json:"skipped"`
}

func NewScheduler(store *db.Queries, jobs *Job, sett *Setting) *Scheduler {
	loc, err := time.LoadLocation("Asia/Makassar")
	if err != nil {
		loc = time.FixedZone("WITA", 8*60*60)
	}
	return &Scheduler{store: store, jobs: jobs, sett: sett, loc: loc}
}

func (s *Scheduler) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(defaultSchedulerInterval)
		defer ticker.Stop()

		s.runTick(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.runTick(ctx)
			}
		}
	}()
}

func (s *Scheduler) Tick(ctx context.Context, now time.Time) (SchedulerResult, error) {
	localNow := now.In(s.loc)
	_ = s.setStatus(ctx, map[string]string{
		"scheduler_last_tick_at": localNow.Format(time.RFC3339),
		"scheduler_last_error":   "",
	})

	today := pgtype.Date{}
	if err := today.Scan(localNow.Format("2006-01-02")); err != nil {
		return SchedulerResult{}, err
	}

	schedules, err := s.store.ClaimDueSchedules(ctx, db.ClaimDueSchedulesParams{
		LastEnqueuedForDate: today,
		RunTime:             localNow.Format("15:04"),
	})
	if err != nil {
		return SchedulerResult{}, err
	}

	result := SchedulerResult{
		CheckedAt: localNow.Format("2006-01-02 15:04:05 MST"),
		Processed: len(schedules),
	}
	maxAttempts := s.defaultMaxAttempts(ctx)

	for _, schedule := range schedules {
		inserted, skipped, runErr := s.jobs.RunAll(ctx, string(schedule.RunType), maxAttempts)
		if runErr != nil {
			_ = s.store.ResetScheduleEnqueueState(ctx, db.ResetScheduleEnqueueStateParams{
				ID:                  schedule.ID,
				LastEnqueuedForDate: today,
			})
			_ = s.setStatus(ctx, map[string]string{
				"scheduler_last_error": runErr.Error(),
			})
			return result, runErr
		}
		result.Enqueued += inserted
		result.Skipped += skipped
	}

	_ = s.setStatus(ctx, map[string]string{
		"scheduler_last_success_at": localNow.Format(time.RFC3339),
		"scheduler_last_processed":  strconv.Itoa(result.Processed),
		"scheduler_last_enqueued":   strconv.Itoa(result.Enqueued),
		"scheduler_last_skipped":    strconv.Itoa(result.Skipped),
	})

	return result, nil
}

func (s *Scheduler) defaultMaxAttempts(ctx context.Context) int32 {
	row, err := s.store.GetSetting(ctx, "default_max_attempts")
	if err != nil {
		return defaultMaxAttempts
	}
	n, err := strconv.Atoi(row.Value)
	if err != nil || n < 1 || n > 10 {
		return defaultMaxAttempts
	}
	return int32(n)
}

func (s *Scheduler) runTick(ctx context.Context) {
	result, err := s.Tick(ctx, time.Now())
	if err != nil {
		log.Printf("scheduler tick failed: %v", err)
		return
	}
	if result.Processed == 0 {
		return
	}
	log.Printf("scheduler tick processed=%d enqueued=%d skipped=%d at=%s", result.Processed, result.Enqueued, result.Skipped, result.CheckedAt)
}

func (s *Scheduler) setStatus(ctx context.Context, values map[string]string) error {
	if s.sett == nil {
		return nil
	}
	for key, value := range values {
		if err := s.sett.Upsert(ctx, key, value); err != nil {
			return err
		}
	}
	return nil
}
