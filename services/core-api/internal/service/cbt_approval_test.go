package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeCbtApprovalStore struct {
	listArg   db.ListCbtApprovalRecordsParams
	listRows  []db.ListCbtApprovalRecordsRow
	listErr   error
	createArg db.CreateCbtApprovalRecordParams
	createRow db.CbtApprovalRecord
	createErr error
	revokeArg db.RevokeCbtApprovalRecordParams
	revokeRow db.CbtApprovalRecord
	revokeErr error
	eventRow  db.GetCbtEventOverviewSummaryRow
	eventErr  error
}

func (f *fakeCbtApprovalStore) ListCbtApprovalRecords(ctx context.Context, arg db.ListCbtApprovalRecordsParams) ([]db.ListCbtApprovalRecordsRow, error) {
	f.listArg = arg
	return f.listRows, f.listErr
}

func (f *fakeCbtApprovalStore) CreateCbtApprovalRecord(ctx context.Context, arg db.CreateCbtApprovalRecordParams) (db.CbtApprovalRecord, error) {
	f.createArg = arg
	return f.createRow, f.createErr
}

func (f *fakeCbtApprovalStore) RevokeCbtApprovalRecord(ctx context.Context, arg db.RevokeCbtApprovalRecordParams) (db.CbtApprovalRecord, error) {
	f.revokeArg = arg
	return f.revokeRow, f.revokeErr
}

func (f *fakeCbtApprovalStore) GetCbtEventOverviewSummary(ctx context.Context, id pgtype.UUID) (db.GetCbtEventOverviewSummaryRow, error) {
	if !f.eventRow.ID.Valid {
		f.eventRow.ID = id
	}
	return f.eventRow, f.eventErr
}

func TestCbtApprovalListFiltersAndNormalizesNilRows(t *testing.T) {
	entityID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000000101")
	store := &fakeCbtApprovalStore{}
	svc := NewCbtApproval(nil)
	svc.q = store

	rows, err := svc.List(context.Background(), ListCbtApprovalInput{EntityType: " event ", EntityID: entityID})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if rows == nil || len(rows) != 0 {
		t.Fatalf("List() rows = %#v, want non-nil empty slice", rows)
	}
	if !store.listArg.EntityType.Valid || store.listArg.EntityType.String != "event" {
		t.Fatalf("EntityType arg = %#v, want valid event", store.listArg.EntityType)
	}
	if store.listArg.EntityID != entityID {
		t.Fatalf("EntityID arg = %v, want %v", store.listArg.EntityID, entityID)
	}

	_, err = svc.List(context.Background(), ListCbtApprovalInput{EntityType: "bogus"})
	if !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("List(invalid) error = %v, want ErrBadRequest", err)
	}
}

func TestCbtApprovalApproveValidatesAndDelegatesTrimmedInput(t *testing.T) {
	entityID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000000102")
	actorID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000000103")
	wantRow := db.CbtApprovalRecord{EntityType: "package", EntityID: entityID, ApprovalType: "package_ready", ApprovedBy: actorID, Notes: "siap"}
	store := &fakeCbtApprovalStore{createRow: wantRow}
	svc := NewCbtApproval(nil)
	svc.q = store
	svc.events = store

	got, err := svc.Approve(context.Background(), SaveCbtApprovalInput{
		EntityType: " package ", EntityID: entityID, ApprovalType: " package_ready ", Notes: " siap ", ActorUserID: actorID,
	})
	if err != nil {
		t.Fatalf("Approve() error = %v", err)
	}
	if got != wantRow {
		t.Fatalf("Approve() = %+v, want %+v", got, wantRow)
	}
	if store.createArg.EntityType != "package" || store.createArg.ApprovalType != "package_ready" || store.createArg.Notes != "siap" {
		t.Fatalf("Create arg not trimmed/delegated: %+v", store.createArg)
	}
	if store.createArg.EntityID != entityID || store.createArg.ApprovedBy != actorID {
		t.Fatalf("Create UUID args = %+v, want entity/actor", store.createArg)
	}
}

func TestCbtApprovalApproveAndRevokeRejectInvalidInputs(t *testing.T) {
	validID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000000104")
	actorID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000000105")
	svc := NewCbtApproval(nil)
	svc.q = &fakeCbtApprovalStore{}
	svc.events = svc.q.(*fakeCbtApprovalStore)

	cases := []struct {
		name string
		in   SaveCbtApprovalInput
		want error
	}{
		{name: "bad entity type", in: SaveCbtApprovalInput{EntityType: "exam", EntityID: validID, ApprovalType: "package_ready", ActorUserID: actorID}, want: domain.ErrBadRequest},
		{name: "missing entity id", in: SaveCbtApprovalInput{EntityType: "event", ApprovalType: "package_ready", ActorUserID: actorID}, want: domain.ErrBadRequest},
		{name: "bad approval type", in: SaveCbtApprovalInput{EntityType: "event", EntityID: validID, ApprovalType: "bad", ActorUserID: actorID}, want: domain.ErrBadRequest},
		{name: "missing actor", in: SaveCbtApprovalInput{EntityType: "event", EntityID: validID, ApprovalType: "package_ready"}, want: domain.ErrUnauthorized},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Approve(context.Background(), tt.in)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Approve() error = %v, want %v", err, tt.want)
			}
		})
	}

	if _, err := svc.Revoke(context.Background(), pgtype.UUID{}, actorID, ""); !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("Revoke(missing id) error = %v, want ErrBadRequest", err)
	}
	if _, err := svc.Revoke(context.Background(), validID, pgtype.UUID{}, ""); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("Revoke(missing actor) error = %v, want ErrUnauthorized", err)
	}
}

func TestCbtApprovalRevokeDelegatesTrimmedNotes(t *testing.T) {
	approvalID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000000106")
	actorID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000000107")
	store := &fakeCbtApprovalStore{revokeRow: db.CbtApprovalRecord{ID: approvalID, RevokedBy: actorID, Notes: "cabut"}}
	svc := NewCbtApproval(nil)
	svc.q = store
	svc.events = store

	got, err := svc.Revoke(context.Background(), approvalID, actorID, " cabut ")
	if err != nil {
		t.Fatalf("Revoke() error = %v", err)
	}
	if got.ID != approvalID || got.RevokedBy != actorID || got.Notes != "cabut" {
		t.Fatalf("Revoke() = %+v, want revoked row", got)
	}
	if store.revokeArg.ID != approvalID || store.revokeArg.RevokedBy != actorID || store.revokeArg.Notes != "cabut" {
		t.Fatalf("Revoke arg = %+v, want trimmed notes and UUIDs", store.revokeArg)
	}
}

func TestCbtApprovalApproveRejectsEventMilestonesBeforeReady(t *testing.T) {
	entityID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000000108")
	actorID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000000109")
	store := &fakeCbtApprovalStore{
		eventRow: db.GetCbtEventOverviewSummaryRow{
			ID:                  entityID,
			Status:              "active",
			TargetQuestionCount: 10,
			PublishedQuestions:  10,
			PackageCount:        1,
			ActivePackageCount:  1,
			SessionCount:        0,
		},
	}
	svc := NewCbtApproval(nil)
	svc.q = store
	svc.events = store

	_, err := svc.Approve(context.Background(), SaveCbtApprovalInput{
		EntityType:   "event",
		EntityID:     entityID,
		ApprovalType: "participants_rooms_ready",
		ActorUserID:  actorID,
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("Approve(blocked event milestone) error = %v, want ErrConflict", err)
	}
}

func TestCbtApprovalApproveAllowsReadyEventMilestones(t *testing.T) {
	entityID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000000110")
	actorID := mustQuestionUUID(t, "00000000-0000-0000-0000-000000000111")
	store := &fakeCbtApprovalStore{
		createRow: db.CbtApprovalRecord{EntityType: "event", EntityID: entityID, ApprovalType: "final_archive"},
		eventRow: db.GetCbtEventOverviewSummaryRow{
			ID:                         entityID,
			Status:                     "finished",
			TargetQuestionCount:        10,
			PublishedQuestions:         10,
			PackageCount:               1,
			ActivePackageCount:         1,
			SessionCount:               1,
			RoomCount:                  1,
			ParticipantCount:           4,
			TokenReadyCount:            4,
			SubmittedCount:             4,
			ScoredCount:                4,
			UnassignedParticipantCount: 0,
			MissingSeatCount:           0,
			RoomsWithoutProctor:        0,
		},
	}
	svc := NewCbtApproval(nil)
	svc.q = store
	svc.events = store

	if _, err := svc.Approve(context.Background(), SaveCbtApprovalInput{
		EntityType:   "event",
		EntityID:     entityID,
		ApprovalType: "final_archive",
		ActorUserID:  actorID,
	}); err != nil {
		t.Fatalf("Approve(ready final archive) error = %v", err)
	}
	if store.createArg.EntityType != "event" || store.createArg.ApprovalType != "final_archive" {
		t.Fatalf("create arg = %+v, want delegated approval write", store.createArg)
	}
}
