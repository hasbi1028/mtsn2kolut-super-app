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
	settings         map[string]string
	claimedSchedules []db.ClaimDueSchedulesRow
	claimErr         error
	resetCalls       int
	lastClaimRunTime string
	lastClaimValid   bool
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
	return nil, nil
}

func (f *fakeSchedulerStore) ResetEmployeeScheduleEnqueueState(_ context.Context, _ db.ResetEmployeeScheduleEnqueueStateParams) error {
	return nil
}

type fakeJobRunner struct {
	inserted int
	skipped  int
	err      error
	calls    int
	runTypes []string
}

func (f *fakeJobRunner) RunAll(ctx context.Context, runType string, maxAttempts int32) (inserted, skipped int, err error) {
	f.calls++
	f.runTypes = append(f.runTypes, runType)
	return f.inserted, f.skipped, f.err
}

func (f *fakeJobRunner) Create(_ context.Context, _ pgtype.UUID, _ string, _ int32) (db.Job, error) {
	return db.Job{}, nil
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
	svc := &Scheduler{
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

func TestSchedulerTickResetsClaimOnRunAllError(t *testing.T) {
	store := &fakeSchedulerStore{
		settings: map[string]string{"default_max_attempts": "3"},
		claimedSchedules: []db.ClaimDueSchedulesRow{
			{ID: pgtype.UUID{Valid: true}, RunType: db.RunTypeEnumMorning},
		},
	}
	sett := &Setting{q: newFakeStore()}
	jobs := &fakeJobRunner{err: errors.New("run failed")}
	svc := &Scheduler{
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
