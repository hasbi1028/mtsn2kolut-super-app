package service

import (
	"context"
	"log/slog"
	"math/rand"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

const auditCleanupInterval = 24 * time.Hour

const (
	defaultSchedulerInterval = 30 * time.Second
	defaultMaxAttempts       = 3
)

type schedulerStore interface {
	ClaimDueSchedules(ctx context.Context, arg db.ClaimDueSchedulesParams) ([]db.ClaimDueSchedulesRow, error)
	ResetScheduleEnqueueState(ctx context.Context, arg db.ResetScheduleEnqueueStateParams) error
	ClaimDueEmployeeSchedules(ctx context.Context, arg db.ClaimDueEmployeeSchedulesParams) ([]db.ClaimDueEmployeeSchedulesRow, error)
	ResetEmployeeScheduleEnqueueState(ctx context.Context, arg db.ResetEmployeeScheduleEnqueueStateParams) error
	GetSetting(ctx context.Context, key string) (db.AppSetting, error)
	CreateDocumentCycleReminderEvents(ctx context.Context, today pgtype.Date) (int64, error)
	CreateDocumentCycleReminderNotifications(ctx context.Context, today pgtype.Date) (int64, error)
}

type pusakaSchedulerJobRunner interface {
	RunAll(ctx context.Context, runType string, maxAttempts int32) (inserted, skipped int, err error)
	Create(ctx context.Context, employeeID pgtype.UUID, runType string, maxAttempts int32) (db.Job, error)
	CreateWithDelay(ctx context.Context, employeeID pgtype.UUID, runType string, maxAttempts int32, notBefore pgtype.Timestamptz) (db.Job, error)
	RecoverStaleRunning(ctx context.Context, olderThan time.Duration) (int64, error)
}

type auditCleaner interface {
	CleanupOld(ctx context.Context) (int64, error)
}

type PusakaScheduler struct {
	store       schedulerStore
	jobs        pusakaSchedulerJobRunner
	sett        *Setting
	loc         *time.Location
	audit       auditCleaner
	lastCleanup time.Time
	cancel      context.CancelFunc
}

type PusakaSchedulerResult struct {
	CheckedAt             string `json:"checked_at"`
	Processed             int    `json:"processed"`
	Enqueued              int    `json:"enqueued"`
	Skipped               int    `json:"skipped"`
	DocumentReminders     int64  `json:"document_reminders"`
	DocumentNotifications int64  `json:"document_notifications"`
}

func NewPusakaScheduler(store *db.Queries, jobs *PusakaJob, sett *Setting, audit *Audit) *PusakaScheduler {
	loc, err := time.LoadLocation("Asia/Makassar")
	if err != nil {
		loc = time.FixedZone("WITA", 8*60*60)
	}
	return &PusakaScheduler{store: store, jobs: jobs, sett: sett, loc: loc, audit: audit}
}

func (s *PusakaScheduler) Start(ctx context.Context) {
	schedCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	go func() {
		ticker := time.NewTicker(defaultSchedulerInterval)
		defer ticker.Stop()

		s.runTick(schedCtx)
		for {
			select {
			case <-schedCtx.Done():
				return
			case <-ticker.C:
				s.runTick(schedCtx)
			}
		}
	}()
}

func (s *PusakaScheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *PusakaScheduler) Tick(ctx context.Context, now time.Time) (PusakaSchedulerResult, error) {
	localNow := now.In(s.loc)
	_ = s.setStatus(ctx, map[string]string{
		"scheduler_last_tick_at": localNow.Format(time.RFC3339),
		"scheduler_last_error":   "",
	})

	today := pgtype.Date{}
	if err := today.Scan(localNow.Format("2006-01-02")); err != nil {
		return PusakaSchedulerResult{}, err
	}

	claimParams := db.ClaimDueSchedulesParams{
		LastEnqueuedForDate: today,
		RunTime:             localNow.Format("15:04"),
	}

	if _, err := s.jobs.RecoverStaleRunning(ctx, defaultRunningJobStaleAfter); err != nil {
		return PusakaSchedulerResult{}, err
	}

	schedules, err := s.store.ClaimDueSchedules(ctx, claimParams)
	if err != nil {
		return PusakaSchedulerResult{}, err
	}

	result := PusakaSchedulerResult{
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

	// Process per-employee checkin/checkout schedules
	empSchedules, err := s.store.ClaimDueEmployeeSchedules(ctx, db.ClaimDueEmployeeSchedulesParams{
		LastEnqueuedForDate: today,
		RunTime:             localNow.Format("15:04"),
	})
	if err != nil {
		slog.Error("scheduler: failed to claim employee schedules", "error", err)
	} else {
		for _, es := range empSchedules {
			result.Processed++
			notBefore := pgtype.Timestamptz{}
			if es.RandomWindowMinutes > 0 {
				delayMinutes := rand.Int31n(int32(es.RandomWindowMinutes) + 1)
				_ = notBefore.Scan(now.Add(time.Duration(delayMinutes) * time.Minute))
			}
			_, createErr := s.jobs.CreateWithDelay(ctx, es.EmployeeID, string(es.RunType), maxAttempts, notBefore)
			if createErr != nil {
				_ = s.store.ResetEmployeeScheduleEnqueueState(ctx, db.ResetEmployeeScheduleEnqueueStateParams{
					ID:                  es.ID,
					LastEnqueuedForDate: today,
				})
				slog.Error("scheduler: failed to create employee job", "employee_id", es.EmployeeID, "run_type", es.RunType, "error", createErr)
				continue
			}
			result.Enqueued++
		}
	}

	reminders, err := s.store.CreateDocumentCycleReminderEvents(ctx, today)
	if err != nil {
		slog.Error("scheduler: failed to create document-cycle reminder events", "error", err)
		_ = s.setStatus(ctx, map[string]string{
			"scheduler_document_cycle_reminder_error": err.Error(),
		})
	} else {
		result.DocumentReminders = reminders
		_ = s.setStatus(ctx, map[string]string{
			"scheduler_document_cycle_reminders":      strconv.FormatInt(reminders, 10),
			"scheduler_document_cycle_reminder_error": "",
		})
	}

	notifications, err := s.store.CreateDocumentCycleReminderNotifications(ctx, today)
	if err != nil {
		slog.Error("scheduler: failed to create document-cycle reminder notifications", "error", err)
		_ = s.setStatus(ctx, map[string]string{
			"scheduler_document_cycle_notification_error": err.Error(),
		})
	} else {
		result.DocumentNotifications = notifications
		_ = s.setStatus(ctx, map[string]string{
			"scheduler_document_cycle_notifications":      strconv.FormatInt(notifications, 10),
			"scheduler_document_cycle_notification_error": "",
		})
	}

	_ = s.setStatus(ctx, map[string]string{
		"scheduler_last_success_at": localNow.Format(time.RFC3339),
		"scheduler_last_processed":  strconv.Itoa(result.Processed),
		"scheduler_last_enqueued":   strconv.Itoa(result.Enqueued),
		"scheduler_last_skipped":    strconv.Itoa(result.Skipped),
	})

	return result, nil
}

func (s *PusakaScheduler) defaultMaxAttempts(ctx context.Context) int32 {
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

func (s *PusakaScheduler) runTick(ctx context.Context) {
	result, err := s.Tick(ctx, time.Now())
	if err != nil {
		slog.Error("scheduler tick failed", "error", err)
		return
	}
	if result.Processed > 0 {
		slog.Info("scheduler tick", "processed", result.Processed, "enqueued", result.Enqueued, "skipped", result.Skipped, "at", result.CheckedAt)
	}

	s.maybeCleanupAudit(ctx)
}

func (s *PusakaScheduler) maybeCleanupAudit(ctx context.Context) {
	if s.audit == nil || s.sett == nil {
		return
	}
	if time.Since(s.lastCleanup) < auditCleanupInterval {
		return
	}

	row, err := s.store.GetSetting(ctx, "last_audit_cleanup_date")
	if err == nil {
		lastDate, parseErr := time.Parse(time.DateOnly, row.Value)
		if parseErr == nil {
			localNow := time.Now().In(s.loc)
			if lastDate.Year() == localNow.Year() && lastDate.YearDay() == localNow.YearDay() {
				s.lastCleanup = localNow
				return
			}
		}
	}

	deleted, err := s.audit.CleanupOld(ctx)
	if err != nil {
		slog.Error("audit cleanup failed", "error", err)
		return
	}

	localNowStr := time.Now().In(s.loc).Format(time.DateOnly)
	if setErr := s.sett.Upsert(ctx, "last_audit_cleanup_date", localNowStr); setErr != nil {
		slog.Error("audit cleanup: failed to update last_audit_cleanup_date", "error", setErr)
	}
	s.lastCleanup = time.Now()

	if deleted > 0 {
		slog.Info("audit cleanup: deleted old entries", "deleted", deleted)
	}
}

func (s *PusakaScheduler) setStatus(ctx context.Context, values map[string]string) error {
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
