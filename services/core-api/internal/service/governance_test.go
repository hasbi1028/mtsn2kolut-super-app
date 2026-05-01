package service

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func governanceTestUUID(seed byte) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte{seed}, Valid: true}
}

func governanceTestDate(year int, month time.Month, day int) pgtype.Date {
	return pgtype.Date{Time: time.Date(year, month, day, 0, 0, 0, 0, time.UTC), Valid: true}
}

func TestGovernanceParseHelpers(t *testing.T) {
	id, err := ParseGovernanceOptionalUUID(" 00000000-0000-0000-0000-000000000001 ")
	if err != nil {
		t.Fatalf("ParseGovernanceOptionalUUID() error = %v", err)
	}
	if !id.Valid {
		t.Fatal("ParseGovernanceOptionalUUID() valid = false, want true")
	}
	if emptyID, err := ParseGovernanceOptionalUUID(" "); err != nil || emptyID.Valid {
		t.Fatalf("ParseGovernanceOptionalUUID(empty) = %v, %v; want invalid nil", emptyID, err)
	}
	if _, err := ParseGovernanceOptionalUUID("bad"); err == nil || err.Error() != "id tidak valid" {
		t.Fatalf("ParseGovernanceOptionalUUID(bad) error = %v, want invalid id", err)
	}

	date, err := ParseGovernanceDate(" 2026-05-01 ")
	if err != nil {
		t.Fatalf("ParseGovernanceDate() error = %v", err)
	}
	if !date.Valid || date.Time.Format("2006-01-02") != "2026-05-01" {
		t.Fatalf("ParseGovernanceDate() = %v, want 2026-05-01", date.Time)
	}
	if _, err := ParseGovernanceDate(" "); err == nil || err.Error() != "tanggal wajib diisi" {
		t.Fatalf("ParseGovernanceDate(empty) error = %v, want required date", err)
	}
	if _, err := ParseGovernanceDate("bad"); err == nil || err.Error() != "format tanggal tidak valid" {
		t.Fatalf("ParseGovernanceDate(bad) error = %v, want invalid date", err)
	}

	if optionalDate, err := ParseGovernanceOptionalDate(" "); err != nil || optionalDate.Valid {
		t.Fatalf("ParseGovernanceOptionalDate(empty) = %v, %v; want invalid nil", optionalDate, err)
	}
	if _, err := ParseGovernanceOptionalDate("bad"); err == nil || err.Error() != "format tanggal tidak valid" {
		t.Fatalf("ParseGovernanceOptionalDate(bad) error = %v, want invalid optional date", err)
	}

	timestamp, err := ParseGovernanceOptionalTimestamp("2026-05-01T10:30:00Z")
	if err != nil {
		t.Fatalf("ParseGovernanceOptionalTimestamp(rfc3339) error = %v", err)
	}
	if !timestamp.Valid || timestamp.Time.Format(time.RFC3339) != "2026-05-01T10:30:00Z" {
		t.Fatalf("ParseGovernanceOptionalTimestamp(rfc3339) = %v, want parsed timestamp", timestamp.Time)
	}
	dateTimestamp, err := ParseGovernanceOptionalTimestamp("2026-05-01")
	if err != nil {
		t.Fatalf("ParseGovernanceOptionalTimestamp(date) error = %v", err)
	}
	if !dateTimestamp.Valid || dateTimestamp.Time.Format("2006-01-02") != "2026-05-01" {
		t.Fatalf("ParseGovernanceOptionalTimestamp(date) = %v, want parsed date", dateTimestamp.Time)
	}
	if emptyTimestamp, err := ParseGovernanceOptionalTimestamp(" "); err != nil || emptyTimestamp.Valid {
		t.Fatalf("ParseGovernanceOptionalTimestamp(empty) = %v, %v; want invalid nil", emptyTimestamp, err)
	}
	if _, err := ParseGovernanceOptionalTimestamp("not-a-time"); err == nil || err.Error() != "format waktu tidak valid" {
		t.Fatalf("ParseGovernanceOptionalTimestamp(bad) error = %v, want invalid time", err)
	}

	if got := normalizeGovernanceText("  teks  ", "fallback"); got != "teks" {
		t.Fatalf("normalizeGovernanceText() = %q, want teks", got)
	}
	if got := normalizeGovernanceText(" ", "fallback"); got != "fallback" {
		t.Fatalf("normalizeGovernanceText(empty) = %q, want fallback", got)
	}
	if got := normalizeSNPStandard(" SKL "); got != "skl" {
		t.Fatalf("normalizeSNPStandard() = %q, want skl", got)
	}
}

func TestGovernanceValidationHelpers(t *testing.T) {
	validID := governanceTestUUID(1)
	startDate := governanceTestDate(2026, time.January, 1)
	endDate := governanceTestDate(2026, time.December, 31)
	beforeStart := governanceTestDate(2025, time.December, 31)

	tests := []struct {
		name    string
		err     error
		wantErr string
	}{
		{name: "unit valid", err: validateGovernanceUnit("KAMAD", "Kepala Madrasah")},
		{name: "unit empty code", err: validateGovernanceUnit("", "Kepala Madrasah"), wantErr: "kode unit wajib diisi"},
		{name: "unit empty name", err: validateGovernanceUnit("KAMAD", ""), wantErr: "nama unit wajib diisi"},
		{name: "position valid", err: validateGovernancePosition(db.CreateGovernancePositionParams{UnitID: validID, Title: "Waka"})},
		{name: "position missing unit", err: validateGovernancePosition(db.CreateGovernancePositionParams{Title: "Waka"}), wantErr: "unit kerja wajib dipilih"},
		{name: "position missing title", err: validateGovernancePosition(db.CreateGovernancePositionParams{UnitID: validID}), wantErr: "nama jabatan wajib diisi"},
		{name: "assignment valid", err: validateGovernanceAssignment(validID, validID, startDate, endDate)},
		{name: "assignment missing position", err: validateGovernanceAssignment(pgtype.UUID{}, validID, startDate, endDate), wantErr: "jabatan wajib dipilih"},
		{name: "assignment missing employee", err: validateGovernanceAssignment(validID, pgtype.UUID{}, startDate, endDate), wantErr: "pegawai wajib dipilih"},
		{name: "assignment missing start", err: validateGovernanceAssignment(validID, validID, pgtype.Date{}, endDate), wantErr: "tanggal mulai wajib diisi"},
		{name: "assignment end before start", err: validateGovernanceAssignment(validID, validID, startDate, beforeStart), wantErr: "tanggal selesai tidak boleh sebelum tanggal mulai"},
		{name: "document valid", err: validateGovernanceDocument("rkt", "RKT 2026", 2026, "draft", "skl")},
		{name: "document invalid type", err: validateGovernanceDocument("memo", "RKT", 2026, "draft", "skl"), wantErr: "jenis dokumen tidak valid"},
		{name: "document missing title", err: validateGovernanceDocument("rkt", "", 2026, "draft", "skl"), wantErr: "judul dokumen wajib diisi"},
		{name: "document invalid year", err: validateGovernanceDocument("rkt", "RKT", 1999, "draft", "skl"), wantErr: "tahun periode tidak valid"},
		{name: "document invalid status", err: validateGovernanceDocument("rkt", "RKT", 2026, "review", "skl"), wantErr: "status dokumen tidak valid"},
		{name: "document invalid snp", err: validateGovernanceDocument("rkt", "RKT", 2026, "draft", "lain"), wantErr: "standar SNP tidak valid"},
		{name: "program valid", err: validateGovernanceProgram(2026, "P-1", "Program", "planned", 10, "skl")},
		{name: "program invalid year", err: validateGovernanceProgram(1999, "P-1", "Program", "planned", 10, "skl"), wantErr: "tahun program tidak valid"},
		{name: "program missing code", err: validateGovernanceProgram(2026, "", "Program", "planned", 10, "skl"), wantErr: "kode program wajib diisi"},
		{name: "program missing name", err: validateGovernanceProgram(2026, "P-1", "", "planned", 10, "skl"), wantErr: "nama program wajib diisi"},
		{name: "program invalid status", err: validateGovernanceProgram(2026, "P-1", "Program", "review", 10, "skl"), wantErr: "status program tidak valid"},
		{name: "program invalid progress", err: validateGovernanceProgram(2026, "P-1", "Program", "planned", 101, "skl"), wantErr: "progres program harus 0 sampai 100"},
		{name: "program invalid snp", err: validateGovernanceProgram(2026, "P-1", "Program", "planned", 10, "lain"), wantErr: "standar SNP tidak valid"},
		{name: "performance target valid", err: validateGovernancePerformanceTarget(2026, validID, "hasil_kerja", "Target", "planned", 50)},
		{name: "performance invalid year", err: validateGovernancePerformanceTarget(1999, validID, "hasil_kerja", "Target", "planned", 50), wantErr: "tahun target kinerja tidak valid"},
		{name: "performance missing employee", err: validateGovernancePerformanceTarget(2026, pgtype.UUID{}, "hasil_kerja", "Target", "planned", 50), wantErr: "pegawai wajib dipilih"},
		{name: "performance invalid aspect", err: validateGovernancePerformanceTarget(2026, validID, "lain", "Target", "planned", 50), wantErr: "aspek target kinerja tidak valid"},
		{name: "performance missing title", err: validateGovernancePerformanceTarget(2026, validID, "hasil_kerja", "", "planned", 50), wantErr: "target kerja wajib diisi"},
		{name: "performance invalid status", err: validateGovernancePerformanceTarget(2026, validID, "hasil_kerja", "Target", "review", 50), wantErr: "status target kinerja tidak valid"},
		{name: "performance invalid progress", err: validateGovernancePerformanceTarget(2026, validID, "hasil_kerja", "Target", "planned", -1), wantErr: "progres target kinerja harus 0 sampai 100"},
		{name: "work plan valid", err: validateGovernanceWorkPlanItem(2026, validID, "A-1", "Kegiatan", "planned", 0, 0, 0, startDate, endDate)},
		{name: "work plan invalid year", err: validateGovernanceWorkPlanItem(1999, validID, "A-1", "Kegiatan", "planned", 0, 0, 0, startDate, endDate), wantErr: "tahun item RKT/RKJM tidak valid"},
		{name: "work plan missing program", err: validateGovernanceWorkPlanItem(2026, pgtype.UUID{}, "A-1", "Kegiatan", "planned", 0, 0, 0, startDate, endDate), wantErr: "program wajib dipilih"},
		{name: "work plan missing code", err: validateGovernanceWorkPlanItem(2026, validID, "", "Kegiatan", "planned", 0, 0, 0, startDate, endDate), wantErr: "kode kegiatan wajib diisi"},
		{name: "work plan missing name", err: validateGovernanceWorkPlanItem(2026, validID, "A-1", "", "planned", 0, 0, 0, startDate, endDate), wantErr: "nama kegiatan wajib diisi"},
		{name: "work plan invalid status", err: validateGovernanceWorkPlanItem(2026, validID, "A-1", "Kegiatan", "review", 0, 0, 0, startDate, endDate), wantErr: "status item RKT/RKJM tidak valid"},
		{name: "work plan invalid progress", err: validateGovernanceWorkPlanItem(2026, validID, "A-1", "Kegiatan", "planned", 101, 0, 0, startDate, endDate), wantErr: "progres item RKT/RKJM harus 0 sampai 100"},
		{name: "work plan invalid budget", err: validateGovernanceWorkPlanItem(2026, validID, "A-1", "Kegiatan", "planned", 0, -1, 0, startDate, endDate), wantErr: "anggaran tidak boleh negatif"},
		{name: "work plan invalid realization", err: validateGovernanceWorkPlanItem(2026, validID, "A-1", "Kegiatan", "planned", 0, 0, -1, startDate, endDate), wantErr: "realisasi anggaran tidak boleh negatif"},
		{name: "work plan invalid date range", err: validateGovernanceWorkPlanItem(2026, validID, "A-1", "Kegiatan", "planned", 0, 0, 0, startDate, beforeStart), wantErr: "tanggal selesai tidak boleh sebelum tanggal mulai"},
		{name: "evidence valid", err: validateGovernanceEvidenceItem(2026, "Bukti", "dokumen", "needed", "skl", "governance")},
		{name: "evidence invalid year", err: validateGovernanceEvidenceItem(1999, "Bukti", "dokumen", "needed", "skl", "governance"), wantErr: "tahun bukti mutu tidak valid"},
		{name: "evidence missing title", err: validateGovernanceEvidenceItem(2026, "", "dokumen", "needed", "skl", "governance"), wantErr: "judul bukti mutu wajib diisi"},
		{name: "evidence invalid type", err: validateGovernanceEvidenceItem(2026, "Bukti", "video", "needed", "skl", "governance"), wantErr: "jenis bukti mutu tidak valid"},
		{name: "evidence invalid status", err: validateGovernanceEvidenceItem(2026, "Bukti", "dokumen", "review", "skl", "governance"), wantErr: "status bukti mutu tidak valid"},
		{name: "evidence invalid snp", err: validateGovernanceEvidenceItem(2026, "Bukti", "dokumen", "needed", "lain", "governance"), wantErr: "standar SNP tidak valid"},
		{name: "evidence missing source", err: validateGovernanceEvidenceItem(2026, "Bukti", "dokumen", "needed", "skl", ""), wantErr: "sumber bukti mutu wajib diisi"},
		{name: "compliance valid", err: validateGovernanceComplianceAction(2026, "manual", "skl", "Tindak lanjut", "medium", "open")},
		{name: "compliance invalid year", err: validateGovernanceComplianceAction(1999, "manual", "skl", "Tindak lanjut", "medium", "open"), wantErr: "tahun tindak lanjut tidak valid"},
		{name: "compliance invalid source", err: validateGovernanceComplianceAction(2026, "lain", "skl", "Tindak lanjut", "medium", "open"), wantErr: "sumber tindak lanjut tidak valid"},
		{name: "compliance invalid snp", err: validateGovernanceComplianceAction(2026, "manual", "lain", "Tindak lanjut", "medium", "open"), wantErr: "standar SNP tidak valid"},
		{name: "compliance missing title", err: validateGovernanceComplianceAction(2026, "manual", "skl", "", "medium", "open"), wantErr: "judul tindak lanjut wajib diisi"},
		{name: "compliance invalid priority", err: validateGovernanceComplianceAction(2026, "manual", "skl", "Tindak lanjut", "normal", "open"), wantErr: "prioritas tindak lanjut tidak valid"},
		{name: "compliance invalid status", err: validateGovernanceComplianceAction(2026, "manual", "skl", "Tindak lanjut", "medium", "review"), wantErr: "status tindak lanjut tidak valid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr == "" {
				if tt.err != nil {
					t.Fatalf("validation error = %v", tt.err)
				}
				return
			}
			if tt.err == nil || tt.err.Error() != tt.wantErr {
				t.Fatalf("validation error = %v, want %q", tt.err, tt.wantErr)
			}
		})
	}
}

func TestGovernanceNormalizeCreateAndUpdateParams(t *testing.T) {
	currentYear := int32(time.Now().Year())

	program := normalizeGovernanceProgramCreate(db.CreateGovernanceProgramParams{
		Code:               "  P-1  ",
		Name:               "  Program Mutu  ",
		SnpStandard:        " SKL ",
		IkuCode:            " IKU-1 ",
		Indicator:          "  Indikator  ",
		TargetValue:        "  90  ",
		TargetUnit:         "  persen  ",
		RealizationSummary: "  berjalan  ",
		EvidenceUrl:        "  /evidence  ",
	})
	if program.PeriodYear != currentYear || program.Code != "P-1" || program.Name != "Program Mutu" || program.SnpStandard != "skl" || program.Status != "planned" {
		t.Fatalf("normalizeGovernanceProgramCreate() = %+v, want defaults and trimmed fields", program)
	}
	if program.IkuCode != "IKU-1" || program.Indicator != "Indikator" || program.TargetValue != "90" || program.TargetUnit != "persen" || program.RealizationSummary != "berjalan" || program.EvidenceUrl != "/evidence" {
		t.Fatalf("normalizeGovernanceProgramCreate() = %+v, want trimmed detail fields", program)
	}

	programUpdate := normalizeGovernanceProgramUpdate(db.UpdateGovernanceProgramParams{
		Code:               " P-2 ",
		Name:               " Program Update ",
		SnpStandard:        " Ptk ",
		IkuCode:            " IKU-2 ",
		Indicator:          " Indikator 2 ",
		TargetValue:        " 80 ",
		TargetUnit:         " persen ",
		RealizationSummary: " ringkas ",
		EvidenceUrl:        " /url ",
	})
	if programUpdate.PeriodYear != currentYear || programUpdate.Code != "P-2" || programUpdate.SnpStandard != "ptk" || programUpdate.Status != "planned" {
		t.Fatalf("normalizeGovernanceProgramUpdate() = %+v, want defaults and trimmed fields", programUpdate)
	}

	workPlan := normalizeGovernanceWorkPlanItemCreate(db.CreateGovernanceWorkPlanItemParams{
		ActivityCode:    " A-1 ",
		ActivityName:    " Kegiatan ",
		OutputIndicator: " Output ",
		TargetVolume:    " 10 ",
		TargetUnit:      " siswa ",
		BudgetSource:    " BOS ",
		EvidenceUrl:     " /bukti ",
		Notes:           " catatan ",
	})
	if workPlan.PeriodYear != currentYear || workPlan.ActivityCode != "A-1" || workPlan.ActivityName != "Kegiatan" || workPlan.Status != "planned" {
		t.Fatalf("normalizeGovernanceWorkPlanItemCreate() = %+v, want defaults and trimmed fields", workPlan)
	}
	if workPlan.OutputIndicator != "Output" || workPlan.TargetVolume != "10" || workPlan.TargetUnit != "siswa" || workPlan.BudgetSource != "BOS" || workPlan.EvidenceUrl != "/bukti" || workPlan.Notes != "catatan" {
		t.Fatalf("normalizeGovernanceWorkPlanItemCreate() = %+v, want trimmed detail fields", workPlan)
	}

	workPlanUpdate := normalizeGovernanceWorkPlanItemUpdate(db.UpdateGovernanceWorkPlanItemParams{
		ActivityCode:    " A-2 ",
		ActivityName:    " Kegiatan Update ",
		OutputIndicator: " Output Update ",
		TargetVolume:    " 12 ",
		TargetUnit:      " guru ",
		BudgetSource:    " Komite ",
		EvidenceUrl:     " /bukti-2 ",
		Notes:           " catatan 2 ",
	})
	if workPlanUpdate.PeriodYear != currentYear || workPlanUpdate.ActivityCode != "A-2" || workPlanUpdate.Status != "planned" {
		t.Fatalf("normalizeGovernanceWorkPlanItemUpdate() = %+v, want defaults and trimmed fields", workPlanUpdate)
	}

	target := normalizeGovernancePerformanceTargetCreate(db.CreateGovernancePerformanceTargetParams{
		Title:       " Target ",
		Indicator:   " Indikator ",
		TargetValue: " 100 ",
		TargetUnit:  " persen ",
		EvidenceUrl: " /target ",
		ReviewNotes: " catatan ",
	})
	if target.PeriodYear != currentYear || target.Aspect != "hasil_kerja" || target.Title != "Target" || target.Status != "planned" {
		t.Fatalf("normalizeGovernancePerformanceTargetCreate() = %+v, want defaults and trimmed fields", target)
	}
	if target.Indicator != "Indikator" || target.TargetValue != "100" || target.TargetUnit != "persen" || target.EvidenceUrl != "/target" || target.ReviewNotes != "catatan" {
		t.Fatalf("normalizeGovernancePerformanceTargetCreate() = %+v, want trimmed detail fields", target)
	}

	targetUpdate := normalizeGovernancePerformanceTargetUpdate(db.UpdateGovernancePerformanceTargetParams{
		Title:       " Target 2 ",
		Indicator:   " Indikator 2 ",
		TargetValue: " 95 ",
		TargetUnit:  " persen ",
		EvidenceUrl: " /target-2 ",
		ReviewNotes: " catatan 2 ",
	})
	if targetUpdate.PeriodYear != currentYear || targetUpdate.Aspect != "hasil_kerja" || targetUpdate.Title != "Target 2" || targetUpdate.Status != "planned" {
		t.Fatalf("normalizeGovernancePerformanceTargetUpdate() = %+v, want defaults and trimmed fields", targetUpdate)
	}

	evidence := normalizeGovernanceEvidenceItemCreate(db.CreateGovernanceEvidenceItemParams{
		Title:       " Bukti ",
		SnpStandard: " ISI ",
		EvidenceUrl: " /bukti ",
		Notes:       " catatan ",
	})
	if evidence.PeriodYear != currentYear || evidence.Title != "Bukti" || evidence.EvidenceType != "dokumen" || evidence.SnpStandard != "isi" || evidence.SourceModule != "governance" || evidence.Status != "needed" {
		t.Fatalf("normalizeGovernanceEvidenceItemCreate() = %+v, want defaults and trimmed fields", evidence)
	}
	if evidence.EvidenceUrl != "/bukti" || evidence.Notes != "catatan" {
		t.Fatalf("normalizeGovernanceEvidenceItemCreate() = %+v, want trimmed details", evidence)
	}

	evidenceUpdate := normalizeGovernanceEvidenceItemUpdate(db.UpdateGovernanceEvidenceItemParams{
		Title:       " Bukti 2 ",
		SnpStandard: " Proses ",
		EvidenceUrl: " /bukti-2 ",
		Notes:       " catatan 2 ",
	})
	if evidenceUpdate.PeriodYear != currentYear || evidenceUpdate.Title != "Bukti 2" || evidenceUpdate.EvidenceType != "dokumen" || evidenceUpdate.SnpStandard != "proses" || evidenceUpdate.SourceModule != "governance" || evidenceUpdate.Status != "needed" {
		t.Fatalf("normalizeGovernanceEvidenceItemUpdate() = %+v, want defaults and trimmed fields", evidenceUpdate)
	}

	action := normalizeGovernanceComplianceActionCreate(db.CreateGovernanceComplianceActionParams{
		SnpStandard:   " PTK ",
		Title:         " Tindak lanjut ",
		Description:   " Deskripsi ",
		FollowUpNotes: " Catatan ",
		EvidenceUrl:   " /aksi ",
	})
	if action.PeriodYear != currentYear || action.SourceType != "manual" || action.SnpStandard != "ptk" || action.Title != "Tindak lanjut" || action.Priority != "medium" || action.Status != "open" || action.CompletedAt.Valid {
		t.Fatalf("normalizeGovernanceComplianceActionCreate() = %+v, want defaults and no completed_at", action)
	}
	if action.Description != "Deskripsi" || action.FollowUpNotes != "Catatan" || action.EvidenceUrl != "/aksi" {
		t.Fatalf("normalizeGovernanceComplianceActionCreate() = %+v, want trimmed details", action)
	}

	actionUpdate := normalizeGovernanceComplianceActionUpdate(db.UpdateGovernanceComplianceActionParams{
		SnpStandard:   " Sarpras ",
		Title:         " Tindak lanjut 2 ",
		Description:   " Deskripsi 2 ",
		Status:        "done",
		FollowUpNotes: " Catatan 2 ",
		EvidenceUrl:   " /aksi-2 ",
	})
	if actionUpdate.PeriodYear != currentYear || actionUpdate.SourceType != "manual" || actionUpdate.SnpStandard != "sarpras" || actionUpdate.Title != "Tindak lanjut 2" || actionUpdate.Priority != "medium" || !actionUpdate.CompletedAt.Valid {
		t.Fatalf("normalizeGovernanceComplianceActionUpdate() = %+v, want done action with completed_at", actionUpdate)
	}

	knownCompletedAt := pgtype.Timestamptz{Time: time.Date(2026, time.May, 1, 8, 0, 0, 0, time.UTC), Valid: true}
	if got := normalizeGovernanceComplianceCompletedAt("done", knownCompletedAt); got != knownCompletedAt {
		t.Fatalf("normalizeGovernanceComplianceCompletedAt(done existing) = %v, want existing timestamp", got)
	}
	if got := normalizeGovernanceComplianceCompletedAt("open", knownCompletedAt); got.Valid {
		t.Fatalf("normalizeGovernanceComplianceCompletedAt(open) valid = true, want false")
	}
}
