package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestCbtPackageDeepCreateStopsWhenQuestionInsertAffectsNoRows(t *testing.T) {
	ctx := context.Background()
	subjectID := pgtype.UUID{Bytes: [16]byte{11, 1}, Valid: true}
	packageID := pgtype.UUID{Bytes: [16]byte{11, 2}, Valid: true}
	questionID := pgtype.UUID{Bytes: [16]byte{11, 3}, Valid: true}
	store := &fakeCbtPackageCreateRowsStore{
		createRow: db.CbtPackage{ID: packageID, SubjectID: subjectID},
		questions: map[pgtype.UUID]db.GetCbtQuestionRow{
			questionID: {ID: questionID, SubjectID: subjectID, Status: db.CbtQuestionStatusEnumPublished},
		},
		addRows: 0,
	}

	_, err := createCbtPackage(ctx, store, CreateCbtPackageInput{
		SubjectID:   subjectID,
		Title:       "PAT IPA",
		QuestionIDs: []pgtype.UUID{questionID},
	})
	if err == nil || !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "Siap Pakai atau sudah terbit") {
		t.Fatalf("createCbtPackage(add rows=0) error = %v, want workflow conflict", err)
	}
	if store.createCalls != 1 || len(store.addArgs) != 1 {
		t.Fatalf("calls create/add = %d/%d, want package created then one failed add", store.createCalls, len(store.addArgs))
	}
}

func TestCbtPackageDeepUpdateMetadataSuccessAndLockedBranch(t *testing.T) {
	ctx := context.Background()
	packageID := pgtype.UUID{Bytes: [16]byte{12, 1}, Valid: true}
	store := &fakeCbtPackageEditStore{
		lockRow:   db.LockCbtPackageForEditRow{ID: packageID},
		detailRow: db.GetCbtPackageDetailRow{ID: packageID, Title: "PAT Trimmed"},
	}

	got, err := (&CbtPackage{q: store}).UpdateMetadata(ctx, UpdateCbtPackageInput{
		ID:                 packageID,
		Title:              "  PAT Trimmed  ",
		Description:        "  desc  ",
		DurationMinutes:    0,
		RandomizeQuestions: true,
		RandomizeOptions:   true,
		RandomSeed:         "  seed-42  ",
		IsActive:           true,
	})
	if err != nil {
		t.Fatalf("UpdateMetadata(success) error = %v", err)
	}
	if got.Package.ID != packageID || store.updateCalls != 1 {
		t.Fatalf("UpdateMetadata(success) result/calls = %+v/%d, want refreshed package and one update", got.Package, store.updateCalls)
	}
	if store.updateArg.Title != "PAT Trimmed" || store.updateArg.Description != "desc" || store.updateArg.RandomSeed != "seed-42" {
		t.Fatalf("UpdateMetadata trims = %+v, want trimmed title/description/random seed", store.updateArg)
	}
	if store.updateArg.DurationMinutes != 60 || store.updateArg.SourceMode != "teacher_class" || !store.updateArg.RandomizeQuestions || !store.updateArg.RandomizeOptions || !store.updateArg.IsActive {
		t.Fatalf("UpdateMetadata normalized params = %+v, want defaults and flags preserved", store.updateArg)
	}

	locked := &fakeCbtPackageEditStore{lockRow: db.LockCbtPackageForEditRow{ID: packageID, LockedAt: pgtype.Timestamptz{Valid: true}}}
	_, err = (&CbtPackage{q: locked}).UpdateMetadata(ctx, UpdateCbtPackageInput{ID: packageID, Title: "PAT", DurationMinutes: 90})
	if err == nil || !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "metadata") || locked.updateCalls != 0 {
		t.Fatalf("UpdateMetadata(locked) err/calls = %v/%d, want metadata conflict before update", err, locked.updateCalls)
	}
}

func TestCbtPackageDeepReplaceQuestionsSuccessOrderingAndLockedBranch(t *testing.T) {
	ctx := context.Background()
	packageID := pgtype.UUID{Bytes: [16]byte{13, 1}, Valid: true}
	eventID := pgtype.UUID{Bytes: [16]byte{13, 2}, Valid: true}
	subjectID := pgtype.UUID{Bytes: [16]byte{13, 3}, Valid: true}
	firstQuestionID := pgtype.UUID{Bytes: [16]byte{13, 4}, Valid: true}
	secondQuestionID := pgtype.UUID{Bytes: [16]byte{13, 5}, Valid: true}
	store := &fakeCbtPackageEditStore{
		lockRow:   db.LockCbtPackageForEditRow{ID: packageID, EventID: eventID, SubjectID: subjectID},
		detailRow: db.GetCbtPackageDetailRow{ID: packageID, SubjectID: subjectID},
		questions: map[pgtype.UUID]db.GetCbtQuestionRow{
			firstQuestionID:  {ID: firstQuestionID, EventID: eventID, SubjectID: subjectID, Status: db.CbtQuestionStatusEnumPublished},
			secondQuestionID: {ID: secondQuestionID, SubjectID: subjectID, Status: db.CbtQuestionStatusEnumDraft, WorkflowStatus: "approved"},
		},
		detailQuestions: []db.ListCbtPackageQuestionsByPackageRow{
			{PackageID: packageID, QuestionID: firstQuestionID, QuestionType: "multiple_choice", Status: db.CbtQuestionStatusEnumPublished, Points: 3, TargetLevel: pgtype.Text{String: "VIII", Valid: true}, CpRef: "CP", TpRef: "TP", CognitiveLevel: "C2"},
		},
	}

	got, err := (&CbtPackage{q: store}).ReplaceQuestions(ctx, ReplaceCbtPackageQuestionsInput{
		PackageID:   packageID,
		QuestionIDs: []pgtype.UUID{firstQuestionID, secondQuestionID},
		QuestionWeights: map[string]int32{
			pgUUIDString(firstQuestionID):  3,
			pgUUIDString(secondQuestionID): 100,
		},
	})
	if err != nil {
		t.Fatalf("ReplaceQuestions(success) error = %v", err)
	}
	if store.deleteQuestionCalls != 1 || len(store.addArgs) != 2 {
		t.Fatalf("ReplaceQuestions mutation counts delete/add = %d/%d, want 1/2", store.deleteQuestionCalls, len(store.addArgs))
	}
	if store.addArgs[0].QuestionID != firstQuestionID || store.addArgs[0].Position != 1 || store.addArgs[0].Points != 3 || store.addArgs[1].QuestionID != secondQuestionID || store.addArgs[1].Position != 2 || store.addArgs[1].Points != 100 {
		t.Fatalf("ReplaceQuestions add args = %+v, want ordered weighted inserts", store.addArgs)
	}
	if got.Package.ID != packageID || len(got.Questions) != 1 {
		t.Fatalf("ReplaceQuestions result = %+v, want refreshed detail", got)
	}

	locked := &fakeCbtPackageEditStore{lockRow: db.LockCbtPackageForEditRow{ID: packageID, SubjectID: subjectID, LockedAt: pgtype.Timestamptz{Valid: true}}}
	_, err = (&CbtPackage{q: locked}).ReplaceQuestions(ctx, ReplaceCbtPackageQuestionsInput{PackageID: packageID, QuestionIDs: []pgtype.UUID{firstQuestionID}})
	if err == nil || !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "mengubah soal") || locked.deleteQuestionCalls != 0 || len(locked.addArgs) != 0 {
		t.Fatalf("ReplaceQuestions(locked) err/delete/add = %v/%d/%d, want conflict before mutation", err, locked.deleteQuestionCalls, len(locked.addArgs))
	}
}

func TestCbtPackageDeepCloneTrimsTitleCopiesQuestionsAndRefreshesTarget(t *testing.T) {
	ctx := context.Background()
	sourceID := pgtype.UUID{Bytes: [16]byte{14, 1}, Valid: true}
	targetID := pgtype.UUID{Bytes: [16]byte{14, 2}, Valid: true}
	store := &fakeCbtPackageEditStore{
		cloneRow:  db.CbtPackage{ID: targetID, Title: "PAT Copy"},
		detailRow: db.GetCbtPackageDetailRow{ID: targetID, Title: "PAT Copy"},
	}

	got, err := (&CbtPackage{q: store}).Clone(ctx, CloneCbtPackageInput{SourceID: sourceID, Title: "  PAT Copy  "})
	if err != nil {
		t.Fatalf("Clone(success) error = %v", err)
	}
	if store.cloneArg.SourceID != sourceID || store.cloneArg.Title != "PAT Copy" {
		t.Fatalf("CloneCbtPackage arg = %+v, want source id and trimmed title", store.cloneArg)
	}
	if store.cloneQuestionsArg.SourceID != sourceID || store.cloneQuestionsArg.TargetID != targetID {
		t.Fatalf("CloneCbtPackageQuestions arg = %+v, want source -> target", store.cloneQuestionsArg)
	}
	if got.Package.ID != targetID {
		t.Fatalf("Clone result package = %v, want refreshed target %v", got.Package.ID, targetID)
	}

	_, err = (&CbtPackage{q: &fakeCbtPackageStore{}}).Clone(ctx, CloneCbtPackageInput{SourceID: sourceID, Title: "PAT Copy"})
	if err == nil || !strings.Contains(err.Error(), "clone store unavailable") {
		t.Fatalf("Clone(unavailable store) error = %v, want unavailable store", err)
	}
}

func TestCbtPackageDeepLockSnapshotDeleteSuccessAndNotFoundMapping(t *testing.T) {
	ctx := context.Background()
	packageID := pgtype.UUID{Bytes: [16]byte{15, 1}, Valid: true}
	lockedBy := pgtype.UUID{Bytes: [16]byte{15, 2}, Valid: true}

	_, err := lockCbtPackageSnapshot(ctx, &fakeCbtPackageSnapshotStore{lockErr: pgx.ErrNoRows}, packageID, lockedBy, "manual")
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("lockCbtPackageSnapshot(no rows) error = %v, want raw pgx.ErrNoRows", err)
	}

	deleteStore := &fakeCbtPackageDeleteStore{deleteRows: 1}
	if err := (&CbtPackage{q: deleteStore}).Delete(ctx, packageID); err != nil {
		t.Fatalf("Delete(success) error = %v", err)
	}
	if deleteStore.deleteID != packageID {
		t.Fatalf("Delete(success) id = %v, want %v", deleteStore.deleteID, packageID)
	}
}

type fakeCbtPackageCreateRowsStore struct {
	createRow   db.CbtPackage
	createCalls int
	createArg   db.CreateCbtPackageParams
	createErr   error

	questions   map[pgtype.UUID]db.GetCbtQuestionRow
	questionErr error

	addRows int64
	addErr  error
	addArgs []db.AddCbtPackageQuestionParams
}

func (f *fakeCbtPackageCreateRowsStore) CreateCbtPackage(_ context.Context, arg db.CreateCbtPackageParams) (db.CbtPackage, error) {
	f.createCalls++
	f.createArg = arg
	return f.createRow, f.createErr
}

func (f *fakeCbtPackageCreateRowsStore) GetCbtQuestion(_ context.Context, id pgtype.UUID) (db.GetCbtQuestionRow, error) {
	if f.questionErr != nil {
		return db.GetCbtQuestionRow{}, f.questionErr
	}
	question, ok := f.questions[id]
	if !ok {
		return db.GetCbtQuestionRow{}, pgx.ErrNoRows
	}
	return question, nil
}

func (f *fakeCbtPackageCreateRowsStore) AddCbtPackageQuestion(_ context.Context, arg db.AddCbtPackageQuestionParams) (int64, error) {
	f.addArgs = append(f.addArgs, arg)
	return f.addRows, f.addErr
}
