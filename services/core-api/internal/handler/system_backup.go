package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/domain"
	"mtsn2kolut-super-app/backend/internal/service"
)

type SystemBackup struct {
	svc systemBackupService
}

type systemBackupService interface {
	Status(ctx context.Context) (service.SystemBackupStatus, error)
	List(ctx context.Context) (service.SystemBackupList, error)
	Download(ctx context.Context, id string) (service.SystemBackupDownload, error)
	RunManual(ctx context.Context, req service.SystemBackupRunRequest) (service.SystemBackupJob, error)
	Job(ctx context.Context, id string) (service.SystemBackupJob, error)
}

func NewSystemBackup(svc systemBackupService) *SystemBackup {
	return &SystemBackup{svc: svc}
}

func (h *SystemBackup) Status(w http.ResponseWriter, r *http.Request) {
	status, err := h.svc.Status(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, status)
}

func (h *SystemBackup) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.List(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, list)
}

func (h *SystemBackup) RunManual(w http.ResponseWriter, r *http.Request) {
	var req service.SystemBackupRunRequest
	if r.Body != nil {
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
			api.BadRequest(w, "payload backup tidak valid")
			return
		}
	}
	job, err := h.svc.RunManual(r.Context(), req)
	if err != nil {
		if job.ID != "" {
			api.JSON(w, http.StatusInternalServerError, api.Response{Data: job, Error: "backup manual gagal"})
			return
		}
		writeSystemBackupError(w, err)
		return
	}
	api.OK(w, job)
}

func (h *SystemBackup) Job(w http.ResponseWriter, r *http.Request) {
	job, err := h.svc.Job(r.Context(), chi.URLParam(r, "job_id"))
	if err != nil {
		writeSystemBackupError(w, err)
		return
	}
	api.OK(w, job)
}

func (h *SystemBackup) Download(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	download, err := h.svc.Download(r.Context(), id)
	if err != nil {
		writeSystemBackupError(w, err)
		return
	}
	file, err := os.Open(download.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			api.NotFound(w)
			return
		}
		api.Internal(w, err)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="`+contentDispositionFilename(download.Filename)+`"`)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, download.Filename, download.ModTime, file)
}

func writeSystemBackupError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrBadRequest):
		api.BadRequest(w, "backup tidak valid")
	case errors.Is(err, domain.ErrForbidden):
		api.Forbidden(w)
	case errors.Is(err, domain.ErrConflict):
		api.Conflict(w, "backup manual sedang berjalan")
	case errors.Is(err, domain.ErrNotFound):
		api.NotFound(w)
	default:
		api.Internal(w, err)
	}
}
