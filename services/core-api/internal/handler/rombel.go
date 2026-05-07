package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	mw "mtsn2kolut-super-app/backend/internal/middleware"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type rombelService interface {
	List(ctx context.Context) ([]db.ListRombelsRow, error)
	Get(ctx context.Context, id pgtype.UUID) (db.GetRombelDetailRow, error)
	ListStudentsWithParents(ctx context.Context, classID pgtype.UUID) ([]db.ListStudentsByClassWithParentsRow, error)
	ListSubjectAssignments(ctx context.Context, classID pgtype.UUID) ([]db.ListRombelSubjectAssignmentsRow, error)
	GetSubjectAssignment(ctx context.Context, arg db.GetRombelSubjectAssignmentParams) (db.GetRombelSubjectAssignmentRow, error)
	CreateSubjectAssignment(ctx context.Context, arg db.CreateRombelSubjectAssignmentParams) (db.CreateRombelSubjectAssignmentRow, error)
	UpdateSubjectAssignment(ctx context.Context, arg db.UpdateRombelSubjectAssignmentParams) (db.UpdateRombelSubjectAssignmentRow, error)
	DeleteSubjectAssignment(ctx context.Context, arg db.DeleteRombelSubjectAssignmentParams) error
	ListTimetableSlots(ctx context.Context, classID pgtype.UUID) ([]db.ListRombelTimetableSlotsRow, error)
	ListHomeroomAssignments(ctx context.Context, classID pgtype.UUID) ([]db.ListHomeroomAssignmentsByClassRow, error)
	CreateHomeroomAssignment(ctx context.Context, arg db.CreateHomeroomAssignmentParams) (db.CreateHomeroomAssignmentRow, error)
	UpdateHomeroomAssignment(ctx context.Context, arg db.UpdateHomeroomAssignmentParams) (db.UpdateHomeroomAssignmentRow, error)
	DeleteHomeroomAssignment(ctx context.Context, id pgtype.UUID) error
}

type Rombel struct {
	svc rombelService
}

func NewRombel(svc *service.Rombel) *Rombel { return &Rombel{svc: svc} }

func (h *Rombel) List(w http.ResponseWriter, r *http.Request) {
	if !rombelReadAllowed(w, r) {
		return
	}
	rows, err := h.svc.List(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *Rombel) Get(w http.ResponseWriter, r *http.Request) {
	if !rombelReadAllowed(w, r) {
		return
	}
	classID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	detail, err := h.svc.Get(r.Context(), classID)
	if err != nil {
		if err == pgx.ErrNoRows {
			api.NotFound(w)
			return
		}
		api.Internal(w, err)
		return
	}
	students, err := h.svc.ListStudentsWithParents(r.Context(), classID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	assignments, err := h.svc.ListSubjectAssignments(r.Context(), classID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	timetable, err := h.svc.ListTimetableSlots(r.Context(), classID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	homerooms, err := h.svc.ListHomeroomAssignments(r.Context(), classID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{
		"rombel":               detail,
		"students":             groupRombelStudents(students),
		"subject_assignments":  assignments,
		"timetable_slots":      timetable,
		"homeroom_assignments": homerooms,
	})
}

func (h *Rombel) ListStudents(w http.ResponseWriter, r *http.Request) {
	if !rombelReadAllowed(w, r) {
		return
	}
	classID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	rows, err := h.svc.ListStudentsWithParents(r.Context(), classID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, groupRombelStudents(rows))
}

func (h *Rombel) ListHomeroomAssignments(w http.ResponseWriter, r *http.Request) {
	if !rombelReadAllowed(w, r) {
		return
	}
	classID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	rows, err := h.svc.ListHomeroomAssignments(r.Context(), classID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *Rombel) ListSubjectAssignments(w http.ResponseWriter, r *http.Request) {
	if !rombelReadAllowed(w, r) {
		return
	}
	classID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	rows, err := h.svc.ListSubjectAssignments(r.Context(), classID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *Rombel) GetSubjectAssignment(w http.ResponseWriter, r *http.Request) {
	if !rombelReadAllowed(w, r) {
		return
	}
	classID, assignmentID, ok := parseRombelAssignmentRoute(w, r)
	if !ok {
		return
	}
	row, err := h.svc.GetSubjectAssignment(r.Context(), db.GetRombelSubjectAssignmentParams{
		ClassID: classID,
		ID:      assignmentID,
	})
	if err != nil {
		writeClientError(w, err, "Penugasan guru mapel tidak ditemukan")
		return
	}
	api.OK(w, row)
}

func (h *Rombel) CreateSubjectAssignment(w http.ResponseWriter, r *http.Request) {
	if !rombelManageAllowed(w, r) {
		return
	}
	classID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body subjectAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	arg, ok := parseCreateSubjectAssignment(w, classID, body)
	if !ok {
		return
	}
	row, err := h.svc.CreateSubjectAssignment(r.Context(), arg)
	if err != nil {
		writeClientError(w, err, "Data guru mapel tidak valid")
		return
	}
	api.Created(w, row)
}

func (h *Rombel) UpdateSubjectAssignment(w http.ResponseWriter, r *http.Request) {
	if !rombelManageAllowed(w, r) {
		return
	}
	classID, assignmentID, ok := parseRombelAssignmentRoute(w, r)
	if !ok {
		return
	}
	var body subjectAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	arg, ok := parseUpdateSubjectAssignment(w, classID, assignmentID, body)
	if !ok {
		return
	}
	row, err := h.svc.UpdateSubjectAssignment(r.Context(), arg)
	if err != nil {
		writeClientError(w, err, "Data guru mapel tidak valid")
		return
	}
	api.OK(w, row)
}

func (h *Rombel) DeleteSubjectAssignment(w http.ResponseWriter, r *http.Request) {
	if !rombelManageAllowed(w, r) {
		return
	}
	classID, assignmentID, ok := parseRombelAssignmentRoute(w, r)
	if !ok {
		return
	}
	if err := h.svc.DeleteSubjectAssignment(r.Context(), db.DeleteRombelSubjectAssignmentParams{
		ClassID: classID,
		ID:      assignmentID,
	}); err != nil {
		writeClientError(w, err, "Penghapusan guru mapel tidak valid")
		return
	}
	api.NoContent(w)
}

func (h *Rombel) CreateHomeroomAssignment(w http.ResponseWriter, r *http.Request) {
	if !rombelManageAllowed(w, r) {
		return
	}
	classID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body homeroomAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	arg, ok := parseCreateHomeroomAssignment(w, classID, body)
	if !ok {
		return
	}
	row, err := h.svc.CreateHomeroomAssignment(r.Context(), arg)
	if err != nil {
		writeClientError(w, err, "Data wali kelas tidak valid")
		return
	}
	api.Created(w, row)
}

func (h *Rombel) UpdateHomeroomAssignment(w http.ResponseWriter, r *http.Request) {
	if !rombelManageAllowed(w, r) {
		return
	}
	assignmentID, err := parseUUID(chi.URLParam(r, "assignmentID"))
	if err != nil {
		api.BadRequest(w, "invalid assignment_id")
		return
	}
	var body homeroomAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	arg, ok := parseUpdateHomeroomAssignment(w, assignmentID, body)
	if !ok {
		return
	}
	row, err := h.svc.UpdateHomeroomAssignment(r.Context(), arg)
	if err != nil {
		writeClientError(w, err, "Data wali kelas tidak valid")
		return
	}
	api.OK(w, row)
}

func (h *Rombel) DeleteHomeroomAssignment(w http.ResponseWriter, r *http.Request) {
	if !rombelManageAllowed(w, r) {
		return
	}
	assignmentID, err := parseUUID(chi.URLParam(r, "assignmentID"))
	if err != nil {
		api.BadRequest(w, "invalid assignment_id")
		return
	}
	if err := h.svc.DeleteHomeroomAssignment(r.Context(), assignmentID); err != nil {
		writeClientError(w, err, "Penghapusan wali kelas tidak valid")
		return
	}
	api.NoContent(w)
}

type homeroomAssignmentRequest struct {
	EmployeeID     string `json:"employee_id"`
	AcademicYearID string `json:"academic_year_id"`
	StartDate      string `json:"start_date"`
	EndDate        string `json:"end_date"`
	IsActive       *bool  `json:"is_active"`
	Notes          string `json:"notes"`
}

type subjectAssignmentRequest struct {
	SubjectID         string `json:"subject_id"`
	TeacherEmployeeID string `json:"teacher_employee_id"`
}

type rombelStudentParent struct {
	ID               pgtype.UUID `json:"id"`
	Nama             string      `json:"nama"`
	Phone            string      `json:"phone"`
	Address          string      `json:"address"`
	Occupation       string      `json:"occupation"`
	IncomeBand       string      `json:"income_band"`
	Nik              string      `json:"nik"`
	Relationship     string      `json:"relationship"`
	IsPrimaryContact bool        `json:"is_primary_contact"`
	Notes            string      `json:"notes"`
}

type rombelStudent struct {
	ID             pgtype.UUID           `json:"id"`
	Nis            string                `json:"nis"`
	Nisn           string                `json:"nisn"`
	Nama           string                `json:"nama"`
	Gender         db.GenderEnum         `json:"gender"`
	ParentName     string                `json:"parent_name"`
	ParentPhone    string                `json:"parent_phone"`
	StudentPhone   string                `json:"student_phone"`
	StudentAddress string                `json:"student_address"`
	IsActive       bool                  `json:"is_active"`
	Status         db.StudentStatusEnum  `json:"status"`
	Parents        []rombelStudentParent `json:"parents"`
}

func groupRombelStudents(rows []db.ListStudentsByClassWithParentsRow) []rombelStudent {
	ordered := make([]rombelStudent, 0)
	indexByID := map[string]int{}
	for _, row := range rows {
		key := fmt.Sprintf("%x", row.StudentID.Bytes)
		index, ok := indexByID[key]
		if !ok {
			ordered = append(ordered, rombelStudent{
				ID:             row.StudentID,
				Nis:            row.Nis,
				Nisn:           row.Nisn,
				Nama:           row.StudentName,
				Gender:         row.Gender,
				ParentName:     row.ParentName,
				ParentPhone:    row.ParentPhone,
				StudentPhone:   row.StudentPhone,
				StudentAddress: row.StudentAddress,
				IsActive:       row.IsActive,
				Status:         row.Status,
				Parents:        []rombelStudentParent{},
			})
			index = len(ordered) - 1
			indexByID[key] = index
		}
		if !row.ParentID.Valid {
			continue
		}
		ordered[index].Parents = append(ordered[index].Parents, rombelStudentParent{
			ID:               row.ParentID,
			Nama:             row.ParentNama,
			Phone:            row.ParentPhoneLinked,
			Address:          row.ParentAddress,
			Occupation:       row.ParentOccupation,
			IncomeBand:       row.ParentIncomeBand,
			Nik:              row.ParentNik,
			Relationship:     relationshipString(row.Relationship),
			IsPrimaryContact: row.IsPrimaryContact,
			Notes:            row.RelationshipNotes,
		})
	}
	return ordered
}

func parseCreateHomeroomAssignment(w http.ResponseWriter, classID pgtype.UUID, body homeroomAssignmentRequest) (db.CreateHomeroomAssignmentParams, bool) {
	employeeID, err := parseUUID(body.EmployeeID)
	if err != nil {
		api.BadRequest(w, "employee_id invalid")
		return db.CreateHomeroomAssignmentParams{}, false
	}
	active := true
	if body.IsActive != nil {
		active = *body.IsActive
	}
	academicYearID, ok := parseOptionalUUIDParam(w, body.AcademicYearID, "academic_year_id")
	if !ok {
		return db.CreateHomeroomAssignmentParams{}, false
	}
	startDate, ok := parseOptionalDateParam(w, body.StartDate, "start_date")
	if !ok {
		return db.CreateHomeroomAssignmentParams{}, false
	}
	endDate, ok := parseOptionalDateParam(w, body.EndDate, "end_date")
	if !ok {
		return db.CreateHomeroomAssignmentParams{}, false
	}
	return db.CreateHomeroomAssignmentParams{
		ClassID:                classID,
		EmployeeID:             employeeID,
		HomeroomAcademicYearID: optionalUUIDArg(academicYearID),
		HomeroomStartDate:      optionalDateArg(startDate),
		HomeroomEndDate:        endDate,
		HomeroomIsActive:       active,
		Notes:                  strings.TrimSpace(body.Notes),
	}, true
}

func parseUpdateHomeroomAssignment(w http.ResponseWriter, assignmentID pgtype.UUID, body homeroomAssignmentRequest) (db.UpdateHomeroomAssignmentParams, bool) {
	employeeID, err := parseUUID(body.EmployeeID)
	if err != nil {
		api.BadRequest(w, "employee_id invalid")
		return db.UpdateHomeroomAssignmentParams{}, false
	}
	active := true
	if body.IsActive != nil {
		active = *body.IsActive
	}
	return db.UpdateHomeroomAssignmentParams{
		ID:               assignmentID,
		EmployeeID:       employeeID,
		HomeroomIsActive: active,
		Notes:            strings.TrimSpace(body.Notes),
	}, true
}

func parseCreateSubjectAssignment(w http.ResponseWriter, classID pgtype.UUID, body subjectAssignmentRequest) (db.CreateRombelSubjectAssignmentParams, bool) {
	subjectID, teacherID, ok := parseSubjectAssignmentBody(w, body)
	if !ok {
		return db.CreateRombelSubjectAssignmentParams{}, false
	}
	return db.CreateRombelSubjectAssignmentParams{
		ClassID:           classID,
		SubjectID:         subjectID,
		TeacherEmployeeID: teacherID,
	}, true
}

func parseUpdateSubjectAssignment(w http.ResponseWriter, classID pgtype.UUID, assignmentID pgtype.UUID, body subjectAssignmentRequest) (db.UpdateRombelSubjectAssignmentParams, bool) {
	subjectID, teacherID, ok := parseSubjectAssignmentBody(w, body)
	if !ok {
		return db.UpdateRombelSubjectAssignmentParams{}, false
	}
	return db.UpdateRombelSubjectAssignmentParams{
		ClassID:           classID,
		ID:                assignmentID,
		SubjectID:         subjectID,
		TeacherEmployeeID: teacherID,
	}, true
}

func parseSubjectAssignmentBody(w http.ResponseWriter, body subjectAssignmentRequest) (pgtype.UUID, pgtype.UUID, bool) {
	subjectID, err := parseUUID(body.SubjectID)
	if err != nil {
		api.BadRequest(w, "subject_id invalid")
		return pgtype.UUID{}, pgtype.UUID{}, false
	}
	teacherID, err := parseUUID(body.TeacherEmployeeID)
	if err != nil {
		api.BadRequest(w, "teacher_employee_id invalid")
		return pgtype.UUID{}, pgtype.UUID{}, false
	}
	return subjectID, teacherID, true
}

func parseRombelAssignmentRoute(w http.ResponseWriter, r *http.Request) (pgtype.UUID, pgtype.UUID, bool) {
	classID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return pgtype.UUID{}, pgtype.UUID{}, false
	}
	assignmentID, err := parseUUID(chi.URLParam(r, "assignmentID"))
	if err != nil {
		api.BadRequest(w, "invalid assignment_id")
		return pgtype.UUID{}, pgtype.UUID{}, false
	}
	return classID, assignmentID, true
}

func parseOptionalUUIDParam(w http.ResponseWriter, value string, field string) (pgtype.UUID, bool) {
	if strings.TrimSpace(value) == "" {
		return pgtype.UUID{}, true
	}
	id, err := parseUUID(value)
	if err != nil {
		api.BadRequest(w, field+" invalid")
		return pgtype.UUID{}, false
	}
	return id, true
}

func parseOptionalDateParam(w http.ResponseWriter, value string, field string) (pgtype.Date, bool) {
	if strings.TrimSpace(value) == "" {
		return pgtype.Date{}, true
	}
	var parsed pgtype.Date
	if err := parsed.Scan(strings.TrimSpace(value)); err != nil {
		api.BadRequest(w, field+" invalid")
		return pgtype.Date{}, false
	}
	return parsed, true
}

func optionalUUIDArg(value pgtype.UUID) any {
	if !value.Valid {
		return nil
	}
	return value
}

func optionalDateArg(value pgtype.Date) any {
	if !value.Valid {
		return nil
	}
	return value
}

func relationshipString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return strings.TrimSpace(fmt.Sprint(value))
	}
}

func rombelReadAllowed(w http.ResponseWriter, r *http.Request) bool {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
		return false
	}
	if mw.HasAnyRole(claims, "admin", "guru", "staf", "kesiswaan") || mw.HasAnyPermission(claims, "academic.read", "students.read") {
		return true
	}
	api.Forbidden(w)
	return false
}

func rombelManageAllowed(w http.ResponseWriter, r *http.Request) bool {
	claims, ok := api.ClaimsFromContext(r.Context())
	if !ok {
		api.Unauthorized(w)
		return false
	}
	if mw.HasAnyRole(claims, "admin") || mw.HasAnyPermission(claims, "academic.manage") {
		return true
	}
	api.Forbidden(w)
	return false
}
