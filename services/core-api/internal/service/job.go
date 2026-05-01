package service

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

const defaultRunningJobStaleAfter = 30 * time.Minute

type PusakaJob struct {
	q jobStore
}

type jobStore interface {
	ListJobsByStatus(ctx context.Context, arg db.ListJobsByStatusParams) ([]db.ListJobsByStatusRow, error)
	CountJobsByStatus(ctx context.Context, status db.JobStatusEnum) (int64, error)
	ListJobs(ctx context.Context, arg db.ListJobsParams) ([]db.ListJobsRow, error)
	CountJobs(ctx context.Context) (int64, error)
	CreateJobIfAbsent(ctx context.Context, arg db.CreateJobIfAbsentParams) (db.Job, error)
	GetJobStats(ctx context.Context) (db.GetJobStatsRow, error)
	ClaimJob(ctx context.Context, workerID string) (db.ClaimJobRow, error)
	RecoverStaleRunningJobs(ctx context.Context, staleAfterSeconds int32) (int64, error)
	CompleteJob(ctx context.Context, id pgtype.UUID) error
	FailJob(ctx context.Context, arg db.FailJobParams) error
	GetJob(ctx context.Context, id pgtype.UUID) (db.GetJobRow, error)
	ListActiveEmployees(ctx context.Context) ([]db.ListActiveEmployeesRow, error)
	CancelEmployeeJobs(ctx context.Context, employeeID pgtype.UUID) (int64, error)
	CancelAllJobs(ctx context.Context) (int64, error)
}

func NewPusakaJob(q *db.Queries) *PusakaJob { return &PusakaJob{q: q} }

func (s *PusakaJob) List(ctx context.Context, status string, limit, offset int32) ([]db.ListJobsRow, int64, error) {
	if status != "" {
		rows, err := s.q.ListJobsByStatus(ctx, db.ListJobsByStatusParams{
			Status: db.JobStatusEnum(status),
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			return nil, 0, err
		}
		count, err := s.q.CountJobsByStatus(ctx, db.JobStatusEnum(status))
		result := make([]db.ListJobsRow, len(rows))
		for i, r := range rows {
			result[i] = db.ListJobsRow{
				ID: r.ID, EmployeeID: r.EmployeeID,
				EmployeeNama: r.EmployeeNama, EmployeeNip: r.EmployeeNip,
				RunType: r.RunType, Status: r.Status, ErrorMessage: r.ErrorMessage,
				ClaimedBy: r.ClaimedBy, ClaimedAt: r.ClaimedAt,
				Attempts: r.Attempts, MaxAttempts: r.MaxAttempts,
				NextRetryAt: r.NextRetryAt, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
			}
		}
		return result, count, err
	}
	rows, err := s.q.ListJobs(ctx, db.ListJobsParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, 0, err
	}
	count, err := s.q.CountJobs(ctx)
	return rows, count, err
}

func (s *PusakaJob) Create(ctx context.Context, employeeID pgtype.UUID, runType string, maxAttempts int32) (db.Job, error) {
	return s.CreateWithDelay(ctx, employeeID, runType, maxAttempts, pgtype.Timestamptz{})
}

func (s *PusakaJob) CreateWithDelay(ctx context.Context, employeeID pgtype.UUID, runType string, maxAttempts int32, notBefore pgtype.Timestamptz) (db.Job, error) {
	job, err := s.q.CreateJobIfAbsent(ctx, db.CreateJobIfAbsentParams{
		EmployeeID:  employeeID,
		RunType:     db.RunTypeEnum(runType),
		MaxAttempts: maxAttempts,
		NotBefore:   notBefore,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return db.Job{}, domain.ErrConflict
	}
	return job, err
}

func (s *PusakaJob) Stats(ctx context.Context) (db.GetJobStatsRow, error) {
	return s.q.GetJobStats(ctx)
}

func (s *PusakaJob) RecoverStaleRunning(ctx context.Context, olderThan time.Duration) (int64, error) {
	if olderThan <= 0 {
		olderThan = defaultRunningJobStaleAfter
	}
	return s.q.RecoverStaleRunningJobs(ctx, int32(olderThan.Seconds()))
}

func (s *PusakaJob) Claim(ctx context.Context, workerID string) (db.ClaimJobRow, error) {
	if _, err := s.RecoverStaleRunning(ctx, defaultRunningJobStaleAfter); err != nil {
		return db.ClaimJobRow{}, err
	}
	row, err := s.q.ClaimJob(ctx, workerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return db.ClaimJobRow{}, domain.ErrNoJob
	}
	return row, err
}

func (s *PusakaJob) Complete(ctx context.Context, id pgtype.UUID) error {
	return s.q.CompleteJob(ctx, id)
}

func (s *PusakaJob) Fail(ctx context.Context, id pgtype.UUID, errMsg string, retryAfterSecs pgtype.Text) error {
	return s.q.FailJob(ctx, db.FailJobParams{
		ID:           id,
		ErrorMessage: errMsg,
		Column3:      retryAfterSecs,
	})
}

func (s *PusakaJob) Get(ctx context.Context, id pgtype.UUID) (db.GetJobRow, error) {
	return s.q.GetJob(ctx, id)
}

func (s *PusakaJob) RunAll(ctx context.Context, runType string, maxAttempts int32) (inserted, skipped int, err error) {
	emps, err := s.q.ListActiveEmployees(ctx)
	if err != nil {
		return 0, 0, err
	}
	for _, emp := range emps {
		_, err := s.q.CreateJobIfAbsent(ctx, db.CreateJobIfAbsentParams{
			EmployeeID:  emp.ID,
			RunType:     db.RunTypeEnum(runType),
			MaxAttempts: maxAttempts,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			skipped++
			continue
		}
		if err != nil {
			return inserted, skipped, err
		}
		inserted++
	}
	return inserted, skipped, nil
}

func (s *PusakaJob) CancelEmployee(ctx context.Context, employeeID pgtype.UUID) (int64, error) {
	return s.q.CancelEmployeeJobs(ctx, employeeID)
}

func (s *PusakaJob) CancelAll(ctx context.Context) (int64, error) {
	return s.q.CancelAllJobs(ctx)
}
