package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestBankSoalReviewerScopeValidators(t *testing.T) {
	validUser := documentCycleTestUUID(70)
	badInputs := []BankSoalReviewerScopeInput{
		{GradeLevel: pgtype.Int2{Int16: 7, Valid: true}},
		{UserID: validUser, GradeLevel: pgtype.Int2{Int16: 6, Valid: true}},
		{UserID: validUser, GradeLevel: pgtype.Int2{Int16: 10, Valid: true}},
	}
	for _, input := range badInputs {
		if err := validateBankSoalReviewerScopeInput(input); !errors.Is(err, domain.ErrBadRequest) {
			t.Fatalf("validateBankSoalReviewerScopeInput(%+v) error = %v, want ErrBadRequest", input, err)
		}
	}
	if err := validateBankSoalReviewerScopeInput(BankSoalReviewerScopeInput{UserID: validUser, GradeLevel: pgtype.Int2{Int16: 9, Valid: true}}); err != nil {
		t.Fatalf("validateBankSoalReviewerScopeInput(valid) error = %v", err)
	}
	if err := validateBankSoalReviewerScopeInput(BankSoalReviewerScopeInput{UserID: validUser}); err != nil {
		t.Fatalf("validateBankSoalReviewerScopeInput(valid nullable grade) error = %v", err)
	}

	badChecks := []BankSoalReviewerScopeCheckInput{
		{GradeLevel: pgtype.Int2{Int16: 7, Valid: true}},
		{UserID: validUser, GradeLevel: pgtype.Int2{Int16: 6, Valid: true}},
		{UserID: validUser, GradeLevel: pgtype.Int2{Int16: 10, Valid: true}},
	}
	for _, input := range badChecks {
		if err := validateBankSoalReviewerScopeCheckInput(input); !errors.Is(err, domain.ErrBadRequest) {
			t.Fatalf("validateBankSoalReviewerScopeCheckInput(%+v) error = %v, want ErrBadRequest", input, err)
		}
	}
	if err := validateBankSoalReviewerScopeCheckInput(BankSoalReviewerScopeCheckInput{UserID: validUser, GradeLevel: pgtype.Int2{Int16: 7, Valid: true}}); err != nil {
		t.Fatalf("validateBankSoalReviewerScopeCheckInput(valid) error = %v", err)
	}
}

func TestBankSoalReviewerScopeMethodsDelegateToStore(t *testing.T) {
	ctx := context.Background()
	store := &fakeBankSoalReviewerScopeStore{
		listRows:  []db.ListBankSoalReviewerScopesRow{{ID: documentCycleTestUUID(71), Username: "reviewer"}},
		upsertRow: db.UpsertBankSoalReviewerScopeRow{ID: documentCycleTestUUID(72), CanReview: true},
		reviewOK:  true,
		approveOK: true,
	}
	svc := &BankSoalReviewerScope{q: store}
	userID := documentCycleTestUUID(73)
	subjectID := documentCycleTestUUID(74)
	assignedBy := documentCycleTestUUID(75)
	grade := pgtype.Int2{Int16: 8, Valid: true}

	rows, err := svc.List(ctx)
	if err != nil || len(rows) != 1 || rows[0].Username != "reviewer" || store.listCalls != 1 {
		t.Fatalf("List() rows = %+v err = %v listCalls = %d, want delegated list row", rows, err, store.listCalls)
	}

	upserted, err := svc.Upsert(ctx, BankSoalReviewerScopeInput{UserID: userID, SubjectID: subjectID, GradeLevel: grade, CanReview: true, CanApprove: false, AssignedBy: assignedBy})
	if err != nil || upserted.ID != store.upsertRow.ID || store.upsertCalls != 1 {
		t.Fatalf("Upsert() row = %+v err = %v upsertCalls = %d, want delegated upsert", upserted, err, store.upsertCalls)
	}
	if store.lastUpsert.UserID != userID || store.lastUpsert.SubjectID != subjectID || store.lastUpsert.GradeLevel != grade || !store.lastUpsert.CanReview || store.lastUpsert.CanApprove || store.lastUpsert.AssignedBy != assignedBy {
		t.Fatalf("Upsert() params = %+v, want input mapped to db params", store.lastUpsert)
	}

	if err := svc.Delete(ctx, documentCycleTestUUID(76)); err != nil || store.deleteCalls != 1 {
		t.Fatalf("Delete(valid) err = %v deleteCalls = %d, want delegated delete", err, store.deleteCalls)
	}
	if err := svc.Delete(ctx, pgtype.UUID{}); !errors.Is(err, domain.ErrBadRequest) || store.deleteCalls != 1 {
		t.Fatalf("Delete(invalid) err = %v deleteCalls = %d, want bad request without store call", err, store.deleteCalls)
	}

	canReview, err := svc.CanReview(ctx, BankSoalReviewerScopeCheckInput{UserID: userID, SubjectID: subjectID, GradeLevel: grade})
	if err != nil || !canReview || store.reviewCalls != 1 || store.lastReview.UserID != userID || store.lastReview.SubjectID != subjectID || store.lastReview.GradeLevel != grade {
		t.Fatalf("CanReview() ok = %v err = %v calls = %d params = %+v, want delegated true", canReview, err, store.reviewCalls, store.lastReview)
	}
	canApprove, err := svc.CanApprove(ctx, BankSoalReviewerScopeCheckInput{UserID: userID, SubjectID: subjectID, GradeLevel: grade})
	if err != nil || !canApprove || store.approveCalls != 1 || store.lastApprove.UserID != userID || store.lastApprove.SubjectID != subjectID || store.lastApprove.GradeLevel != grade {
		t.Fatalf("CanApprove() ok = %v err = %v calls = %d params = %+v, want delegated true", canApprove, err, store.approveCalls, store.lastApprove)
	}
}

func TestBankSoalReviewerScopeMethodsDoNotCallStoreOnInvalidInput(t *testing.T) {
	ctx := context.Background()
	store := &fakeBankSoalReviewerScopeStore{}
	svc := &BankSoalReviewerScope{q: store}

	if _, err := svc.Upsert(ctx, BankSoalReviewerScopeInput{}); !errors.Is(err, domain.ErrBadRequest) || store.upsertCalls != 0 {
		t.Fatalf("Upsert(invalid) err = %v upsertCalls = %d, want bad request without store call", err, store.upsertCalls)
	}
	if _, err := svc.CanReview(ctx, BankSoalReviewerScopeCheckInput{}); !errors.Is(err, domain.ErrBadRequest) || store.reviewCalls != 0 {
		t.Fatalf("CanReview(invalid) err = %v reviewCalls = %d, want bad request without store call", err, store.reviewCalls)
	}
	if _, err := svc.CanApprove(ctx, BankSoalReviewerScopeCheckInput{}); !errors.Is(err, domain.ErrBadRequest) || store.approveCalls != 0 {
		t.Fatalf("CanApprove(invalid) err = %v approveCalls = %d, want bad request without store call", err, store.approveCalls)
	}
}

type fakeBankSoalReviewerScopeStore struct {
	listRows  []db.ListBankSoalReviewerScopesRow
	upsertRow db.UpsertBankSoalReviewerScopeRow
	reviewOK  bool
	approveOK bool

	listCalls    int
	upsertCalls  int
	deleteCalls  int
	reviewCalls  int
	approveCalls int
	lastUpsert   db.UpsertBankSoalReviewerScopeParams
	lastDelete   pgtype.UUID
	lastReview   db.CanBankSoalUserReviewParams
	lastApprove  db.CanBankSoalUserApproveParams
}

func (f *fakeBankSoalReviewerScopeStore) ListBankSoalReviewerScopes(ctx context.Context) ([]db.ListBankSoalReviewerScopesRow, error) {
	f.listCalls++
	return f.listRows, nil
}

func (f *fakeBankSoalReviewerScopeStore) UpsertBankSoalReviewerScope(ctx context.Context, arg db.UpsertBankSoalReviewerScopeParams) (db.UpsertBankSoalReviewerScopeRow, error) {
	f.upsertCalls++
	f.lastUpsert = arg
	return f.upsertRow, nil
}

func (f *fakeBankSoalReviewerScopeStore) DeleteBankSoalReviewerScope(ctx context.Context, id pgtype.UUID) (pgtype.UUID, error) {
	f.deleteCalls++
	f.lastDelete = id
	return id, nil
}

func (f *fakeBankSoalReviewerScopeStore) CanBankSoalUserReview(ctx context.Context, arg db.CanBankSoalUserReviewParams) (bool, error) {
	f.reviewCalls++
	f.lastReview = arg
	return f.reviewOK, nil
}

func (f *fakeBankSoalReviewerScopeStore) CanBankSoalUserApprove(ctx context.Context, arg db.CanBankSoalUserApproveParams) (bool, error) {
	f.approveCalls++
	f.lastApprove = arg
	return f.approveOK, nil
}
