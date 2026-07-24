package handler

import (
	"encoding/json"
	"net/http"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type CurriculumHandler struct {
	svc         *service.CurriculumService
	semesterSvc *service.SemesterService
}

func NewCurriculumHandler(svc *service.CurriculumService, semesterSvc *service.SemesterService) *CurriculumHandler {
	return &CurriculumHandler{svc: svc, semesterSvc: semesterSvc}
}

// ─── Profiles ────────────────────────────────────────────────────

// GET /api/academic/curriculum/profiles
func (h *CurriculumHandler) ListProfiles(w http.ResponseWriter, r *http.Request) {
	profiles, err := h.svc.ListProfiles(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, profiles)
}

// GET /api/academic/curriculum/profiles/active
func (h *CurriculumHandler) GetActiveProfile(w http.ResponseWriter, r *http.Request) {
	profile, err := h.svc.GetActiveProfile(r.Context())
	if err != nil {
		writeDomainOrInternal(w, err, "gagal memuat kurikulum aktif")
		return
	}
	api.OK(w, profile)
}

// POST /api/academic/curriculum/profiles
func (h *CurriculumHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code                string `json:"code"`
		Name                string `json:"name"`
		RegulationReference string `json:"regulation_reference"`
		EducationLevel      string `json:"education_level"`
		EffectiveYearID     string `json:"effective_academic_year_id"`
		Status              string `json:"status"`
		Notes               string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "format data tidak valid")
		return
	}
	if body.Code == "" || body.Name == "" {
		api.BadRequest(w, "code dan name wajib diisi")
		return
	}
	if body.Status == "" {
		body.Status = "draft"
	}
	if body.EducationLevel == "" {
		body.EducationLevel = "MTs"
	}

	profile, err := h.svc.CreateProfile(r.Context(), service.CurriculumProfileCreateParams{
		Code:                body.Code,
		Name:                body.Name,
		RegulationReference: body.RegulationReference,
		EducationLevel:      body.EducationLevel,
		EffectiveYearID:     body.EffectiveYearID,
		Status:              body.Status,
		Notes:               body.Notes,
	})
	if err != nil {
		writeDomainOrInternal(w, err, "gagal membuat kurikulum")
		return
	}
	api.Created(w, profile)
}

// POST /api/academic/curriculum/profiles/{id}/activate
func (h *CurriculumHandler) ActivateProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.BadRequest(w, "id kurikulum wajib diisi")
		return
	}
	profile, err := h.svc.ActivateProfile(r.Context(), id)
	if err != nil {
		writeDomainOrInternal(w, err, "gagal mengaktifkan kurikulum")
		return
	}
	api.OK(w, profile)
}

// DELETE /api/academic/curriculum/profiles/{id}
func (h *CurriculumHandler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.BadRequest(w, "id kurikulum wajib diisi")
		return
	}
	if err := h.svc.DeleteProfile(r.Context(), id); err != nil {
		writeDomainOrInternal(w, err, "gagal menghapus kurikulum")
		return
	}
	api.NoContent(w)
}

// ─── Subject Allocations ─────────────────────────────────────────

// GET /api/academic/curriculum/profiles/{id}/allocations?level=VII
func (h *CurriculumHandler) ListAllocations(w http.ResponseWriter, r *http.Request) {
	profileID := r.PathValue("id")
	level := r.URL.Query().Get("level")
	allocations, err := h.svc.ListAllocations(r.Context(), profileID, level)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, allocations)
}

// POST /api/academic/curriculum/profiles/{id}/allocations
func (h *CurriculumHandler) CreateAllocation(w http.ResponseWriter, r *http.Request) {
	profileID := r.PathValue("id")
	var body struct {
		SubjectID string  `json:"subject_id"`
		Level     string  `json:"level"`
		Group     string  `json:"subject_group"`
		Intra     float64 `json:"intra_weekly_hours"`
		Koku      float64 `json:"koku_weekly_hours"`
		Total     float64 `json:"total_weekly_hours"`
		Notes     string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "format data tidak valid")
		return
	}
	if body.SubjectID == "" || body.Level == "" {
		api.BadRequest(w, "subject_id dan level wajib diisi")
		return
	}
	if body.Group == "" {
		body.Group = "wajib"
	}
	if body.Total == 0 {
		body.Total = body.Intra + body.Koku
	}

	allocation, err := h.svc.CreateAllocation(r.Context(), profileID, body.SubjectID, body.Level, body.Group, body.Intra, body.Koku, body.Total, body.Notes)
	if err != nil {
		writeDomainOrInternal(w, err, "gagal menambah alokasi mapel")
		return
	}
	api.Created(w, allocation)
}

// DELETE /api/academic/curriculum/allocations/{id}
func (h *CurriculumHandler) DeleteAllocation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.BadRequest(w, "id alokasi wajib diisi")
		return
	}
	if err := h.svc.DeleteAllocation(r.Context(), id); err != nil {
		writeDomainOrInternal(w, err, "gagal menghapus alokasi")
		return
	}
	api.NoContent(w)
}

// ─── Class Assignments ───────────────────────────────────────────

// GET /api/academic/curriculum/assignments?class_id=
func (h *CurriculumHandler) ListAssignments(w http.ResponseWriter, r *http.Request) {
	classID := r.URL.Query().Get("class_id")
	assignments, err := h.svc.ListActiveAssignments(r.Context(), classID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, assignments)
}

// POST /api/academic/curriculum/assignments
func (h *CurriculumHandler) CreateAssignment(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ClassID   string `json:"class_id"`
		ProfileID string `json:"curriculum_profile_id"`
		Notes     string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "format data tidak valid")
		return
	}
	if body.ClassID == "" || body.ProfileID == "" {
		api.BadRequest(w, "class_id dan curriculum_profile_id wajib diisi")
		return
	}
	assignment, err := h.svc.CreateAssignment(r.Context(), body.ClassID, body.ProfileID, body.Notes)
	if err != nil {
		writeDomainOrInternal(w, err, "gagal assign kurikulum")
		return
	}
	api.Created(w, assignment)
}

// DELETE /api/academic/curriculum/assignments/{id}
func (h *CurriculumHandler) DeleteAssignment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.BadRequest(w, "id assignment wajib diisi")
		return
	}
	if err := h.svc.DeleteAssignment(r.Context(), id); err != nil {
		writeDomainOrInternal(w, err, "gagal menghapus assignment")
		return
	}
	api.NoContent(w)
}
