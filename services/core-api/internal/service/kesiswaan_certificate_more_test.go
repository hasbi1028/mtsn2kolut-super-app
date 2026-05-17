package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestKesiswaanCreateStudentTransferMoreValidatesBeforePool(t *testing.T) {
	svc := &Kesiswaan{}
	_, err := svc.CreateStudentTransfer(context.Background(), db.CreateStudentTransferParams{
		TransferDate:      kesiswaanTestDate(),
		TransferType:      "out",
		DestinationSchool: "MTs Tujuan",
	})
	if err == nil || !strings.Contains(err.Error(), "siswa wajib dipilih") {
		t.Fatalf("CreateStudentTransfer(missing student) error = %v, want student validation", err)
	}

	_, err = svc.CreateStudentTransfer(context.Background(), db.CreateStudentTransferParams{
		StudentID:         kesiswaanTestUUID(41),
		TransferDate:      kesiswaanTestDate(),
		TransferType:      "stay",
		DestinationSchool: "MTs Tujuan",
	})
	if err == nil || !strings.Contains(err.Error(), "jenis mutasi") {
		t.Fatalf("CreateStudentTransfer(invalid type) error = %v, want type validation", err)
	}
}

func TestKesiswaanCreateStudentTransferMoreRequiresPoolAfterNormalization(t *testing.T) {
	svc := &Kesiswaan{}
	_, err := svc.CreateStudentTransfer(context.Background(), db.CreateStudentTransferParams{
		StudentID:         kesiswaanTestUUID(42),
		DestinationSchool: "  MTs Tujuan  ",
		Reason:            "  pindah domisili  ",
	})
	if err == nil || !strings.Contains(err.Error(), "layanan mutasi siswa belum siap") {
		t.Fatalf("CreateStudentTransfer(nil pool) error = %v, want service readiness error", err)
	}
}

func TestCreateStudentTransferMoreReturnsCreatedRowAfterLifecycleUpdate(t *testing.T) {
	studentID := kesiswaanTestUUID(43)
	rowID := kesiswaanTestUUID(44)
	store := &fakeKesiswaanStore{transferRow: db.StudentTransfer{ID: rowID, StudentID: studentID, TransferType: "out"}}

	row, err := createStudentTransfer(context.Background(), store, db.CreateStudentTransferParams{
		StudentID:         studentID,
		TransferDate:      kesiswaanTestDate(),
		TransferType:      "out",
		DestinationSchool: "MTs Tujuan",
	})
	if err != nil {
		t.Fatalf("createStudentTransfer() error = %v", err)
	}
	if row.ID != rowID || row.StudentID != studentID {
		t.Fatalf("createStudentTransfer() row = %+v, want created transfer row", row)
	}
	if store.transferCreate.StudentID != studentID || store.lifecycleArg.ID != studentID || store.lifecycleArg.Status != db.StudentStatusEnumMutated || store.lifecycleArg.IsActive {
		t.Fatalf("transfer/lifecycle args = %+v/%+v, want outgoing transfer and mutated inactive lifecycle", store.transferCreate, store.lifecycleArg)
	}
}

func TestStudentCertificateCreateMoreNormalizesInputBeforePool(t *testing.T) {
	templateID := studentCertificateTestUUID(50)
	studentID := studentCertificateTestUUID(51)
	svc := &StudentCertificate{}

	_, err := svc.Create(context.Background(), CreateStudentCertificateInput{
		TemplateID:   "  " + templateID.String() + "  ",
		StudentID:    "  " + studentID.String() + "  ",
		TanggalSurat: " 2026-05-01 ",
		Purpose:      "  Surat aktif  ",
		Recipient:    "   ",
		Remarks:      "  catatan  ",
	})
	if err == nil || !strings.Contains(err.Error(), "layanan surat keterangan siswa belum siap") {
		t.Fatalf("StudentCertificate.Create(nil pool) error = %v, want service readiness error after valid normalized input", err)
	}
}

func TestCreateStudentCertificateMoreStopsBeforeCertificateWhenOutgoingFails(t *testing.T) {
	templateID := studentCertificateTestUUID(52)
	studentID := studentCertificateTestUUID(53)
	outErr := errors.New("outgoing unavailable")
	store := &fakeStudentCertificateStore{
		template:    db.CertificateTemplate{ID: templateID, Code: "AKTIF", Name: "Surat Aktif", IsActive: true},
		student:     db.GetCertificateStudentSnapshotRow{ID: studentID, Nama: "Alya", Gender: db.GenderEnumP, Status: db.StudentStatusEnumActive},
		sequence:    9,
		outgoingErr: outErr,
	}

	_, err := createStudentCertificate(context.Background(), store, normalizedStudentCertificateInput{
		TemplateID:   templateID,
		StudentID:    studentID,
		TanggalSurat: studentCertificateTestDate(),
		Purpose:      "Beasiswa",
		Recipient:    "Komite",
	})
	if !errors.Is(err, outErr) {
		t.Fatalf("createStudentCertificate(outgoing error) = %v, want %v", err, outErr)
	}
	if store.createArg.TemplateID.Valid || store.getID.Valid {
		t.Fatalf("certificate write/detail lookup occurred despite outgoing error: create=%+v get=%v", store.createArg, store.getID)
	}
}

func TestNormalizeStudentCertificateInputMoreRejectsBadDateAndPurpose(t *testing.T) {
	templateID := studentCertificateTestUUID(54).String()
	studentID := studentCertificateTestUUID(55).String()
	tests := []struct {
		name  string
		input CreateStudentCertificateInput
		want  string
	}{
		{name: "bad date", input: CreateStudentCertificateInput{TemplateID: templateID, StudentID: studentID, TanggalSurat: "2026/05/01", Purpose: "Aktif"}, want: "format tanggal"},
		{name: "blank purpose", input: CreateStudentCertificateInput{TemplateID: templateID, StudentID: studentID, TanggalSurat: "2026-05-01", Purpose: "   "}, want: "keperluan"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := normalizeStudentCertificateInput(tt.input)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("normalizeStudentCertificateInput() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestBuildStudentCertificateSnapshotMoreIncludesOptionalDates(t *testing.T) {
	date := studentCertificateTestDate()
	snapshot, err := buildStudentCertificateSnapshot(
		db.CertificateTemplate{ID: studentCertificateTestUUID(56), Code: "AKTIF", Name: "Surat Aktif"},
		db.GetCertificateStudentSnapshotRow{ID: studentCertificateTestUUID(57), Nama: "Alya", TanggalLahir: date, Gender: db.GenderEnumP, Status: db.StudentStatusEnumActive},
		"009/PP.00.4/MTs.20.05/V/2026",
		normalizedStudentCertificateInput{TanggalSurat: date, Purpose: "Aktif", Recipient: "Komite"},
	)
	if err != nil {
		t.Fatalf("buildStudentCertificateSnapshot() error = %v", err)
	}
	text := string(snapshot)
	for _, want := range []string{"\"tanggal_lahir\":\"2026-05-01\"", "\"tanggal_surat\":\"2026-05-01\"", "\"recipient\":\"Komite\""} {
		if !strings.Contains(text, want) {
			t.Fatalf("snapshot %s missing %s", text, want)
		}
	}
}
