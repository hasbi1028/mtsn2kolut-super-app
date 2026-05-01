package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeStudentCertificateStore struct {
	templatesArg      bool
	templates         []db.CertificateTemplate
	template          db.CertificateTemplate
	templateID        pgtype.UUID
	templateErr       error
	studentOptionsArg db.ListStudentCertificateOptionsParams
	studentOptions    []db.ListStudentCertificateOptionsRow
	studentID         pgtype.UUID
	student           db.GetCertificateStudentSnapshotRow
	studentErr        error
	listArg           db.ListStudentCertificatesParams
	listRows          []db.ListStudentCertificatesRow
	getID             pgtype.UUID
	getRow            db.GetStudentCertificateRow
	getErr            error
	createArg         db.CreateStudentCertificateParams
	createRow         db.StudentCertificate
	createErr         error
	cancelArg         db.CancelStudentCertificateParams
	cancelRow         db.StudentCertificate
	outgoingArg       db.CreateOutgoingLetterParams
	outgoingRow       db.OutgoingLetter
	outgoingErr       error
	sequenceArg       db.IssueOutgoingLetterSequenceParams
	sequence          int32
	sequenceErr       error
}

func (f *fakeStudentCertificateStore) ListCertificateTemplates(ctx context.Context, activeOnly bool) ([]db.CertificateTemplate, error) {
	f.templatesArg = activeOnly
	return f.templates, nil
}

func (f *fakeStudentCertificateStore) GetCertificateTemplateByID(ctx context.Context, id pgtype.UUID) (db.CertificateTemplate, error) {
	f.templateID = id
	return f.template, f.templateErr
}

func (f *fakeStudentCertificateStore) ListStudentCertificateOptions(ctx context.Context, arg db.ListStudentCertificateOptionsParams) ([]db.ListStudentCertificateOptionsRow, error) {
	f.studentOptionsArg = arg
	return f.studentOptions, nil
}

func (f *fakeStudentCertificateStore) GetCertificateStudentSnapshot(ctx context.Context, id pgtype.UUID) (db.GetCertificateStudentSnapshotRow, error) {
	f.studentID = id
	return f.student, f.studentErr
}

func (f *fakeStudentCertificateStore) ListStudentCertificates(ctx context.Context, arg db.ListStudentCertificatesParams) ([]db.ListStudentCertificatesRow, error) {
	f.listArg = arg
	return f.listRows, nil
}

func (f *fakeStudentCertificateStore) GetStudentCertificate(ctx context.Context, id pgtype.UUID) (db.GetStudentCertificateRow, error) {
	f.getID = id
	return f.getRow, f.getErr
}

func (f *fakeStudentCertificateStore) CreateStudentCertificate(ctx context.Context, arg db.CreateStudentCertificateParams) (db.StudentCertificate, error) {
	f.createArg = arg
	return f.createRow, f.createErr
}

func (f *fakeStudentCertificateStore) CancelStudentCertificate(ctx context.Context, arg db.CancelStudentCertificateParams) (db.StudentCertificate, error) {
	f.cancelArg = arg
	if f.cancelRow.ID.Valid {
		return f.cancelRow, nil
	}
	return db.StudentCertificate{ID: arg.ID, Remarks: arg.Remarks, Status: "canceled"}, nil
}

func (f *fakeStudentCertificateStore) CreateOutgoingLetter(ctx context.Context, arg db.CreateOutgoingLetterParams) (db.OutgoingLetter, error) {
	f.outgoingArg = arg
	return f.outgoingRow, f.outgoingErr
}

func (f *fakeStudentCertificateStore) IssueOutgoingLetterSequence(ctx context.Context, arg db.IssueOutgoingLetterSequenceParams) (int32, error) {
	f.sequenceArg = arg
	return f.sequence, f.sequenceErr
}

func studentCertificateTestUUID(seed byte) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte{seed}, Valid: true}
}

func studentCertificateTestDate() pgtype.Date {
	return pgtype.Date{Time: time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC), Valid: true}
}

func TestStudentCertificateListMethodsValidateFilters(t *testing.T) {
	store := &fakeStudentCertificateStore{
		templates:      []db.CertificateTemplate{{Code: "AKTIF", IsActive: true}},
		studentOptions: []db.ListStudentCertificateOptionsRow{{Nama: "Siswa A"}},
		listRows:       []db.ListStudentCertificatesRow{{Status: "issued"}},
	}
	svc := &StudentCertificate{q: store}

	templates, err := svc.ListTemplates(context.Background(), true)
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}
	if !store.templatesArg || len(templates) != 1 {
		t.Fatalf("ListTemplates() active=%v rows=%+v, want active rows", store.templatesArg, templates)
	}

	students, err := svc.ListStudents(context.Background(), "  Ali  ", "active")
	if err != nil {
		t.Fatalf("ListStudents() error = %v", err)
	}
	if store.studentOptionsArg.Search != "Ali" || store.studentOptionsArg.Status != "active" || len(students) != 1 {
		t.Fatalf("ListStudents() arg=%+v rows=%+v, want trimmed active query", store.studentOptionsArg, students)
	}
	if _, err := svc.ListStudents(context.Background(), "", "dropout"); err == nil || err.Error() != "status siswa tidak valid" {
		t.Fatalf("ListStudents(invalid status) error = %v, want status validation", err)
	}

	rows, err := svc.List(context.Background(), "  nomor  ", "issued", "  SKA  ")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if store.listArg.Search != "nomor" || store.listArg.Status != "issued" || store.listArg.TemplateCode != "SKA" || len(rows) != 1 {
		t.Fatalf("List() arg=%+v rows=%+v, want trimmed issued query", store.listArg, rows)
	}
	if _, err := svc.List(context.Background(), "", "draft", ""); err == nil || err.Error() != "status surat tidak valid" {
		t.Fatalf("List(invalid status) error = %v, want status validation", err)
	}
}

func TestStudentCertificateGetAndCancelDelegate(t *testing.T) {
	id := studentCertificateTestUUID(1)
	store := &fakeStudentCertificateStore{
		getRow: db.GetStudentCertificateRow{ID: id, Status: "issued"},
	}
	svc := &StudentCertificate{q: store}

	row, err := svc.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if store.getID != id || row.ID != id {
		t.Fatalf("Get() id=%v row=%+v, want delegated row", store.getID, row)
	}

	canceled, err := svc.Cancel(context.Background(), id, "  salah data  ")
	if err != nil {
		t.Fatalf("Cancel() error = %v", err)
	}
	if store.cancelArg.ID != id || store.cancelArg.Remarks != "salah data" || canceled.Status != "canceled" {
		t.Fatalf("Cancel() arg=%+v row=%+v, want trimmed cancel", store.cancelArg, canceled)
	}
}

func TestCreateStudentCertificateIssuesOutgoingLetterAndSnapshot(t *testing.T) {
	templateID := studentCertificateTestUUID(11)
	studentID := studentCertificateTestUUID(12)
	classID := studentCertificateTestUUID(13)
	outgoingID := studentCertificateTestUUID(14)
	certificateID := studentCertificateTestUUID(15)
	createdBy := studentCertificateTestUUID(16)
	issuedBy := studentCertificateTestUUID(17)
	tanggalSurat := studentCertificateTestDate()
	store := &fakeStudentCertificateStore{
		template: db.CertificateTemplate{
			ID:             templateID,
			Code:           "SKA",
			Name:           "Surat Aktif",
			DefaultPurpose: "Administrasi",
			BodyTemplate:   "Isi surat",
			IsActive:       true,
		},
		student: db.GetCertificateStudentSnapshotRow{
			ID:           studentID,
			Nis:          "123",
			Nisn:         "456",
			Nama:         "Siswa A",
			Gender:       db.GenderEnumL,
			Status:       db.StudentStatusEnumActive,
			ClassID:      classID,
			ClassName:    "VII A",
			ClassLevel:   7,
			ParentName:   "Orang Tua",
			ParentPhone:  "0812",
			Nik:          "1234567890123456",
			TempatLahir:  "Kolaka Utara",
			TanggalLahir: tanggalSurat,
			Alamat:       "Lasusua",
			Agama:        "Islam",
			Phone:        "0822",
		},
		sequence:    42,
		outgoingRow: db.OutgoingLetter{ID: outgoingID},
		createRow:   db.StudentCertificate{ID: certificateID},
		getRow:      db.GetStudentCertificateRow{ID: certificateID, NomorSurat: "042/PP.00.4/MTs.20.05/V/2026", Status: "issued"},
	}
	arg := normalizedStudentCertificateInput{
		TemplateID:         templateID,
		StudentID:          studentID,
		TanggalSurat:       tanggalSurat,
		Purpose:            "Beasiswa",
		Recipient:          "Komite",
		Remarks:            "Prioritas",
		CreatedByUserID:    createdBy,
		IssuedByEmployeeID: issuedBy,
	}

	got, err := createStudentCertificate(context.Background(), store, arg)
	if err != nil {
		t.Fatalf("createStudentCertificate() error = %v", err)
	}
	if got.ID != certificateID || store.getID != certificateID {
		t.Fatalf("createStudentCertificate() got/getID = %+v/%v, want detail by created certificate", got, store.getID)
	}
	if store.templateID != templateID || store.studentID != studentID {
		t.Fatalf("template/student ids = %v/%v, want %v/%v", store.templateID, store.studentID, templateID, studentID)
	}
	if store.sequenceArg.Year != 2026 || store.sequenceArg.ClassificationCode != studentCertificateClassificationCode {
		t.Fatalf("sequence arg = %+v, want 2026/%s", store.sequenceArg, studentCertificateClassificationCode)
	}
	wantNomor := "042/PP.00.4/MTs.20.05/V/2026"
	if store.outgoingArg.NomorSurat != wantNomor || store.outgoingArg.Tujuan != "Komite" || store.outgoingArg.Perihal != "Surat Aktif - Siswa A" {
		t.Fatalf("outgoing arg = %+v, want generated outgoing letter", store.outgoingArg)
	}
	wantNote := "Dibuat otomatis dari modul Surat Keterangan Siswa.\nTemplate: SKA\nKeperluan: Beasiswa\nCatatan: Prioritas"
	if store.outgoingArg.Catatan != wantNote || store.outgoingArg.IssuedByEmployeeID != issuedBy {
		t.Fatalf("outgoing note/issuer = %q/%v, want expected note/%v", store.outgoingArg.Catatan, store.outgoingArg.IssuedByEmployeeID, issuedBy)
	}
	if store.createArg.TemplateID != templateID || store.createArg.StudentID != studentID || store.createArg.OutgoingLetterID != outgoingID || store.createArg.CreatedByUserID != createdBy {
		t.Fatalf("create certificate arg = %+v, want linked template/student/outgoing/actor", store.createArg)
	}

	var payload map[string]any
	if err := json.Unmarshal(store.createArg.SnapshotData, &payload); err != nil {
		t.Fatalf("snapshot json unmarshal error = %v", err)
	}
	letter := payload["letter"].(map[string]any)
	if letter["nomor_surat"] != wantNomor || letter["recipient"] != "Komite" {
		t.Fatalf("snapshot letter = %+v, want generated letter data", letter)
	}
}

func TestCreateStudentCertificatePropagatesErrors(t *testing.T) {
	templateID := studentCertificateTestUUID(18)
	studentID := studentCertificateTestUUID(19)
	outgoingID := studentCertificateTestUUID(20)
	certificateID := studentCertificateTestUUID(21)
	arg := normalizedStudentCertificateInput{
		TemplateID:   templateID,
		StudentID:    studentID,
		TanggalSurat: studentCertificateTestDate(),
		Purpose:      "Beasiswa",
		Recipient:    "Komite",
	}
	newStore := func() *fakeStudentCertificateStore {
		return &fakeStudentCertificateStore{
			template: db.CertificateTemplate{
				ID:       templateID,
				Code:     "SKA",
				Name:     "Surat Aktif",
				IsActive: true,
			},
			student: db.GetCertificateStudentSnapshotRow{
				ID:     studentID,
				Nama:   "Siswa A",
				Gender: db.GenderEnumL,
				Status: db.StudentStatusEnumActive,
			},
			sequence:    7,
			outgoingRow: db.OutgoingLetter{ID: outgoingID},
			createRow:   db.StudentCertificate{ID: certificateID},
			getRow:      db.GetStudentCertificateRow{ID: certificateID, Status: "issued"},
		}
	}

	tests := []struct {
		name    string
		mutate  func(*fakeStudentCertificateStore)
		wantErr string
	}{
		{name: "template lookup", mutate: func(s *fakeStudentCertificateStore) { s.templateErr = errors.New("template failed") }, wantErr: "template failed"},
		{name: "inactive template", mutate: func(s *fakeStudentCertificateStore) { s.template.IsActive = false }, wantErr: "template surat tidak aktif"},
		{name: "student snapshot", mutate: func(s *fakeStudentCertificateStore) { s.studentErr = errors.New("student failed") }, wantErr: "student failed"},
		{name: "sequence", mutate: func(s *fakeStudentCertificateStore) { s.sequenceErr = errors.New("sequence failed") }, wantErr: "gagal generate nomor surat: sequence failed"},
		{name: "outgoing letter", mutate: func(s *fakeStudentCertificateStore) { s.outgoingErr = errors.New("outgoing failed") }, wantErr: "outgoing failed"},
		{name: "certificate create", mutate: func(s *fakeStudentCertificateStore) { s.createErr = errors.New("create failed") }, wantErr: "create failed"},
		{name: "detail lookup", mutate: func(s *fakeStudentCertificateStore) { s.getErr = errors.New("detail failed") }, wantErr: "detail failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newStore()
			tt.mutate(store)
			_, err := createStudentCertificate(context.Background(), store, arg)
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("createStudentCertificate() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestStudentCertificateCreateValidatesInputBeforeTransaction(t *testing.T) {
	svc := &StudentCertificate{}
	_, err := svc.Create(context.Background(), CreateStudentCertificateInput{TemplateID: "bad"})
	if err == nil || err.Error() != "template_id tidak valid" {
		t.Fatalf("Create(invalid input) error = %v, want template_id validation", err)
	}
}

func TestStudentCertificateCreateRequiresPoolAfterValidation(t *testing.T) {
	svc := &StudentCertificate{}
	_, err := svc.Create(context.Background(), CreateStudentCertificateInput{
		TemplateID:   "00000000-0000-0000-0000-000000000001",
		StudentID:    "00000000-0000-0000-0000-000000000002",
		TanggalSurat: "2026-05-01",
		Purpose:      "Beasiswa",
	})
	if err == nil || err.Error() != "layanan surat keterangan siswa belum siap" {
		t.Fatalf("Create(valid input without pool) error = %v, want service not ready", err)
	}
}

func TestNormalizeStudentCertificateInput(t *testing.T) {
	templateID := "00000000-0000-0000-0000-000000000001"
	studentID := "00000000-0000-0000-0000-000000000002"
	createdBy := studentCertificateTestUUID(3)
	issuedBy := studentCertificateTestUUID(4)

	arg, err := normalizeStudentCertificateInput(CreateStudentCertificateInput{
		TemplateID:         " " + templateID + " ",
		StudentID:          " " + studentID + " ",
		TanggalSurat:       " 2026-05-01 ",
		Purpose:            "  Beasiswa  ",
		Recipient:          "  Komite  ",
		Remarks:            "  Prioritas  ",
		CreatedByUserID:    createdBy,
		IssuedByEmployeeID: issuedBy,
	})
	if err != nil {
		t.Fatalf("normalizeStudentCertificateInput() error = %v", err)
	}
	if !arg.TemplateID.Valid || !arg.StudentID.Valid || !arg.TanggalSurat.Valid {
		t.Fatalf("normalizeStudentCertificateInput() = %+v, want parsed ids/date", arg)
	}
	if arg.Purpose != "Beasiswa" || arg.Recipient != "Komite" || arg.Remarks != "Prioritas" {
		t.Fatalf("normalizeStudentCertificateInput() = %+v, want trimmed text", arg)
	}
	if arg.CreatedByUserID != createdBy || arg.IssuedByEmployeeID != issuedBy {
		t.Fatalf("normalizeStudentCertificateInput() = %+v, want actor ids preserved", arg)
	}

	defaultRecipient, err := normalizeStudentCertificateInput(CreateStudentCertificateInput{
		TemplateID:   templateID,
		StudentID:    studentID,
		TanggalSurat: "2026-05-01",
		Purpose:      "Aktif sekolah",
	})
	if err != nil {
		t.Fatalf("normalizeStudentCertificateInput(default recipient) error = %v", err)
	}
	if defaultRecipient.Recipient != "Yang berkepentingan" {
		t.Fatalf("recipient = %q, want default recipient", defaultRecipient.Recipient)
	}

	tests := []struct {
		name    string
		input   CreateStudentCertificateInput
		wantErr string
	}{
		{name: "bad template", input: CreateStudentCertificateInput{TemplateID: "bad"}, wantErr: "template_id tidak valid"},
		{name: "bad student", input: CreateStudentCertificateInput{TemplateID: templateID, StudentID: "bad"}, wantErr: "student_id tidak valid"},
		{name: "bad date", input: CreateStudentCertificateInput{TemplateID: templateID, StudentID: studentID, TanggalSurat: "bad"}, wantErr: "format tanggal surat tidak valid"},
		{name: "empty purpose", input: CreateStudentCertificateInput{TemplateID: templateID, StudentID: studentID, TanggalSurat: "2026-05-01"}, wantErr: "keperluan surat wajib diisi"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := normalizeStudentCertificateInput(tt.input)
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("normalizeStudentCertificateInput() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestStudentCertificateSnapshotHelpers(t *testing.T) {
	templateID := studentCertificateTestUUID(5)
	studentID := studentCertificateTestUUID(6)
	classID := studentCertificateTestUUID(7)
	arg := normalizedStudentCertificateInput{
		TemplateID:   templateID,
		StudentID:    studentID,
		TanggalSurat: studentCertificateTestDate(),
		Purpose:      "Beasiswa",
		Recipient:    "Komite",
		Remarks:      "Prioritas",
	}

	note := certificateOutgoingNote("SKA", "Beasiswa", "  Prioritas  ")
	if note != "Dibuat otomatis dari modul Surat Keterangan Siswa.\nTemplate: SKA\nKeperluan: Beasiswa\nCatatan: Prioritas" {
		t.Fatalf("certificateOutgoingNote() = %q, want full note", note)
	}
	if note := certificateOutgoingNote("SKA", "Beasiswa", " "); note != "Dibuat otomatis dari modul Surat Keterangan Siswa.\nTemplate: SKA\nKeperluan: Beasiswa" {
		t.Fatalf("certificateOutgoingNote(no remarks) = %q, want note without catatan", note)
	}

	snapshot, err := buildStudentCertificateSnapshot(
		db.CertificateTemplate{
			ID:             templateID,
			Code:           "SKA",
			Name:           "Surat Aktif",
			DefaultPurpose: "Administrasi",
			BodyTemplate:   "Isi",
		},
		db.GetCertificateStudentSnapshotRow{
			ID:           studentID,
			Nis:          "123",
			Nisn:         "456",
			Nama:         "Siswa A",
			Gender:       db.GenderEnumL,
			Status:       db.StudentStatusEnumActive,
			ClassID:      classID,
			ClassName:    "VII A",
			ClassLevel:   7,
			ParentName:   "Orang Tua",
			ParentPhone:  "0812",
			Nik:          "1234567890123456",
			TempatLahir:  "Kolaka Utara",
			TanggalLahir: studentCertificateTestDate(),
			Alamat:       "Lasusua",
			Agama:        "Islam",
			Phone:        "0822",
		},
		"001/PP.00.4/MTs.20.05/V/2026",
		arg,
	)
	if err != nil {
		t.Fatalf("buildStudentCertificateSnapshot() error = %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(snapshot, &payload); err != nil {
		t.Fatalf("snapshot json unmarshal error = %v", err)
	}
	template := payload["template"].(map[string]any)
	student := payload["student"].(map[string]any)
	letter := payload["letter"].(map[string]any)
	if template["code"] != "SKA" || student["nama"] != "Siswa A" || letter["purpose"] != "Beasiswa" {
		t.Fatalf("snapshot payload = %+v, want template/student/letter fields", payload)
	}
	if letter["tanggal_surat"] != "2026-05-01" || student["tanggal_lahir"] != "2026-05-01" {
		t.Fatalf("snapshot dates = %v/%v, want formatted dates", letter["tanggal_surat"], student["tanggal_lahir"])
	}

	if got := uuidString(pgtype.UUID{}); got != "" {
		t.Fatalf("uuidString(invalid) = %q, want empty", got)
	}
	if got := dateString(pgtype.Date{}); got != "" {
		t.Fatalf("dateString(invalid) = %q, want empty", got)
	}
}
