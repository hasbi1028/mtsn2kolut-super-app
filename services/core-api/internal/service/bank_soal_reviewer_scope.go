package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type bankSoalReviewerScopeStore interface {
	ListBankSoalReviewerScopes(ctx context.Context) ([]db.ListBankSoalReviewerScopesRow, error)
	UpsertBankSoalReviewerScope(ctx context.Context, arg db.UpsertBankSoalReviewerScopeParams) (db.UpsertBankSoalReviewerScopeRow, error)
	DeleteBankSoalReviewerScope(ctx context.Context, id pgtype.UUID) (pgtype.UUID, error)
	CanBankSoalUserReview(ctx context.Context, arg db.CanBankSoalUserReviewParams) (bool, error)
	CanBankSoalUserApprove(ctx context.Context, arg db.CanBankSoalUserApproveParams) (bool, error)
}

type BankSoalReviewerScope struct {
	q bankSoalReviewerScopeStore
}

type BankSoalReviewerScopeInput struct {
	UserID     pgtype.UUID
	SubjectID  pgtype.UUID
	GradeLevel pgtype.Int2
	CanReview  bool
	CanApprove bool
	AssignedBy pgtype.UUID
}

type BankSoalReviewerScopeCheckInput struct {
	UserID     pgtype.UUID
	SubjectID  pgtype.UUID
	GradeLevel pgtype.Int2
}

func NewBankSoalReviewerScope(q *db.Queries) *BankSoalReviewerScope {
	return &BankSoalReviewerScope{q: q}
}

func (s *BankSoalReviewerScope) List(ctx context.Context) ([]db.ListBankSoalReviewerScopesRow, error) {
	return s.q.ListBankSoalReviewerScopes(ctx)
}

func (s *BankSoalReviewerScope) Upsert(ctx context.Context, input BankSoalReviewerScopeInput) (db.UpsertBankSoalReviewerScopeRow, error) {
	if err := validateBankSoalReviewerScopeInput(input); err != nil {
		return db.UpsertBankSoalReviewerScopeRow{}, err
	}
	return s.q.UpsertBankSoalReviewerScope(ctx, db.UpsertBankSoalReviewerScopeParams{
		UserID:     input.UserID,
		SubjectID:  input.SubjectID,
		GradeLevel: input.GradeLevel,
		CanReview:  input.CanReview,
		CanApprove: input.CanApprove,
		AssignedBy: input.AssignedBy,
	})
}

func (s *BankSoalReviewerScope) Delete(ctx context.Context, id pgtype.UUID) error {
	if !id.Valid {
		return fmt.Errorf("%w: scope reviewer wajib diisi", domain.ErrBadRequest)
	}
	_, err := s.q.DeleteBankSoalReviewerScope(ctx, id)
	return err
}

func (s *BankSoalReviewerScope) CanReview(ctx context.Context, input BankSoalReviewerScopeCheckInput) (bool, error) {
	if err := validateBankSoalReviewerScopeCheckInput(input); err != nil {
		return false, err
	}
	return s.q.CanBankSoalUserReview(ctx, db.CanBankSoalUserReviewParams{
		UserID:     input.UserID,
		SubjectID:  input.SubjectID,
		GradeLevel: input.GradeLevel,
	})
}

func (s *BankSoalReviewerScope) CanApprove(ctx context.Context, input BankSoalReviewerScopeCheckInput) (bool, error) {
	if err := validateBankSoalReviewerScopeCheckInput(input); err != nil {
		return false, err
	}
	return s.q.CanBankSoalUserApprove(ctx, db.CanBankSoalUserApproveParams{
		UserID:     input.UserID,
		SubjectID:  input.SubjectID,
		GradeLevel: input.GradeLevel,
	})
}

func validateBankSoalReviewerScopeInput(input BankSoalReviewerScopeInput) error {
	if !input.UserID.Valid {
		return fmt.Errorf("%w: user reviewer wajib dipilih", domain.ErrBadRequest)
	}
	if input.GradeLevel.Valid && (input.GradeLevel.Int16 < 7 || input.GradeLevel.Int16 > 9) {
		return fmt.Errorf("%w: tingkat harus VII, VIII, atau IX", domain.ErrBadRequest)
	}
	return nil
}

func validateBankSoalReviewerScopeCheckInput(input BankSoalReviewerScopeCheckInput) error {
	if !input.UserID.Valid {
		return fmt.Errorf("%w: user reviewer wajib dipilih", domain.ErrBadRequest)
	}
	if input.GradeLevel.Valid && (input.GradeLevel.Int16 < 7 || input.GradeLevel.Int16 > 9) {
		return fmt.Errorf("%w: tingkat harus VII, VIII, atau IX", domain.ErrBadRequest)
	}
	return nil
}
