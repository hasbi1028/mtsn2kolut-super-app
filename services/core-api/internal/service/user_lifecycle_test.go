package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeUserLifecycleStore struct {
	statusCalls []db.UpdateUserStatusParams
	statusErr   error
	revokeCalls []pgtype.UUID
	revokeErr   error
}

func userLifecycleTestUUID(seed byte) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte{seed, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, seed}, Valid: true}
}

func (f *fakeUserLifecycleStore) UpdateUserStatus(ctx context.Context, arg db.UpdateUserStatusParams) error {
	f.statusCalls = append(f.statusCalls, arg)
	return f.statusErr
}

func (f *fakeUserLifecycleStore) RevokeAllAuthSessionsForUser(ctx context.Context, userID pgtype.UUID) (int64, error) {
	f.revokeCalls = append(f.revokeCalls, userID)
	return 1, f.revokeErr
}

func TestUserLifecycleDeactivateInvalidatesSessions(t *testing.T) {
	userID := userLifecycleTestUUID(120)
	store := &fakeUserLifecycleStore{}
	svc := &UserLifecycle{q: store}

	if err := svc.UpdateStatus(context.Background(), userID, false); err != nil {
		t.Fatalf("UpdateStatus(false) error = %v", err)
	}
	if len(store.statusCalls) != 1 || store.statusCalls[0].ID != userID || store.statusCalls[0].IsActive {
		t.Fatalf("status calls = %+v, want inactive update", store.statusCalls)
	}
	if len(store.revokeCalls) != 1 || store.revokeCalls[0] != userID {
		t.Fatalf("revoke calls = %+v, want user sessions revoked", store.revokeCalls)
	}
}

func TestUserLifecycleDeleteIsNonDestructiveDeactivate(t *testing.T) {
	userID := userLifecycleTestUUID(121)
	store := &fakeUserLifecycleStore{}
	svc := &UserLifecycle{q: store}

	if err := svc.DeleteAsDeactivate(context.Background(), userID); err != nil {
		t.Fatalf("DeleteAsDeactivate() error = %v", err)
	}
	if len(store.statusCalls) != 1 || store.statusCalls[0].ID != userID || store.statusCalls[0].IsActive {
		t.Fatalf("status calls = %+v, want non-destructive inactive update", store.statusCalls)
	}
	if len(store.revokeCalls) != 1 || store.revokeCalls[0] != userID {
		t.Fatalf("revoke calls = %+v, want session invalidation", store.revokeCalls)
	}
}

func TestUserLifecycleReactivateDoesNotRevokeSessions(t *testing.T) {
	userID := userLifecycleTestUUID(122)
	store := &fakeUserLifecycleStore{}
	svc := &UserLifecycle{q: store}

	if err := svc.UpdateStatus(context.Background(), userID, true); err != nil {
		t.Fatalf("UpdateStatus(true) error = %v", err)
	}
	if len(store.statusCalls) != 1 || store.statusCalls[0].ID != userID || !store.statusCalls[0].IsActive {
		t.Fatalf("status calls = %+v, want active update", store.statusCalls)
	}
	if len(store.revokeCalls) != 0 {
		t.Fatalf("revoke calls = %+v, want no revoke on reactivation", store.revokeCalls)
	}
}

func TestUserLifecycleStopsBeforeRevokeWhenStatusFails(t *testing.T) {
	expected := errors.New("status failed")
	store := &fakeUserLifecycleStore{statusErr: expected}
	svc := &UserLifecycle{q: store}

	if err := svc.UpdateStatus(context.Background(), userLifecycleTestUUID(123), false); !errors.Is(err, expected) {
		t.Fatalf("UpdateStatus(error) = %v, want %v", err, expected)
	}
	if len(store.revokeCalls) != 0 {
		t.Fatalf("revoke calls = %+v, want no revoke after status failure", store.revokeCalls)
	}
}
