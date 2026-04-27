package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type CbtQuestion struct {
	q *db.Queries
}

func NewCbtQuestion(q *db.Queries) *CbtQuestion { return &CbtQuestion{q: q} }

func (s *CbtQuestion) List(ctx context.Context) ([]db.ListCbtQuestionsRow, error) {
	return s.q.ListCbtQuestions(ctx)
}

func (s *CbtQuestion) Get(ctx context.Context, id pgtype.UUID) (db.CbtQuestion, error) {
	return s.q.GetCbtQuestion(ctx, id)
}

func (s *CbtQuestion) Create(ctx context.Context, p db.CreateCbtQuestionParams) (db.CbtQuestion, error) {
	return s.q.CreateCbtQuestion(ctx, p)
}

func (s *CbtQuestion) Delete(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteCbtQuestion(ctx, id)
}
