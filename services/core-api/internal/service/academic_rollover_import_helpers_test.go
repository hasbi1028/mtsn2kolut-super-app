package service

import (
	"context"
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

func TestAcademicActivateYearRequiresConfirmationAndSkipsAlreadyActive(t *testing.T) {
	yearID := documentCycleTestUUID(31)
	activeYear := db.AcademicYear{ID: yearID, Name: "2026/2027", IsActive: true}
	svc := &Academic{q: &fakeAcademicStore{yearByID: map[string]db.AcademicYear{pgUUIDString(yearID): activeYear}}}
	if _, err := svc.ActivateYear(context.Background(), yearID, "salah"); !errors.Is(err, domain.ErrBadRequest) || !strings.Contains(err.Error(), "konfirmasi aktivasi") {
		t.Fatalf("ActivateYear(bad confirmation) = %v, want bad request confirmation error", err)
	}

	store := &fakeAcademicStore{yearByID: map[string]db.AcademicYear{pgUUIDString(yearID): activeYear}}
	svc = &Academic{q: store}
	got, err := svc.ActivateYear(context.Background(), yearID, " aktifkan ")
	if err != nil {
		t.Fatalf("ActivateYear(already active) error = %v", err)
	}
	if got != activeYear {
		t.Fatalf("ActivateYear(already active) = %+v, want current row", got)
	}
	if store.deactivateCalled || store.activateYearID.Valid {
		t.Fatalf("ActivateYear(already active) deactivate=%v activateID=%v, want no writes", store.deactivateCalled, store.activateYearID)
	}
}

func TestAcademicActivateYearDeactivatesThenActivatesInactiveYear(t *testing.T) {
	yearID := documentCycleTestUUID(32)
	store := &fakeAcademicStore{yearByID: map[string]db.AcademicYear{pgUUIDString(yearID): {ID: yearID, Name: "2026/2027"}}}
	svc := &Academic{q: store}
	got, err := svc.ActivateYear(context.Background(), yearID, academicYearActivationChallenge)
	if err != nil {
		t.Fatalf("ActivateYear() error = %v", err)
	}
	if !got.IsActive || got.ID != yearID {
		t.Fatalf("ActivateYear() = %+v, want activated target year", got)
	}
	if !store.deactivateCalled || store.activateYearID != yearID {
		t.Fatalf("ActivateYear() deactivate=%v activateID=%v, want deactivate then activate target", store.deactivateCalled, store.activateYearID)
	}
}

func TestAcademicPreviewYearRolloverCountsCreatesReusesAndWarnings(t *testing.T) {
	sourceYearID := documentCycleTestUUID(33)
	targetYearID := documentCycleTestUUID(34)
	sourceVIIID := documentCycleTestUUID(35)
	sourceVIIIID := documentCycleTestUUID(36)
	sourceIXID := documentCycleTestUUID(37)
	targetVIIIID := documentCycleTestUUID(38)
	targetIXInactiveID := documentCycleTestUUID(39)
	assignmentID := documentCycleTestUUID(40)
	subjectID := documentCycleTestUUID(41)
	start, _ := ParseAcademicTimeInput("07:00")
	end, _ := ParseAcademicTimeInput("07:40")
	store := &fakeAcademicStore{
		yearByID: map[string]db.AcademicYear{
			pgUUIDString(sourceYearID): {ID: sourceYearID, Name: "2025/2026", IsActive: true},
			pgUUIDString(targetYearID): {ID: targetYearID, Name: "2026/2027"},
		},
		classes: []db.ListSchoolClassesRow{
			{ID: sourceVIIID, AcademicYearID: sourceYearID, Code: "VII-A", Name: "VII A", Level: "VII", IsActive: true},
			{ID: sourceVIIIID, AcademicYearID: sourceYearID, Code: "VIII-A", Name: "VIII A", Level: "VIII", IsActive: true},
			{ID: sourceIXID, AcademicYearID: sourceYearID, Code: "IX-A", Name: "IX A", Level: "IX", IsActive: true},
			{ID: targetVIIIID, AcademicYearID: targetYearID, Code: "VIII-A", Name: "VIII A", Level: "VIII", IsActive: true},
			{ID: targetIXInactiveID, AcademicYearID: targetYearID, Code: "IX-A", Name: "IX A", Level: "IX", IsActive: false},
		},
		rolloverStudents: []db.ListYearRolloverStudentsRow{
			{ID: documentCycleTestUUID(42), Nis: "001", Nama: "Alya", ClassID: sourceVIIID, ClassCode: "VII-A", ClassLevel: "VII"},
			{ID: documentCycleTestUUID(43), Nis: "002", Nama: "Bimo", ClassID: sourceVIIIID, ClassCode: "VIII-A", ClassLevel: "VIII"},
			{ID: documentCycleTestUUID(44), Nis: "003", Nama: "Cici", ClassID: sourceIXID, ClassCode: "IX-A", ClassLevel: "IX"},
		},
		rolloverHomerooms: []db.ListYearRolloverHomeroomAssignmentsRow{{ClassID: sourceVIIID, Total: 1}, {ClassID: sourceVIIIID, Total: 1}},
		assignments:       []db.ListClassSubjectAssignmentsRow{{ID: assignmentID, ClassID: sourceVIIID, SubjectID: subjectID}, {ID: documentCycleTestUUID(45), ClassID: sourceVIIIID, SubjectID: subjectID}},
		timetableSlots:    []db.ListTimetableSlotsRow{{AssignmentID: assignmentID, DayOfWeek: 1, StartTime: start, EndTime: end}},
	}
	svc := &Academic{q: store}
	preview, err := svc.PreviewYearRollover(context.Background(), YearRolloverPreviewInput{SourceAcademicYearID: sourceYearID, TargetAcademicYearID: targetYearID})
	if err != nil {
		t.Fatalf("PreviewYearRollover() error = %v", err)
	}
	if preview.Counts.ClassesToCreate != 0 || preview.Counts.StudentsToPromote != 1 || preview.Counts.StudentsWithoutNextClass != 2 {
		t.Fatalf("PreviewYearRollover() counts = %+v, want 0 creates, 1 promote, 2 without next", preview.Counts)
	}
	if preview.Counts.HomeroomAssignmentsCopy != 1 || preview.Counts.SubjectAssignmentsCopy != 1 || preview.Counts.TimetableSlotsCopy != 1 {
		t.Fatalf("PreviewYearRollover() copy counts = %+v, want reusable target counts only", preview.Counts)
	}
	if len(preview.StudentsWithoutNext) != 2 || preview.StudentsWithoutNext[0].Reason != "Rombel tujuan ada tetapi nonaktif" || !strings.Contains(preview.StudentsWithoutNext[1].Reason, "Tingkat akhir") {
		t.Fatalf("PreviewYearRollover() skipped = %+v, want inactive target and final level warnings", preview.StudentsWithoutNext)
	}
}

func TestAcademicDryRunImportReadsCSVReferencesAndSummarizesRows(t *testing.T) {
	yearID := documentCycleTestUUID(54)
	classID := documentCycleTestUUID(55)
	store := &fakeAcademicStore{
		years:          []db.AcademicYear{{ID: yearID, Name: "2026/2027", IsActive: true}},
		classes:        []db.ListSchoolClassesRow{{ID: classID, AcademicYearID: yearID, Code: "VII-A", Name: "VII A", Level: "VII", IsActive: true}},
		importStudents: []db.ListAcademicImportStudentsRow{{ID: documentCycleTestUUID(56), Nis: "1001", Nama: "Existing"}},
	}
	svc := &Academic{q: store}
	raw := "\ufeffNIS,NISN,Nama,Jenis Kelamin,Kode Rombel,Status\n1001,111,Alya,L,VII-A,aktif\n1002,222,Bimo,P,VII-A,aktif\n,,,,,\n1002,333,Cici,X,MISSING,aktif\n"
	result, err := svc.DryRunAcademicImport(context.Background(), AcademicImportDryRunInput{Kind: "students", CSV: raw})
	if err != nil {
		t.Fatalf("DryRunAcademicImport() error = %v", err)
	}
	if result.Kind != "siswa" || result.TotalRows != 3 || result.UpdateCount != 1 || result.AddCount != 1 || result.SkipCount != 1 || result.ErrorCount != 1 {
		t.Fatalf("DryRunAcademicImport() = %+v, want update/add/skip/error summary", result)
	}
	if len(result.Rows) != 4 || result.Rows[2].Action != "skip" || result.Rows[3].Action != "needs_review" {
		t.Fatalf("DryRunAcademicImport() rows = %+v, want blank skip and review row", result.Rows)
	}
	if !academicImportErrorsContain(result.RowErrors, "Jenis Kelamin") || !academicImportErrorsContain(result.RowErrors, "Kode Rombel") || !academicImportErrorsContainMessage(result.RowErrors, "duplikat dengan baris 3") {
		t.Fatalf("DryRunAcademicImport() row errors = %+v, want gender/class/duplicate errors", result.RowErrors)
	}
}

func TestReadAcademicCSVRejectsEmptyAndMalformedInput(t *testing.T) {
	if _, err := readAcademicCSV(" \n\t "); !errors.Is(err, domain.ErrBadRequest) || !strings.Contains(err.Error(), "file impor kosong") {
		t.Fatalf("readAcademicCSV(empty) = %v, want empty bad request", err)
	}
	if _, err := readAcademicCSV("nis,nama\n\"unterminated"); !errors.Is(err, domain.ErrBadRequest) || !strings.Contains(err.Error(), "baris 2") {
		t.Fatalf("readAcademicCSV(malformed) = %v, want row bad request", err)
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
