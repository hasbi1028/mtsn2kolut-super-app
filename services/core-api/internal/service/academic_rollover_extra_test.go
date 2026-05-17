package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func academicExtraDate(year int, month time.Month, day int) pgtype.Date {
	return pgtype.Date{Time: time.Date(year, month, day, 0, 0, 0, 0, time.UTC), Valid: true}
}

func TestNormalizeAcademicYearParamsExtraValidation(t *testing.T) {
	valid, err := normalizeAcademicYearParams(db.CreateAcademicYearParams{
		Name:      " 2026/2027 ",
		StartDate: academicExtraDate(2026, time.July, 1),
		EndDate:   academicExtraDate(2027, time.June, 30),
	})
	if err != nil {
		t.Fatalf("normalizeAcademicYearParams(valid) error = %v", err)
	}
	if valid.Name != "2026/2027" || valid.IsActive {
		t.Fatalf("normalizeAcademicYearParams(valid) = %+v, want trimmed non-active year", valid)
	}

	cases := []struct {
		name string
		arg  db.CreateAcademicYearParams
		want string
	}{
		{"bad format", db.CreateAcademicYearParams{Name: "2026-2027", StartDate: academicExtraDate(2026, time.July, 1), EndDate: academicExtraDate(2027, time.June, 30)}, "format tahun ajaran"},
		{"non sequential", db.CreateAcademicYearParams{Name: "2026/2028", StartDate: academicExtraDate(2026, time.July, 1), EndDate: academicExtraDate(2027, time.June, 30)}, "berurutan"},
		{"missing date", db.CreateAcademicYearParams{Name: "2026/2027", StartDate: academicExtraDate(2026, time.July, 1)}, "tanggal tahun ajaran wajib"},
		{"end before start", db.CreateAcademicYearParams{Name: "2026/2027", StartDate: academicExtraDate(2027, time.June, 30), EndDate: academicExtraDate(2026, time.July, 1)}, "tanggal selesai"},
		{"active create", db.CreateAcademicYearParams{Name: "2026/2027", StartDate: academicExtraDate(2026, time.July, 1), EndDate: academicExtraDate(2027, time.June, 30), IsActive: true}, "nonaktif"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := normalizeAcademicYearParams(tc.arg)
			if !errors.Is(err, domain.ErrBadRequest) || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("normalizeAcademicYearParams(%s) error = %v, want ErrBadRequest containing %q", tc.name, err, tc.want)
			}
		})
	}
}

func TestResolveYearRolloverYearsUsesActiveSourceAndRejectsSameYear(t *testing.T) {
	sourceID := documentCycleTestUUID(101)
	targetID := documentCycleTestUUID(102)
	store := &fakeAcademicStore{
		activeYear: db.AcademicYear{ID: sourceID, Name: "2025/2026", IsActive: true},
		yearByID: map[string]db.AcademicYear{
			pgUUIDString(targetID): {ID: targetID, Name: "2026/2027"},
		},
	}
	source, target, err := resolveYearRolloverYears(context.Background(), store, pgtype.UUID{}, targetID)
	if err != nil {
		t.Fatalf("resolveYearRolloverYears(active source) error = %v", err)
	}
	if source.ID != sourceID || target.ID != targetID {
		t.Fatalf("resolveYearRolloverYears() source=%+v target=%+v, want active source and target", source, target)
	}

	store.activeYear = db.AcademicYear{ID: targetID, Name: "2026/2027", IsActive: true}
	_, _, err = resolveYearRolloverYears(context.Background(), store, pgtype.UUID{}, targetID)
	if !errors.Is(err, domain.ErrBadRequest) || !strings.Contains(err.Error(), "harus berbeda") {
		t.Fatalf("resolveYearRolloverYears(same source target) error = %v, want different years bad request", err)
	}
}

func TestResolveYearRolloverYearsMapsMissingSourceAndTargetToBadRequest(t *testing.T) {
	targetID := documentCycleTestUUID(103)
	store := &fakeAcademicStore{yearByID: map[string]db.AcademicYear{}}
	_, _, err := resolveYearRolloverYears(context.Background(), store, pgtype.UUID{}, targetID)
	if !errors.Is(err, domain.ErrBadRequest) || !strings.Contains(err.Error(), "tahun ajaran tujuan tidak ditemukan") {
		t.Fatalf("resolveYearRolloverYears(missing target) error = %v, want target bad request", err)
	}

	store = &fakeAcademicStore{yearByID: map[string]db.AcademicYear{pgUUIDString(targetID): {ID: targetID, Name: "2026/2027"}}, activeYearErr: pgx.ErrNoRows}
	_, _, err = resolveYearRolloverYears(context.Background(), store, pgtype.UUID{}, targetID)
	if !errors.Is(err, domain.ErrBadRequest) || !strings.Contains(err.Error(), "tahun ajaran sumber tidak ditemukan") {
		t.Fatalf("resolveYearRolloverYears(missing source) error = %v, want source bad request", err)
	}
}

func TestImportDryRunReferencesIndexesOnlyActiveSubjectsAndActiveYearClasses(t *testing.T) {
	activeYearID := documentCycleTestUUID(110)
	oldYearID := documentCycleTestUUID(111)
	activeClassID := documentCycleTestUUID(112)
	inactiveClassID := documentCycleTestUUID(113)
	oldClassID := documentCycleTestUUID(114)
	subjectID := documentCycleTestUUID(115)
	inactiveSubjectID := documentCycleTestUUID(116)
	studentID := documentCycleTestUUID(117)
	teacherID := documentCycleTestUUID(118)
	assignmentID := documentCycleTestUUID(119)
	start, _ := ParseAcademicTimeInput("07:00")
	end, _ := ParseAcademicTimeInput("08:20")
	store := &fakeAcademicStore{
		years: []db.AcademicYear{
			{ID: oldYearID, Name: "2025/2026"},
			{ID: activeYearID, Name: "2026/2027", IsActive: true},
		},
		classes: []db.ListSchoolClassesRow{
			{ID: activeClassID, AcademicYearID: activeYearID, Code: "VII-A", Name: "VII A", IsActive: true},
			{ID: inactiveClassID, AcademicYearID: activeYearID, Code: "VII-B", Name: "VII B", IsActive: false},
			{ID: oldClassID, AcademicYearID: oldYearID, Code: "VII-OLD", Name: "VII Old", IsActive: true},
		},
		subjects:       []db.ListSubjectsRow{{ID: subjectID, Code: "MTK", Name: "Matematika", IsActive: true}, {ID: inactiveSubjectID, Code: "IPA", Name: "IPA", IsActive: false}},
		importStudents: []db.ListAcademicImportStudentsRow{{ID: studentID, Nis: "1001", Nama: "Alya"}},
		importTeachers: []db.ListAcademicImportTeachersRow{{ID: teacherID, Nip: "1970", Nama: "Guru MTK", IsActive: true}},
		assignments:    []db.ListClassSubjectAssignmentsRow{{ID: assignmentID, ClassID: activeClassID, SubjectID: subjectID, TeacherEmployeeID: teacherID, ClassCode: "VII-A", SubjectCode: "MTK"}},
		timetableSlots: []db.ListTimetableSlotsRow{{ID: documentCycleTestUUID(120), AssignmentID: assignmentID, DayOfWeek: 1, StartTime: start, EndTime: end}},
	}
	svc := &Academic{q: store}
	refs, err := svc.importDryRunReferences(context.Background())
	if err != nil {
		t.Fatalf("importDryRunReferences() error = %v", err)
	}
	if refs.activeYearID != pgUUIDString(activeYearID) || refs.yearsByName["2026/2027"].ID != activeYearID {
		t.Fatalf("refs activeYearID=%q years=%+v, want active year indexed", refs.activeYearID, refs.yearsByName)
	}
	if refs.activeClassesByCode["vii-a"].ID != activeClassID {
		t.Fatalf("activeClassesByCode = %+v, want active class for active year", refs.activeClassesByCode)
	}
	if _, ok := refs.activeClassesByCode["vii-b"]; ok {
		t.Fatalf("activeClassesByCode includes inactive active-year class: %+v", refs.activeClassesByCode["vii-b"])
	}
	if _, ok := refs.activeClassesByCode["vii-old"]; ok {
		t.Fatalf("activeClassesByCode includes class from inactive year: %+v", refs.activeClassesByCode["vii-old"])
	}
	if refs.subjectsByCode["mtk"].ID != subjectID {
		t.Fatalf("subjectsByCode = %+v, want active subject", refs.subjectsByCode)
	}
	if _, ok := refs.subjectsByCode["ipa"]; ok {
		t.Fatalf("subjectsByCode includes inactive subject: %+v", refs.subjectsByCode["ipa"])
	}
	if refs.studentsByNIS["1001"].ID != studentID || refs.teachersByNIP["1970"].ID != teacherID || refs.teachersByName["guru mtk"].ID != teacherID {
		t.Fatalf("student/teacher indexes not populated: students=%+v teachersNIP=%+v teachersName=%+v", refs.studentsByNIS, refs.teachersByNIP, refs.teachersByName)
	}
	if refs.assignmentsByClassSubject[importPairKey(pgUUIDString(activeClassID), pgUUIDString(subjectID))].ID != assignmentID || refs.assignmentsByClassSubjectTeacher[importTripleKey(pgUUIDString(activeClassID), pgUUIDString(subjectID), pgUUIDString(teacherID))].ID != assignmentID {
		t.Fatalf("assignment indexes not populated: classSubject=%+v classSubjectTeacher=%+v", refs.assignmentsByClassSubject, refs.assignmentsByClassSubjectTeacher)
	}
	if refs.slotsByAssignmentDayTime[importTripleKey(pgUUIDString(assignmentID), "1", "07:00")+"|08:20"].AssignmentID != assignmentID {
		t.Fatalf("slot index = %+v, want timetable slot by assignment/day/time", refs.slotsByAssignmentDayTime)
	}
}

func TestValidateAcademicImportRowDispatchesKindsAndRejectsUnknown(t *testing.T) {
	classID := documentCycleTestUUID(121)
	subjectID := documentCycleTestUUID(122)
	teacherID := documentCycleTestUUID(123)
	refs := academicImportAssignmentRefs(classID, subjectID, teacherID)
	refs.studentsByNIS = map[string]db.ListAcademicImportStudentsRow{}
	seen := map[string]int{}
	plan, rowErrs := validateAcademicImportRow("siswa", 2, []string{"1001", "Alya", "P", "VII-A"}, map[string]int{"nis": 0, "nama": 1, "jenis kelamin": 2, "kode rombel": 3}, refs, seen)
	if len(rowErrs) != 0 || plan.Action != "add" {
		t.Fatalf("validateAcademicImportRow(siswa) plan=%+v errs=%+v, want add", plan, rowErrs)
	}
	_, rowErrs = validateAcademicImportRow("tidak_ada", 3, nil, nil, refs, seen)
	if len(rowErrs) != 1 || !strings.Contains(rowErrs[0].Message, "Jenis data") {
		t.Fatalf("validateAcademicImportRow(unknown) errs=%+v, want invalid kind error", rowErrs)
	}
}
