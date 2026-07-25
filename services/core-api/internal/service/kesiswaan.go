package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type KesiswaanService struct {
	q *db.Queries
}

func NewKesiswaanService(q *db.Queries) *KesiswaanService {
	return &KesiswaanService{q: q}
}

type StudentItem struct {
	ID                   string     `json:"id"`
	NIS                  string     `json:"nis"`
	NISN                 string     `json:"nisn"`
	Nama                 string     `json:"nama"`
	Gender               string     `json:"gender"`
	ParentName           string     `json:"parent_name"`
	ParentPhone          string     `json:"parent_phone"`
	ClassID              string     `json:"class_id,omitempty"`
	ClassName            string     `json:"class_name"`
	ClassCode            string     `json:"class_code"`
	IsActive             bool       `json:"is_active"`
	Status               string     `json:"status"`
	NIK                  string     `json:"nik"`
	TempatLahir          string     `json:"tempat_lahir"`
	TanggalLahir         *time.Time `json:"tanggal_lahir,omitempty"`
	Alamat               string     `json:"alamat"`
	Agama                string     `json:"agama"`
	AnakKe               *int32     `json:"anak_ke,omitempty"`
	Phone                string     `json:"phone"`
	PhotoURL             string     `json:"photo_url"`
	TotalViolationPoints int32      `json:"total_violation_points"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

func (s *KesiswaanService) ListStudents(ctx context.Context, search, status, classID string) ([]StudentItem, error) {
	var cid pgtype.UUID
	if classID != "" {
		cid = pgUUID(classID)
	}
	rows, err := s.q.ListKesiswaanStudents(ctx, db.ListKesiswaanStudentsParams{
		Search:            search,
		Status:            status,
		ClassID:           cid,
		TeacherEmployeeID: pgUUID("00000000-0000-0000-0000-000000000000"),
	})
	if err != nil {
		return nil, fmt.Errorf("list students: %w", err)
	}
	items := make([]StudentItem, len(rows))
	for i, r := range rows {
		item := StudentItem{
			ID: pgUUIDString(r.ID), NIS: r.Nis, NISN: r.Nisn, Nama: r.Nama,
			Gender: string(r.Gender), ParentName: r.ParentName, ParentPhone: r.ParentPhone,
			IsActive: r.IsActive, Status: string(r.Status),
			NIK: r.Nik, TempatLahir: r.TempatLahir, Alamat: r.Alamat,
			Agama: r.Agama, Phone: r.Phone, PhotoURL: r.PhotoUrl,
			TotalViolationPoints: r.TotalViolationPoints,
			CreatedAt: r.CreatedAt.Time, UpdatedAt: r.UpdatedAt.Time,
			ClassName: r.ClassName.String, ClassCode: r.ClassCode.String,
		}
		if r.ClassID.Valid {
			item.ClassID = pgUUIDString(r.ClassID)
		}
		if r.TanggalLahir.Valid {
			t := r.TanggalLahir.Time
			item.TanggalLahir = &t
		}
		if r.AnakKe.Valid {
			v := r.AnakKe.Int32
			item.AnakKe = &v
		}
		items[i] = item
	}
	return items, nil
}

func (s *KesiswaanService) UpdateProfile(ctx context.Context, id, nik, tempatLahir, alamat, agama, phone, parentName, parentPhone string, tanggalLahir string, anakKe int) error {
	var tl pgtype.Date
	if tanggalLahir != "" {
		parsed, err := time.Parse("2006-01-02", tanggalLahir)
		if err == nil {
			tl = pgtype.Date{Time: parsed, Valid: true}
		}
	}
	ak := pgtype.Int4{Int32: int32(anakKe), Valid: true}
	if anakKe <= 0 {
		ak.Valid = false
	}
	_, err := s.q.UpdateKesiswaanStudentProfile(ctx, db.UpdateKesiswaanStudentProfileParams{
		ID:           pgUUID(id),
		Nik:          nik,
		TempatLahir:  tempatLahir,
		TanggalLahir: tl,
		Alamat:       alamat,
		Agama:        agama,
		AnakKe:       ak,
		Phone:        phone,
		ParentName:   parentName,
		ParentPhone:  parentPhone,
	})
	if err != nil {
		return fmt.Errorf("update student profile: %w", err)
	}
	return nil
}

// ─── Create Student ───

type CreateStudentRequest struct {
	NIS          string `json:"nis"`
	NISN         string `json:"nisn"`
	Nama         string `json:"nama"`
	Gender       string `json:"gender"`
	ClassID      string `json:"class_id"`
	Status       string `json:"status"`
	NIK          string `json:"nik"`
	TempatLahir  string `json:"tempat_lahir"`
	TanggalLahir string `json:"tanggal_lahir"`
	Alamat       string `json:"alamat"`
	Agama        string `json:"agama"`
	Phone        string `json:"phone"`
	ParentName   string `json:"parent_name"`
	ParentPhone  string `json:"parent_phone"`
}

func (s *KesiswaanService) CreateStudent(ctx context.Context, req CreateStudentRequest) (string, error) {
	var classID pgtype.UUID
	if req.ClassID != "" {
		classID = pgUUID(req.ClassID)
	}
	var tgl pgtype.Date
	if req.TanggalLahir != "" {
		parsed, err := time.Parse("2006-01-02", req.TanggalLahir)
		if err == nil {
			tgl = pgtype.Date{Time: parsed, Valid: true}
		}
	}
	status := req.Status
	if status == "" {
		status = "active"
	}
	id, err := s.q.CreateKesiswaanStudent(ctx, db.CreateKesiswaanStudentParams{
		NIS:          req.NIS,
		NISN:         req.NISN,
		Nama:         req.Nama,
		Gender:       req.Gender,
		ClassID:      classID,
		Status:       status,
		IsActive:     status == "active",
		NIK:          req.NIK,
		TempatLahir:  req.TempatLahir,
		TanggalLahir: tgl,
		Alamat:       req.Alamat,
		Agama:        req.Agama,
		Phone:        req.Phone,
		ParentName:   req.ParentName,
		ParentPhone:  req.ParentPhone,
	})
	if err != nil {
		return "", fmt.Errorf("create student: %w", err)
	}
	return pgUUIDString(id), nil
}

// ─── Delete Student ───

func (s *KesiswaanService) DeleteStudent(ctx context.Context, id string) error {
	err := s.q.DeleteKesiswaanStudent(ctx, pgUUID(id))
	if err != nil {
		return fmt.Errorf("delete student: %w", err)
	}
	return nil
}
