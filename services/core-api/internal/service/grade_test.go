package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeGradeStore struct {
	highestScore           float64
	updateArg              db.UpdateGradeComponentParams
	updateComponent        db.GradeComponent
	component              db.GradeComponent
	listComponents         []db.ListGradeComponentsRow
	listSummary            []db.ListGradebookSummaryRow
	finalization           db.GradeAssignmentFinalization
	finalizationErr        error
	finalizationUpsertArg  db.UpsertGradeAssignmentFinalizationParams
	finalizationDeleteID   pgtype.UUID
}

func (f *fakeGradeStore) ListClassSubjectAssignments(ctx context.Context) ([]db.ListClassSubjectAssignmentsRow, error) {
	return nil, nil
}

func (f *fakeGradeStore) ListGradeComponents(ctx context.Context, arg db.ListGradeComponentsParams) ([]db.ListGradeComponentsRow, error) {
	return f.listComponents, nil
}

func (f *fakeGradeStore) GetGradeComponent(ctx context.Context, id pgtype.UUID) (db.GradeComponent, error) {
	return f.component, nil
}

func (f *fakeGradeStore) GetGradeComponentHighestScore(ctx context.Context, componentID pgtype.UUID) (float64, error) {
	return f.highestScore, nil
}

func (f *fakeGradeStore) GetGradeAssignmentFinalization(ctx context.Context, assignmentID pgtype.UUID) (db.GradeAssignmentFinalization, error) {
	if f.finalizationErr != nil {
		return db.GradeAssignmentFinalization{}, f.finalizationErr
	}
	return f.finalization, nil
}

func (f *fakeGradeStore) UpsertGradeAssignmentFinalization(ctx context.Context, arg db.UpsertGradeAssignmentFinalizationParams) (db.GradeAssignmentFinalization, error) {
	f.finalizationUpsertArg = arg
	return db.GradeAssignmentFinalization{
		AssignmentID: arg.AssignmentID,
		FinalizedBy:  arg.FinalizedBy,
		Notes:        arg.Notes,
	}, nil
}

func (f *fakeGradeStore) DeleteGradeAssignmentFinalization(ctx context.Context, assignmentID pgtype.UUID) error {
	f.finalizationDeleteID = assignmentID
	return nil
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
	return f.listSummary, nil
}

func (f *fakeGradeStore) ListGradeEntriesByComponent(ctx context.Context, componentID pgtype.UUID) ([]db.ListGradeEntriesByComponentRow, error) {
	return nil, nil
}

func (f *fakeGradeStore) UpsertGradeEntry(ctx context.Context, arg db.UpsertGradeEntryParams) (db.GradeEntry, error) {
	return db.GradeEntry{}, nil
}

func TestGradeUpdateComponentRejectsMaxScoreBelowHighestExistingScore(t *testing.T) {
	store := &fakeGradeStore{
		highestScore:    88,
		component:       db.GradeComponent{AssignmentID: pgtype.UUID{}},
		finalizationErr: pgx.ErrNoRows,
	}
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
		highestScore:    -1,
		component:       db.GradeComponent{AssignmentID: pgtype.UUID{}},
		finalizationErr: pgx.ErrNoRows,
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

func TestGradeFinalizeAssignmentRejectsWhenNotReady(t *testing.T) {
	assignmentID := pgtype.UUID{Valid: true}
	store := &fakeGradeStore{
		listComponents: []db.ListGradeComponentsRow{
			{IsPublished: false},
		},
		listSummary: []db.ListGradebookSummaryRow{
			{ComponentCount: 1, FilledCount: 0},
		},
		finalizationErr: pgx.ErrNoRows,
	}
	svc := &Grade{q: store}

	_, err := svc.FinalizeAssignment(context.Background(), assignmentID, "guru-a", "cek awal")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "belum siap difinalisasi") {
		t.Fatalf("error = %q, want readiness guard", err.Error())
	}
}

func TestGradeFinalizeAssignmentPersistsCheckpointWhenReady(t *testing.T) {
	assignmentID := pgtype.UUID{Valid: true}
	store := &fakeGradeStore{
		listComponents: []db.ListGradeComponentsRow{
			{IsPublished: true},
		},
		listSummary: []db.ListGradebookSummaryRow{
			{ComponentCount: 1, FilledCount: 1},
		},
		finalizationErr: pgx.ErrNoRows,
	}
	svc := &Grade{q: store}

	_, err := svc.FinalizeAssignment(context.Background(), assignmentID, "guru-a", "siap cetak")
	if err != nil {
		t.Fatalf("FinalizeAssignment() error = %v", err)
	}
	if store.finalizationUpsertArg.FinalizedBy != "guru-a" {
		t.Fatalf("finalized_by = %q, want %q", store.finalizationUpsertArg.FinalizedBy, "guru-a")
	}
	if store.finalizationUpsertArg.Notes != "siap cetak" {
		t.Fatalf("notes = %q, want %q", store.finalizationUpsertArg.Notes, "siap cetak")
	}
}

func TestGradeCreateComponentRejectsWhenAssignmentAlreadyFinalized(t *testing.T) {
	assignmentID := pgtype.UUID{Valid: true}
	store := &fakeGradeStore{
		finalizationErr: nil,
		finalization:    db.GradeAssignmentFinalization{AssignmentID: assignmentID, FinalizedBy: "guru-a"},
	}
	svc := &Grade{q: store}

	_, err := svc.CreateComponent(context.Background(), db.CreateGradeComponentParams{
		AssignmentID: assignmentID,
		Title:        "UH 1",
		Category:     "quiz",
		Weight:       1,
		MaxScore:     100,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "sudah difinalisasi") {
		t.Fatalf("error = %q, want finalized guard", err.Error())
	}
}

func TestGradeReopenAssignmentDeletesFinalization(t *testing.T) {
	assignmentID := pgtype.UUID{Valid: true}
	store := &fakeGradeStore{finalizationErr: errors.New("unused")}
	svc := &Grade{q: store}

	if err := svc.ReopenAssignment(context.Background(), assignmentID); err != nil {
		t.Fatalf("ReopenAssignment() error = %v", err)
	}
	if store.finalizationDeleteID != assignmentID {
		t.Fatal("delete finalization assignment id was not forwarded")
	}
}
