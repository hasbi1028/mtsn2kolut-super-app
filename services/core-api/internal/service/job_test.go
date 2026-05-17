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
	completeRows    int64
	completeErr     error
	failArg         db.FailJobParams
	failRows        int64
	failRowsSet     bool
	failErr         error
	getID           pgtype.UUID
	getRow          db.GetJobRow
	getErr          error
	runningArg      db.GetRunningJobForWorkerParams
	runningRow      db.GetRunningJobForWorkerRow
	runningErr      error
	upsertArg       db.UpsertAttendanceParams
	upsertErr       error
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
	if f.completeErr != nil {
		return 0, f.completeErr
	}
	if f.completeRows != 0 {
		return f.completeRows, nil
	}
	return 1, nil
}

func (f *fakeJobStore) FailJob(ctx context.Context, arg db.FailJobParams) (int64, error) {
	f.failArg = arg
	if f.failErr != nil {
		return 0, f.failErr
	}
	if f.failRowsSet {
		return f.failRows, nil
	}
	return 1, nil
}

func (f *fakeJobStore) GetJob(ctx context.Context, id pgtype.UUID) (db.GetJobRow, error) {
	f.getID = id
	if f.getErr != nil {
		return db.GetJobRow{}, f.getErr
	}
	return f.getRow, nil
}

func (f *fakeJobStore) GetRunningJobForWorker(ctx context.Context, arg db.GetRunningJobForWorkerParams) (db.GetRunningJobForWorkerRow, error) {
	f.runningArg = arg
	if f.runningErr != nil {
		return db.GetRunningJobForWorkerRow{}, f.runningErr
	}
	return f.runningRow, nil
}

func (f *fakeJobStore) UpsertAttendance(ctx context.Context, arg db.UpsertAttendanceParams) (db.AttendanceRecord, error) {
	f.upsertArg = arg
	if f.upsertErr != nil {
		return db.AttendanceRecord{}, f.upsertErr
	}
	return db.AttendanceRecord{EmployeeID: arg.EmployeeID, Tanggal: arg.Tanggal, JamMasuk: arg.JamMasuk, JamPulang: arg.JamPulang, SourceJobID: arg.SourceJobID}, nil
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

func TestJobCompleteWithAttendanceLocksJobUpsertsAttendanceAndCompletes(t *testing.T) {
	jobID := pgtype.UUID{Bytes: [16]byte{10}, Valid: true}
	employeeID := pgtype.UUID{Bytes: [16]byte{11}, Valid: true}
	tanggal := pgtype.Date{Time: time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC), Valid: true}
	store := &fakeJobStore{
		runningRow: db.GetRunningJobForWorkerRow{ID: jobID, EmployeeID: employeeID, Status: db.JobStatusEnumRunning, ClaimedBy: "worker-1"},
	}
	svc := &PusakaJob{q: store}

	err := svc.CompleteWithAttendance(context.Background(), jobID, "worker-1", &PusakaJobAttendanceInput{
		Tanggal:   tanggal,
		JamMasuk:  "07:10",
		JamPulang: "15:00",
	})
	if err != nil {
		t.Fatalf("CompleteWithAttendance() error = %v", err)
	}
	if store.runningArg.ID != jobID || store.runningArg.ClaimedBy != "worker-1" {
		t.Fatalf("GetRunningJobForWorker() arg = %+v, want job/worker lock", store.runningArg)
	}
	if store.upsertArg.EmployeeID != employeeID || store.upsertArg.Tanggal != tanggal || store.upsertArg.JamMasuk != "07:10" || store.upsertArg.JamPulang != "15:00" || store.upsertArg.SourceJobID != jobID {
		t.Fatalf("UpsertAttendance() arg = %+v, want locked job employee attendance", store.upsertArg)
	}
	if store.completeArg.ID != jobID || store.completeArg.ClaimedBy != "worker-1" {
		t.Fatalf("CompleteJob() arg = %+v, want job/worker", store.completeArg)
	}
}

func TestJobCompleteWithAttendanceReturnsSeedErrorBeforeComplete(t *testing.T) {
	jobID := pgtype.UUID{Bytes: [16]byte{12}, Valid: true}
	employeeID := pgtype.UUID{Bytes: [16]byte{13}, Valid: true}
	expectedErr := errors.New("attendance failed")
	store := &fakeJobStore{
		runningRow: db.GetRunningJobForWorkerRow{ID: jobID, EmployeeID: employeeID, Status: db.JobStatusEnumRunning, ClaimedBy: "worker-1"},
		upsertErr:  expectedErr,
	}
	svc := &PusakaJob{q: store}

	err := svc.CompleteWithAttendance(context.Background(), jobID, "worker-1", &PusakaJobAttendanceInput{
		Tanggal: pgtype.Date{Time: time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC), Valid: true},
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("CompleteWithAttendance() error = %v, want %v", err, expectedErr)
	}
	if store.completeArg.ID.Valid {
		t.Fatalf("CompleteJob() arg = %+v, want not called after attendance failure", store.completeArg)
	}
}

func TestJobCompleteWithAttendanceMapsMissingRunningJobToConflictOrIdempotentSuccess(t *testing.T) {
	jobID := pgtype.UUID{Bytes: [16]byte{14}, Valid: true}
	store := &fakeJobStore{
		runningErr: pgx.ErrNoRows,
		getRow:     db.GetJobRow{ID: jobID, Status: db.JobStatusEnumRunning},
	}
	svc := &PusakaJob{q: store}

	err := svc.CompleteWithAttendance(context.Background(), jobID, "worker-1", &PusakaJobAttendanceInput{Tanggal: pgtype.Date{Valid: true}})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("CompleteWithAttendance(conflict) error = %v, want ErrConflict", err)
	}

	store.getRow.Status = db.JobStatusEnumSuccess
	err = svc.CompleteWithAttendance(context.Background(), jobID, "worker-1", &PusakaJobAttendanceInput{Tanggal: pgtype.Date{Valid: true}})
	if err != nil {
		t.Fatalf("CompleteWithAttendance(already success) error = %v, want nil", err)
	}
}

type completionlessJobStore struct {
	inner *fakeJobStore
}

func (s completionlessJobStore) ListJobsByStatus(ctx context.Context, arg db.ListJobsByStatusParams) ([]db.ListJobsByStatusRow, error) {
	return s.inner.ListJobsByStatus(ctx, arg)
}
func (s completionlessJobStore) CountJobsByStatus(ctx context.Context, status db.JobStatusEnum) (int64, error) {
	return s.inner.CountJobsByStatus(ctx, status)
}
func (s completionlessJobStore) ListJobs(ctx context.Context, arg db.ListJobsParams) ([]db.ListJobsRow, error) {
	return s.inner.ListJobs(ctx, arg)
}
func (s completionlessJobStore) CountJobs(ctx context.Context) (int64, error) {
	return s.inner.CountJobs(ctx)
}
func (s completionlessJobStore) CreateJobIfAbsent(ctx context.Context, arg db.CreateJobIfAbsentParams) (db.Job, error) {
	return s.inner.CreateJobIfAbsent(ctx, arg)
}
func (s completionlessJobStore) GetJobStats(ctx context.Context) (db.GetJobStatsRow, error) {
	return s.inner.GetJobStats(ctx)
}
func (s completionlessJobStore) ClaimJob(ctx context.Context, workerID string) (db.ClaimJobRow, error) {
	return s.inner.ClaimJob(ctx, workerID)
}
func (s completionlessJobStore) RecoverStaleRunningJobs(ctx context.Context, staleAfterSeconds int32) (int64, error) {
	return s.inner.RecoverStaleRunningJobs(ctx, staleAfterSeconds)
}
func (s completionlessJobStore) CompleteJob(ctx context.Context, arg db.CompleteJobParams) (int64, error) {
	return s.inner.CompleteJob(ctx, arg)
}
func (s completionlessJobStore) FailJob(ctx context.Context, arg db.FailJobParams) (int64, error) {
	return s.inner.FailJob(ctx, arg)
}
func (s completionlessJobStore) GetJob(ctx context.Context, id pgtype.UUID) (db.GetJobRow, error) {
	return s.inner.GetJob(ctx, id)
}
func (s completionlessJobStore) ListActiveEmployees(ctx context.Context) ([]db.ListActiveEmployeesRow, error) {
	return s.inner.ListActiveEmployees(ctx)
}
func (s completionlessJobStore) CancelEmployeeJobs(ctx context.Context, employeeID pgtype.UUID) (int64, error) {
	return s.inner.CancelEmployeeJobs(ctx, employeeID)
}
func (s completionlessJobStore) CancelAllJobs(ctx context.Context) (int64, error) {
	return s.inner.CancelAllJobs(ctx)
}

func TestJobCompleteWithAttendanceRequiresCompletionStoreWhenNoTx(t *testing.T) {
	svc := &PusakaJob{q: completionlessJobStore{inner: &fakeJobStore{}}}

	err := svc.CompleteWithAttendance(context.Background(), pgtype.UUID{Bytes: [16]byte{15}, Valid: true}, "worker-1", &PusakaJobAttendanceInput{Tanggal: pgtype.Date{Valid: true}})
	if err == nil || err.Error() != "job store does not support attendance completion" {
		t.Fatalf("CompleteWithAttendance(completionless store) error = %v, want unsupported completion store", err)
	}
}

func TestJobFailMapsAffectedRowsAndStoreErrors(t *testing.T) {
	jobID := pgtype.UUID{Bytes: [16]byte{16}, Valid: true}
	retryAfter := pgtype.Text{String: "30", Valid: true}

	store := &fakeJobStore{failErr: errors.New("fail update failed")}
	svc := &PusakaJob{q: store}
	if err := svc.Fail(context.Background(), jobID, "worker-1", "timeout", retryAfter); err == nil || err.Error() != "fail update failed" {
		t.Fatalf("Fail(store error) error = %v, want fail update failed", err)
	}

	for _, status := range []db.JobStatusEnum{db.JobStatusEnumSuccess, db.JobStatusEnumFailed} {
		store = &fakeJobStore{failRowsSet: true, getRow: db.GetJobRow{ID: jobID, Status: status}}
		svc = &PusakaJob{q: store}
		if err := svc.Fail(context.Background(), jobID, "worker-1", "timeout", retryAfter); err != nil {
			t.Fatalf("Fail(already terminal %s) error = %v, want nil", status, err)
		}
		if store.getID != jobID {
			t.Fatalf("Fail(already terminal %s) GetJob id = %v, want %v", status, store.getID, jobID)
		}
	}

	store = &fakeJobStore{failRowsSet: true, getRow: db.GetJobRow{ID: jobID, Status: db.JobStatusEnumRunning}}
	svc = &PusakaJob{q: store}
	if err := svc.Fail(context.Background(), jobID, "worker-1", "timeout", retryAfter); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("Fail(non-terminal affected 0) error = %v, want ErrConflict", err)
	}

	store = &fakeJobStore{failRowsSet: true, getErr: pgx.ErrNoRows}
	svc = &PusakaJob{q: store}
	if err := svc.Fail(context.Background(), jobID, "worker-1", "timeout", retryAfter); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("Fail(missing job affected 0) error = %v, want ErrConflict", err)
	}
}
