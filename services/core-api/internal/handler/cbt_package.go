package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type CbtPackage struct {
	svc   cbtPackageService
	audit cbtAuthoringAuditWriter
}

func NewCbtPackage(svc *service.CbtPackage, audit ...cbtAuthoringAuditWriter) *CbtPackage {
	var writer cbtAuthoringAuditWriter
	if len(audit) > 0 {
		writer = audit[0]
	}
	return &CbtPackage{svc: svc, audit: writer}
}

type cbtPackageService interface {
	List(ctx context.Context) ([]db.ListCbtPackagesRow, []db.ListCbtPackageQuestionsRow, error)
	Create(ctx context.Context, input service.CreateCbtPackageInput) (db.CbtPackage, error)
	Delete(ctx context.Context, id pgtype.UUID) error
}

func (h *CbtPackage) List(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	packages, questions, err := h.svc.List(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{
		"packages":  packages,
		"questions": questions,
	})
}

func (h *CbtPackage) Create(w http.ResponseWriter, r *http.Request) {
	if !hasAnyRole(r, "admin") {
		api.Forbidden(w)
		return
	}
	var body struct {
		SubjectID          string   `json:"subject_id"`
		Title              string   `json:"title"`
		Description        string   `json:"description"`
		DurationMinutes    int32    `json:"duration_minutes"`
		RandomizeQuestions bool     `json:"randomize_questions"`
		IsActive           bool     `json:"is_active"`
		QuestionIDs        []string `json:"question_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	subjectID, err := parseUUID(body.SubjectID)
	if err != nil {
		api.BadRequest(w, "subject_id invalid")
		return
	}
	questionIDs := make([]pgtype.UUID, 0, len(body.QuestionIDs))
	for _, rawID := range body.QuestionIDs {
		id, err := parseUUID(rawID)
		if err != nil {
			api.BadRequest(w, "question_id invalid")
			return
		}
		questionIDs = append(questionIDs, id)
	}
	row, err := h.svc.Create(r.Context(), service.CreateCbtPackageInput{
		SubjectID:          subjectID,
		Title:              body.Title,
		Description:        body.Description,
		DurationMinutes:    body.DurationMinutes,
		RandomizeQuestions: body.RandomizeQuestions,
		IsActive:           body.IsActive,
		QuestionIDs:        questionIDs,
	})
	if err != nil {
		writeClientError(w, err, "Paket CBT tidak valid")
		return
	}
	h.auditPackageEvent(r.Context(), "CBT_PACKAGE_CREATE", "cbt_package", pgUUIDString(row.ID), map[string]any{
		"subject_id":     pgUUIDString(row.SubjectID),
		"title":          row.Title,
		"question_count": len(questionIDs),
	})
	api.Created(w, row)
}

func (h *CbtPackage) Delete(w http.ResponseWriter, r *http.Request) {
	if !hasAnyRole(r, "admin") {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		api.Internal(w, err)
		return
	}
	h.auditPackageEvent(r.Context(), "CBT_PACKAGE_DELETE", "cbt_package", pgUUIDString(id), nil)
	api.NoContent(w)
}

func (h *CbtPackage) auditPackageEvent(ctx context.Context, action, entityType, entityID string, extra map[string]any) {
	if h.audit == nil {
		return
	}
	claims, ok := api.ClaimsFromContext(ctx)
	if !ok {
		return
	}
	userID, err := authUserID(claims)
	if err != nil {
		return
	}
	meta := map[string]any{
		"username":   claims["usr"],
		"user_id":    claims["uid"],
		"session_id": claims["ssid"],
	}
	for key, value := range extra {
		meta[key] = value
	}
	rawMeta, err := json.Marshal(meta)
	if err != nil {
		return
	}
	_, _ = h.audit.CreateAuditLog(ctx, db.CreateAuditLogParams{
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Metadata:   rawMeta,
	})
}
