package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type kesiswaanStore interface {
	GetKesiswaanStats(ctx context.Context) (db.GetKesiswaanStatsRow, error)
	GetKesiswaanStatsByTeacher(ctx context.Context, teacherEmployeeID pgtype.UUID) (db.GetKesiswaanStatsByTeacherRow, error)
	ListKesiswaanClassOptions(ctx context.Context) ([]db.ListKesiswaanClassOptionsRow, error)
	ListKesiswaanStudents(ctx context.Context, arg db.ListKesiswaanStudentsParams) ([]db.ListKesiswaanStudentsRow, error)
	UpdateKesiswaanStudentProfile(ctx context.Context, arg db.UpdateKesiswaanStudentProfileParams) (db.Student, error)
	UpdateKesiswaanStudentPhoto(ctx context.Context, arg db.UpdateKesiswaanStudentPhotoParams) (db.Student, error)

	ListViolationCategories(ctx context.Context, search string) ([]db.ViolationCategory, error)
	CreateViolationCategory(ctx context.Context, arg db.CreateViolationCategoryParams) (db.ViolationCategory, error)
	UpdateViolationCategory(ctx context.Context, arg db.UpdateViolationCategoryParams) (db.ViolationCategory, error)
	DeleteViolationCategory(ctx context.Context, id pgtype.UUID) error

	ListStudentViolations(ctx context.Context, arg db.ListStudentViolationsParams) ([]db.ListStudentViolationsRow, error)
	CreateStudentViolation(ctx context.Context, arg db.CreateStudentViolationParams) (db.StudentViolation, error)
	UpdateStudentViolation(ctx context.Context, arg db.UpdateStudentViolationParams) (db.StudentViolation, error)
	DeleteStudentViolation(ctx context.Context, id pgtype.UUID) error

	ListStudentAchievements(ctx context.Context, arg db.ListStudentAchievementsParams) ([]db.ListStudentAchievementsRow, error)
	CreateStudentAchievement(ctx context.Context, arg db.CreateStudentAchievementParams) (db.StudentAchievement, error)
	UpdateStudentAchievement(ctx context.Context, arg db.UpdateStudentAchievementParams) (db.StudentAchievement, error)
	DeleteStudentAchievement(ctx context.Context, id pgtype.UUID) error

	ListExtracurriculars(ctx context.Context, arg db.ListExtracurricularsParams) ([]db.ListExtracurricularsRow, error)
	CreateExtracurricular(ctx context.Context, arg db.CreateExtracurricularParams) (db.Extracurricular, error)
	UpdateExtracurricular(ctx context.Context, arg db.UpdateExtracurricularParams) (db.Extracurricular, error)
	DeleteExtracurricular(ctx context.Context, id pgtype.UUID) error

	ListExtracurricularMembers(ctx context.Context, arg db.ListExtracurricularMembersParams) ([]db.ListExtracurricularMembersRow, error)
	CreateExtracurricularMember(ctx context.Context, arg db.CreateExtracurricularMemberParams) (db.ExtracurricularMember, error)
	UpdateExtracurricularMember(ctx context.Context, arg db.UpdateExtracurricularMemberParams) (db.ExtracurricularMember, error)
	DeleteExtracurricularMember(ctx context.Context, id pgtype.UUID) error

	ListCounselingSessions(ctx context.Context, arg db.ListCounselingSessionsParams) ([]db.ListCounselingSessionsRow, error)
	CreateCounselingSession(ctx context.Context, arg db.CreateCounselingSessionParams) (db.CounselingSession, error)
	UpdateCounselingSession(ctx context.Context, arg db.UpdateCounselingSessionParams) (db.CounselingSession, error)
	DeleteCounselingSession(ctx context.Context, id pgtype.UUID) error

	ListStudentTransfers(ctx context.Context, arg db.ListStudentTransfersParams) ([]db.ListStudentTransfersRow, error)
}

type Kesiswaan struct {
	pool     *pgxpool.Pool
	q        kesiswaanStore
	photoDir string
}

func NewKesiswaan(q *db.Queries, photoDir string) *Kesiswaan {
	if strings.TrimSpace(photoDir) == "" {
		photoDir = "data/student-photos"
	}
	return &Kesiswaan{q: q, photoDir: photoDir}
}

func NewKesiswaanWithPool(pool *pgxpool.Pool, photoDir string) *Kesiswaan {
	if strings.TrimSpace(photoDir) == "" {
		photoDir = "data/student-photos"
	}
	return &Kesiswaan{pool: pool, q: db.New(pool), photoDir: photoDir}
}

type KesiswaanStats struct {
	ActiveStudents         int64 `json:"active_students"`
	ProspectiveStudents    int64 `json:"prospective_students"`
	StudentsWithPoints     int64 `json:"students_with_points"`
	TotalViolationPoints   int64 `json:"total_violation_points"`
	OpenViolations         int64 `json:"open_violations"`
	AchievementsThisYear   int64 `json:"achievements_this_year"`
	ActiveCategories       int64 `json:"active_categories"`
	ActiveExtracurriculars int64 `json:"active_extracurriculars"`
	OpenCounselingSessions int64 `json:"open_counseling_sessions"`
	TransfersThisYear      int64 `json:"transfers_this_year"`
}

type UploadStudentPhotoInput struct {
	StudentID    pgtype.UUID
	OriginalName string
	MimeType     string
	FileSize     int64
	File         io.Reader
}

func (s *Kesiswaan) Stats(ctx context.Context, teacherEmployeeID pgtype.UUID) (KesiswaanStats, error) {
	if teacherEmployeeID.Valid {
		row, err := s.q.GetKesiswaanStatsByTeacher(ctx, teacherEmployeeID)
		if err != nil {
			return KesiswaanStats{}, err
		}
		return KesiswaanStats(row), nil
	}
	row, err := s.q.GetKesiswaanStats(ctx)
	if err != nil {
		return KesiswaanStats{}, err
	}
	return KesiswaanStats(row), nil
}

func (s *Kesiswaan) ClassOptions(ctx context.Context) ([]db.ListKesiswaanClassOptionsRow, error) {
	return s.q.ListKesiswaanClassOptions(ctx)
}

func (s *Kesiswaan) ListStudents(ctx context.Context, search, status, classID string, teacherEmployeeID pgtype.UUID) ([]db.ListKesiswaanStudentsRow, error) {
	parsedClassID, err := ParseKesiswaanOptionalUUID(classID)
	if err != nil {
		return nil, err
	}
	status = normalizeKesiswaanText(status, "")
	if status != "" && !validKesiswaanStudentStatuses[status] {
		return nil, fmt.Errorf("status siswa tidak valid")
	}
	return s.q.ListKesiswaanStudents(ctx, db.ListKesiswaanStudentsParams{
		Search:            strings.TrimSpace(search),
		Status:            status,
		ClassID:           parsedClassID,
		TeacherEmployeeID: teacherEmployeeID,
	})
}

func (s *Kesiswaan) UpdateStudentProfile(ctx context.Context, arg db.UpdateKesiswaanStudentProfileParams) (db.Student, error) {
	arg.Nik = strings.TrimSpace(arg.Nik)
	arg.TempatLahir = strings.TrimSpace(arg.TempatLahir)
	arg.Alamat = strings.TrimSpace(arg.Alamat)
	arg.Agama = strings.TrimSpace(arg.Agama)
	arg.Phone = strings.TrimSpace(arg.Phone)
	arg.ParentName = strings.TrimSpace(arg.ParentName)
	arg.ParentPhone = strings.TrimSpace(arg.ParentPhone)
	if err := validateKesiswaanStudentProfile(arg); err != nil {
		return db.Student{}, err
	}
	return s.q.UpdateKesiswaanStudentProfile(ctx, arg)
}

func (s *Kesiswaan) SaveStudentPhoto(ctx context.Context, input UploadStudentPhotoInput) (db.Student, error) {
	if err := validateStudentPhoto(input); err != nil {
		return db.Student{}, err
	}
	if err := os.MkdirAll(s.photoDir, 0o755); err != nil {
		return db.Student{}, err
	}
	token, err := randomHex(12)
	if err != nil {
		return db.Student{}, err
	}
	safeName := sanitizeFilename(input.OriginalName)
	storedName := token + "_" + safeName
	absPath, err := filepath.Abs(filepath.Join(s.photoDir, storedName))
	if err != nil {
		return db.Student{}, err
	}
	out, err := os.Create(absPath)
	if err != nil {
		return db.Student{}, err
	}
	defer out.Close()
	if _, err := io.Copy(out, input.File); err != nil {
		return db.Student{}, err
	}
	return s.q.UpdateKesiswaanStudentPhoto(ctx, db.UpdateKesiswaanStudentPhotoParams{
		ID:       input.StudentID,
		PhotoUrl: "/api/kesiswaan/student-photos/" + storedName,
	})
}

func (s *Kesiswaan) StudentPhotoPath(filename string) (string, bool) {
	filename = strings.TrimSpace(filename)
	if filename == "" || strings.Contains(filename, "/") || strings.Contains(filename, "..") {
		return "", false
	}
	return filepath.Join(s.photoDir, filename), true
}

func (s *Kesiswaan) ListCategories(ctx context.Context, search string) ([]db.ViolationCategory, error) {
	return s.q.ListViolationCategories(ctx, strings.TrimSpace(search))
}

func (s *Kesiswaan) CreateCategory(ctx context.Context, arg db.CreateViolationCategoryParams) (db.ViolationCategory, error) {
	arg = normalizeViolationCategoryCreate(arg)
	if err := validateViolationCategory(arg.Code, arg.Name, arg.Point, arg.Severity); err != nil {
		return db.ViolationCategory{}, err
	}
	return s.q.CreateViolationCategory(ctx, arg)
}

func (s *Kesiswaan) UpdateCategory(ctx context.Context, arg db.UpdateViolationCategoryParams) (db.ViolationCategory, error) {
	arg.Code = strings.TrimSpace(arg.Code)
	arg.Name = strings.TrimSpace(arg.Name)
	arg.Severity = normalizeKesiswaanText(arg.Severity, "ringan")
	arg.Description = strings.TrimSpace(arg.Description)
	if err := validateViolationCategory(arg.Code, arg.Name, arg.Point, arg.Severity); err != nil {
		return db.ViolationCategory{}, err
	}
	return s.q.UpdateViolationCategory(ctx, arg)
}

func (s *Kesiswaan) DeleteCategory(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteViolationCategory(ctx, id)
}

func (s *Kesiswaan) ListViolations(ctx context.Context, search, status, studentID string, teacherEmployeeID pgtype.UUID) ([]db.ListStudentViolationsRow, error) {
	parsedStudentID, err := ParseKesiswaanOptionalUUID(studentID)
	if err != nil {
		return nil, err
	}
	status = normalizeKesiswaanText(status, "")
	if status != "" && !validStudentViolationStatuses[status] {
		return nil, fmt.Errorf("status pelanggaran tidak valid")
	}
	return s.q.ListStudentViolations(ctx, db.ListStudentViolationsParams{
		Search:            strings.TrimSpace(search),
		Status:            status,
		StudentID:         parsedStudentID,
		TeacherEmployeeID: teacherEmployeeID,
	})
}

func (s *Kesiswaan) CreateViolation(ctx context.Context, arg db.CreateStudentViolationParams) (db.StudentViolation, error) {
	arg = normalizeStudentViolationCreate(arg)
	if err := validateStudentViolation(arg.StudentID, arg.IncidentDate, arg.Points, arg.Status); err != nil {
		return db.StudentViolation{}, err
	}
	return s.q.CreateStudentViolation(ctx, arg)
}

func (s *Kesiswaan) UpdateViolation(ctx context.Context, arg db.UpdateStudentViolationParams) (db.StudentViolation, error) {
	arg.Description = strings.TrimSpace(arg.Description)
	arg.ActionTaken = strings.TrimSpace(arg.ActionTaken)
	arg.Status = normalizeKesiswaanText(arg.Status, "open")
	if err := validateStudentViolation(arg.StudentID, arg.IncidentDate, arg.Points, arg.Status); err != nil {
		return db.StudentViolation{}, err
	}
	return s.q.UpdateStudentViolation(ctx, arg)
}

func (s *Kesiswaan) DeleteViolation(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteStudentViolation(ctx, id)
}

func (s *Kesiswaan) ListAchievements(ctx context.Context, search, level, studentID string, teacherEmployeeID pgtype.UUID) ([]db.ListStudentAchievementsRow, error) {
	parsedStudentID, err := ParseKesiswaanOptionalUUID(studentID)
	if err != nil {
		return nil, err
	}
	level = normalizeKesiswaanText(level, "")
	if level != "" && !validAchievementLevels[level] {
		return nil, fmt.Errorf("tingkat prestasi tidak valid")
	}
	return s.q.ListStudentAchievements(ctx, db.ListStudentAchievementsParams{
		Search:            strings.TrimSpace(search),
		Level:             level,
		StudentID:         parsedStudentID,
		TeacherEmployeeID: teacherEmployeeID,
	})
}

func (s *Kesiswaan) CreateAchievement(ctx context.Context, arg db.CreateStudentAchievementParams) (db.StudentAchievement, error) {
	arg = normalizeStudentAchievementCreate(arg)
	if err := validateStudentAchievement(arg.StudentID, arg.AchievementDate, arg.Title, arg.Level); err != nil {
		return db.StudentAchievement{}, err
	}
	return s.q.CreateStudentAchievement(ctx, arg)
}

func (s *Kesiswaan) UpdateAchievement(ctx context.Context, arg db.UpdateStudentAchievementParams) (db.StudentAchievement, error) {
	arg.Title = strings.TrimSpace(arg.Title)
	arg.Level = normalizeKesiswaanText(arg.Level, "school")
	arg.Category = strings.TrimSpace(arg.Category)
	arg.Organizer = strings.TrimSpace(arg.Organizer)
	arg.Description = strings.TrimSpace(arg.Description)
	arg.DocumentUrl = strings.TrimSpace(arg.DocumentUrl)
	if err := validateStudentAchievement(arg.StudentID, arg.AchievementDate, arg.Title, arg.Level); err != nil {
		return db.StudentAchievement{}, err
	}
	return s.q.UpdateStudentAchievement(ctx, arg)
}

func (s *Kesiswaan) DeleteAchievement(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteStudentAchievement(ctx, id)
}

func (s *Kesiswaan) ListExtracurriculars(ctx context.Context, search string, activeOnly bool) ([]db.ListExtracurricularsRow, error) {
	return s.q.ListExtracurriculars(ctx, db.ListExtracurricularsParams{
		Search:     strings.TrimSpace(search),
		ActiveOnly: activeOnly,
	})
}

func (s *Kesiswaan) CreateExtracurricular(ctx context.Context, arg db.CreateExtracurricularParams) (db.Extracurricular, error) {
	arg = normalizeExtracurricularCreate(arg)
	if err := validateExtracurricular(arg.Code, arg.Name); err != nil {
		return db.Extracurricular{}, err
	}
	return s.q.CreateExtracurricular(ctx, arg)
}

func (s *Kesiswaan) UpdateExtracurricular(ctx context.Context, arg db.UpdateExtracurricularParams) (db.Extracurricular, error) {
	arg.Code = strings.TrimSpace(arg.Code)
	arg.Name = strings.TrimSpace(arg.Name)
	arg.Category = strings.TrimSpace(arg.Category)
	arg.Description = strings.TrimSpace(arg.Description)
	arg.ScheduleText = strings.TrimSpace(arg.ScheduleText)
	if err := validateExtracurricular(arg.Code, arg.Name); err != nil {
		return db.Extracurricular{}, err
	}
	return s.q.UpdateExtracurricular(ctx, arg)
}

func (s *Kesiswaan) DeleteExtracurricular(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteExtracurricular(ctx, id)
}

func (s *Kesiswaan) ListExtracurricularMembers(ctx context.Context, search, status, extracurricularID, studentID string, teacherEmployeeID pgtype.UUID) ([]db.ListExtracurricularMembersRow, error) {
	parsedExtracurricularID, err := ParseKesiswaanOptionalUUID(extracurricularID)
	if err != nil {
		return nil, err
	}
	parsedStudentID, err := ParseKesiswaanOptionalUUID(studentID)
	if err != nil {
		return nil, err
	}
	status = normalizeKesiswaanText(status, "")
	if status != "" && !validExtracurricularMemberStatuses[status] {
		return nil, fmt.Errorf("status anggota ekskul tidak valid")
	}
	return s.q.ListExtracurricularMembers(ctx, db.ListExtracurricularMembersParams{
		Search:            strings.TrimSpace(search),
		Status:            status,
		ExtracurricularID: parsedExtracurricularID,
		StudentID:         parsedStudentID,
		TeacherEmployeeID: teacherEmployeeID,
	})
}

func (s *Kesiswaan) CreateExtracurricularMember(ctx context.Context, arg db.CreateExtracurricularMemberParams) (db.ExtracurricularMember, error) {
	arg = normalizeExtracurricularMemberCreate(arg)
	if err := validateExtracurricularMember(arg.ExtracurricularID, arg.StudentID, arg.JoinedAt, arg.Role, arg.Status); err != nil {
		return db.ExtracurricularMember{}, err
	}
	return s.q.CreateExtracurricularMember(ctx, arg)
}

func (s *Kesiswaan) UpdateExtracurricularMember(ctx context.Context, arg db.UpdateExtracurricularMemberParams) (db.ExtracurricularMember, error) {
	arg.Role = normalizeKesiswaanText(arg.Role, "member")
	arg.Status = normalizeKesiswaanText(arg.Status, "active")
	arg.Notes = strings.TrimSpace(arg.Notes)
	if err := validateExtracurricularMember(arg.ExtracurricularID, arg.StudentID, arg.JoinedAt, arg.Role, arg.Status); err != nil {
		return db.ExtracurricularMember{}, err
	}
	return s.q.UpdateExtracurricularMember(ctx, arg)
}

func (s *Kesiswaan) DeleteExtracurricularMember(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteExtracurricularMember(ctx, id)
}

func (s *Kesiswaan) ListCounselingSessions(ctx context.Context, search, status, studentID string, canReadConfidential bool, teacherEmployeeID pgtype.UUID) ([]db.ListCounselingSessionsRow, error) {
	parsedStudentID, err := ParseKesiswaanOptionalUUID(studentID)
	if err != nil {
		return nil, err
	}
	status = normalizeKesiswaanText(status, "")
	if status != "" && !validCounselingStatuses[status] {
		return nil, fmt.Errorf("status konseling tidak valid")
	}
	return s.q.ListCounselingSessions(ctx, db.ListCounselingSessionsParams{
		Search:              strings.TrimSpace(search),
		Status:              status,
		StudentID:           parsedStudentID,
		CanReadConfidential: canReadConfidential,
		TeacherEmployeeID:   teacherEmployeeID,
	})
}

func (s *Kesiswaan) CreateCounselingSession(ctx context.Context, arg db.CreateCounselingSessionParams) (db.CounselingSession, error) {
	arg = normalizeCounselingCreate(arg)
	if err := validateCounselingSession(arg.StudentID, arg.SessionDate, arg.Topic, arg.Status); err != nil {
		return db.CounselingSession{}, err
	}
	return s.q.CreateCounselingSession(ctx, arg)
}

func (s *Kesiswaan) UpdateCounselingSession(ctx context.Context, arg db.UpdateCounselingSessionParams) (db.CounselingSession, error) {
	arg.Topic = strings.TrimSpace(arg.Topic)
	arg.Summary = strings.TrimSpace(arg.Summary)
	arg.FollowUp = strings.TrimSpace(arg.FollowUp)
	arg.Status = normalizeKesiswaanText(arg.Status, "open")
	if err := validateCounselingSession(arg.StudentID, arg.SessionDate, arg.Topic, arg.Status); err != nil {
		return db.CounselingSession{}, err
	}
	return s.q.UpdateCounselingSession(ctx, arg)
}

func (s *Kesiswaan) DeleteCounselingSession(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteCounselingSession(ctx, id)
}

func (s *Kesiswaan) ListStudentTransfers(ctx context.Context, search, transferType, studentID string, teacherEmployeeID pgtype.UUID) ([]db.ListStudentTransfersRow, error) {
	parsedStudentID, err := ParseKesiswaanOptionalUUID(studentID)
	if err != nil {
		return nil, err
	}
	transferType = normalizeKesiswaanText(transferType, "")
	if transferType != "" && !validStudentTransferTypes[transferType] {
		return nil, fmt.Errorf("jenis mutasi tidak valid")
	}
	return s.q.ListStudentTransfers(ctx, db.ListStudentTransfersParams{
		Search:            strings.TrimSpace(search),
		TransferType:      transferType,
		StudentID:         parsedStudentID,
		TeacherEmployeeID: teacherEmployeeID,
	})
}

func (s *Kesiswaan) CreateStudentTransfer(ctx context.Context, arg db.CreateStudentTransferParams) (db.StudentTransfer, error) {
	arg = normalizeStudentTransferCreate(arg)
	if err := validateStudentTransfer(arg); err != nil {
		return db.StudentTransfer{}, err
	}
	if s.pool == nil {
		return db.StudentTransfer{}, fmt.Errorf("layanan mutasi siswa belum siap")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return db.StudentTransfer{}, err
	}
	defer tx.Rollback(ctx)
	q := db.New(tx)

	row, err := q.CreateStudentTransfer(ctx, arg)
	if err != nil {
		return db.StudentTransfer{}, err
	}
	status := db.StudentStatusEnumActive
	isActive := true
	if arg.TransferType == "out" {
		status = db.StudentStatusEnumMutated
		isActive = false
	}
	if _, err := q.UpdateKesiswaanStudentLifecycle(ctx, db.UpdateKesiswaanStudentLifecycleParams{
		ID:       arg.StudentID,
		Status:   status,
		IsActive: isActive,
	}); err != nil {
		return db.StudentTransfer{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return db.StudentTransfer{}, err
	}
	return row, nil
}

func ParseKesiswaanOptionalUUID(raw string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if strings.TrimSpace(raw) == "" {
		return id, nil
	}
	if err := id.Scan(strings.TrimSpace(raw)); err != nil {
		return pgtype.UUID{}, fmt.Errorf("id tidak valid")
	}
	return id, nil
}

func ParseKesiswaanDate(raw string) (pgtype.Date, error) {
	var date pgtype.Date
	if strings.TrimSpace(raw) == "" {
		return date, fmt.Errorf("tanggal wajib diisi")
	}
	if err := date.Scan(strings.TrimSpace(raw)); err != nil {
		return pgtype.Date{}, fmt.Errorf("format tanggal tidak valid")
	}
	return date, nil
}

func ParseKesiswaanOptionalDate(raw string) (pgtype.Date, error) {
	var date pgtype.Date
	if strings.TrimSpace(raw) == "" {
		return date, nil
	}
	if err := date.Scan(strings.TrimSpace(raw)); err != nil {
		return pgtype.Date{}, fmt.Errorf("format tanggal tidak valid")
	}
	return date, nil
}

func KesiswaanOptionalInt4(value *int32) pgtype.Int4 {
	if value == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *value, Valid: true}
}

func normalizeKesiswaanText(value, fallback string) string {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func normalizeViolationCategoryCreate(arg db.CreateViolationCategoryParams) db.CreateViolationCategoryParams {
	arg.Code = strings.TrimSpace(arg.Code)
	arg.Name = strings.TrimSpace(arg.Name)
	arg.Severity = normalizeKesiswaanText(arg.Severity, "ringan")
	arg.Description = strings.TrimSpace(arg.Description)
	return arg
}

func normalizeStudentViolationCreate(arg db.CreateStudentViolationParams) db.CreateStudentViolationParams {
	arg.Description = strings.TrimSpace(arg.Description)
	arg.ActionTaken = strings.TrimSpace(arg.ActionTaken)
	arg.Status = normalizeKesiswaanText(arg.Status, "open")
	if !arg.IncidentDate.Valid {
		arg.IncidentDate = pgtype.Date{Time: time.Now(), Valid: true}
	}
	return arg
}

func normalizeStudentAchievementCreate(arg db.CreateStudentAchievementParams) db.CreateStudentAchievementParams {
	arg.Title = strings.TrimSpace(arg.Title)
	arg.Level = normalizeKesiswaanText(arg.Level, "school")
	arg.Category = strings.TrimSpace(arg.Category)
	arg.Organizer = strings.TrimSpace(arg.Organizer)
	arg.Description = strings.TrimSpace(arg.Description)
	arg.DocumentUrl = strings.TrimSpace(arg.DocumentUrl)
	if !arg.AchievementDate.Valid {
		arg.AchievementDate = pgtype.Date{Time: time.Now(), Valid: true}
	}
	return arg
}

func normalizeExtracurricularCreate(arg db.CreateExtracurricularParams) db.CreateExtracurricularParams {
	arg.Code = strings.TrimSpace(arg.Code)
	arg.Name = strings.TrimSpace(arg.Name)
	arg.Category = strings.TrimSpace(arg.Category)
	arg.Description = strings.TrimSpace(arg.Description)
	arg.ScheduleText = strings.TrimSpace(arg.ScheduleText)
	return arg
}

func normalizeExtracurricularMemberCreate(arg db.CreateExtracurricularMemberParams) db.CreateExtracurricularMemberParams {
	arg.Role = normalizeKesiswaanText(arg.Role, "member")
	arg.Status = normalizeKesiswaanText(arg.Status, "active")
	arg.Notes = strings.TrimSpace(arg.Notes)
	if !arg.JoinedAt.Valid {
		arg.JoinedAt = pgtype.Date{Time: time.Now(), Valid: true}
	}
	return arg
}

func normalizeCounselingCreate(arg db.CreateCounselingSessionParams) db.CreateCounselingSessionParams {
	arg.Topic = strings.TrimSpace(arg.Topic)
	arg.Summary = strings.TrimSpace(arg.Summary)
	arg.FollowUp = strings.TrimSpace(arg.FollowUp)
	arg.Status = normalizeKesiswaanText(arg.Status, "open")
	if !arg.SessionDate.Valid {
		arg.SessionDate = pgtype.Date{Time: time.Now(), Valid: true}
	}
	return arg
}

func normalizeStudentTransferCreate(arg db.CreateStudentTransferParams) db.CreateStudentTransferParams {
	arg.TransferType = normalizeKesiswaanText(arg.TransferType, "out")
	arg.PreviousSchool = strings.TrimSpace(arg.PreviousSchool)
	arg.DestinationSchool = strings.TrimSpace(arg.DestinationSchool)
	arg.Reason = strings.TrimSpace(arg.Reason)
	arg.DocumentRef = strings.TrimSpace(arg.DocumentRef)
	arg.Notes = strings.TrimSpace(arg.Notes)
	if !arg.TransferDate.Valid {
		arg.TransferDate = pgtype.Date{Time: time.Now(), Valid: true}
	}
	return arg
}

func validateKesiswaanStudentProfile(arg db.UpdateKesiswaanStudentProfileParams) error {
	if arg.Nik != "" {
		if len(arg.Nik) != 16 {
			return fmt.Errorf("NIK harus 16 digit")
		}
		for _, r := range arg.Nik {
			if !unicode.IsDigit(r) {
				return fmt.Errorf("NIK hanya boleh berisi angka")
			}
		}
	}
	if arg.AnakKe.Valid && arg.AnakKe.Int32 <= 0 {
		return fmt.Errorf("anak ke harus lebih dari 0")
	}
	return nil
}

func validateStudentPhoto(input UploadStudentPhotoInput) error {
	if !input.StudentID.Valid {
		return fmt.Errorf("siswa wajib dipilih")
	}
	if input.File == nil {
		return fmt.Errorf("file wajib diisi")
	}
	if input.FileSize <= 0 {
		return fmt.Errorf("ukuran file tidak valid")
	}
	if input.FileSize > 3*1024*1024 {
		return fmt.Errorf("ukuran foto maksimal 3MB")
	}
	if !strings.HasPrefix(input.MimeType, "image/") {
		return fmt.Errorf("hanya file gambar yang diperbolehkan")
	}
	return nil
}

func validateViolationCategory(code, name string, point int32, severity string) error {
	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("kode kategori wajib diisi")
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("nama kategori wajib diisi")
	}
	if point < 0 {
		return fmt.Errorf("poin kategori tidak valid")
	}
	if !validViolationSeverities[severity] {
		return fmt.Errorf("tingkat pelanggaran tidak valid")
	}
	return nil
}

func validateStudentViolation(studentID pgtype.UUID, incidentDate pgtype.Date, points int32, status string) error {
	if !studentID.Valid {
		return fmt.Errorf("siswa wajib dipilih")
	}
	if !incidentDate.Valid {
		return fmt.Errorf("tanggal kejadian wajib diisi")
	}
	if points < 0 {
		return fmt.Errorf("poin pelanggaran tidak valid")
	}
	if !validStudentViolationStatuses[status] {
		return fmt.Errorf("status pelanggaran tidak valid")
	}
	return nil
}

func validateStudentAchievement(studentID pgtype.UUID, achievementDate pgtype.Date, title, level string) error {
	if !studentID.Valid {
		return fmt.Errorf("siswa wajib dipilih")
	}
	if !achievementDate.Valid {
		return fmt.Errorf("tanggal prestasi wajib diisi")
	}
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("judul prestasi wajib diisi")
	}
	if !validAchievementLevels[level] {
		return fmt.Errorf("tingkat prestasi tidak valid")
	}
	return nil
}

func validateExtracurricular(code, name string) error {
	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("kode ekskul wajib diisi")
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("nama ekskul wajib diisi")
	}
	return nil
}

func validateExtracurricularMember(extracurricularID, studentID pgtype.UUID, joinedAt pgtype.Date, role, status string) error {
	if !extracurricularID.Valid {
		return fmt.Errorf("ekskul wajib dipilih")
	}
	if !studentID.Valid {
		return fmt.Errorf("siswa wajib dipilih")
	}
	if !joinedAt.Valid {
		return fmt.Errorf("tanggal bergabung wajib diisi")
	}
	if !validExtracurricularMemberRoles[role] {
		return fmt.Errorf("peran anggota ekskul tidak valid")
	}
	if !validExtracurricularMemberStatuses[status] {
		return fmt.Errorf("status anggota ekskul tidak valid")
	}
	return nil
}

func validateCounselingSession(studentID pgtype.UUID, sessionDate pgtype.Date, topic, status string) error {
	if !studentID.Valid {
		return fmt.Errorf("siswa wajib dipilih")
	}
	if !sessionDate.Valid {
		return fmt.Errorf("tanggal konseling wajib diisi")
	}
	if strings.TrimSpace(topic) == "" {
		return fmt.Errorf("topik konseling wajib diisi")
	}
	if !validCounselingStatuses[status] {
		return fmt.Errorf("status konseling tidak valid")
	}
	return nil
}

func validateStudentTransfer(arg db.CreateStudentTransferParams) error {
	if !arg.StudentID.Valid {
		return fmt.Errorf("siswa wajib dipilih")
	}
	if !arg.TransferDate.Valid {
		return fmt.Errorf("tanggal mutasi wajib diisi")
	}
	if !validStudentTransferTypes[arg.TransferType] {
		return fmt.Errorf("jenis mutasi tidak valid")
	}
	if arg.TransferType == "out" && arg.DestinationSchool == "" {
		return fmt.Errorf("sekolah tujuan wajib diisi untuk mutasi keluar")
	}
	if arg.TransferType == "in" && arg.PreviousSchool == "" {
		return fmt.Errorf("sekolah asal wajib diisi untuk mutasi masuk")
	}
	return nil
}

var validKesiswaanStudentStatuses = map[string]bool{
	"prospective": true,
	"active":      true,
	"alumni":      true,
	"mutated":     true,
}

var validViolationSeverities = map[string]bool{
	"ringan": true,
	"sedang": true,
	"berat":  true,
}

var validStudentViolationStatuses = map[string]bool{
	"open":     true,
	"resolved": true,
	"canceled": true,
}

var validAchievementLevels = map[string]bool{
	"school":        true,
	"district":      true,
	"province":      true,
	"national":      true,
	"international": true,
}

var validExtracurricularMemberRoles = map[string]bool{
	"member":    true,
	"leader":    true,
	"assistant": true,
}

var validExtracurricularMemberStatuses = map[string]bool{
	"active":   true,
	"inactive": true,
	"alumni":   true,
}

var validCounselingStatuses = map[string]bool{
	"open":       true,
	"monitoring": true,
	"resolved":   true,
	"referred":   true,
	"canceled":   true,
}

var validStudentTransferTypes = map[string]bool{
	"in":  true,
	"out": true,
}
