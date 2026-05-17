package db

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestIntegrationNotificationsSettingsRepository(t *testing.T) {
	t.Parallel()

	tdb := setupIntegrationTestDB(t)
	ctx := context.Background()
	q := tdb.Q
	suffix := integrationSuffix()

	user, err := q.CreateUserWithMustChangePassword(ctx, CreateUserWithMustChangePasswordParams{
		Username:           "it-notify-" + suffix,
		PasswordHash:       "hash:v1:" + suffix,
		DisplayName:        pgtype.Text{String: "Integration Notify " + suffix, Valid: true},
		IsActive:           true,
		MustChangePassword: false,
	})
	if err != nil {
		t.Fatalf("create notification user: %v", err)
	}

	firstNotificationID := randomTestUUID(t)
	insertNotification(t, ctx, tdb, firstNotificationID, user.ID, "system", "First", "first body", "system", suffix+"-first", "/system", "it-notify:"+suffix+":first", nil)
	readAt := time.Now().Add(-time.Minute)
	insertNotification(t, ctx, tdb, randomTestUUID(t), user.ID, "system", "Already read", "read body", "system", suffix+"-read", "/system", "it-notify:"+suffix+":read", &readAt)

	unreadCount, err := q.CountUnreadAppNotifications(ctx, user.ID)
	if err != nil {
		t.Fatalf("count unread app notifications: %v", err)
	}
	if unreadCount != 1 {
		t.Fatalf("unread app notifications = %d, want 1", unreadCount)
	}

	unread, err := q.ListAppNotifications(ctx, ListAppNotificationsParams{UserID: user.ID, UnreadOnly: true, LimitCount: 10})
	if err != nil {
		t.Fatalf("list unread app notifications: %v", err)
	}
	if len(unread) != 1 || unread[0].ID != firstNotificationID || unread[0].ReadAt.Valid {
		t.Fatalf("unread notifications = %#v, want only first unread notification", unread)
	}

	allNotifications, err := q.ListAppNotifications(ctx, ListAppNotificationsParams{UserID: user.ID, LimitCount: 10})
	if err != nil {
		t.Fatalf("list all app notifications: %v", err)
	}
	if len(allNotifications) != 2 {
		t.Fatalf("all notifications len = %d, want 2", len(allNotifications))
	}

	marked, err := q.MarkAppNotificationRead(ctx, MarkAppNotificationReadParams{ID: firstNotificationID, UserID: user.ID})
	if err != nil {
		t.Fatalf("mark app notification read: %v", err)
	}
	if marked.ID != firstNotificationID || !marked.ReadAt.Valid {
		t.Fatalf("marked notification = %#v, want matching id with read_at", marked)
	}

	unreadCount, err = q.CountUnreadAppNotifications(ctx, user.ID)
	if err != nil {
		t.Fatalf("count unread app notifications after mark: %v", err)
	}
	if unreadCount != 0 {
		t.Fatalf("unread app notifications after mark = %d, want 0", unreadCount)
	}

	insertNotification(t, ctx, tdb, randomTestUUID(t), user.ID, "system", "Second unread", "second body", "system", suffix+"-second", "/system", "it-notify:"+suffix+":second", nil)
	markedRows, err := q.MarkAllAppNotificationsRead(ctx, user.ID)
	if err != nil {
		t.Fatalf("mark all app notifications read: %v", err)
	}
	if markedRows != 1 {
		t.Fatalf("mark all app notifications rows = %d, want 1", markedRows)
	}

	settingKeyA := "it.notifications." + suffix + ".a"
	settingKeyB := "it.notifications." + suffix + ".b"
	if err := q.UpsertSetting(ctx, UpsertSettingParams{Key: settingKeyB, Value: "value-b"}); err != nil {
		t.Fatalf("upsert setting b: %v", err)
	}
	if err := q.UpsertSetting(ctx, UpsertSettingParams{Key: settingKeyA, Value: "value-a"}); err != nil {
		t.Fatalf("upsert setting a: %v", err)
	}
	if err := q.UpsertSetting(ctx, UpsertSettingParams{Key: settingKeyA, Value: "value-a-updated"}); err != nil {
		t.Fatalf("update setting a: %v", err)
	}

	settingA, err := q.GetSetting(ctx, settingKeyA)
	if err != nil {
		t.Fatalf("get setting a: %v", err)
	}
	if settingA.Value != "value-a-updated" || !settingA.UpdatedAt.Valid {
		t.Fatalf("setting a = %#v, want updated value with updated_at", settingA)
	}

	settings, err := q.ListSettings(ctx)
	if err != nil {
		t.Fatalf("list settings: %v", err)
	}
	indexA := settingIndex(settings, settingKeyA)
	indexB := settingIndex(settings, settingKeyB)
	if indexA == -1 || indexB == -1 {
		t.Fatalf("settings did not include inserted keys %q/%q", settingKeyA, settingKeyB)
	}
	if indexA > indexB {
		t.Fatalf("settings order indexA=%d indexB=%d, want key order", indexA, indexB)
	}
}

func TestIntegrationJobsRepository(t *testing.T) {
	t.Parallel()

	tdb := setupIntegrationTestDB(t)
	ctx := context.Background()
	q := tdb.Q
	suffix := integrationSuffix()
	employeeID := createIntegrationEmployee(t, ctx, q, suffix)

	account, err := q.UpsertPusakaAccount(ctx, UpsertPusakaAccountParams{
		EmployeeID:     employeeID,
		PusakaUsername: "pusaka-" + suffix,
		PusakaPassword: "secret-" + suffix,
		IsEnabled:      true,
	})
	if err != nil {
		t.Fatalf("upsert pusaka account: %v", err)
	}
	if account.EmployeeID != employeeID || !account.IsEnabled {
		t.Fatalf("pusaka account = %#v, want enabled account for employee %v", account, employeeID)
	}

	job, err := q.CreateJobIfAbsent(ctx, CreateJobIfAbsentParams{
		EmployeeID:  employeeID,
		RunType:     RunTypeEnumCheckin,
		MaxAttempts: 3,
		NotBefore:   pgtype.Timestamptz{Time: time.Now().Add(-time.Minute), Valid: true},
	})
	if err != nil {
		t.Fatalf("create job if absent: %v", err)
	}
	if job.Status != JobStatusEnumQueued || job.Attempts != 0 || job.EmployeeID != employeeID {
		t.Fatalf("created job = %#v", job)
	}

	if _, err := q.CreateJobIfAbsent(ctx, CreateJobIfAbsentParams{EmployeeID: employeeID, RunType: RunTypeEnumCheckin, MaxAttempts: 3}); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("duplicate active job err = %v, want pgx.ErrNoRows", err)
	}

	queuedCount, err := q.CountJobsByStatus(ctx, JobStatusEnumQueued)
	if err != nil {
		t.Fatalf("count queued jobs: %v", err)
	}
	if queuedCount != 1 {
		t.Fatalf("queued jobs = %d, want 1", queuedCount)
	}

	claimed, err := q.ClaimJob(ctx, "worker-a")
	if err != nil {
		t.Fatalf("claim job: %v", err)
	}
	if claimed.ID != job.ID || claimed.Attempts != 1 || claimed.PusakaUsername != "pusaka-"+suffix || claimed.PusakaPassword != "secret-"+suffix {
		t.Fatalf("claimed job = %#v", claimed)
	}

	running, err := q.GetRunningJobForWorker(ctx, GetRunningJobForWorkerParams{ID: job.ID, ClaimedBy: "worker-a"})
	if err != nil {
		t.Fatalf("get running job for worker: %v", err)
	}
	if running.Status != JobStatusEnumRunning || running.ClaimedBy != "worker-a" || !running.ClaimedAt.Valid {
		t.Fatalf("running job = %#v", running)
	}

	wrongCompleteRows, err := q.CompleteJob(ctx, CompleteJobParams{ID: job.ID, ClaimedBy: "worker-b"})
	if err != nil {
		t.Fatalf("complete job with wrong worker: %v", err)
	}
	if wrongCompleteRows != 0 {
		t.Fatalf("wrong worker complete rows = %d, want 0", wrongCompleteRows)
	}
	completeRows, err := q.CompleteJob(ctx, CompleteJobParams{ID: job.ID, ClaimedBy: "worker-a"})
	if err != nil {
		t.Fatalf("complete job: %v", err)
	}
	if completeRows != 1 {
		t.Fatalf("complete job rows = %d, want 1", completeRows)
	}
	completed, err := q.GetJob(ctx, job.ID)
	if err != nil {
		t.Fatalf("get completed job: %v", err)
	}
	if completed.Status != JobStatusEnumSuccess || completed.ClaimedBy != "" || completed.ClaimedAt.Valid {
		t.Fatalf("completed job = %#v", completed)
	}

	staleJob, err := q.CreateJobIfAbsent(ctx, CreateJobIfAbsentParams{EmployeeID: employeeID, RunType: RunTypeEnumCheckout, MaxAttempts: 2})
	if err != nil {
		t.Fatalf("create stale job: %v", err)
	}
	staleClaimed, err := q.ClaimJob(ctx, "worker-stale")
	if err != nil {
		t.Fatalf("claim stale job: %v", err)
	}
	if staleClaimed.ID != staleJob.ID {
		t.Fatalf("claimed stale job id = %v, want %v", staleClaimed.ID, staleJob.ID)
	}
	if _, err := tdb.Pool.Exec(ctx, `UPDATE jobs SET claimed_at = NOW() - INTERVAL '2 hours' WHERE id = $1`, staleJob.ID); err != nil {
		t.Fatalf("age claimed stale job: %v", err)
	}
	recoveredRows, err := q.RecoverStaleRunningJobs(ctx, 60)
	if err != nil {
		t.Fatalf("recover stale running jobs: %v", err)
	}
	if recoveredRows != 1 {
		t.Fatalf("recover stale running jobs rows = %d, want 1", recoveredRows)
	}
	recovered, err := q.GetJob(ctx, staleJob.ID)
	if err != nil {
		t.Fatalf("get recovered stale job: %v", err)
	}
	if recovered.Status != JobStatusEnumFailed || recovered.ClaimedBy != "" || !recovered.NextRetryAt.Valid {
		t.Fatalf("recovered stale job = %#v", recovered)
	}

	cancelJob, err := q.CreateJobIfAbsent(ctx, CreateJobIfAbsentParams{EmployeeID: employeeID, RunType: RunTypeEnumMorning, MaxAttempts: 1})
	if err != nil {
		t.Fatalf("create cancel job: %v", err)
	}
	cancelRows, err := q.CancelEmployeeJobs(ctx, employeeID)
	if err != nil {
		t.Fatalf("cancel employee jobs: %v", err)
	}
	if cancelRows != 1 {
		t.Fatalf("cancel employee jobs rows = %d, want 1", cancelRows)
	}
	cancelled, err := q.GetJob(ctx, cancelJob.ID)
	if err != nil {
		t.Fatalf("get cancelled job: %v", err)
	}
	if cancelled.Status != JobStatusEnumFailed || cancelled.ErrorMessage != "Dibatalkan manual" {
		t.Fatalf("cancelled job = %#v", cancelled)
	}

	allJobs, err := q.ListJobs(ctx, ListJobsParams{Limit: 10})
	if err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if !listJobsContains(allJobs, job.ID) || !listJobsContains(allJobs, staleJob.ID) || !listJobsContains(allJobs, cancelJob.ID) {
		t.Fatalf("list jobs missing one of created jobs: %#v", allJobs)
	}
	failedJobs, err := q.ListJobsByStatus(ctx, ListJobsByStatusParams{Status: JobStatusEnumFailed, Limit: 10})
	if err != nil {
		t.Fatalf("list failed jobs: %v", err)
	}
	if !listJobsByStatusContains(failedJobs, staleJob.ID) || !listJobsByStatusContains(failedJobs, cancelJob.ID) {
		t.Fatalf("failed jobs missing recovered/cancelled jobs: %#v", failedJobs)
	}

	stats, err := q.GetJobStats(ctx)
	if err != nil {
		t.Fatalf("get job stats: %v", err)
	}
	if stats.Success != 1 || stats.Failed != 2 || stats.Queued != 0 || stats.Running != 0 {
		t.Fatalf("job stats = %#v, want success=1 failed=2 queued=0 running=0", stats)
	}
	totalJobs, err := q.CountJobs(ctx)
	if err != nil {
		t.Fatalf("count jobs: %v", err)
	}
	if totalJobs != 3 {
		t.Fatalf("total jobs = %d, want 3", totalJobs)
	}
}

func TestIntegrationSystemMaintenanceRepository(t *testing.T) {
	t.Parallel()

	tdb := setupIntegrationTestDB(t)
	ctx := context.Background()
	q := tdb.Q
	suffix := integrationSuffix()

	actor, err := q.CreateUserWithMustChangePassword(ctx, CreateUserWithMustChangePasswordParams{
		Username:           "it-maint-" + suffix,
		PasswordHash:       "hash:v1:" + suffix,
		DisplayName:        pgtype.Text{String: "Integration Maintenance " + suffix, Valid: true},
		IsActive:           true,
		MustChangePassword: false,
	})
	if err != nil {
		t.Fatalf("create maintenance actor: %v", err)
	}

	dbTime, err := q.GetMaintenanceDatabaseTime(ctx)
	if err != nil {
		t.Fatalf("get maintenance database time: %v", err)
	}
	if !dbTime.Valid {
		t.Fatalf("maintenance database time is invalid")
	}
	activeSessions, err := q.CountActiveCbtSessionsForMaintenance(ctx)
	if err != nil {
		t.Fatalf("count active cbt sessions for maintenance: %v", err)
	}
	if activeSessions < 0 {
		t.Fatalf("active cbt sessions = %d, want non-negative", activeSessions)
	}

	readOnlyWindow, err := q.CreateMaintenanceWindow(ctx, CreateMaintenanceWindowParams{
		Title:            "Read-only maintenance " + suffix,
		Message:          "read-only message",
		Mode:             "read_only",
		AffectedModules:  []string{"cbt", "pusaka"},
		StartsAt:         pgtype.Timestamptz{Time: time.Now().Add(-time.Minute), Valid: true},
		EndsAt:           pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
		IsActive:         true,
		AllowAdminBypass: true,
		BypassRoles:      []string{"admin"},
		Severity:         "warning",
		ActorUserID:      actor.ID,
	})
	if err != nil {
		t.Fatalf("create read-only maintenance window: %v", err)
	}
	globalWindow, err := q.CreateMaintenanceWindow(ctx, CreateMaintenanceWindowParams{
		Title:            "Global maintenance " + suffix,
		Message:          "global message",
		Mode:             "global",
		AffectedModules:  []string{"global"},
		StartsAt:         pgtype.Timestamptz{Time: time.Now().Add(-time.Minute), Valid: true},
		EndsAt:           pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
		IsActive:         true,
		AllowAdminBypass: false,
		BypassRoles:      []string{"superadmin"},
		Severity:         "critical",
		ActorUserID:      actor.ID,
	})
	if err != nil {
		t.Fatalf("create global maintenance window: %v", err)
	}

	activeWindow, err := q.GetActiveMaintenanceWindow(ctx)
	if err != nil {
		t.Fatalf("get active maintenance window: %v", err)
	}
	if activeWindow.ID != globalWindow.ID || activeWindow.Mode != "global" || activeWindow.Severity != "critical" {
		t.Fatalf("active maintenance window = %#v, want global window %#v", activeWindow, globalWindow)
	}

	gotGlobal, err := q.GetMaintenanceWindow(ctx, globalWindow.ID)
	if err != nil {
		t.Fatalf("get maintenance window: %v", err)
	}
	if gotGlobal.ID != globalWindow.ID || gotGlobal.Title != globalWindow.Title || !slices.Equal(gotGlobal.AffectedModules, []string{"global"}) {
		t.Fatalf("got global maintenance window = %#v", gotGlobal)
	}

	updated, err := q.UpdateMaintenanceWindow(ctx, UpdateMaintenanceWindowParams{
		Title:            "Global maintenance updated " + suffix,
		Message:          "updated message",
		Mode:             "module",
		AffectedModules:  []string{"cbt"},
		StartsAt:         pgtype.Timestamptz{Time: time.Now().Add(-time.Minute), Valid: true},
		EndsAt:           pgtype.Timestamptz{Time: time.Now().Add(2 * time.Hour), Valid: true},
		AllowAdminBypass: true,
		BypassRoles:      []string{"admin", "staf"},
		Severity:         "info",
		ActorUserID:      actor.ID,
		ID:               globalWindow.ID,
	})
	if err != nil {
		t.Fatalf("update maintenance window: %v", err)
	}
	if updated.Title != "Global maintenance updated "+suffix || updated.Mode != "module" || !slices.Equal(updated.BypassRoles, []string{"admin", "staf"}) {
		t.Fatalf("updated maintenance window = %#v", updated)
	}

	deactivated, err := q.SetMaintenanceWindowActive(ctx, SetMaintenanceWindowActiveParams{ID: globalWindow.ID, ActorUserID: actor.ID, IsActive: false})
	if err != nil {
		t.Fatalf("set maintenance window inactive: %v", err)
	}
	if deactivated.IsActive {
		t.Fatalf("deactivated window is still active: %#v", deactivated)
	}
	activeWindow, err = q.GetActiveMaintenanceWindow(ctx)
	if err != nil {
		t.Fatalf("get active maintenance window after deactivating global: %v", err)
	}
	if activeWindow.ID != readOnlyWindow.ID || activeWindow.Mode != "read_only" {
		t.Fatalf("active maintenance window after global deactivate = %#v, want read-only %#v", activeWindow, readOnlyWindow)
	}

	windows, err := q.ListMaintenanceWindows(ctx, ListMaintenanceWindowsParams{LimitCount: 10})
	if err != nil {
		t.Fatalf("list maintenance windows: %v", err)
	}
	if !maintenanceWindowsContains(windows, globalWindow.ID) || !maintenanceWindowsContains(windows, readOnlyWindow.ID) {
		t.Fatalf("maintenance windows missing created windows: %#v", windows)
	}

	audit, err := q.CreateMaintenanceAuditLog(ctx, CreateMaintenanceAuditLogParams{
		MaintenanceID: globalWindow.ID,
		ActorUserID:   actor.ID,
		Action:        "deactivated",
		Reason:        pgtype.Text{String: "integration test", Valid: true},
		Metadata:      []byte(`{"source":"integration"}`),
	})
	if err != nil {
		t.Fatalf("create maintenance audit log: %v", err)
	}
	if audit.MaintenanceID != globalWindow.ID || audit.ActorUserID != actor.ID || audit.Action != "deactivated" || !audit.Reason.Valid {
		t.Fatalf("maintenance audit log = %#v", audit)
	}
	auditLogs, err := q.ListMaintenanceAuditLogs(ctx, ListMaintenanceAuditLogsParams{ActionFilter: "deactivated", LimitCount: 10})
	if err != nil {
		t.Fatalf("list filtered maintenance audit logs: %v", err)
	}
	if !maintenanceAuditLogsContains(auditLogs, audit.ID) {
		t.Fatalf("filtered maintenance audit logs missing audit %v: %#v", audit.ID, auditLogs)
	}
}

func insertNotification(t *testing.T, ctx context.Context, tdb *integrationTestDB, id pgtype.UUID, userID pgtype.UUID, category, title, body, entityType, entityID, linkPath, dedupeKey string, readAt *time.Time) {
	t.Helper()
	var readAtValue any
	if readAt != nil {
		readAtValue = *readAt
	}
	_, err := tdb.Pool.Exec(ctx, `
		INSERT INTO app_notifications (id, user_id, category, title, body, entity_type, entity_id, link_path, dedupe_key, read_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, id, userID, category, title, body, entityType, entityID, linkPath, dedupeKey, readAtValue)
	if err != nil {
		t.Fatalf("insert app notification %q: %v", dedupeKey, err)
	}
}

func createIntegrationEmployee(t *testing.T, ctx context.Context, q *Queries, suffix string) pgtype.UUID {
	t.Helper()
	employeeID, err := q.CreateEmployee(ctx, CreateEmployeeParams{
		Npsn:           "40404224",
		Nip:            "19" + suffix,
		Nama:           "Integration Employee " + suffix,
		UnitKerja:      "MTsN 2 Kolaka Utara",
		EmploymentType: "pns",
		TanggalLahir:   pgDate(1988, time.January, 2),
		JenisKelamin:   "L",
		TempatLahir:    "Kolaka Utara",
		IsActive:       true,
	})
	if err != nil {
		t.Fatalf("create integration employee: %v", err)
	}
	if !employeeID.Valid {
		t.Fatalf("created employee id is invalid")
	}
	return employeeID
}

func settingIndex(settings []AppSetting, key string) int {
	for i, setting := range settings {
		if setting.Key == key {
			return i
		}
	}
	return -1
}

func listJobsContains(jobs []ListJobsRow, id pgtype.UUID) bool {
	return slices.ContainsFunc(jobs, func(job ListJobsRow) bool { return job.ID == id })
}

func listJobsByStatusContains(jobs []ListJobsByStatusRow, id pgtype.UUID) bool {
	return slices.ContainsFunc(jobs, func(job ListJobsByStatusRow) bool { return job.ID == id })
}

func maintenanceWindowsContains(windows []SystemMaintenanceWindow, id pgtype.UUID) bool {
	return slices.ContainsFunc(windows, func(window SystemMaintenanceWindow) bool { return window.ID == id })
}

func maintenanceAuditLogsContains(logs []SystemMaintenanceAuditLog, id pgtype.UUID) bool {
	return slices.ContainsFunc(logs, func(log SystemMaintenanceAuditLog) bool { return log.ID == id })
}
