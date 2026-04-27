package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type CbtPackage struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewCbtPackage(pool *pgxpool.Pool) *CbtPackage {
	return &CbtPackage{pool: pool, q: db.New(pool)}
}

func (s *CbtPackage) List(ctx context.Context) ([]db.ListCbtPackagesRow, []db.ListCbtPackageQuestionsRow, error) {
	packages, err := s.q.ListCbtPackages(ctx)
	if err != nil {
		return nil, nil, err
	}
	questions, err := s.q.ListCbtPackageQuestions(ctx)
	if err != nil {
		return nil, nil, err
	}
	return packages, questions, nil
}

type CreateCbtPackageInput struct {
	SubjectID          pgtype.UUID
	Title              string
	Description        string
	DurationMinutes    int32
	RandomizeQuestions bool
	IsActive           bool
	QuestionIDs        []pgtype.UUID
}

func (s *CbtPackage) Create(ctx context.Context, input CreateCbtPackageInput) (db.CbtPackage, error) {
	if len(input.QuestionIDs) == 0 {
		return db.CbtPackage{}, fmt.Errorf("question_ids wajib diisi")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return db.CbtPackage{}, err
	}
	defer tx.Rollback(ctx)

	qtx := s.q.WithTx(tx)
	pkg, err := qtx.CreateCbtPackage(ctx, db.CreateCbtPackageParams{
		SubjectID:          input.SubjectID,
		Title:              input.Title,
		Description:        input.Description,
		DurationMinutes:    input.DurationMinutes,
		RandomizeQuestions: input.RandomizeQuestions,
		IsActive:           input.IsActive,
	})
	if err != nil {
		return db.CbtPackage{}, err
	}

	for i, questionID := range input.QuestionIDs {
		q, err := qtx.GetCbtQuestion(ctx, questionID)
		if err != nil {
			return db.CbtPackage{}, err
		}
		if q.SubjectID != input.SubjectID {
			return db.CbtPackage{}, fmt.Errorf("semua soal harus dari mapel yang sama")
		}
		if err := qtx.AddCbtPackageQuestion(ctx, db.AddCbtPackageQuestionParams{
			PackageID:  pkg.ID,
			QuestionID: questionID,
			Position:   int32(i + 1),
			Points:     1,
		}); err != nil {
			return db.CbtPackage{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return db.CbtPackage{}, err
	}

	return pkg, nil
}

func (s *CbtPackage) Delete(ctx context.Context, id pgtype.UUID) error {
	return s.q.DeleteCbtPackage(ctx, id)
}
