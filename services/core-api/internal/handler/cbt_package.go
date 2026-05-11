package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

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

func optionalUUIDQuery(r *http.Request, key string) (pgtype.UUID, error) {
	return parseOptionalUUID(r.URL.Query().Get(key))
}

func parseOptionalUUID(raw string) (pgtype.UUID, error) {
	if strings.TrimSpace(raw) == "" {
		return pgtype.UUID{}, nil
	}
	return parseUUID(raw)
}

func NewCbtPackage(svc *service.CbtPackage, audit ...cbtAuthoringAuditWriter) *CbtPackage {
	var writer cbtAuthoringAuditWriter
	if len(audit) > 0 {
		writer = audit[0]
	}
	return &CbtPackage{svc: svc, audit: writer}
}

type cbtPackageService interface {
	List(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtPackagesRow, []db.ListCbtPackageQuestionsRow, error)
	Create(ctx context.Context, input service.CreateCbtPackageInput) (db.CbtPackage, error)
	Delete(ctx context.Context, id pgtype.UUID) error
}

func (h *CbtPackage) List(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	eventID, err := optionalUUIDQuery(r, "event_id")
	if err != nil {
		api.BadRequest(w, "event_id invalid")
		return
	}
	packages, questions, err := h.svc.List(r.Context(), eventID)
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
		EventID            string           `json:"event_id"`
		SubjectID          string           `json:"subject_id"`
		Title              string           `json:"title"`
		Description        string           `json:"description"`
		DurationMinutes    int32            `json:"duration_minutes"`
		RandomizeQuestions bool             `json:"randomize_questions"`
		RandomizeOptions   bool             `json:"randomize_options"`
		SourceMode         string           `json:"source_mode"`
		DrawPgCount        int32            `json:"draw_pg_count"`
		DrawEssayCount     int32            `json:"draw_essay_count"`
		RandomSeed         string           `json:"random_seed"`
		IsActive           bool             `json:"is_active"`
		QuestionIDs        []string         `json:"question_ids"`
		QuestionWeights    map[string]int32 `json:"question_weights"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	if body.DurationMinutes < 1 || body.DurationMinutes > 360 {
		api.BadRequest(w, "duration_minutes harus 1-360")
		return
	}
	eventID, err := parseOptionalUUID(body.EventID)
	if err != nil {
		api.BadRequest(w, "event_id invalid")
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
		EventID:            eventID,
		SubjectID:          subjectID,
		Title:              body.Title,
		Description:        body.Description,
		DurationMinutes:    body.DurationMinutes,
		RandomizeQuestions: body.RandomizeQuestions,
		RandomizeOptions:   body.RandomizeOptions,
		SourceMode:         body.SourceMode,
		DrawPgCount:        body.DrawPgCount,
		DrawEssayCount:     body.DrawEssayCount,
		RandomSeed:         body.RandomSeed,
		IsActive:           body.IsActive,
		QuestionIDs:        questionIDs,
		QuestionWeights:    body.QuestionWeights,
	})
	if err != nil {
		writeClientError(w, err, "Paket CBT tidak valid")
		return
	}
	h.auditPackageEvent(r.Context(), "CBT_PACKAGE_CREATE", "cbt_package", pgUUIDString(row.ID), map[string]any{
		"event_id":            pgUUIDString(row.EventID),
		"subject_id":          pgUUIDString(row.SubjectID),
		"title":               row.Title,
		"question_count":      len(questionIDs),
		"source_mode":         body.SourceMode,
		"randomize_questions": body.RandomizeQuestions,
		"randomize_options":   body.RandomizeOptions,
		"draw_pg_count":       body.DrawPgCount,
		"draw_essay_count":    body.DrawEssayCount,
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
		writeDomainOrInternal(w, err, "Hapus paket CBT tidak valid")
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
