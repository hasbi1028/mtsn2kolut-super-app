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

func TestCreateCbtPackageBuildsMetadataAndValidatesInput(t *testing.T) {
	subjectID := pgtype.UUID{Bytes: [16]byte{111}, Valid: true}
	packageID := pgtype.UUID{Bytes: [16]byte{112}, Valid: true}
	firstQuestionID := pgtype.UUID{Bytes: [16]byte{113}, Valid: true}
	secondQuestionID := pgtype.UUID{Bytes: [16]byte{114}, Valid: true}
	store := &fakeCbtPackageCreateStore{
		createRow: db.CbtPackage{ID: packageID, SubjectID: subjectID},
		questions: map[pgtype.UUID]db.GetCbtQuestionRow{
			firstQuestionID:  {ID: firstQuestionID, SubjectID: subjectID, Status: db.CbtQuestionStatusEnumPublished},
			secondQuestionID: {ID: secondQuestionID, SubjectID: subjectID, Status: db.CbtQuestionStatusEnumPublished},
		},
	}

	got, err := createCbtPackage(context.Background(), store, CreateCbtPackageInput{
		SubjectID:          subjectID,
		Title:              "  PAT IPA  ",
		Description:        "  semester genap  ",
		DurationMinutes:    0,
		RandomizeQuestions: true,
		RandomizeOptions:   true,
		QuestionIDs:        []pgtype.UUID{firstQuestionID, secondQuestionID},
		QuestionWeights: map[string]int32{
			pgUUIDString(secondQuestionID): 7,
		},
		DrawPgCount:    20,
		DrawEssayCount: 5,
		RandomSeed:     "  seed-a  ",
		IsActive:       true,
	})
	if err != nil {
		t.Fatalf("createCbtPackage() error = %v", err)
	}
	if got.ID != packageID || store.createCalls != 1 {
		t.Fatalf("createCbtPackage() package/calls = %v/%d, want %v/1", got.ID, store.createCalls, packageID)
	}
	if store.createParams.Title != "PAT IPA" || store.createParams.Description != "semester genap" || store.createParams.RandomSeed != "seed-a" {
		t.Fatalf("CreateCbtPackage trims = %+v, want trimmed title/description/random seed", store.createParams)
	}
	if store.createParams.DurationMinutes != 60 || store.createParams.SourceMode != "teacher_class" {
		t.Fatalf("CreateCbtPackage normalized = %+v, want duration 60/source teacher_class", store.createParams)
	}
	if !store.createParams.RandomizeQuestions || !store.createParams.RandomizeOptions || !store.createParams.IsActive || store.createParams.DrawPgCount != 20 || store.createParams.DrawEssayCount != 5 {
		t.Fatalf("CreateCbtPackage flags = %+v, want input flags preserved", store.createParams)
	}
	if !strings.Contains(string(store.createParams.CompositionLog), `"selected_question_count":2`) || !strings.Contains(string(store.createParams.CompositionLog), `"source_mode":"teacher_class"`) {
		t.Fatalf("CompositionLog = %s, want source mode and selected count", string(store.createParams.CompositionLog))
	}
	if len(store.addParams) != 2 || store.addParams[0].Position != 1 || store.addParams[0].Points != 1 || store.addParams[1].Position != 2 || store.addParams[1].Points != 7 {
		t.Fatalf("AddCbtPackageQuestion params = %+v, want ordered default/weighted points", store.addParams)
	}

	tests := []struct {
		name  string
		input CreateCbtPackageInput
		want  string
	}{
		{name: "blank title", input: CreateCbtPackageInput{SubjectID: subjectID, Title: " ", QuestionIDs: []pgtype.UUID{firstQuestionID}}, want: "nama paket wajib diisi"},
		{name: "duration too high", input: CreateCbtPackageInput{SubjectID: subjectID, Title: "PAT", DurationMinutes: 361, QuestionIDs: []pgtype.UUID{firstQuestionID}}, want: "durasi paket CBT"},
		{name: "invalid source mode", input: CreateCbtPackageInput{SubjectID: subjectID, Title: "PAT", SourceMode: "upload", QuestionIDs: []pgtype.UUID{firstQuestionID}}, want: "mode sumber paket tidak valid"},
		{name: "negative draw", input: CreateCbtPackageInput{SubjectID: subjectID, Title: "PAT", DrawPgCount: -1, QuestionIDs: []pgtype.UUID{firstQuestionID}}, want: "jumlah draw soal tidak boleh negatif"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationStore := &fakeCbtPackageCreateStore{questions: map[pgtype.UUID]db.GetCbtQuestionRow{firstQuestionID: {ID: firstQuestionID, SubjectID: subjectID, Status: db.CbtQuestionStatusEnumPublished}}}
			_, err := createCbtPackage(context.Background(), validationStore, tt.input)
			if err == nil || !errors.Is(err, domain.ErrBadRequest) || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("createCbtPackage() error = %v, want ErrBadRequest containing %q", err, tt.want)
			}
			if validationStore.createCalls != 0 {
				t.Fatalf("CreateCbtPackage calls = %d, want 0 on validation failure", validationStore.createCalls)
			}
		})
	}
}

func TestCbtPackageReplaceQuestionsEdgeCases(t *testing.T) {
	packageID := pgtype.UUID{Bytes: [16]byte{121}, Valid: true}
	subjectID := pgtype.UUID{Bytes: [16]byte{122}, Valid: true}
	questionID := pgtype.UUID{Bytes: [16]byte{123}, Valid: true}

	t.Run("allows empty replacement by clearing package questions", func(t *testing.T) {
		store := &fakeCbtPackageEditStore{
			lockRow:   db.LockCbtPackageForEditRow{ID: packageID, SubjectID: subjectID},
			detailRow: db.GetCbtPackageDetailRow{ID: packageID, SubjectID: subjectID},
		}
		got, err := (&CbtPackage{q: store}).ReplaceQuestions(context.Background(), ReplaceCbtPackageQuestionsInput{PackageID: packageID})
		if err != nil {
			t.Fatalf("ReplaceQuestions(empty) error = %v", err)
		}
		if store.deleteQuestionCalls != 1 || len(store.addArgs) != 0 {
			t.Fatalf("ReplaceQuestions(empty) delete/add = %d/%d, want 1/0", store.deleteQuestionCalls, len(store.addArgs))
		}
		if got.Package.ID != packageID || got.Readiness.Status != "kosong" {
			t.Fatalf("ReplaceQuestions(empty) result = %+v, want empty refreshed detail", got)
		}
	})

	t.Run("rejects invalid weight before mutating", func(t *testing.T) {
		store := &fakeCbtPackageEditStore{
			lockRow: db.LockCbtPackageForEditRow{ID: packageID, SubjectID: subjectID},
			questions: map[pgtype.UUID]db.GetCbtQuestionRow{
				questionID: {ID: questionID, SubjectID: subjectID, Status: db.CbtQuestionStatusEnumPublished},
			},
		}
		_, err := (&CbtPackage{q: store}).ReplaceQuestions(context.Background(), ReplaceCbtPackageQuestionsInput{
			PackageID:   packageID,
			QuestionIDs: []pgtype.UUID{questionID},
			QuestionWeights: map[string]int32{
				pgUUIDString(questionID): 101,
			},
		})
		if err == nil || !strings.Contains(err.Error(), "bobot soal") {
			t.Fatalf("ReplaceQuestions(invalid weight) error = %v, want weight validation", err)
		}
		if store.deleteQuestionCalls != 0 || len(store.addArgs) != 0 {
			t.Fatalf("ReplaceQuestions(invalid weight) delete/add = %d/%d, want no mutation", store.deleteQuestionCalls, len(store.addArgs))
		}
	})
}

func TestCbtPackageCloneValidationAndDetailErrors(t *testing.T) {
	sourceID := pgtype.UUID{Bytes: [16]byte{131}, Valid: true}
	targetID := pgtype.UUID{Bytes: [16]byte{132}, Valid: true}

	_, err := (&CbtPackage{q: &fakeCbtPackageEditStore{}}).Clone(context.Background(), CloneCbtPackageInput{SourceID: sourceID, Title: "  "})
	if err != nil {
		t.Fatalf("Clone(blank title) error = %v, want allowed blank title", err)
	}

	_, err = (&CbtPackage{q: &fakeCbtPackageEditStore{cloneRow: db.CbtPackage{ID: targetID}, detailErr: errors.New("detail failed")}}).Clone(context.Background(), CloneCbtPackageInput{SourceID: sourceID, Title: "PAT Copy"})
	if err == nil || err.Error() != "detail failed" {
		t.Fatalf("Clone(detail error) error = %v, want detail failed", err)
	}
}

func TestCbtPackageUpdateMetadataValidationAndDetailErrors(t *testing.T) {
	packageID := pgtype.UUID{Bytes: [16]byte{141}, Valid: true}
	store := &fakeCbtPackageEditStore{lockRow: db.LockCbtPackageForEditRow{ID: packageID}}
	_, err := (&CbtPackage{q: store}).UpdateMetadata(context.Background(), UpdateCbtPackageInput{ID: packageID, Title: "", DurationMinutes: 90})
	if err == nil || !errors.Is(err, domain.ErrBadRequest) || store.updateCalls != 0 {
		t.Fatalf("UpdateMetadata(validation) err/calls = %v/%d, want bad request before update", err, store.updateCalls)
	}

	detailErr := errors.New("detail refresh failed")
	store = &fakeCbtPackageEditStore{lockRow: db.LockCbtPackageForEditRow{ID: packageID}, detailErr: detailErr}
	_, err = (&CbtPackage{q: store}).UpdateMetadata(context.Background(), UpdateCbtPackageInput{ID: packageID, Title: "PAT", DurationMinutes: 90})
	if !errors.Is(err, detailErr) || store.updateCalls != 1 {
		t.Fatalf("UpdateMetadata(detail error) err/calls = %v/%d, want detail error after update", err, store.updateCalls)
	}
}

type fakeCbtPackageDeleteStore struct {
	usageCount int32
	usageErr   error
	deleteRows int64
	deleteErr  error
	deleteID   pgtype.UUID
}

func (f *fakeCbtPackageDeleteStore) ListCbtPackages(context.Context, pgtype.UUID) ([]db.ListCbtPackagesRow, error) {
	return nil, nil
}

func (f *fakeCbtPackageDeleteStore) ListCbtPackageQuestions(context.Context, pgtype.UUID) ([]db.ListCbtPackageQuestionsRow, error) {
	return nil, nil
}

func (f *fakeCbtPackageDeleteStore) GetCbtPackageUsage(context.Context, pgtype.UUID) (int32, error) {
	return f.usageCount, f.usageErr
}

func (f *fakeCbtPackageDeleteStore) DeleteCbtPackage(_ context.Context, id pgtype.UUID) (int64, error) {
	f.deleteID = id
	return f.deleteRows, f.deleteErr
}

func (f *fakeCbtPackageDeleteStore) ArchiveCbtPackage(context.Context, db.ArchiveCbtPackageParams) (int64, error) {
	return 1, nil
}

func (f *fakeCbtPackageDeleteStore) WithTx(pgx.Tx) *db.Queries { return nil }

func TestCbtPackageDeleteErrorPaths(t *testing.T) {
	packageID := pgtype.UUID{Bytes: [16]byte{151}, Valid: true}

	err := (&CbtPackage{q: &fakeCbtPackageDeleteStore{usageErr: pgx.ErrNoRows}}).Delete(context.Background(), packageID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Delete(usage missing) error = %v, want ErrNotFound", err)
	}

	usageFailed := errors.New("usage failed")
	err = (&CbtPackage{q: &fakeCbtPackageDeleteStore{usageErr: usageFailed}}).Delete(context.Background(), packageID)
	if !errors.Is(err, usageFailed) {
		t.Fatalf("Delete(usage error) error = %v, want usage error", err)
	}

	store := &fakeCbtPackageDeleteStore{deleteRows: 0}
	err = (&CbtPackage{q: store}).Delete(context.Background(), packageID)
	if err == nil || !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "tidak ditemukan atau sudah terkunci") || store.deleteID != packageID {
		t.Fatalf("Delete(no rows) err/id = %v/%v, want conflict after delete attempt", err, store.deleteID)
	}

	deleteFailed := errors.New("delete failed")
	err = (&CbtPackage{q: &fakeCbtPackageDeleteStore{deleteRows: 1, deleteErr: deleteFailed}}).Delete(context.Background(), packageID)
	if !errors.Is(err, deleteFailed) {
		t.Fatalf("Delete(delete error) error = %v, want delete error", err)
	}
}

func TestCbtPackageLockAndSnapshotAdditionalErrorPaths(t *testing.T) {
	packageID := pgtype.UUID{Bytes: [16]byte{161}, Valid: true}
	lockedBy := pgtype.UUID{Bytes: [16]byte{162}, Valid: true}

	_, err := (&CbtPackage{q: &fakeCbtPackageStore{}}).LockAndSnapshot(context.Background(), packageID, lockedBy, "manual")
	if err == nil || !strings.Contains(err.Error(), "snapshot store unavailable") {
		t.Fatalf("LockAndSnapshot(unavailable store) error = %v, want unavailable store", err)
	}

	store := &fakeCbtPackageSnapshotStore{
		lockRow:      db.LockCbtPackageForSnapshotRow{ID: packageID, LockReason: "manual", SnapshotVersion: 3},
		snapshotRows: 0,
	}
	got, err := (&CbtPackage{q: store}).LockAndSnapshot(context.Background(), packageID, lockedBy, " manual ")
	if err != nil {
		t.Fatalf("LockAndSnapshot(no locked_at) error = %v", err)
	}
	if got.LockedAt != "" || got.LockReason != "manual" || got.SnapshotVersion != 3 || got.SnapshotRowsAdded != 0 {
		t.Fatalf("LockAndSnapshot(no locked_at) = %+v, want empty locked_at and snapshot metadata", got)
	}
}
