package handler

import (
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	mw "mtsn2kolut-super-app/backend/internal/middleware"
	"mtsn2kolut-super-app/backend/internal/service"
)

type CbtQuestionAsset struct {
	svc *service.CbtQuestionAsset
}

func NewCbtQuestionAsset(svc *service.CbtQuestionAsset) *CbtQuestionAsset {
	return &CbtQuestionAsset{svc: svc}
}

func (h *CbtQuestionAsset) Upload(w http.ResponseWriter, r *http.Request) {
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
		api.BadRequest(w, err.Error())
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
