package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type gradeStore interface {
	ListClassSubjectAssignments(ctx context.Context) ([]db.ListClassSubjectAssignmentsRow, error)
	ListGradeAssignmentStatuses(ctx context.Context) ([]db.ListGradeAssignmentStatusesRow, error)
	ListGradeComponents(ctx context.Context, arg db.ListGradeComponentsParams) ([]db.ListGradeComponentsRow, error)
	GetGradeComponent(ctx context.Context, id pgtype.UUID) (db.GradeComponent, error)
	GetGradeComponentHighestScore(ctx context.Context, componentID pgtype.UUID) (float64, error)
	GetGradeAssignmentFinalization(ctx context.Context, assignmentID pgtype.UUID) (db.GradeAssignmentFinalization, error)
	UpsertGradeAssignmentFinalization(ctx context.Context, arg db.UpsertGradeAssignmentFinalizationParams) (db.GradeAssignmentFinalization, error)
	DeleteGradeAssignmentFinalization(ctx context.Context, assignmentID pgtype.UUID) error
	CreateGradeComponent(ctx context.Context, arg db.CreateGradeComponentParams) (db.GradeComponent, error)
	UpdateGradeComponent(ctx context.Context, arg db.UpdateGradeComponentParams) (db.GradeComponent, error)
	UpdateGradeComponentPublishState(ctx context.Context, arg db.UpdateGradeComponentPublishStateParams) (db.GradeComponent, error)
	DeleteGradeComponent(ctx context.Context, id pgtype.UUID) error
	ListGradebookSummary(ctx context.Context, arg db.ListGradebookSummaryParams) ([]db.ListGradebookSummaryRow, error)
	ListGradeEntriesByComponent(ctx context.Context, componentID pgtype.UUID) ([]db.ListGradeEntriesByComponentRow, error)
	UpsertGradeEntry(ctx context.Context, arg db.UpsertGradeEntryParams) (db.GradeEntry, error)
}

type Grade struct {
	q gradeStore
}

func NewGrade(q *db.Queries) *Grade { return &Grade{q: q} }

type GradeOverview struct {
	Assignments        []db.ListClassSubjectAssignmentsRow `json:"assignments"`
	AssignmentStatuses []GradeAssignmentStatus             `json:"assignment_statuses"`
	Components         []db.ListGradeComponentsRow         `json:"components"`
	Summary            []db.ListGradebookSummaryRow        `json:"summary"`
	Entries            []db.ListGradeEntriesByComponentRow `json:"entries"`
	Readiness          GradeReadiness                      `json:"readiness"`
	Finalization       *GradeFinalization                  `json:"finalization,omitempty"`
}

type GradeAssignmentStatus struct {
	AssignmentID            string             `json:"assignment_id"`
	ClassName               string             `json:"class_name"`
	ClassCode               string             `json:"class_code"`
	SubjectName             string             `json:"subject_name"`
	SubjectCode             string             `json:"subject_code"`
	TeacherName             string             `json:"teacher_name"`
	ComponentCount          int                `json:"component_count"`
	PublishedComponentCount int                `json:"published_component_count"`
	DraftComponentCount     int                `json:"draft_component_count"`
	StudentCount            int                `json:"student_count"`
	ReadyStudentCount       int                `json:"ready_student_count"`
	IncompleteStudentCount  int                `json:"incomplete_student_count"`
	MissingGradeCount       int                `json:"missing_grade_count"`
	Ready                   bool               `json:"ready"`
	IsFinalized             bool               `json:"is_finalized"`
	Finalization            *GradeFinalization `json:"finalization,omitempty"`
}

type GradeReadiness struct {
	Ready                   bool `json:"ready"`
	PublishedComponentCount int  `json:"published_component_count"`
	DraftComponentCount     int  `json:"draft_component_count"`
	ReadyStudentCount       int  `json:"ready_student_count"`
	IncompleteStudentCount  int  `json:"incomplete_student_count"`
	MissingGradeCount       int  `json:"missing_grade_count"`
}

type GradeFinalization struct {
	AssignmentID string `json:"assignment_id"`
	FinalizedBy  string `json:"finalized_by"`
	Notes        string `json:"notes"`
	FinalizedAt  string `json:"finalized_at"`
}

func (s *Grade) Overview(ctx context.Context, assignmentID, componentID pgtype.UUID, publishedOnly bool, teacherEmployeeID pgtype.UUID) (GradeOverview, error) {
	assignments, err := s.q.ListClassSubjectAssignments(ctx)
	if err != nil {
		return GradeOverview{}, err
	}
	if teacherEmployeeID.Valid {
		assignments = filterGradeAssignmentsByTeacher(assignments, teacherEmployeeID)
	}
	statusRows, err := s.q.ListGradeAssignmentStatuses(ctx)
	if err != nil {
		return GradeOverview{}, err
	}
	if teacherEmployeeID.Valid {
		statusRows = filterGradeAssignmentStatusesByTeacher(statusRows, teacherEmployeeID)
	}
	out := GradeOverview{
		Assignments:        assignments,
		AssignmentStatuses: buildAssignmentStatuses(statusRows),
	}
	if assignmentID.Valid {
		if err := s.ensureAssignmentAccess(ctx, assignmentID, teacherEmployeeID); err != nil {
			return GradeOverview{}, err
		}
		out.Components, err = s.q.ListGradeComponents(ctx, db.ListGradeComponentsParams{
			AssignmentID:  assignmentID,
			PublishedOnly: publishedOnly,
		})
		if err != nil {
			return GradeOverview{}, err
		}
		out.Summary, err = s.q.ListGradebookSummary(ctx, db.ListGradebookSummaryParams{
			AssignmentID:  assignmentID,
			PublishedOnly: publishedOnly,
		})
		if err != nil {
			return GradeOverview{}, err
		}
		out.Readiness = buildGradeReadiness(out.Components, out.Summary)
		finalization, err := s.q.GetGradeAssignmentFinalization(ctx, assignmentID)
		if err == nil {
			out.Finalization = &GradeFinalization{
				AssignmentID: finalization.AssignmentID.String(),
				FinalizedBy:  finalization.FinalizedBy,
				Notes:        finalization.Notes,
				FinalizedAt:  finalization.FinalizedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			}
		} else if err != pgx.ErrNoRows {
			return GradeOverview{}, err
		}
	}
	if componentID.Valid {
		if err := s.ensureComponentAccess(ctx, componentID, teacherEmployeeID); err != nil {
			return GradeOverview{}, err
		}
		out.Entries, err = s.q.ListGradeEntriesByComponent(ctx, componentID)
		if err != nil {
			return GradeOverview{}, err
		}
	}
	return out, nil
}

func (s *Grade) CreateComponent(ctx context.Context, arg db.CreateGradeComponentParams, teacherEmployeeID pgtype.UUID) (db.GradeComponent, error) {
	if err := s.ensureAssignmentAccess(ctx, arg.AssignmentID, teacherEmployeeID); err != nil {
		return db.GradeComponent{}, err
	}
	if err := s.ensureAssignmentEditable(ctx, arg.AssignmentID); err != nil {
		return db.GradeComponent{}, err
	}
	arg.Title = strings.TrimSpace(arg.Title)
	arg.Category = normalizeGradeCategory(arg.Category)
	if arg.Title == "" {
		return db.GradeComponent{}, fmt.Errorf("judul komponen wajib diisi")
	}
	if arg.Weight < 0 {
		return db.GradeComponent{}, fmt.Errorf("bobot tidak boleh negatif")
	}
	if arg.MaxScore <= 0 {
		return db.GradeComponent{}, fmt.Errorf("skor maksimum harus lebih dari 0")
	}
	return s.q.CreateGradeComponent(ctx, arg)
}

func (s *Grade) UpdateComponent(ctx context.Context, arg db.UpdateGradeComponentParams, teacherEmployeeID pgtype.UUID) (db.GradeComponent, error) {
	component, err := s.q.GetGradeComponent(ctx, arg.ID)
	if err != nil {
		return db.GradeComponent{}, err
	}
	if err := s.ensureAssignmentAccess(ctx, component.AssignmentID, teacherEmployeeID); err != nil {
		return db.GradeComponent{}, err
	}
	if err := s.ensureAssignmentEditable(ctx, component.AssignmentID); err != nil {
		return db.GradeComponent{}, err
	}
	arg.Title = strings.TrimSpace(arg.Title)
	arg.Category = normalizeGradeCategory(arg.Category)
	if arg.Title == "" {
		return db.GradeComponent{}, fmt.Errorf("judul komponen wajib diisi")
	}
	if arg.Weight < 0 {
		return db.GradeComponent{}, fmt.Errorf("bobot tidak boleh negatif")
	}
	if arg.MaxScore <= 0 {
		return db.GradeComponent{}, fmt.Errorf("skor maksimum harus lebih dari 0")
	}
	highestScore, err := s.q.GetGradeComponentHighestScore(ctx, arg.ID)
	if err != nil {
		return db.GradeComponent{}, err
	}
	if highestScore >= 0 && arg.MaxScore < highestScore {
		return db.GradeComponent{}, fmt.Errorf("skor maksimum tidak boleh lebih kecil dari nilai tertinggi %.2f", highestScore)
	}
	return s.q.UpdateGradeComponent(ctx, arg)
}

func (s *Grade) SetComponentPublished(ctx context.Context, id pgtype.UUID, isPublished bool, teacherEmployeeID pgtype.UUID) (db.GradeComponent, error) {
	component, err := s.q.GetGradeComponent(ctx, id)
	if err != nil {
		return db.GradeComponent{}, err
	}
	if err := s.ensureAssignmentAccess(ctx, component.AssignmentID, teacherEmployeeID); err != nil {
		return db.GradeComponent{}, err
	}
	if err := s.ensureAssignmentEditable(ctx, component.AssignmentID); err != nil {
		return db.GradeComponent{}, err
	}
	return s.q.UpdateGradeComponentPublishState(ctx, db.UpdateGradeComponentPublishStateParams{
		ID:          id,
		IsPublished: isPublished,
	})
}

func (s *Grade) DeleteComponent(ctx context.Context, id pgtype.UUID, teacherEmployeeID pgtype.UUID) error {
	component, err := s.q.GetGradeComponent(ctx, id)
	if err != nil {
		return err
	}
	if err := s.ensureAssignmentAccess(ctx, component.AssignmentID, teacherEmployeeID); err != nil {
		return err
	}
	if err := s.ensureAssignmentEditable(ctx, component.AssignmentID); err != nil {
		return err
	}
	return s.q.DeleteGradeComponent(ctx, id)
}

func (s *Grade) UpsertEntry(ctx context.Context, componentID, studentID, teacherEmployeeID pgtype.UUID, score float64, notes, gradedBy string) (db.GradeEntry, error) {
	component, err := s.q.GetGradeComponent(ctx, componentID)
	if err != nil {
		return db.GradeEntry{}, err
	}
	if err := s.ensureAssignmentAccess(ctx, component.AssignmentID, teacherEmployeeID); err != nil {
		return db.GradeEntry{}, err
	}
	if err := s.ensureAssignmentEditable(ctx, component.AssignmentID); err != nil {
		return db.GradeEntry{}, err
	}
	if score < 0 {
		return db.GradeEntry{}, fmt.Errorf("nilai tidak boleh negatif")
	}
	if score > component.MaxScore {
		return db.GradeEntry{}, fmt.Errorf("nilai melebihi skor maksimum %.2f", component.MaxScore)
	}
	return s.q.UpsertGradeEntry(ctx, db.UpsertGradeEntryParams{
		ComponentID: componentID,
		StudentID:   studentID,
		Score:       pgtype.Float8{Float64: score, Valid: true},
		Notes:       strings.TrimSpace(notes),
		GradedBy:    strings.TrimSpace(gradedBy),
	})
}

func (s *Grade) FinalizeAssignment(ctx context.Context, assignmentID pgtype.UUID, finalizedBy, notes string, teacherEmployeeID pgtype.UUID) (db.GradeAssignmentFinalization, error) {
	if err := s.ensureAssignmentAccess(ctx, assignmentID, teacherEmployeeID); err != nil {
		return db.GradeAssignmentFinalization{}, err
	}
	overview, err := s.Overview(ctx, assignmentID, pgtype.UUID{}, false, teacherEmployeeID)
	if err != nil {
		return db.GradeAssignmentFinalization{}, err
	}
	if !overview.Readiness.Ready {
		return db.GradeAssignmentFinalization{}, fmt.Errorf("assignment belum siap difinalisasi")
	}
	return s.q.UpsertGradeAssignmentFinalization(ctx, db.UpsertGradeAssignmentFinalizationParams{
		AssignmentID: assignmentID,
		FinalizedBy:  strings.TrimSpace(finalizedBy),
		Notes:        strings.TrimSpace(notes),
	})
}

func (s *Grade) ReopenAssignment(ctx context.Context, assignmentID, teacherEmployeeID pgtype.UUID) error {
	if err := s.ensureAssignmentAccess(ctx, assignmentID, teacherEmployeeID); err != nil {
		return err
	}
	return s.q.DeleteGradeAssignmentFinalization(ctx, assignmentID)
}

func (s *Grade) ensureAssignmentAccess(ctx context.Context, assignmentID, teacherEmployeeID pgtype.UUID) error {
	if !teacherEmployeeID.Valid {
		return nil
	}
	assignments, err := s.q.ListClassSubjectAssignments(ctx)
	if err != nil {
		return err
	}
	for _, assignment := range assignments {
		if assignment.ID == assignmentID {
			if assignment.TeacherEmployeeID == teacherEmployeeID {
				return nil
			}
			return fmt.Errorf("akses ditolak")
		}
	}
	return fmt.Errorf("assignment tidak ditemukan")
}

func (s *Grade) ensureComponentAccess(ctx context.Context, componentID, teacherEmployeeID pgtype.UUID) error {
	if !teacherEmployeeID.Valid {
		return nil
	}
	component, err := s.q.GetGradeComponent(ctx, componentID)
	if err != nil {
		return err
	}
	return s.ensureAssignmentAccess(ctx, component.AssignmentID, teacherEmployeeID)
}

func (s *Grade) ensureAssignmentEditable(ctx context.Context, assignmentID pgtype.UUID) error {
	_, err := s.q.GetGradeAssignmentFinalization(ctx, assignmentID)
	if err == nil {
		return fmt.Errorf("assignment sudah difinalisasi, buka finalisasi terlebih dahulu")
	}
	if err == pgx.ErrNoRows {
		return nil
	}
	return err
}

func buildGradeReadiness(components []db.ListGradeComponentsRow, summary []db.ListGradebookSummaryRow) GradeReadiness {
	readiness := GradeReadiness{}
	readiness.PublishedComponentCount = 0
	readiness.DraftComponentCount = 0
	for _, component := range components {
		if component.IsPublished {
			readiness.PublishedComponentCount += 1
		} else {
			readiness.DraftComponentCount += 1
		}
	}
	for _, row := range summary {
		missing := int(row.ComponentCount - row.FilledCount)
		if missing < 0 {
			missing = 0
		}
		readiness.MissingGradeCount += missing
		if row.ComponentCount > 0 && row.FilledCount == row.ComponentCount {
			readiness.ReadyStudentCount += 1
		} else {
			readiness.IncompleteStudentCount += 1
		}
	}
	readiness.Ready = len(components) > 0 &&
		readiness.DraftComponentCount == 0 &&
		len(summary) > 0 &&
		readiness.IncompleteStudentCount == 0
	return readiness
}

func buildAssignmentStatuses(rows []db.ListGradeAssignmentStatusesRow) []GradeAssignmentStatus {
	statuses := make([]GradeAssignmentStatus, 0, len(rows))
	for _, row := range rows {
		ready := row.ComponentCount > 0 &&
			row.DraftComponentCount == 0 &&
			row.StudentCount > 0 &&
			row.IncompleteStudentCount == 0

		status := GradeAssignmentStatus{
			AssignmentID:            row.AssignmentID.String(),
			ClassName:               row.ClassName,
			ClassCode:               row.ClassCode,
			SubjectName:             row.SubjectName,
			SubjectCode:             row.SubjectCode,
			TeacherName:             row.TeacherName,
			ComponentCount:          int(row.ComponentCount),
			PublishedComponentCount: int(row.PublishedComponentCount),
			DraftComponentCount:     int(row.DraftComponentCount),
			StudentCount:            int(row.StudentCount),
			ReadyStudentCount:       int(row.ReadyStudentCount),
			IncompleteStudentCount:  int(row.IncompleteStudentCount),
			MissingGradeCount:       int(row.MissingGradeCount),
			Ready:                   ready,
			IsFinalized:             row.IsFinalized,
		}
		if row.IsFinalized && row.FinalizedAt.Valid {
			status.Finalization = &GradeFinalization{
				AssignmentID: row.AssignmentID.String(),
				FinalizedBy:  row.FinalizedBy.String,
				Notes:        row.Notes.String,
				FinalizedAt:  row.FinalizedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			}
		}
		statuses = append(statuses, status)
	}
	return statuses
}

func filterGradeAssignmentsByTeacher(all []db.ListClassSubjectAssignmentsRow, teacherEmployeeID pgtype.UUID) []db.ListClassSubjectAssignmentsRow {
	out := make([]db.ListClassSubjectAssignmentsRow, 0, len(all))
	for _, assignment := range all {
		if assignment.TeacherEmployeeID == teacherEmployeeID {
			out = append(out, assignment)
		}
	}
	return out
}

func filterGradeAssignmentStatusesByTeacher(all []db.ListGradeAssignmentStatusesRow, teacherEmployeeID pgtype.UUID) []db.ListGradeAssignmentStatusesRow {
	out := make([]db.ListGradeAssignmentStatusesRow, 0, len(all))
	for _, row := range all {
		if row.TeacherEmployeeID == teacherEmployeeID {
			out = append(out, row)
		}
	}
	return out
}

func normalizeGradeCategory(v string) string {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "assignment", "quiz", "midterm", "final", "project", "practice", "attitude", "attendance":
		return strings.TrimSpace(strings.ToLower(v))
	default:
		return "other"
	}
}
