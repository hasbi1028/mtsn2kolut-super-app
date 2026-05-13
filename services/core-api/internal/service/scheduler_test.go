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
	claimCh                  chan struct{}
	resetCalls               int
	resetEmployeeCalls       int
	lastClaimRunTime         string
	lastClaimValid           bool
	reminderRows             int64
	reminderErr              error
	lastReminderDate         pgtype.Date
	notificationRows         int64
	notificationErr          error
	lastNotificationDate     pgtype.Date
}

func (f *fakeSchedulerStore) ClaimDueSchedules(ctx context.Context, arg db.ClaimDueSchedulesParams) ([]db.ClaimDueSchedulesRow, error) {
	f.lastClaimRunTime = arg.RunTime
	f.lastClaimValid = arg.LastEnqueuedForDate.Valid
	if f.claimCh != nil {
		select {
		case f.claimCh <- struct{}{}:
		default:
		}
	}
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

func (f *fakeSchedulerStore) CreateDocumentCycleReminderEvents(_ context.Context, today pgtype.Date) (int64, error) {
	f.lastReminderDate = today
	if f.reminderErr != nil {
		return 0, f.reminderErr
	}
	return f.reminderRows, nil
}

func (f *fakeSchedulerStore) CreateDocumentCycleReminderNotifications(_ context.Context, today pgtype.Date) (int64, error) {
	f.lastNotificationDate = today
	if f.notificationErr != nil {
		return 0, f.notificationErr
	}
	return f.notificationRows, nil
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

type fakeAuditCleaner struct {
	deleted int64
	err     error
	calls   int
}

func (f *fakeAuditCleaner) CleanupOld(ctx context.Context) (int64, error) {
	f.calls++
	if f.err != nil {
		return 0, f.err
	}
	return f.deleted, nil
}

func TestNewPusakaSchedulerAndStop(t *testing.T) {
	svc := NewPusakaScheduler(nil, nil, nil, nil)
	if svc == nil {
		t.Fatal("NewPusakaScheduler() = nil")
	}
	if svc.loc == nil {
		t.Fatal("NewPusakaScheduler() location = nil")
	}

	cancelled := false
	svc.cancel = func() { cancelled = true }
	svc.Stop()
	if !cancelled {
		t.Fatal("Stop() did not call cancel")
	}
}

func TestSchedulerStartRunsImmediateTickAndStopCancels(t *testing.T) {
	claimCh := make(chan struct{}, 1)
	store := &fakeSchedulerStore{
		settings: map[string]string{"default_max_attempts": "3"},
		claimCh:  claimCh,
	}
	svc := &PusakaScheduler{
		store: store,
		jobs:  &fakeJobRunner{},
		loc:   time.FixedZone("WITA", 8*60*60),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc.Start(ctx)
	select {
	case <-claimCh:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Start() did not run the immediate scheduler tick")
	}

	svc.Stop()
	if svc.cancel == nil {
		t.Fatal("Start() did not install cancel function")
	}
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
	if !store.lastReminderDate.Valid {
		t.Fatalf("CreateDocumentCycleReminderEvents not called with local date")
	}
	if !store.lastNotificationDate.Valid {
		t.Fatalf("CreateDocumentCycleReminderNotifications not called with local date")
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

func TestSchedulerEmployeeScheduleWindow(t *testing.T) {
	loc := time.FixedZone("WITA", 8*60*60)

	t.Run("builds not before from schedule base", func(t *testing.T) {
		localNow := time.Date(2026, 5, 13, 14, 40, 0, 0, loc)
		window, err := buildEmployeeScheduleWindow(localNow, "14:33", 28, 15, 3*time.Minute, loc)
		if err != nil {
			t.Fatalf("buildEmployeeScheduleWindow() error = %v", err)
		}
		wantNotBefore := time.Date(2026, 5, 13, 14, 48, 0, 0, loc)
		wantLatest := time.Date(2026, 5, 13, 15, 4, 0, 0, loc)
		if !window.NotBefore.Equal(wantNotBefore) {
			t.Fatalf("not_before = %v, want %v", window.NotBefore, wantNotBefore)
		}
		if !window.Latest.Equal(wantLatest) {
			t.Fatalf("latest = %v, want %v", window.Latest, wantLatest)
		}
		if window.Expired {
			t.Fatal("window marked expired before latest boundary")
		}
	})

	t.Run("expires late clock jumps", func(t *testing.T) {
		localNow := time.Date(2026, 5, 13, 16, 39, 0, 0, loc)
		window, err := buildEmployeeScheduleWindow(localNow, "14:33", 28, 0, 3*time.Minute, loc)
		if err != nil {
			t.Fatalf("buildEmployeeScheduleWindow() error = %v", err)
		}
		if !window.Expired {
			t.Fatalf("window expired = false, want true for now=%v latest=%v", localNow, window.Latest)
		}
	})
}

func TestSchedulerTickEmployeeSchedulesWithoutRandomWindowStayImmediate(t *testing.T) {
	store := &fakeSchedulerStore{
		settings: map[string]string{"default_max_attempts": "3"},
		claimedEmployeeSchedules: []db.ClaimDueEmployeeSchedulesRow{{
			ID:                  pgtype.UUID{Bytes: [16]byte{9}, Valid: true},
			EmployeeID:          pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			RunType:             db.RunTypeEnumCheckin,
			RunTime:             "15:00",
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
			RunTime:             "14:50",
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
	base := time.Date(2026, 4, 28, 14, 50, 0, 0, time.FixedZone("WITA", 8*60*60))
	diff := call.NotBefore.Time.Sub(base)
	if diff < 0 || diff > 15*time.Minute {
		t.Fatalf("CreateWithDelay() not_before diff from schedule base = %v, want 0..15m", diff)
	}
}

func TestSchedulerTickEmployeeSchedulesExpiredWindowSkipsJob(t *testing.T) {
	store := &fakeSchedulerStore{
		settings: map[string]string{"default_max_attempts": "3"},
		claimedEmployeeSchedules: []db.ClaimDueEmployeeSchedulesRow{{
			ID:                  pgtype.UUID{Bytes: [16]byte{8}, Valid: true},
			EmployeeID:          pgtype.UUID{Bytes: [16]byte{3}, Valid: true},
			RunType:             db.RunTypeEnumCheckout,
			RunTime:             "14:33",
			RandomWindowMinutes: 28,
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

	now := time.Date(2026, 5, 13, 8, 39, 0, 0, time.UTC) // 16:39 WITA
	got, err := svc.Tick(context.Background(), now)
	if err != nil {
		t.Fatalf("Tick() error = %v", err)
	}
	if got.Processed != 1 || got.Enqueued != 0 || got.Skipped != 1 {
		t.Fatalf("Tick() = %+v, want processed=1 enqueued=0 skipped=1", got)
	}
	if len(jobs.createCalls) != 0 {
		t.Fatalf("CreateWithDelay() calls = %d, want 0 for expired window", len(jobs.createCalls))
	}
}

func TestSchedulerTickReportsDocumentCycleReminders(t *testing.T) {
	store := &fakeSchedulerStore{
		settings:         map[string]string{"default_max_attempts": "3"},
		reminderRows:     3,
		notificationRows: 2,
	}
	jobs := &fakeJobRunner{}
	svc := &PusakaScheduler{
		store: store,
		jobs:  jobs,
		loc:   time.FixedZone("WITA", 8*60*60),
	}

	got, err := svc.Tick(context.Background(), time.Date(2026, 4, 28, 7, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Tick() error = %v", err)
	}
	if got.DocumentReminders != 3 {
		t.Fatalf("Tick() document reminders = %d, want 3", got.DocumentReminders)
	}
	if got.DocumentNotifications != 2 {
		t.Fatalf("Tick() document notifications = %d, want 2", got.DocumentNotifications)
	}
}

func TestSchedulerTickHandlesClaimAndDocumentCycleErrors(t *testing.T) {
	now := time.Date(2026, 4, 28, 7, 0, 0, 0, time.UTC)

	t.Run("schedule claim error stops tick", func(t *testing.T) {
		expectedErr := errors.New("claim failed")
		store := &fakeSchedulerStore{
			settings: map[string]string{"default_max_attempts": "3"},
			claimErr: expectedErr,
		}
		svc := &PusakaScheduler{
			store: store,
			jobs:  &fakeJobRunner{},
			loc:   time.FixedZone("WITA", 8*60*60),
		}

		_, err := svc.Tick(context.Background(), now)
		if !errors.Is(err, expectedErr) {
			t.Fatalf("Tick() error = %v, want %v", err, expectedErr)
		}
	})

	t.Run("employee claim error does not stop tick", func(t *testing.T) {
		store := &fakeSchedulerStore{
			settings:         map[string]string{"default_max_attempts": "3"},
			claimEmployeeErr: errors.New("employee claim failed"),
		}
		svc := &PusakaScheduler{
			store: store,
			jobs:  &fakeJobRunner{},
			loc:   time.FixedZone("WITA", 8*60*60),
		}

		got, err := svc.Tick(context.Background(), now)
		if err != nil {
			t.Fatalf("Tick() error = %v, want nil for employee claim failure", err)
		}
		if got.Processed != 0 || got.Enqueued != 0 {
			t.Fatalf("Tick() = %+v, want no employee jobs enqueued", got)
		}
	})

	t.Run("employee job create error resets schedule and continues", func(t *testing.T) {
		store := &fakeSchedulerStore{
			settings: map[string]string{"default_max_attempts": "3"},
			claimedEmployeeSchedules: []db.ClaimDueEmployeeSchedulesRow{{
				ID:         pgtype.UUID{Bytes: [16]byte{11}, Valid: true},
				EmployeeID: pgtype.UUID{Bytes: [16]byte{12}, Valid: true},
				RunType:    db.RunTypeEnumCheckin,
			}},
		}
		svc := &PusakaScheduler{
			store: store,
			jobs:  &fakeJobRunner{createErr: errors.New("create failed")},
			loc:   time.FixedZone("WITA", 8*60*60),
		}

		got, err := svc.Tick(context.Background(), now)
		if err != nil {
			t.Fatalf("Tick() error = %v", err)
		}
		if got.Processed != 1 || got.Enqueued != 0 || store.resetEmployeeCalls != 1 {
			t.Fatalf("Tick() = %+v resetEmployeeCalls=%d, want processed reset without enqueue", got, store.resetEmployeeCalls)
		}
	})

	t.Run("document cycle errors are recorded but do not stop tick", func(t *testing.T) {
		settingStore := newFakeStore()
		store := &fakeSchedulerStore{
			settings:        map[string]string{"default_max_attempts": "3"},
			reminderErr:     errors.New("reminder failed"),
			notificationErr: errors.New("notification failed"),
		}
		svc := &PusakaScheduler{
			store: store,
			jobs:  &fakeJobRunner{},
			sett:  &Setting{q: settingStore},
			loc:   time.FixedZone("WITA", 8*60*60),
		}

		got, err := svc.Tick(context.Background(), now)
		if err != nil {
			t.Fatalf("Tick() error = %v", err)
		}
		if got.DocumentReminders != 0 || got.DocumentNotifications != 0 {
			t.Fatalf("Tick() reminders/notifications = %d/%d, want zero on errors", got.DocumentReminders, got.DocumentNotifications)
		}
		if settingStore.settings["scheduler_document_cycle_reminder_error"] != "reminder failed" {
			t.Fatalf("reminder status = %q, want reminder failed", settingStore.settings["scheduler_document_cycle_reminder_error"])
		}
		if settingStore.settings["scheduler_document_cycle_notification_error"] != "notification failed" {
			t.Fatalf("notification status = %q, want notification failed", settingStore.settings["scheduler_document_cycle_notification_error"])
		}
	})
}

func TestSchedulerDefaultMaxAttemptsBounds(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  int32
	}{
		{name: "missing", value: "", want: defaultMaxAttempts},
		{name: "non numeric", value: "bad", want: defaultMaxAttempts},
		{name: "too low", value: "0", want: defaultMaxAttempts},
		{name: "too high", value: "11", want: defaultMaxAttempts},
		{name: "valid", value: "7", want: 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := map[string]string{}
			if tt.value != "" {
				settings["default_max_attempts"] = tt.value
			}
			svc := &PusakaScheduler{store: &fakeSchedulerStore{settings: settings}}
			if got := svc.defaultMaxAttempts(context.Background()); got != tt.want {
				t.Fatalf("defaultMaxAttempts() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSchedulerRunTickExecutesTickAndAuditGate(t *testing.T) {
	store := &fakeSchedulerStore{settings: map[string]string{"default_max_attempts": "3"}}
	svc := &PusakaScheduler{
		store: store,
		jobs:  &fakeJobRunner{},
		loc:   time.FixedZone("WITA", 8*60*60),
	}

	svc.runTick(context.Background())
	if store.lastClaimRunTime == "" {
		t.Fatal("runTick() did not execute Tick()")
	}
}

func TestSchedulerRunTickLogsProcessedResult(t *testing.T) {
	store := &fakeSchedulerStore{
		settings: map[string]string{"default_max_attempts": "3"},
		claimedSchedules: []db.ClaimDueSchedulesRow{
			{ID: documentCycleTestUUID(71), RunType: db.RunTypeEnumCheckin},
		},
	}
	jobs := &fakeJobRunner{inserted: 1}
	svc := &PusakaScheduler{
		store: store,
		jobs:  jobs,
		loc:   time.FixedZone("WITA", 8*60*60),
	}

	svc.runTick(context.Background())
	if jobs.calls != 1 {
		t.Fatalf("RunAll() calls = %d, want 1", jobs.calls)
	}
}

func TestSchedulerRunTickHandlesTickError(t *testing.T) {
	store := &fakeSchedulerStore{
		settings: map[string]string{"default_max_attempts": "3"},
		claimErr: errors.New("claim failed"),
	}
	svc := &PusakaScheduler{
		store: store,
		jobs:  &fakeJobRunner{},
		loc:   time.FixedZone("WITA", 8*60*60),
	}

	svc.runTick(context.Background())
	if store.lastClaimRunTime == "" {
		t.Fatal("runTick() did not attempt Tick() before handling error")
	}
}

func TestSchedulerSetStatusHandlesNilAndUpsertError(t *testing.T) {
	svc := &PusakaScheduler{}
	if err := svc.setStatus(context.Background(), map[string]string{"scheduler_last_error": ""}); err != nil {
		t.Fatalf("setStatus(nil setting) error = %v, want nil", err)
	}

	expectedErr := errors.New("setting write failed")
	svc = &PusakaScheduler{sett: &Setting{q: failingSettingStore{upsertErr: expectedErr}}}
	if err := svc.setStatus(context.Background(), map[string]string{"scheduler_last_error": "failed"}); !errors.Is(err, expectedErr) {
		t.Fatalf("setStatus() error = %v, want %v", err, expectedErr)
	}
}

func TestSchedulerMaybeCleanupAudit(t *testing.T) {
	t.Run("skips when audit or settings service is missing", func(t *testing.T) {
		store := &fakeSchedulerStore{settings: map[string]string{}}
		audit := &fakeAuditCleaner{}
		svc := &PusakaScheduler{
			store: store,
			loc:   time.FixedZone("WITA", 8*60*60),
			audit: audit,
		}

		svc.maybeCleanupAudit(context.Background())
		if audit.calls != 0 {
			t.Fatalf("CleanupOld() calls = %d, want 0 without settings service", audit.calls)
		}
	})

	t.Run("skips inside cleanup interval", func(t *testing.T) {
		audit := &fakeAuditCleaner{}
		svc := &PusakaScheduler{
			store:       &fakeSchedulerStore{settings: map[string]string{}},
			sett:        &Setting{q: newFakeStore()},
			loc:         time.FixedZone("WITA", 8*60*60),
			audit:       audit,
			lastCleanup: time.Now(),
		}

		svc.maybeCleanupAudit(context.Background())
		if audit.calls != 0 {
			t.Fatalf("CleanupOld() calls = %d, want 0 inside interval", audit.calls)
		}
	})

	t.Run("skips when already cleaned today", func(t *testing.T) {
		loc := time.FixedZone("WITA", 8*60*60)
		today := time.Now().In(loc).Format(time.DateOnly)
		audit := &fakeAuditCleaner{}
		svc := &PusakaScheduler{
			store: &fakeSchedulerStore{settings: map[string]string{"last_audit_cleanup_date": today}},
			sett:  &Setting{q: newFakeStore()},
			loc:   loc,
			audit: audit,
		}

		svc.maybeCleanupAudit(context.Background())
		if audit.calls != 0 {
			t.Fatalf("CleanupOld() calls = %d, want 0 when already cleaned today", audit.calls)
		}
		if svc.lastCleanup.IsZero() {
			t.Fatal("lastCleanup was not updated after same-day cleanup setting")
		}
	})

	t.Run("cleans and records cleanup date", func(t *testing.T) {
		settingStore := newFakeStore()
		audit := &fakeAuditCleaner{deleted: 4}
		svc := &PusakaScheduler{
			store: &fakeSchedulerStore{settings: map[string]string{}},
			sett:  &Setting{q: settingStore},
			loc:   time.FixedZone("WITA", 8*60*60),
			audit: audit,
		}

		svc.maybeCleanupAudit(context.Background())
		if audit.calls != 1 {
			t.Fatalf("CleanupOld() calls = %d, want 1", audit.calls)
		}
		if settingStore.settings["last_audit_cleanup_date"] == "" {
			t.Fatalf("last_audit_cleanup_date was not recorded: %+v", settingStore.settings)
		}
		if svc.lastCleanup.IsZero() {
			t.Fatal("lastCleanup was not updated after cleanup")
		}
	})

	t.Run("does not record date when cleanup fails", func(t *testing.T) {
		settingStore := newFakeStore()
		audit := &fakeAuditCleaner{err: errors.New("cleanup failed")}
		svc := &PusakaScheduler{
			store: &fakeSchedulerStore{settings: map[string]string{}},
			sett:  &Setting{q: settingStore},
			loc:   time.FixedZone("WITA", 8*60*60),
			audit: audit,
		}

		svc.maybeCleanupAudit(context.Background())
		if audit.calls != 1 {
			t.Fatalf("CleanupOld() calls = %d, want 1", audit.calls)
		}
		if settingStore.settings["last_audit_cleanup_date"] != "" {
			t.Fatalf("last_audit_cleanup_date = %q, want empty on cleanup failure", settingStore.settings["last_audit_cleanup_date"])
		}
		if !svc.lastCleanup.IsZero() {
			t.Fatalf("lastCleanup = %v, want zero on cleanup failure", svc.lastCleanup)
		}
	})
}
