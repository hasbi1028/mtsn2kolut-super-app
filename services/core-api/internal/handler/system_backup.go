package handler

import (
	"context"
	"errors"
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
	case errors.Is(err, domain.ErrNotFound):
		api.NotFound(w)
	default:
		api.Internal(w, err)
	}
}
