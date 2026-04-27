package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/pusaka/backend/internal/domain"
	db "github.com/pusaka/backend/internal/repository/postgres"
)

type Job struct {
	q *db.Queries
}

func NewJob(q *db.Queries) *Job { return &Job{q: q} }

func (s *Job) List(ctx context.Context, status string, limit, offset int32) ([]db.ListJobsRow, int64, error) {
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

func (s *Job) Create(ctx context.Context, employeeID pgtype.UUID, runType string, maxAttempts int32) (db.Job, error) {
	return s.q.CreateJob(ctx, db.CreateJobParams{
		EmployeeID:  employeeID,
		RunType:     db.RunTypeEnum(runType),
		MaxAttempts: maxAttempts,
	})
}

func (s *Job) Stats(ctx context.Context) (db.GetJobStatsRow, error) {
	return s.q.GetJobStats(ctx)
}

func (s *Job) Claim(ctx context.Context, workerID string) (db.ClaimJobRow, error) {
	row, err := s.q.ClaimJob(ctx, workerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return db.ClaimJobRow{}, domain.ErrNoJob
	}
	return row, err
}

func (s *Job) Complete(ctx context.Context, id pgtype.UUID) error {
	return s.q.CompleteJob(ctx, id)
}

func (s *Job) Fail(ctx context.Context, id pgtype.UUID, errMsg string, retryAfterSecs pgtype.Text) error {
	return s.q.FailJob(ctx, db.FailJobParams{
		ID:           id,
		ErrorMessage: errMsg,
		Column3:      retryAfterSecs,
	})
}
