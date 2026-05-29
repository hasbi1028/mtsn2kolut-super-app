package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestCbtPackageUltraCreateSelectionValidationBranches(t *testing.T) {
	ctx := context.Background()
	eventID := pgtype.UUID{Bytes: [16]byte{21, 1}, Valid: true}
	subjectID := pgtype.UUID{Bytes: [16]byte{21, 2}, Valid: true}
	packageID := pgtype.UUID{Bytes: [16]byte{21, 3}, Valid: true}
	questionID := pgtype.UUID{Bytes: [16]byte{21, 4}, Valid: true}
	otherEventQuestionID := pgtype.UUID{Bytes: [16]byte{21, 5}, Valid: true}
	otherEventID := pgtype.UUID{Bytes: [16]byte{21, 6}, Valid: true}
	archivedQuestionID := pgtype.UUID{Bytes: [16]byte{21, 7}, Valid: true}

	tests := []struct {
		name         string
		questionIDs  []pgtype.UUID
		questions    map[pgtype.UUID]db.GetCbtQuestionRow
		weights      map[string]int32
		wantIs       error
		wantContains string
		wantAdds     int
		wantLookups  bool
	}{
		{
			name:         "duplicate question rejected before second lookup",
			questionIDs:  []pgtype.UUID{questionID, questionID},
			questions:    map[pgtype.UUID]db.GetCbtQuestionRow{questionID: {ID: questionID, SubjectID: subjectID, Status: db.CbtQuestionStatusEnumPublished}},
			wantIs:       domain.ErrBadRequest,
			wantContains: "duplikat",
			wantLookups:  true,
		},
		{
			name:         "event package rejects question from another event",
			questionIDs:  []pgtype.UUID{otherEventQuestionID},
			questions:    map[pgtype.UUID]db.GetCbtQuestionRow{otherEventQuestionID: {ID: otherEventQuestionID, EventID: otherEventID, SubjectID: subjectID, Status: db.CbtQuestionStatusEnumPublished}},
			wantIs:       domain.ErrBadRequest,
			wantContains: "event yang sama",
			wantLookups:  true,
		},
		{
			name:         "archived question rejected even when workflow published",
			questionIDs:  []pgtype.UUID{archivedQuestionID},
			questions:    map[pgtype.UUID]db.GetCbtQuestionRow{archivedQuestionID: {ID: archivedQuestionID, SubjectID: subjectID, Status: db.CbtQuestionStatusEnumArchived, WorkflowStatus: "published"}},
			wantIs:       domain.ErrConflict,
			wantContains: "Siap Pakai atau sudah terbit",
			wantLookups:  true,
		},
		{
			name:         "weight below one rejected",
			questionIDs:  []pgtype.UUID{questionID},
			questions:    map[pgtype.UUID]db.GetCbtQuestionRow{questionID: {ID: questionID, SubjectID: subjectID, Status: db.CbtQuestionStatusEnumPublished}},
			weights:      map[string]int32{pgUUIDString(questionID): 0},
			wantContains: "bobot soal",
			wantLookups:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeCbtPackageCreateStore{
				createRow: db.CbtPackage{ID: packageID, EventID: eventID, SubjectID: subjectID},
				questions: tt.questions,
			}
			_, err := createCbtPackage(ctx, store, CreateCbtPackageInput{
				EventID:         eventID,
				SubjectID:       subjectID,
				Title:           "PAT IPA",
				QuestionIDs:     tt.questionIDs,
				QuestionWeights: tt.weights,
			})
			if err == nil {
				t.Fatalf("createCbtPackage() error = nil, want %q", tt.wantContains)
			}
			if tt.wantIs != nil && !errors.Is(err, tt.wantIs) {
				t.Fatalf("createCbtPackage() error = %v, want errors.Is %v", err, tt.wantIs)
			}
			if !strings.Contains(err.Error(), tt.wantContains) {
				t.Fatalf("createCbtPackage() error = %v, want contains %q", err, tt.wantContains)
			}
			if store.createCalls != 1 {
				t.Fatalf("CreateCbtPackage calls = %d, want package row created before selection validation", store.createCalls)
			}
			if len(store.addParams) != tt.wantAdds {
				t.Fatalf("AddCbtPackageQuestion calls = %d, want %d", len(store.addParams), tt.wantAdds)
			}
		})
	}

	missingStore := &fakeCbtPackageCreateRowsStore{
		createRow: db.CbtPackage{ID: packageID, EventID: eventID, SubjectID: subjectID},
		questions: map[pgtype.UUID]db.GetCbtQuestionRow{},
	}
	_, err := createCbtPackage(ctx, missingStore, CreateCbtPackageInput{
		EventID:     eventID,
		SubjectID:   subjectID,
		Title:       "PAT IPA",
		QuestionIDs: []pgtype.UUID{questionID},
	})
	if !errors.Is(err, domain.ErrNotFound) || missingStore.createCalls != 1 || len(missingStore.addArgs) != 0 {
		t.Fatalf("createCbtPackage(missing question) err/create/add = %v/%d/%d, want not found after create before add", err, missingStore.createCalls, len(missingStore.addArgs))
	}
}

func TestCbtPackageUltraReplaceQuestionsValidationStopsBeforeMutation(t *testing.T) {
	ctx := context.Background()
	packageID := pgtype.UUID{Bytes: [16]byte{22, 1}, Valid: true}
	eventID := pgtype.UUID{Bytes: [16]byte{22, 2}, Valid: true}
	subjectID := pgtype.UUID{Bytes: [16]byte{22, 3}, Valid: true}
	questionID := pgtype.UUID{Bytes: [16]byte{22, 4}, Valid: true}
	otherSubjectID := pgtype.UUID{Bytes: [16]byte{22, 5}, Valid: true}

	tests := []struct {
		name         string
		store        *fakeCbtPackageEditStore
		input        ReplaceCbtPackageQuestionsInput
		wantIs       error
		wantContains string
	}{
		{
			name: "question lookup no rows maps not found",
			store: &fakeCbtPackageEditStore{
				lockRow:   db.LockCbtPackageForEditRow{ID: packageID, EventID: eventID, SubjectID: subjectID},
				questions: map[pgtype.UUID]db.GetCbtQuestionRow{},
			},
			input:  ReplaceCbtPackageQuestionsInput{PackageID: packageID, QuestionIDs: []pgtype.UUID{questionID}},
			wantIs: domain.ErrNotFound,
		},
		{
			name: "question subject mismatch rejected",
			store: &fakeCbtPackageEditStore{
				lockRow:   db.LockCbtPackageForEditRow{ID: packageID, EventID: eventID, SubjectID: subjectID},
				questions: map[pgtype.UUID]db.GetCbtQuestionRow{questionID: {ID: questionID, SubjectID: otherSubjectID, Status: db.CbtQuestionStatusEnumPublished}},
			},
			input:        ReplaceCbtPackageQuestionsInput{PackageID: packageID, QuestionIDs: []pgtype.UUID{questionID}},
			wantContains: "mapel yang sama",
		},
		{
			name: "duplicate replacement rejected",
			store: &fakeCbtPackageEditStore{
				lockRow:   db.LockCbtPackageForEditRow{ID: packageID, EventID: eventID, SubjectID: subjectID},
				questions: map[pgtype.UUID]db.GetCbtQuestionRow{questionID: {ID: questionID, EventID: eventID, SubjectID: subjectID, Status: db.CbtQuestionStatusEnumPublished}},
			},
			input:        ReplaceCbtPackageQuestionsInput{PackageID: packageID, QuestionIDs: []pgtype.UUID{questionID, questionID}},
			wantIs:       domain.ErrBadRequest,
			wantContains: "duplikat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := (&CbtPackage{q: tt.store}).ReplaceQuestions(ctx, tt.input)
			if err == nil {
				t.Fatalf("ReplaceQuestions() error = nil, want failure")
			}
			if tt.wantIs != nil && !errors.Is(err, tt.wantIs) {
				t.Fatalf("ReplaceQuestions() error = %v, want errors.Is %v", err, tt.wantIs)
			}
			if tt.wantContains != "" && !strings.Contains(err.Error(), tt.wantContains) {
				t.Fatalf("ReplaceQuestions() error = %v, want contains %q", err, tt.wantContains)
			}
			if tt.store.deleteQuestionCalls != 0 || len(tt.store.addArgs) != 0 {
				t.Fatalf("ReplaceQuestions() delete/add = %d/%d, want no mutation after validation failure", tt.store.deleteQuestionCalls, len(tt.store.addArgs))
			}
		})
	}
}

func TestCbtPackageUltraCloneErrorAndBlankTitleBranches(t *testing.T) {
	ctx := context.Background()
	sourceID := pgtype.UUID{Bytes: [16]byte{23, 1}, Valid: true}
	targetID := pgtype.UUID{Bytes: [16]byte{23, 2}, Valid: true}

	cloneErr := errors.New("clone insert failed")
	store := &fakeCbtPackageEditStore{cloneErr: cloneErr}
	_, err := (&CbtPackage{q: store}).Clone(ctx, CloneCbtPackageInput{SourceID: sourceID, Title: "PAT Copy"})
	if !errors.Is(err, cloneErr) {
		t.Fatalf("Clone(clone error) = %v, want %v", err, cloneErr)
	}
	if store.cloneQuestionsArg.SourceID.Valid || store.detailRow.ID.Valid {
		t.Fatalf("Clone(clone error) unexpectedly copied questions/detail: %+v/%+v", store.cloneQuestionsArg, store.detailRow)
	}

	store = &fakeCbtPackageEditStore{
		cloneRow:  db.CbtPackage{ID: targetID, Title: ""},
		detailRow: db.GetCbtPackageDetailRow{ID: targetID, Title: ""},
	}
	got, err := (&CbtPackage{q: store}).Clone(ctx, CloneCbtPackageInput{SourceID: sourceID, Title: "   "})
	if err != nil {
		t.Fatalf("Clone(blank title) error = %v", err)
	}
	if store.cloneArg.Title != "" || store.cloneQuestionsArg.TargetID != targetID || got.Package.ID != targetID {
		t.Fatalf("Clone(blank title) clone/detail = arg:%+v copy:%+v got:%+v, want blank title copied to target and detail refresh", store.cloneArg, store.cloneQuestionsArg, got.Package)
	}
}

func TestCbtPackageUltraLockAndSnapshotDefaultReasonSuccessAndSnapshotError(t *testing.T) {
	ctx := context.Background()
	packageID := pgtype.UUID{Bytes: [16]byte{24, 1}, Valid: true}
	lockedBy := pgtype.UUID{Bytes: [16]byte{24, 2}, Valid: true}
	lockedAt := time.Date(2026, 5, 17, 12, 34, 56, 0, time.UTC)

	store := &fakeCbtPackageSnapshotStore{
		lockRow: db.LockCbtPackageForSnapshotRow{
			ID:              packageID,
			LockedAt:        pgtype.Timestamptz{Time: lockedAt, Valid: true},
			LockReason:      "session_scheduled",
			SnapshotVersion: 7,
		},
		snapshotRows: 11,
	}
	got, err := (&CbtPackage{q: store}).LockAndSnapshot(ctx, packageID, lockedBy, "  ")
	if err != nil {
		t.Fatalf("LockAndSnapshot(default reason) error = %v", err)
	}
	if store.lockArg.PackageID != packageID || store.lockArg.LockedBy != lockedBy || store.lockArg.LockReason != "session_scheduled" {
		t.Fatalf("LockCbtPackageForSnapshot arg = %+v, want ids and default reason", store.lockArg)
	}
	if store.snapshotPackageID != packageID || store.snapshotCalls != 1 {
		t.Fatalf("CreateCbtPackageQuestionSnapshots package/calls = %v/%d, want %v/1", store.snapshotPackageID, store.snapshotCalls, packageID)
	}
	if got.PackageID != pgUUIDString(packageID) || got.LockedAt != "2026-05-17T12:34:56Z" || got.LockReason != "session_scheduled" || got.SnapshotVersion != 7 || got.SnapshotRowsAdded != 11 {
		t.Fatalf("LockAndSnapshot(default reason) = %+v, want formatted snapshot result", got)
	}

	snapshotErr := errors.New("snapshot insert failed")
	store = &fakeCbtPackageSnapshotStore{
		lockRow:     db.LockCbtPackageForSnapshotRow{ID: packageID, LockReason: "manual", SnapshotVersion: 1},
		snapshotErr: snapshotErr,
	}
	_, err = (&CbtPackage{q: store}).LockAndSnapshot(ctx, packageID, lockedBy, "manual")
	if !errors.Is(err, snapshotErr) || store.snapshotCalls != 1 {
		t.Fatalf("LockAndSnapshot(snapshot error) err/calls = %v/%d, want snapshot error after one snapshot attempt", err, store.snapshotCalls)
	}

	_, err = lockCbtPackageSnapshot(ctx, &fakeCbtPackageSnapshotStore{lockErr: pgx.ErrNoRows}, packageID, lockedBy, "manual")
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("lockCbtPackageSnapshot(no rows passthrough) = %v, want pgx.ErrNoRows", err)
	}
}
