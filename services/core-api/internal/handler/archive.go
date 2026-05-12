package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

const maxArchiveUploadBytes = 25 * 1024 * 1024

type Archive struct {
	svc   archiveService
	audit cbtAuthoringAuditWriter
}

type archiveService interface {
	Stats(ctx context.Context) (db.GetArchiveStatsRow, error)
	ListCategories(ctx context.Context, search string, activeOnly bool) ([]db.ListArchiveCategoriesRow, error)
	CreateCategory(ctx context.Context, arg db.CreateArchiveCategoryParams) (db.ArchiveCategory, error)
	UpdateCategory(ctx context.Context, arg db.UpdateArchiveCategoryParams) (db.ArchiveCategory, error)
	DeleteCategory(ctx context.Context, id pgtype.UUID) error
	ListDocuments(ctx context.Context, search string, categoryID pgtype.UUID, status, classificationCode string) ([]db.ListArchiveDocumentsRow, error)
	SaveDocument(ctx context.Context, input service.UploadArchiveDocumentInput) (db.ArchiveDocument, error)
	GetDocument(ctx context.Context, id pgtype.UUID) (db.GetArchiveDocumentDetailRow, error)
	UpdateDocument(ctx context.Context, arg db.UpdateArchiveDocumentParams) (db.ArchiveDocument, error)
	DeleteDocument(ctx context.Context, id pgtype.UUID) error
	GetDocumentFile(ctx context.Context, id pgtype.UUID) (db.ArchiveDocument, error)
}

func NewArchive(svc *service.Archive, audit ...cbtAuthoringAuditWriter) *Archive {
	var writer cbtAuthoringAuditWriter
	if len(audit) > 0 {
		writer = audit[0]
	}
	return &Archive{svc: svc, audit: writer}
}

func (h *Archive) Stats(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	data, err := h.svc.Stats(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

func (h *Archive) ListCategories(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	activeOnly := r.URL.Query().Get("active_only") == "true"
	data, err := h.svc.ListCategories(r.Context(), r.URL.Query().Get("search"), activeOnly)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

func (h *Archive) CreateCategory(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body archiveCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	active := true
	if body.IsActive != nil {
		active = *body.IsActive
	}
	row, err := h.svc.CreateCategory(r.Context(), db.CreateArchiveCategoryParams{
		Code:               body.Code,
		Name:               body.Name,
		ClassificationCode: body.ClassificationCode,
		Description:        body.Description,
		RetentionYears:     body.RetentionYears,
		IsActive:           active,
	})
	if err != nil {
		api.BadRequest(w, archiveClientMessage(err))
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "ARCHIVE_CATEGORY_CREATE", "archive_category", pgUUIDString(row.ID), map[string]any{
		"code": row.Code,
		"name": row.Name,
	})
	api.Created(w, row)
}

func (h *Archive) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	var body archiveCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	active := true
	if body.IsActive != nil {
		active = *body.IsActive
	}
	row, err := h.svc.UpdateCategory(r.Context(), db.UpdateArchiveCategoryParams{
		ID:                 id,
		Code:               body.Code,
		Name:               body.Name,
		ClassificationCode: body.ClassificationCode,
		Description:        body.Description,
		RetentionYears:     body.RetentionYears,
		IsActive:           active,
	})
	if err != nil {
		api.BadRequest(w, archiveClientMessage(err))
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "ARCHIVE_CATEGORY_UPDATE", "archive_category", pgUUIDString(row.ID), map[string]any{
		"code": row.Code,
		"name": row.Name,
	})
	api.OK(w, row)
}

func (h *Archive) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	if err := h.svc.DeleteCategory(r.Context(), id); err != nil {
		api.BadRequest(w, archiveClientMessage(err))
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "ARCHIVE_CATEGORY_DELETE", "archive_category", pgUUIDString(id), map[string]any{
		"deleted_by": currentUsername(r),
	})
	api.NoContent(w)
}

func (h *Archive) ListDocuments(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	categoryID, err := service.ParseArchiveOptionalUUID(r.URL.Query().Get("category_id"))
	if err != nil {
		writeClientError(w, err, "Filter arsip tidak valid")
		return
	}
	data, err := h.svc.ListDocuments(
		r.Context(),
		r.URL.Query().Get("search"),
		categoryID,
		r.URL.Query().Get("status"),
		r.URL.Query().Get("classification_code"),
	)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

func (h *Archive) UploadDocument(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 28<<20)
	if err := r.ParseMultipartForm(28 << 20); err != nil {
		api.BadRequest(w, "Berkas/formulir yang dikirim tidak valid")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		api.BadRequest(w, "file arsip wajib diisi")
		return
	}
	defer file.Close()
	validated, err := validateUploadedFile(header.Filename, file, maxArchiveUploadBytes, false)
	if err != nil {
		api.BadRequest(w, archiveClientMessage(err))
		return
	}

	categoryID, err := service.ParseArchiveOptionalUUID(r.FormValue("category_id"))
	if err != nil {
		writeClientError(w, err, "Data arsip tidak valid")
		return
	}
	documentDate, err := service.ParseArchiveOptionalDate(r.FormValue("document_date"))
	if err != nil {
		writeClientError(w, err, "Data arsip tidak valid")
		return
	}
	receivedDate, err := service.ParseArchiveOptionalDate(r.FormValue("received_date"))
	if err != nil {
		writeClientError(w, err, "Data arsip tidak valid")
		return
	}
	retentionUntil, err := service.ParseArchiveOptionalDate(r.FormValue("retention_until"))
	if err != nil {
		writeClientError(w, err, "Data arsip tidak valid")
		return
	}

	row, err := h.svc.SaveDocument(r.Context(), service.UploadArchiveDocumentInput{
		CategoryID:       categoryID,
		Title:            r.FormValue("title"),
		ArchiveNumber:    r.FormValue("archive_number"),
		DocumentDate:     documentDate,
		ReceivedDate:     receivedDate,
		Summary:          r.FormValue("summary"),
		Tags:             r.FormValue("tags"),
		Status:           r.FormValue("status"),
		StorageLocation:  r.FormValue("storage_location"),
		RetentionUntil:   retentionUntil,
		OriginalName:     header.Filename,
		MimeType:         validated.MimeType,
		FileSize:         int64(len(validated.Data)),
		UploadedByUserID: inventoryActorUserID(r),
		File:             bytes.NewReader(validated.Data),
	})
	if err != nil {
		api.BadRequest(w, archiveClientMessage(err))
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "ARCHIVE_DOCUMENT_UPLOAD", "archive_document", pgUUIDString(row.ID), map[string]any{
		"title":          row.Title,
		"archive_number": row.ArchiveNumber,
		"original_name":  row.OriginalName,
	})
	api.Created(w, row)
}

func (h *Archive) GetDocument(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	row, err := h.svc.GetDocument(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			api.NotFound(w)
			return
		}
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
}

func (h *Archive) UpdateDocument(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	var body archiveDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	categoryID, err := service.ParseArchiveOptionalUUID(body.CategoryID)
	if err != nil {
		writeClientError(w, err, "Data arsip tidak valid")
		return
	}
	documentDate, err := service.ParseArchiveOptionalDate(body.DocumentDate)
	if err != nil {
		writeClientError(w, err, "Data arsip tidak valid")
		return
	}
	receivedDate, err := service.ParseArchiveOptionalDate(body.ReceivedDate)
	if err != nil {
		writeClientError(w, err, "Data arsip tidak valid")
		return
	}
	retentionUntil, err := service.ParseArchiveOptionalDate(body.RetentionUntil)
	if err != nil {
		writeClientError(w, err, "Data arsip tidak valid")
		return
	}
	row, err := h.svc.UpdateDocument(r.Context(), db.UpdateArchiveDocumentParams{
		ID:              id,
		CategoryID:      categoryID,
		Title:           body.Title,
		ArchiveNumber:   body.ArchiveNumber,
		DocumentDate:    documentDate,
		ReceivedDate:    receivedDate,
		Summary:         body.Summary,
		Tags:            body.Tags,
		Status:          body.Status,
		StorageLocation: body.StorageLocation,
		RetentionUntil:  retentionUntil,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			api.NotFound(w)
			return
		}
		api.BadRequest(w, archiveClientMessage(err))
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "ARCHIVE_DOCUMENT_UPDATE", "archive_document", pgUUIDString(row.ID), map[string]any{
		"title":          row.Title,
		"archive_number": row.ArchiveNumber,
		"status":         row.Status,
	})
	api.OK(w, row)
}

func (h *Archive) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	if err := h.svc.DeleteDocument(r.Context(), id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			api.NotFound(w)
			return
		}
		api.BadRequest(w, archiveClientMessage(err))
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "ARCHIVE_DOCUMENT_DELETE", "archive_document", pgUUIDString(id), map[string]any{
		"deleted_by": currentUsername(r),
	})
	api.NoContent(w)
}

func (h *Archive) File(w http.ResponseWriter, r *http.Request) {
	if !tuAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "id tidak valid")
		return
	}
	document, err := h.svc.GetDocumentFile(r.Context(), id)
	if err != nil {
		api.NotFound(w)
		return
	}
	f, err := os.Open(document.FilePath)
	if err != nil {
		api.NotFound(w)
		return
	}
	defer f.Close()

	secureFileResponseHeaders(w, document.MimeType, document.OriginalName)
	w.Header().Set("Content-Length", strconv.FormatInt(document.FileSize, 10))
	w.Header().Set("Cache-Control", "private, max-age=300")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, f)
}

type archiveCategoryRequest struct {
	Code               string `json:"code"`
	Name               string `json:"name"`
	ClassificationCode string `json:"classification_code"`
	Description        string `json:"description"`
	RetentionYears     int32  `json:"retention_years"`
	IsActive           *bool  `json:"is_active"`
}

type archiveDocumentRequest struct {
	CategoryID      string `json:"category_id"`
	Title           string `json:"title"`
	ArchiveNumber   string `json:"archive_number"`
	DocumentDate    string `json:"document_date"`
	ReceivedDate    string `json:"received_date"`
	Summary         string `json:"summary"`
	Tags            string `json:"tags"`
	Status          string `json:"status"`
	StorageLocation string `json:"storage_location"`
	RetentionUntil  string `json:"retention_until"`
}

func archiveClientMessage(err error) string {
	return safeClientMessage(err, "data arsip tidak valid")
}
