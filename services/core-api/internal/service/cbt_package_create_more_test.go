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

func TestCbtPackageCreateMoreValidationBranches(t *testing.T) {
	ctx := context.Background()
	subjectID := pgtype.UUID{Bytes: [16]byte{171}, Valid: true}
	otherSubjectID := pgtype.UUID{Bytes: [16]byte{172}, Valid: true}
	packageID := pgtype.UUID{Bytes: [16]byte{173}, Valid: true}
	questionID := pgtype.UUID{Bytes: [16]byte{174}, Valid: true}
	baseQuestion := db.GetCbtQuestionRow{ID: questionID, SubjectID: subjectID, Status: db.CbtQuestionStatusEnumPublished}

	if _, err := (&CbtPackage{}).Create(ctx, CreateCbtPackageInput{Title: "PAT IPA"}); err == nil || !strings.Contains(err.Error(), "question_ids wajib diisi") {
		t.Fatalf("Create(no questions) error = %v, want question_ids validation before transaction", err)
	}

	for _, mode := range []string{"teacher_class", "level_subject_teachers", "event_pool"} {
		t.Run("accepts source mode "+mode, func(t *testing.T) {
			store := &fakeCbtPackageCreateStore{
				createRow: db.CbtPackage{ID: packageID, SubjectID: subjectID},
				questions: map[pgtype.UUID]db.GetCbtQuestionRow{questionID: baseQuestion},
			}
			_, err := createCbtPackage(ctx, store, CreateCbtPackageInput{SubjectID: subjectID, Title: "PAT IPA", SourceMode: mode, DurationMinutes: 1, QuestionIDs: []pgtype.UUID{questionID}})
			if err != nil {
				t.Fatalf("createCbtPackage(source %s) error = %v", mode, err)
			}
			if store.createParams.SourceMode != mode || store.createParams.DurationMinutes != 1 {
				t.Fatalf("CreateCbtPackage params = %+v, want source %s and duration 1", store.createParams, mode)
			}
		})
	}

	tests := []struct {
		name      string
		input     CreateCbtPackageInput
		store     *fakeCbtPackageCreateStore
		wantIs    error
		wantError string
	}{
		{
			name:      "duration too low",
			input:     CreateCbtPackageInput{SubjectID: subjectID, Title: "PAT IPA", DurationMinutes: -1, QuestionIDs: []pgtype.UUID{questionID}},
			store:     &fakeCbtPackageCreateStore{},
			wantIs:    domain.ErrBadRequest,
			wantError: "durasi paket CBT",
		},
		{
			name:      "zero uuid question",
			input:     CreateCbtPackageInput{SubjectID: subjectID, Title: "PAT IPA", QuestionIDs: []pgtype.UUID{{}}},
			store:     &fakeCbtPackageCreateStore{createRow: db.CbtPackage{ID: packageID, SubjectID: subjectID}},
			wantIs:    domain.ErrBadRequest,
			wantError: "soal tidak valid",
		},
		{
			name:      "subject mismatch",
			input:     CreateCbtPackageInput{SubjectID: subjectID, Title: "PAT IPA", QuestionIDs: []pgtype.UUID{questionID}},
			store:     &fakeCbtPackageCreateStore{createRow: db.CbtPackage{ID: packageID, SubjectID: subjectID}, questions: map[pgtype.UUID]db.GetCbtQuestionRow{questionID: {ID: questionID, SubjectID: otherSubjectID, Status: db.CbtQuestionStatusEnumPublished}}},
			wantError: "mapel yang sama",
		},
		{
			name:      "question lookup error",
			input:     CreateCbtPackageInput{SubjectID: subjectID, Title: "PAT IPA", QuestionIDs: []pgtype.UUID{questionID}},
			store:     &fakeCbtPackageCreateStore{createRow: db.CbtPackage{ID: packageID, SubjectID: subjectID}, questionErr: errors.New("question lookup failed")},
			wantError: "question lookup failed",
		},
		{
			name:      "create package error",
			input:     CreateCbtPackageInput{SubjectID: subjectID, Title: "PAT IPA", QuestionIDs: []pgtype.UUID{questionID}},
			store:     &fakeCbtPackageCreateStore{createErr: errors.New("create failed"), questions: map[pgtype.UUID]db.GetCbtQuestionRow{questionID: baseQuestion}},
			wantError: "create failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := createCbtPackage(ctx, tt.store, tt.input)
			if err == nil {
				t.Fatalf("createCbtPackage() error = nil, want %q", tt.wantError)
			}
			if tt.wantIs != nil && !errors.Is(err, tt.wantIs) {
				t.Fatalf("createCbtPackage() error = %v, want errors.Is %v", err, tt.wantIs)
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("createCbtPackage() error = %v, want contains %q", err, tt.wantError)
			}
			if len(tt.store.addParams) != 0 {
				t.Fatalf("AddCbtPackageQuestion calls = %d, want 0 on failure", len(tt.store.addParams))
			}
		})
	}
}

func TestCbtPackageReplaceQuestionsMoreErrorBranches(t *testing.T) {
	ctx := context.Background()
	packageID := pgtype.UUID{Bytes: [16]byte{181}, Valid: true}
	subjectID := pgtype.UUID{Bytes: [16]byte{182}, Valid: true}
	questionID := pgtype.UUID{Bytes: [16]byte{183}, Valid: true}

	_, err := (&CbtPackage{q: &fakeCbtPackageStore{}}).ReplaceQuestions(ctx, ReplaceCbtPackageQuestionsInput{PackageID: packageID, QuestionIDs: []pgtype.UUID{questionID}})
	if err == nil || !strings.Contains(err.Error(), "edit store unavailable") {
		t.Fatalf("ReplaceQuestions(unavailable store) error = %v, want unavailable store", err)
	}

	lockErr := errors.New("lock failed")
	_, err = (&CbtPackage{q: &fakeCbtPackageEditStore{lockErr: lockErr}}).ReplaceQuestions(ctx, ReplaceCbtPackageQuestionsInput{PackageID: packageID, QuestionIDs: []pgtype.UUID{questionID}})
	if !errors.Is(err, lockErr) {
		t.Fatalf("ReplaceQuestions(lock generic error) = %v, want %v", err, lockErr)
	}

	addErr := errors.New("add failed")
	store := &fakeCbtPackageEditStore{
		lockRow: db.LockCbtPackageForEditRow{ID: packageID, SubjectID: subjectID},
		questions: map[pgtype.UUID]db.GetCbtQuestionRow{
			questionID: {ID: questionID, SubjectID: subjectID, Status: db.CbtQuestionStatusEnumPublished},
		},
		addErr: addErr,
	}
	_, err = (&CbtPackage{q: store}).ReplaceQuestions(ctx, ReplaceCbtPackageQuestionsInput{PackageID: packageID, QuestionIDs: []pgtype.UUID{questionID}})
	if !errors.Is(err, addErr) || store.deleteQuestionCalls != 1 || len(store.addArgs) != 1 {
		t.Fatalf("ReplaceQuestions(add error) err/delete/add = %v/%d/%d, want add error after delete/add attempt", err, store.deleteQuestionCalls, len(store.addArgs))
	}

	detailErr := errors.New("detail questions failed")
	store = &fakeCbtPackageEditStore{
		lockRow: db.LockCbtPackageForEditRow{ID: packageID, SubjectID: subjectID},
		questions: map[pgtype.UUID]db.GetCbtQuestionRow{
			questionID: {ID: questionID, SubjectID: subjectID, Status: db.CbtQuestionStatusEnumPublished},
		},
		detailQErr: detailErr,
	}
	_, err = (&CbtPackage{q: store}).ReplaceQuestions(ctx, ReplaceCbtPackageQuestionsInput{PackageID: packageID, QuestionIDs: []pgtype.UUID{questionID}})
	if !errors.Is(err, detailErr) {
		t.Fatalf("ReplaceQuestions(detail questions error) = %v, want %v", err, detailErr)
	}
}

func TestCbtPackageLockSnapshotAndAddQuestionMoreBranches(t *testing.T) {
	ctx := context.Background()
	packageID := pgtype.UUID{Bytes: [16]byte{191}, Valid: true}
	lockedBy := pgtype.UUID{Bytes: [16]byte{192}, Valid: true}

	lockErr := errors.New("lock snapshot failed")
	_, err := (&CbtPackage{q: &fakeCbtPackageSnapshotStore{lockErr: lockErr}}).LockAndSnapshot(ctx, packageID, lockedBy, "manual")
	if !errors.Is(err, lockErr) {
		t.Fatalf("LockAndSnapshot(lock error) = %v, want %v", err, lockErr)
	}

	if err := addCbtPackageQuestion(ctx, &fakeCbtPackageAddRowsStore{rows: 0}, db.AddCbtPackageQuestionParams{PackageID: packageID}); err == nil || !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "approved/published") {
		t.Fatalf("addCbtPackageQuestion(rows=0) error = %v, want workflow conflict", err)
	}

	addErr := errors.New("insert failed")
	if err := addCbtPackageQuestion(ctx, &fakeCbtPackageAddRowsStore{err: addErr}, db.AddCbtPackageQuestionParams{PackageID: packageID}); !errors.Is(err, addErr) {
		t.Fatalf("addCbtPackageQuestion(error) = %v, want %v", err, addErr)
	}

	if _, err := (&CbtPackage{q: &fakeCbtPackageEditStore{detailErr: pgx.ErrNoRows}}).Detail(ctx, packageID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Detail(no rows) error = %v, want ErrNotFound", err)
	}
}

type fakeCbtPackageAddRowsStore struct {
	rows int64
	err  error
}

func (f *fakeCbtPackageAddRowsStore) AddCbtPackageQuestion(context.Context, db.AddCbtPackageQuestionParams) (int64, error) {
	return f.rows, f.err
}
