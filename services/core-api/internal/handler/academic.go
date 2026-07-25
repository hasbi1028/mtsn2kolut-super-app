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

// ─── SchoolClass (Rombel) ────────────────────────────────────────

// GET /api/academic/rombels/{id}
func (h *AcademicHandler) GetSchoolClass(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.BadRequest(w, "id rombel wajib diisi")
		return
	}
	class, err := h.semesterSvc.GetSchoolClass(r.Context(), id)
	if err != nil {
		writeDomainOrInternal(w, err, "gagal memuat rombel")
		return
	}
	api.OK(w, class)
}

// GET /api/academic/rombels
func (h *AcademicHandler) ListSchoolClasses(w http.ResponseWriter, r *http.Request) {
	classes, err := h.semesterSvc.ListSchoolClasses(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, classes)
}

// POST /api/academic/rombels
func (h *AcademicHandler) CreateSchoolClass(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AcademicYearID string `json:"academic_year_id"`
		Code           string `json:"code"`
		Name           string `json:"name"`
		Level          string `json:"level"`
		IsActive       bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "format data tidak valid")
		return
	}
	if body.Code == "" || body.Name == "" || body.Level == "" {
		api.BadRequest(w, "code, name, dan level wajib diisi")
		return
	}
	class, err := h.semesterSvc.CreateSchoolClass(r.Context(), service.SchoolClassCreateParams{
		AcademicYearID: body.AcademicYearID,
		Code:           body.Code,
		Name:           body.Name,
		Level:          body.Level,
		IsActive:       body.IsActive,
	})
	if err != nil {
		writeDomainOrInternal(w, err, "gagal membuat rombel")
		return
	}
	api.Created(w, class)
}

// DELETE /api/academic/rombels/{id}
func (h *AcademicHandler) DeleteSchoolClass(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.BadRequest(w, "id rombel wajib diisi")
		return
	}
	if err := h.semesterSvc.DeleteSchoolClass(r.Context(), id); err != nil {
		writeDomainOrInternal(w, err, "gagal menghapus rombel")
		return
	}
	api.NoContent(w)
}

// ─── Rombel Student Management ────────────────────────────────────

// GET /api/academic/rombels/{id}/students
func (h *AcademicHandler) ListRombelStudents(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.BadRequest(w, "id rombel wajib diisi")
		return
	}
	students, err := h.semesterSvc.ListStudentsByClass(r.Context(), id)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, students)
}

// GET /api/academic/rombels/unassigned-students
func (h *AcademicHandler) ListUnassignedStudents(w http.ResponseWriter, r *http.Request) {
	students, err := h.semesterSvc.ListUnassignedStudents(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, students)
}

// POST /api/academic/rombels/{id}/students
func (h *AcademicHandler) AssignStudent(w http.ResponseWriter, r *http.Request) {
	classID := r.PathValue("id")
	if classID == "" {
		api.BadRequest(w, "id rombel wajib diisi")
		return
	}
	var body struct {
		StudentID   string   `json:"student_id"`
		StudentIDs  []string `json:"student_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "format data tidak valid")
		return
	}
	if len(body.StudentIDs) > 0 {
		if err := h.semesterSvc.BulkAssignStudentsToClass(r.Context(), body.StudentIDs, classID); err != nil {
			writeDomainOrInternal(w, err, "gagal menambahkan siswa")
			return
		}
	} else if body.StudentID != "" {
		if err := h.semesterSvc.AssignStudentToClass(r.Context(), body.StudentID, classID); err != nil {
			writeDomainOrInternal(w, err, "gagal menambahkan siswa")
			return
		}
	} else {
		api.BadRequest(w, "student_id atau student_ids wajib diisi")
		return
	}
	api.OK(w, map[string]string{"status": "ok"})
}

// DELETE /api/academic/rombels/{id}/students/{studentId}
func (h *AcademicHandler) RemoveStudentFromClass(w http.ResponseWriter, r *http.Request) {
	studentID := r.PathValue("studentId")
	if studentID == "" {
		api.BadRequest(w, "id siswa wajib diisi")
		return
	}
	if err := h.semesterSvc.RemoveStudentFromClass(r.Context(), studentID); err != nil {
		writeDomainOrInternal(w, err, "gagal menghapus siswa dari rombel")
		return
	}
	api.NoContent(w)
}
