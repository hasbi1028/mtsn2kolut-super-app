package service

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeBankSoalReviewerScopeListStore struct {
	rows []db.ListBankSoalReviewerScopesRow
}

func (f *fakeBankSoalReviewerScopeListStore) ListBankSoalReviewerScopes(context.Context) ([]db.ListBankSoalReviewerScopesRow, error) {
	return f.rows, nil
}

func (f *fakeBankSoalReviewerScopeListStore) UpsertBankSoalReviewerScope(context.Context, db.UpsertBankSoalReviewerScopeParams) (db.UpsertBankSoalReviewerScopeRow, error) {
	return db.UpsertBankSoalReviewerScopeRow{}, nil
}

func (f *fakeBankSoalReviewerScopeListStore) DeleteBankSoalReviewerScope(context.Context, pgtype.UUID) (pgtype.UUID, error) {
	return pgtype.UUID{}, nil
}

func (f *fakeBankSoalReviewerScopeListStore) CanBankSoalUserReview(context.Context, db.CanBankSoalUserReviewParams) (bool, error) {
	return false, nil
}

func (f *fakeBankSoalReviewerScopeListStore) CanBankSoalUserApprove(context.Context, db.CanBankSoalUserApproveParams) (bool, error) {
	return false, nil
}

func TestBankSoalReviewerScopeListForwardsStore(t *testing.T) {
	store := &fakeBankSoalReviewerScopeListStore{rows: []db.ListBankSoalReviewerScopesRow{{}}}
	svc := &BankSoalReviewerScope{q: store}

	rows, err := svc.List(context.Background())
	if err != nil || len(rows) != 1 {
		t.Fatalf("List() = %d rows, %v; want 1 nil", len(rows), err)
	}
}

type fakeParentPortalListStore struct {
	previewRows []db.ListParentPortalPreviewParentsRow
}

func (f *fakeParentPortalListStore) GetPortalParentIDByUserID(context.Context, pgtype.UUID) (pgtype.UUID, error) {
	return pgtype.UUID{}, nil
}

func (f *fakeParentPortalListStore) ListParentPortalPreviewParents(context.Context) ([]db.ListParentPortalPreviewParentsRow, error) {
	return f.previewRows, nil
}

func (f *fakeParentPortalListStore) ListParentChildren(context.Context, pgtype.UUID) ([]db.ListParentChildrenRow, error) {
	return nil, nil
}

func (f *fakeParentPortalListStore) GetParentPortalChildAccess(context.Context, db.GetParentPortalChildAccessParams) (pgtype.UUID, error) {
	return pgtype.UUID{}, nil
}

func (f *fakeParentPortalListStore) GetParentPortalChildProfile(context.Context, db.GetParentPortalChildProfileParams) (db.GetParentPortalChildProfileRow, error) {
	return db.GetParentPortalChildProfileRow{}, nil
}

func (f *fakeParentPortalListStore) ListParentPortalChildTimetable(context.Context, db.ListParentPortalChildTimetableParams) ([]db.ListParentPortalChildTimetableRow, error) {
	return nil, nil
}

func (f *fakeParentPortalListStore) ListParentPortalChildExamSessions(context.Context, db.ListParentPortalChildExamSessionsParams) ([]db.ListParentPortalChildExamSessionsRow, error) {
	return nil, nil
}

func TestParentPortalPreviewListForwardsStore(t *testing.T) {
	store := &fakeParentPortalListStore{previewRows: []db.ListParentPortalPreviewParentsRow{{}}}
	svc := &ParentPortal{q: store}

	rows, err := svc.ListParentPortalPreviewParents(context.Background())
	if err != nil || len(rows) != 1 {
		t.Fatalf("ListParentPortalPreviewParents() = %d rows, %v; want 1 nil", len(rows), err)
	}
}
