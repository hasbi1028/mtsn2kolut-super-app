package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type Grade struct {
	svc *service.Grade
}

func NewGrade(svc *service.Grade) *Grade { return &Grade{svc: svc} }

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
	data, err := h.svc.Overview(r.Context(), assignmentID, componentID, publishedOnly)
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
	})
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
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
	})
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
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
	row, err := h.svc.SetComponentPublished(r.Context(), id, body.IsPublished)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
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
	if err := h.svc.DeleteComponent(r.Context(), id); err != nil {
		api.Internal(w, err)
		return
	}
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
	row, err := h.svc.UpsertEntry(r.Context(), componentID, studentID, body.Score, body.Notes, gradedBy)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	api.OK(w, row)
}

func gradeAccessAllowed(r *http.Request) bool {
	if claims, ok := api.ClaimsFromContext(r.Context()); ok {
		if rawRoles, ok := claims["roles"].([]any); ok {
			for _, role := range rawRoles {
				if role == "admin" || role == "guru" {
					return true
				}
			}
		}
		if role, _ := claims["role"].(string); role == "admin" || role == "guru" {
			return true
		}
		return false
	}
	return true
}

func currentGradeUsername(r *http.Request) string {
	if claims, ok := api.ClaimsFromContext(r.Context()); ok {
		if sub, _ := claims["sub"].(string); sub != "" {
			return sub
		}
	}
	return ""
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
