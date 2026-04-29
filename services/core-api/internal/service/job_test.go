package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeJobStore struct {
	activeEmployees []db.Employee
	createResults   []error
	createCalls     int
}

func (f *fakeJobStore) ListJobsByStatus(ctx context.Context, arg db.ListJobsByStatusParams) ([]db.ListJobsByStatusRow, error) {
	return nil, nil
}

func (f *fakeJobStore) CountJobsByStatus(ctx context.Context, status db.JobStatusEnum) (int64, error) {
	return 0, nil
}

func (f *fakeJobStore) ListJobs(ctx context.Context, arg db.ListJobsParams) ([]db.ListJobsRow, error) {
	return nil, nil
}

func (f *fakeJobStore) CountJobs(ctx context.Context) (int64, error) {
	return 0, nil
}

func (f *fakeJobStore) CreateJobIfAbsent(ctx context.Context, arg db.CreateJobIfAbsentParams) (db.Job, error) {
	var err error
	if f.createCalls < len(f.createResults) {
		err = f.createResults[f.createCalls]
	}
	f.createCalls++
	if err != nil {
		return db.Job{}, err
	}
	return db.Job{
		EmployeeID:  arg.EmployeeID,
		RunType:     arg.RunType,
		MaxAttempts: arg.MaxAttempts,
	}, nil
}

func (f *fakeJobStore) GetJobStats(ctx context.Context) (db.GetJobStatsRow, error) {
	return db.GetJobStatsRow{}, nil
}

func (f *fakeJobStore) ClaimJob(ctx context.Context, workerID string) (db.ClaimJobRow, error) {
	return db.ClaimJobRow{}, nil
}

func (f *fakeJobStore) CompleteJob(ctx context.Context, id pgtype.UUID) error {
	return nil
}

func (f *fakeJobStore) FailJob(ctx context.Context, arg db.FailJobParams) error {
	return nil
}

func (f *fakeJobStore) GetJob(ctx context.Context, id pgtype.UUID) (db.GetJobRow, error) {
	return db.GetJobRow{}, nil
}

func (f *fakeJobStore) ListActiveEmployees(ctx context.Context) ([]db.Employee, error) {
	return f.activeEmployees, nil
}

func (f *fakeJobStore) CancelEmployeeJobs(ctx context.Context, employeeID pgtype.UUID) (int64, error) {
	return 0, nil
}

func (f *fakeJobStore) CancelAllJobs(ctx context.Context) (int64, error) {
	return 0, nil
}

func TestJobCreateReturnsConflictWhenActiveJobExists(t *testing.T) {
	store := &fakeJobStore{createResults: []error{pgx.ErrNoRows}}
	svc := &Job{q: store}

	_, err := svc.Create(context.Background(), pgtype.UUID{Valid: true}, "morning", 3)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("Create() error = %v, want ErrConflict", err)
	}
}

func TestJobRunAllCountsInsertedAndSkipped(t *testing.T) {
	store := &fakeJobStore{
		activeEmployees: []db.Employee{
			{ID: pgtype.UUID{Bytes: [16]byte{1}, Valid: true}},
			{ID: pgtype.UUID{Bytes: [16]byte{2}, Valid: true}},
			{ID: pgtype.UUID{Bytes: [16]byte{3}, Valid: true}},
		},
		createResults: []error{nil, pgx.ErrNoRows, nil},
	}
	svc := &Job{q: store}

	inserted, skipped, err := svc.RunAll(context.Background(), "morning", 3)
	if err != nil {
		t.Fatalf("RunAll() error = %v", err)
	}
	if inserted != 2 || skipped != 1 {
		t.Fatalf("RunAll() = inserted:%d skipped:%d, want inserted:2 skipped:1", inserted, skipped)
	}
}

func TestJobRunAllReturnsUnexpectedStoreError(t *testing.T) {
	store := &fakeJobStore{
		activeEmployees: []db.Employee{
			{ID: pgtype.UUID{Bytes: [16]byte{1}, Valid: true}},
		},
		createResults: []error{errors.New("db down")},
	}
	svc := &Job{q: store}

	_, _, err := svc.RunAll(context.Background(), "morning", 3)
	if err == nil || err.Error() != "db down" {
		t.Fatalf("RunAll() error = %v, want db down", err)
	}
}
