package handler

import (
	"encoding/json"
	"net/http"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type AcademicHandler struct {
	semesterSvc *service.SemesterService
}

func NewAcademicHandler(semesterSvc *service.SemesterService) *AcademicHandler {
	return &AcademicHandler{semesterSvc: semesterSvc}
}

func (h *AcademicHandler) ListSemesters(w http.ResponseWriter, r *http.Request) {
	semesters, err := h.semesterSvc.List(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, semesters)
}

func (h *AcademicHandler) GetSemester(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.BadRequest(w, "id semester wajib diisi")
		return
	}
	semester, err := h.semesterSvc.Get(r.Context(), id)
	if err != nil {
		writeDomainOrInternal(w, err, "gagal memuat semester")
		return
	}
	api.OK(w, semester)
}

func (h *AcademicHandler) GetActiveSemester(w http.ResponseWriter, r *http.Request) {
	semester, err := h.semesterSvc.GetActive(r.Context())
	if err != nil {
		writeDomainOrInternal(w, err, "gagal memuat semester aktif")
		return
	}
	api.OK(w, semester)
}

func (h *AcademicHandler) CreateSemester(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AcademicYearID string `json:"academic_year_id"`
		Name           string `json:"name"`
		Label          string `json:"label"`
		StartDate      string `json:"start_date"`
		EndDate        string `json:"end_date"`
		IsActive       bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "format data tidak valid")
		return
	}
	if body.Name == "" || body.Label == "" || body.StartDate == "" || body.EndDate == "" {
		api.BadRequest(w, "nama, label, start_date, dan end_date wajib diisi")
		return
	}

	semester, err := h.semesterSvc.Create(r.Context(), service.SemesterCreateParams{
		AcademicYearID: body.AcademicYearID,
		Name:           body.Name,
		Label:          body.Label,
		StartDate:      body.StartDate,
		EndDate:        body.EndDate,
		IsActive:       body.IsActive,
	})
	if err != nil {
		writeDomainOrInternal(w, err, "gagal membuat semester")
		return
	}
	api.Created(w, semester)
}

func (h *AcademicHandler) ActivateSemester(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.BadRequest(w, "id semester wajib diisi")
		return
	}
	semester, err := h.semesterSvc.Activate(r.Context(), id)
	if err != nil {
		writeDomainOrInternal(w, err, "gagal mengaktifkan semester")
		return
	}
	api.OK(w, semester)
}

func (h *AcademicHandler) DeleteSemester(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.BadRequest(w, "id semester wajib diisi")
		return
	}
	if err := h.semesterSvc.Delete(r.Context(), id); err != nil {
		writeDomainOrInternal(w, err, "gagal menghapus semester")
		return
	}
	api.NoContent(w)
}
