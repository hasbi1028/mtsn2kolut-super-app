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
	UpdateIdentity(ctx context.Context, arg db.UpdateRombelIdentityParams) (db.UpdateRombelIdentityRow, error)
	ListStudentsWithParents(ctx context.Context, classID pgtype.UUID) ([]db.ListStudentsByClassWithParentsRow, error)
	ListSubjectAssignments(ctx context.Context, classID pgtype.UUID) ([]db.ListRombelSubjectAssignmentsRow, error)
	GetSubjectAssignment(ctx context.Context, arg db.GetRombelSubjectAssignmentParams) (db.GetRombelSubjectAssignmentRow, error)
	CreateSubjectAssignment(ctx context.Context, arg db.CreateRombelSubjectAssignmentParams) (db.CreateRombelSubjectAssignmentRow, error)
	UpdateSubjectAssignment(ctx context.Context, arg db.UpdateRombelSubjectAssignmentParams) (db.UpdateRombelSubjectAssignmentRow, error)
	DeleteSubjectAssignment(ctx context.Context, arg db.DeleteRombelSubjectAssignmentParams) error
	ListTimetableSlots(ctx context.Context, classID pgtype.UUID) ([]db.ListRombelTimetableSlotsRow, error)
	GetTimetableSlot(ctx context.Context, arg db.GetRombelTimetableSlotParams) (db.GetRombelTimetableSlotRow, error)
	CreateTimetableSlot(ctx context.Context, arg db.CreateRombelTimetableSlotParams) (db.CreateRombelTimetableSlotRow, error)
	UpdateTimetableSlot(ctx context.Context, arg db.UpdateRombelTimetableSlotParams) (db.UpdateRombelTimetableSlotRow, error)
	DeleteTimetableSlot(ctx context.Context, arg db.DeleteRombelTimetableSlotParams) error
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

func (h *Rombel) UpdateIdentity(w http.ResponseWriter, r *http.Request) {
	if !rombelManageAllowed(w, r) {
		return
	}
	classID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body rombelIdentityRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	if body.IsActive == nil {
		api.BadRequest(w, "is_active wajib diisi")
		return
	}
	row, err := h.svc.UpdateIdentity(r.Context(), db.UpdateRombelIdentityParams{
		ID:       classID,
		Code:     body.Code,
		Name:     body.Name,
		Level:    body.Level,
		IsActive: *body.IsActive,
	})
	if err != nil {
		writeClientError(w, err, "Identitas rombel tidak valid")
		return
	}
	api.OK(w, row)
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

func (h *Rombel) ListTimetableSlots(w http.ResponseWriter, r *http.Request) {
	if !rombelReadAllowed(w, r) {
		return
	}
	classID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	rows, err := h.svc.ListTimetableSlots(r.Context(), classID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *Rombel) GetTimetableSlot(w http.ResponseWriter, r *http.Request) {
	if !rombelReadAllowed(w, r) {
		return
	}
	classID, slotID, ok := parseRombelTimetableSlotRoute(w, r)
	if !ok {
		return
	}
	row, err := h.svc.GetTimetableSlot(r.Context(), db.GetRombelTimetableSlotParams{
		ClassID: classID,
		ID:      slotID,
	})
	if err != nil {
		writeClientError(w, err, "Slot jadwal rombel tidak ditemukan")
		return
	}
	api.OK(w, row)
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

func (h *Rombel) CreateTimetableSlot(w http.ResponseWriter, r *http.Request) {
	if !rombelManageAllowed(w, r) {
		return
	}
	classID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return
	}
	var body timetableSlotRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	arg, ok := parseCreateTimetableSlot(w, classID, body)
	if !ok {
		return
	}
	row, err := h.svc.CreateTimetableSlot(r.Context(), arg)
	if err != nil {
		writeClientError(w, err, "Jadwal rombel tidak valid")
		return
	}
	api.Created(w, row)
}

func (h *Rombel) UpdateTimetableSlot(w http.ResponseWriter, r *http.Request) {
	if !rombelManageAllowed(w, r) {
		return
	}
	classID, slotID, ok := parseRombelTimetableSlotRoute(w, r)
	if !ok {
		return
	}
	var body timetableSlotRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.BadRequest(w, "invalid json")
		return
	}
	arg, ok := parseUpdateTimetableSlot(w, classID, slotID, body)
	if !ok {
		return
	}
	row, err := h.svc.UpdateTimetableSlot(r.Context(), arg)
	if err != nil {
		writeClientError(w, err, "Jadwal rombel tidak valid")
		return
	}
	api.OK(w, row)
}

func (h *Rombel) DeleteTimetableSlot(w http.ResponseWriter, r *http.Request) {
	if !rombelManageAllowed(w, r) {
		return
	}
	classID, slotID, ok := parseRombelTimetableSlotRoute(w, r)
	if !ok {
		return
	}
	if err := h.svc.DeleteTimetableSlot(r.Context(), db.DeleteRombelTimetableSlotParams{
		ClassID: classID,
		ID:      slotID,
	}); err != nil {
		writeClientError(w, err, "Penghapusan jadwal rombel tidak valid")
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

type timetableSlotRequest struct {
	AssignmentID string `json:"assignment_id"`
	DayOfWeek    int16  `json:"day_of_week"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	Room         string `json:"room"`
	RoomLabel    string `json:"room_label"`
	Notes        string `json:"notes"`
}

type rombelIdentityRequest struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Level    string `json:"level"`
	IsActive *bool  `json:"is_active"`
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

func parseCreateTimetableSlot(w http.ResponseWriter, classID pgtype.UUID, body timetableSlotRequest) (db.CreateRombelTimetableSlotParams, bool) {
	assignmentID, dayOfWeek, startTime, endTime, roomLabel, notes, ok := parseTimetableSlotBody(w, body)
	if !ok {
		return db.CreateRombelTimetableSlotParams{}, false
	}
	return db.CreateRombelTimetableSlotParams{
		ClassID:      classID,
		AssignmentID: assignmentID,
		DayOfWeek:    dayOfWeek,
		StartTime:    startTime,
		EndTime:      endTime,
		RoomLabel:    roomLabel,
		Notes:        notes,
	}, true
}

func parseUpdateTimetableSlot(w http.ResponseWriter, classID pgtype.UUID, slotID pgtype.UUID, body timetableSlotRequest) (db.UpdateRombelTimetableSlotParams, bool) {
	assignmentID, dayOfWeek, startTime, endTime, roomLabel, notes, ok := parseTimetableSlotBody(w, body)
	if !ok {
		return db.UpdateRombelTimetableSlotParams{}, false
	}
	return db.UpdateRombelTimetableSlotParams{
		ClassID:      classID,
		ID:           slotID,
		AssignmentID: assignmentID,
		DayOfWeek:    dayOfWeek,
		StartTime:    startTime,
		EndTime:      endTime,
		RoomLabel:    roomLabel,
		Notes:        notes,
	}, true
}

func parseTimetableSlotBody(w http.ResponseWriter, body timetableSlotRequest) (pgtype.UUID, int16, pgtype.Time, pgtype.Time, string, string, bool) {
	assignmentID, err := parseUUID(body.AssignmentID)
	if err != nil {
		api.BadRequest(w, "assignment_id invalid")
		return pgtype.UUID{}, 0, pgtype.Time{}, pgtype.Time{}, "", "", false
	}
	if body.DayOfWeek < 1 || body.DayOfWeek > 6 {
		api.BadRequest(w, "day_of_week harus 1-6")
		return pgtype.UUID{}, 0, pgtype.Time{}, pgtype.Time{}, "", "", false
	}
	startTime, err := service.ParseAcademicTimeInput(body.StartTime)
	if err != nil {
		api.BadRequest(w, "start_time invalid")
		return pgtype.UUID{}, 0, pgtype.Time{}, pgtype.Time{}, "", "", false
	}
	endTime, err := service.ParseAcademicTimeInput(body.EndTime)
	if err != nil {
		api.BadRequest(w, "end_time invalid")
		return pgtype.UUID{}, 0, pgtype.Time{}, pgtype.Time{}, "", "", false
	}
	if startTime.Microseconds >= endTime.Microseconds {
		api.BadRequest(w, "rentang waktu tidak valid")
		return pgtype.UUID{}, 0, pgtype.Time{}, pgtype.Time{}, "", "", false
	}
	return assignmentID, body.DayOfWeek, startTime, endTime, timetableRoomLabel(body), strings.TrimSpace(body.Notes), true
}

func timetableRoomLabel(body timetableSlotRequest) string {
	if room := strings.TrimSpace(body.Room); room != "" {
		return room
	}
	return strings.TrimSpace(body.RoomLabel)
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

func parseRombelTimetableSlotRoute(w http.ResponseWriter, r *http.Request) (pgtype.UUID, pgtype.UUID, bool) {
	classID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "invalid id")
		return pgtype.UUID{}, pgtype.UUID{}, false
	}
	slotID, err := parseUUID(chi.URLParam(r, "slotID"))
	if err != nil {
		api.BadRequest(w, "invalid slot_id")
		return pgtype.UUID{}, pgtype.UUID{}, false
	}
	return classID, slotID, true
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
