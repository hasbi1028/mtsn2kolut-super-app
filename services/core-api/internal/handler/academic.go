package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/api"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type Academic struct {
	svc academicService
}

type academicService interface {
	ListYears(ctx context.Context) ([]db.AcademicYear, error)
	ListClasses(ctx context.Context) ([]db.ListSchoolClassesRow, error)
	ListSubjects(ctx context.Context) ([]db.ListSubjectsRow, error)
	ListAssignments(ctx context.Context) ([]db.ListClassSubjectAssignmentsRow, error)
	ListTimetableSlots(ctx context.Context) ([]db.ListTimetableSlotsRow, error)
	GetWeeklyTimetable(ctx context.Context) (service.WeeklyTimetable, error)
	GetTimetableConflicts(ctx context.Context) ([]db.ListTimetableConflictsRow, error)
	GetLessonPeriodOverview(ctx context.Context) (service.LessonPeriodOverview, error)
	CreateLessonPeriodTemplate(ctx context.Context, p db.CreateLessonPeriodTemplateParams) (db.LessonPeriodTemplate, error)
	UpdateLessonPeriodTemplate(ctx context.Context, p db.UpdateLessonPeriodTemplateParams) (db.LessonPeriodTemplate, error)
	DeleteLessonPeriodTemplate(ctx context.Context, id pgtype.UUID) error
	GetTeacherWorkload(ctx context.Context) (service.TeacherWorkloadOverview, error)
	GetStats(ctx context.Context) (db.GetAcademicStatsRow, error)
	GetDashboardSummary(ctx context.Context) (db.GetAcademicDashboardSummaryRow, error)
	GetReadinessSummary(ctx context.Context) (db.GetAcademicReadinessSummaryRow, error)
	GetCurriculumOverview(ctx context.Context, profileID pgtype.UUID, level string) (service.CurriculumOverview, error)
	ListCurriculumProfiles(ctx context.Context) ([]db.CurriculumProfile, error)
	ListCurriculumAllocations(ctx context.Context, profileID pgtype.UUID, level string) ([]service.CurriculumAllocation, error)
	GetCurriculumSummary(ctx context.Context, profileID pgtype.UUID) ([]service.CurriculumLevelSummary, error)
	CreateYear(ctx context.Context, p db.CreateAcademicYearParams) (db.AcademicYear, error)
	ActivateYear(ctx context.Context, id pgtype.UUID, confirmation string) (db.AcademicYear, error)
	PreviewYearRollover(ctx context.Context, input service.YearRolloverPreviewInput) (service.YearRolloverPreview, error)
	ApplyYearRollover(ctx context.Context, input service.YearRolloverApplyInput) (service.YearRolloverApplyResult, error)
	DryRunAcademicImport(ctx context.Context, input service.AcademicImportDryRunInput) (service.AcademicImportDryRunResult, error)
	CreateClass(ctx context.Context, p db.CreateSchoolClassParams) (db.SchoolClass, error)
	CreateSubject(ctx context.Context, p db.CreateSubjectParams) (db.Subject, error)
	UpdateSubject(ctx context.Context, p db.UpdateSubjectParams) (db.Subject, error)
	CreateAssignment(ctx context.Context, p db.CreateClassSubjectAssignmentParams) (db.ClassSubjectAssignment, error)
	CreateTimetableSlot(ctx context.Context, p db.CreateTimetableSlotParams) (db.TimetableSlot, error)
	UpdateTimetableSlot(ctx context.Context, p db.UpdateTimetableSlotParams) (db.TimetableSlot, error)
	DeleteYear(ctx context.Context, id pgtype.UUID) error
	DeleteClass(ctx context.Context, id pgtype.UUID) error
	DeleteSubject(ctx context.Context, id pgtype.UUID) error
	DeleteAssignment(ctx context.Context, id pgtype.UUID) error
	DeleteTimetableSlot(ctx context.Context, id pgtype.UUID) error
}

func NewAcademic(svc *service.Academic) *Academic { return &Academic{svc: svc} }

func (h *Academic) Overview(w http.ResponseWriter, r *http.Request) {
	if !academicReadAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
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
	timetableSlots, err := h.svc.ListTimetableSlots(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{
		"years":          years,
		"classes":        classes,
		"subjects":       subjects,
		"assignments":    assigns,
		"timetableSlots": timetableSlots,
	})
}

func (h *Academic) ListSubjectsOnly(w http.ResponseWriter, r *http.Request) {
	if !academicSubjectListAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	rows, err := h.svc.ListSubjects(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, map[string]any{"subjects": rows})
}

func (h *Academic) GetStats(w http.ResponseWriter, r *http.Request) {
	if !academicReadAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	row, err := h.svc.GetStats(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
}

func (h *Academic) GetDashboard(w http.ResponseWriter, r *http.Request) {
	if !academicReadAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	row, err := h.svc.GetDashboardSummary(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
}

func (h *Academic) GetReadiness(w http.ResponseWriter, r *http.Request) {
	if !academicReadAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	row, err := h.svc.GetReadinessSummary(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
}

func (h *Academic) GetCurriculumOverview(w http.ResponseWriter, r *http.Request) {
	if !academicReadAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	profileID, err := parseOptionalUUID(r.URL.Query().Get("profile_id"))
	if err != nil {
		api.BadRequest(w, "Profil kurikulum tidak valid")
		return
	}
	level := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("level")))
	if level != "" && level != "VII" && level != "VIII" && level != "IX" {
		api.BadRequest(w, "Tingkat kelas tidak valid")
		return
	}
	row, err := h.svc.GetCurriculumOverview(r.Context(), profileID, level)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
}

func (h *Academic) ListCurriculumProfiles(w http.ResponseWriter, r *http.Request) {
	if !academicReadAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	rows, err := h.svc.ListCurriculumProfiles(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *Academic) ListCurriculumAllocations(w http.ResponseWriter, r *http.Request) {
	if !academicReadAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	profileID, err := parseUUID(r.URL.Query().Get("profile_id"))
	if err != nil {
		api.BadRequest(w, "Profil kurikulum tidak valid")
		return
	}
	level := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("level")))
	if level != "" && level != "VII" && level != "VIII" && level != "IX" {
		api.BadRequest(w, "Tingkat kelas tidak valid")
		return
	}
	rows, err := h.svc.ListCurriculumAllocations(r.Context(), profileID, level)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *Academic) GetCurriculumSummary(w http.ResponseWriter, r *http.Request) {
	if !academicReadAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	profileID, err := parseUUID(r.URL.Query().Get("profile_id"))
	if err != nil {
		api.BadRequest(w, "Profil kurikulum tidak valid")
		return
	}
	rows, err := h.svc.GetCurriculumSummary(r.Context(), profileID)
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *Academic) GetWeeklyTimetable(w http.ResponseWriter, r *http.Request) {
	if !academicReadAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	row, err := h.svc.GetWeeklyTimetable(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
}

func (h *Academic) GetTimetableConflicts(w http.ResponseWriter, r *http.Request) {
	if !academicReadAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	rows, err := h.svc.GetTimetableConflicts(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, rows)
}

func (h *Academic) GetLessonPeriods(w http.ResponseWriter, r *http.Request) {
	if !academicReadAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	row, err := h.svc.GetLessonPeriodOverview(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
}

func (h *Academic) CreateLessonPeriod(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body lessonPeriodRequest
	if !decodeJSON(w, r, &body, defaultJSONBodyLimit) {
		return
	}
	params, ok := parseCreateLessonPeriod(w, body)
	if !ok {
		return
	}
	row, err := h.svc.CreateLessonPeriodTemplate(r.Context(), params)
	if err != nil {
		writeClientError(w, err, "Jam pelajaran tidak valid")
		return
	}
	api.Created(w, row)
}

func (h *Academic) UpdateLessonPeriod(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	var body lessonPeriodRequest
	if !decodeJSON(w, r, &body, defaultJSONBodyLimit) {
		return
	}
	params, ok := parseUpdateLessonPeriod(w, id, body)
	if !ok {
		return
	}
	row, err := h.svc.UpdateLessonPeriodTemplate(r.Context(), params)
	if err != nil {
		writeClientError(w, err, "Jam pelajaran tidak valid")
		return
	}
	api.OK(w, row)
}

func (h *Academic) DeleteLessonPeriod(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	if err := h.svc.DeleteLessonPeriodTemplate(r.Context(), id); err != nil {
		writeClientError(w, err, "Jam pelajaran tidak dapat dihapus")
		return
	}
	api.NoContent(w)
}

func (h *Academic) GetTeacherWorkload(w http.ResponseWriter, r *http.Request) {
	if !academicReadAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	row, err := h.svc.GetTeacherWorkload(r.Context())
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.OK(w, row)
}

func (h *Academic) ActivateYear(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	var body struct {
		Confirmation string `json:"confirmation"`
	}
	if !decodeJSON(w, r, &body, defaultJSONBodyLimit) {
		return
	}
	row, err := h.svc.ActivateYear(r.Context(), id, body.Confirmation)
	if err != nil {
		writeDomainOrInternal(w, err, "Aktivasi tahun ajaran tidak valid")
		return
	}
	api.OK(w, row)
}

func (h *Academic) PreviewYearRollover(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body struct {
		SourceAcademicYearID string `json:"source_academic_year_id"`
		TargetAcademicYearID string `json:"target_academic_year_id"`
	}
	if !decodeJSON(w, r, &body, defaultJSONBodyLimit) {
		return
	}
	sourceID, err := parseOptionalUUID(body.SourceAcademicYearID)
	if err != nil {
		api.BadRequest(w, "Tahun ajaran sumber tidak valid")
		return
	}
	targetID, err := parseUUID(body.TargetAcademicYearID)
	if err != nil {
		api.BadRequest(w, "Tahun ajaran tujuan tidak valid")
		return
	}
	preview, err := h.svc.PreviewYearRollover(r.Context(), service.YearRolloverPreviewInput{
		SourceAcademicYearID: sourceID,
		TargetAcademicYearID: targetID,
	})
	if err != nil {
		writeDomainOrInternal(w, err, "Pratinjau kenaikan kelas tidak valid")
		return
	}
	api.OK(w, preview)
}

func (h *Academic) ApplyYearRollover(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body struct {
		SourceAcademicYearID string `json:"source_academic_year_id"`
		TargetAcademicYearID string `json:"target_academic_year_id"`
		Confirmation         string `json:"confirmation"`
		SafetyToken          string `json:"safety_token"`
	}
	if !decodeJSON(w, r, &body, defaultJSONBodyLimit) {
		return
	}
	sourceID, err := parseOptionalUUID(body.SourceAcademicYearID)
	if err != nil {
		api.BadRequest(w, "Tahun ajaran sumber tidak valid")
		return
	}
	targetID, err := parseUUID(body.TargetAcademicYearID)
	if err != nil {
		api.BadRequest(w, "Tahun ajaran tujuan tidak valid")
		return
	}
	result, err := h.svc.ApplyYearRollover(r.Context(), service.YearRolloverApplyInput{
		SourceAcademicYearID: sourceID,
		TargetAcademicYearID: targetID,
		Confirmation:         body.Confirmation,
		SafetyToken:          body.SafetyToken,
	})
	if err != nil {
		writeDomainOrInternal(w, err, "Penerapan kenaikan kelas tidak valid")
		return
	}
	api.OK(w, result)
}

func (h *Academic) DryRunAcademicImport(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	var body service.AcademicImportDryRunInput
	if !decodeJSON(w, r, &body, defaultJSONBodyLimit) {
		return
	}
	result, err := h.svc.DryRunAcademicImport(r.Context(), body)
	if err != nil {
		writeDomainOrInternal(w, err, "Cek data sebelum impor tidak valid")
		return
	}
	api.OK(w, result)
}

func (h *Academic) Create(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	entity := chi.URLParam(r, "entity")
	switch entity {
	case "years":
		var body struct {
			Name      string `json:"name"`
			StartDate string `json:"start_date"`
			EndDate   string `json:"end_date"`
			IsActive  bool   `json:"is_active"`
		}
		if !decodeJSON(w, r, &body, defaultJSONBodyLimit) {
			return
		}
		var startDate, endDate pgtype.Date
		if err := startDate.Scan(body.StartDate); err != nil {
			api.BadRequest(w, "Tanggal mulai tidak valid")
			return
		}
		if err := endDate.Scan(body.EndDate); err != nil {
			api.BadRequest(w, "Tanggal selesai tidak valid")
			return
		}
		row, err := h.svc.CreateYear(r.Context(), db.CreateAcademicYearParams{
			Name:      body.Name,
			StartDate: startDate,
			EndDate:   endDate,
			IsActive:  body.IsActive,
		})
		if err != nil {
			writeDomainOrInternal(w, err, "Tahun ajaran tidak valid")
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
		if !decodeJSON(w, r, &body, defaultJSONBodyLimit) {
			return
		}
		yearID, err := parseUUID(body.AcademicYearID)
		if err != nil {
			api.BadRequest(w, "Tahun ajaran tidak valid")
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
		var body subjectRequest
		if !decodeJSON(w, r, &body, defaultJSONBodyLimit) {
			return
		}
		row, err := h.svc.CreateSubject(r.Context(), createSubjectParams(body))
		if err != nil {
			writeDomainOrInternal(w, err, "Data mapel tidak valid")
			return
		}
		api.Created(w, row)
	case "assignments":
		var body struct {
			ClassID           string `json:"class_id"`
			SubjectID         string `json:"subject_id"`
			TeacherEmployeeID string `json:"teacher_employee_id"`
		}
		if !decodeJSON(w, r, &body, defaultJSONBodyLimit) {
			return
		}
		classID, err := parseUUID(body.ClassID)
		if err != nil {
			api.BadRequest(w, "Rombel tidak valid")
			return
		}
		subjectID, err := parseUUID(body.SubjectID)
		if err != nil {
			api.BadRequest(w, "Mata pelajaran tidak valid")
			return
		}
		teacherID, err := parseUUID(body.TeacherEmployeeID)
		if err != nil {
			api.BadRequest(w, "Guru tidak valid")
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
	case "timetables":
		var body struct {
			AssignmentID string `json:"assignment_id"`
			DayOfWeek    int16  `json:"day_of_week"`
			StartTime    string `json:"start_time"`
			EndTime      string `json:"end_time"`
			RoomLabel    string `json:"room_label"`
			Notes        string `json:"notes"`
		}
		if !decodeJSON(w, r, &body, defaultJSONBodyLimit) {
			return
		}
		assignmentID, err := parseUUID(body.AssignmentID)
		if err != nil {
			api.BadRequest(w, "Penugasan guru mapel tidak valid")
			return
		}
		if body.DayOfWeek < 1 || body.DayOfWeek > 6 {
			api.BadRequest(w, "Hari jadwal harus Senin sampai Sabtu")
			return
		}
		startTime, err := service.ParseAcademicTimeInput(body.StartTime)
		if err != nil {
			api.BadRequest(w, "Jam mulai tidak valid")
			return
		}
		endTime, err := service.ParseAcademicTimeInput(body.EndTime)
		if err != nil {
			api.BadRequest(w, "Jam selesai tidak valid")
			return
		}
		if startTime.Microseconds >= endTime.Microseconds {
			api.BadRequest(w, "rentang waktu tidak valid")
			return
		}
		row, err := h.svc.CreateTimetableSlot(r.Context(), db.CreateTimetableSlotParams{
			AssignmentID: assignmentID,
			DayOfWeek:    body.DayOfWeek,
			StartTime:    startTime,
			EndTime:      endTime,
			RoomLabel:    strings.TrimSpace(body.RoomLabel),
			Notes:        strings.TrimSpace(body.Notes),
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

func (h *Academic) Update(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	entity := chi.URLParam(r, "entity")
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
		return
	}
	if entity == "subjects" {
		var body subjectRequest
		if !decodeJSON(w, r, &body, defaultJSONBodyLimit) {
			return
		}
		row, err := h.svc.UpdateSubject(r.Context(), updateSubjectParams(id, body))
		if err != nil {
			writeClientError(w, err, "Data mapel tidak valid")
			return
		}
		api.OK(w, row)
		return
	}
	if entity != "timetables" {
		api.NotFound(w)
		return
	}
	var body struct {
		AssignmentID string `json:"assignment_id"`
		DayOfWeek    int16  `json:"day_of_week"`
		StartTime    string `json:"start_time"`
		EndTime      string `json:"end_time"`
		RoomLabel    string `json:"room_label"`
		Notes        string `json:"notes"`
	}
	if !decodeJSON(w, r, &body, defaultJSONBodyLimit) {
		return
	}
	assignmentID, err := parseUUID(body.AssignmentID)
	if err != nil {
		api.BadRequest(w, "Penugasan guru mapel tidak valid")
		return
	}
	if body.DayOfWeek < 1 || body.DayOfWeek > 6 {
		api.BadRequest(w, "Hari jadwal harus Senin sampai Sabtu")
		return
	}
	startTime, err := service.ParseAcademicTimeInput(body.StartTime)
	if err != nil {
		api.BadRequest(w, "Jam mulai tidak valid")
		return
	}
	endTime, err := service.ParseAcademicTimeInput(body.EndTime)
	if err != nil {
		api.BadRequest(w, "Jam selesai tidak valid")
		return
	}
	if startTime.Microseconds >= endTime.Microseconds {
		api.BadRequest(w, "rentang waktu tidak valid")
		return
	}
	row, err := h.svc.UpdateTimetableSlot(r.Context(), db.UpdateTimetableSlotParams{
		ID:           id,
		AssignmentID: assignmentID,
		DayOfWeek:    body.DayOfWeek,
		StartTime:    startTime,
		EndTime:      endTime,
		RoomLabel:    strings.TrimSpace(body.RoomLabel),
		Notes:        strings.TrimSpace(body.Notes),
	})
	if err != nil {
		writeClientError(w, err, "Perubahan jadwal akademik tidak valid")
		return
	}
	api.OK(w, row)
}

func (h *Academic) Delete(w http.ResponseWriter, r *http.Request) {
	if !adminAccessAllowed(r) {
		api.Forbidden(w)
		return
	}
	entity := chi.URLParam(r, "entity")
	switch entity {
	case "years", "classes", "subjects", "assignments", "timetables":
	default:
		api.NotFound(w)
		return
	}
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		api.BadRequest(w, "ID data tidak valid")
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
	case "timetables":
		err = h.svc.DeleteTimetableSlot(r.Context(), id)
	}
	if err != nil {
		api.Internal(w, err)
		return
	}
	api.NoContent(w)
}

type subjectRequest struct {
	Code                string `json:"code"`
	Name                string `json:"name"`
	Category            string `json:"category"`
	IsAssessmentSubject *bool  `json:"is_assessment_subject"`
	IsReportSubject     *bool  `json:"is_report_subject"`
	IsScheduleActivity  *bool  `json:"is_schedule_activity"`
	CountsForRanking    *bool  `json:"counts_for_ranking"`
	IsLocalContent      *bool  `json:"is_local_content"`
	IsChoiceSubject     *bool  `json:"is_choice_subject"`
	DefaultWeeklyHours  int32  `json:"default_weekly_hours"`
	DisplayOrder        int32  `json:"display_order"`
	IsActive            *bool  `json:"is_active"`
}

func createSubjectParams(body subjectRequest) db.CreateSubjectParams {
	return db.CreateSubjectParams{
		Code:                body.Code,
		Name:                body.Name,
		Category:            body.Category,
		IsAssessmentSubject: subjectBoolDefault(body.IsAssessmentSubject, true),
		IsReportSubject:     subjectBoolDefault(body.IsReportSubject, true),
		IsScheduleActivity:  subjectBoolDefault(body.IsScheduleActivity, false),
		CountsForRanking:    subjectBoolDefault(body.CountsForRanking, true),
		IsLocalContent:      subjectBoolDefault(body.IsLocalContent, false),
		IsChoiceSubject:     subjectBoolDefault(body.IsChoiceSubject, false),
		DefaultWeeklyHours:  body.DefaultWeeklyHours,
		DisplayOrder:        body.DisplayOrder,
		IsActive:            subjectBoolDefault(body.IsActive, true),
	}
}

func updateSubjectParams(id pgtype.UUID, body subjectRequest) db.UpdateSubjectParams {
	return db.UpdateSubjectParams{
		ID:                  id,
		Code:                body.Code,
		Name:                body.Name,
		Category:            body.Category,
		IsAssessmentSubject: subjectBoolDefault(body.IsAssessmentSubject, true),
		IsReportSubject:     subjectBoolDefault(body.IsReportSubject, true),
		IsScheduleActivity:  subjectBoolDefault(body.IsScheduleActivity, false),
		CountsForRanking:    subjectBoolDefault(body.CountsForRanking, true),
		IsLocalContent:      subjectBoolDefault(body.IsLocalContent, false),
		IsChoiceSubject:     subjectBoolDefault(body.IsChoiceSubject, false),
		DefaultWeeklyHours:  body.DefaultWeeklyHours,
		DisplayOrder:        body.DisplayOrder,
		IsActive:            subjectBoolDefault(body.IsActive, true),
	}
}

func subjectBoolDefault(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

type lessonPeriodRequest struct {
	AcademicYearID    string `json:"academic_year_id"`
	DayOfWeek         int16  `json:"day_of_week"`
	PeriodNumber      int32  `json:"period_number"`
	StartTime         string `json:"start_time"`
	EndTime           string `json:"end_time"`
	ActivityType      string `json:"activity_type"`
	Label             string `json:"label"`
	IsCountedAsLesson *bool  `json:"is_counted_as_lesson"`
}

func parseCreateLessonPeriod(w http.ResponseWriter, body lessonPeriodRequest) (db.CreateLessonPeriodTemplateParams, bool) {
	yearID, err := parseUUID(body.AcademicYearID)
	if err != nil {
		api.BadRequest(w, "Tahun ajaran tidak valid")
		return db.CreateLessonPeriodTemplateParams{}, false
	}
	start, end, ok := parseLessonPeriodTimes(w, body)
	if !ok {
		return db.CreateLessonPeriodTemplateParams{}, false
	}
	return db.CreateLessonPeriodTemplateParams{
		AcademicYearID:    yearID,
		DayOfWeek:         body.DayOfWeek,
		PeriodNumber:      body.PeriodNumber,
		StartTime:         start,
		EndTime:           end,
		ActivityType:      strings.TrimSpace(body.ActivityType),
		Label:             strings.TrimSpace(body.Label),
		IsCountedAsLesson: subjectBoolDefault(body.IsCountedAsLesson, true),
	}, true
}

func parseUpdateLessonPeriod(w http.ResponseWriter, id pgtype.UUID, body lessonPeriodRequest) (db.UpdateLessonPeriodTemplateParams, bool) {
	start, end, ok := parseLessonPeriodTimes(w, body)
	if !ok {
		return db.UpdateLessonPeriodTemplateParams{}, false
	}
	return db.UpdateLessonPeriodTemplateParams{
		ID:                id,
		DayOfWeek:         body.DayOfWeek,
		PeriodNumber:      body.PeriodNumber,
		StartTime:         start,
		EndTime:           end,
		ActivityType:      strings.TrimSpace(body.ActivityType),
		Label:             strings.TrimSpace(body.Label),
		IsCountedAsLesson: subjectBoolDefault(body.IsCountedAsLesson, true),
	}, true
}

func parseLessonPeriodTimes(w http.ResponseWriter, body lessonPeriodRequest) (pgtype.Time, pgtype.Time, bool) {
	start, err := service.ParseAcademicTimeInput(body.StartTime)
	if err != nil {
		api.BadRequest(w, "Jam mulai tidak valid")
		return pgtype.Time{}, pgtype.Time{}, false
	}
	end, err := service.ParseAcademicTimeInput(body.EndTime)
	if err != nil {
		api.BadRequest(w, "Jam selesai tidak valid")
		return pgtype.Time{}, pgtype.Time{}, false
	}
	return start, end, true
}
