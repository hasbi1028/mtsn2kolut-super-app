package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type SemesterService struct {
	q  *db.Queries
	db *pgxpool.Pool
}

func NewSemesterService(q *db.Queries, pool *pgxpool.Pool) *SemesterService {
	return &SemesterService{q: q, db: pool}
}

type Semester struct {
	ID               string    `json:"id"`
	AcademicYearID   string    `json:"academic_year_id"`
	AcademicYearName string    `json:"academic_year_name"`
	Name             string    `json:"name"`
	Label            string    `json:"label"`
	StartDate        string    `json:"start_date"`
	EndDate          string    `json:"end_date"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type SemesterCreateParams struct {
	AcademicYearID string
	Name           string
	Label          string
	StartDate      string
	EndDate        string
	IsActive       bool
}

func pgUUID(s string) pgtype.UUID {
	var u pgtype.UUID
	if s == "" {
		return u
	}
	_ = u.Scan(s)
	return u
}

func dateString(d pgtype.Date) string {
	if !d.Valid {
		return ""
	}
	return d.Time.Format("2006-01-02")
}

func timestamptzTime(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

func rowToSemester(id pgtype.UUID, academicYearID pgtype.UUID, academicYearName string, name string, label string, startDate pgtype.Date, endDate pgtype.Date, isActive bool, createdAt pgtype.Timestamptz, updatedAt pgtype.Timestamptz) Semester {
	return Semester{
		ID:               pgUUIDString(id),
		AcademicYearID:   pgUUIDString(academicYearID),
		AcademicYearName: academicYearName,
		Name:             name,
		Label:            label,
		StartDate:        dateString(startDate),
		EndDate:          dateString(endDate),
		IsActive:         isActive,
		CreatedAt:        timestamptzTime(createdAt),
		UpdatedAt:        timestamptzTime(updatedAt),
	}
}

func (s *SemesterService) List(ctx context.Context) ([]Semester, error) {
	rows, err := s.q.ListSemesters(ctx)
	if err != nil {
		return nil, fmt.Errorf("list semesters: %w", err)
	}
	result := make([]Semester, len(rows))
	for i, row := range rows {
		result[i] = rowToSemester(row.ID, row.AcademicYearID, row.AcademicYearName, row.Name, row.Label, row.StartDate, row.EndDate, row.IsActive, row.CreatedAt, row.UpdatedAt)
	}
	return result, nil
}

func (s *SemesterService) Get(ctx context.Context, id string) (*Semester, error) {
	row, err := s.q.GetSemester(ctx, pgUUID(id))
	if err != nil {
		return nil, fmt.Errorf("get semester: %w", err)
	}
	result := rowToSemester(row.ID, row.AcademicYearID, row.AcademicYearName, row.Name, row.Label, row.StartDate, row.EndDate, row.IsActive, row.CreatedAt, row.UpdatedAt)
	return &result, nil
}

func (s *SemesterService) GetActive(ctx context.Context) (*Semester, error) {
	row, err := s.q.GetActiveSemester(ctx)
	if err != nil {
		return nil, fmt.Errorf("get active semester: %w", err)
	}
	result := rowToSemester(row.ID, row.AcademicYearID, row.AcademicYearName, row.Name, row.Label, row.StartDate, row.EndDate, row.IsActive, row.CreatedAt, row.UpdatedAt)
	return &result, nil
}

func (s *SemesterService) Activate(ctx context.Context, id string) (*Semester, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := s.q.WithTx(tx)

	if err := qtx.DeactivateSemesters(ctx); err != nil {
		return nil, fmt.Errorf("deactivate semesters: %w", err)
	}
	if err := qtx.DeactivateSemesterAcademicYears(ctx); err != nil {
		return nil, fmt.Errorf("deactivate academic years: %w", err)
	}

	sem, err := qtx.ActivateSemester(ctx, pgUUID(id))
	if err != nil {
		return nil, fmt.Errorf("activate semester: %w", err)
	}
	if sem.AcademicYearID.Valid {
		if _, err = qtx.ActivateSemesterAcademicYear(ctx, sem.AcademicYearID); err != nil {
			return nil, fmt.Errorf("activate academic year: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	// Reload with JOIN
	return s.Get(ctx, id)
}

func (s *SemesterService) Create(ctx context.Context, params SemesterCreateParams) (*Semester, error) {
	startDate, err := time.Parse("2006-01-02", params.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date: %w", err)
	}
	endDate, err := time.Parse("2006-01-02", params.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date: %w", err)
	}

	var pgStart, pgEnd pgtype.Date
	if err := pgStart.Scan(startDate); err != nil {
		return nil, fmt.Errorf("scan start_date: %w", err)
	}
	if err := pgEnd.Scan(endDate); err != nil {
		return nil, fmt.Errorf("scan end_date: %w", err)
	}

	arg := db.CreateSemesterParams{
		AcademicYearID: pgUUID(params.AcademicYearID),
		Name:           params.Name,
		Label:          params.Label,
		StartDate:      pgStart,
		EndDate:        pgEnd,
		IsActive:       params.IsActive,
	}

	_, err = s.q.CreateSemester(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("create semester: %w", err)
	}

	// Read back with JOIN to populate AcademicYearName
	// We need to find the semester we just created - query by unique constraint
	semesters, err := s.q.ListSemesters(ctx)
	if err != nil {
		return nil, fmt.Errorf("list after create: %w", err)
	}
	for _, row := range semesters {
		if row.AcademicYearID == arg.AcademicYearID && row.Name == arg.Name {
			result := rowToSemester(row.ID, row.AcademicYearID, row.AcademicYearName, row.Name, row.Label, row.StartDate, row.EndDate, row.IsActive, row.CreatedAt, row.UpdatedAt)
			return &result, nil
		}
	}
	return nil, fmt.Errorf("semester created but not found")
}

func (s *SemesterService) Delete(ctx context.Context, id string) error {
	return s.q.DeleteSemester(ctx, pgUUID(id))
}

// ─── SchoolClass (Rombel) ────────────────────────────────────────

type SchoolClass struct {
	ID               string `json:"id"`
	Code             string `json:"code"`
	Name             string `json:"name"`
	Level            string `json:"level"`
	IsActive         bool   `json:"is_active"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
	AcademicYearID   string `json:"academic_year_id"`
	AcademicYearName string `json:"academic_year_name"`
}

type SchoolClassCreateParams struct {
	AcademicYearID string
	Code           string
	Name           string
	Level          string
	IsActive       bool
}

func rowToSchoolClass(row db.ListSchoolClassesRow) SchoolClass {
	return SchoolClass{
		ID:               pgUUIDString(row.ID),
		Code:             row.Code,
		Name:             row.Name,
		Level:            row.Level,
		IsActive:         row.IsActive,
		CreatedAt:        row.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:        row.UpdatedAt.Time.Format(time.RFC3339),
		AcademicYearID:   pgUUIDString(row.AcademicYearID),
		AcademicYearName: row.AcademicYearName,
	}
}

func (s *SemesterService) ListSchoolClasses(ctx context.Context) ([]SchoolClass, error) {
	rows, err := s.q.ListSchoolClasses(ctx)
	if err != nil {
		return nil, fmt.Errorf("list school classes: %w", err)
	}
	result := make([]SchoolClass, len(rows))
	for i, row := range rows {
		result[i] = rowToSchoolClass(row)
	}
	return result, nil
}

func (s *SemesterService) CreateSchoolClass(ctx context.Context, params SchoolClassCreateParams) (*SchoolClass, error) {
	arg := db.CreateSchoolClassParams{
		AcademicYearID: pgUUID(params.AcademicYearID),
		Code:           params.Code,
		Name:           params.Name,
		Level:          params.Level,
		IsActive:       params.IsActive,
	}
	_, err := s.q.CreateSchoolClass(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("create school class: %w", err)
	}
	// Read back via list
	rows, err := s.q.ListSchoolClasses(ctx)
	if err != nil {
		return nil, fmt.Errorf("list after create: %w", err)
	}
	for _, row := range rows {
		if row.AcademicYearID == arg.AcademicYearID && row.Code == arg.Code {
			result := rowToSchoolClass(row)
			return &result, nil
		}
	}
	return nil, fmt.Errorf("school class created but not found")
}

func (s *SemesterService) DeleteSchoolClass(ctx context.Context, id string) error {
	return s.q.DeleteSchoolClass(ctx, pgUUID(id))
}

func (s *SemesterService) GetSchoolClass(ctx context.Context, id string) (*SchoolClass, error) {
	row, err := s.q.GetSchoolClass(ctx, pgUUID(id))
	if err != nil {
		return nil, fmt.Errorf("get school class: %w", err)
	}
	result := SchoolClass{
		ID:               pgUUIDString(row.ID),
		Code:             row.Code,
		Name:             row.Name,
		Level:            row.Level,
		IsActive:         row.IsActive,
		CreatedAt:        row.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:        row.UpdatedAt.Time.Format(time.RFC3339),
		AcademicYearID:   pgUUIDString(row.AcademicYearID),
		AcademicYearName: row.AcademicYearName,
	}
	return &result, nil
}

// ─── Rombel Student Management ────────────────────────────────────

type RombelStudentItem struct {
	ID       string `json:"id"`
	NIS      string `json:"nis"`
	NISN     string `json:"nisn"`
	Nama     string `json:"nama"`
	Gender   string `json:"gender"`
	IsActive bool   `json:"is_active"`
	Status   string `json:"status"`
}

type StudentWithRelationsRow struct {
	StudentID          string `json:"student_id"`
	NIS                string `json:"nis"`
	NISN               string `json:"nisn"`
	StudentName        string `json:"student_name"`
	Gender             string `json:"gender"`
	ParentName         string `json:"parent_name"`
	ParentPhone        string `json:"parent_phone"`
	StudentPhone       string `json:"student_phone"`
	StudentAddress     string `json:"student_address"`
	IsActive           bool   `json:"is_active"`
	Status             string `json:"status"`
	ParentID           string `json:"parent_id"`
	ParentNama         string `json:"parent_nama"`
	ParentPhoneLinked  string `json:"parent_phone_linked"`
	ParentAddress      string `json:"parent_address"`
	ParentOccupation   string `json:"parent_occupation"`
	ParentIncomeBand   string `json:"parent_income_band"`
	ParentNIK          string `json:"parent_nik"`
	Relationship       string `json:"relationship"`
	IsPrimaryContact   bool   `json:"is_primary_contact"`
	RelationshipNotes  string `json:"relationship_notes"`
}

func (s *SemesterService) ListStudentsByClass(ctx context.Context, classID string) ([]StudentWithRelationsRow, error) {
	rows, err := s.q.ListStudentsByClassWithParents(ctx, pgUUID(classID))
	if err != nil {
		return nil, fmt.Errorf("list students by class: %w", err)
	}
	items := make([]StudentWithRelationsRow, len(rows))
	for i, r := range rows {
		items[i] = StudentWithRelationsRow{
			StudentID:         pgUUIDString(r.StudentID),
			NIS:               r.Nis,
			NISN:              r.Nisn,
			StudentName:       r.StudentName,
			Gender:            string(r.Gender),
			ParentName:        r.ParentName,
			ParentPhone:       r.ParentPhone,
			StudentPhone:      r.StudentPhone,
			StudentAddress:    r.StudentAddress,
			IsActive:          r.IsActive,
			Status:            string(r.Status),
			ParentID:          pgUUIDString(r.ParentID),
			ParentNama:        r.ParentNama,
			ParentPhoneLinked: r.ParentPhoneLinked,
			ParentAddress:     r.ParentAddress,
			ParentOccupation:  r.ParentOccupation,
			ParentIncomeBand:  r.ParentIncomeBand,
			ParentNIK:         r.ParentNik,
			Relationship:      fmt.Sprintf("%v", r.Relationship),
			IsPrimaryContact:  r.IsPrimaryContact,
			RelationshipNotes: r.RelationshipNotes,
		}
	}
	return items, nil
}

func (s *SemesterService) ListUnassignedStudents(ctx context.Context) ([]RombelStudentItem, error) {
	rows, err := s.q.ListUnassignedStudents(ctx)
	if err != nil {
		return nil, fmt.Errorf("list unassigned students: %w", err)
	}
	items := make([]RombelStudentItem, len(rows))
	for i, r := range rows {
		items[i] = RombelStudentItem{
			ID:       pgUUIDString(r.ID),
			NIS:      r.Nis,
			NISN:     r.Nisn,
			Nama:     r.Nama,
			Gender:   string(r.Gender),
			IsActive: r.IsActive,
			Status:   string(r.Status),
		}
	}
	return items, nil
}

func (s *SemesterService) AssignStudentToClass(ctx context.Context, studentID, classID string) error {
	return s.q.AssignStudentToClass(ctx, db.AssignStudentToClassParams{
		ID:      pgUUID(studentID),
		ClassID: pgUUID(classID),
	})
}

func (s *SemesterService) BulkAssignStudentsToClass(ctx context.Context, studentIDs []string, classID string) error {
	ids := make([]pgtype.UUID, len(studentIDs))
	for i, id := range studentIDs {
		ids[i] = pgUUID(id)
	}
	return s.q.BulkAssignStudentsToClass(ctx, db.BulkAssignStudentsToClassParams{
		Column1: ids,
		ClassID: pgUUID(classID),
	})
}

func (s *SemesterService) RemoveStudentFromClass(ctx context.Context, studentID string) error {
	return s.q.RemoveStudentFromClass(ctx, pgUUID(studentID))
}

// ─── Homeroom (Wali Kelas) ────────────────────────────────────────

type HomeroomAssignmentItem struct {
	ID              string `json:"id"`
	ClassID         string `json:"class_id"`
	ClassCode       string `json:"class_code"`
	ClassName       string `json:"class_name"`
	EmployeeID      string `json:"employee_id"`
	EmployeeName    string `json:"employee_name"`
	AcademicYearID  string `json:"academic_year_id"`
	AcademicYearName string `json:"academic_year_name"`
	StartDate       string `json:"start_date"`
	EndDate         string `json:"end_date"`
	IsActive        bool   `json:"is_active"`
	Notes           string `json:"notes"`
}

func (s *SemesterService) GetActiveHomeroom(ctx context.Context, classID string) (*HomeroomAssignmentItem, error) {
	assignments, err := s.q.ListHomeroomAssignmentsByClass(ctx, pgUUID(classID))
	if err != nil {
		return nil, fmt.Errorf("list homeroom assignments: %w", err)
	}
	for _, a := range assignments {
		if a.IsActive {
			item := HomeroomAssignmentItem{
				ID:              pgUUIDString(a.ID),
				ClassID:         pgUUIDString(a.ClassID),
				ClassCode:       a.ClassCode,
				ClassName:       a.ClassName,
				EmployeeID:      pgUUIDString(a.EmployeeID),
				EmployeeName:    a.EmployeeName,
				AcademicYearID:  pgUUIDString(a.AcademicYearID),
				AcademicYearName: a.AcademicYearName,
				StartDate:       a.StartDate.Time.Format("2006-01-02"),
				EndDate:         a.EndDate.Time.Format("2006-01-02"),
				IsActive:        a.IsActive,
				Notes:           a.Notes,
			}
			return &item, nil
		}
	}
	return nil, nil
}

func (s *SemesterService) SetHomeroomTeacher(ctx context.Context, classID, employeeID string) (*HomeroomAssignmentItem, error) {
	assignment, err := s.q.CreateHomeroomAssignment(ctx, db.CreateHomeroomAssignmentParams{
		ClassID:                pgUUID(classID),
		EmployeeID:             pgUUID(employeeID),
		HomeroomIsActive:       true,
		HomeroomAcademicYearID: nil,
		HomeroomStartDate:      nil,
		HomeroomEndDate:        pgtype.Date{},
		Notes:                  "",
	})
	if err != nil {
		return nil, fmt.Errorf("create homeroom assignment: %w", err)
	}
	return &HomeroomAssignmentItem{
		ID:              pgUUIDString(assignment.ID),
		ClassID:         pgUUIDString(assignment.ClassID),
		ClassCode:       assignment.ClassCode,
		ClassName:       assignment.ClassName,
		EmployeeID:      pgUUIDString(assignment.EmployeeID),
		EmployeeName:    assignment.EmployeeName,
		AcademicYearID:  pgUUIDString(assignment.AcademicYearID),
		AcademicYearName: assignment.AcademicYearName,
		StartDate:       assignment.StartDate.Time.Format("2006-01-02"),
		IsActive:        assignment.IsActive,
	}, nil
}
