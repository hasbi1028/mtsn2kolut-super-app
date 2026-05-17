package db

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestIntegrationRombelAssignmentsHomeroomAndTimetableQueries(t *testing.T) {
	t.Parallel()

	tdb := setupIntegrationTestDB(t)
	ctx := context.Background()
	q := tdb.Q

	suffix := integrationSuffix()
	year, err := q.CreateAcademicYear(ctx, CreateAcademicYearParams{
		Name:      "Integration Rombel Year " + suffix,
		StartDate: pgDate(2026, time.July, 1),
		EndDate:   pgDate(2027, time.June, 30),
		IsActive:  false,
	})
	if err != nil {
		t.Fatalf("create academic year: %v", err)
	}

	class, err := q.CreateSchoolClass(ctx, CreateSchoolClassParams{
		AcademicYearID: year.ID,
		Code:           "IT-ROM-" + suffix,
		Name:           "Integration Rombel " + suffix,
		Level:          "7",
		IsActive:       true,
	})
	if err != nil {
		t.Fatalf("create school class: %v", err)
	}

	subject, err := q.CreateSubject(ctx, CreateSubjectParams{
		Code:                "IT-ROM-SUB-" + suffix,
		Name:                "Integration Rombel Subject " + suffix,
		Category:            "umum",
		IsAssessmentSubject: true,
		IsReportSubject:     true,
		IsScheduleActivity:  false,
		CountsForRanking:    true,
		IsLocalContent:      false,
		IsChoiceSubject:     false,
		DefaultWeeklyHours:  2,
		DisplayOrder:        9998,
		IsActive:            true,
	})
	if err != nil {
		t.Fatalf("create subject: %v", err)
	}

	teacherID := insertIntegrationEmployee(t, ctx, tdb, "IT-TEA-"+suffix, "Integration Teacher "+suffix, employeeUID('1', suffix))
	homeroomTeacherID := insertIntegrationEmployee(t, ctx, tdb, "IT-HOM-"+suffix, "Integration Homeroom "+suffix, employeeUID('2', suffix))
	studentID := insertIntegrationStudent(t, ctx, tdb, "IT-NIS-"+suffix, "Integration Student "+suffix, class.ID)

	homeroom, err := q.CreateHomeroomAssignment(ctx, CreateHomeroomAssignmentParams{
		ClassID:                class.ID,
		HomeroomIsActive:       true,
		EmployeeID:             homeroomTeacherID,
		HomeroomAcademicYearID: year.ID,
		HomeroomStartDate:      pgDate(2026, time.July, 15),
		HomeroomEndDate:        pgtype.Date{},
		Notes:                  "integration homeroom",
	})
	if err != nil {
		t.Fatalf("create homeroom assignment: %v", err)
	}
	if homeroom.ClassID != class.ID || homeroom.EmployeeID != homeroomTeacherID || !homeroom.IsActive {
		t.Fatalf("created homeroom = %#v", homeroom)
	}

	homerooms, err := q.ListHomeroomAssignmentsByClass(ctx, class.ID)
	if err != nil {
		t.Fatalf("list homeroom assignments by class: %v", err)
	}
	if !homeroomAssignmentListed(homerooms, homeroom.ID) {
		t.Fatalf("created homeroom assignment %v not returned by ListHomeroomAssignmentsByClass", homeroom.ID)
	}

	assignment, err := q.CreateRombelSubjectAssignment(ctx, CreateRombelSubjectAssignmentParams{
		ClassID:           class.ID,
		SubjectID:         subject.ID,
		TeacherEmployeeID: teacherID,
	})
	if err != nil {
		t.Fatalf("create rombel subject assignment: %v", err)
	}
	if assignment.ClassID != class.ID || assignment.SubjectID != subject.ID || assignment.TeacherEmployeeID != teacherID {
		t.Fatalf("created assignment = %#v", assignment)
	}

	gotAssignment, err := q.GetRombelSubjectAssignment(ctx, GetRombelSubjectAssignmentParams{ClassID: class.ID, ID: assignment.ID})
	if err != nil {
		t.Fatalf("get rombel subject assignment: %v", err)
	}
	if gotAssignment.ID != assignment.ID || gotAssignment.SubjectCode != subject.Code || gotAssignment.TeacherName == "" {
		t.Fatalf("got assignment = %#v", gotAssignment)
	}

	assignments, err := q.ListRombelSubjectAssignments(ctx, class.ID)
	if err != nil {
		t.Fatalf("list rombel subject assignments: %v", err)
	}
	if !rombelSubjectAssignmentListed(assignments, assignment.ID) {
		t.Fatalf("created assignment %v not returned by ListRombelSubjectAssignments", assignment.ID)
	}

	weeklyAssignments, err := q.ListWeeklyTimetableAssignments(ctx, year.ID)
	if err != nil {
		t.Fatalf("list weekly timetable assignments: %v", err)
	}
	if !weeklyTimetableAssignmentListed(weeklyAssignments, assignment.ID) {
		t.Fatalf("created assignment %v not returned by ListWeeklyTimetableAssignments", assignment.ID)
	}

	period, err := q.CreateLessonPeriodTemplate(ctx, CreateLessonPeriodTemplateParams{
		AcademicYearID:    year.ID,
		DayOfWeek:         1,
		PeriodNumber:      1,
		StartTime:         pgTime(7, 30),
		EndTime:           pgTime(8, 50),
		ActivityType:      "pelajaran",
		Label:             "Jam 1-2",
		IsCountedAsLesson: true,
	})
	if err != nil {
		t.Fatalf("create lesson period template: %v", err)
	}

	slot, err := q.CreateRombelTimetableSlot(ctx, CreateRombelTimetableSlotParams{
		AssignmentID:   assignment.ID,
		DayOfWeek:      1,
		StartTime:      pgTime(7, 30),
		EndTime:        pgTime(8, 50),
		RoomLabel:      "Lab IT " + suffix,
		Notes:          "integration timetable slot",
		LessonPeriodID: period.ID,
		SlotType:       "pelajaran",
		LessonHours:    numericInt(2),
		ClassID:        class.ID,
	})
	if err != nil {
		t.Fatalf("create rombel timetable slot: %v", err)
	}
	if slot.AssignmentID != assignment.ID || slot.ClassID != class.ID || slot.PeriodNumber != 1 || slot.LessonHours != 2 {
		t.Fatalf("created timetable slot = %#v", slot)
	}

	gotSlot, err := q.GetRombelTimetableSlot(ctx, GetRombelTimetableSlotParams{ClassID: class.ID, ID: slot.ID})
	if err != nil {
		t.Fatalf("get rombel timetable slot: %v", err)
	}
	if gotSlot.ID != slot.ID || gotSlot.SubjectID != subject.ID || gotSlot.TeacherEmployeeID != teacherID {
		t.Fatalf("got rombel timetable slot = %#v", gotSlot)
	}

	rombelSlots, err := q.ListRombelTimetableSlots(ctx, class.ID)
	if err != nil {
		t.Fatalf("list rombel timetable slots: %v", err)
	}
	if !rombelTimetableSlotListed(rombelSlots, slot.ID) {
		t.Fatalf("created slot %v not returned by ListRombelTimetableSlots", slot.ID)
	}

	weeklySlots, err := q.ListWeeklyTimetableSlots(ctx, year.ID)
	if err != nil {
		t.Fatalf("list weekly timetable slots: %v", err)
	}
	if !weeklyTimetableSlotListed(weeklySlots, slot.ID) {
		t.Fatalf("created slot %v not returned by ListWeeklyTimetableSlots", slot.ID)
	}

	studentSlots, err := q.ListStudentTimetable(ctx, studentID)
	if err != nil {
		t.Fatalf("list student timetable: %v", err)
	}
	if !studentTimetableSlotListed(studentSlots, slot.ID) {
		t.Fatalf("created slot %v not returned by ListStudentTimetable", slot.ID)
	}

	teacherSlots, err := q.ListTeacherTimetable(ctx, teacherID)
	if err != nil {
		t.Fatalf("list teacher timetable: %v", err)
	}
	if !teacherTimetableSlotListed(teacherSlots, slot.ID) {
		t.Fatalf("created slot %v not returned by ListTeacherTimetable", slot.ID)
	}

	dependents, err := q.CountRombelSubjectAssignmentDependents(ctx, CountRombelSubjectAssignmentDependentsParams{ClassID: class.ID, ID: assignment.ID})
	if err != nil {
		t.Fatalf("count rombel subject assignment dependents: %v", err)
	}
	if dependents.TotalTimetableSlots != 1 {
		t.Fatalf("assignment timetable dependents = %d, want 1", dependents.TotalTimetableSlots)
	}

	detail, err := q.GetRombelDetail(ctx, class.ID)
	if err != nil {
		t.Fatalf("get rombel detail: %v", err)
	}
	if detail.HomeroomAssignmentID != homeroom.ID || detail.TotalSubjectAssignments != 1 || detail.TotalTimetableSlots != 1 || detail.TotalStudents != 1 {
		t.Fatalf("rombel detail aggregates = %#v", detail)
	}

	rombels, err := q.ListRombels(ctx)
	if err != nil {
		t.Fatalf("list rombels: %v", err)
	}
	if !rombelListedWithAggregates(rombels, class.ID, homeroom.ID) {
		t.Fatalf("created rombel %v with homeroom %v not returned by ListRombels", class.ID, homeroom.ID)
	}
}

func insertIntegrationEmployee(t *testing.T, ctx context.Context, tdb *integrationTestDB, nip, nama, pegawaiUID string) pgtype.UUID {
	t.Helper()

	var id pgtype.UUID
	if err := tdb.Pool.QueryRow(ctx, `
		INSERT INTO employees (nip, nama, unit_kerja, is_active, employment_type, pegawai_uid)
		VALUES ($1, $2, 'Integration Test', TRUE, 'honorer', $3)
		RETURNING id
	`, nip, nama, pegawaiUID).Scan(&id); err != nil {
		t.Fatalf("insert integration employee: %v", err)
	}
	return id
}

func employeeUID(prefix byte, suffix string) string {
	if len(suffix) > 12 {
		suffix = suffix[len(suffix)-12:]
	}
	return string([]byte{prefix}) + suffix
}

func insertIntegrationStudent(t *testing.T, ctx context.Context, tdb *integrationTestDB, nis, nama string, classID pgtype.UUID) pgtype.UUID {
	t.Helper()

	var id pgtype.UUID
	if err := tdb.Pool.QueryRow(ctx, `
		INSERT INTO students (nis, nisn, nama, gender, class_id, is_active, status)
		VALUES ($1, $2, $3, 'L', $4, TRUE, 'active')
		RETURNING id
	`, nis, "NISN-"+nis, nama, classID).Scan(&id); err != nil {
		t.Fatalf("insert integration student: %v", err)
	}
	return id
}

func pgTime(hour, minute int) pgtype.Time {
	duration := time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute
	return pgtype.Time{Microseconds: int64(duration / time.Microsecond), Valid: true}
}

func numericInt(value int64) pgtype.Numeric {
	return pgtype.Numeric{Int: big.NewInt(value), Exp: 0, Valid: true}
}

func homeroomAssignmentListed(items []ListHomeroomAssignmentsByClassRow, id pgtype.UUID) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func rombelSubjectAssignmentListed(items []ListRombelSubjectAssignmentsRow, id pgtype.UUID) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func weeklyTimetableAssignmentListed(items []ListWeeklyTimetableAssignmentsRow, id pgtype.UUID) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func rombelTimetableSlotListed(items []ListRombelTimetableSlotsRow, id pgtype.UUID) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func weeklyTimetableSlotListed(items []ListWeeklyTimetableSlotsRow, id pgtype.UUID) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func studentTimetableSlotListed(items []ListStudentTimetableRow, id pgtype.UUID) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func teacherTimetableSlotListed(items []ListTeacherTimetableRow, id pgtype.UUID) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func rombelListedWithAggregates(items []ListRombelsRow, classID, homeroomID pgtype.UUID) bool {
	for _, item := range items {
		if item.ID == classID && item.HomeroomAssignmentID == homeroomID && item.TotalSubjectAssignments == 1 && item.TotalTimetableSlots == 1 && item.TotalStudents == 1 {
			return true
		}
	}
	return false
}
