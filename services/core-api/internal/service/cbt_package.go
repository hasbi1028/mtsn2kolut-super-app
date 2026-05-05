package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type CbtPackage struct {
	pool *pgxpool.Pool
	q    cbtPackageStore
}

type cbtPackageStore interface {
	ListCbtPackages(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtPackagesRow, error)
	ListCbtPackageQuestions(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtPackageQuestionsRow, error)
	GetCbtPackageUsage(ctx context.Context, id pgtype.UUID) (int32, error)
	DeleteCbtPackage(ctx context.Context, id pgtype.UUID) (int64, error)
	WithTx(tx pgx.Tx) *db.Queries
}

type cbtPackageCreateStore interface {
	CreateCbtPackage(ctx context.Context, arg db.CreateCbtPackageParams) (db.CbtPackage, error)
	GetCbtQuestion(ctx context.Context, id pgtype.UUID) (db.GetCbtQuestionRow, error)
	AddCbtPackageQuestion(ctx context.Context, arg db.AddCbtPackageQuestionParams) error
}

func NewCbtPackage(pool *pgxpool.Pool) *CbtPackage {
	return &CbtPackage{pool: pool, q: db.New(pool)}
}

func (s *CbtPackage) List(ctx context.Context, eventID pgtype.UUID) ([]db.ListCbtPackagesRow, []db.ListCbtPackageQuestionsRow, error) {
	packages, err := s.q.ListCbtPackages(ctx, eventID)
	if err != nil {
		return nil, nil, err
	}
	questions, err := s.q.ListCbtPackageQuestions(ctx, eventID)
	if err != nil {
		return nil, nil, err
	}
	return packages, questions, nil
}

type CreateCbtPackageInput struct {
	EventID            pgtype.UUID
	SubjectID          pgtype.UUID
	Title              string
	Description        string
	DurationMinutes    int32
	RandomizeQuestions bool
	IsActive           bool
	QuestionIDs        []pgtype.UUID
	QuestionWeights    map[string]int32
}

func (s *CbtPackage) Create(ctx context.Context, input CreateCbtPackageInput) (db.CbtPackage, error) {
	if err := validateCbtPackageDuration(input.DurationMinutes); err != nil {
		return db.CbtPackage{}, err
	}
	input.DurationMinutes = normalizeCbtPackageDuration(input.DurationMinutes)
	if len(input.QuestionIDs) == 0 {
		return db.CbtPackage{}, fmt.Errorf("question_ids wajib diisi")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return db.CbtPackage{}, err
	}
	defer tx.Rollback(ctx)

	qtx := s.q.WithTx(tx)
	pkg, err := createCbtPackage(ctx, qtx, input)
	if err != nil {
		return db.CbtPackage{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return db.CbtPackage{}, err
	}

	return pkg, nil
}

func createCbtPackage(ctx context.Context, q cbtPackageCreateStore, input CreateCbtPackageInput) (db.CbtPackage, error) {
	if err := validateCbtPackageDuration(input.DurationMinutes); err != nil {
		return db.CbtPackage{}, err
	}
	input.DurationMinutes = normalizeCbtPackageDuration(input.DurationMinutes)
	pkg, err := q.CreateCbtPackage(ctx, db.CreateCbtPackageParams{
		EventID:            input.EventID,
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
		question, err := q.GetCbtQuestion(ctx, questionID)
		if err != nil {
			return db.CbtPackage{}, err
		}
		if question.SubjectID != input.SubjectID {
			return db.CbtPackage{}, fmt.Errorf("semua soal harus dari mapel yang sama")
		}
		if question.Status != db.CbtQuestionStatusEnumPublished {
			return db.CbtPackage{}, fmt.Errorf("semua soal paket harus berstatus terbit")
		}
		switch {
		case input.EventID.Valid && question.EventID.Valid && !sameUUID(question.EventID, input.EventID):
			return db.CbtPackage{}, fmt.Errorf("%w: soal paket event harus berasal dari event yang sama", domain.ErrBadRequest)
		case !input.EventID.Valid && question.EventID.Valid:
			return db.CbtPackage{}, fmt.Errorf("%w: paket umum tidak boleh memakai soal khusus event", domain.ErrBadRequest)
		}
		points := int32(1)
		if input.QuestionWeights != nil {
			if value, ok := input.QuestionWeights[pgUUIDString(questionID)]; ok {
				points = value
			}
		}
		if points < 1 || points > 100 {
			return db.CbtPackage{}, fmt.Errorf("bobot soal harus 1-100")
		}
		if err := q.AddCbtPackageQuestion(ctx, db.AddCbtPackageQuestionParams{
			PackageID:  pkg.ID,
			QuestionID: questionID,
			Position:   int32(i + 1),
			Points:     points,
		}); err != nil {
			return db.CbtPackage{}, err
		}
	}
	return pkg, nil
}

func (s *CbtPackage) Delete(ctx context.Context, id pgtype.UUID) error {
	sessionCount, err := s.q.GetCbtPackageUsage(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	if sessionCount > 0 {
		return fmt.Errorf("%w: paket CBT sudah digunakan oleh sesi ujian dan tidak dapat dihapus", domain.ErrConflict)
	}
	rows, err := s.q.DeleteCbtPackage(ctx, id)
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func validateCbtPackageDuration(duration int32) error {
	if duration == 0 {
		return nil
	}
	if duration < 1 || duration > 360 {
		return fmt.Errorf("%w: durasi paket CBT harus 1-360 menit", domain.ErrBadRequest)
	}
	return nil
}

func normalizeCbtPackageDuration(duration int32) int32 {
	if duration == 0 {
		return 60
	}
	return duration
}
