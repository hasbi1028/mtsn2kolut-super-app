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

func stringContainsAll(value string, needles []string) bool {
	for _, needle := range needles {
		if !strings.Contains(value, needle) {
			return false
		}
	}
	return true
}
