package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"mtsn2kolut-super-app/backend/internal/domain"
)

func TestSystemBackupListSortsAndMarksLatest(t *testing.T) {
	dir := t.TempDir()
	older := filepath.Join(dir, "pusaka_20260512_000001.dump")
	newer := filepath.Join(dir, "manual_20260513_090000.dump")
	preChange := filepath.Join(dir, "pre-migration_20260513_100000.dump")
	writeBackupFile(t, older, "older")
	writeBackupFile(t, newer, "newer")
	writeBackupFile(t, preChange, "pre")
	writeBackupFile(t, filepath.Join(dir, "notes.txt"), "ignored")
	writeBackupFile(t, newer+".sha256", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef  manual_20260513_090000.dump\n")

	base := time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC)
	mustChtimes(t, older, base.Add(-2*time.Hour))
	mustChtimes(t, newer, base.Add(-1*time.Hour))
	mustChtimes(t, preChange, base)
	if err := os.Symlink(newer, filepath.Join(dir, systemBackupLatestSymlinkName)); err != nil {
		t.Fatalf("symlink latest: %v", err)
	}

	svc := NewSystemBackup(SystemBackupConfig{BackupDir: dir})
	list, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	names := make([]string, 0, len(list.Items))
	kinds := map[string]string{}
	latest := ""
	for _, item := range list.Items {
		names = append(names, item.Name)
		kinds[item.Name] = item.Kind
		if item.IsLatest {
			latest = item.Name
		}
	}
	wantNames := []string{"pre-migration_20260513_100000.dump", "manual_20260513_090000.dump", "pusaka_20260512_000001.dump"}
	if !reflect.DeepEqual(names, wantNames) {
		t.Fatalf("List() names = %v, want %v", names, wantNames)
	}
	if latest != "manual_20260513_090000.dump" {
		t.Fatalf("latest = %q, want manual backup target", latest)
	}
	if kinds["pusaka_20260512_000001.dump"] != "scheduled" || kinds["pre-migration_20260513_100000.dump"] != "manual-pre-change" {
		t.Fatalf("backup kinds = %+v", kinds)
	}
	if got := list.Items[1].SHA256; got != "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" {
		t.Fatalf("SHA256 = %q", got)
	}
}

func TestSystemBackupListMissingDirectoryIsEmpty(t *testing.T) {
	svc := NewSystemBackup(SystemBackupConfig{BackupDir: filepath.Join(t.TempDir(), "missing")})
	list, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List(missing) error = %v", err)
	}
	if list.Meta.Total != 0 || len(list.Items) != 0 {
		t.Fatalf("List(missing) = %+v, want empty", list)
	}
}

func TestSystemBackupDownloadBlocksTraversalAndUnsafeSymlinks(t *testing.T) {
	dir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.dump")
	writeBackupFile(t, outside, "outside")
	if err := os.Symlink(outside, filepath.Join(dir, systemBackupLatestSymlinkName)); err != nil {
		t.Fatalf("symlink outside latest: %v", err)
	}
	svc := NewSystemBackup(SystemBackupConfig{BackupDir: dir})

	for _, id := range []string{"../outside.dump", "nested/file.dump", "file.sql"} {
		if _, err := svc.Download(context.Background(), id); !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("Download(%q) error = %v, want forbidden", id, err)
		}
	}
	if _, err := svc.Download(context.Background(), systemBackupLatestSymlinkName); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("Download(latest outside) error = %v, want forbidden", err)
	}
}

func TestSystemBackupDownloadAllowsSafeLatestSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "pusaka_20260513_000001.dump")
	writeBackupFile(t, target, "backup")
	if err := os.Symlink(target, filepath.Join(dir, systemBackupLatestSymlinkName)); err != nil {
		t.Fatalf("symlink latest: %v", err)
	}
	svc := NewSystemBackup(SystemBackupConfig{BackupDir: dir})
	download, err := svc.Download(context.Background(), systemBackupLatestSymlinkName)
	if err != nil {
		t.Fatalf("Download(latest) error = %v", err)
	}
	if download.Filename != filepath.Base(target) || download.Path != target {
		t.Fatalf("Download(latest) = %+v, want resolved target", download)
	}
}

func TestSystemBackupStatusUsesSystemdFieldsAndWarnings(t *testing.T) {
	dir := t.TempDir()
	backup := filepath.Join(dir, "pusaka_20260513_000001.dump")
	writeBackupFile(t, backup, "backup")
	now := time.Date(2026, 5, 13, 12, 0, 0, 0, time.FixedZone("WITA", 8*60*60))
	mustChtimes(t, backup, now.Add(-2*time.Hour))
	runner := func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if len(args) >= 2 && args[1] == DefaultSystemBackupTimerName {
			return []byte("UnitFileState=enabled\nActiveState=active\nNextElapseUSecRealtime=Thu 2026-05-14 00:00:00 WITA\nLastTriggerUSec=Wed 2026-05-13 00:00:01 WITA\n"), nil
		}
		return []byte("Result=success\nExecMainStatus=0\n"), nil
	}
	svc := NewSystemBackup(SystemBackupConfig{BackupDir: dir, Now: func() time.Time { return now }, CommandRunner: runner})

	status, err := svc.Status(context.Background())
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if status.Health != "ok" || !status.TimerEnabled || !status.TimerActive || status.BackupCount != 1 {
		t.Fatalf("Status() = %+v, want healthy active timer with one backup", status)
	}
	if status.NextRunAt == nil || status.NextRunAt.Location().String() != "WITA" {
		t.Fatalf("NextRunAt = %v, want parsed WITA timestamp", status.NextRunAt)
	}
	if status.LastRunSuccess == nil || !*status.LastRunSuccess {
		t.Fatalf("LastRunSuccess = %v, want true", status.LastRunSuccess)
	}
}

func TestSystemBackupStatusWarnsWhenTimerUnavailable(t *testing.T) {
	svc := NewSystemBackup(SystemBackupConfig{
		BackupDir: t.TempDir(),
		CommandRunner: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return nil, errors.New("systemctl unavailable")
		},
	})
	status, err := svc.Status(context.Background())
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if status.Health != "error" || len(status.Warnings) == 0 {
		t.Fatalf("Status() = %+v, want warning/error when timer cannot be verified", status)
	}
}

func TestSystemBackupRunManualExecutesFixedScriptAndSanitizesOutput(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "backup.sh")
	writeBackupFile(t, script, "#!/usr/bin/env bash\n")
	if err := os.Chmod(script, 0o700); err != nil {
		t.Fatalf("chmod script: %v", err)
	}
	now := time.Date(2026, 5, 13, 11, 0, 0, 0, time.UTC)
	var gotName string
	var gotArgs []string
	svc := NewSystemBackup(SystemBackupConfig{
		BackupDir:  dir,
		ScriptPath: script,
		Now:        func() time.Time { return now },
		CommandRunner: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			gotName = name
			gotArgs = append([]string{}, args...)
			return []byte("Starting backup\npassword=secret\nBackup completed\n"), nil
		},
	})

	job, err := svc.RunManual(context.Background(), SystemBackupRunRequest{Reason: " before deploy  "})
	if err != nil {
		t.Fatalf("RunManual() error = %v", err)
	}
	if job.Status != "success" || job.ID != "backup-20260513-110000" || job.Reason != "before deploy" {
		t.Fatalf("job = %+v", job)
	}
	if gotName != "bash" || len(gotArgs) != 1 || gotArgs[0] != script {
		t.Fatalf("command = %s %v, want bash fixed script", gotName, gotArgs)
	}
	if strings.Contains(strings.ToLower(job.Output), "secret") || !strings.Contains(job.Output, "[REDACTED]") {
		t.Fatalf("output not sanitized: %q", job.Output)
	}
}

func TestSystemBackupRunManualRejectsConcurrentRun(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "backup.sh")
	writeBackupFile(t, script, "#!/usr/bin/env bash\n")
	started := make(chan struct{})
	release := make(chan struct{})
	svc := NewSystemBackup(SystemBackupConfig{
		BackupDir:  dir,
		ScriptPath: script,
		CommandRunner: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			close(started)
			<-release
			return []byte("done"), nil
		},
	})
	done := make(chan error, 1)
	go func() {
		_, err := svc.RunManual(context.Background(), SystemBackupRunRequest{})
		done <- err
	}()
	<-started
	if _, err := svc.RunManual(context.Background(), SystemBackupRunRequest{}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("second RunManual error = %v, want conflict", err)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatalf("first RunManual error = %v", err)
	}
}

func writeBackupFile(t *testing.T, path string, value string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mustChtimes(t *testing.T, path string, ts time.Time) {
	t.Helper()
	if err := os.Chtimes(path, ts, ts); err != nil {
		t.Fatalf("chtimes %s: %v", path, err)
	}
}
