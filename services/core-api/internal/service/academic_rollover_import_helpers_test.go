package service

import (
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestAcademicImportHeaderNormalizesColumnsAndReportsMissingRequired(t *testing.T) {
	header, err := academicImportHeader("siswa", []string{" NIS ", "NISN", "Nama", "Jenis Kelamin", "Kode Rombel", "Status", "Catatan"})
	if err != nil {
		t.Fatalf("academicImportHeader(valid siswa) error = %v", err)
	}
	if header["nis"] != 0 || header["jenis kelamin"] != 3 || header["kode rombel"] != 4 {
		t.Fatalf("academicImportHeader() = %+v, want normalized header positions", header)
	}

	_, err = academicImportHeader("jadwal", []string{"Kode Rombel", "Kode Mapel", "Hari", "Jam Mulai"})
	if !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("academicImportHeader(missing jadwal fields) error = %v, want ErrBadRequest", err)
	}
	if err == nil || !stringContainsAll(err.Error(), []string{"nip guru", "nama guru", "jam selesai", "ruang", "catatan"}) {
		t.Fatalf("academicImportHeader(missing jadwal fields) error = %v, want missing field list", err)
	}
}

func TestNormalizeImportKindAcceptsAliasesAndRejectsUnknown(t *testing.T) {
	cases := map[string]string{
		" Students ":          "siswa",
		"kelas":               "rombel",
		"ASSIGNMENTS":         "guru_mapel",
		"jadwal pelajaran":    "jadwal",
		"unknown import kind": "",
	}
	for input, want := range cases {
		if got := normalizeImportKind(input); got != want {
			t.Fatalf("normalizeImportKind(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestPlannedRolloverClassBuildsActiveTargetClassRow(t *testing.T) {
	targetID := documentCycleTestUUID(30)
	got := plannedRolloverClass(db.AcademicYear{ID: targetID, Name: "2027/2028"}, "VIII-A", "VIII A", "VIII")
	if got.Code != "VIII-A" || got.Name != "VIII A" || got.Level != "VIII" || !got.IsActive || got.AcademicYearID != targetID || got.AcademicYearName != "2027/2028" {
		t.Fatalf("plannedRolloverClass() = %+v, want active target class row", got)
	}
	if got.ID != (pgtype.UUID{}) {
		t.Fatalf("plannedRolloverClass() ID = %v, want zero UUID for planned/uncreated class", got.ID)
	}
}

func TestImportKeysCellsAndRequiredCellsNormalizeAndTrim(t *testing.T) {
	if got := importTripleKey(" Class A ", "MTK", " Teacher "); got != "class a|mtk|teacher" {
		t.Fatalf("importTripleKey() = %q, want normalized triple key", got)
	}
	row := []string{"  VII-A ", " MTK "}
	header := map[string]int{"kode rombel": 0, "kode mapel": 1, "nip guru": 4}
	if got := cell(row, header, "Kode Rombel"); got != "VII-A" {
		t.Fatalf("cell(existing) = %q, want trimmed VII-A", got)
	}
	if got := cell(row, header, "NIP Guru"); got != "" {
		t.Fatalf("cell(out of range) = %q, want empty string", got)
	}
	if got := cell(row, header, "Nama Guru"); got != "" {
		t.Fatalf("cell(missing header) = %q, want empty string", got)
	}

	errs := []AcademicImportRowError{}
	requireCell(&errs, 7, "NIS", "  ")
	requireCell(&errs, 7, "Nama", "Siswa")
	if len(errs) != 1 || errs[0].Row != 7 || errs[0].Field != "NIS" || errs[0].Message != "NIS wajib diisi" {
		t.Fatalf("requireCell() errors = %+v, want one required NIS error", errs)
	}
}

func TestAcademicImportPrimitiveHelpers(t *testing.T) {
	if !csvRowBlank([]string{" ", "\t", ""}) {
		t.Fatalf("csvRowBlank(all whitespace) = false, want true")
	}
	if csvRowBlank([]string{" ", "NIS"}) {
		t.Fatalf("csvRowBlank(non-empty cell) = true, want false")
	}

	genderCases := map[string]string{
		" L ":         "male",
		"laki laki":   "male",
		"Laki-laki":   "male",
		"MALE":        "male",
		"p":           "female",
		"Perempuan":   "female",
		"female":      "female",
		"tidak valid": "",
	}
	for input, want := range genderCases {
		if got := normalizeGender(input); got != want {
			t.Fatalf("normalizeGender(%q) = %q, want %q", input, got, want)
		}
	}

	dayCases := map[string]int{" Senin ": 1, "2": 2, "RABU": 3, "kamis": 4, "Jum'at": 5, "jumat": 5, "Sabtu": 6}
	for input, want := range dayCases {
		got, ok := parseAcademicDay(input)
		if !ok || got != want {
			t.Fatalf("parseAcademicDay(%q) = (%d, %v), want (%d, true)", input, got, ok, want)
		}
	}
	if got, ok := parseAcademicDay("minggu"); ok || got != 0 {
		t.Fatalf("parseAcademicDay(invalid) = (%d, %v), want (0, false)", got, ok)
	}
}

func TestAcademicImportTimeHelpers(t *testing.T) {
	if got := normalizeImportTime(" 07:05 "); got != "07:05" {
		t.Fatalf("normalizeImportTime(valid) = %q, want 07:05", got)
	}
	if got := normalizeImportTime("bad time"); got != "bad time" {
		t.Fatalf("normalizeImportTime(invalid) = %q, want trimmed original", got)
	}
	if got := formatAcademicTime(pgtype.Time{}); got != "" {
		t.Fatalf("formatAcademicTime(invalid) = %q, want empty", got)
	}
	if got := formatAcademicTime(pgtype.Time{Microseconds: (9*3600 + 30*60 + 59) * 1_000_000, Valid: true}); got != "09:30" {
		t.Fatalf("formatAcademicTime(valid) = %q, want 09:30", got)
	}
}

func TestValidateStudentImportRowPlansAddUpdateAndDuplicate(t *testing.T) {
	refs := importDryRunReferences{
		activeClassesByCode: map[string]db.ListSchoolClassesRow{"vii-a": {ID: documentCycleTestUUID(41), Code: "VII-A", IsActive: true}},
		studentsByNIS:       map[string]db.ListAcademicImportStudentsRow{"1001": {ID: documentCycleTestUUID(42), Nis: "1001", Nama: "Existing"}},
	}
	header := map[string]int{"nis": 0, "nama": 1, "jenis kelamin": 2, "kode rombel": 3}
	seen := map[string]int{}

	plan, errs := validateStudentImportRow(2, []string{"1001", "Ali", "L", "VII-A"}, header, refs, seen)
	if len(errs) != 0 || plan.Action != "update" || !strings.Contains(plan.Summary, "Perbarui siswa Ali") {
		t.Fatalf("validateStudentImportRow(existing) plan = %+v errs = %+v, want update", plan, errs)
	}
	plan, errs = validateStudentImportRow(3, []string{"1002", "Budi", "Perempuan", "VII-A"}, header, refs, seen)
	if len(errs) != 0 || plan.Action != "add" || !strings.Contains(plan.Summary, "Tambah siswa Budi") {
		t.Fatalf("validateStudentImportRow(new) plan = %+v errs = %+v, want add", plan, errs)
	}
	_, errs = validateStudentImportRow(4, []string{"1002", "Budi", "X", "MISSING"}, header, refs, seen)
	if !academicImportErrorsContain(errs, "Jenis Kelamin") || !academicImportErrorsContain(errs, "Kode Rombel") || !academicImportErrorsContainMessage(errs, "duplikat dengan baris 3") {
		t.Fatalf("validateStudentImportRow(invalid duplicate) errs = %+v, want gender/class/duplicate errors", errs)
	}
}

func TestValidateClassImportRowPlansAddUpdateAndErrors(t *testing.T) {
	yearID := documentCycleTestUUID(43)
	refs := importDryRunReferences{
		yearsByName:       map[string]db.AcademicYear{"2026/2027": {ID: yearID, Name: "2026/2027"}},
		classesByYearCode: map[string]db.ListSchoolClassesRow{importPairKey(pgUUIDString(yearID), "VII-A"): {ID: documentCycleTestUUID(44), Code: "VII-A"}},
	}
	header := map[string]int{"kode rombel": 0, "nama rombel": 1, "tingkat": 2, "tahun ajaran": 3}
	seen := map[string]int{}

	plan, errs := validateClassImportRow(2, []string{"VII-A", "VII A", "7", "2026/2027"}, header, refs, seen)
	if len(errs) != 0 || plan.Action != "update" {
		t.Fatalf("validateClassImportRow(existing) plan = %+v errs = %+v, want update", plan, errs)
	}
	plan, errs = validateClassImportRow(3, []string{"VII-B", "VII B", "VII", "2026/2027"}, header, refs, seen)
	if len(errs) != 0 || plan.Action != "add" {
		t.Fatalf("validateClassImportRow(new) plan = %+v errs = %+v, want add", plan, errs)
	}
	_, errs = validateClassImportRow(4, []string{"VII-B", "", "X", "2026/2027"}, header, refs, seen)
	if !academicImportErrorsContain(errs, "Nama Rombel") || !academicImportErrorsContain(errs, "Tingkat") || !academicImportErrorsContainMessage(errs, "duplikat dengan baris 3") {
		t.Fatalf("validateClassImportRow(invalid duplicate) errs = %+v, want required/level/duplicate errors", errs)
	}
}

func TestValidateSubjectTeacherImportRowPlansAddSkipUpdateAndDuplicate(t *testing.T) {
	classID := documentCycleTestUUID(45)
	subjectID := documentCycleTestUUID(46)
	teacherID := documentCycleTestUUID(47)
	otherTeacherID := documentCycleTestUUID(48)
	refs := academicImportAssignmentRefs(classID, subjectID, teacherID)
	header := map[string]int{"kode rombel": 0, "kode mapel": 1, "nip guru": 2, "nama guru": 3}
	seen := map[string]int{}

	plan, errs := validateSubjectTeacherImportRow(2, []string{"VII-A", "MTK", "1970", ""}, header, refs, seen)
	if len(errs) != 0 || plan.Action != "skip" {
		t.Fatalf("validateSubjectTeacherImportRow(same teacher) plan = %+v errs = %+v, want skip", plan, errs)
	}
	refs.teachersByNIP["1971"] = db.ListAcademicImportTeachersRow{ID: otherTeacherID, Nip: "1971", Nama: "Other", IsActive: true}
	_, errs = validateSubjectTeacherImportRow(3, []string{"VII-A", "MTK", "1971", ""}, header, refs, seen)
	if !academicImportErrorsContainMessage(errs, "duplikat dengan baris 2") {
		t.Fatalf("validateSubjectTeacherImportRow(duplicate) errs = %+v, want duplicate error", errs)
	}

	seen = map[string]int{}
	plan, errs = validateSubjectTeacherImportRow(4, []string{"VII-A", "MTK", "1971", ""}, header, refs, seen)
	if len(errs) != 0 || plan.Action != "update" {
		t.Fatalf("validateSubjectTeacherImportRow(other teacher) plan = %+v errs = %+v, want update", plan, errs)
	}
	refs.assignmentsByClassSubject = map[string]db.ListClassSubjectAssignmentsRow{}
	plan, errs = validateSubjectTeacherImportRow(5, []string{"VII-A", "MTK", "1970", ""}, header, refs, map[string]int{})
	if len(errs) != 0 || plan.Action != "add" {
		t.Fatalf("validateSubjectTeacherImportRow(no assignment) plan = %+v errs = %+v, want add", plan, errs)
	}
}

func TestValidateTimetableImportRowPlansAddUpdateAndErrors(t *testing.T) {
	classID := documentCycleTestUUID(49)
	subjectID := documentCycleTestUUID(50)
	teacherID := documentCycleTestUUID(51)
	assignmentID := documentCycleTestUUID(52)
	refs := academicImportAssignmentRefs(classID, subjectID, teacherID)
	assignment := db.ListClassSubjectAssignmentsRow{ID: assignmentID, ClassID: classID, SubjectID: subjectID, TeacherEmployeeID: teacherID, ClassCode: "VII-A", SubjectCode: "MTK"}
	refs.assignmentsByClassSubjectTeacher[importTripleKey(pgUUIDString(classID), pgUUIDString(subjectID), pgUUIDString(teacherID))] = assignment
	refs.slotsByAssignmentDayTime[importTripleKey(pgUUIDString(assignmentID), "1", "07:00")+"|"+"08:20"] = db.ListTimetableSlotsRow{ID: documentCycleTestUUID(53)}
	header := map[string]int{"kode rombel": 0, "kode mapel": 1, "nip guru": 2, "nama guru": 3, "hari": 4, "jam mulai": 5, "jam selesai": 6}
	seen := map[string]int{}

	plan, errs := validateTimetableImportRow(2, []string{"VII-A", "MTK", "1970", "", "Senin", "07:00", "08:20"}, header, refs, seen)
	if len(errs) != 0 || plan.Action != "update" {
		t.Fatalf("validateTimetableImportRow(existing slot) plan = %+v errs = %+v, want update", plan, errs)
	}
	plan, errs = validateTimetableImportRow(3, []string{"VII-A", "MTK", "1970", "", "2", "09:00", "10:20"}, header, refs, seen)
	if len(errs) != 0 || plan.Action != "add" {
		t.Fatalf("validateTimetableImportRow(new slot) plan = %+v errs = %+v, want add", plan, errs)
	}
	_, errs = validateTimetableImportRow(4, []string{"VII-A", "MTK", "1970", "", "2", "09:00", "10:20"}, header, refs, seen)
	if !academicImportErrorsContainMessage(errs, "duplikat dengan baris 3") {
		t.Fatalf("validateTimetableImportRow(duplicate) errs = %+v, want duplicate error", errs)
	}
	_, errs = validateTimetableImportRow(5, []string{"VII-A", "MTK", "1970", "", "2", "09:00", "09:00"}, header, refs, seen)
	if !academicImportErrorsContain(errs, "Jam Selesai") {
		t.Fatalf("validateTimetableImportRow(invalid end) errs = %+v, want end error", errs)
	}
	refs.assignmentsByClassSubjectTeacher = map[string]db.ListClassSubjectAssignmentsRow{}
	_, errs = validateTimetableImportRow(6, []string{"VII-A", "MTK", "1970", "", "Minggu", "bad", "08:00"}, header, refs, map[string]int{})
	if !academicImportErrorsContain(errs, "Hari") || !academicImportErrorsContain(errs, "Jam Mulai") {
		t.Fatalf("validateTimetableImportRow(invalid day/time) errs = %+v, want day/start errors", errs)
	}
}

func academicImportAssignmentRefs(classID, subjectID, teacherID pgtype.UUID) importDryRunReferences {
	class := db.ListSchoolClassesRow{ID: classID, Code: "VII-A", Name: "VII A", IsActive: true}
	subject := db.ListSubjectsRow{ID: subjectID, Code: "MTK", Name: "Matematika", IsActive: true}
	teacher := db.ListAcademicImportTeachersRow{ID: teacherID, Nip: "1970", Nama: "Guru MTK", IsActive: true}
	assignment := db.ListClassSubjectAssignmentsRow{ID: documentCycleTestUUID(60), ClassID: classID, SubjectID: subjectID, TeacherEmployeeID: teacherID, ClassCode: "VII-A", SubjectCode: "MTK"}
	return importDryRunReferences{
		activeClassesByCode:              map[string]db.ListSchoolClassesRow{"vii-a": class},
		subjectsByCode:                   map[string]db.ListSubjectsRow{"mtk": subject},
		teachersByNIP:                    map[string]db.ListAcademicImportTeachersRow{"1970": teacher},
		teachersByName:                   map[string]db.ListAcademicImportTeachersRow{"guru mtk": teacher},
		assignmentsByClassSubject:        map[string]db.ListClassSubjectAssignmentsRow{importPairKey(pgUUIDString(classID), pgUUIDString(subjectID)): assignment},
		assignmentsByClassSubjectTeacher: map[string]db.ListClassSubjectAssignmentsRow{importTripleKey(pgUUIDString(classID), pgUUIDString(subjectID), pgUUIDString(teacherID)): assignment},
		slotsByAssignmentDayTime:         map[string]db.ListTimetableSlotsRow{},
	}
}

func academicImportErrorsContain(errs []AcademicImportRowError, field string) bool {
	for _, err := range errs {
		if err.Field == field {
			return true
		}
	}
	return false
}

func academicImportErrorsContainMessage(errs []AcademicImportRowError, text string) bool {
	for _, err := range errs {
		if strings.Contains(err.Message, text) {
			return true
		}
	}
	return false
}

func stringContainsAll(value string, needles []string) bool {
	for _, needle := range needles {
		if !strings.Contains(value, needle) {
			return false
		}
	}
	return true
}
