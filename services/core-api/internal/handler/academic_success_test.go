package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
	"mtsn2kolut-super-app/backend/internal/service"
)

type fakeAcademicService struct {
	listYearsCalled     bool
	listClassesCalled   bool
	listSubjectsCalled  bool
	listAssignsCalled   bool
	listTimetableCalled bool
	listErr             error
	listYearsErr        error
	listClassesErr      error
	listSubjectsErr     error
	listAssignsErr      error
	listTimetableErr    error

	statsCalled bool
	statsErr    error

	dashboardCalled bool
	dashboardErr    error
	readinessCalled bool
	readinessErr    error

	weeklyCalled    bool
	weeklyErr       error
	conflictsCalled bool
	conflictsErr    error
	lessonCalled    bool
	lessonErr       error
	createLessonArg db.CreateLessonPeriodTemplateParams
	updateLessonArg db.UpdateLessonPeriodTemplateParams
	deleteLessonID  pgtype.UUID
	workloadCalled  bool
	workloadErr     error

	curriculumOverviewProfileID pgtype.UUID
	curriculumOverviewLevel     string
	curriculumAllocProfileID    pgtype.UUID
	curriculumAllocLevel        string
	curriculumSummaryProfileID  pgtype.UUID
	curriculumErr               error

	createYearArg      db.CreateAcademicYearParams
	activateYearID     pgtype.UUID
	activateConfirm    string
	previewInput       service.YearRolloverPreviewInput
	applyInput         service.YearRolloverApplyInput
	dryRunInput        service.AcademicImportDryRunInput
	createClassArg     db.CreateSchoolClassParams
	createSubjectArg   db.CreateSubjectParams
	createAssignArg    db.CreateClassSubjectAssignmentParams
	createTimetableArg db.CreateTimetableSlotParams
	createErr          error
	activateErr        error
	previewErr         error
	dryRunErr          error

	updateSubjectArg   db.UpdateSubjectParams
	updateTimetableArg db.UpdateTimetableSlotParams
	updateErr          error

	deleteYearID      pgtype.UUID
	deleteClassID     pgtype.UUID
	deleteSubjectID   pgtype.UUID
	deleteAssignID    pgtype.UUID
	deleteTimetableID pgtype.UUID
	deleteErr         error
}

func (f *fakeAcademicService) ListYears(context.Context) ([]db.AcademicYear, error) {
	f.listYearsCalled = true
	if f.listYearsErr != nil {
		return nil, f.listYearsErr
	}
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []db.AcademicYear{{Name: "2026/2027"}}, nil
}

func (f *fakeAcademicService) ListClasses(context.Context) ([]db.ListSchoolClassesRow, error) {
	f.listClassesCalled = true
	if f.listClassesErr != nil {
		return nil, f.listClassesErr
	}
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []db.ListSchoolClassesRow{{Name: "VII A"}}, nil
}

func (f *fakeAcademicService) ListSubjects(context.Context) ([]db.ListSubjectsRow, error) {
	f.listSubjectsCalled = true
	if f.listSubjectsErr != nil {
		return nil, f.listSubjectsErr
	}
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []db.ListSubjectsRow{{Name: "IPA"}}, nil
}

func (f *fakeAcademicService) ListAssignments(context.Context) ([]db.ListClassSubjectAssignmentsRow, error) {
	f.listAssignsCalled = true
	if f.listAssignsErr != nil {
		return nil, f.listAssignsErr
	}
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []db.ListClassSubjectAssignmentsRow{}, nil
}

func (f *fakeAcademicService) ListTimetableSlots(context.Context) ([]db.ListTimetableSlotsRow, error) {
	f.listTimetableCalled = true
	if f.listTimetableErr != nil {
		return nil, f.listTimetableErr
	}
	if f.listErr != nil {
		return nil, f.listErr
	}
	return []db.ListTimetableSlotsRow{}, nil
}

func (f *fakeAcademicService) GetLessonPeriodOverview(context.Context) (service.LessonPeriodOverview, error) {
	f.lessonCalled = true
	if f.lessonErr != nil {
		return service.LessonPeriodOverview{}, f.lessonErr
	}
	return service.LessonPeriodOverview{Items: []db.ListLessonPeriodTemplatesRow{}}, nil
}

func (f *fakeAcademicService) CreateLessonPeriodTemplate(_ context.Context, p db.CreateLessonPeriodTemplateParams) (db.LessonPeriodTemplate, error) {
	f.createLessonArg = p
	if f.createErr != nil {
		return db.LessonPeriodTemplate{}, f.createErr
	}
	return db.LessonPeriodTemplate{AcademicYearID: p.AcademicYearID, DayOfWeek: p.DayOfWeek, PeriodNumber: p.PeriodNumber, StartTime: p.StartTime, EndTime: p.EndTime, ActivityType: p.ActivityType, Label: p.Label, IsCountedAsLesson: p.IsCountedAsLesson}, nil
}

func (f *fakeAcademicService) UpdateLessonPeriodTemplate(_ context.Context, p db.UpdateLessonPeriodTemplateParams) (db.LessonPeriodTemplate, error) {
	f.updateLessonArg = p
	if f.updateErr != nil {
		return db.LessonPeriodTemplate{}, f.updateErr
	}
	return db.LessonPeriodTemplate{ID: p.ID, DayOfWeek: p.DayOfWeek, PeriodNumber: p.PeriodNumber, StartTime: p.StartTime, EndTime: p.EndTime, ActivityType: p.ActivityType, Label: p.Label, IsCountedAsLesson: p.IsCountedAsLesson}, nil
}

func (f *fakeAcademicService) DeleteLessonPeriodTemplate(_ context.Context, id pgtype.UUID) error {
	f.deleteLessonID = id
	return f.deleteErr
}

func (f *fakeAcademicService) GetTeacherWorkload(context.Context) (service.TeacherWorkloadOverview, error) {
	f.workloadCalled = true
	if f.workloadErr != nil {
		return service.TeacherWorkloadOverview{}, f.workloadErr
	}
	return service.TeacherWorkloadOverview{Items: []service.TeacherWorkloadRow{}}, nil
}

func (f *fakeAcademicService) GetStats(context.Context) (db.GetAcademicStatsRow, error) {
	f.statsCalled = true
	if f.statsErr != nil {
		return db.GetAcademicStatsRow{}, f.statsErr
	}
	return db.GetAcademicStatsRow{TotalYears: 1, TotalClasses: 2, TotalSubjects: 3}, nil
}

func (f *fakeAcademicService) GetDashboardSummary(context.Context) (db.GetAcademicDashboardSummaryRow, error) {
	f.dashboardCalled = true
	if f.dashboardErr != nil {
		return db.GetAcademicDashboardSummaryRow{}, f.dashboardErr
	}
	return db.GetAcademicDashboardSummaryRow{
		ActiveAcademicYear:     "2026/2027",
		ActiveSemester:         "Ganjil",
		TotalClasses:           7,
		TotalActiveStudents:    210,
		StudentsWithoutClass:   3,
		ClassesWithoutHomeroom: 1,
	}, nil
}

func (f *fakeAcademicService) GetReadinessSummary(context.Context) (db.GetAcademicReadinessSummaryRow, error) {
	f.readinessCalled = true
	if f.readinessErr != nil {
		return db.GetAcademicReadinessSummaryRow{}, f.readinessErr
	}
	return db.GetAcademicReadinessSummaryRow{
		ActiveAcademicYear:  "2026/2027",
		ActiveSemester:      "Ganjil",
		TotalClasses:        7,
		TotalActiveStudents: 210,
	}, nil
}

func (f *fakeAcademicService) GetCurriculumOverview(_ context.Context, profileID pgtype.UUID, level string) (service.CurriculumOverview, error) {
	f.curriculumOverviewProfileID = profileID
	f.curriculumOverviewLevel = level
	if f.curriculumErr != nil {
		return service.CurriculumOverview{}, f.curriculumErr
	}
	return service.CurriculumOverview{
		Profiles:       []db.CurriculumProfile{{ID: handlerTestUUID(201), Name: "Kurikulum Merdeka MTs KMA 1503 Tahun 2025", Status: "active"}},
		SummaryByLevel: []service.CurriculumLevelSummary{{GetCurriculumSummaryByLevelRow: db.GetCurriculumSummaryByLevelRow{Level: "VII", TotalAnnualHours: 1512, ComplianceStatus: "sesuai"}, TotalWeeklyHours: 42, StatusLabel: "Sesuai KMA"}},
		Allocations:    []service.CurriculumAllocation{{ListCurriculumSubjectAllocationsRow: db.ListCurriculumSubjectAllocationsRow{Level: "VII", SubjectName: "Al-Qur'an Hadis", TotalAnnualHours: 108}, TotalWeeklyHours: 3}},
	}, nil
}

func (f *fakeAcademicService) ListCurriculumProfiles(context.Context) ([]db.CurriculumProfile, error) {
	if f.curriculumErr != nil {
		return nil, f.curriculumErr
	}
	return []db.CurriculumProfile{{ID: handlerTestUUID(201), Name: "Kurikulum Merdeka MTs KMA 1503 Tahun 2025", Status: "active"}}, nil
}

func (f *fakeAcademicService) ListCurriculumAllocations(_ context.Context, profileID pgtype.UUID, level string) ([]service.CurriculumAllocation, error) {
	f.curriculumAllocProfileID = profileID
	f.curriculumAllocLevel = level
	if f.curriculumErr != nil {
		return nil, f.curriculumErr
	}
	return []service.CurriculumAllocation{{ListCurriculumSubjectAllocationsRow: db.ListCurriculumSubjectAllocationsRow{Level: "VII", SubjectName: "Al-Qur'an Hadis", TotalAnnualHours: 108}, TotalWeeklyHours: 3}}, nil
}

func (f *fakeAcademicService) GetCurriculumSummary(_ context.Context, profileID pgtype.UUID) ([]service.CurriculumLevelSummary, error) {
	f.curriculumSummaryProfileID = profileID
	if f.curriculumErr != nil {
		return nil, f.curriculumErr
	}
	return []service.CurriculumLevelSummary{{GetCurriculumSummaryByLevelRow: db.GetCurriculumSummaryByLevelRow{Level: "VII", TotalAnnualHours: 1512, ComplianceStatus: "sesuai"}, TotalWeeklyHours: 42, StatusLabel: "Sesuai KMA"}}, nil
}

func (f *fakeAcademicService) GetWeeklyTimetable(context.Context) (service.WeeklyTimetable, error) {
	f.weeklyCalled = true
	if f.weeklyErr != nil {
		return service.WeeklyTimetable{}, f.weeklyErr
	}
	return service.WeeklyTimetable{
		ActiveAcademicYearName: "2026/2027",
		Slots: []service.WeeklyTimetableSlot{{
			ListWeeklyTimetableSlotsRow: db.ListWeeklyTimetableSlotsRow{SubjectName: "Matematika"},
			ConflictStatus:              "ok",
			ConflictLabel:               "Aman",
		}},
	}, nil
}

func (f *fakeAcademicService) GetTimetableConflicts(context.Context) ([]db.ListTimetableConflictsRow, error) {
	f.conflictsCalled = true
	if f.conflictsErr != nil {
		return nil, f.conflictsErr
	}
	return []db.ListTimetableConflictsRow{{ConflictType: "same_teacher", Message: "Guru bentrok"}}, nil
}

func (f *fakeAcademicService) CreateYear(_ context.Context, p db.CreateAcademicYearParams) (db.AcademicYear, error) {
	f.createYearArg = p
	if f.createErr != nil {
		return db.AcademicYear{}, f.createErr
	}
	return db.AcademicYear{ID: handlerTestUUID(120), Name: p.Name}, nil
}

func (f *fakeAcademicService) ActivateYear(_ context.Context, id pgtype.UUID, confirmation string) (db.AcademicYear, error) {
	f.activateYearID = id
	f.activateConfirm = confirmation
	if f.activateErr != nil {
		return db.AcademicYear{}, f.activateErr
	}
	return db.AcademicYear{ID: id, Name: "2026/2027", IsActive: true}, nil
}

func (f *fakeAcademicService) PreviewYearRollover(_ context.Context, input service.YearRolloverPreviewInput) (service.YearRolloverPreview, error) {
	f.previewInput = input
	if f.previewErr != nil {
		return service.YearRolloverPreview{}, f.previewErr
	}
	return service.YearRolloverPreview{
		SourceAcademicYearName: "2025/2026",
		TargetAcademicYearName: "2026/2027",
		Counts: service.YearRolloverPreviewCounts{
			StudentsToPromote: 12,
		},
		Warnings: []string{"Preview ini tidak mengubah database."},
	}, nil
}

func (f *fakeAcademicService) ApplyYearRollover(_ context.Context, input service.YearRolloverApplyInput) (service.YearRolloverApplyResult, error) {
	f.applyInput = input
	if f.previewErr != nil {
		return service.YearRolloverApplyResult{}, f.previewErr
	}
	return service.YearRolloverApplyResult{
		SourceAcademicYearName: "2025/2026",
		TargetAcademicYearName: "2026/2027",
		Counts: service.YearRolloverApplyCounts{
			StudentsPromoted: 12,
		},
		Warnings: []string{"Apply rollover selesai."},
	}, nil
}

func (f *fakeAcademicService) DryRunAcademicImport(_ context.Context, input service.AcademicImportDryRunInput) (service.AcademicImportDryRunResult, error) {
	f.dryRunInput = input
	if f.dryRunErr != nil {
		return service.AcademicImportDryRunResult{}, f.dryRunErr
	}
	return service.AcademicImportDryRunResult{Kind: input.Kind, TotalRows: 1, AddCount: 1}, nil
}

func (f *fakeAcademicService) CreateClass(_ context.Context, p db.CreateSchoolClassParams) (db.SchoolClass, error) {
	f.createClassArg = p
	if f.createErr != nil {
		return db.SchoolClass{}, f.createErr
	}
	return db.SchoolClass{ID: handlerTestUUID(121), Name: p.Name}, nil
}

func (f *fakeAcademicService) CreateSubject(_ context.Context, p db.CreateSubjectParams) (db.Subject, error) {
	f.createSubjectArg = p
	if f.createErr != nil {
		return db.Subject{}, f.createErr
	}
	return db.Subject{ID: handlerTestUUID(122), Name: p.Name}, nil
}

func (f *fakeAcademicService) UpdateSubject(_ context.Context, p db.UpdateSubjectParams) (db.Subject, error) {
	f.updateSubjectArg = p
	if f.updateErr != nil {
		return db.Subject{}, f.updateErr
	}
	return db.Subject{ID: p.ID, Code: p.Code, Name: p.Name, IsActive: p.IsActive}, nil
}

func (f *fakeAcademicService) CreateAssignment(_ context.Context, p db.CreateClassSubjectAssignmentParams) (db.ClassSubjectAssignment, error) {
	f.createAssignArg = p
	if f.createErr != nil {
		return db.ClassSubjectAssignment{}, f.createErr
	}
	return db.ClassSubjectAssignment{ID: handlerTestUUID(123), ClassID: p.ClassID, SubjectID: p.SubjectID}, nil
}

func (f *fakeAcademicService) CreateTimetableSlot(_ context.Context, p db.CreateTimetableSlotParams) (db.TimetableSlot, error) {
	f.createTimetableArg = p
	if f.createErr != nil {
		return db.TimetableSlot{}, f.createErr
	}
	return db.TimetableSlot{ID: handlerTestUUID(124), AssignmentID: p.AssignmentID}, nil
}

func (f *fakeAcademicService) UpdateTimetableSlot(_ context.Context, p db.UpdateTimetableSlotParams) (db.TimetableSlot, error) {
	f.updateTimetableArg = p
	if f.updateErr != nil {
		return db.TimetableSlot{}, f.updateErr
	}
	return db.TimetableSlot{ID: p.ID, AssignmentID: p.AssignmentID}, nil
}

func (f *fakeAcademicService) DeleteYear(_ context.Context, id pgtype.UUID) error {
	f.deleteYearID = id
	return f.deleteErr
}

func (f *fakeAcademicService) DeleteClass(_ context.Context, id pgtype.UUID) error {
	f.deleteClassID = id
	return f.deleteErr
}

func (f *fakeAcademicService) DeleteSubject(_ context.Context, id pgtype.UUID) error {
	f.deleteSubjectID = id
	return f.deleteErr
}

func (f *fakeAcademicService) DeleteAssignment(_ context.Context, id pgtype.UUID) error {
	f.deleteAssignID = id
	return f.deleteErr
}

func (f *fakeAcademicService) DeleteTimetableSlot(_ context.Context, id pgtype.UUID) error {
	f.deleteTimetableID = id
	return f.deleteErr
}

func TestAcademicOverviewAndStatsSuccess(t *testing.T) {
	fake := &fakeAcademicService{}
	h := &Academic{svc: fake}

	rec := httptest.NewRecorder()
	h.Overview(rec, adminRequest(http.MethodGet, "/api/academic", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("Overview() status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if !fake.listYearsCalled || !fake.listClassesCalled || !fake.listSubjectsCalled || !fake.listAssignsCalled || !fake.listTimetableCalled {
		t.Fatalf("Overview calls years/classes/subjects/assigns/timetable = %v/%v/%v/%v/%v, want all true", fake.listYearsCalled, fake.listClassesCalled, fake.listSubjectsCalled, fake.listAssignsCalled, fake.listTimetableCalled)
	}
	if !strings.Contains(rec.Body.String(), "timetableSlots") {
		t.Fatalf("Overview() body missing timetableSlots: %s", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.GetStats(rec, adminRequest(http.MethodGet, "/api/academic/stats", ""))
	if rec.Code != http.StatusOK || !fake.statsCalled {
		t.Fatalf("GetStats() status/called = %d/%v, want 200/true; body=%s", rec.Code, fake.statsCalled, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.GetDashboard(rec, adminRequest(http.MethodGet, "/api/academic/dashboard", ""))
	if rec.Code != http.StatusOK || !fake.dashboardCalled || !strings.Contains(rec.Body.String(), "active_academic_year") {
		t.Fatalf("GetDashboard() status/called/body = %d/%v/%s, want 200/true/summary", rec.Code, fake.dashboardCalled, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.GetReadiness(rec, adminRequest(http.MethodGet, "/api/academic/readiness", ""))
	if rec.Code != http.StatusOK || !fake.readinessCalled || !strings.Contains(rec.Body.String(), "active_academic_year") {
		t.Fatalf("GetReadiness() status/called/body = %d/%v/%s, want 200/true/summary", rec.Code, fake.readinessCalled, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.GetWeeklyTimetable(rec, adminRequest(http.MethodGet, "/api/academic/timetable/weekly", ""))
	if rec.Code != http.StatusOK || !fake.weeklyCalled || !strings.Contains(rec.Body.String(), "conflict_status") {
		t.Fatalf("GetWeeklyTimetable() status/called/body = %d/%v/%s, want 200/true/weekly", rec.Code, fake.weeklyCalled, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	h.GetTimetableConflicts(rec, adminRequest(http.MethodGet, "/api/academic/timetable/conflicts", ""))
	if rec.Code != http.StatusOK || !fake.conflictsCalled || !strings.Contains(rec.Body.String(), "same_teacher") {
		t.Fatalf("GetTimetableConflicts() status/called/body = %d/%v/%s, want 200/true/conflict list", rec.Code, fake.conflictsCalled, rec.Body.String())
	}
}

func TestAcademicAdditionalValidationBranches(t *testing.T) {
	errDB := errors.New("db down")
	yearID := handlerTestUUID(142)
	classID := handlerTestUUID(143)
	subjectID := handlerTestUUID(144)
	teacherID := handlerTestUUID(145)
	assignmentID := handlerTestUUID(146)
	slotID := handlerTestUUID(147)
	validTimetableBody := `{"assignment_id":"` + assignmentID.String() + `","day_of_week":2,"start_time":"07:30","end_time":"08:50"}`

	plainRequest := func(method, target, body string) *http.Request {
		return httptest.NewRequest(method, target, strings.NewReader(body))
	}
	tests := []struct {
		name       string
		fn         func(*Academic, http.ResponseWriter, *http.Request)
		svc        *fakeAcademicService
		req        *http.Request
		wantStatus int
	}{
		{name: "overview class error", fn: (*Academic).Overview, svc: &fakeAcademicService{listClassesErr: errDB}, req: adminRequest(http.MethodGet, "/api/academic", ""), wantStatus: http.StatusInternalServerError},
		{name: "overview subject error", fn: (*Academic).Overview, svc: &fakeAcademicService{listSubjectsErr: errDB}, req: adminRequest(http.MethodGet, "/api/academic", ""), wantStatus: http.StatusInternalServerError},
		{name: "overview assignment error", fn: (*Academic).Overview, svc: &fakeAcademicService{listAssignsErr: errDB}, req: adminRequest(http.MethodGet, "/api/academic", ""), wantStatus: http.StatusInternalServerError},
		{name: "overview timetable error", fn: (*Academic).Overview, svc: &fakeAcademicService{listTimetableErr: errDB}, req: adminRequest(http.MethodGet, "/api/academic", ""), wantStatus: http.StatusInternalServerError},
		{name: "weekly timetable error", fn: (*Academic).GetWeeklyTimetable, svc: &fakeAcademicService{weeklyErr: errDB}, req: adminRequest(http.MethodGet, "/api/academic/timetable/weekly", ""), wantStatus: http.StatusInternalServerError},
		{name: "timetable conflicts error", fn: (*Academic).GetTimetableConflicts, svc: &fakeAcademicService{conflictsErr: errDB}, req: adminRequest(http.MethodGet, "/api/academic/timetable/conflicts", ""), wantStatus: http.StatusInternalServerError},
		{name: "create forbidden", fn: (*Academic).Create, req: withRouteParam(plainRequest(http.MethodPost, "/api/academic/subjects", `{}`), "entity", "subjects"), wantStatus: http.StatusForbidden},
		{name: "create year invalid end", fn: (*Academic).Create, req: withRouteParam(adminRequest(http.MethodPost, "/api/academic/years", `{"name":"2026","start_date":"2026-07-01","end_date":"bad"}`), "entity", "years"), wantStatus: http.StatusBadRequest},
		{name: "create class invalid json", fn: (*Academic).Create, req: withRouteParam(adminRequest(http.MethodPost, "/api/academic/classes", `{`), "entity", "classes"), wantStatus: http.StatusBadRequest},
		{name: "create subject invalid json", fn: (*Academic).Create, req: withRouteParam(adminRequest(http.MethodPost, "/api/academic/subjects", `{`), "entity", "subjects"), wantStatus: http.StatusBadRequest},
		{name: "create assignment invalid json", fn: (*Academic).Create, req: withRouteParam(adminRequest(http.MethodPost, "/api/academic/assignments", `{`), "entity", "assignments"), wantStatus: http.StatusBadRequest},
		{name: "create assignment invalid subject", fn: (*Academic).Create, req: withRouteParam(adminRequest(http.MethodPost, "/api/academic/assignments", `{"class_id":"`+classID.String()+`","subject_id":"bad"}`), "entity", "assignments"), wantStatus: http.StatusBadRequest},
		{name: "create assignment invalid teacher", fn: (*Academic).Create, req: withRouteParam(adminRequest(http.MethodPost, "/api/academic/assignments", `{"class_id":"`+classID.String()+`","subject_id":"`+subjectID.String()+`","teacher_employee_id":"bad"}`), "entity", "assignments"), wantStatus: http.StatusBadRequest},
		{name: "create timetable invalid assignment", fn: (*Academic).Create, req: withRouteParam(adminRequest(http.MethodPost, "/api/academic/timetables", `{"assignment_id":"bad"}`), "entity", "timetables"), wantStatus: http.StatusBadRequest},
		{name: "create timetable invalid start", fn: (*Academic).Create, req: withRouteParam(adminRequest(http.MethodPost, "/api/academic/timetables", `{"assignment_id":"`+assignmentID.String()+`","day_of_week":2,"start_time":"bad","end_time":"08:50"}`), "entity", "timetables"), wantStatus: http.StatusBadRequest},
		{name: "create timetable invalid end", fn: (*Academic).Create, req: withRouteParam(adminRequest(http.MethodPost, "/api/academic/timetables", `{"assignment_id":"`+assignmentID.String()+`","day_of_week":2,"start_time":"07:30","end_time":"bad"}`), "entity", "timetables"), wantStatus: http.StatusBadRequest},
		{name: "create timetable invalid range", fn: (*Academic).Create, req: withRouteParam(adminRequest(http.MethodPost, "/api/academic/timetables", `{"assignment_id":"`+assignmentID.String()+`","day_of_week":2,"start_time":"08:50","end_time":"07:30"}`), "entity", "timetables"), wantStatus: http.StatusBadRequest},
		{name: "create year service error", fn: (*Academic).Create, svc: &fakeAcademicService{createErr: errDB}, req: withRouteParam(adminRequest(http.MethodPost, "/api/academic/years", `{"name":"2026","start_date":"2026-07-01","end_date":"2027-06-30"}`), "entity", "years"), wantStatus: http.StatusInternalServerError},
		{name: "create class service error", fn: (*Academic).Create, svc: &fakeAcademicService{createErr: errDB}, req: withRouteParam(adminRequest(http.MethodPost, "/api/academic/classes", `{"academic_year_id":"`+yearID.String()+`","code":"VII-A"}`), "entity", "classes"), wantStatus: http.StatusInternalServerError},
		{name: "create assignment service error", fn: (*Academic).Create, svc: &fakeAcademicService{createErr: errDB}, req: withRouteParam(adminRequest(http.MethodPost, "/api/academic/assignments", `{"class_id":"`+classID.String()+`","subject_id":"`+subjectID.String()+`","teacher_employee_id":"`+teacherID.String()+`"}`), "entity", "assignments"), wantStatus: http.StatusInternalServerError},
		{name: "create timetable service error", fn: (*Academic).Create, svc: &fakeAcademicService{createErr: errDB}, req: withRouteParam(adminRequest(http.MethodPost, "/api/academic/timetables", validTimetableBody), "entity", "timetables"), wantStatus: http.StatusInternalServerError},
		{name: "update forbidden", fn: (*Academic).Update, req: withRouteParams(plainRequest(http.MethodPatch, "/api/academic/timetables/"+slotID.String(), validTimetableBody), "entity", "timetables", "id", slotID.String()), wantStatus: http.StatusForbidden},
		{name: "update invalid json", fn: (*Academic).Update, req: withRouteParams(adminRequest(http.MethodPatch, "/api/academic/timetables/"+slotID.String(), `{`), "entity", "timetables", "id", slotID.String()), wantStatus: http.StatusBadRequest},
		{name: "update invalid assignment", fn: (*Academic).Update, req: withRouteParams(adminRequest(http.MethodPatch, "/api/academic/timetables/"+slotID.String(), `{"assignment_id":"bad"}`), "entity", "timetables", "id", slotID.String()), wantStatus: http.StatusBadRequest},
		{name: "update invalid day", fn: (*Academic).Update, req: withRouteParams(adminRequest(http.MethodPatch, "/api/academic/timetables/"+slotID.String(), `{"assignment_id":"`+assignmentID.String()+`","day_of_week":7}`), "entity", "timetables", "id", slotID.String()), wantStatus: http.StatusBadRequest},
		{name: "update invalid start", fn: (*Academic).Update, req: withRouteParams(adminRequest(http.MethodPatch, "/api/academic/timetables/"+slotID.String(), `{"assignment_id":"`+assignmentID.String()+`","day_of_week":2,"start_time":"bad","end_time":"08:50"}`), "entity", "timetables", "id", slotID.String()), wantStatus: http.StatusBadRequest},
		{name: "update invalid end", fn: (*Academic).Update, req: withRouteParams(adminRequest(http.MethodPatch, "/api/academic/timetables/"+slotID.String(), `{"assignment_id":"`+assignmentID.String()+`","day_of_week":2,"start_time":"07:30","end_time":"bad"}`), "entity", "timetables", "id", slotID.String()), wantStatus: http.StatusBadRequest},
		{name: "update invalid range", fn: (*Academic).Update, req: withRouteParams(adminRequest(http.MethodPatch, "/api/academic/timetables/"+slotID.String(), `{"assignment_id":"`+assignmentID.String()+`","day_of_week":2,"start_time":"08:50","end_time":"07:30"}`), "entity", "timetables", "id", slotID.String()), wantStatus: http.StatusBadRequest},
		{name: "delete forbidden", fn: (*Academic).Delete, req: withRouteParams(plainRequest(http.MethodDelete, "/api/academic/subjects/"+slotID.String(), ""), "entity", "subjects", "id", slotID.String()), wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.svc
			if svc == nil {
				svc = &fakeAcademicService{}
			}
			rec := httptest.NewRecorder()
			tt.fn(&Academic{svc: svc}, rec, tt.req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestAcademicCreateUpdateAndDeleteSuccess(t *testing.T) {
	yearID := handlerTestUUID(130)
	classID := handlerTestUUID(131)
	subjectID := handlerTestUUID(132)
	teacherID := handlerTestUUID(133)
	assignmentID := handlerTestUUID(134)
	slotID := handlerTestUUID(135)
	fake := &fakeAcademicService{}
	h := &Academic{svc: fake}

	createCases := []struct {
		entity string
		body   string
		check  func(t *testing.T)
	}{
		{
			entity: "years",
			body:   `{"name":"2026/2027","start_date":"2026-07-01","end_date":"2027-06-30","is_active":true}`,
			check: func(t *testing.T) {
				t.Helper()
				if fake.createYearArg.Name != "2026/2027" || !fake.createYearArg.StartDate.Valid || !fake.createYearArg.EndDate.Valid || !fake.createYearArg.IsActive {
					t.Fatalf("CreateYear arg = %+v, want mapped year", fake.createYearArg)
				}
			},
		},
		{
			entity: "classes",
			body:   `{"academic_year_id":"` + yearID.String() + `","code":"VII-A","name":"VII A","level":"VII","is_active":true}`,
			check: func(t *testing.T) {
				t.Helper()
				if fake.createClassArg.AcademicYearID != yearID || fake.createClassArg.Code != "VII-A" || fake.createClassArg.Level != "VII" {
					t.Fatalf("CreateClass arg = %+v, want mapped class", fake.createClassArg)
				}
			},
		},
		{
			entity: "subjects",
			body:   `{"code":"IPA","name":"Ilmu Pengetahuan Alam","category":"intrakurikuler","is_assessment_subject":true,"is_report_subject":true,"is_schedule_activity":false,"default_weekly_hours":5,"display_order":20,"is_active":true}`,
			check: func(t *testing.T) {
				t.Helper()
				if fake.createSubjectArg.Code != "IPA" || fake.createSubjectArg.Name != "Ilmu Pengetahuan Alam" || fake.createSubjectArg.Category != "intrakurikuler" || fake.createSubjectArg.DefaultWeeklyHours != 5 || fake.createSubjectArg.DisplayOrder != 20 || !fake.createSubjectArg.IsAssessmentSubject || !fake.createSubjectArg.IsReportSubject || fake.createSubjectArg.IsScheduleActivity || !fake.createSubjectArg.IsActive {
					t.Fatalf("CreateSubject arg = %+v, want mapped subject", fake.createSubjectArg)
				}
			},
		},
		{
			entity: "assignments",
			body:   `{"class_id":"` + classID.String() + `","subject_id":"` + subjectID.String() + `","teacher_employee_id":"` + teacherID.String() + `"}`,
			check: func(t *testing.T) {
				t.Helper()
				if fake.createAssignArg.ClassID != classID || fake.createAssignArg.SubjectID != subjectID || fake.createAssignArg.TeacherEmployeeID != teacherID {
					t.Fatalf("CreateAssignment arg = %+v, want mapped ids", fake.createAssignArg)
				}
			},
		},
		{
			entity: "timetables",
			body:   `{"assignment_id":"` + assignmentID.String() + `","day_of_week":2,"start_time":"07:30","end_time":"08:50","room_label":" Lab IPA ","notes":" Praktikum "}`,
			check: func(t *testing.T) {
				t.Helper()
				if fake.createTimetableArg.AssignmentID != assignmentID || fake.createTimetableArg.DayOfWeek != 2 || fake.createTimetableArg.RoomLabel != "Lab IPA" || fake.createTimetableArg.Notes != "Praktikum" {
					t.Fatalf("CreateTimetableSlot arg = %+v, want mapped timetable", fake.createTimetableArg)
				}
				if fake.createTimetableArg.StartTime.Microseconds >= fake.createTimetableArg.EndTime.Microseconds {
					t.Fatalf("CreateTimetableSlot times = %+v/%+v, want increasing range", fake.createTimetableArg.StartTime, fake.createTimetableArg.EndTime)
				}
			},
		},
	}

	for _, tt := range createCases {
		t.Run("create "+tt.entity, func(t *testing.T) {
			rec := httptest.NewRecorder()
			h.Create(rec, withRouteParam(adminRequest(http.MethodPost, "/api/academic/"+tt.entity, tt.body), "entity", tt.entity))
			if rec.Code != http.StatusCreated {
				t.Fatalf("Create(%s) status = %d, want 201; body=%s", tt.entity, rec.Code, rec.Body.String())
			}
			tt.check(t)
		})
	}

	rec := httptest.NewRecorder()
	updateReq := withRouteParams(adminRequest(http.MethodPut, "/api/academic/subjects/"+subjectID.String(), `{"code":"MTK","name":"Matematika","category":"intrakurikuler","is_assessment_subject":true,"is_report_subject":true,"is_schedule_activity":false,"default_weekly_hours":6,"display_order":10,"is_active":true}`), "entity", "subjects", "id", subjectID.String())
	h.Update(rec, updateReq)
	if rec.Code != http.StatusOK {
		t.Fatalf("Update(subjects) status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.updateSubjectArg.ID != subjectID || fake.updateSubjectArg.Code != "MTK" || fake.updateSubjectArg.DefaultWeeklyHours != 6 || fake.updateSubjectArg.DisplayOrder != 10 || !fake.updateSubjectArg.IsActive {
		t.Fatalf("UpdateSubject arg = %+v, want mapped subject", fake.updateSubjectArg)
	}

	rec = httptest.NewRecorder()
	updateReq = withRouteParams(adminRequest(http.MethodPatch, "/api/academic/timetables/"+slotID.String(), `{"assignment_id":"`+assignmentID.String()+`","day_of_week":3,"start_time":"09:00","end_time":"10:20","room_label":"Ruang 2","notes":"Ulangan"}`), "entity", "timetables", "id", slotID.String())
	h.Update(rec, updateReq)
	if rec.Code != http.StatusOK {
		t.Fatalf("Update(timetables) status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if fake.updateTimetableArg.ID != slotID || fake.updateTimetableArg.AssignmentID != assignmentID || fake.updateTimetableArg.DayOfWeek != 3 || fake.updateTimetableArg.RoomLabel != "Ruang 2" {
		t.Fatalf("UpdateTimetableSlot arg = %+v, want mapped update", fake.updateTimetableArg)
	}

	deleteCases := []struct {
		entity string
		check  func(t *testing.T)
	}{
		{entity: "years", check: func(t *testing.T) {
			t.Helper()
			if fake.deleteYearID != slotID {
				t.Fatalf("DeleteYear id = %v, want %v", fake.deleteYearID, slotID)
			}
		}},
		{entity: "classes", check: func(t *testing.T) {
			t.Helper()
			if fake.deleteClassID != slotID {
				t.Fatalf("DeleteClass id = %v, want %v", fake.deleteClassID, slotID)
			}
		}},
		{entity: "subjects", check: func(t *testing.T) {
			t.Helper()
			if fake.deleteSubjectID != slotID {
				t.Fatalf("DeleteSubject id = %v, want %v", fake.deleteSubjectID, slotID)
			}
		}},
		{entity: "assignments", check: func(t *testing.T) {
			t.Helper()
			if fake.deleteAssignID != slotID {
				t.Fatalf("DeleteAssignment id = %v, want %v", fake.deleteAssignID, slotID)
			}
		}},
		{entity: "timetables", check: func(t *testing.T) {
			t.Helper()
			if fake.deleteTimetableID != slotID {
				t.Fatalf("DeleteTimetableSlot id = %v, want %v", fake.deleteTimetableID, slotID)
			}
		}},
	}
	for _, tt := range deleteCases {
		t.Run("delete "+tt.entity, func(t *testing.T) {
			rec := httptest.NewRecorder()
			h.Delete(rec, withRouteParams(adminRequest(http.MethodDelete, "/api/academic/"+tt.entity+"/"+slotID.String(), ""), "entity", tt.entity, "id", slotID.String()))
			if rec.Code != http.StatusNoContent {
				t.Fatalf("Delete(%s) status = %d, want 204; body=%s", tt.entity, rec.Code, rec.Body.String())
			}
			tt.check(t)
		})
	}
}

func TestAcademicHandlersMapServiceErrors(t *testing.T) {
	slotID := handlerTestUUID(140)
	assignmentID := handlerTestUUID(141)
	tests := []struct {
		name string
		fn   func(http.ResponseWriter, *http.Request)
		req  *http.Request
		want int
	}{
		{
			name: "overview internal",
			fn:   (&Academic{svc: &fakeAcademicService{listErr: errors.New("db down")}}).Overview,
			req:  adminRequest(http.MethodGet, "/api/academic", ""),
			want: http.StatusInternalServerError,
		},
		{
			name: "stats internal",
			fn:   (&Academic{svc: &fakeAcademicService{statsErr: errors.New("db down")}}).GetStats,
			req:  adminRequest(http.MethodGet, "/api/academic/stats", ""),
			want: http.StatusInternalServerError,
		},
		{
			name: "create internal",
			fn:   (&Academic{svc: &fakeAcademicService{createErr: errors.New("db down")}}).Create,
			req:  withRouteParam(adminRequest(http.MethodPost, "/api/academic/subjects", `{"code":"IPA","name":"IPA"}`), "entity", "subjects"),
			want: http.StatusInternalServerError,
		},
		{
			name: "update client error",
			fn:   (&Academic{svc: &fakeAcademicService{updateErr: errors.New("slot bentrok")}}).Update,
			req:  withRouteParams(adminRequest(http.MethodPatch, "/api/academic/timetables/"+slotID.String(), `{"assignment_id":"`+assignmentID.String()+`","day_of_week":2,"start_time":"07:30","end_time":"08:50"}`), "entity", "timetables", "id", slotID.String()),
			want: http.StatusBadRequest,
		},
		{
			name: "delete internal",
			fn:   (&Academic{svc: &fakeAcademicService{deleteErr: errors.New("db down")}}).Delete,
			req:  withRouteParams(adminRequest(http.MethodDelete, "/api/academic/subjects/"+slotID.String(), ""), "entity", "subjects", "id", slotID.String()),
			want: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.fn(rec, tt.req)
			if rec.Code != tt.want {
				t.Fatalf("%s status = %d, want %d; body=%s", tt.name, rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}
