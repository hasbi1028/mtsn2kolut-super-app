package service

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestAcademicLessonPeriodValidationAndOverview(t *testing.T) {
	yearID := documentCycleTestUUID(181)
	lessonID := documentCycleTestUUID(182)
	start := mustAcademicTime(t, "07:30")
	end := mustAcademicTime(t, "08:10")
	store := &fakeAcademicStore{
		activeYear:    db.AcademicYear{ID: yearID, Name: "2026/2027"},
		lessonPeriods: []db.ListLessonPeriodTemplatesRow{{ID: lessonID, PeriodNumber: 1, ActivityType: "pelajaran"}},
	}
	svc := &Academic{q: store}

	overview, err := svc.GetLessonPeriodOverview(context.Background())
	if err != nil {
		t.Fatalf("GetLessonPeriodOverview() error = %v", err)
	}
	if overview.ActiveAcademicYearID != yearID || overview.ActiveAcademicYearName != "2026/2027" || len(overview.Items) != 1 {
		t.Fatalf("GetLessonPeriodOverview() = %+v, want active year and one template", overview)
	}

	created, err := svc.CreateLessonPeriodTemplate(context.Background(), db.CreateLessonPeriodTemplateParams{
		AcademicYearID:    yearID,
		DayOfWeek:         1,
		PeriodNumber:      1,
		StartTime:         start,
		EndTime:           end,
		ActivityType:      " pelajaran ",
		Label:             "Jam 1",
		IsCountedAsLesson: true,
	})
	if err != nil {
		t.Fatalf("CreateLessonPeriodTemplate(valid) error = %v", err)
	}
	if created.DayOfWeek != 1 || created.PeriodNumber != 1 || store.createLessonArg.ActivityType != " pelajaran " {
		t.Fatalf("CreateLessonPeriodTemplate() = %+v arg=%+v, want forwarded validated period", created, store.createLessonArg)
	}

	updated, err := svc.UpdateLessonPeriodTemplate(context.Background(), db.UpdateLessonPeriodTemplateParams{
		ID:                lessonID,
		DayOfWeek:         6,
		PeriodNumber:      2,
		StartTime:         start,
		EndTime:           end,
		ActivityType:      "kokurikuler",
		Label:             "P5RA",
		IsCountedAsLesson: true,
	})
	if err != nil {
		t.Fatalf("UpdateLessonPeriodTemplate(valid) error = %v", err)
	}
	if updated.ID != lessonID || store.updateLessonArg.ActivityType != "kokurikuler" {
		t.Fatalf("UpdateLessonPeriodTemplate() = %+v arg=%+v, want forwarded update", updated, store.updateLessonArg)
	}
	if err := svc.DeleteLessonPeriodTemplate(context.Background(), lessonID); err != nil || store.deleteLessonID != lessonID {
		t.Fatalf("DeleteLessonPeriodTemplate() = %v id=%v, want nil/%v", err, store.deleteLessonID, lessonID)
	}

	invalidCases := []struct {
		name string
		arg  db.CreateLessonPeriodTemplateParams
	}{
		{name: "invalid day", arg: db.CreateLessonPeriodTemplateParams{DayOfWeek: 0, PeriodNumber: 1, StartTime: start, EndTime: end, ActivityType: "pelajaran"}},
		{name: "invalid number", arg: db.CreateLessonPeriodTemplateParams{DayOfWeek: 1, PeriodNumber: 0, StartTime: start, EndTime: end, ActivityType: "pelajaran"}},
		{name: "invalid range", arg: db.CreateLessonPeriodTemplateParams{DayOfWeek: 1, PeriodNumber: 1, StartTime: end, EndTime: start, ActivityType: "pelajaran"}},
		{name: "invalid activity", arg: db.CreateLessonPeriodTemplateParams{DayOfWeek: 1, PeriodNumber: 1, StartTime: start, EndTime: end, ActivityType: "rapat"}},
	}
	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := svc.CreateLessonPeriodTemplate(context.Background(), tc.arg); !errors.Is(err, domain.ErrBadRequest) {
				t.Fatalf("CreateLessonPeriodTemplate(%s) error = %v, want ErrBadRequest", tc.name, err)
			}
		})
	}

	emptySvc := &Academic{q: &fakeAcademicStore{activeYearErr: pgx.ErrNoRows}}
	empty, err := emptySvc.GetLessonPeriodOverview(context.Background())
	if err != nil || len(empty.Items) != 0 || empty.ActiveAcademicYearName != "" {
		t.Fatalf("GetLessonPeriodOverview(no active year) = %+v, %v; want empty nil", empty, err)
	}
}

func TestAcademicTeacherWorkloadGroupsAssignmentsAndLabelsStatus(t *testing.T) {
	yearID := documentCycleTestUUID(191)
	teacherA := documentCycleTestUUID(192)
	teacherB := documentCycleTestUUID(193)
	teacherC := documentCycleTestUUID(194)
	store := &fakeAcademicStore{
		activeYear: db.AcademicYear{ID: yearID, Name: "2026/2027"},
		weeklyAssignments: []db.ListWeeklyTimetableAssignmentsRow{
			{ID: documentCycleTestUUID(195), TeacherEmployeeID: teacherA, TeacherName: "Guru Kurang", TeacherNip: "A", ExpectedWeeklyHours: 10},
			{ID: documentCycleTestUUID(196), TeacherEmployeeID: teacherA, TeacherName: "Guru Kurang", TeacherNip: "A", ExpectedWeeklyHours: 12},
			{ID: documentCycleTestUUID(197), TeacherEmployeeID: teacherB, TeacherName: "Guru Cukup", TeacherNip: "B", ExpectedWeeklyHours: 24},
			{ID: documentCycleTestUUID(198), TeacherEmployeeID: teacherC, TeacherName: "Guru Lebih", TeacherNip: "C", ExpectedWeeklyHours: 41},
		},
	}
	svc := &Academic{q: store}

	overview, err := svc.GetTeacherWorkload(context.Background())
	if err != nil {
		t.Fatalf("GetTeacherWorkload() error = %v", err)
	}
	if overview.ActiveAcademicYearID != yearID || overview.ActiveAcademicYearName != "2026/2027" || len(overview.Items) != 3 {
		t.Fatalf("GetTeacherWorkload() = %+v, want active year and three teacher rows", overview)
	}
	byTeacher := map[string]TeacherWorkloadRow{}
	for _, item := range overview.Items {
		byTeacher[item.TeacherEmployeeID.String()] = item
	}
	assertWorkloadRow(t, byTeacher[teacherA.String()], 22, 2, "kurang", "Kurang dari 24 JP")
	assertWorkloadRow(t, byTeacher[teacherB.String()], 24, 1, "cukup", "Cukup")
	assertWorkloadRow(t, byTeacher[teacherC.String()], 41, 1, "lebih", "Lebih dari 40 JP")

	emptySvc := &Academic{q: &fakeAcademicStore{activeYearErr: pgx.ErrNoRows}}
	empty, err := emptySvc.GetTeacherWorkload(context.Background())
	if err != nil || len(empty.Items) != 0 {
		t.Fatalf("GetTeacherWorkload(no active year) = %+v, %v; want empty nil", empty, err)
	}
}

func TestAcademicCurriculumOverviewConvertsNumbersAndHandlesMissingActiveProfile(t *testing.T) {
	profileID := documentCycleTestUUID(201)
	store := &fakeAcademicStore{
		curriculumProfiles: []db.CurriculumProfile{{ID: profileID, Code: "KMA", Name: "Kurikulum Madrasah", Status: "active"}},
		activeCurriculum:   db.CurriculumProfile{ID: profileID, Code: "KMA", Name: "Kurikulum Madrasah", Status: "active"},
		curriculumAllocs: []db.ListCurriculumSubjectAllocationsRow{{
			CurriculumProfileID: profileID,
			SubjectCode:         "MTK",
			SubjectName:         "Matematika",
			Level:               "VII",
			IntraWeeklyHours:    testNumeric(1234, -2),
			KokuWeeklyHours:     testNumeric(5, 0),
			TotalWeeklyHours:    testNumeric(1734, -2),
		}},
		curriculumSummary: []db.GetCurriculumSummaryByLevelRow{
			{CurriculumProfileID: profileID, Level: "VII", SubjectCount: 12, IntraWeeklyHours: testNumeric(3215, -2), KokuWeeklyHours: testNumeric(100, -2), TotalWeeklyHours: testNumeric(3315, -2), ComplianceStatus: "sesuai"},
			{CurriculumProfileID: profileID, Level: "VIII", SubjectCount: 12, TotalWeeklyHours: testNumeric(21, 0), ComplianceStatus: "kurang"},
		},
		classCurricula: []db.ListClassCurriculumAssignmentsRow{{ClassName: "VII A", CurriculumProfileID: profileID}},
	}
	svc := &Academic{q: store}

	profiles, err := svc.ListCurriculumProfiles(context.Background())
	if err != nil || len(profiles) != 1 || profiles[0].ID != profileID {
		t.Fatalf("ListCurriculumProfiles() = %+v, %v; want seeded profile", profiles, err)
	}
	activeProfile, err := svc.GetActiveCurriculumProfile(context.Background())
	if err != nil || activeProfile.ID != profileID {
		t.Fatalf("GetActiveCurriculumProfile() = %+v, %v; want active profile", activeProfile, err)
	}
	classAssignments, err := svc.ListClassCurriculumAssignments(context.Background(), profileID)
	if err != nil || len(classAssignments) != 1 || classAssignments[0].CurriculumProfileID != profileID {
		t.Fatalf("ListClassCurriculumAssignments() = %+v, %v; want seeded class assignment", classAssignments, err)
	}

	overview, err := svc.GetCurriculumOverview(context.Background(), pgtype.UUID{}, " vii ")
	if err != nil {
		t.Fatalf("GetCurriculumOverview() error = %v", err)
	}
	if overview.ActiveProfile == nil || overview.ActiveProfile.ID != profileID || len(overview.Profiles) != 1 || len(overview.Allocations) != 1 || len(overview.SummaryByLevel) != 2 || len(overview.ClassAssignments) != 1 {
		t.Fatalf("GetCurriculumOverview() = %+v, want active profile, allocation, summary, class assignment", overview)
	}
	if overview.Allocations[0].IntraWeeklyHours != 12.34 || overview.Allocations[0].KokuWeeklyHours != 5 || overview.Allocations[0].TotalWeeklyHours != 17.34 {
		t.Fatalf("allocation hours = %+v, want converted rounded numeric hours", overview.Allocations[0])
	}
	if overview.SummaryByLevel[0].StatusLabel != "Sesuai KMA" || overview.SummaryByLevel[1].StatusLabel != "Perlu ditinjau" || overview.SummaryByLevel[0].TotalWeeklyHours != 33.15 {
		t.Fatalf("summary labels/hours = %+v, want Indonesian compliance labels and numeric conversion", overview.SummaryByLevel)
	}

	emptySvc := &Academic{q: &fakeAcademicStore{}}
	empty, err := emptySvc.GetCurriculumOverview(context.Background(), pgtype.UUID{}, "")
	if err != nil || len(empty.Allocations) != 0 || len(empty.SummaryByLevel) != 0 || len(empty.ClassAssignments) != 0 || empty.ActiveProfile != nil {
		t.Fatalf("GetCurriculumOverview(no active profile) = %+v, %v; want empty nil", empty, err)
	}
}

func TestAcademicSubjectValidationAndDuplicateCode(t *testing.T) {
	subjectID := documentCycleTestUUID(211)
	store := &fakeAcademicStore{}
	svc := &Academic{q: store}

	created, err := svc.CreateSubject(context.Background(), db.CreateSubjectParams{Code: " MTK ", Name: " Matematika ", Category: "", DefaultWeeklyHours: 4, DisplayOrder: 2, IsActive: true})
	if err != nil {
		t.Fatalf("CreateSubject(valid) error = %v", err)
	}
	if created.Code != "MTK" || store.createSubjectArg.Name != "Matematika" || store.createSubjectArg.Category != "intrakurikuler" {
		t.Fatalf("CreateSubject() = %+v arg=%+v, want trimmed fields and default category", created, store.createSubjectArg)
	}

	updated, err := svc.UpdateSubject(context.Background(), db.UpdateSubjectParams{ID: subjectID, Code: " IPA ", Name: " Ilmu Pengetahuan Alam ", Category: " kokurikuler ", DefaultWeeklyHours: 3, DisplayOrder: 5, IsActive: true})
	if err != nil {
		t.Fatalf("UpdateSubject(valid) error = %v", err)
	}
	if updated.ID != subjectID || store.updateSubjectArg.Code != "IPA" || store.updateSubjectArg.Name != "Ilmu Pengetahuan Alam" || store.updateSubjectArg.Category != "kokurikuler" {
		t.Fatalf("UpdateSubject() = %+v arg=%+v, want trimmed update", updated, store.updateSubjectArg)
	}

	invalidCreates := []struct {
		name string
		arg  db.CreateSubjectParams
	}{
		{name: "blank code", arg: db.CreateSubjectParams{Name: "Matematika"}},
		{name: "blank name", arg: db.CreateSubjectParams{Code: "MTK"}},
		{name: "negative hours", arg: db.CreateSubjectParams{Code: "MTK", Name: "Matematika", DefaultWeeklyHours: -1}},
		{name: "too many hours", arg: db.CreateSubjectParams{Code: "MTK", Name: "Matematika", DefaultWeeklyHours: 61}},
		{name: "negative order", arg: db.CreateSubjectParams{Code: "MTK", Name: "Matematika", DisplayOrder: -1}},
	}
	for _, tc := range invalidCreates {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := svc.CreateSubject(context.Background(), tc.arg); !errors.Is(err, domain.ErrBadRequest) {
				t.Fatalf("CreateSubject(%s) error = %v, want ErrBadRequest", tc.name, err)
			}
		})
	}

	conflictSvc := &Academic{q: &fakeAcademicStore{subjectConflictCount: 1}}
	if _, err := conflictSvc.UpdateSubject(context.Background(), db.UpdateSubjectParams{ID: subjectID, Code: "MTK", Name: "Matematika"}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("UpdateSubject(duplicate code) error = %v, want ErrConflict", err)
	}
	missingSvc := &Academic{q: &fakeAcademicStore{getSubjectErr: pgx.ErrNoRows}}
	if _, err := missingSvc.UpdateSubject(context.Background(), db.UpdateSubjectParams{ID: subjectID, Code: "MTK", Name: "Matematika"}); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("UpdateSubject(missing subject) error = %v, want pgx.ErrNoRows", err)
	}
}

func mustAcademicTime(t *testing.T, value string) pgtype.Time {
	t.Helper()
	parsed, err := ParseAcademicTimeInput(value)
	if err != nil {
		t.Fatalf("ParseAcademicTimeInput(%q) error = %v", value, err)
	}
	return parsed
}

func testNumeric(value int64, exp int32) pgtype.Numeric {
	return pgtype.Numeric{Int: big.NewInt(value), Exp: exp, Valid: true}
}

func assertWorkloadRow(t *testing.T, got TeacherWorkloadRow, hours float64, assignments int, status string, label string) {
	t.Helper()
	if got.TotalWeeklyHours != hours || got.AssignmentCount != assignments || got.Status != status || got.StatusLabel != label || len(got.Assignments) != assignments {
		t.Fatalf("workload row = %+v, want hours %.2f assignments %d status %q label %q", got, hours, assignments, status, label)
	}
}
