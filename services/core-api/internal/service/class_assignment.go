package service

import (
	"context"
	"fmt"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type SubjectAssignmentService struct {
	q *db.Queries
}

func NewSubjectAssignmentService(q *db.Queries) *SubjectAssignmentService {
	return &SubjectAssignmentService{q: q}
}

// ─── Matrix Data ─────────────────────────────────────────────────

type AssignmentMatrixClass struct {
	ID    string `json:"id"`
	Code  string `json:"code"`
	Name  string `json:"name"`
	Level string `json:"level"`
}

type AssignmentMatrixSubject struct {
	ID                    string  `json:"id"`
	Code                  string  `json:"code"`
	Name                  string  `json:"name"`
	Category              string  `json:"category"`
	IsAssessmentSubject   bool    `json:"is_assessment_subject"`
	IsReportSubject       bool    `json:"is_report_subject"`
	IsScheduleActivity    bool    `json:"is_schedule_activity"`
	CountsForRanking      bool    `json:"counts_for_ranking"`
	IsLocalContent        bool    `json:"is_local_content"`
	IsChoiceSubject       bool    `json:"is_choice_subject"`
	DefaultWeeklyHours    float64 `json:"default_weekly_hours"`
	DisplayOrder          int32   `json:"display_order"`
}

type AssignmentMatrixTeacher struct {
	ID        string `json:"id"`
	NIP       string `json:"nip"`
	Nama      string `json:"nama"`
	UnitKerja string `json:"unit_kerja"`
}

type AssignmentMatrixCell struct {
	ClassID              string  `json:"class_id"`
	SubjectID            string  `json:"subject_id"`
	AssignmentID         string  `json:"assignment_id,omitempty"`
	TeacherEmployeeID    string  `json:"teacher_employee_id,omitempty"`
	TeacherName          string  `json:"teacher_name"`
	Status               string  `json:"status"`
	IntraWeeklyHours     float64 `json:"intra_weekly_hours"`
	KokuWeeklyHours      float64 `json:"koku_weekly_hours"`
	AdditionalHours      float64 `json:"additional_weekly_hours"`
	TotalWeeklyHours     float64 `json:"total_weekly_hours"`
	IsCustomized         bool    `json:"is_customized"`
	ComplianceStatus     string  `json:"compliance_status"`
}

type AssignmentMatrixData struct {
	Classes  []AssignmentMatrixClass   `json:"classes"`
	Subjects []AssignmentMatrixSubject `json:"subjects"`
	Teachers []AssignmentMatrixTeacher `json:"teachers"`
	Cells    []AssignmentMatrixCell    `json:"cells"`
}

func (s *SubjectAssignmentService) GetMatrixData(ctx context.Context, academicYearID string) (*AssignmentMatrixData, error) {
	ayID := pgUUID(academicYearID)

	classes, err := s.q.ListSubjectAssignmentMatrixClasses(ctx, ayID)
	if err != nil {
		return nil, fmt.Errorf("list classes: %w", err)
	}

	// Fallback: if no classes for the requested academic year, get all active classes
	if len(classes) == 0 {
		allRows, err2 := s.q.ScanAllActiveClasses(ctx)
		if err2 == nil {
			for _, r := range allRows {
				classes = append(classes, db.ListSubjectAssignmentMatrixClassesRow{
					ID:    r.ID,
					Code:  r.Code,
					Name:  r.Name,
					Level: r.Level,
				})
			}
		}
	}

	subjects, err := s.q.ListSubjectAssignmentMatrixSubjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("list subjects: %w", err)
	}
	teachers, err := s.q.ListSubjectAssignmentMatrixTeachers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list teachers: %w", err)
	}
	cells, err := s.q.ListSubjectAssignmentMatrixCells(ctx, ayID)
	if err != nil {
		return nil, fmt.Errorf("list cells: %w", err)
	}

	result := &AssignmentMatrixData{
		Classes:  make([]AssignmentMatrixClass, len(classes)),
		Subjects: make([]AssignmentMatrixSubject, len(subjects)),
		Teachers: make([]AssignmentMatrixTeacher, len(teachers)),
		Cells:    make([]AssignmentMatrixCell, len(cells)),
	}
	for i, c := range classes {
		result.Classes[i] = AssignmentMatrixClass{ID: pgUUIDString(c.ID), Code: c.Code, Name: c.Name, Level: c.Level}
	}
	for i, s := range subjects {
		result.Subjects[i] = AssignmentMatrixSubject{
			ID: pgUUIDString(s.ID), Code: s.Code, Name: s.Name, Category: s.Category,
			IsAssessmentSubject: s.IsAssessmentSubject, IsReportSubject: s.IsReportSubject,
			IsScheduleActivity: s.IsScheduleActivity, CountsForRanking: s.CountsForRanking,
			IsLocalContent: s.IsLocalContent, IsChoiceSubject: s.IsChoiceSubject,
			DefaultWeeklyHours: float64(s.DefaultWeeklyHours), DisplayOrder: s.DisplayOrder,
		}
	}
	for i, t := range teachers {
		result.Teachers[i] = AssignmentMatrixTeacher{ID: pgUUIDString(t.ID), NIP: t.Nip, Nama: t.Nama, UnitKerja: t.UnitKerja}
	}
	for i, cell := range cells {
		c := AssignmentMatrixCell{
			ClassID: pgUUIDString(cell.ClassID), SubjectID: pgUUIDString(cell.SubjectID),
			Status: cell.Status, TeacherName: cell.TeacherName,
			ComplianceStatus: cell.ComplianceStatus, IsCustomized: cell.IsCustomized,
		}
		if cell.AssignmentID.Valid {
			c.AssignmentID = pgUUIDString(cell.AssignmentID)
		}
		if cell.TeacherEmployeeID.Valid {
			c.TeacherEmployeeID = pgUUIDString(cell.TeacherEmployeeID)
		}
		if cell.IntraWeeklyHours.Valid {
			fv, _ := cell.IntraWeeklyHours.Float64Value()
			c.IntraWeeklyHours = fv.Float64
		}
		if cell.KokuWeeklyHours.Valid {
			fv, _ := cell.KokuWeeklyHours.Float64Value()
			c.KokuWeeklyHours = fv.Float64
		}
		if cell.AdditionalWeeklyHours.Valid {
			fv, _ := cell.AdditionalWeeklyHours.Float64Value()
			c.AdditionalHours = fv.Float64
		}
		if cell.TotalWeeklyHours.Valid {
			fv, _ := cell.TotalWeeklyHours.Float64Value()
			c.TotalWeeklyHours = fv.Float64
		}
		result.Cells[i] = c
	}
	return result, nil
}

type UpsertCellRequest struct {
	ClassID           string `json:"class_id"`
	SubjectID         string `json:"subject_id"`
	TeacherEmployeeID string `json:"teacher_employee_id"`
}

func (s *SubjectAssignmentService) UpsertCell(ctx context.Context, req UpsertCellRequest) error {
	_, err := s.q.UpsertSubjectAssignmentMatrixCell(ctx, db.UpsertSubjectAssignmentMatrixCellParams{
		ClassID:           pgUUID(req.ClassID),
		SubjectID:         pgUUID(req.SubjectID),
		TeacherEmployeeID: pgUUID(req.TeacherEmployeeID),
	})
	if err != nil {
		return fmt.Errorf("upsert assignment: %w", err)
	}
	return nil
}

func (s *SubjectAssignmentService) DeleteCell(ctx context.Context, id string) error {
	return s.q.DeleteClassSubjectAssignment(ctx, pgUUID(id))
}

func (s *SubjectAssignmentService) GetActiveAcademicYear(ctx context.Context) (string, error) {
	row, err := s.q.GetActiveAcademicYear(ctx)
	if err != nil {
		return "", fmt.Errorf("get active academic year: %w", err)
	}
	return pgUUIDString(row.ID), nil
}
