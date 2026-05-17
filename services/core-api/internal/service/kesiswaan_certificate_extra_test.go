package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeKesiswaanTransferStore struct {
	createArg       db.CreateStudentTransferParams
	createRow       db.StudentTransfer
	createErr       error
	lifecycleArg    db.UpdateKesiswaanStudentLifecycleParams
	lifecycleCalled bool
	lifecycleErr    error
}

func (f *fakeKesiswaanTransferStore) CreateStudentTransfer(ctx context.Context, arg db.CreateStudentTransferParams) (db.StudentTransfer, error) {
	f.createArg = arg
	if f.createErr != nil {
		return db.StudentTransfer{}, f.createErr
	}
	if f.createRow.ID.Valid {
		return f.createRow, nil
	}
	return db.StudentTransfer{ID: kesiswaanTestUUID(91), StudentID: arg.StudentID, TransferDate: arg.TransferDate, TransferType: arg.TransferType, PreviousSchool: arg.PreviousSchool, DestinationSchool: arg.DestinationSchool, Reason: arg.Reason, DocumentRef: arg.DocumentRef, Notes: arg.Notes, RecordedByUserID: arg.RecordedByUserID}, nil
}

func (f *fakeKesiswaanTransferStore) UpdateKesiswaanStudentLifecycle(ctx context.Context, arg db.UpdateKesiswaanStudentLifecycleParams) (db.Student, error) {
	f.lifecycleCalled = true
	f.lifecycleArg = arg
	if f.lifecycleErr != nil {
		return db.Student{}, f.lifecycleErr
	}
	return db.Student{ID: arg.ID, Status: arg.Status, IsActive: arg.IsActive}, nil
}

func TestCreateStudentTransferOutUpdatesLifecycleToMutatedInactive(t *testing.T) {
	studentID := kesiswaanTestUUID(1)
	store := &fakeKesiswaanTransferStore{}
	row, err := createStudentTransfer(context.Background(), store, db.CreateStudentTransferParams{
		StudentID:         studentID,
		TransferDate:      kesiswaanTestDate(),
		TransferType:      "out",
		DestinationSchool: " MTs Tujuan ",
		Reason:            "Pindah domisili",
	})
	if err != nil {
		t.Fatalf("createStudentTransfer(out) error = %v", err)
	}
	if row.StudentID != studentID || row.TransferType != "out" {
		t.Fatalf("createStudentTransfer(out) row = %+v, want transfer for student", row)
	}
	if !store.lifecycleCalled || store.lifecycleArg.ID != studentID || store.lifecycleArg.Status != db.StudentStatusEnumMutated || store.lifecycleArg.IsActive {
		t.Fatalf("lifecycle arg = %+v called=%v, want mutated inactive", store.lifecycleArg, store.lifecycleCalled)
	}
}

func TestCreateStudentTransferInUpdatesLifecycleToActive(t *testing.T) {
	studentID := kesiswaanTestUUID(2)
	store := &fakeKesiswaanTransferStore{}
	_, err := createStudentTransfer(context.Background(), store, db.CreateStudentTransferParams{
		StudentID:        studentID,
		TransferDate:     kesiswaanTestDate(),
		TransferType:     "in",
		PreviousSchool:   "SMP Asal",
		RecordedByUserID: kesiswaanTestUUID(3),
	})
	if err != nil {
		t.Fatalf("createStudentTransfer(in) error = %v", err)
	}
	if !store.lifecycleCalled || store.lifecycleArg.ID != studentID || store.lifecycleArg.Status != db.StudentStatusEnumActive || !store.lifecycleArg.IsActive {
		t.Fatalf("lifecycle arg = %+v called=%v, want active", store.lifecycleArg, store.lifecycleCalled)
	}
}

func TestCreateStudentTransferStopsWhenCreateOrLifecycleFails(t *testing.T) {
	createErr := errors.New("create failed")
	store := &fakeKesiswaanTransferStore{createErr: createErr}
	_, err := createStudentTransfer(context.Background(), store, db.CreateStudentTransferParams{StudentID: kesiswaanTestUUID(4), TransferDate: kesiswaanTestDate(), TransferType: "out", DestinationSchool: "Tujuan"})
	if !errors.Is(err, createErr) || store.lifecycleCalled {
		t.Fatalf("createStudentTransfer(create err) err=%v lifecycleCalled=%v, want create error and no lifecycle", err, store.lifecycleCalled)
	}

	lifecycleErr := errors.New("lifecycle failed")
	store = &fakeKesiswaanTransferStore{lifecycleErr: lifecycleErr}
	_, err = createStudentTransfer(context.Background(), store, db.CreateStudentTransferParams{StudentID: kesiswaanTestUUID(5), TransferDate: kesiswaanTestDate(), TransferType: "out", DestinationSchool: "Tujuan"})
	if !errors.Is(err, lifecycleErr) || !store.lifecycleCalled {
		t.Fatalf("createStudentTransfer(lifecycle err) err=%v lifecycleCalled=%v, want lifecycle error", err, store.lifecycleCalled)
	}
}

func TestNormalizeStudentTransferCreateDefaultsTrimsAndValidatesDirectionRequirements(t *testing.T) {
	arg := normalizeStudentTransferCreate(db.CreateStudentTransferParams{StudentID: kesiswaanTestUUID(6), PreviousSchool: " Asal ", DestinationSchool: " Tujuan ", Reason: " Alasan ", DocumentRef: " DOC ", Notes: " Catatan "})
	if arg.TransferType != "out" || !arg.TransferDate.Valid || arg.PreviousSchool != "Asal" || arg.DestinationSchool != "Tujuan" || arg.Reason != "Alasan" || arg.DocumentRef != "DOC" || arg.Notes != "Catatan" {
		t.Fatalf("normalizeStudentTransferCreate() = %+v, want defaults and trimmed text", arg)
	}
	if err := validateStudentTransfer(arg); err != nil {
		t.Fatalf("validateStudentTransfer(normalized out) error = %v", err)
	}
	if err := validateStudentTransfer(db.CreateStudentTransferParams{StudentID: kesiswaanTestUUID(7), TransferDate: kesiswaanTestDate(), TransferType: "out"}); err == nil || !strings.Contains(err.Error(), "sekolah tujuan") {
		t.Fatalf("validateStudentTransfer(out missing destination) = %v, want destination error", err)
	}
	if err := validateStudentTransfer(db.CreateStudentTransferParams{StudentID: kesiswaanTestUUID(8), TransferDate: kesiswaanTestDate(), TransferType: "in"}); err == nil || !strings.Contains(err.Error(), "sekolah asal") {
		t.Fatalf("validateStudentTransfer(in missing previous) = %v, want previous school error", err)
	}
}

func TestStudentCertificateCreateNilPoolRejectsAfterInputNormalization(t *testing.T) {
	templateID := kesiswaanTestUUID(10)
	studentID := kesiswaanTestUUID(11)
	svc := &StudentCertificate{q: &fakeStudentCertificateStore{}}
	_, err := svc.Create(context.Background(), CreateStudentCertificateInput{TemplateID: templateID.String(), StudentID: studentID.String(), TanggalSurat: "2026-05-02", Purpose: "  Beasiswa  "})
	if err == nil || !strings.Contains(err.Error(), "belum siap") {
		t.Fatalf("StudentCertificate.Create(nil pool) error = %v, want service readiness error", err)
	}
}

func TestCreateStudentCertificateIssuesOutgoingLetterAndBuildsSnapshot(t *testing.T) {
	templateID := kesiswaanTestUUID(12)
	studentID := kesiswaanTestUUID(13)
	userID := kesiswaanTestUUID(14)
	employeeID := kesiswaanTestUUID(15)
	arg, err := normalizeStudentCertificateInput(CreateStudentCertificateInput{TemplateID: templateID.String(), StudentID: studentID.String(), TanggalSurat: "2026-05-02", Purpose: " Beasiswa ", Recipient: " ", Remarks: " Prestasi ", CreatedByUserID: userID, IssuedByEmployeeID: employeeID})
	if err != nil {
		t.Fatalf("normalizeStudentCertificateInput() error = %v", err)
	}
	store := &fakeStudentCertificateStore{
		template:    db.CertificateTemplate{ID: templateID, Code: "AKTIF", Name: "Surat Aktif", DefaultPurpose: "Administrasi", BodyTemplate: "Body", IsActive: true},
		student:     db.GetCertificateStudentSnapshotRow{ID: studentID, Nis: "1001", Nisn: "999", Nama: "Alya", Gender: db.GenderEnumP, Status: db.StudentStatusEnumActive, ClassID: kesiswaanTestUUID(16), ClassName: "VII A", ClassLevel: 7, ParentName: "Ortu", TanggalLahir: kesiswaanTestDate(), Agama: "Islam"},
		outgoingRow: db.OutgoingLetter{ID: kesiswaanTestUUID(93)},
		createRow:   db.StudentCertificate{ID: kesiswaanTestUUID(92)},
		getRow:      db.GetStudentCertificateRow{ID: kesiswaanTestUUID(92), TemplateID: templateID, StudentID: studentID, NomorSurat: "007/PP.00.4/MTs.20.05/V/2026", Purpose: "Beasiswa", Recipient: "Yang berkepentingan"},
		sequence:    7,
	}
	detail, err := createStudentCertificate(context.Background(), store, arg)
	if err != nil {
		t.Fatalf("createStudentCertificate() error = %v", err)
	}
	if detail.NomorSurat != "007/PP.00.4/MTs.20.05/V/2026" || store.sequenceArg.Year != 2026 || store.sequenceArg.ClassificationCode != studentCertificateClassificationCode {
		t.Fatalf("issued number=%q sequenceArg=%+v, want generated certificate number", detail.NomorSurat, store.sequenceArg)
	}
	if store.outgoingArg.Tujuan != "Yang berkepentingan" || store.outgoingArg.Perihal != "Surat Aktif - Alya" || store.outgoingArg.IssuedByEmployeeID != employeeID || !strings.Contains(store.outgoingArg.Catatan, "Catatan: Prestasi") {
		t.Fatalf("outgoing arg = %+v, want default recipient/perihal/note/issuer", store.outgoingArg)
	}
	if store.createArg.TemplateID != templateID || store.createArg.StudentID != studentID || store.createArg.CreatedByUserID != userID || store.createArg.Purpose != "Beasiswa" || store.createArg.Recipient != "Yang berkepentingan" {
		t.Fatalf("certificate create arg = %+v, want normalized IDs and fields", store.createArg)
	}
	var snapshot map[string]any
	if err := json.Unmarshal(store.createArg.SnapshotData, &snapshot); err != nil {
		t.Fatalf("snapshot json unmarshal error = %v", err)
	}
	letter := snapshot["letter"].(map[string]any)
	student := snapshot["student"].(map[string]any)
	if letter["nomor_surat"] != detail.NomorSurat || letter["recipient"] != "Yang berkepentingan" || student["nama"] != "Alya" || student["class_name"] != "VII A" {
		t.Fatalf("snapshot = %+v, want letter and student payload", snapshot)
	}
}

func TestCreateStudentCertificateRejectsInactiveTemplateAndPropagatesSequenceError(t *testing.T) {
	arg := normalizedStudentCertificateInput{TemplateID: kesiswaanTestUUID(20), StudentID: kesiswaanTestUUID(21), TanggalSurat: kesiswaanTestDate(), Purpose: "Keperluan", Recipient: "Penerima"}
	store := &fakeStudentCertificateStore{template: db.CertificateTemplate{ID: arg.TemplateID, Code: "X", Name: "Inactive", IsActive: false}}
	_, err := createStudentCertificate(context.Background(), store, arg)
	if err == nil || !strings.Contains(err.Error(), "template surat tidak aktif") {
		t.Fatalf("createStudentCertificate(inactive template) error = %v, want inactive error", err)
	}

	seqErr := errors.New("sequence down")
	store = &fakeStudentCertificateStore{template: db.CertificateTemplate{ID: arg.TemplateID, Code: "X", Name: "Active", IsActive: true}, student: db.GetCertificateStudentSnapshotRow{ID: arg.StudentID, Nama: "Alya"}, sequenceErr: seqErr}
	_, err = createStudentCertificate(context.Background(), store, arg)
	if !errors.Is(err, seqErr) || store.outgoingArg.NomorSurat != "" || store.createArg.TemplateID.Valid {
		t.Fatalf("createStudentCertificate(sequence err) err=%v outgoing=%+v cert=%+v, want wrapped sequence error before writes", err, store.outgoingArg, store.createArg)
	}
}
