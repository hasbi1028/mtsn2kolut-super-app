package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type legacyAccountRoleExtraStore struct {
	calls int
	arg   db.AddUserRoleParams
	err   error
}

func (s *legacyAccountRoleExtraStore) AddUserRole(ctx context.Context, arg db.AddUserRoleParams) error {
	s.calls++
	s.arg = arg
	return s.err
}

func TestAddLegacyAccountRoleIfSafeSwallowsEnumCompatibilityErrors(t *testing.T) {
	userID := testGenerationUUID(81)
	cases := []struct {
		name string
		err  error
	}{
		{name: "constraint name", err: errors.New(`ERROR: insert or update on table "user_roles" violates foreign key constraint "user_role_role_check"`)},
		{name: "postgres enum", err: errors.New(`ERROR: invalid input value for enum user_role: "wali" (SQLSTATE 22P02)`)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := &legacyAccountRoleExtraStore{err: tc.err}
			if err := addLegacyAccountRoleIfSafe(context.Background(), store, userID, "wali"); err != nil {
				t.Fatalf("addLegacyAccountRoleIfSafe() error = %v, want nil compatibility swallow", err)
			}
			if store.calls != 1 {
				t.Fatalf("AddUserRole calls = %d, want 1", store.calls)
			}
			if store.arg.UserID != userID || store.arg.Role != db.UserRole("wali") {
				t.Fatalf("AddUserRole arg = %+v, want userID=%s role=wali", store.arg, userID.String())
			}
		})
	}
}

func TestAddLegacyAccountRoleIfSafeReturnsUnexpectedErrors(t *testing.T) {
	wantErr := errors.New("database connection lost")
	store := &legacyAccountRoleExtraStore{err: wantErr}
	if err := addLegacyAccountRoleIfSafe(context.Background(), store, pgtype.UUID{}, StudentAccountRole); !errors.Is(err, wantErr) {
		t.Fatalf("addLegacyAccountRoleIfSafe(unexpected) error = %v, want %v", err, wantErr)
	}
	if store.calls != 1 {
		t.Fatalf("AddUserRole calls = %d, want 1", store.calls)
	}
}

func TestAddLegacyAccountRoleIfSafeSuccess(t *testing.T) {
	userID := testGenerationUUID(82)
	store := &legacyAccountRoleExtraStore{}
	if err := addLegacyAccountRoleIfSafe(context.Background(), store, userID, ParentAccountRole); err != nil {
		t.Fatalf("addLegacyAccountRoleIfSafe(success) error = %v", err)
	}
	if store.calls != 1 || store.arg.UserID != userID || store.arg.Role != db.UserRoleOrtu {
		t.Fatalf("AddUserRole calls=%d arg=%+v, want one ortu role add", store.calls, store.arg)
	}
}
