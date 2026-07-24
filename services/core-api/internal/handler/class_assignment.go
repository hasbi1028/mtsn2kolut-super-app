package handler

import (
	"encoding/json"
	"net/http"

	"mtsn2kolut-super-app/backend/internal/api"
	"mtsn2kolut-super-app/backend/internal/service"
)

type SubjectAssignmentHandler struct {
	svc *service.SubjectAssignmentService
}

func NewSubjectAssignmentHandler(svc *service.SubjectAssignmentService) *SubjectAssignmentHandler {
	return &SubjectAssignmentHandler{svc: svc}
}

// GET /api/academic/subject-assignments?academic_year_id=
func (h *SubjectAssignmentHandler) GetMatrix(w http.ResponseWriter, r *http.Request) {
	ayID := r.URL.Query().Get("academic_year_id")
	if ayID == "" {
		// fallback to active academic year
		activeAY, err := h.svc.GetActiveAcademicYear(r.Context())
		if err != nil {
			// still try with empty UUID — will return empty classes
			ayID = "00000000-0000-0000-0000-000000000000"
		} else {
			ayID = activeAY
		}
	}
	data, err := h.svc.GetMatrixData(r.Context(), ayID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, data)
}

// POST /api/academic/subject-assignments
func (h *SubjectAssignmentHandler) UpsertCell(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ClassID           string `json:"class_id"`
		SubjectID         string `json:"subject_id"`
		TeacherEmployeeID string `json:"teacher_employee_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "format data tidak valid")
		return
	}
	if body.ClassID == "" || body.SubjectID == "" || body.TeacherEmployeeID == "" {
		api.BadRequest(w, "class_id, subject_id, dan teacher_employee_id wajib diisi")
		return
	}
	if err := h.svc.UpsertCell(r.Context(), service.UpsertCellRequest{
		ClassID:           body.ClassID,
		SubjectID:         body.SubjectID,
		TeacherEmployeeID: body.TeacherEmployeeID,
	}); err != nil {
		writeDomainOrInternal(w, err, "gagal menyimpan assignment")
		return
	}
	api.OK(w, map[string]string{"status": "ok"})
}

// DELETE /api/academic/subject-assignments/{id}
func (h *SubjectAssignmentHandler) DeleteCell(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.BadRequest(w, "id wajib diisi")
		return
	}
	if err := h.svc.DeleteCell(r.Context(), id); err != nil {
		writeDomainOrInternal(w, err, "gagal menghapus assignment")
		return
	}
	api.NoContent(w)
}
