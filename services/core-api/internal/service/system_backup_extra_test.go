package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"mtsn2kolut-super-app/backend/internal/domain"
)

func TestSystemBackupExtraParseSystemdHelpers(t *testing.T) {
	fields := parseSystemctlShow(" UnitFileState = enabled \nNoEqualsLine\nActiveState= active\nActiveState=inactive\n")
	if fields["UnitFileState"] != "enabled" || fields["ActiveState"] != "inactive" {
		t.Fatalf("parseSystemctlShow() = %#v, want trimmed fields with later value", fields)
	}

	if got := parseSystemdTimestamp("n/a"); got != nil {
		t.Fatalf("parseSystemdTimestamp(n/a) = %v, want nil", got)
	}
	if got := parseSystemdTimestamp("Thu 2026-05-14 00:00:00 WITA"); got == nil || got.Location().String() != "WITA" || got.Hour() != 0 {
		t.Fatalf("parseSystemdTimestamp(WITA) = %v, want WITA midnight", got)
	}
	if got := parseSystemdTimestamp("2026-05-14T01:02:03Z"); got == nil || got.UTC().Hour() != 1 {
		t.Fatalf("parseSystemdTimestamp(RFC3339) = %v, want parsed UTC", got)
	}
	if got := parseSystemdTimestamp("definitely not a timestamp"); got != nil {
		t.Fatalf("parseSystemdTimestamp(invalid) = %v, want nil", got)
	}
	if !isWeekday("Mon") || isWeekday("Monday") {
		t.Fatalf("isWeekday returned unexpected values")
	}
}

func TestSystemBackupExtraOffsiteStatusFileBranches(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC)

	t.Run("missing status file is configured error", func(t *testing.T) {
		svc := NewSystemBackup(SystemBackupConfig{BackupDir: dir, OffsiteStatusFile: filepath.Join(dir, "missing.json")})
		status, err := svc.OffsiteStatus(context.Background())
		if err != nil {
			t.Fatalf("OffsiteStatus() error = %v", err)
		}
		if !status.Configured || status.Source != "status_file" || status.Health != "error" || len(status.Warnings) == 0 {
			t.Fatalf("OffsiteStatus(missing) = %+v, want configured status_file error", status)
		}
	})

	t.Run("invalid json status file is non-fatal error health", func(t *testing.T) {
		statusFile := filepath.Join(dir, "invalid.json")
		writeBackupFile(t, statusFile, `{not-json`)
		svc := NewSystemBackup(SystemBackupConfig{BackupDir: dir, OffsiteStatusFile: statusFile})
		status, err := svc.OffsiteStatus(context.Background())
		if err != nil {
			t.Fatalf("OffsiteStatus() error = %v", err)
		}
		if !status.Configured || status.Health != "error" || !strings.Contains(strings.Join(status.Warnings, " "), "JSON valid") {
			t.Fatalf("OffsiteStatus(invalid json) = %+v, want JSON warning", status)
		}
	})

	t.Run("failed stale payload sanitizes and reports error", func(t *testing.T) {
		statusFile := filepath.Join(dir, "failed.json")
		writeBackupFile(t, statusFile, `{"provider":"rclone token secret","target_label":"label","last_sync_at":"not-a-time","last_sync_success":false,"remote_backup_count":0}`)
		svc := NewSystemBackup(SystemBackupConfig{BackupDir: dir, OffsiteStatusFile: statusFile, Now: func() time.Time { return now }})
		status, err := svc.OffsiteStatus(context.Background())
		if err != nil {
			t.Fatalf("OffsiteStatus() error = %v", err)
		}
		if status.Health != "error" || status.Provider != "[REDACTED]" || len(status.Warnings) < 3 {
			t.Fatalf("OffsiteStatus(failed) = %+v, want redacted error with warnings", status)
		}
	})
}

func TestSystemBackupExtraRemoteDirAndDownloadBranches(t *testing.T) {
	dir := t.TempDir()

	t.Run("missing remote dir reports mount warning", func(t *testing.T) {
		svc := NewSystemBackup(SystemBackupConfig{BackupDir: dir, OffsiteRemoteDir: filepath.Join(dir, "missing-remote")})
		status, err := svc.OffsiteStatus(context.Background())
		if err != nil {
			t.Fatalf("OffsiteStatus() error = %v", err)
		}
		if !status.Configured || status.Source != "remote_dir" || status.Health != "error" {
			t.Fatalf("OffsiteStatus(missing remote) = %+v, want remote_dir error", status)
		}
	})

	t.Run("empty remote dir warns about no dumps", func(t *testing.T) {
		remote := filepath.Join(dir, "empty-remote")
		if err := os.Mkdir(remote, 0o700); err != nil {
			t.Fatalf("mkdir remote: %v", err)
		}
		writeBackupFile(t, filepath.Join(remote, "notes.txt"), "ignored")
		svc := NewSystemBackup(SystemBackupConfig{BackupDir: dir, OffsiteRemoteDir: remote})
		status, err := svc.OffsiteStatus(context.Background())
		if err != nil {
			t.Fatalf("OffsiteStatus() error = %v", err)
		}
		if status.Health != "error" || status.RemoteBackupCount != 0 || status.LastSyncAt != nil || status.LastSyncSuccess == nil || *status.LastSyncSuccess {
			t.Fatalf("OffsiteStatus(empty remote) = %+v, want failed no-dump health", status)
		}
	})

	t.Run("latest regular file is listed as latest and downloadable", func(t *testing.T) {
		backupDir := filepath.Join(dir, "regular-latest")
		if err := os.Mkdir(backupDir, 0o700); err != nil {
			t.Fatalf("mkdir backupDir: %v", err)
		}
		latest := filepath.Join(backupDir, systemBackupLatestSymlinkName)
		writeBackupFile(t, latest, "dump")
		svc := NewSystemBackup(SystemBackupConfig{BackupDir: backupDir})
		list, err := svc.List(context.Background())
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
		if len(list.Items) != 1 || !list.Items[0].IsLatest || !list.Items[0].Downloadable {
			t.Fatalf("List(regular latest) = %+v, want latest downloadable item", list)
		}
		download, err := svc.Download(context.Background(), systemBackupLatestSymlinkName)
		if err != nil {
			t.Fatalf("Download(regular latest) error = %v", err)
		}
		if download.Filename != systemBackupLatestSymlinkName || download.SizeBytes != 4 {
			t.Fatalf("Download(regular latest) = %+v", download)
		}
	})

	t.Run("download rejects bad ids and non-regular targets", func(t *testing.T) {
		backupDir := filepath.Join(dir, "download-errors")
		if err := os.Mkdir(backupDir, 0o700); err != nil {
			t.Fatalf("mkdir backupDir: %v", err)
		}
		if err := os.Mkdir(filepath.Join(backupDir, "folder.dump"), 0o700); err != nil {
			t.Fatalf("mkdir folder.dump: %v", err)
		}
		writeBackupFile(t, filepath.Join(backupDir, "target.dump"), "dump")
		if err := os.Symlink(filepath.Join(backupDir, "target.dump"), filepath.Join(backupDir, "alias.dump")); err != nil {
			t.Fatalf("symlink alias: %v", err)
		}
		svc := NewSystemBackup(SystemBackupConfig{BackupDir: backupDir})
		if _, err := svc.Download(context.Background(), " "); !errors.Is(err, domain.ErrBadRequest) {
			t.Fatalf("Download(empty) error = %v, want bad request", err)
		}
		if _, err := svc.Download(context.Background(), "missing.dump"); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("Download(missing) error = %v, want not found", err)
		}
		if _, err := svc.Download(context.Background(), "folder.dump"); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("Download(directory) error = %v, want not found", err)
		}
		if _, err := svc.Download(context.Background(), "alias.dump"); !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("Download(non-latest symlink) error = %v, want forbidden", err)
		}
	})
}
