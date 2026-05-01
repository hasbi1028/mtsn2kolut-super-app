package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeSchedulerStore struct {
	settings                 map[string]string
	claimedSchedules         []db.ClaimDueSchedulesRow
	claimErr                 error
	claimedEmployeeSchedules []db.ClaimDueEmployeeSchedulesRow
	claimEmployeeErr         error
	resetCalls               int
	resetEmployeeCalls       int
	lastClaimRunTime         string
	lastClaimValid           bool
}

func (f *fakeSchedulerStore) ClaimDueSchedules(ctx context.Context, arg db.ClaimDueSchedulesParams) ([]db.ClaimDueSchedulesRow, error) {
	f.lastClaimRunTime = arg.RunTime
	f.lastClaimValid = arg.LastEnqueuedForDate.Valid
	if f.claimErr != nil {
		return nil, f.claimErr
	}
	return f.claimedSchedules, nil
}

func (f *fakeSchedulerStore) ResetScheduleEnqueueState(ctx context.Context, arg db.ResetScheduleEnqueueStateParams) error {
	f.resetCalls++
	return nil
}

func (f *fakeSchedulerStore) GetSetting(ctx context.Context, key string) (db.AppSetting, error) {
	if val, ok := f.settings[key]; ok {
		return db.AppSetting{Key: key, Value: val}, nil
	}
	return db.AppSetting{}, errors.New("not found")
}

func (f *fakeSchedulerStore) ClaimDueEmployeeSchedules(_ context.Context, _ db.ClaimDueEmployeeSchedulesParams) ([]db.ClaimDueEmployeeSchedulesRow, error) {
	if f.claimEmployeeErr != nil {
		return nil, f.claimEmployeeErr
	}
	return f.claimedEmployeeSchedules, nil
}

func (f *fakeSchedulerStore) ResetEmployeeScheduleEnqueueState(_ context.Context, _ db.ResetEmployeeScheduleEnqueueStateParams) error {
	f.resetEmployeeCalls++
	return nil
}

type fakeJobRunner struct {
	inserted    int
	skipped     int
	err         error
	calls       int
	runTypes    []string
	createCalls []createCall
	createErr   error
	recovered   int64
	recoverErr  error
}

func (f *fakeJobRunner) RunAll(ctx context.Context, runType string, maxAttempts int32) (inserted, skipped int, err error) {
	f.calls++
	f.runTypes = append(f.runTypes, runType)
	return f.inserted, f.skipped, f.err
}

type createCall struct {
	EmployeeID  pgtype.UUID
	RunType     string
	MaxAttempts int32
	NotBefore   pgtype.Timestamptz
}

func (f *fakeJobRunner) Create(_ context.Context, _ pgtype.UUID, _ string, _ int32) (db.Job, error) {
	return db.Job{}, nil
}

func (f *fakeJobRunner) CreateWithDelay(_ context.Context, employeeID pgtype.UUID, runType string, maxAttempts int32, notBefore pgtype.Timestamptz) (db.Job, error) {
	f.createCalls = append(f.createCalls, createCall{
		EmployeeID:  employeeID,
		RunType:     runType,
		MaxAttempts: maxAttempts,
		NotBefore:   notBefore,
	})
	if f.createErr != nil {
		return db.Job{}, f.createErr
	}
	return db.Job{}, nil
}

func (f *fakeJobRunner) RecoverStaleRunning(_ context.Context, _ time.Duration) (int64, error) {
	return f.recovered, f.recoverErr
}

func TestSchedulerTickProcessesClaimedSchedules(t *testing.T) {
	store := &fakeSchedulerStore{
		settings: map[string]string{"default_max_attempts": "4"},
		claimedSchedules: []db.ClaimDueSchedulesRow{
			{RunType: db.RunTypeEnumMorning},
			{RunType: db.RunTypeEnumCheckout},
		},
	}
	sett := &Setting{q: newFakeStore()}
	jobs := &fakeJobRunner{inserted: 2, skipped: 1}
	svc := &PusakaScheduler{
		store: store,
		jobs:  jobs,
		sett:  sett,
		loc:   time.FixedZone("WITA", 8*60*60),
	}

	now := time.Date(2026, 4, 28, 7, 5, 0, 0, time.UTC)
	got, err := svc.Tick(context.Background(), now)
	if err != nil {
		t.Fatalf("Tick() error = %v", err)
	}
	if got.Processed != 2 || got.Enqueued != 4 || got.Skipped != 2 {
		t.Fatalf("Tick() = %+v", got)
	}
	if jobs.calls != 2 {
		t.Fatalf("RunAll() calls = %d, want 2", jobs.calls)
	}
	if store.lastClaimRunTime == "" || !store.lastClaimValid {
		t.Fatalf("ClaimDueSchedules not called with runtime/date")
	}
}

func TestSchedulerTickRecoversStaleRunningJobsBeforeScheduleClaim(t *testing.T) {
	store := &fakeSchedulerStore{settings: map[string]string{}}
	jobs := &fakeJobRunner{recoverErr: errors.New("recover failed")}
	svc := &PusakaScheduler{
		store: store,
		jobs:  jobs,
		loc:   time.FixedZone("WITA", 8*60*60),
	}

	_, err := svc.Tick(context.Background(), time.Date(2026, 4, 28, 7, 5, 0, 0, time.UTC))
	if err == nil || err.Error() != "recover failed" {
		t.Fatalf("Tick() error = %v, want recover failed", err)
	}
	if store.lastClaimRunTime != "" {
		t.Fatalf("ClaimDueSchedules ran before stale recovery succeeded")
	}
}

func TestSchedulerTickResetsClaimOnRunAllError(t *testing.T) {
	store := &fakeSchedulerStore{
		settings: map[string]string{"default_max_attempts": "3"},
		claimedSchedules: []db.ClaimDueSchedulesRow{
			{ID: pgtype.UUID{Valid: true}, RunType: db.RunTypeEnumMorning},
		},
	}
	sett := &Setting{q: newFakeStore()}
	jobs := &fakeJobRunner{err: errors.New("run failed")}
	svc := &PusakaScheduler{
		store: store,
		jobs:  jobs,
		sett:  sett,
		loc:   time.FixedZone("WITA", 8*60*60),
	}

	_, err := svc.Tick(context.Background(), time.Date(2026, 4, 28, 7, 0, 0, 0, time.UTC))
	if err == nil {
		t.Fatalf("Tick() error = nil, want error")
	}
	if store.resetCalls != 1 {
		t.Fatalf("ResetScheduleEnqueueState() calls = %d, want 1", store.resetCalls)
	}
}

func TestSchedulerTickEmployeeSchedulesWithoutRandomWindowStayImmediate(t *testing.T) {
	store := &fakeSchedulerStore{
		settings: map[string]string{"default_max_attempts": "3"},
		claimedEmployeeSchedules: []db.ClaimDueEmployeeSchedulesRow{{
			ID:                  pgtype.UUID{Bytes: [16]byte{9}, Valid: true},
			EmployeeID:          pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			RunType:             db.RunTypeEnumCheckin,
			RandomWindowMinutes: 0,
		}},
	}
	sett := &Setting{q: newFakeStore()}
	jobs := &fakeJobRunner{}
	svc := &PusakaScheduler{
		store: store,
		jobs:  jobs,
		sett:  sett,
		loc:   time.FixedZone("WITA", 8*60*60),
	}

	now := time.Date(2026, 4, 28, 7, 0, 0, 0, time.UTC)
	got, err := svc.Tick(context.Background(), now)
	if err != nil {
		t.Fatalf("Tick() error = %v", err)
	}
	if got.Processed != 1 || got.Enqueued != 1 || got.Skipped != 0 {
		t.Fatalf("Tick() = %+v", got)
	}
	if len(jobs.createCalls) != 1 {
		t.Fatalf("CreateWithDelay() calls = %d, want 1", len(jobs.createCalls))
	}
	if jobs.createCalls[0].NotBefore.Valid {
		t.Fatalf("CreateWithDelay() not_before = %+v, want zero/invalid for immediate schedules", jobs.createCalls[0].NotBefore)
	}
	if store.resetEmployeeCalls != 0 {
		t.Fatalf("ResetEmployeeScheduleEnqueueState() calls = %d, want 0", store.resetEmployeeCalls)
	}
}

func TestSchedulerTickEmployeeSchedulesRandomWindowSetsNotBefore(t *testing.T) {
	store := &fakeSchedulerStore{
		settings: map[string]string{"default_max_attempts": "3"},
		claimedEmployeeSchedules: []db.ClaimDueEmployeeSchedulesRow{{
			ID:                  pgtype.UUID{Bytes: [16]byte{7}, Valid: true},
			EmployeeID:          pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
			RunType:             db.RunTypeEnumCheckout,
			RandomWindowMinutes: 15,
		}},
	}
	sett := &Setting{q: newFakeStore()}
	jobs := &fakeJobRunner{}
	svc := &PusakaScheduler{
		store: store,
		jobs:  jobs,
		sett:  sett,
		loc:   time.FixedZone("WITA", 8*60*60),
	}

	now := time.Date(2026, 4, 28, 7, 0, 0, 0, time.UTC)
	got, err := svc.Tick(context.Background(), now)
	if err != nil {
		t.Fatalf("Tick() error = %v", err)
	}
	if got.Processed != 1 || got.Enqueued != 1 || got.Skipped != 0 {
		t.Fatalf("Tick() = %+v", got)
	}
	if len(jobs.createCalls) != 1 {
		t.Fatalf("CreateWithDelay() calls = %d, want 1", len(jobs.createCalls))
	}
	call := jobs.createCalls[0]
	if !call.NotBefore.Valid {
		t.Fatalf("CreateWithDelay() not_before should be set when random window is enabled")
	}
	diff := call.NotBefore.Time.Sub(now)
	if diff < 0 || diff > 15*time.Minute {
		t.Fatalf("CreateWithDelay() not_before diff = %v, want 0..15m", diff)
	}
}
