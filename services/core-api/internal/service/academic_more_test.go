package service

import (
	"context"
	"math/big"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestAcademicMoreNewAcademicWithPoolNilAndCurriculumEmpty(t *testing.T) {
	svc := NewAcademicWithPool(nil)
	if svc == nil {
		t.Fatal("NewAcademicWithPool(nil) = nil, want service")
	}
	if svc.q == nil {
		t.Fatal("NewAcademicWithPool(nil).q = nil, want query store")
	}
	if svc.tx != nil {
		t.Fatalf("NewAcademicWithPool(nil).tx = %T, want nil", svc.tx)
	}

	store := &fakeAcademicStore{}
	overview, err := (&Academic{q: store}).GetCurriculumOverview(context.Background(), pgtype.UUID{}, "VII")
	if err != nil {
		t.Fatalf("GetCurriculumOverview(no active profile) error = %v", err)
	}
	if len(overview.Profiles) != 0 || overview.ActiveProfile != nil || len(overview.Allocations) != 0 || len(overview.SummaryByLevel) != 0 || len(overview.ClassAssignments) != 0 {
		t.Fatalf("GetCurriculumOverview(no active profile) = %+v, want empty slices and nil active profile", overview)
	}
}

func TestAcademicMoreCurriculumOverviewConvertsNumbersAndActiveProfile(t *testing.T) {
	profileID := documentCycleTestUUID(62)
	classID := documentCycleTestUUID(63)
	alloc := db.ListCurriculumSubjectAllocationsRow{
		CurriculumProfileID: profileID,
		Level:               "vii",
		SubjectName:         "Matematika",
		IntraWeeklyHours:    academicMoreNumeric(425, -2),
		KokuWeeklyHours:     academicMoreNumeric(15, -1),
		TotalWeeklyHours:    academicMoreNumeric(575, -2),
		TotalAnnualHours:    207,
	}
	store := &fakeAcademicStore{
		curriculumProfiles: []db.CurriculumProfile{{ID: profileID, Name: "Kurikulum Merdeka", Status: "active"}},
		activeCurriculum:   db.CurriculumProfile{ID: profileID, Name: "Kurikulum Merdeka", Status: "active"},
		curriculumAllocs:   []db.ListCurriculumSubjectAllocationsRow{alloc},
		curriculumSummary: []db.GetCurriculumSummaryByLevelRow{
			{CurriculumProfileID: profileID, Level: "VII", IntraWeeklyHours: academicMoreNumeric(405, -1), KokuWeeklyHours: academicMoreNumeric(15, -1), TotalWeeklyHours: academicMoreNumeric(42, 0), ComplianceStatus: "sesuai"},
			{CurriculumProfileID: profileID, Level: "VIII", IntraWeeklyHours: academicMoreNumeric(39, 0), KokuWeeklyHours: academicMoreNumeric(1, 0), TotalWeeklyHours: academicMoreNumeric(40, 0), ComplianceStatus: "kurang"},
		},
		classCurricula: []db.ListClassCurriculumAssignmentsRow{{ClassID: classID, ClassName: "VII A", CurriculumProfileID: profileID}},
	}

	overview, err := (&Academic{q: store}).GetCurriculumOverview(context.Background(), pgtype.UUID{}, " vii ")
	if err != nil {
		t.Fatalf("GetCurriculumOverview(active profile) error = %v", err)
	}
	if overview.ActiveProfile == nil || overview.ActiveProfile.ID != profileID {
		t.Fatalf("ActiveProfile = %+v, want profile %v", overview.ActiveProfile, profileID)
	}
	if len(overview.Allocations) != 1 || overview.Allocations[0].IntraWeeklyHours != 4.25 || overview.Allocations[0].KokuWeeklyHours != 1.5 || overview.Allocations[0].TotalWeeklyHours != 5.75 {
		t.Fatalf("Allocations = %+v, want converted numeric weekly hours", overview.Allocations)
	}
	if len(overview.SummaryByLevel) != 2 {
		t.Fatalf("SummaryByLevel len = %d, want 2", len(overview.SummaryByLevel))
	}
	if overview.SummaryByLevel[0].IntraWeeklyHours != 40.5 || overview.SummaryByLevel[0].StatusLabel != "Sesuai KMA" {
		t.Fatalf("SummaryByLevel[0] = %+v, want converted numeric and compliant label", overview.SummaryByLevel[0])
	}
	if overview.SummaryByLevel[1].StatusLabel != "Perlu ditinjau" {
		t.Fatalf("SummaryByLevel[1].StatusLabel = %q, want Perlu ditinjau", overview.SummaryByLevel[1].StatusLabel)
	}
	if len(overview.ClassAssignments) != 1 || overview.ClassAssignments[0].ClassID != classID {
		t.Fatalf("ClassAssignments = %+v, want class %v", overview.ClassAssignments, classID)
	}

	selected, err := (&Academic{q: store}).GetCurriculumOverview(context.Background(), profileID, "IX")
	if err != nil {
		t.Fatalf("GetCurriculumOverview(explicit profile) error = %v", err)
	}
	if selected.ActiveProfile == nil || selected.ActiveProfile.ID != profileID {
		t.Fatalf("explicit ActiveProfile = %+v, want profile %v", selected.ActiveProfile, profileID)
	}
}

func TestAcademicMoreCurriculumWrappersAndNumericFallbacks(t *testing.T) {
	profileID := documentCycleTestUUID(64)
	store := &fakeAcademicStore{
		curriculumAllocs:  []db.ListCurriculumSubjectAllocationsRow{{CurriculumProfileID: profileID, Level: "IX", SubjectName: "BK", IntraWeeklyHours: pgtype.Numeric{}, KokuWeeklyHours: pgtype.Numeric{Valid: true}, TotalWeeklyHours: academicMoreNumeric(3335, -3)}},
		curriculumSummary: []db.GetCurriculumSummaryByLevelRow{{CurriculumProfileID: profileID, Level: "IX", IntraWeeklyHours: pgtype.Numeric{}, KokuWeeklyHours: academicMoreNumeric(25, -1), TotalWeeklyHours: academicMoreNumeric(405, -1), ComplianceStatus: "lebih"}},
	}
	svc := &Academic{q: store}

	allocs, err := svc.ListCurriculumAllocations(context.Background(), profileID, " ix ")
	if err != nil {
		t.Fatalf("ListCurriculumAllocations() error = %v", err)
	}
	if len(allocs) != 1 || allocs[0].IntraWeeklyHours != 0 || allocs[0].KokuWeeklyHours != 0 || allocs[0].TotalWeeklyHours != 3.34 {
		t.Fatalf("ListCurriculumAllocations() = %+v, want zero invalid numerics and rounded total", allocs)
	}

	summary, err := svc.GetCurriculumSummary(context.Background(), profileID)
	if err != nil {
		t.Fatalf("GetCurriculumSummary() error = %v", err)
	}
	if len(summary) != 1 || summary[0].IntraWeeklyHours != 0 || summary[0].KokuWeeklyHours != 2.5 || summary[0].TotalWeeklyHours != 40.5 || summary[0].StatusLabel != "Perlu ditinjau" {
		t.Fatalf("GetCurriculumSummary() = %+v, want converted numbers and review label", summary)
	}
}

func academicMoreNumeric(value int64, exp int32) pgtype.Numeric {
	return pgtype.Numeric{Int: big.NewInt(value), Exp: exp, Valid: true}
}
