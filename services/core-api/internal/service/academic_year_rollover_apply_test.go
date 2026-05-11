package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestAcademicApplyYearRolloverCreatesCopiesAndPromotesSafely(t *testing.T) {
	sourceYearID := documentCycleTestUUID(201)
	targetYearID := documentCycleTestUUID(202)
	sourceClassID := documentCycleTestUUID(203)
	sourceFinalClassID := documentCycleTestUUID(204)
	targetClassID := documentCycleTestUUID(205)
	subjectID := documentCycleTestUUID(206)
	teacherID := documentCycleTestUUID(207)
	sourceAssignmentID := documentCycleTestUUID(208)
	targetAssignmentID := documentCycleTestUUID(209)
	studentID := documentCycleTestUUID(210)
	finalStudentID := documentCycleTestUUID(211)
	slotID := documentCycleTestUUID(212)
	start, err := ParseAcademicTimeInput("07:30")
	if err != nil {
		t.Fatalf("ParseAcademicTimeInput(start) error = %v", err)
	}
	end, err := ParseAcademicTimeInput("08:10")
	if err != nil {
		t.Fatalf("ParseAcademicTimeInput(end) error = %v", err)
	}
	targetStart := pgtype.Date{Time: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), Valid: true}
	sourceYear := db.AcademicYear{ID: sourceYearID, Name: "2025/2026"}
	targetYear := db.AcademicYear{ID: targetYearID, Name: "2026/2027", StartDate: targetStart}
	store := &fakeAcademicStore{
		yearByID: map[string]db.AcademicYear{
			pgUUIDString(sourceYearID): sourceYear,
			pgUUIDString(targetYearID): targetYear,
		},
		activeYear: sourceYear,
		classes: []db.ListSchoolClassesRow{
			{ID: sourceClassID, AcademicYearID: sourceYearID, Code: "VII-A", Name: "VII A", Level: "VII", IsActive: true},
			{ID: sourceFinalClassID, AcademicYearID: sourceYearID, Code: "IX-A", Name: "IX A", Level: "IX", IsActive: true},
		},
		rolloverStudents: []db.ListYearRolloverStudentsRow{
			{ID: studentID, Nis: "001", Nama: "Alya", ClassID: sourceClassID, ClassCode: "VII-A", ClassLevel: "VII"},
			{ID: finalStudentID, Nis: "009", Nama: "Bima", ClassID: sourceFinalClassID, ClassCode: "IX-A", ClassLevel: "IX"},
		},
		homeroomDetails: []db.ListYearRolloverHomeroomAssignmentDetailsRow{
			{ClassID: sourceClassID, EmployeeID: teacherID, StartDate: targetStart, Notes: "wali aktif"},
		},
		assignments: []db.ListClassSubjectAssignmentsRow{
			{ID: sourceAssignmentID, ClassID: sourceClassID, SubjectID: subjectID, TeacherEmployeeID: teacherID},
		},
		timetableSlots: []db.ListTimetableSlotsRow{
			{ID: slotID, AssignmentID: sourceAssignmentID, DayOfWeek: 1, StartTime: start, EndTime: end, RoomLabel: "R1", Notes: "pagi"},
		},
		createClassResult:  db.SchoolClass{ID: targetClassID, AcademicYearID: targetYearID, Code: "VIII-A", Name: "VIII A", Level: "VIII", IsActive: true},
		createAssignResult: db.ClassSubjectAssignment{ID: targetAssignmentID, ClassID: targetClassID, SubjectID: subjectID, TeacherEmployeeID: teacherID},
	}
	svc := &Academic{q: store}

	result, err := svc.ApplyYearRollover(context.Background(), YearRolloverApplyInput{
		SourceAcademicYearID: sourceYearID,
		TargetAcademicYearID: targetYearID,
		Confirmation:         yearRolloverApplyChallenge(sourceYear.Name, targetYear.Name),
	})
	if err != nil {
		t.Fatalf("ApplyYearRollover() error = %v", err)
	}
	if result.Counts.ClassesCreated != 1 || result.Counts.ClassesReused != 0 {
		t.Fatalf("class counts = %+v, want created 1 reused 0", result.Counts)
	}
	if result.Counts.HomeroomsCopied != 1 || result.Counts.AssignmentsCopied != 1 || result.Counts.TimetableSlotsCopied != 1 {
		t.Fatalf("copy counts = %+v, want homeroom/assignment/slot copied", result.Counts)
	}
	if result.Counts.StudentsPromoted != 1 || result.Counts.StudentsSkipped != 1 {
		t.Fatalf("student counts = %+v, want promoted 1 skipped 1", result.Counts)
	}
	if store.createClassArg.Code != "VIII-A" || store.createClassArg.Level != "VIII" {
		t.Fatalf("created class arg = %+v, want VIII-A/VIII", store.createClassArg)
	}
	if len(store.createHomerooms) != 1 || store.createHomerooms[0].ClassID != targetClassID || store.createHomerooms[0].HomeroomAcademicYearID != targetYearID {
		t.Fatalf("created homerooms = %+v, want target class/year", store.createHomerooms)
	}
	if store.createAssignArg.ClassID != targetClassID || store.createAssignArg.SubjectID != subjectID {
		t.Fatalf("created assignment arg = %+v, want target class subject", store.createAssignArg)
	}
	if len(store.createSlotArgs) != 1 || store.createSlotArgs[0].AssignmentID != targetAssignmentID {
		t.Fatalf("created slot args = %+v, want target assignment", store.createSlotArgs)
	}
	if len(store.promoteArgs) != 1 || store.promoteArgs[0].StudentID != studentID || store.promoteArgs[0].TargetClassID != targetClassID {
		t.Fatalf("promote args = %+v, want student promoted to target class", store.promoteArgs)
	}
	if len(result.StudentsSkipped) != 1 || !strings.Contains(result.StudentsSkipped[0].Reason, "Tingkat akhir") {
		t.Fatalf("students skipped = %+v, want final level skipped", result.StudentsSkipped)
	}
}

func TestAcademicApplyYearRolloverRequiresChallengeAndReusesExistingTargets(t *testing.T) {
	sourceYearID := documentCycleTestUUID(221)
	targetYearID := documentCycleTestUUID(222)
	sourceClassID := documentCycleTestUUID(223)
	targetClassID := documentCycleTestUUID(224)
	subjectID := documentCycleTestUUID(225)
	teacherID := documentCycleTestUUID(226)
	sourceAssignmentID := documentCycleTestUUID(227)
	targetAssignmentID := documentCycleTestUUID(228)
	sourceYear := db.AcademicYear{ID: sourceYearID, Name: "2025/2026"}
	targetYear := db.AcademicYear{ID: targetYearID, Name: "2026/2027"}
	start, err := ParseAcademicTimeInput("09:00")
	if err != nil {
		t.Fatalf("ParseAcademicTimeInput(start) error = %v", err)
	}
	end, err := ParseAcademicTimeInput("09:40")
	if err != nil {
		t.Fatalf("ParseAcademicTimeInput(end) error = %v", err)
	}
	store := &fakeAcademicStore{
		yearByID: map[string]db.AcademicYear{
			pgUUIDString(sourceYearID): sourceYear,
			pgUUIDString(targetYearID): targetYear,
		},
		activeYear: sourceYear,
		classes: []db.ListSchoolClassesRow{
			{ID: sourceClassID, AcademicYearID: sourceYearID, Code: "VII-B", Name: "VII B", Level: "VII", IsActive: true},
			{ID: targetClassID, AcademicYearID: targetYearID, Code: "VIII-B", Name: "VIII B", Level: "VIII", IsActive: true},
		},
		homeroomCounts: map[string]int32{pgUUIDString(targetClassID): 1},
		homeroomDetails: []db.ListYearRolloverHomeroomAssignmentDetailsRow{
			{ClassID: sourceClassID, EmployeeID: teacherID},
		},
		assignments: []db.ListClassSubjectAssignmentsRow{
			{ID: sourceAssignmentID, ClassID: sourceClassID, SubjectID: subjectID, TeacherEmployeeID: teacherID},
			{ID: targetAssignmentID, ClassID: targetClassID, SubjectID: subjectID, TeacherEmployeeID: teacherID},
		},
		timetableSlots: []db.ListTimetableSlotsRow{
			{AssignmentID: sourceAssignmentID, DayOfWeek: 2, StartTime: start, EndTime: end, RoomLabel: "R2"},
			{AssignmentID: targetAssignmentID, DayOfWeek: 2, StartTime: start, EndTime: end, RoomLabel: "R2"},
		},
	}
	svc := &Academic{q: store}
	if _, err := svc.ApplyYearRollover(context.Background(), YearRolloverApplyInput{TargetAcademicYearID: targetYearID, Confirmation: "SALAH"}); err == nil || !strings.Contains(err.Error(), "kalimat konfirmasi") {
		t.Fatalf("ApplyYearRollover(bad challenge) = %v, want confirmation error", err)
	}

	result, err := svc.ApplyYearRollover(context.Background(), YearRolloverApplyInput{
		TargetAcademicYearID: targetYearID,
		Confirmation:         yearRolloverApplyChallenge(sourceYear.Name, targetYear.Name),
	})
	if err != nil {
		t.Fatalf("ApplyYearRollover(reuse) error = %v", err)
	}
	if result.Counts.ClassesReused != 1 || result.Counts.ClassesCreated != 0 {
		t.Fatalf("class counts = %+v, want reused 1 created 0", result.Counts)
	}
	if result.Counts.HomeroomsCopied != 0 || result.Counts.AssignmentsCopied != 0 || result.Counts.TimetableSlotsCopied != 0 {
		t.Fatalf("copy counts = %+v, want no duplicates copied", result.Counts)
	}
	if len(store.createHomerooms) != 0 || len(store.createSlotArgs) != 0 || store.createAssignArg.ClassID.Valid {
		t.Fatalf("created homeroom/slot/assignment = %+v/%+v/%+v, want no duplicate creates", store.createHomerooms, store.createSlotArgs, store.createAssignArg)
	}
}
