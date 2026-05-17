package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"mtsn2kolut-super-app/backend/internal/domain"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeSystemBackupService struct {
	status        service.SystemBackupStatus
	list          service.SystemBackupList
	download      service.SystemBackupDownload
	job           service.SystemBackupJob
	validation    service.SystemBackupRestoreValidation
	command       service.SystemBackupRestoreCommand
	offsite       service.SystemBackupOffsiteStatus
	err           error
	runReq        service.SystemBackupRunRequest
	lastJobID     string
	lastRestoreID string
}

func (f *fakeSystemBackupService) Status(ctx context.Context) (service.SystemBackupStatus, error) {
	return f.status, f.err
}

func (f *fakeSystemBackupService) List(ctx context.Context) (service.SystemBackupList, error) {
	return f.list, f.err
}

func (f *fakeSystemBackupService) Download(ctx context.Context, id string) (service.SystemBackupDownload, error) {
	if f.err != nil {
		return service.SystemBackupDownload{}, f.err
	}
	return f.download, nil
}

func (f *fakeSystemBackupService) RunManual(ctx context.Context, req service.SystemBackupRunRequest) (service.SystemBackupJob, error) {
	f.runReq = req
	if f.err != nil {
		return f.job, f.err
	}
	return f.job, nil
}

func (f *fakeSystemBackupService) Job(ctx context.Context, id string) (service.SystemBackupJob, error) {
	f.lastJobID = id
	if f.err != nil {
		return service.SystemBackupJob{}, f.err
	}
	return f.job, nil
}

func (f *fakeSystemBackupService) ValidateRestore(ctx context.Context, id string) (service.SystemBackupRestoreValidation, error) {
	f.lastRestoreID = id
	if f.err != nil {
		return service.SystemBackupRestoreValidation{}, f.err
	}
	return f.validation, nil
}

func (f *fakeSystemBackupService) RestoreCommand(ctx context.Context, id string) (service.SystemBackupRestoreCommand, error) {
	f.lastRestoreID = id
	if f.err != nil {
		return service.SystemBackupRestoreCommand{}, f.err
	}
	return f.command, nil
}

func (f *fakeSystemBackupService) OffsiteStatus(ctx context.Context) (service.SystemBackupOffsiteStatus, error) {
	if f.err != nil {
		return service.SystemBackupOffsiteStatus{}, f.err
	}
	return f.offsite, nil
}

func TestSystemBackupHandlerStatusAndList(t *testing.T) {
	now := time.Date(2026, 5, 13, 0, 0, 1, 0, time.UTC)
	h := NewSystemBackup(&fakeSystemBackupService{
		status: service.SystemBackupStatus{
			TimerName:     service.DefaultSystemBackupTimerName,
			ServiceName:   service.DefaultSystemBackupServiceName,
			TimerEnabled:  true,
			TimerActive:   true,
			RetentionDays: 30,
			BackupCount:   1,
			Health:        "ok",
			Warnings:      []string{},
		},
		list: service.SystemBackupList{
			Items: []service.SystemBackupFile{{ID: "pusaka_20260513_000001.dump", Name: "pusaka_20260513_000001.dump", CreatedAt: now, Downloadable: true}},
			Meta:  service.SystemBackupListMeta{Total: 1},
		},
	})

	statusRec := httptest.NewRecorder()
	h.Status(statusRec, httptest.NewRequest(http.MethodGet, "/api/system/backups/status", nil))
	if statusRec.Code != http.StatusOK {
		t.Fatalf("Status() code = %d body=%s", statusRec.Code, statusRec.Body.String())
	}
	var statusBody struct {
		Data service.SystemBackupStatus `json:"data"`
	}
	if err := json.Unmarshal(statusRec.Body.Bytes(), &statusBody); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	if statusBody.Data.Health != "ok" || statusBody.Data.BackupCount != 1 {
		t.Fatalf("status body = %+v", statusBody.Data)
	}

	listRec := httptest.NewRecorder()
	h.List(listRec, httptest.NewRequest(http.MethodGet, "/api/system/backups", nil))
	if listRec.Code != http.StatusOK {
		t.Fatalf("List() code = %d body=%s", listRec.Code, listRec.Body.String())
	}
	if !strings.Contains(listRec.Body.String(), "pusaka_20260513_000001.dump") {
		t.Fatalf("List() body missing backup name: %s", listRec.Body.String())
	}
}

func TestSystemBackupActionEndpoints(t *testing.T) {
	now := time.Date(2026, 5, 17, 20, 0, 0, 0, time.UTC)
	backupID := "pusaka_20260517_200000.dump"
	fake := &fakeSystemBackupService{
		job:        service.SystemBackupJob{ID: "job-1", Status: "running", Reason: "before deploy", StartedAt: now},
		validation: service.SystemBackupRestoreValidation{BackupID: backupID, Valid: true, ObjectCount: 12, Preview: []string{"public.users"}, CheckedAt: now, Command: "pg_restore --list"},
		command:    service.SystemBackupRestoreCommand{BackupID: backupID, GeneratedAt: now, SafetyLevel: "manual", Commands: []string{"pg_restore ..."}},
		offsite:    service.SystemBackupOffsiteStatus{Configured: true, Provider: "rclone", TargetLabel: "school-drive", Source: "status-file", RemoteBackupCount: 3, RemoteSizeBytes: 4096, Health: "ok"},
	}
	h := NewSystemBackup(fake)

	rec := httptest.NewRecorder()
	h.OffsiteStatus(rec, httptest.NewRequest(http.MethodGet, "/api/system/backups/offsite", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "school-drive") {
		t.Fatalf("OffsiteStatus status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.RunManual(rec, httptest.NewRequest(http.MethodPost, "/api/system/backups/run", strings.NewReader(`{"reason":"before deploy"}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("RunManual status=%d body=%s", rec.Code, rec.Body.String())
	}
	if fake.runReq.Reason != "before deploy" {
		t.Fatalf("RunManual req = %+v", fake.runReq)
	}

	rec = httptest.NewRecorder()
	h.Job(rec, withRouteParam(httptest.NewRequest(http.MethodGet, "/api/system/backups/jobs/job-1", nil), "job_id", "job-1"))
	if rec.Code != http.StatusOK || fake.lastJobID != "job-1" {
		t.Fatalf("Job status=%d lastID=%q body=%s", rec.Code, fake.lastJobID, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.ValidateRestore(rec, withRouteParam(httptest.NewRequest(http.MethodPost, "/api/system/backups/"+backupID+"/validate-restore", nil), "id", backupID))
	if rec.Code != http.StatusOK || fake.lastRestoreID != backupID || !strings.Contains(rec.Body.String(), "public.users") {
		t.Fatalf("ValidateRestore status=%d id=%q body=%s", rec.Code, fake.lastRestoreID, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.RestoreCommand(rec, withRouteParam(httptest.NewRequest(http.MethodGet, "/api/system/backups/"+backupID+"/restore-command", nil), "id", backupID))
	if rec.Code != http.StatusOK || fake.lastRestoreID != backupID || !strings.Contains(rec.Body.String(), "pg_restore") {
		t.Fatalf("RestoreCommand status=%d id=%q body=%s", rec.Code, fake.lastRestoreID, rec.Body.String())
	}
}

func TestSystemBackupActionErrors(t *testing.T) {
	for _, tc := range []struct {
		name string
		fn   func(*SystemBackup, http.ResponseWriter, *http.Request)
		err  error
		want int
	}{
		{"run invalid json", (*SystemBackup).RunManual, nil, http.StatusBadRequest},
		{"run conflict", (*SystemBackup).RunManual, domain.ErrConflict, http.StatusConflict},
		{"job not found", (*SystemBackup).Job, domain.ErrNotFound, http.StatusNotFound},
		{"validate bad request", (*SystemBackup).ValidateRestore, domain.ErrBadRequest, http.StatusBadRequest},
		{"restore forbidden", (*SystemBackup).RestoreCommand, domain.ErrForbidden, http.StatusForbidden},
		{"offsite internal", (*SystemBackup).OffsiteStatus, errors.New("status read failed"), http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeSystemBackupService{err: tc.err}
			h := NewSystemBackup(fake)
			req := httptest.NewRequest(http.MethodPost, "/api/system/backups/run", strings.NewReader(`{"reason":"x"}`))
			if strings.Contains(tc.name, "invalid json") {
				req = httptest.NewRequest(http.MethodPost, "/api/system/backups/run", strings.NewReader(`{`))
			}
			if strings.HasPrefix(tc.name, "job") {
				req = withRouteParam(httptest.NewRequest(http.MethodGet, "/api/system/backups/jobs/missing", nil), "job_id", "missing")
			}
			if strings.HasPrefix(tc.name, "validate") || strings.HasPrefix(tc.name, "restore") {
				req = withRouteParam(httptest.NewRequest(http.MethodGet, "/api/system/backups/x/restore", nil), "id", "x.dump")
			}
			if strings.HasPrefix(tc.name, "offsite") {
				req = httptest.NewRequest(http.MethodGet, "/api/system/backups/offsite", nil)
			}
			rec := httptest.NewRecorder()
			tc.fn(h, rec, req)
			if rec.Code != tc.want {
				t.Fatalf("status=%d want=%d body=%s", rec.Code, tc.want, rec.Body.String())
			}
		})
	}

	fake := &fakeSystemBackupService{job: service.SystemBackupJob{ID: "job-2"}, err: errors.New("partial failure")}
	rec := httptest.NewRecorder()
	NewSystemBackup(fake).RunManual(rec, httptest.NewRequest(http.MethodPost, "/api/system/backups/run", strings.NewReader(`{}`)))
	if rec.Code != http.StatusInternalServerError || !strings.Contains(rec.Body.String(), "job-2") {
		t.Fatalf("partial RunManual status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestSystemBackupHandlerDownloadStreamsAttachment(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pusaka_20260513_000001.dump")
	if err := os.WriteFile(path, []byte("dump-data"), 0o600); err != nil {
		t.Fatalf("write dump: %v", err)
	}
	h := NewSystemBackup(&fakeSystemBackupService{
		download: service.SystemBackupDownload{
			Path:     path,
			Filename: filepath.Base(path),
			ModTime:  time.Date(2026, 5, 13, 0, 0, 1, 0, time.UTC),
		},
	})

	req := withRouteParam(httptest.NewRequest(http.MethodGet, "/api/system/backups/pusaka_20260513_000001.dump/download", nil), "id", filepath.Base(path))
	rec := httptest.NewRecorder()
	h.Download(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Download() code = %d body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Disposition"); !strings.Contains(got, `attachment; filename="pusaka_20260513_000001.dump"`) {
		t.Fatalf("Content-Disposition = %q", got)
	}
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q", got)
	}
	if rec.Body.String() != "dump-data" {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestSystemBackupHandlerDownloadMapsSafeErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{name: "forbidden", err: domain.ErrForbidden, wantStatus: http.StatusForbidden},
		{name: "not found", err: domain.ErrNotFound, wantStatus: http.StatusNotFound},
		{name: "bad request", err: domain.ErrBadRequest, wantStatus: http.StatusBadRequest},
		{name: "internal", err: errors.New("disk failure"), wantStatus: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewSystemBackup(&fakeSystemBackupService{err: tt.err})
			req := withRouteParam(httptest.NewRequest(http.MethodGet, "/api/system/backups/x.dump/download", nil), "id", "x.dump")
			rec := httptest.NewRecorder()
			h.Download(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("Download() code = %d body=%s, want %d", rec.Code, rec.Body.String(), tt.wantStatus)
			}
			if tt.wantStatus == http.StatusInternalServerError && strings.Contains(rec.Body.String(), "disk failure") {
				t.Fatalf("internal detail leaked: %s", rec.Body.String())
			}
		})
	}
}
