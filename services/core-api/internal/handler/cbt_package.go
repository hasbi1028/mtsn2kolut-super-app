package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type CbtPackage struct {
	svc *service.CbtPackage
}

func NewCbtPackage(svc *service.CbtPackage) *CbtPackage { return &CbtPackage{svc: svc} }

func (h *CbtPackage) List(w http.ResponseWriter, r *http.Request) {
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
		api.BadRequest(w, err.Error())
		return
	}
	api.Created(w, row)
}

func (h *CbtPackage) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		api.Internal(w, err)
		return
	}
	api.NoContent(w)
}
