package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestIntegrationAcademicYearAndSchoolClassRepository(t *testing.T) {
	t.Parallel()

	tdb := setupIntegrationTestDB(t)
	ctx := context.Background()
	q := tdb.Q

	beforeStats, err := q.GetAcademicStats(ctx)
	if err != nil {
		t.Fatalf("get initial academic stats: %v", err)
	}

	suffix := integrationSuffix()
	year, err := q.CreateAcademicYear(ctx, CreateAcademicYearParams{
		Name:      "Integration Year " + suffix,
		StartDate: pgDate(2026, time.July, 1),
		EndDate:   pgDate(2027, time.June, 30),
		IsActive:  false,
	})
	if err != nil {
		t.Fatalf("create academic year: %v", err)
	}
	if !year.ID.Valid {
		t.Fatalf("created academic year has invalid id")
	}
	if year.Name != "Integration Year "+suffix {
		t.Fatalf("created academic year name = %q", year.Name)
	}

	conflicts, err := q.CountAcademicYearNameConflicts(ctx, "integration year "+suffix)
	if err != nil {
		t.Fatalf("count academic year name conflicts: %v", err)
	}
	if conflicts != 1 {
		t.Fatalf("academic year name conflicts = %d, want 1", conflicts)
	}

	gotYear, err := q.GetAcademicYearByID(ctx, year.ID)
	if err != nil {
		t.Fatalf("get academic year by id: %v", err)
	}
	if gotYear.ID != year.ID || gotYear.Name != year.Name {
		t.Fatalf("got academic year = (%v, %q), want (%v, %q)", gotYear.ID, gotYear.Name, year.ID, year.Name)
	}

	if err := q.DeactivateAcademicYears(ctx); err != nil {
		t.Fatalf("deactivate academic years: %v", err)
	}
	activeYear, err := q.ActivateAcademicYear(ctx, year.ID)
	if err != nil {
		t.Fatalf("activate academic year: %v", err)
	}
	if !activeYear.IsActive {
		t.Fatalf("activated academic year is not active")
	}
	currentActiveYear, err := q.GetActiveAcademicYear(ctx)
	if err != nil {
		t.Fatalf("get active academic year: %v", err)
	}
	if currentActiveYear.ID != year.ID {
		t.Fatalf("active academic year id = %v, want %v", currentActiveYear.ID, year.ID)
	}

	class, err := q.CreateSchoolClass(ctx, CreateSchoolClassParams{
		AcademicYearID: year.ID,
		Code:           "IT-" + suffix,
		Name:           "Integration Class " + suffix,
		Level:          "7",
		IsActive:       true,
	})
	if err != nil {
		t.Fatalf("create school class: %v", err)
	}
	if class.AcademicYearID != year.ID || class.Code != "IT-"+suffix || class.Level != "7" {
		t.Fatalf("created school class = %#v", class)
	}

	classes, err := q.ListSchoolClasses(ctx)
	if err != nil {
		t.Fatalf("list school classes: %v", err)
	}
	if !schoolClassListed(classes, class.ID) {
		t.Fatalf("created school class %v not returned by ListSchoolClasses", class.ID)
	}

	matrixClasses, err := q.ListSubjectAssignmentMatrixClasses(ctx, year.ID)
	if err != nil {
		t.Fatalf("list subject assignment matrix classes: %v", err)
	}
	if !matrixClassListed(matrixClasses, class.ID) {
		t.Fatalf("created school class %v not returned by ListSubjectAssignmentMatrixClasses", class.ID)
	}

	afterStats, err := q.GetAcademicStats(ctx)
	if err != nil {
		t.Fatalf("get final academic stats: %v", err)
	}
	if afterStats.TotalYears != beforeStats.TotalYears+1 {
		t.Fatalf("total academic years = %d, want %d", afterStats.TotalYears, beforeStats.TotalYears+1)
	}
	if afterStats.TotalClasses != beforeStats.TotalClasses+1 {
		t.Fatalf("total active classes = %d, want %d", afterStats.TotalClasses, beforeStats.TotalClasses+1)
	}
}

func TestIntegrationSubjectRepository(t *testing.T) {
	t.Parallel()

	tdb := setupIntegrationTestDB(t)
	ctx := context.Background()
	q := tdb.Q

	beforeStats, err := q.GetAcademicStats(ctx)
	if err != nil {
		t.Fatalf("get initial academic stats: %v", err)
	}

	suffix := integrationSuffix()
	subject, err := q.CreateSubject(ctx, CreateSubjectParams{
		Code:                "IT-SUB-" + suffix,
		Name:                "Integration Subject " + suffix,
		Category:            "umum",
		IsAssessmentSubject: true,
		IsReportSubject:     true,
		IsScheduleActivity:  false,
		CountsForRanking:    true,
		IsLocalContent:      false,
		IsChoiceSubject:     false,
		DefaultWeeklyHours:  2,
		DisplayOrder:        9999,
		IsActive:            true,
	})
	if err != nil {
		t.Fatalf("create subject: %v", err)
	}
	if !subject.ID.Valid {
		t.Fatalf("created subject has invalid id")
	}

	gotSubject, err := q.GetSubject(ctx, subject.ID)
	if err != nil {
		t.Fatalf("get subject: %v", err)
	}
	if gotSubject.ID != subject.ID || gotSubject.Code != subject.Code || gotSubject.Name != subject.Name {
		t.Fatalf("got subject = %#v, want id %v code %q name %q", gotSubject, subject.ID, subject.Code, subject.Name)
	}
	if !gotSubject.IsAssessmentSubject || !gotSubject.IsReportSubject || !gotSubject.CountsForRanking {
		t.Fatalf("got subject flags not preserved: %#v", gotSubject)
	}

	conflicts, err := q.CountSubjectCodeConflicts(ctx, CountSubjectCodeConflictsParams{
		ID:   nonMatchingUUID(),
		Code: subject.Code,
	})
	if err != nil {
		t.Fatalf("count subject code conflicts: %v", err)
	}
	if conflicts != 1 {
		t.Fatalf("subject code conflicts = %d, want 1", conflicts)
	}

	subjects, err := q.ListSubjects(ctx)
	if err != nil {
		t.Fatalf("list subjects: %v", err)
	}
	if !subjectListed(subjects, subject.ID) {
		t.Fatalf("created subject %v not returned by ListSubjects", subject.ID)
	}

	matrixSubjects, err := q.ListSubjectAssignmentMatrixSubjects(ctx)
	if err != nil {
		t.Fatalf("list subject assignment matrix subjects: %v", err)
	}
	if !matrixSubjectListed(matrixSubjects, subject.ID) {
		t.Fatalf("created subject %v not returned by ListSubjectAssignmentMatrixSubjects", subject.ID)
	}

	afterStats, err := q.GetAcademicStats(ctx)
	if err != nil {
		t.Fatalf("get final academic stats: %v", err)
	}
	if afterStats.TotalSubjects != beforeStats.TotalSubjects+1 {
		t.Fatalf("total active subjects = %d, want %d", afterStats.TotalSubjects, beforeStats.TotalSubjects+1)
	}
}

func pgDate(year int, month time.Month, day int) pgtype.Date {
	return pgtype.Date{Time: time.Date(year, month, day, 0, 0, 0, 0, time.UTC), Valid: true}
}

func integrationSuffix() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func nonMatchingUUID() pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte{0x42}, Valid: true}
}

func schoolClassListed(classes []ListSchoolClassesRow, id pgtype.UUID) bool {
	for _, class := range classes {
		if class.ID == id {
			return true
		}
	}
	return false
}

func matrixClassListed(classes []ListSubjectAssignmentMatrixClassesRow, id pgtype.UUID) bool {
	for _, class := range classes {
		if class.ID == id {
			return true
		}
	}
	return false
}

func subjectListed(subjects []ListSubjectsRow, id pgtype.UUID) bool {
	for _, subject := range subjects {
		if subject.ID == id {
			return true
		}
	}
	return false
}

func matrixSubjectListed(subjects []ListSubjectAssignmentMatrixSubjectsRow, id pgtype.UUID) bool {
	for _, subject := range subjects {
		if subject.ID == id {
			return true
		}
	}
	return false
}
