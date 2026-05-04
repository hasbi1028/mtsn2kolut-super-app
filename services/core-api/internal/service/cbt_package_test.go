package service

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeCbtPackageCreateStore struct {
	createParams db.CreateCbtPackageParams
	createRow    db.CbtPackage
	createErr    error
	createCalls  int

	questions   map[pgtype.UUID]db.GetCbtQuestionRow
	questionErr error

	addParams []db.AddCbtPackageQuestionParams
	addErr    error
}

func (f *fakeCbtPackageCreateStore) CreateCbtPackage(_ context.Context, arg db.CreateCbtPackageParams) (db.CbtPackage, error) {
	f.createParams = arg
	f.createCalls++
	if f.createErr != nil {
		return db.CbtPackage{}, f.createErr
	}
	return f.createRow, nil
}

func (f *fakeCbtPackageCreateStore) GetCbtQuestion(_ context.Context, id pgtype.UUID) (db.GetCbtQuestionRow, error) {
	if f.questionErr != nil {
		return db.GetCbtQuestionRow{}, f.questionErr
	}
	return f.questions[id], nil
}

func (f *fakeCbtPackageCreateStore) AddCbtPackageQuestion(_ context.Context, arg db.AddCbtPackageQuestionParams) error {
	f.addParams = append(f.addParams, arg)
	return f.addErr
}

func TestCreateCbtPackageRequiresPublishedQuestions(t *testing.T) {
	subjectID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	packageID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	publishedQuestionID := pgtype.UUID{Bytes: [16]byte{3}, Valid: true}
	draftQuestionID := pgtype.UUID{Bytes: [16]byte{4}, Valid: true}

	t.Run("accepts published questions", func(t *testing.T) {
		store := &fakeCbtPackageCreateStore{
			createRow: db.CbtPackage{ID: packageID, SubjectID: subjectID},
			questions: map[pgtype.UUID]db.GetCbtQuestionRow{
				publishedQuestionID: {
					ID:        publishedQuestionID,
					SubjectID: subjectID,
					Status:    db.CbtQuestionStatusEnumPublished,
				},
			},
		}

		_, err := createCbtPackage(context.Background(), store, CreateCbtPackageInput{
			SubjectID:   subjectID,
			Title:       "PAT IPA",
			QuestionIDs: []pgtype.UUID{publishedQuestionID},
		})
		if err != nil {
			t.Fatalf("createCbtPackage() error = %v", err)
		}
		if len(store.addParams) != 1 || store.addParams[0].QuestionID != publishedQuestionID {
			t.Fatalf("AddCbtPackageQuestion() params = %+v, want published question added", store.addParams)
		}
	})

	t.Run("rejects draft questions", func(t *testing.T) {
		store := &fakeCbtPackageCreateStore{
			createRow: db.CbtPackage{ID: packageID, SubjectID: subjectID},
			questions: map[pgtype.UUID]db.GetCbtQuestionRow{
				draftQuestionID: {
					ID:        draftQuestionID,
					SubjectID: subjectID,
					Status:    db.CbtQuestionStatusEnumDraft,
				},
			},
		}

		_, err := createCbtPackage(context.Background(), store, CreateCbtPackageInput{
			SubjectID:   subjectID,
			Title:       "PAT IPA",
			QuestionIDs: []pgtype.UUID{draftQuestionID},
		})
		if err == nil || !strings.Contains(err.Error(), "terbit") {
			t.Fatalf("createCbtPackage() error = %v, want unpublished question rejection", err)
		}
		if len(store.addParams) != 0 {
			t.Fatalf("AddCbtPackageQuestion() calls = %d, want 0", len(store.addParams))
		}
	})

	t.Run("rejects event questions in global package", func(t *testing.T) {
		eventID := pgtype.UUID{Bytes: [16]byte{5}, Valid: true}
		store := &fakeCbtPackageCreateStore{
			createRow: db.CbtPackage{ID: packageID, SubjectID: subjectID},
			questions: map[pgtype.UUID]db.GetCbtQuestionRow{
				publishedQuestionID: {
					ID:        publishedQuestionID,
					EventID:   eventID,
					SubjectID: subjectID,
					Status:    db.CbtQuestionStatusEnumPublished,
				},
			},
		}

		_, err := createCbtPackage(context.Background(), store, CreateCbtPackageInput{
			SubjectID:   subjectID,
			Title:       "Paket Umum IPA",
			QuestionIDs: []pgtype.UUID{publishedQuestionID},
		})
		if err == nil || !strings.Contains(err.Error(), "paket umum") {
			t.Fatalf("createCbtPackage() error = %v, want global package event-question rejection", err)
		}
		if len(store.addParams) != 0 {
			t.Fatalf("AddCbtPackageQuestion() calls = %d, want 0", len(store.addParams))
		}
	})
}
