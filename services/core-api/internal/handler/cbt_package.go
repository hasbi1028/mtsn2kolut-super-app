package handler

import (
	"context"
	"encoding/json"
	"io"
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
	Detail(ctx context.Context, id pgtype.UUID) (service.CbtPackageDetailResult, error)
	Readiness(ctx context.Context, eventID pgtype.UUID) (service.CbtPackageReadinessOverview, error)
	Create(ctx context.Context, input service.CreateCbtPackageInput) (db.CbtPackage, error)
	UpdateMetadata(ctx context.Context, input service.UpdateCbtPackageInput) (service.CbtPackageDetailResult, error)
	ReplaceQuestions(ctx context.Context, input service.ReplaceCbtPackageQuestionsInput) (service.CbtPackageDetailResult, error)
	Clone(ctx context.Context, input service.CloneCbtPackageInput) (service.CbtPackageDetailResult, error)
	LockAndSnapshot(ctx context.Context, packageID, lockedBy pgtype.UUID, reason string) (service.CbtPackageSnapshotResult, error)
	Archive(ctx context.Context, id, archivedBy pgtype.UUID, reason string) error
	Delete(ctx context.Context, id pgtype.UUID) error
}

func (h *CbtPackage) List(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	eventID, err := optionalUUIDQuery(r, "event_id")
	if err != nil {
		api.BadRequest(w, "Kegiatan asesmen tidak valid")
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

func (h *CbtPackage) Readiness(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	eventID, err := optionalUUIDQuery(r, "event_id")
	if err != nil {
		api.BadRequest(w, "Kegiatan asesmen tidak valid")
		return
	}
	overview, err := h.svc.Readiness(r.Context(), eventID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, overview)
}

func (h *CbtPackage) Get(w http.ResponseWriter, r *http.Request) {
	if !cbtAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID paket tidak valid")
		return
	}
	detail, err := h.svc.Detail(r.Context(), id)
	if err != nil {
		writeDomainOrInternal(w, err, "Paket CBT tidak ditemukan")
		return
	}
	api.OK(w, detail)
}

func (h *CbtPackage) Create(w http.ResponseWriter, r *http.Request) {
	if !packageWriteAllowed(r) {
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
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	if body.DurationMinutes < 1 || body.DurationMinutes > 360 {
		api.BadRequest(w, "duration_minutes harus 1-360")
		return
	}
	eventID, err := parseOptionalUUID(body.EventID)
	if err != nil {
		api.BadRequest(w, "Kegiatan asesmen tidak valid")
		return
	}
	subjectID, err := parseUUID(body.SubjectID)
	if err != nil {
		api.BadRequest(w, "Mata pelajaran tidak valid")
		return
	}
	questionIDs := make([]pgtype.UUID, 0, len(body.QuestionIDs))
	for _, rawID := range body.QuestionIDs {
		id, err := parseUUID(rawID)
		if err != nil {
			api.BadRequest(w, "Soal tidak valid")
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

func (h *CbtPackage) Update(w http.ResponseWriter, r *http.Request) {
	if !packageWriteAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID paket tidak valid")
		return
	}
	var body struct {
		Title              string `json:"title"`
		Description        string `json:"description"`
		DurationMinutes    int32  `json:"duration_minutes"`
		RandomizeQuestions bool   `json:"randomize_questions"`
		RandomizeOptions   bool   `json:"randomize_options"`
		SourceMode         string `json:"source_mode"`
		DrawPgCount        int32  `json:"draw_pg_count"`
		DrawEssayCount     int32  `json:"draw_essay_count"`
		RandomSeed         string `json:"random_seed"`
		IsActive           bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	detail, err := h.svc.UpdateMetadata(r.Context(), service.UpdateCbtPackageInput{
		ID:                 id,
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
	})
	if err != nil {
		writeDomainOrInternal(w, err, "Update paket CBT tidak valid")
		return
	}
	h.auditPackageEvent(r.Context(), "CBT_PACKAGE_UPDATE", "cbt_package", pgUUIDString(id), map[string]any{
		"title":               body.Title,
		"duration_minutes":    body.DurationMinutes,
		"randomize_questions": body.RandomizeQuestions,
		"randomize_options":   body.RandomizeOptions,
		"is_active":           body.IsActive,
	})
	api.OK(w, detail)
}

func (h *CbtPackage) ReplaceQuestions(w http.ResponseWriter, r *http.Request) {
	if !packageWriteAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID paket tidak valid")
		return
	}
	var body struct {
		QuestionIDs     []string         `json:"question_ids"`
		QuestionWeights map[string]int32 `json:"question_weights"`
		Questions       []struct {
			QuestionID string `json:"question_id"`
			Points     int32  `json:"points"`
		} `json:"questions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	questionIDs, weights, ok := packageQuestionInput(body.QuestionIDs, body.QuestionWeights, body.Questions)
	if !ok {
		api.BadRequest(w, "Soal paket tidak valid")
		return
	}
	detail, err := h.svc.ReplaceQuestions(r.Context(), service.ReplaceCbtPackageQuestionsInput{
		PackageID:       id,
		QuestionIDs:     questionIDs,
		QuestionWeights: weights,
	})
	if err != nil {
		writeDomainOrInternal(w, err, "Update soal paket CBT tidak valid")
		return
	}
	h.auditPackageEvent(r.Context(), "CBT_PACKAGE_REPLACE_QUESTIONS", "cbt_package", pgUUIDString(id), map[string]any{
		"question_count": len(questionIDs),
	})
	api.OK(w, detail)
}

func (h *CbtPackage) Clone(w http.ResponseWriter, r *http.Request) {
	if !packageWriteAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID paket tidak valid")
		return
	}
	var body struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	detail, err := h.svc.Clone(r.Context(), service.CloneCbtPackageInput{SourceID: id, Title: body.Title})
	if err != nil {
		writeDomainOrInternal(w, err, "Clone paket CBT tidak valid")
		return
	}
	h.auditPackageEvent(r.Context(), "CBT_PACKAGE_CLONE", "cbt_package", pgUUIDString(detail.Package.ID), map[string]any{
		"source_package_id": pgUUIDString(id),
		"title":             detail.Package.Title,
	})
	api.Created(w, detail)
}

func (h *CbtPackage) Lock(w http.ResponseWriter, r *http.Request) {
	if !packageWriteAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID paket tidak valid")
		return
	}
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
		return
	}
	userID, err := authUserID(claims)
	if err != nil {
		api.Unauthorized(w)
		return
	}
	var body struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil && err != io.EOF {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	result, err := h.svc.LockAndSnapshot(r.Context(), id, userID, body.Reason)
	if err != nil {
		writeDomainOrInternal(w, err, "Lock paket CBT tidak valid")
		return
	}
	h.auditPackageEvent(r.Context(), "CBT_PACKAGE_LOCK", "cbt_package", pgUUIDString(id), map[string]any{
		"lock_reason":         result.LockReason,
		"snapshot_version":    result.SnapshotVersion,
		"snapshot_rows_added": result.SnapshotRowsAdded,
	})
	api.OK(w, result)
}

func (h *CbtPackage) Delete(w http.ResponseWriter, r *http.Request) {
	if !packageWriteAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeDomainOrInternal(w, err, "Hapus paket CBT tidak valid")
		return
	}
	h.auditPackageEvent(r.Context(), "CBT_PACKAGE_DELETE", "cbt_package", pgUUIDString(id), nil)
	api.NoContent(w)
}

func (h *CbtPackage) Archive(w http.ResponseWriter, r *http.Request) {
	if !packageWriteAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID paket tidak valid")
		return
	}
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
		return
	}
	userID, err := authUserID(claims)
	if err != nil {
		api.Unauthorized(w)
		return
	}
	var body struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil && err != io.EOF {
		api.BadRequest(w, "Data yang dikirim tidak valid")
		return
	}
	if err := h.svc.Archive(r.Context(), id, userID, body.Reason); err != nil {
		writeDomainOrInternal(w, err, "Arsipkan paket CBT tidak valid")
		return
	}
	h.auditPackageEvent(r.Context(), "CBT_PACKAGE_ARCHIVE", "cbt_package", pgUUIDString(id), map[string]any{
		"reason": body.Reason,
	})
	api.OK(w, map[string]any{"archived": true})
}

func packageWriteAllowed(r *http.Request) bool {
	return hasAnyRole(r, "admin") || hasAnyPermission(r, "asesmen.package_manage", "asesmen.manage")
}

func packageQuestionInput(rawIDs []string, rawWeights map[string]int32, rawQuestions []struct {
	QuestionID string `json:"question_id"`
	Points     int32  `json:"points"`
}) ([]pgtype.UUID, map[string]int32, bool) {
	ids := rawIDs
	weights := rawWeights
	if len(rawQuestions) > 0 {
		ids = make([]string, 0, len(rawQuestions))
		weights = map[string]int32{}
		for _, item := range rawQuestions {
			ids = append(ids, item.QuestionID)
			weights[item.QuestionID] = item.Points
		}
	}
	questionIDs := make([]pgtype.UUID, 0, len(ids))
	for _, rawID := range ids {
		id, err := parseUUID(rawID)
		if err != nil {
			return nil, nil, false
		}
		questionIDs = append(questionIDs, id)
	}
	if weights == nil {
		weights = map[string]int32{}
	}
	return questionIDs, weights, true
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
