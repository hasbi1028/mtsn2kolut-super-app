package handler

import (
	"context"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	mw "mtsn2kolut-super-app/backend/internal/middleware"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type CbtQuestionAsset struct {
	svc cbtQuestionAssetService
}

type cbtQuestionAssetService interface {
	Get(ctx context.Context, id pgtype.UUID) (db.CbtQuestionAsset, error)
	Open(asset db.CbtQuestionAsset) (*os.File, error)
	AccessibleByPackage(ctx context.Context, questionID, packageID pgtype.UUID) (bool, error)
}

func NewCbtQuestionAsset(svc *service.CbtQuestionAsset) *CbtQuestionAsset {
	return &CbtQuestionAsset{svc: svc}
}

func (h *CbtQuestionAsset) File(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	asset, err := h.svc.Get(r.Context(), id)
	if err != nil {
		api.NotFound(w)
		return
	}
	if participant, ok := mw.ParticipantFromContext(r.Context()); ok {
		allowed, err := h.svc.AccessibleByPackage(r.Context(), asset.QuestionID, participant.PackageID)
		if err != nil {
			api.Internal(w, err)
			return
		}
		if !allowed {
			api.Forbidden(w)
			return
		}
	} else if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	f, err := h.svc.Open(asset)
	if err != nil {
		api.NotFound(w)
		return
	}
	defer f.Close()

	secureFileResponseHeaders(w, asset.MimeType, asset.StoredName)
	w.Header().Set("Content-Length", strconv.FormatInt(asset.FileSize, 10))
	w.Header().Set("Cache-Control", "private, max-age=300")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, f)
}
