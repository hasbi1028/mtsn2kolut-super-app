package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type CurriculumService struct {
	q *db.Queries
}

func NewCurriculumService(q *db.Queries) *CurriculumService {
	return &CurriculumService{q: q}
}

func numericFromFloat64(v float64) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan(v)
	return n
}

// ─── Curriculum Profile ──────────────────────────────────────────

type CurriculumProfile struct {
	ID                       string `json:"id"`
	Code                     string `json:"code"`
	Name                     string `json:"name"`
	RegulationReference      string `json:"regulation_reference"`
	EducationLevel           string `json:"education_level"`
	EffectiveAcademicYearID  string `json:"effective_academic_year_id"`
	Status                   string `json:"status"`
	Notes                    string `json:"notes"`
	CreatedAt                string `json:"created_at"`
	UpdatedAt                string `json:"updated_at"`
}

type CurriculumProfileCreateParams struct {
	Code                string
	Name                string
	RegulationReference string
	EducationLevel      string
	EffectiveYearID     string
	Status              string
	Notes               string
}

func dbProfileToService(row db.CurriculumProfile) CurriculumProfile {
	return CurriculumProfile{
		ID:                  pgUUIDString(row.ID),
		Code:                row.Code,
		Name:                row.Name,
		RegulationReference: row.RegulationReference,
		EducationLevel:      row.EducationLevel,
		EffectiveAcademicYearID: pgUUIDString(row.EffectiveAcademicYearID),
		Status:             row.Status,
		Notes:              row.Notes,
		CreatedAt:          row.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:          row.UpdatedAt.Time.Format(time.RFC3339),
	}
}

func (s *CurriculumService) ListProfiles(ctx context.Context) ([]CurriculumProfile, error) {
	rows, err := s.q.ListCurriculumProfiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("list curriculum profiles: %w", err)
	}
	result := make([]CurriculumProfile, len(rows))
	for i, row := range rows {
		result[i] = dbProfileToService(row)
	}
	return result, nil
}

func (s *CurriculumService) GetActiveProfile(ctx context.Context) (*CurriculumProfile, error) {
	row, err := s.q.GetActiveCurriculumProfile(ctx)
	if err != nil {
		return nil, fmt.Errorf("get active curriculum profile: %w", err)
	}
	result := dbProfileToService(row)
	return &result, nil
}

func (s *CurriculumService) CreateProfile(ctx context.Context, params CurriculumProfileCreateParams) (*CurriculumProfile, error) {
	row, err := s.q.CreateCurriculumProfile(ctx, db.CreateCurriculumProfileParams{
		Code:                    params.Code,
		Name:                    params.Name,
		RegulationReference:     params.RegulationReference,
		EducationLevel:          params.EducationLevel,
		EffectiveAcademicYearID: pgUUID(params.EffectiveYearID),
		Status:                  params.Status,
		Notes:                   params.Notes,
	})
	if err != nil {
		return nil, fmt.Errorf("create curriculum profile: %w", err)
	}
	result := dbProfileToService(row)
	return &result, nil
}

func (s *CurriculumService) ActivateProfile(ctx context.Context, id string) (*CurriculumProfile, error) {
	if err := s.q.DeactivateCurriculumProfiles(ctx, pgUUID(id)); err != nil {
		return nil, fmt.Errorf("deactivate profiles: %w", err)
	}
	row, err := s.q.ActivateCurriculumProfile(ctx, pgUUID(id))
	if err != nil {
		return nil, fmt.Errorf("activate profile: %w", err)
	}
	result := dbProfileToService(row)
	return &result, nil
}

func (s *CurriculumService) DeleteProfile(ctx context.Context, id string) error {
	return s.q.DeleteCurriculumProfile(ctx, pgUUID(id))
}

func (s *CurriculumService) UpdateProfile(ctx context.Context, id string, params CurriculumProfileCreateParams) (*CurriculumProfile, error) {
	row, err := s.q.UpdateCurriculumProfile(ctx, db.UpdateCurriculumProfileParams{
		Code:                    params.Code,
		Name:                    params.Name,
		RegulationReference:     params.RegulationReference,
		EducationLevel:          params.EducationLevel,
		EffectiveAcademicYearID: pgUUID(params.EffectiveYearID),
		Status:                  params.Status,
		Notes:                   params.Notes,
		ID:                      pgUUID(id),
	})
	if err != nil {
		return nil, fmt.Errorf("update curriculum profile: %w", err)
	}
	result := dbProfileToService(row)
	return &result, nil
}

// ─── Subject Allocations ─────────────────────────────────────────

type SubjectAllocation struct {
	ID                   string  `json:"id"`
	CurriculumProfileID  string  `json:"curriculum_profile_id"`
	CurriculumCode       string  `json:"curriculum_code"`
	CurriculumName       string  `json:"curriculum_name"`
	SubjectID            string  `json:"subject_id"`
	SubjectCode          string  `json:"subject_code"`
	SubjectName          string  `json:"subject_name"`
	Level                string  `json:"level"`
	SubjectGroup         string  `json:"subject_group"`
	IntraWeeklyHours     float64 `json:"intra_weekly_hours"`
	KokuWeeklyHours      float64 `json:"koku_weekly_hours"`
	TotalWeeklyHours     float64 `json:"total_weekly_hours"`
	CountsForSchedule    bool    `json:"counts_for_schedule"`
	CountsForReport      bool    `json:"counts_for_report"`
	CountsForAssessment  bool    `json:"counts_for_assessment"`
	CountsForRanking     bool    `json:"counts_for_ranking"`
	IsRequired           bool    `json:"is_required"`
	Notes                string  `json:"notes"`
}

func (s *CurriculumService) ListAllocations(ctx context.Context, profileID string, level string) ([]SubjectAllocation, error) {
	var levelArg pgtype.Text
	if level != "" {
		levelArg = pgtype.Text{String: level, Valid: true}
	}
	rows, err := s.q.ListCurriculumSubjectAllocations(ctx, db.ListCurriculumSubjectAllocationsParams{
		CurriculumProfileID: pgUUID(profileID),
		Level:               levelArg,
	})
	if err != nil {
		return nil, fmt.Errorf("list allocations: %w", err)
	}
	result := make([]SubjectAllocation, len(rows))
	for i, row := range rows {
		sa := SubjectAllocation{
			ID:                   pgUUIDString(row.ID),
			CurriculumProfileID:  pgUUIDString(row.CurriculumProfileID),
			CurriculumCode:       row.CurriculumProfileCode,
			CurriculumName:       row.CurriculumProfileName,
			SubjectID:            pgUUIDString(row.SubjectID),
			SubjectCode:          row.SubjectCode,
			SubjectName:          row.SubjectName,
			Level:                row.Level,
			SubjectGroup:         row.SubjectGroup,
			CountsForSchedule:    row.CountsForSchedule,
			CountsForReport:      row.CountsForReport,
			CountsForAssessment:  row.CountsForAssessment,
			CountsForRanking:     row.CountsForRanking,
			IsRequired:           row.IsRequired,
			Notes:                row.Notes,
		}
		if row.IntraWeeklyHours.Valid {
			fv, _ := row.IntraWeeklyHours.Float64Value()
			sa.IntraWeeklyHours = fv.Float64
		}
		if row.KokuWeeklyHours.Valid {
			fv, _ := row.KokuWeeklyHours.Float64Value()
			sa.KokuWeeklyHours = fv.Float64
		}
		if row.TotalWeeklyHours.Valid {
			fv, _ := row.TotalWeeklyHours.Float64Value()
			sa.TotalWeeklyHours = fv.Float64
		}
		result[i] = sa
	}
	return result, nil
}

func (s *CurriculumService) CreateAllocation(ctx context.Context, profileID, subjectID, level, group string, intra, koku, total float64, notes string) (*SubjectAllocation, error) {
	_, err := s.q.CreateCurriculumAllocation(ctx, db.CreateCurriculumAllocationParams{
		CurriculumProfileID: pgUUID(profileID),
		SubjectID:           pgUUID(subjectID),
		Level:               level,
		SubjectGroup:        group,
		IntraAnnualHours:    0,
		KokuAnnualHours:     0,
		TotalAnnualHours:    0,
		IntraWeeklyHours:    numericFromFloat64(intra),
		KokuWeeklyHours:     numericFromFloat64(koku),
		TotalWeeklyHours:    numericFromFloat64(total),
		LessonMinutes:       40,
		DisplayOrder:        0,
		CountsForSchedule:   true,
		CountsForReport:     true,
		CountsForAssessment: true,
		CountsForRanking:    true,
		IsRequired:          true,
		Notes:               notes,
	})
	if err != nil {
		return nil, fmt.Errorf("create allocation: %w", err)
	}
	// Read back by listing with level
	rows, err := s.ListAllocations(ctx, profileID, level)
	if err != nil {
		return nil, err
	}
	return &rows[len(rows)-1], nil
}

func (s *CurriculumService) DeleteAllocation(ctx context.Context, id string) error {
	return s.q.DeleteCurriculumAllocation(ctx, pgUUID(id))
}

func (s *CurriculumService) UpdateAllocation(ctx context.Context, id string, params db.UpdateCurriculumAllocationParams) (*SubjectAllocation, error) {
	_, err := s.q.UpdateCurriculumAllocation(ctx, db.UpdateCurriculumAllocationParams{
		SubjectID:           params.SubjectID,
		Level:               params.Level,
		SubjectGroup:        params.SubjectGroup,
		IntraAnnualHours:    params.IntraAnnualHours,
		KokuAnnualHours:     params.KokuAnnualHours,
		TotalAnnualHours:    params.TotalAnnualHours,
		IntraWeeklyHours:    params.IntraWeeklyHours,
		KokuWeeklyHours:     params.KokuWeeklyHours,
		TotalWeeklyHours:    params.TotalWeeklyHours,
		LessonMinutes:       params.LessonMinutes,
		DisplayOrder:        params.DisplayOrder,
		CountsForSchedule:   params.CountsForSchedule,
		CountsForReport:     params.CountsForReport,
		CountsForAssessment: params.CountsForAssessment,
		CountsForRanking:    params.CountsForRanking,
		IsRequired:          params.IsRequired,
		Notes:               params.Notes,
		ID:                  pgUUID(id),
	})
	if err != nil {
		return nil, fmt.Errorf("update allocation: %w", err)
	}
	return nil, fmt.Errorf("not implemented: readback")
}

// ─── Class Assignments ───────────────────────────────────────────

type ClassCurriculumAssignment struct {
	ID                 string `json:"id"`
	ClassID            string `json:"class_id"`
	ClassCode          string `json:"class_code"`
	ClassName          string `json:"class_name"`
	ClassLevel         string `json:"class_level"`
	CurriculumProfileID string `json:"curriculum_profile_id"`
	CurriculumCode     string `json:"curriculum_code"`
	CurriculumName     string `json:"curriculum_name"`
	IsActive           bool   `json:"is_active"`
	Notes              string `json:"notes"`
}

func (s *CurriculumService) ListActiveAssignments(ctx context.Context, classID string) ([]ClassCurriculumAssignment, error) {
	var classIDArg pgtype.UUID
	if classID != "" {
		classIDArg = pgUUID(classID)
	}
	rows, err := s.q.ListActiveAssignmentsBySemester(ctx, classIDArg)
	if err != nil {
		return nil, fmt.Errorf("list assignments: %w", err)
	}
	result := make([]ClassCurriculumAssignment, len(rows))
	for i, row := range rows {
		result[i] = ClassCurriculumAssignment{
			ID:                 pgUUIDString(row.ID),
			ClassID:            pgUUIDString(row.ClassID),
			ClassCode:          row.ClassCode,
			ClassName:          row.ClassName,
			ClassLevel:         row.ClassLevel,
			CurriculumProfileID: pgUUIDString(row.CurriculumProfileID),
			CurriculumCode:     row.CurriculumCode,
			CurriculumName:     row.CurriculumName,
			IsActive:           row.IsActive,
			Notes:              row.Notes,
		}
	}
	return result, nil
}

func (s *CurriculumService) CreateAssignment(ctx context.Context, classID, profileID, notes string) (*ClassCurriculumAssignment, error) {
	_, err := s.q.CreateClassCurriculumAssignment(ctx, db.CreateClassCurriculumAssignmentParams{
		ClassID:             pgUUID(classID),
		CurriculumProfileID: pgUUID(profileID),
		IsActive:            true,
		Notes:               notes,
	})
	if err != nil {
		return nil, fmt.Errorf("create assignment: %w", err)
	}
	rows, err := s.ListActiveAssignments(ctx, classID)
	if err != nil {
		return nil, err
	}
	if len(rows) > 0 {
		return &rows[0], nil
	}
	return nil, fmt.Errorf("assignment created but not found")
}

func (s *CurriculumService) DeleteAssignment(ctx context.Context, id string) error {
	return s.q.DeleteClassCurriculumAssignment(ctx, pgUUID(id))
}
