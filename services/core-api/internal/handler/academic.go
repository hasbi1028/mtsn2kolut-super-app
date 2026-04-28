package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type Academic struct {
	svc *service.Academic
}

func NewAcademic(svc *service.Academic) *Academic { return &Academic{svc: svc} }

func (h *Academic) Overview(w http.ResponseWriter, r *http.Request) {
	years, err := h.svc.ListYears(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	classes, err := h.svc.ListClasses(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	subjects, err := h.svc.ListSubjects(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	assigns, err := h.svc.ListAssignments(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{
		"years":       years,
		"classes":     classes,
		"subjects":    subjects,
		"assignments": assigns,
	})
	}

	func (h *Academic) GetStats(w http.ResponseWriter, r *http.Request) {
	row, err := h.svc.GetStats(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
	}

	func (h *Academic) Create(w http.ResponseWriter, r *http.Request) {
	entity := chi.URLParam(r, "entity")
	switch entity {
	case "years":
		var body struct {
			Name      string `json:"name"`
			StartDate string `json:"start_date"`
			EndDate   string `json:"end_date"`
			IsActive  bool   `json:"is_active"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			api.BadRequest(w, "invalid json")
			return
		}
		var startDate, endDate pgtype.Date
		if err := startDate.Scan(body.StartDate); err != nil {
			api.BadRequest(w, "start_date invalid")
			return
		}
		if err := endDate.Scan(body.EndDate); err != nil {
			api.BadRequest(w, "end_date invalid")
			return
		}
		row, err := h.svc.CreateYear(r.Context(), db.CreateAcademicYearParams{
			Name:      body.Name,
			StartDate: startDate,
			EndDate:   endDate,
			IsActive:  body.IsActive,
		})
		if err != nil {
			api.Internal(w, err)
			return
		}
		api.Created(w, row)
	case "classes":
		var body struct {
			AcademicYearID string `json:"academic_year_id"`
			Code           string `json:"code"`
			Name           string `json:"name"`
			Level          string `json:"level"`
			IsActive       bool   `json:"is_active"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			api.BadRequest(w, "invalid json")
			return
		}
		yearID, err := parseUUID(body.AcademicYearID)
		if err != nil {
			api.BadRequest(w, "academic_year_id invalid")
			return
		}
		row, err := h.svc.CreateClass(r.Context(), db.CreateSchoolClassParams{
			AcademicYearID: yearID,
			Code:           body.Code,
			Name:           body.Name,
			Level:          body.Level,
			IsActive:       body.IsActive,
		})
		if err != nil {
			api.Internal(w, err)
			return
		}
		api.Created(w, row)
	case "subjects":
		var body struct {
			Code     string `json:"code"`
			Name     string `json:"name"`
			IsActive bool   `json:"is_active"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			api.BadRequest(w, "invalid json")
			return
		}
		row, err := h.svc.CreateSubject(r.Context(), db.CreateSubjectParams{
			Code:     body.Code,
			Name:     body.Name,
			IsActive: body.IsActive,
		})
		if err != nil {
			api.Internal(w, err)
			return
		}
		api.Created(w, row)
	case "assignments":
		var body struct {
			ClassID           string `json:"class_id"`
			SubjectID         string `json:"subject_id"`
			TeacherEmployeeID string `json:"teacher_employee_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			api.BadRequest(w, "invalid json")
			return
		}
		classID, err := parseUUID(body.ClassID)
		if err != nil {
			api.BadRequest(w, "class_id invalid")
			return
		}
		subjectID, err := parseUUID(body.SubjectID)
		if err != nil {
			api.BadRequest(w, "subject_id invalid")
			return
		}
		teacherID, err := parseUUID(body.TeacherEmployeeID)
		if err != nil {
			api.BadRequest(w, "teacher_employee_id invalid")
			return
		}
		row, err := h.svc.CreateAssignment(r.Context(), db.CreateClassSubjectAssignmentParams{
			ClassID:           classID,
			SubjectID:         subjectID,
			TeacherEmployeeID: teacherID,
		})
		if err != nil {
			api.Internal(w, err)
			return
		}
		api.Created(w, row)
	default:
		api.NotFound(w)
	}
}

func (h *Academic) Delete(w http.ResponseWriter, r *http.Request) {
	entity := chi.URLParam(r, "entity")
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	switch entity {
	case "years":
		err = h.svc.DeleteYear(r.Context(), id)
	case "classes":
		err = h.svc.DeleteClass(r.Context(), id)
	case "subjects":
		err = h.svc.DeleteSubject(r.Context(), id)
	case "assignments":
		err = h.svc.DeleteAssignment(r.Context(), id)
	default:
		api.NotFound(w)
		return
	}
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.NoContent(w)
}
