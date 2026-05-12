package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func TestParentPortalDerivesParentIDFromUserAccountAndScopesChildren(t *testing.T) {
	userID := testGenerationUUID(41)
	parentID := testGenerationUUID(42)
	studentID := testGenerationUUID(43)
	store := &fakeParentPortalStore{parentID: parentID}
	svc := &ParentPortal{q: store}

	children, err := svc.Children(context.Background(), userID)
	if err != nil {
		t.Fatalf("Children() error = %v", err)
	}
	if store.parentIDUserID != userID || store.childrenParentID != parentID || len(children) != 1 {
		t.Fatalf("Children() user/parent/len = %s/%s/%d, want %s/%s/1", store.parentIDUserID.String(), store.childrenParentID.String(), len(children), userID.String(), parentID.String())
	}

	profile, err := svc.ChildProfile(context.Background(), userID, studentID)
	if err != nil {
		t.Fatalf("ChildProfile() error = %v", err)
	}
	if store.profileArg.ParentID != parentID || store.profileArg.StudentID != studentID || profile.ID != studentID {
		t.Fatalf("ChildProfile() arg/profile = %+v/%s, want parent %s student %s", store.profileArg, profile.ID.String(), parentID.String(), studentID.String())
	}

	if _, err := svc.ChildSchedule(context.Background(), userID, studentID); err != nil {
		t.Fatalf("ChildSchedule() error = %v", err)
	}
	if store.accessArg.ParentID != parentID || store.accessArg.StudentID != studentID || store.scheduleArg.ParentID != parentID || store.scheduleArg.StudentID != studentID {
		t.Fatalf("ChildSchedule() access/schedule arg = %+v/%+v, want parent %s student %s", store.accessArg, store.scheduleArg, parentID.String(), studentID.String())
	}

	if _, err := svc.ChildResults(context.Background(), userID, studentID); err != nil {
		t.Fatalf("ChildResults() error = %v", err)
	}
	if store.resultsArg.ParentID != parentID || store.resultsArg.StudentID != studentID {
		t.Fatalf("ChildResults() arg = %+v, want parent %s student %s", store.resultsArg, parentID.String(), studentID.String())
	}
}

func TestParentPortalRejectsUsersWithoutLinkedParent(t *testing.T) {
	svc := &ParentPortal{q: &fakeParentPortalStore{parentIDErr: pgx.ErrNoRows}}

	if _, err := svc.Children(context.Background(), testGenerationUUID(44)); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("Children() error = %v, want ErrForbidden", err)
	}
}

func TestParentPortalForbidsUnlinkedChildUUIDBeforeReturningEmptyData(t *testing.T) {
	userID := testGenerationUUID(45)
	parentID := testGenerationUUID(46)
	studentID := testGenerationUUID(47)
	store := &fakeParentPortalStore{parentID: parentID, accessErr: pgx.ErrNoRows}
	svc := &ParentPortal{q: store}

	if _, err := svc.ChildSchedule(context.Background(), userID, studentID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("ChildSchedule(unlinked) error = %v, want ErrForbidden", err)
	}
	if store.scheduleCalled {
		t.Fatal("ChildSchedule(unlinked) called scoped schedule query after failed access check")
	}

	if _, err := svc.ChildResults(context.Background(), userID, studentID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("ChildResults(unlinked) error = %v, want ErrForbidden", err)
	}
	if store.resultsCalled {
		t.Fatal("ChildResults(unlinked) called scoped results query after failed access check")
	}

	store = &fakeParentPortalStore{parentID: parentID, profileErr: pgx.ErrNoRows}
	svc = &ParentPortal{q: store}
	if _, err := svc.ChildProfile(context.Background(), userID, studentID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("ChildProfile(unlinked) error = %v, want ErrForbidden", err)
	}
}

type fakeParentPortalStore struct {
	parentID       pgtype.UUID
	parentIDErr    error
	parentIDUserID pgtype.UUID

	childrenParentID pgtype.UUID

	accessArg db.GetParentPortalChildAccessParams
	accessErr error

	profileArg db.GetParentPortalChildProfileParams
	profileErr error

	scheduleArg    db.ListParentPortalChildTimetableParams
	scheduleCalled bool

	resultsArg    db.ListParentPortalChildExamSessionsParams
	resultsCalled bool
}

func (f *fakeParentPortalStore) GetPortalParentIDByUserID(ctx context.Context, userID pgtype.UUID) (pgtype.UUID, error) {
	f.parentIDUserID = userID
	if f.parentIDErr != nil {
		return pgtype.UUID{}, f.parentIDErr
	}
	return f.parentID, nil
}

func (f *fakeParentPortalStore) ListParentPortalPreviewParents(ctx context.Context) ([]db.ListParentPortalPreviewParentsRow, error) {
	return []db.ListParentPortalPreviewParentsRow{{ID: testGenerationUUID(49), Nama: "Wali A", LinkedStudentCount: 1}}, nil
}

func (f *fakeParentPortalStore) ListParentChildren(ctx context.Context, parentID pgtype.UUID) ([]db.ListParentChildrenRow, error) {
	f.childrenParentID = parentID
	return []db.ListParentChildrenRow{{ID: testGenerationUUID(48), Nama: "Siswa A"}}, nil
}

func (f *fakeParentPortalStore) GetParentPortalChildAccess(ctx context.Context, arg db.GetParentPortalChildAccessParams) (pgtype.UUID, error) {
	f.accessArg = arg
	if f.accessErr != nil {
		return pgtype.UUID{}, f.accessErr
	}
	return arg.StudentID, nil
}

func (f *fakeParentPortalStore) GetParentPortalChildProfile(ctx context.Context, arg db.GetParentPortalChildProfileParams) (db.GetParentPortalChildProfileRow, error) {
	f.profileArg = arg
	if f.profileErr != nil {
		return db.GetParentPortalChildProfileRow{}, f.profileErr
	}
	return db.GetParentPortalChildProfileRow{ID: arg.StudentID, Nama: "Siswa A"}, nil
}

func (f *fakeParentPortalStore) ListParentPortalChildTimetable(ctx context.Context, arg db.ListParentPortalChildTimetableParams) ([]db.ListParentPortalChildTimetableRow, error) {
	f.scheduleCalled = true
	f.scheduleArg = arg
	return []db.ListParentPortalChildTimetableRow{{ID: testGenerationUUID(49), SubjectName: "IPA"}}, nil
}

func (f *fakeParentPortalStore) ListParentPortalChildExamSessions(ctx context.Context, arg db.ListParentPortalChildExamSessionsParams) ([]db.ListParentPortalChildExamSessionsRow, error) {
	f.resultsCalled = true
	f.resultsArg = arg
	return []db.ListParentPortalChildExamSessionsRow{{SessionID: testGenerationUUID(50), SessionTitle: "Ujian IPA"}}, nil
}
