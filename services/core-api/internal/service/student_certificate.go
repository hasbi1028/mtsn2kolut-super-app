package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

const studentCertificateClassificationCode = "PP.00.4"

type studentCertificateStore interface {
	ListCertificateTemplates(ctx context.Context, activeOnly bool) ([]db.CertificateTemplate, error)
	GetCertificateTemplateByID(ctx context.Context, id pgtype.UUID) (db.CertificateTemplate, error)
	ListStudentCertificateOptions(ctx context.Context, arg db.ListStudentCertificateOptionsParams) ([]db.ListStudentCertificateOptionsRow, error)
	GetCertificateStudentSnapshot(ctx context.Context, id pgtype.UUID) (db.GetCertificateStudentSnapshotRow, error)
	ListStudentCertificates(ctx context.Context, arg db.ListStudentCertificatesParams) ([]db.ListStudentCertificatesRow, error)
	GetStudentCertificate(ctx context.Context, id pgtype.UUID) (db.GetStudentCertificateRow, error)
	CreateStudentCertificate(ctx context.Context, arg db.CreateStudentCertificateParams) (db.StudentCertificate, error)
	CancelStudentCertificate(ctx context.Context, arg db.CancelStudentCertificateParams) (db.StudentCertificate, error)
	CreateOutgoingLetter(ctx context.Context, arg db.CreateOutgoingLetterParams) (db.OutgoingLetter, error)
	IssueOutgoingLetterSequence(ctx context.Context, arg db.IssueOutgoingLetterSequenceParams) (int32, error)
}

type StudentCertificate struct {
	pool *pgxpool.Pool
	q    studentCertificateStore
}

func NewStudentCertificate(pool *pgxpool.Pool) *StudentCertificate {
	return &StudentCertificate{pool: pool, q: db.New(pool)}
}

type CreateStudentCertificateInput struct {
	TemplateID         string
	StudentID          string
	TanggalSurat       string
	Purpose            string
	Recipient          string
	Remarks            string
	CreatedByUserID    pgtype.UUID
	IssuedByEmployeeID pgtype.UUID
}

func (s *StudentCertificate) ListTemplates(ctx context.Context, activeOnly bool) ([]db.CertificateTemplate, error) {
	return s.q.ListCertificateTemplates(ctx, activeOnly)
}

func (s *StudentCertificate) ListStudents(ctx context.Context, search, status string) ([]db.ListStudentCertificateOptionsRow, error) {
	status = strings.TrimSpace(status)
	switch status {
	case "", string(db.StudentStatusEnumProspective), string(db.StudentStatusEnumActive), string(db.StudentStatusEnumAlumni), string(db.StudentStatusEnumMutated):
	default:
		return nil, fmt.Errorf("status siswa tidak valid")
	}
	return s.q.ListStudentCertificateOptions(ctx, db.ListStudentCertificateOptionsParams{
		Search: strings.TrimSpace(search),
		Status: status,
	})
}

func (s *StudentCertificate) List(ctx context.Context, search, status, templateCode string) ([]db.ListStudentCertificatesRow, error) {
	status = strings.TrimSpace(status)
	switch status {
	case "", "issued", "canceled":
	default:
		return nil, fmt.Errorf("status surat tidak valid")
	}
	return s.q.ListStudentCertificates(ctx, db.ListStudentCertificatesParams{
		Search:       strings.TrimSpace(search),
		Status:       status,
		TemplateCode: strings.TrimSpace(templateCode),
	})
}

func (s *StudentCertificate) Get(ctx context.Context, id pgtype.UUID) (db.GetStudentCertificateRow, error) {
	return s.q.GetStudentCertificate(ctx, id)
}

func (s *StudentCertificate) Create(ctx context.Context, input CreateStudentCertificateInput) (db.GetStudentCertificateRow, error) {
	arg, err := normalizeStudentCertificateInput(input)
	if err != nil {
		return db.GetStudentCertificateRow{}, err
	}
	if s.pool == nil {
		return db.GetStudentCertificateRow{}, fmt.Errorf("layanan surat keterangan siswa belum siap")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return db.GetStudentCertificateRow{}, err
	}
	defer tx.Rollback(ctx)
	q := db.New(tx)

	detail, err := createStudentCertificate(ctx, q, arg)
	if err != nil {
		return db.GetStudentCertificateRow{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return db.GetStudentCertificateRow{}, err
	}
	return detail, nil
}

func createStudentCertificate(ctx context.Context, q studentCertificateStore, arg normalizedStudentCertificateInput) (db.GetStudentCertificateRow, error) {
	template, err := q.GetCertificateTemplateByID(ctx, arg.TemplateID)
	if err != nil {
		return db.GetStudentCertificateRow{}, err
	}
	if !template.IsActive {
		return db.GetStudentCertificateRow{}, fmt.Errorf("template surat tidak aktif")
	}
	student, err := q.GetCertificateStudentSnapshot(ctx, arg.StudentID)
	if err != nil {
		return db.GetStudentCertificateRow{}, err
	}

	nomorSurat, err := issueOutgoingLetterNumber(ctx, q, studentCertificateClassificationCode, arg.TanggalSurat)
	if err != nil {
		return db.GetStudentCertificateRow{}, err
	}
	outgoing, err := q.CreateOutgoingLetter(ctx, db.CreateOutgoingLetterParams{
		NomorSurat:         nomorSurat,
		ClassificationCode: studentCertificateClassificationCode,
		TanggalSurat:       arg.TanggalSurat,
		Tujuan:             arg.Recipient,
		Perihal:            fmt.Sprintf("%s - %s", template.Name, student.Nama),
		Sifat:              db.LetterSifatBiasa,
		FilePath:           "",
		Catatan:            certificateOutgoingNote(template.Code, arg.Purpose, arg.Remarks),
		IssuedByEmployeeID: arg.IssuedByEmployeeID,
	})
	if err != nil {
		return db.GetStudentCertificateRow{}, err
	}

	snapshot, err := buildStudentCertificateSnapshot(template, student, nomorSurat, arg)
	if err != nil {
		return db.GetStudentCertificateRow{}, err
	}
	certificate, err := q.CreateStudentCertificate(ctx, db.CreateStudentCertificateParams{
		TemplateID:       arg.TemplateID,
		StudentID:        arg.StudentID,
		OutgoingLetterID: outgoing.ID,
		TanggalSurat:     arg.TanggalSurat,
		Purpose:          arg.Purpose,
		Recipient:        arg.Recipient,
		Remarks:          arg.Remarks,
		SnapshotData:     snapshot,
		CreatedByUserID:  arg.CreatedByUserID,
	})
	if err != nil {
		return db.GetStudentCertificateRow{}, err
	}
	detail, err := q.GetStudentCertificate(ctx, certificate.ID)
	if err != nil {
		return db.GetStudentCertificateRow{}, err
	}
	return detail, nil
}

func (s *StudentCertificate) Cancel(ctx context.Context, id pgtype.UUID, remarks string) (db.StudentCertificate, error) {
	return s.q.CancelStudentCertificate(ctx, db.CancelStudentCertificateParams{
		ID:      id,
		Remarks: strings.TrimSpace(remarks),
	})
}

type normalizedStudentCertificateInput struct {
	TemplateID         pgtype.UUID
	StudentID          pgtype.UUID
	TanggalSurat       pgtype.Date
	Purpose            string
	Recipient          string
	Remarks            string
	CreatedByUserID    pgtype.UUID
	IssuedByEmployeeID pgtype.UUID
}

func normalizeStudentCertificateInput(input CreateStudentCertificateInput) (normalizedStudentCertificateInput, error) {
	templateID, err := parseLetterUUID(strings.TrimSpace(input.TemplateID))
	if err != nil {
		return normalizedStudentCertificateInput{}, fmt.Errorf("template_id tidak valid")
	}
	studentID, err := parseLetterUUID(strings.TrimSpace(input.StudentID))
	if err != nil {
		return normalizedStudentCertificateInput{}, fmt.Errorf("student_id tidak valid")
	}
	var tanggal pgtype.Date
	if err := tanggal.Scan(strings.TrimSpace(input.TanggalSurat)); err != nil {
		return normalizedStudentCertificateInput{}, fmt.Errorf("format tanggal surat tidak valid")
	}
	purpose := strings.TrimSpace(input.Purpose)
	if purpose == "" {
		return normalizedStudentCertificateInput{}, fmt.Errorf("keperluan surat wajib diisi")
	}
	recipient := strings.TrimSpace(input.Recipient)
	if recipient == "" {
		recipient = "Yang berkepentingan"
	}
	return normalizedStudentCertificateInput{
		TemplateID:         templateID,
		StudentID:          studentID,
		TanggalSurat:       tanggal,
		Purpose:            purpose,
		Recipient:          recipient,
		Remarks:            strings.TrimSpace(input.Remarks),
		CreatedByUserID:    input.CreatedByUserID,
		IssuedByEmployeeID: input.IssuedByEmployeeID,
	}, nil
}

func certificateOutgoingNote(templateCode, purpose, remarks string) string {
	parts := []string{"Dibuat otomatis dari modul Surat Keterangan Siswa.", "Template: " + templateCode, "Keperluan: " + purpose}
	if strings.TrimSpace(remarks) != "" {
		parts = append(parts, "Catatan: "+strings.TrimSpace(remarks))
	}
	return strings.Join(parts, "\n")
}

func buildStudentCertificateSnapshot(template db.CertificateTemplate, student db.GetCertificateStudentSnapshotRow, nomorSurat string, arg normalizedStudentCertificateInput) ([]byte, error) {
	payload := map[string]any{
		"template": map[string]any{
			"id":              uuidString(template.ID),
			"code":            template.Code,
			"name":            template.Name,
			"default_purpose": template.DefaultPurpose,
			"body_template":   template.BodyTemplate,
		},
		"student": map[string]any{
			"id":            uuidString(student.ID),
			"nis":           student.Nis,
			"nisn":          student.Nisn,
			"nama":          student.Nama,
			"gender":        string(student.Gender),
			"status":        string(student.Status),
			"class_id":      uuidString(student.ClassID),
			"class_name":    student.ClassName,
			"class_level":   student.ClassLevel,
			"parent_name":   student.ParentName,
			"parent_phone":  student.ParentPhone,
			"nik":           student.Nik,
			"tempat_lahir":  student.TempatLahir,
			"tanggal_lahir": dateString(student.TanggalLahir),
			"alamat":        student.Alamat,
			"agama":         student.Agama,
			"phone":         student.Phone,
		},
		"letter": map[string]any{
			"nomor_surat":    nomorSurat,
			"tanggal_surat":  dateString(arg.TanggalSurat),
			"classification": studentCertificateClassificationCode,
			"purpose":        arg.Purpose,
			"recipient":      arg.Recipient,
			"remarks":        arg.Remarks,
		},
		"issued_at": time.Now().Format(time.RFC3339),
	}
	return json.Marshal(payload)
}

func uuidString(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	return id.String()
}

func dateString(date pgtype.Date) string {
	if !date.Valid {
		return ""
	}
	return date.Time.Format("2006-01-02")
}
