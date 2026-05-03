package handler

import (
	"context"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

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
	Save(ctx context.Context, input service.UploadCbtQuestionAssetInput) (db.CbtQuestionAsset, error)
	ListByQuestion(ctx context.Context, questionID pgtype.UUID) ([]db.CbtQuestionAsset, error)
	GetQuestion(ctx context.Context, id pgtype.UUID) (db.GetCbtQuestionRow, error)
	Get(ctx context.Context, id pgtype.UUID) (db.CbtQuestionAsset, error)
	AccessibleByPackage(ctx context.Context, questionID, packageID pgtype.UUID) (bool, error)
}

func NewCbtQuestionAsset(svc *service.CbtQuestionAsset) *CbtQuestionAsset {
	return &CbtQuestionAsset{svc: svc}
}

func (h *CbtQuestionAsset) Upload(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	if err := r.ParseMultipartForm(12 << 20); err != nil {
		api.BadRequest(w, "multipart form invalid")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		api.BadRequest(w, "file wajib diisi")
		return
	}
	defer file.Close()

	questionID := pgtype.UUID{}
	if rawID := r.FormValue("question_id"); rawID != "" {
		parsed, err := parseUUID(rawID)
		if err != nil {
			api.BadRequest(w, "question_id invalid")
			return
		}
		questionID = parsed
	}
	if !h.requireQuestionAssetScope(w, r, questionID) {
		return
	}

	row, err := h.svc.Save(r.Context(), service.UploadCbtQuestionAssetInput{
		QuestionID:   questionID,
		OriginalName: header.Filename,
		MimeType:     header.Header.Get("Content-Type"),
		FileSize:     header.Size,
		Purpose:      r.FormValue("purpose"),
		UploadedBy:   currentUsername(r),
		File:         file,
	})
	if err != nil {
		writeClientError(w, err, "Upload aset soal CBT tidak valid")
		return
	}

	api.Created(w, map[string]any{
		"id":            pgUUIDString(row.ID),
		"question_id":   pgUUIDString(row.QuestionID),
		"original_name": row.OriginalName,
		"mime_type":     row.MimeType,
		"file_size":     row.FileSize,
		"purpose":       row.Purpose,
		"url":           "/api/cbt/assets/" + pgUUIDString(row.ID) + "/file",
	})
}

func (h *CbtQuestionAsset) List(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	rawID := r.URL.Query().Get("question_id")
	if rawID == "" {
		api.BadRequest(w, "question_id wajib diisi")
		return
	}
	questionID, err := parseUUID(rawID)
	if err != nil {
		api.BadRequest(w, "question_id invalid")
		return
	}
	if !h.requireQuestionAssetScope(w, r, questionID) {
		return
	}
	rows, err := h.svc.ListByQuestion(r.Context(), questionID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	items := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		items = append(items, map[string]any{
			"id":            pgUUIDString(row.ID),
			"question_id":   pgUUIDString(row.QuestionID),
			"original_name": row.OriginalName,
			"mime_type":     row.MimeType,
			"file_size":     row.FileSize,
			"purpose":       row.Purpose,
			"url":           "/api/cbt/assets/" + pgUUIDString(row.ID) + "/file",
		})
	}
	api.OK(w, items)
}

func (h *CbtQuestionAsset) requireQuestionAssetScope(w http.ResponseWriter, r *http.Request, questionID pgtype.UUID) bool {
	if hasAnyRole(r, "admin") {
		return true
	}
	if !hasAnyRole(r, "guru") {
		api.Forbidden(w)
		return false
	}
	if !questionID.Valid {
		api.Forbidden(w)
		return false
	}
	username := currentUsername(r)
	if strings.TrimSpace(username) == "" {
		api.Forbidden(w)
		return false
	}
	question, err := h.svc.GetQuestion(r.Context(), questionID)
	if err != nil {
		api.Internal(w, err)
		return false
	}
	if strings.TrimSpace(question.AuthorUsername) != username {
		api.Forbidden(w)
		return false
	}
	return true
}

func (h *CbtQuestionAsset) File(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
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
	} else if !h.requireQuestionAssetScope(w, r, asset.QuestionID) {
		return
	}
	f, err := os.Open(asset.StoragePath)
	if err != nil {
		api.NotFound(w)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", asset.MimeType)
	w.Header().Set("Content-Length", strconv.FormatInt(asset.FileSize, 10))
	w.Header().Set("Content-Disposition", `inline; filename="`+asset.StoredName+`"`)
	w.Header().Set("Cache-Control", "private, max-age=300")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, f)
}
