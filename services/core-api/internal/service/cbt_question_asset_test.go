package service

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeCbtQuestionAssetStore struct {
	questions []db.GetExamQuestionsRow
}

func (f *fakeCbtQuestionAssetStore) CreateCbtQuestionAsset(ctx context.Context, arg db.CreateCbtQuestionAssetParams) (db.CbtQuestionAsset, error) {
	return db.CbtQuestionAsset{}, nil
}

func (f *fakeCbtQuestionAssetStore) GetCbtQuestionAsset(ctx context.Context, id pgtype.UUID) (db.CbtQuestionAsset, error) {
	return db.CbtQuestionAsset{}, nil
}

func (f *fakeCbtQuestionAssetStore) ListCbtQuestionAssetsByQuestion(ctx context.Context, questionID pgtype.UUID) ([]db.CbtQuestionAsset, error) {
	return nil, nil
}

func (f *fakeCbtQuestionAssetStore) GetExamQuestions(ctx context.Context, packageID pgtype.UUID) ([]db.GetExamQuestionsRow, error) {
	return f.questions, nil
}

func TestCbtQuestionAssetAccessibleByPackageReturnsTrueWhenQuestionIncluded(t *testing.T) {
	questionID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	packageID := mustUUID(t, "22222222-2222-2222-2222-222222222222")
	svc := &CbtQuestionAsset{
		q: &fakeCbtQuestionAssetStore{
			questions: []db.GetExamQuestionsRow{
				{ID: questionID},
			},
		},
	}

	allowed, err := svc.AccessibleByPackage(context.Background(), questionID, packageID)
	if err != nil {
		t.Fatalf("AccessibleByPackage() error = %v", err)
	}
	if !allowed {
		t.Fatal("AccessibleByPackage() = false, want true")
	}
}

func TestCbtQuestionAssetAccessibleByPackageReturnsFalseWhenQuestionExcluded(t *testing.T) {
	questionID := mustUUID(t, "33333333-3333-3333-3333-333333333333")
	packageID := mustUUID(t, "44444444-4444-4444-4444-444444444444")
	svc := &CbtQuestionAsset{
		q: &fakeCbtQuestionAssetStore{
			questions: []db.GetExamQuestionsRow{
				{ID: mustUUID(t, "55555555-5555-5555-5555-555555555555")},
			},
		},
	}

	allowed, err := svc.AccessibleByPackage(context.Background(), questionID, packageID)
	if err != nil {
		t.Fatalf("AccessibleByPackage() error = %v", err)
	}
	if allowed {
		t.Fatal("AccessibleByPackage() = true, want false")
	}
}

func TestCbtQuestionAssetAccessibleByPackageReturnsFalseForInvalidIDs(t *testing.T) {
	svc := &CbtQuestionAsset{q: &fakeCbtQuestionAssetStore{}}

	allowed, err := svc.AccessibleByPackage(context.Background(), pgtype.UUID{}, pgtype.UUID{})
	if err != nil {
		t.Fatalf("AccessibleByPackage() error = %v", err)
	}
	if allowed {
		t.Fatal("AccessibleByPackage() = true, want false")
	}
}
