package service

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeGradeStore struct {
	highestScore    float64
	updateArg       db.UpdateGradeComponentParams
	updateComponent db.GradeComponent
}

func (f *fakeGradeStore) ListClassSubjectAssignments(ctx context.Context) ([]db.ListClassSubjectAssignmentsRow, error) {
	return nil, nil
}

func (f *fakeGradeStore) ListGradeComponents(ctx context.Context, arg db.ListGradeComponentsParams) ([]db.ListGradeComponentsRow, error) {
	return nil, nil
}

func (f *fakeGradeStore) GetGradeComponent(ctx context.Context, id pgtype.UUID) (db.GradeComponent, error) {
	return db.GradeComponent{}, nil
}

func (f *fakeGradeStore) GetGradeComponentHighestScore(ctx context.Context, componentID pgtype.UUID) (float64, error) {
	return f.highestScore, nil
}

func (f *fakeGradeStore) CreateGradeComponent(ctx context.Context, arg db.CreateGradeComponentParams) (db.GradeComponent, error) {
	return db.GradeComponent{}, nil
}

func (f *fakeGradeStore) UpdateGradeComponent(ctx context.Context, arg db.UpdateGradeComponentParams) (db.GradeComponent, error) {
	f.updateArg = arg
	return f.updateComponent, nil
}

func (f *fakeGradeStore) UpdateGradeComponentPublishState(ctx context.Context, arg db.UpdateGradeComponentPublishStateParams) (db.GradeComponent, error) {
	return db.GradeComponent{}, nil
}

func (f *fakeGradeStore) DeleteGradeComponent(ctx context.Context, id pgtype.UUID) error {
	return nil
}

func (f *fakeGradeStore) ListGradebookSummary(ctx context.Context, arg db.ListGradebookSummaryParams) ([]db.ListGradebookSummaryRow, error) {
	return nil, nil
}

func (f *fakeGradeStore) ListGradeEntriesByComponent(ctx context.Context, componentID pgtype.UUID) ([]db.ListGradeEntriesByComponentRow, error) {
	return nil, nil
}

func (f *fakeGradeStore) UpsertGradeEntry(ctx context.Context, arg db.UpsertGradeEntryParams) (db.GradeEntry, error) {
	return db.GradeEntry{}, nil
}

func TestGradeUpdateComponentRejectsMaxScoreBelowHighestExistingScore(t *testing.T) {
	store := &fakeGradeStore{highestScore: 88}
	svc := &Grade{q: store}

	_, err := svc.UpdateComponent(context.Background(), db.UpdateGradeComponentParams{
		ID:       pgtype.UUID{},
		Title:    "Ulangan Harian 1",
		Category: "quiz",
		Weight:   1,
		MaxScore: 80,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "nilai tertinggi 88.00") {
		t.Fatalf("error = %q, want highest-score guard", err.Error())
	}
}

func TestGradeUpdateComponentNormalizesAndForwardsValues(t *testing.T) {
	store := &fakeGradeStore{
		highestScore: -1,
		updateComponent: db.GradeComponent{
			Title:      "Tugas Proyek",
			Category:   "project",
			Weight:     2,
			MaxScore:   100,
			IsPublished: false,
		},
	}
	svc := &Grade{q: store}

	_, err := svc.UpdateComponent(context.Background(), db.UpdateGradeComponentParams{
		ID:       pgtype.UUID{},
		Title:    "  Tugas Proyek  ",
		Category: "PROJECT",
		Weight:   2,
		MaxScore: 100,
	})
	if err != nil {
		t.Fatalf("UpdateComponent() error = %v", err)
	}
	if store.updateArg.Title != "Tugas Proyek" {
		t.Fatalf("title = %q, want %q", store.updateArg.Title, "Tugas Proyek")
	}
	if store.updateArg.Category != "project" {
		t.Fatalf("category = %q, want %q", store.updateArg.Category, "project")
	}
}
