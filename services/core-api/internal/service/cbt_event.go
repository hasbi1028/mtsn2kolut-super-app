package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type CbtEvent struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

func NewCbtEvent(pool *pgxpool.Pool) *CbtEvent {
	return &CbtEvent{q: db.New(pool), pool: pool}
}

func (s *CbtEvent) List(ctx context.Context) ([]db.ListCbtExamEventsRow, error) {
	rows, err := s.q.ListCbtExamEvents(ctx)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.ListCbtExamEventsRow{}, nil
	}
	return rows, nil
}

func (s *CbtEvent) Get(ctx context.Context, id pgtype.UUID) (db.GetCbtExamEventRow, error) {
	return s.q.GetCbtExamEvent(ctx, id)
}

func (s *CbtEvent) GetResults(ctx context.Context, id pgtype.UUID) ([]db.GetEventResultsRow, error) {
	rows, err := s.q.GetEventResults(ctx, id)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.GetEventResultsRow{}, nil
	}
	return rows, nil
}

type CreateCbtEventInput struct {
	Title          string
	ExamType       db.CbtExamType
	Scope          string
	AcademicYearID pgtype.UUID
	Status         string
}

func (s *CbtEvent) Create(ctx context.Context, in CreateCbtEventInput) (db.CbtExamEvent, error) {
	status := in.Status
	if status == "" {
		status = "draft"
	}
	return s.q.CreateCbtExamEvent(ctx, db.CreateCbtExamEventParams{
		Title:          in.Title,
		ExamType:       in.ExamType,
		Scope:          in.Scope,
		AcademicYearID: in.AcademicYearID,
		Status:         status,
	})
}

func (s *CbtEvent) UpdateStatus(ctx context.Context, id pgtype.UUID, status string) (db.CbtExamEvent, error) {
	return s.q.UpdateCbtExamEventStatus(ctx, db.UpdateCbtExamEventStatusParams{
		ID:     id,
		Status: status,
	})
}

func (s *CbtEvent) Update(ctx context.Context, id pgtype.UUID, in CreateCbtEventInput) (db.CbtExamEvent, error) {
	return s.q.UpdateCbtExamEvent(ctx, db.UpdateCbtExamEventParams{
		ID:             id,
		Title:          in.Title,
		ExamType:       in.ExamType,
		Scope:          in.Scope,
		AcademicYearID: in.AcademicYearID,
	})
}

func (s *CbtEvent) Delete(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteCbtExamEvent(ctx, id)
}
