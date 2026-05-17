package service

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeMaintenanceStore struct {
	activeWindow db.SystemMaintenanceWindow
	window       db.SystemMaintenanceWindow
	windows      []db.SystemMaintenanceWindow
	auditLogs    []db.SystemMaintenanceAuditLog

	activeErr    error
	windowErr    error
	listErr      error
	createErr    error
	updateErr    error
	setErr       error
	auditErr     error
	auditListErr error
	dbTimeErr    error
	cbtErr       error

	dbTime         pgtype.Timestamptz
	activeSessions int64
	created        []db.CreateMaintenanceWindowParams
	updated        []db.UpdateMaintenanceWindowParams
	setActives     []db.SetMaintenanceWindowActiveParams
	audits         []db.CreateMaintenanceAuditLogParams
	listParams     []db.ListMaintenanceWindowsParams
	auditParams    []db.ListMaintenanceAuditLogsParams
}

func maintenanceTestUUID(seed byte) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte{seed, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, seed}, Valid: true}
}

func (f *fakeMaintenanceStore) GetActiveMaintenanceWindow(context.Context) (db.SystemMaintenanceWindow, error) {
	return f.activeWindow, f.activeErr
}
func (f *fakeMaintenanceStore) GetMaintenanceWindow(context.Context, pgtype.UUID) (db.SystemMaintenanceWindow, error) {
	return f.window, f.windowErr
}
func (f *fakeMaintenanceStore) ListMaintenanceWindows(ctx context.Context, arg db.ListMaintenanceWindowsParams) ([]db.SystemMaintenanceWindow, error) {
	f.listParams = append(f.listParams, arg)
	return f.windows, f.listErr
}
func (f *fakeMaintenanceStore) CreateMaintenanceWindow(ctx context.Context, arg db.CreateMaintenanceWindowParams) (db.SystemMaintenanceWindow, error) {
	f.created = append(f.created, arg)
	return maintenanceRowFromCreate(arg), f.createErr
}
func (f *fakeMaintenanceStore) UpdateMaintenanceWindow(ctx context.Context, arg db.UpdateMaintenanceWindowParams) (db.SystemMaintenanceWindow, error) {
	f.updated = append(f.updated, arg)
	row := maintenanceRowFromUpdate(arg)
	row.ID = arg.ID
	return row, f.updateErr
}
func (f *fakeMaintenanceStore) SetMaintenanceWindowActive(ctx context.Context, arg db.SetMaintenanceWindowActiveParams) (db.SystemMaintenanceWindow, error) {
	f.setActives = append(f.setActives, arg)
	row := f.window
	if !row.ID.Valid {
		row.ID = arg.ID
	}
	row.IsActive = arg.IsActive
	row.UpdatedBy = arg.ActorUserID
	return row, f.setErr
}
func (f *fakeMaintenanceStore) CreateMaintenanceAuditLog(ctx context.Context, arg db.CreateMaintenanceAuditLogParams) (db.SystemMaintenanceAuditLog, error) {
	f.audits = append(f.audits, arg)
	return db.SystemMaintenanceAuditLog{}, f.auditErr
}
func (f *fakeMaintenanceStore) ListMaintenanceAuditLogs(ctx context.Context, arg db.ListMaintenanceAuditLogsParams) ([]db.SystemMaintenanceAuditLog, error) {
	f.auditParams = append(f.auditParams, arg)
	return f.auditLogs, f.auditListErr
}
func (f *fakeMaintenanceStore) GetMaintenanceDatabaseTime(context.Context) (pgtype.Timestamptz, error) {
	return f.dbTime, f.dbTimeErr
}
func (f *fakeMaintenanceStore) CountActiveCbtSessionsForMaintenance(context.Context) (int64, error) {
	return f.activeSessions, f.cbtErr
}

func maintenanceRowFromCreate(arg db.CreateMaintenanceWindowParams) db.SystemMaintenanceWindow {
	now := pgtype.Timestamptz{Time: time.Date(2026, 5, 17, 9, 0, 0, 0, time.UTC), Valid: true}
	return db.SystemMaintenanceWindow{ID: maintenanceTestUUID(1), Title: arg.Title, Message: arg.Message, Mode: arg.Mode, AffectedModules: arg.AffectedModules, StartsAt: arg.StartsAt, EndsAt: arg.EndsAt, IsActive: arg.IsActive, AllowAdminBypass: arg.AllowAdminBypass, BypassRoles: arg.BypassRoles, Severity: arg.Severity, CreatedBy: arg.ActorUserID, UpdatedBy: arg.ActorUserID, CreatedAt: now, UpdatedAt: now}
}
func maintenanceRowFromUpdate(arg db.UpdateMaintenanceWindowParams) db.SystemMaintenanceWindow {
	now := pgtype.Timestamptz{Time: time.Date(2026, 5, 17, 10, 0, 0, 0, time.UTC), Valid: true}
	return db.SystemMaintenanceWindow{Title: arg.Title, Message: arg.Message, Mode: arg.Mode, AffectedModules: arg.AffectedModules, StartsAt: arg.StartsAt, EndsAt: arg.EndsAt, IsActive: false, AllowAdminBypass: arg.AllowAdminBypass, BypassRoles: arg.BypassRoles, Severity: arg.Severity, UpdatedBy: arg.ActorUserID, CreatedAt: now, UpdatedAt: now}
}

func TestSystemMaintenanceConstructorsInitializeService(t *testing.T) {
	store := &fakeMaintenanceStore{}
	if svc := NewSystemMaintenanceWithStore(store, nil); svc == nil || svc.q != store || svc.now == nil {
		t.Fatalf("NewSystemMaintenanceWithStore() = %+v", svc)
	}
	if svc := NewSystemMaintenance(nil, nil); svc == nil || svc.now == nil {
		t.Fatalf("NewSystemMaintenance() = %+v", svc)
	}
}

func TestMaintenanceModuleHelpersMatchExactSegments(t *testing.T) {
	modules := MaintenanceModules()
	if len(modules) != 10 || modules[0] != MaintenanceModuleGlobal || modules[len(modules)-1] != MaintenanceModuleSettings {
		t.Fatalf("MaintenanceModules() = %v", modules)
	}
	if !IsValidMaintenanceModule(" PUSAKA ") || !IsValidMaintenanceModule("backup_restore") || IsValidMaintenanceModule("pusaka-worker") {
		t.Fatalf("IsValidMaintenanceModule did not normalize/validate as expected")
	}
	cases := []struct {
		module string
		path   string
		want   bool
	}{
		{MaintenanceModuleGlobal, "/api/ping", true},
		{MaintenanceModuleGlobal, "/public/api/ping", false},
		{MaintenanceModulePusaka, "/api/pusaka/employees", true},
		{MaintenanceModulePusaka, "/api/pusaka-extra", false},
		{MaintenanceModuleBackupRestore, "/api/system/backups/latest", true},
		{MaintenanceModuleSettings, "/api/rbac/roles", true},
		{MaintenanceModuleCBT, "/api/cbt/questions", false},
	}
	for _, tc := range cases {
		if got := MaintenanceModuleMatchesAPIPath(tc.module, tc.path); got != tc.want {
			t.Fatalf("MaintenanceModuleMatchesAPIPath(%q, %q) = %v, want %v", tc.module, tc.path, got, tc.want)
		}
	}
	if !MaintenanceModulesMatchAPIPath([]string{MaintenanceModuleStudents, MaintenanceModulePusaka}, "/api/kesiswaan/photos/1") {
		t.Fatalf("MaintenanceModulesMatchAPIPath did not match student/kesiswaan prefix")
	}
}

func TestSystemMaintenanceCreateWindowNormalizesAndAudits(t *testing.T) {
	store := &fakeMaintenanceStore{}
	svc := NewSystemMaintenanceWithStore(store, nil)
	svc.now = func() time.Time { return time.Date(2026, 5, 17, 8, 0, 0, 0, time.UTC) }
	start := svc.now().Add(time.Hour)
	end := start.Add(2 * time.Hour)
	allow := false
	actor := maintenanceTestUUID(9)

	view, err := svc.CreateWindow(context.Background(), MaintenanceWindowInput{
		Title: "  Upgrade inti  ", Message: "  Maintenance PUSAKA  ", Mode: " MODULE ",
		AffectedModules: []string{" pusaka ", "backup-restore", "pusaka", ""},
		StartsAt:        &start, EndsAt: &end, IsActive: true, AllowAdminBypass: &allow,
		BypassRoles: []string{" Admin ", "admin", "superadmin"}, Severity: " WARNING ", ActorUserID: actor, Reason: " deploy ",
	})
	if err != nil {
		t.Fatalf("CreateWindow() error = %v", err)
	}
	if len(store.created) != 1 || len(store.audits) != 1 {
		t.Fatalf("created/audits = %d/%d, want 1/1", len(store.created), len(store.audits))
	}
	created := store.created[0]
	if created.Title != "Upgrade inti" || created.Message != "Maintenance PUSAKA" || created.Mode != MaintenanceModeModule || created.Severity != MaintenanceSeverityWarning {
		t.Fatalf("created params not normalized: %+v", created)
	}
	if !reflect.DeepEqual(created.AffectedModules, []string{MaintenanceModulePusaka, MaintenanceModuleBackupRestore}) {
		t.Fatalf("AffectedModules = %v", created.AffectedModules)
	}
	if created.AllowAdminBypass || !reflect.DeepEqual(created.BypassRoles, []string{"admin", "superadmin"}) {
		t.Fatalf("bypass params = allow %v roles %v", created.AllowAdminBypass, created.BypassRoles)
	}
	if view.Status != "scheduled" || view.ID == "" {
		t.Fatalf("view = %+v, want scheduled with id", view)
	}
	if !store.audits[0].Reason.Valid || store.audits[0].Action != maintenanceActionCreate {
		t.Fatalf("audit = %+v", store.audits[0])
	}
}

func TestSystemMaintenanceListWindowsDefaultsLimitAndMapsViews(t *testing.T) {
	now := time.Date(2026, 5, 17, 8, 0, 0, 0, time.UTC)
	store := &fakeMaintenanceStore{windows: []db.SystemMaintenanceWindow{
		{ID: maintenanceTestUUID(2), Title: "Global", Message: "patch", Mode: MaintenanceModeGlobal, AffectedModules: []string{MaintenanceModuleGlobal}, IsActive: true, AllowAdminBypass: true, BypassRoles: []string{"admin"}, Severity: MaintenanceSeverityInfo, CreatedAt: pgtype.Timestamptz{Time: now.Add(-time.Hour), Valid: true}, UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true}},
	}}
	svc := NewSystemMaintenanceWithStore(store, nil)
	svc.now = func() time.Time { return now }

	views, err := svc.ListWindows(context.Background(), 0, 7)
	if err != nil {
		t.Fatalf("ListWindows() error = %v", err)
	}
	if len(store.listParams) != 1 || store.listParams[0].LimitCount != 50 || store.listParams[0].OffsetCount != 7 {
		t.Fatalf("list params = %+v", store.listParams)
	}
	if len(views) != 1 || views[0].Title != "Global" || views[0].Status != "active_now" || views[0].CreatedAt.IsZero() {
		t.Fatalf("views = %+v", views)
	}

	store.listErr = errors.New("boom")
	if _, err := svc.ListWindows(context.Background(), 10, 0); err == nil {
		t.Fatalf("ListWindows() error = nil, want store error")
	}
}

func TestSystemMaintenanceUpdateWindowNormalizesAuditsAndMapsNotFound(t *testing.T) {
	now := time.Date(2026, 5, 17, 8, 0, 0, 0, time.UTC)
	store := &fakeMaintenanceStore{}
	svc := NewSystemMaintenanceWithStore(store, nil)
	svc.now = func() time.Time { return now }
	start := now.Add(time.Hour)
	end := start.Add(time.Hour)
	actor := maintenanceTestUUID(9)

	view, err := svc.UpdateWindow(context.Background(), maintenanceTestUUID(3), MaintenanceWindowInput{
		Title: "  Update  ", Message: " msg ", Mode: "read_only", AffectedModules: []string{},
		StartsAt: &start, EndsAt: &end, IsActive: true, Severity: "", ActorUserID: actor, Reason: " change ",
	})
	if err != nil {
		t.Fatalf("UpdateWindow() error = %v", err)
	}
	if len(store.updated) != 1 || len(store.audits) != 1 {
		t.Fatalf("updated/audits = %d/%d, want 1/1", len(store.updated), len(store.audits))
	}
	updated := store.updated[0]
	if updated.Title != "Update" || updated.Message != "msg" || updated.Mode != MaintenanceModeReadOnly || updated.Severity != MaintenanceSeverityInfo {
		t.Fatalf("updated params not normalized: %+v", updated)
	}
	if !reflect.DeepEqual(updated.AffectedModules, []string{MaintenanceModuleGlobal}) || !reflect.DeepEqual(updated.BypassRoles, []string{"superadmin", "admin"}) || !updated.AllowAdminBypass {
		t.Fatalf("updated defaults = modules %v roles %v allow %v", updated.AffectedModules, updated.BypassRoles, updated.AllowAdminBypass)
	}
	if view.IsActive || view.Status != "inactive" || store.audits[0].Action != maintenanceActionUpdate || store.audits[0].Reason.String != "change" {
		t.Fatalf("view/audit = %+v / %+v", view, store.audits[0])
	}

	if _, err := svc.UpdateWindow(context.Background(), pgtype.UUID{}, MaintenanceWindowInput{}); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("UpdateWindow(invalid id) error = %v, want bad request", err)
	}
	store.updateErr = pgx.ErrNoRows
	if _, err := svc.UpdateWindow(context.Background(), maintenanceTestUUID(3), MaintenanceWindowInput{Title: "x", Message: "x", Mode: MaintenanceModeGlobal}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("UpdateWindow(no rows) error = %v, want not found", err)
	}
}

func TestSystemMaintenanceValidationRejectsInvalidWindows(t *testing.T) {
	svc := NewSystemMaintenanceWithStore(&fakeMaintenanceStore{}, nil)
	now := time.Date(2026, 5, 17, 8, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return now }
	past := now.Add(-time.Hour)
	start := now.Add(time.Hour)
	endBeforeStart := start.Add(-time.Minute)

	cases := []MaintenanceWindowInput{
		{Title: "", Message: "x", Mode: MaintenanceModeGlobal},
		{Title: "x", Message: "", Mode: MaintenanceModeGlobal},
		{Title: "x", Message: "x", Mode: "offline"},
		{Title: "x", Message: "x", Mode: MaintenanceModeGlobal, Severity: "fatal"},
		{Title: "x", Message: "x", Mode: MaintenanceModeGlobal, StartsAt: &start, EndsAt: &endBeforeStart},
		{Title: "x", Message: "x", Mode: MaintenanceModeModule},
		{Title: "x", Message: "x", Mode: MaintenanceModeModule, AffectedModules: []string{"unknown"}},
		{Title: "x", Message: "x", Mode: MaintenanceModeGlobal, IsActive: true, EndsAt: &past},
	}
	for i, in := range cases {
		if _, err := svc.CreateWindow(context.Background(), in); !errors.Is(err, domain.ErrBadRequest) && !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("case %d error = %v, want bad request/conflict", i, err)
		}
	}
}

func TestSystemMaintenanceStatusAndLifecycleViews(t *testing.T) {
	now := time.Date(2026, 5, 17, 8, 0, 0, 0, time.UTC)
	row := db.SystemMaintenanceWindow{ID: maintenanceTestUUID(2), Title: "Patch", Message: "Read only", Mode: MaintenanceModeReadOnly, AffectedModules: []string{MaintenanceModuleGlobal}, IsActive: true, StartsAt: pgtype.Timestamptz{Time: now.Add(-time.Hour), Valid: true}, EndsAt: pgtype.Timestamptz{Time: now.Add(time.Hour), Valid: true}, AllowAdminBypass: true, BypassRoles: []string{"admin"}, Severity: MaintenanceSeverityCritical, CreatedAt: pgtype.Timestamptz{Time: now.Add(-2 * time.Hour), Valid: true}, UpdatedAt: pgtype.Timestamptz{Time: now.Add(-time.Hour), Valid: true}}
	store := &fakeMaintenanceStore{activeWindow: row}
	svc := NewSystemMaintenanceWithStore(store, nil)
	svc.now = func() time.Time { return now }
	status, err := svc.Status(context.Background())
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if !status.Active || status.Mode != MaintenanceModeReadOnly || status.Window == nil || status.Window.Status != "active_now" {
		t.Fatalf("Status() = %+v", status)
	}

	store.activeErr = pgx.ErrNoRows
	status, err = svc.Status(context.Background())
	if err != nil || status.Active || status.Mode != MaintenanceModeOff {
		t.Fatalf("Status(no rows) = %+v, %v", status, err)
	}

	row.StartsAt.Time = now.Add(time.Hour)
	if got := maintenanceWindowLifecycleStatus(row, now); got != "scheduled" {
		t.Fatalf("scheduled lifecycle = %q", got)
	}
	row.StartsAt.Time = now.Add(-2 * time.Hour)
	row.EndsAt.Time = now.Add(-time.Minute)
	if got := maintenanceWindowLifecycleStatus(row, now); got != "ended" {
		t.Fatalf("ended lifecycle = %q", got)
	}
	row.IsActive = false
	if got := maintenanceWindowLifecycleStatus(row, now); got != "inactive" {
		t.Fatalf("inactive lifecycle = %q", got)
	}
}

func TestSystemMaintenanceActivateWindowSetsActiveAndAudits(t *testing.T) {
	now := time.Date(2026, 5, 17, 8, 0, 0, 0, time.UTC)
	id := maintenanceTestUUID(4)
	actor := maintenanceTestUUID(9)
	store := &fakeMaintenanceStore{window: db.SystemMaintenanceWindow{ID: id, Title: "Patch", Message: "open", Mode: MaintenanceModeModule, AffectedModules: []string{MaintenanceModulePusaka}, EndsAt: pgtype.Timestamptz{Time: now.Add(time.Hour), Valid: true}, AllowAdminBypass: true, BypassRoles: []string{"admin"}, Severity: MaintenanceSeverityWarning, CreatedAt: pgtype.Timestamptz{Time: now.Add(-time.Hour), Valid: true}, UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true}}}
	svc := NewSystemMaintenanceWithStore(store, nil)
	svc.now = func() time.Time { return now }

	view, err := svc.ActivateWindow(context.Background(), id, actor, "start")
	if err != nil {
		t.Fatalf("ActivateWindow() error = %v", err)
	}
	if len(store.setActives) != 1 || !store.setActives[0].IsActive || store.setActives[0].ActorUserID != actor {
		t.Fatalf("set active params = %+v", store.setActives)
	}
	if !view.IsActive || view.Status != "active_now" || len(store.audits) != 1 || store.audits[0].Action != maintenanceActionActivate {
		t.Fatalf("view/audits = %+v / %+v", view, store.audits)
	}

	if _, err := svc.ActivateWindow(context.Background(), pgtype.UUID{}, actor, ""); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("ActivateWindow(invalid id) error = %v, want bad request", err)
	}
	store.windowErr = pgx.ErrNoRows
	if _, err := svc.ActivateWindow(context.Background(), id, actor, ""); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("ActivateWindow(no rows) error = %v, want not found", err)
	}
}

func TestSystemMaintenanceDeactivateWindowSetsInactiveAndAudits(t *testing.T) {
	now := time.Date(2026, 5, 17, 8, 0, 0, 0, time.UTC)
	id := maintenanceTestUUID(5)
	actor := maintenanceTestUUID(9)
	store := &fakeMaintenanceStore{window: db.SystemMaintenanceWindow{ID: id, Title: "Patch", Message: "close", Mode: MaintenanceModeGlobal, AffectedModules: []string{MaintenanceModuleGlobal}, IsActive: true, AllowAdminBypass: true, BypassRoles: []string{"admin"}, Severity: MaintenanceSeverityCritical, CreatedAt: pgtype.Timestamptz{Time: now.Add(-time.Hour), Valid: true}, UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true}}}
	svc := NewSystemMaintenanceWithStore(store, nil)
	svc.now = func() time.Time { return now }

	view, err := svc.DeactivateWindow(context.Background(), id, actor, "done")
	if err != nil {
		t.Fatalf("DeactivateWindow() error = %v", err)
	}
	if len(store.setActives) != 1 || store.setActives[0].IsActive || store.setActives[0].ActorUserID != actor {
		t.Fatalf("set inactive params = %+v", store.setActives)
	}
	if view.IsActive || view.Status != "inactive" || len(store.audits) != 1 || store.audits[0].Action != maintenanceActionDeactivate {
		t.Fatalf("view/audits = %+v / %+v", view, store.audits)
	}

	if _, err := svc.DeactivateWindow(context.Background(), pgtype.UUID{}, actor, ""); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("DeactivateWindow(invalid id) error = %v, want bad request", err)
	}
	store.setErr = pgx.ErrNoRows
	if _, err := svc.DeactivateWindow(context.Background(), id, actor, ""); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("DeactivateWindow(no rows) error = %v, want not found", err)
	}
}

func TestSystemMaintenanceActivateRejectsEndedWindow(t *testing.T) {
	now := time.Date(2026, 5, 17, 8, 0, 0, 0, time.UTC)
	store := &fakeMaintenanceStore{window: db.SystemMaintenanceWindow{ID: maintenanceTestUUID(3), Title: "Old", Message: "done", Mode: MaintenanceModeGlobal, EndsAt: pgtype.Timestamptz{Time: now.Add(-time.Second), Valid: true}}}
	svc := NewSystemMaintenanceWithStore(store, nil)
	svc.now = func() time.Time { return now }
	_, err := svc.ActivateWindow(context.Background(), maintenanceTestUUID(3), maintenanceTestUUID(9), "late")
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("ActivateWindow(ended) error = %v, want conflict", err)
	}
	if len(store.setActives) != 0 {
		t.Fatalf("SetMaintenanceWindowActive called for ended window")
	}
}

func TestSystemMaintenanceHealthSummaryCombinesDBBackupAndCBT(t *testing.T) {
	now := time.Date(2026, 5, 17, 8, 0, 0, 0, time.UTC)
	latest := now.Add(-2 * time.Hour)
	ok := true
	store := &fakeMaintenanceStore{dbTime: pgtype.Timestamptz{Time: now, Valid: true}, activeSessions: 2}
	backup := fakeMaintenanceBackupProvider{status: SystemBackupStatus{Health: "ok", BackupCount: 3, LastRunAt: &latest, LastRunSuccess: &ok, LatestBackup: &SystemBackupFile{CreatedAt: latest}, BackupDirSizeBytes: 2048}}
	svc := NewSystemMaintenanceWithStore(store, backup)
	svc.now = func() time.Time { return now }

	summary, err := svc.HealthSummary(context.Background())
	if err != nil {
		t.Fatalf("HealthSummary() error = %v", err)
	}
	if !summary.Database.Connected || summary.Database.DBTime == nil {
		t.Fatalf("Database health = %+v", summary.Database)
	}
	if !summary.Backup.Available || summary.Backup.LatestBackupAt == nil || summary.Disk.BackupDirSizeBytes != 2048 {
		t.Fatalf("Backup/disk health = %+v / %+v", summary.Backup, summary.Disk)
	}
	if summary.CBT.Health != "warning" || !strings.Contains(summary.CBT.Message, "2 sesi") {
		t.Fatalf("CBT health = %+v", summary.CBT)
	}
	if len(summary.Checklist) != 6 || summary.Checklist[4].OK {
		t.Fatalf("checklist = %+v", summary.Checklist)
	}
}

type fakeMaintenanceBackupProvider struct {
	status SystemBackupStatus
	err    error
}

func (f fakeMaintenanceBackupProvider) Status(context.Context) (SystemBackupStatus, error) {
	return f.status, f.err
}

func TestSystemMaintenanceListAuditLogsNormalizesFilterAndMapsRows(t *testing.T) {
	now := time.Date(2026, 5, 17, 8, 0, 0, 0, time.UTC)
	from := now.Add(-24 * time.Hour)
	to := now
	store := &fakeMaintenanceStore{auditLogs: []db.SystemMaintenanceAuditLog{
		{ID: maintenanceTestUUID(7), MaintenanceID: maintenanceTestUUID(8), ActorUserID: maintenanceTestUUID(9), Action: maintenanceActionUpdate, Reason: pgtype.Text{String: "checked", Valid: true}, Metadata: []byte(`{"safe":"ok","secret_token":"abc"}`), CreatedAt: pgtype.Timestamptz{Time: now, Valid: true}},
	}}
	svc := NewSystemMaintenanceWithStore(store, nil)

	logs, err := svc.ListAuditLogs(context.Background(), MaintenanceAuditFilter{Action: " UPDATE ", FromAt: &from, ToAt: &to, Offset: 3})
	if err != nil {
		t.Fatalf("ListAuditLogs() error = %v", err)
	}
	if len(store.auditParams) != 1 || store.auditParams[0].ActionFilter != maintenanceActionUpdate || store.auditParams[0].LimitCount != 50 || store.auditParams[0].OffsetCount != 3 || !store.auditParams[0].FromAt.Valid || !store.auditParams[0].ToAt.Valid {
		t.Fatalf("audit params = %+v", store.auditParams)
	}
	if len(logs) != 1 || logs[0].Action != maintenanceActionUpdate || logs[0].Reason != "checked" || logs[0].Metadata["secret_token"] != "[REDACTED]" {
		t.Fatalf("logs = %+v", logs)
	}

	store.auditListErr = errors.New("audit list failed")
	if _, err := svc.ListAuditLogs(context.Background(), MaintenanceAuditFilter{Limit: 10}); err == nil {
		t.Fatalf("ListAuditLogs() error = nil, want store error")
	}
}

func TestMaintenanceAuditLogRedactsNestedSecrets(t *testing.T) {
	row := db.SystemMaintenanceAuditLog{ID: maintenanceTestUUID(4), MaintenanceID: maintenanceTestUUID(5), ActorUserID: maintenanceTestUUID(6), Action: "update", Reason: pgtype.Text{String: "credential rotation", Valid: true}, Metadata: []byte(`{"token":"abc","safe":"ok","nested":{"password":"secret","count":2}}`), CreatedAt: pgtype.Timestamptz{Time: time.Date(2026, 5, 17, 8, 0, 0, 0, time.UTC), Valid: true}}
	view := auditLogView(row)
	if view.Metadata["token"] != "[REDACTED]" || view.Metadata["safe"] != "ok" {
		t.Fatalf("metadata redaction = %+v", view.Metadata)
	}
	nested, ok := view.Metadata["nested"].(map[string]any)
	if !ok || nested["password"] != "[REDACTED]" || nested["count"].(float64) != 2 {
		t.Fatalf("nested metadata redaction = %+v", view.Metadata["nested"])
	}
}
