package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type CbtEvent struct {
	svc *service.CbtEvent
}

func NewCbtEvent(svc *service.CbtEvent) *CbtEvent { return &CbtEvent{svc: svc} }

func (h *CbtEvent) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.List(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *CbtEvent) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	row, err := h.svc.Get(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
}

func (h *CbtEvent) GetResults(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	rows, err := h.svc.GetResults(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *CbtEvent) GetExamCards(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	rows, err := h.svc.GetExamCards(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *CbtEvent) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title          string `json:"title"`
		ExamType       string `json:"exam_type"`
		Scope          string `json:"scope"`
		TargetLevels   []string `json:"target_levels"`
		AcademicYearID string `json:"academic_year_id"`
		Status         string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	if body.Title == "" {
		api.BadRequest(w, "title required")
		return
	}
	examType := db.CbtExamType(body.ExamType)
	if examType == "" {
		examType = db.CbtExamTypeLainnya
	}
	scope := body.Scope
	if scope == "" {
		scope = "class"
	}
	targetLevels := normalizeTargetLevels(body.TargetLevels)
	if err := validateTargetLevels(targetLevels); err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	var ayID pgtype.UUID
	if body.AcademicYearID != "" {
		var err error
		ayID, err = parseUUID(body.AcademicYearID)
		if err != nil {
			api.BadRequest(w, "academic_year_id invalid")
			return
		}
	}

	row, err := h.svc.Create(r.Context(), service.CreateCbtEventInput{
		Title:          body.Title,
		ExamType:       examType,
		Scope:          scope,
		TargetLevels:   targetLevels,
		AcademicYearID: ayID,
		Status:         body.Status,
	})
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.Created(w, row)
}

func (h *CbtEvent) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body struct {
		Title          string `json:"title"`
		ExamType       string `json:"exam_type"`
		Scope          string `json:"scope"`
		TargetLevels   []string `json:"target_levels"`
		AcademicYearID string `json:"academic_year_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	targetLevels := normalizeTargetLevels(body.TargetLevels)
	if err := validateTargetLevels(targetLevels); err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	var ayID pgtype.UUID
	if body.AcademicYearID != "" {
		ayID, err = parseUUID(body.AcademicYearID)
		if err != nil {
			api.BadRequest(w, "academic_year_id invalid")
			return
		}
	}
	row, err := h.svc.Update(r.Context(), id, service.CreateCbtEventInput{
		Title:          body.Title,
		ExamType:       db.CbtExamType(body.ExamType),
		Scope:          body.Scope,
		TargetLevels:   targetLevels,
		AcademicYearID: ayID,
	})
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
}

func (h *CbtEvent) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	row, err := h.svc.UpdateStatus(r.Context(), id, body.Status)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
}

func (h *CbtEvent) Delete(w http.ResponseWriter, r *http.Request) {
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

func normalizeTargetLevels(levels []string) []string {
	if len(levels) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(levels))
	seen := map[string]struct{}{}
	for _, level := range levels {
		normalized := strings.ToUpper(strings.TrimSpace(level))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	slices.Sort(out)
	return out
}

func validateTargetLevels(levels []string) error {
	allowed := map[string]struct{}{
		"VII":  {},
		"VIII": {},
		"IX":   {},
	}
	for _, level := range levels {
		if _, ok := allowed[level]; !ok {
			return errors.New("target_levels hanya boleh berisi VII, VIII, atau IX")
		}
	}
	return nil
}
