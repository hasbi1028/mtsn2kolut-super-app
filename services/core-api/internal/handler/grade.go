package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	mw "mtsn2kolut-super-app/backend/internal/middleware"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type Grade struct {
	svc   gradeService
	audit cbtAuthoringAuditWriter
}

type gradeService interface {
	Overview(ctx context.Context, assignmentID, componentID pgtype.UUID, publishedOnly bool, teacherEmployeeID pgtype.UUID) (service.GradeOverview, error)
	CreateComponent(ctx context.Context, arg db.CreateGradeComponentParams, teacherEmployeeID pgtype.UUID) (db.GradeComponent, error)
	UpdateComponent(ctx context.Context, arg db.UpdateGradeComponentParams, teacherEmployeeID pgtype.UUID) (db.GradeComponent, error)
	SetComponentPublished(ctx context.Context, id pgtype.UUID, isPublished bool, teacherEmployeeID pgtype.UUID) (db.GradeComponent, error)
	FinalizeAssignment(ctx context.Context, assignmentID pgtype.UUID, finalizedBy, notes string, teacherEmployeeID pgtype.UUID) (db.GradeAssignmentFinalization, error)
	ReopenAssignment(ctx context.Context, assignmentID, teacherEmployeeID pgtype.UUID) error
	DeleteComponent(ctx context.Context, id, teacherEmployeeID pgtype.UUID) error
	UpsertEntry(ctx context.Context, componentID, studentID, teacherEmployeeID pgtype.UUID, score float64, notes, gradedBy string) (db.GradeEntry, error)
}

func NewGrade(svc *service.Grade, audit ...cbtAuthoringAuditWriter) *Grade {
	var writer cbtAuthoringAuditWriter
	if len(audit) > 0 {
		writer = audit[0]
	}
	return &Grade{svc: svc, audit: writer}
}

func (h *Grade) Overview(w http.ResponseWriter, r *http.Request) {
	if !gradeAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	assignmentID, err := optionalUUID(r.URL.Query().Get("assignment_id"))
	if err != nil {
		api.BadRequest(w, "assignment_id invalid")
		return
	}
	componentID, err := optionalUUID(r.URL.Query().Get("component_id"))
	if err != nil {
		api.BadRequest(w, "component_id invalid")
		return
	}
	publishedOnly := parseGradePublishedOnly(r.URL.Query().Get("published_only"))
	data, err := h.svc.Overview(r.Context(), assignmentID, componentID, publishedOnly, gradeTeacherEmployeeID(r))
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

func (h *Grade) CreateComponent(w http.ResponseWriter, r *http.Request) {
	if !gradeAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body struct {
		AssignmentID string  `json:"assignment_id"`
		Title        string  `json:"title"`
		Category     string  `json:"category"`
		Weight       float64 `json:"weight"`
		MaxScore     float64 `json:"max_score"`
		IsPublished  bool    `json:"is_published"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	assignmentID, err := parseUUID(body.AssignmentID)
	if err != nil {
		api.BadRequest(w, "assignment_id invalid")
		return
	}
	row, err := h.svc.CreateComponent(r.Context(), db.CreateGradeComponentParams{
		AssignmentID: assignmentID,
		Title:        body.Title,
		Category:     body.Category,
		Weight:       body.Weight,
		MaxScore:     body.MaxScore,
		IsPublished:  body.IsPublished,
	}, gradeTeacherEmployeeID(r))
	if err != nil {
		writeClientError(w, err, "Komponen nilai tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "GRADE_COMPONENT_CREATE", "grade_component", pgUUIDString(row.ID), map[string]any{
		"assignment_id": pgUUIDString(row.AssignmentID),
		"title":         row.Title,
		"category":      row.Category,
	})
	api.Created(w, row)
}

func (h *Grade) UpdateComponent(w http.ResponseWriter, r *http.Request) {
	if !gradeAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body struct {
		Title    string  `json:"title"`
		Category string  `json:"category"`
		Weight   float64 `json:"weight"`
		MaxScore float64 `json:"max_score"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	row, err := h.svc.UpdateComponent(r.Context(), db.UpdateGradeComponentParams{
		ID:       id,
		Title:    body.Title,
		Category: body.Category,
		Weight:   body.Weight,
		MaxScore: body.MaxScore,
	}, gradeTeacherEmployeeID(r))
	if err != nil {
		writeClientError(w, err, "Perubahan komponen nilai tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "GRADE_COMPONENT_UPDATE", "grade_component", pgUUIDString(row.ID), map[string]any{
		"assignment_id": pgUUIDString(row.AssignmentID),
		"title":         row.Title,
		"category":      row.Category,
	})
	api.OK(w, row)
}

func (h *Grade) SetComponentPublished(w http.ResponseWriter, r *http.Request) {
	if !gradeAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body struct {
		IsPublished bool `json:"is_published"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	row, err := h.svc.SetComponentPublished(r.Context(), id, body.IsPublished, gradeTeacherEmployeeID(r))
	if err != nil {
		api.Internal(w, err)
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "GRADE_COMPONENT_PUBLISH_TOGGLE", "grade_component", pgUUIDString(row.ID), map[string]any{
		"assignment_id": pgUUIDString(row.AssignmentID),
		"is_published":  row.IsPublished,
	})
	api.OK(w, row)
}

func (h *Grade) FinalizeAssignment(w http.ResponseWriter, r *http.Request) {
	if !gradeAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	assignmentID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body struct {
		Notes string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	row, err := h.svc.FinalizeAssignment(r.Context(), assignmentID, currentGradeUsername(r), body.Notes, gradeTeacherEmployeeID(r))
	if err != nil {
		writeClientError(w, err, "Finalisasi penilaian tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "GRADE_ASSIGNMENT_FINALIZE", "grade_assignment", pgUUIDString(row.AssignmentID), map[string]any{
		"finalized_by": row.FinalizedBy,
		"notes":        row.Notes,
	})
	api.OK(w, row)
}

func (h *Grade) ReopenAssignment(w http.ResponseWriter, r *http.Request) {
	if !gradeAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	assignmentID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	if err := h.svc.ReopenAssignment(r.Context(), assignmentID, gradeTeacherEmployeeID(r)); err != nil {
		api.Internal(w, err)
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "GRADE_ASSIGNMENT_REOPEN", "grade_assignment", pgUUIDString(assignmentID), map[string]any{
		"reopened_by": currentGradeUsername(r),
	})
	api.NoContent(w)
}

func (h *Grade) DeleteComponent(w http.ResponseWriter, r *http.Request) {
	if !gradeAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	if err := h.svc.DeleteComponent(r.Context(), id, gradeTeacherEmployeeID(r)); err != nil {
		api.Internal(w, err)
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "GRADE_COMPONENT_DELETE", "grade_component", pgUUIDString(id), map[string]any{
		"deleted_by": currentGradeUsername(r),
	})
	api.NoContent(w)
}

func (h *Grade) UpsertEntry(w http.ResponseWriter, r *http.Request) {
	if !gradeAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	componentID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "component id invalid")
		return
	}
	var body struct {
		StudentID string  `json:"student_id"`
		Score     float64 `json:"score"`
		Notes     string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	studentID, err := parseUUID(body.StudentID)
	if err != nil {
		api.BadRequest(w, "student_id invalid")
		return
	}
	gradedBy := currentGradeUsername(r)
	row, err := h.svc.UpsertEntry(r.Context(), componentID, studentID, gradeTeacherEmployeeID(r), body.Score, body.Notes, gradedBy)
	if err != nil {
		writeClientError(w, err, "Entri nilai tidak valid")
		return
	}
	cbtAuditAuthoringEvent(h.audit, r.Context(), "GRADE_ENTRY_UPSERT", "grade_entry", pgUUIDString(row.ID), map[string]any{
		"component_id": pgUUIDString(row.ComponentID),
		"student_id":   pgUUIDString(row.StudentID),
		"graded_by":    row.GradedBy,
	})
	api.OK(w, row)
}

func gradeAccessAllowed(r *http.Request) bool {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		return false
	}
	return mw.HasAnyRole(claims, "admin", "guru")
}

func currentGradeUsername(r *http.Request) string {
	if claims, ok := api.ClaimsFromContext(r.Context()); ok {
		if usr, _ := claims["usr"].(string); usr != "" {
			return usr
		}
		if sub, _ := claims["sub"].(string); sub != "" {
			return sub
		}
	}
	return ""
}

func gradeTeacherEmployeeID(r *http.Request) pgtype.UUID {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok || gradeHasRole(claims, "admin") || !gradeHasRole(claims, "guru") {
		return pgtype.UUID{}
	}
	if eid, _ := claims["eid"].(string); strings.TrimSpace(eid) != "" {
		var id pgtype.UUID
		if err := id.Scan(strings.TrimSpace(eid)); err == nil {
			return id
		}
	}
	return pgtype.UUID{}
}

func gradeHasRole(claims jwt.MapClaims, expected string) bool {
	return mw.HasAnyRole(claims, expected)
}

func optionalUUID(raw string) (pgtype.UUID, error) {
	if strings.TrimSpace(raw) == "" {
		return pgtype.UUID{}, nil
	}
	return parseUUID(raw)
}

func parseGradePublishedOnly(raw string) bool {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "1", "true", "yes", "published":
		return true
	default:
		return false
	}
}
