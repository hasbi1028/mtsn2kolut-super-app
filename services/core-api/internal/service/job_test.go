package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeJobStore struct {
	activeEmployees []db.ListActiveEmployeesRow
	activeErr       error
	createResults   []error
	createArgs      []db.CreateJobIfAbsentParams
	createCalls     int
	recoverCalls    int
	recoverSeconds  int32
	recoverErr      error
	claimRow        db.ClaimJobRow
	claimErr        error
	claimCalls      int
	listArg         db.ListJobsParams
	listRows        []db.ListJobsRow
	listErr         error
	listCount       int64
	listCountErr    error
	listStatusArg   db.ListJobsByStatusParams
	listStatusRows  []db.ListJobsByStatusRow
	listStatusErr   error
	listStatusCount int64
	statusCountErr  error
	statsRow        db.GetJobStatsRow
	completeArg     db.CompleteJobParams
	failArg         db.FailJobParams
	getID           pgtype.UUID
	getRow          db.GetJobRow
	cancelEmployee  pgtype.UUID
	cancelCount     int64
	cancelAllCount  int64
}

func (f *fakeJobStore) ListJobsByStatus(ctx context.Context, arg db.ListJobsByStatusParams) ([]db.ListJobsByStatusRow, error) {
	f.listStatusArg = arg
	return f.listStatusRows, f.listStatusErr
}

func (f *fakeJobStore) CountJobsByStatus(ctx context.Context, status db.JobStatusEnum) (int64, error) {
	return f.listStatusCount, f.statusCountErr
}

func (f *fakeJobStore) ListJobs(ctx context.Context, arg db.ListJobsParams) ([]db.ListJobsRow, error) {
	f.listArg = arg
	return f.listRows, f.listErr
}

func (f *fakeJobStore) CountJobs(ctx context.Context) (int64, error) {
	return f.listCount, f.listCountErr
}

func (f *fakeJobStore) CreateJobIfAbsent(ctx context.Context, arg db.CreateJobIfAbsentParams) (db.Job, error) {
	f.createArgs = append(f.createArgs, arg)
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
	return f.statsRow, nil
}

func (f *fakeJobStore) ClaimJob(ctx context.Context, workerID string) (db.ClaimJobRow, error) {
	f.claimCalls++
	return f.claimRow, f.claimErr
}

func (f *fakeJobStore) RecoverStaleRunningJobs(ctx context.Context, staleAfterSeconds int32) (int64, error) {
	f.recoverCalls++
	f.recoverSeconds = staleAfterSeconds
	return 0, f.recoverErr
}

func (f *fakeJobStore) CompleteJob(ctx context.Context, arg db.CompleteJobParams) (int64, error) {
	f.completeArg = arg
	return 1, nil
}

func (f *fakeJobStore) FailJob(ctx context.Context, arg db.FailJobParams) (int64, error) {
	f.failArg = arg
	return 1, nil
}

func (f *fakeJobStore) GetJob(ctx context.Context, id pgtype.UUID) (db.GetJobRow, error) {
	f.getID = id
	return f.getRow, nil
}

func (f *fakeJobStore) ListActiveEmployees(ctx context.Context) ([]db.ListActiveEmployeesRow, error) {
	return f.activeEmployees, f.activeErr
}

func (f *fakeJobStore) CancelEmployeeJobs(ctx context.Context, employeeID pgtype.UUID) (int64, error) {
	f.cancelEmployee = employeeID
	return f.cancelCount, nil
}

func (f *fakeJobStore) CancelAllJobs(ctx context.Context) (int64, error) {
	return f.cancelAllCount, nil
}

func TestJobCreateReturnsConflictWhenActiveJobExists(t *testing.T) {
	store := &fakeJobStore{createResults: []error{pgx.ErrNoRows}}
	svc := &PusakaJob{q: store}

	_, err := svc.Create(context.Background(), pgtype.UUID{Valid: true}, "morning", 3)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("Create() error = %v, want ErrConflict", err)
	}
}

func TestJobCreateWithDelayForwardsPayloadAndErrors(t *testing.T) {
	employeeID := pgtype.UUID{Bytes: [16]byte{9}, Valid: true}
	notBefore := pgtype.Timestamptz{Time: time.Date(2026, time.May, 1, 8, 30, 0, 0, time.UTC), Valid: true}
	store := &fakeJobStore{}
	svc := &PusakaJob{q: store}

	job, err := svc.CreateWithDelay(context.Background(), employeeID, "checkout", 5, notBefore)
	if err != nil {
		t.Fatalf("CreateWithDelay() error = %v", err)
	}
	if job.EmployeeID != employeeID || job.RunType != db.RunTypeEnumCheckout || job.MaxAttempts != 5 {
		t.Fatalf("CreateWithDelay() job = %+v, want forwarded job fields", job)
	}
	if len(store.createArgs) != 1 || store.createArgs[0].EmployeeID != employeeID || store.createArgs[0].RunType != db.RunTypeEnumCheckout || store.createArgs[0].MaxAttempts != 5 || store.createArgs[0].NotBefore != notBefore {
		t.Fatalf("CreateWithDelay() arg = %+v, want forwarded payload", store.createArgs)
	}

	store = &fakeJobStore{createResults: []error{errors.New("insert failed")}}
	svc = &PusakaJob{q: store}
	_, err = svc.CreateWithDelay(context.Background(), employeeID, "checkout", 5, notBefore)
	if err == nil || err.Error() != "insert failed" {
		t.Fatalf("CreateWithDelay(unexpected error) error = %v, want insert failed", err)
	}
}

func TestJobRunAllCountsInsertedAndSkipped(t *testing.T) {
	store := &fakeJobStore{
		activeEmployees: []db.ListActiveEmployeesRow{
			{ID: pgtype.UUID{Bytes: [16]byte{1}, Valid: true}},
			{ID: pgtype.UUID{Bytes: [16]byte{2}, Valid: true}},
			{ID: pgtype.UUID{Bytes: [16]byte{3}, Valid: true}},
		},
		createResults: []error{nil, pgx.ErrNoRows, nil},
	}
	svc := &PusakaJob{q: store}

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
		activeEmployees: []db.ListActiveEmployeesRow{
			{ID: pgtype.UUID{Bytes: [16]byte{1}, Valid: true}},
		},
		createResults: []error{errors.New("db down")},
	}
	svc := &PusakaJob{q: store}

	_, _, err := svc.RunAll(context.Background(), "morning", 3)
	if err == nil || err.Error() != "db down" {
		t.Fatalf("RunAll() error = %v, want db down", err)
	}
}

func TestJobRunAllReturnsActiveEmployeeError(t *testing.T) {
	store := &fakeJobStore{activeErr: errors.New("employee lookup failed")}
	svc := &PusakaJob{q: store}

	inserted, skipped, err := svc.RunAll(context.Background(), "morning", 3)
	if err == nil || err.Error() != "employee lookup failed" {
		t.Fatalf("RunAll() = %d/%d/%v, want active employee error", inserted, skipped, err)
	}
}

func TestJobClaimRecoversStaleRunningBeforeClaim(t *testing.T) {
	store := &fakeJobStore{}
	svc := &PusakaJob{q: store}

	if _, err := svc.Claim(context.Background(), "worker-1"); err != nil {
		t.Fatalf("Claim() error = %v", err)
	}
	if store.recoverCalls != 1 {
		t.Fatalf("RecoverStaleRunningJobs() calls = %d, want 1", store.recoverCalls)
	}
	if store.claimCalls != 1 {
		t.Fatalf("ClaimJob() calls = %d, want 1", store.claimCalls)
	}
}

func TestJobClaimStopsWhenStaleRecoveryFails(t *testing.T) {
	store := &fakeJobStore{recoverErr: errors.New("recover failed")}
	svc := &PusakaJob{q: store}

	_, err := svc.Claim(context.Background(), "worker-1")
	if err == nil || err.Error() != "recover failed" {
		t.Fatalf("Claim() error = %v, want recover failed", err)
	}
	if store.claimCalls != 0 {
		t.Fatalf("ClaimJob() calls = %d, want 0", store.claimCalls)
	}
}

func TestJobClaimMapsNoRowsAndUnexpectedErrors(t *testing.T) {
	store := &fakeJobStore{claimErr: pgx.ErrNoRows}
	svc := &PusakaJob{q: store}

	_, err := svc.Claim(context.Background(), "worker-1")
	if !errors.Is(err, domain.ErrNoJob) {
		t.Fatalf("Claim(no rows) error = %v, want ErrNoJob", err)
	}
	if store.claimCalls != 1 {
		t.Fatalf("ClaimJob(no rows) calls = %d, want 1", store.claimCalls)
	}

	store = &fakeJobStore{claimErr: errors.New("claim failed")}
	svc = &PusakaJob{q: store}
	_, err = svc.Claim(context.Background(), "worker-1")
	if err == nil || err.Error() != "claim failed" {
		t.Fatalf("Claim(unexpected error) error = %v, want claim failed", err)
	}
}

func TestJobListPropagatesStoreErrors(t *testing.T) {
	tests := []struct {
		name      string
		status    string
		store     *fakeJobStore
		wantError string
	}{
		{name: "all list", store: &fakeJobStore{listErr: errors.New("list failed")}, wantError: "list failed"},
		{name: "all count", store: &fakeJobStore{listCountErr: errors.New("count failed")}, wantError: "count failed"},
		{name: "status list", status: "queued", store: &fakeJobStore{listStatusErr: errors.New("status list failed")}, wantError: "status list failed"},
		{name: "status count", status: "queued", store: &fakeJobStore{statusCountErr: errors.New("status count failed")}, wantError: "status count failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &PusakaJob{q: tt.store}
			_, _, err := svc.List(context.Background(), tt.status, 10, 0)
			if err == nil || err.Error() != tt.wantError {
				t.Fatalf("List(%q) error = %v, want %q", tt.status, err, tt.wantError)
			}
		})
	}
}

func TestJobListStatsAndStateMutationsForwardStoreCalls(t *testing.T) {
	jobID := pgtype.UUID{Bytes: [16]byte{7}, Valid: true}
	employeeID := pgtype.UUID{Bytes: [16]byte{8}, Valid: true}
	store := &fakeJobStore{
		listRows: []db.ListJobsRow{
			{ID: jobID, EmployeeID: employeeID, RunType: db.RunTypeEnumMorning, Status: db.JobStatusEnumQueued},
		},
		listCount: 9,
		listStatusRows: []db.ListJobsByStatusRow{
			{ID: jobID, EmployeeID: employeeID, RunType: db.RunTypeEnumMorning, Status: db.JobStatusEnumQueued},
		},
		listStatusCount: 2,
		statsRow:        db.GetJobStatsRow{Queued: 3},
		getRow:          db.GetJobRow{ID: jobID, EmployeeID: employeeID},
		cancelCount:     4,
		cancelAllCount:  5,
	}
	svc := &PusakaJob{q: store}

	rows, total, err := svc.List(context.Background(), "", 25, 50)
	if err != nil {
		t.Fatalf("List(all) error = %v", err)
	}
	if len(rows) != 1 || total != 9 || store.listArg.Limit != 25 || store.listArg.Offset != 50 {
		t.Fatalf("List(all) rows/total/arg = %d/%d/%+v, want 1/9/limit-offset", len(rows), total, store.listArg)
	}

	rows, total, err = svc.List(context.Background(), "queued", 10, 20)
	if err != nil {
		t.Fatalf("List(status) error = %v", err)
	}
	if len(rows) != 1 || total != 2 || store.listStatusArg.Status != db.JobStatusEnumQueued || store.listStatusArg.Limit != 10 || store.listStatusArg.Offset != 20 {
		t.Fatalf("List(status) rows/total/arg = %d/%d/%+v, want queued filters", len(rows), total, store.listStatusArg)
	}

	stats, err := svc.Stats(context.Background())
	if err != nil {
		t.Fatalf("Stats() error = %v", err)
	}
	if stats.Queued != 3 {
		t.Fatalf("Stats().Queued = %d, want 3", stats.Queued)
	}

	if _, err := svc.RecoverStaleRunning(context.Background(), 0); err != nil {
		t.Fatalf("RecoverStaleRunning(default) error = %v", err)
	}
	if store.recoverSeconds != int32(defaultRunningJobStaleAfter.Seconds()) {
		t.Fatalf("RecoverStaleRunning(default) seconds = %d, want default", store.recoverSeconds)
	}
	if _, err := svc.RecoverStaleRunning(context.Background(), 2*time.Minute); err != nil {
		t.Fatalf("RecoverStaleRunning(custom) error = %v", err)
	}
	if store.recoverSeconds != 120 {
		t.Fatalf("RecoverStaleRunning(custom) seconds = %d, want 120", store.recoverSeconds)
	}

	if err := svc.Complete(context.Background(), jobID, "worker-1"); err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if store.completeArg.ID != jobID || store.completeArg.ClaimedBy != "worker-1" {
		t.Fatalf("Complete() arg = %+v, want id and worker", store.completeArg)
	}
	retryAfter := pgtype.Text{String: "60", Valid: true}
	if err := svc.Fail(context.Background(), jobID, "worker-1", "network", retryAfter); err != nil {
		t.Fatalf("Fail() error = %v", err)
	}
	if store.failArg.ID != jobID || store.failArg.ClaimedBy != "worker-1" || store.failArg.ErrorMessage != "network" || store.failArg.Column4 != retryAfter {
		t.Fatalf("Fail() arg = %+v, want forwarded failure", store.failArg)
	}
	got, err := svc.Get(context.Background(), jobID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.ID != jobID || store.getID != jobID {
		t.Fatalf("Get() row/id = %+v/%v, want %v", got, store.getID, jobID)
	}
	canceled, err := svc.CancelEmployee(context.Background(), employeeID)
	if err != nil {
		t.Fatalf("CancelEmployee() error = %v", err)
	}
	if canceled != 4 || store.cancelEmployee != employeeID {
		t.Fatalf("CancelEmployee() = %d/%v, want 4/%v", canceled, store.cancelEmployee, employeeID)
	}
	canceled, err = svc.CancelAll(context.Background())
	if err != nil {
		t.Fatalf("CancelAll() error = %v", err)
	}
	if canceled != 5 {
		t.Fatalf("CancelAll() = %d, want 5", canceled)
	}
}
