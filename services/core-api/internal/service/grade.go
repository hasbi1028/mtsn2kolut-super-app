package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type gradeStore interface {
	ListClassSubjectAssignments(ctx context.Context) ([]db.ListClassSubjectAssignmentsRow, error)
	ListGradeComponents(ctx context.Context, arg db.ListGradeComponentsParams) ([]db.ListGradeComponentsRow, error)
	GetGradeComponent(ctx context.Context, id pgtype.UUID) (db.GradeComponent, error)
	GetGradeComponentHighestScore(ctx context.Context, componentID pgtype.UUID) (float64, error)
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
	Assignments []db.ListClassSubjectAssignmentsRow `json:"assignments"`
	Components  []db.ListGradeComponentsRow         `json:"components"`
	Summary     []db.ListGradebookSummaryRow        `json:"summary"`
	Entries     []db.ListGradeEntriesByComponentRow `json:"entries"`
}

func (s *Grade) Overview(ctx context.Context, assignmentID, componentID pgtype.UUID, publishedOnly bool) (GradeOverview, error) {
	assignments, err := s.q.ListClassSubjectAssignments(ctx)
	if err != nil {
		return GradeOverview{}, err
	}
	out := GradeOverview{Assignments: assignments}
	if assignmentID.Valid {
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
	}
	if componentID.Valid {
		out.Entries, err = s.q.ListGradeEntriesByComponent(ctx, componentID)
		if err != nil {
			return GradeOverview{}, err
		}
	}
	return out, nil
}

func (s *Grade) CreateComponent(ctx context.Context, arg db.CreateGradeComponentParams) (db.GradeComponent, error) {
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

func (s *Grade) UpdateComponent(ctx context.Context, arg db.UpdateGradeComponentParams) (db.GradeComponent, error) {
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

func (s *Grade) SetComponentPublished(ctx context.Context, id pgtype.UUID, isPublished bool) (db.GradeComponent, error) {
	return s.q.UpdateGradeComponentPublishState(ctx, db.UpdateGradeComponentPublishStateParams{
		ID:          id,
		IsPublished: isPublished,
	})
}

func (s *Grade) DeleteComponent(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteGradeComponent(ctx, id)
}

func (s *Grade) UpsertEntry(ctx context.Context, componentID, studentID pgtype.UUID, score float64, notes, gradedBy string) (db.GradeEntry, error) {
	component, err := s.q.GetGradeComponent(ctx, componentID)
	if err != nil {
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

func normalizeGradeCategory(v string) string {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "assignment", "quiz", "midterm", "final", "project", "practice", "attitude", "attendance":
		return strings.TrimSpace(strings.ToLower(v))
	default:
		return "other"
	}
}
