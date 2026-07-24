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
