package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeUserLifecycleStore struct {
	statusCalls     []db.UpdateUserStatusParams
	softDeleteCalls []pgtype.UUID
	statusErr       error
	revokeCalls     []pgtype.UUID
	revokeErr       error

	userByID db.GetUserByIDRow
	userErr  error

	userHasAdmin bool
	adminCount   int64
	roleErr      error

	passwordCalls []db.UpdateUserPasswordParams
	passwordErr   error
	versionCalls  []pgtype.UUID
	versionErr    error

	profileCalls []db.UpdateUserProfileLinkParams
	profileErr   error
	auditCalls   []db.CreateAuditLogParams
	auditErr     error
}

func userLifecycleTestUUID(seed byte) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte{seed, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, seed}, Valid: true}
}

func (f *fakeUserLifecycleStore) UpdateUserStatus(ctx context.Context, arg db.UpdateUserStatusParams) error {
	f.statusCalls = append(f.statusCalls, arg)
	return f.statusErr
}

func (f *fakeUserLifecycleStore) SoftDeleteUser(ctx context.Context, id pgtype.UUID) error {
	f.softDeleteCalls = append(f.softDeleteCalls, id)
	return f.statusErr
}

func (f *fakeUserLifecycleStore) RevokeAllAuthSessionsForUser(ctx context.Context, userID pgtype.UUID) (int64, error) {
	f.revokeCalls = append(f.revokeCalls, userID)
	return 1, f.revokeErr
}

func (f *fakeUserLifecycleStore) GetUserByID(ctx context.Context, id pgtype.UUID) (db.GetUserByIDRow, error) {
	return f.userByID, f.userErr
}

func (f *fakeUserLifecycleStore) UserHasRbacRole(ctx context.Context, arg db.UserHasRbacRoleParams) (bool, error) {
	return f.userHasAdmin, f.roleErr
}

func (f *fakeUserLifecycleStore) CountActiveAdminsByRbac(ctx context.Context) (int64, error) {
	return f.adminCount, f.roleErr
}

func (f *fakeUserLifecycleStore) UpdateUserPassword(ctx context.Context, arg db.UpdateUserPasswordParams) error {
	f.passwordCalls = append(f.passwordCalls, arg)
	return f.passwordErr
}

func (f *fakeUserLifecycleStore) IncrementUserAuthVersion(ctx context.Context, userID pgtype.UUID) (int32, error) {
	f.versionCalls = append(f.versionCalls, userID)
	return 2, f.versionErr
}

func (f *fakeUserLifecycleStore) UpdateUserProfileLink(ctx context.Context, arg db.UpdateUserProfileLinkParams) error {
	f.profileCalls = append(f.profileCalls, arg)
	return f.profileErr
}

func (f *fakeUserLifecycleStore) CreateAuditLog(ctx context.Context, arg db.CreateAuditLogParams) (db.AuditLog, error) {
	f.auditCalls = append(f.auditCalls, arg)
	return db.AuditLog{}, f.auditErr
}

func TestUserLifecycleDeactivateInvalidatesSessions(t *testing.T) {
	userID := userLifecycleTestUUID(120)
	actorID := userLifecycleTestUUID(125)
	store := &fakeUserLifecycleStore{adminCount: 2}
	svc := &UserLifecycle{q: store}

	if err := svc.UpdateStatus(context.Background(), userID, false, actorID); err != nil {
		t.Fatalf("UpdateStatus(false) error = %v", err)
	}
	if len(store.statusCalls) != 1 || store.statusCalls[0].ID != userID || store.statusCalls[0].IsActive {
		t.Fatalf("status calls = %+v, want inactive update", store.statusCalls)
	}
	if len(store.revokeCalls) != 1 || store.revokeCalls[0] != userID {
		t.Fatalf("revoke calls = %+v, want user sessions revoked", store.revokeCalls)
	}
	if len(store.auditCalls) != 1 || store.auditCalls[0].Action != "USER_STATUS_UPDATED" {
		t.Fatalf("audit calls = %+v, want status audit", store.auditCalls)
	}
}

func TestUserLifecycleRefusesToDeactivateLastActiveAdmin(t *testing.T) {
	userID := userLifecycleTestUUID(126)
	store := &fakeUserLifecycleStore{userHasAdmin: true, adminCount: 1}
	svc := &UserLifecycle{q: store}

	if err := svc.UpdateStatus(context.Background(), userID, false, userLifecycleTestUUID(127)); err == nil {
		t.Fatal("UpdateStatus(last admin inactive) error = nil, want error")
	}
	if len(store.statusCalls) != 0 || len(store.revokeCalls) != 0 {
		t.Fatalf("calls = status %+v revoke %+v, want no mutation", store.statusCalls, store.revokeCalls)
	}
}

func TestUserLifecycleDeleteSoftDeletesAndRevokesSessions(t *testing.T) {
	userID := userLifecycleTestUUID(121)
	store := &fakeUserLifecycleStore{adminCount: 2}
	svc := &UserLifecycle{q: store}

	if err := svc.DeleteAsDeactivate(context.Background(), userID, userLifecycleTestUUID(128)); err != nil {
		t.Fatalf("DeleteAsDeactivate() error = %v", err)
	}
	if len(store.softDeleteCalls) != 1 || store.softDeleteCalls[0] != userID {
		t.Fatalf("soft delete calls = %+v, want one target soft-delete", store.softDeleteCalls)
	}
	if len(store.statusCalls) != 0 {
		t.Fatalf("status calls = %+v, want delete to use deleted_at soft delete", store.statusCalls)
	}
	if len(store.revokeCalls) != 1 || store.revokeCalls[0] != userID {
		t.Fatalf("revoke calls = %+v, want session invalidation", store.revokeCalls)
	}
	if len(store.auditCalls) != 1 || store.auditCalls[0].Action != "USER_DELETED" {
		t.Fatalf("audit calls = %+v, want user deleted audit", store.auditCalls)
	}
}

func TestUserLifecycleReactivateDoesNotRevokeSessions(t *testing.T) {
	userID := userLifecycleTestUUID(122)
	store := &fakeUserLifecycleStore{}
	svc := &UserLifecycle{q: store}

	if err := svc.UpdateStatus(context.Background(), userID, true, userLifecycleTestUUID(129)); err != nil {
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

	if err := svc.UpdateStatus(context.Background(), userLifecycleTestUUID(123), false, userLifecycleTestUUID(130)); !errors.Is(err, expected) {
		t.Fatalf("UpdateStatus(error) = %v, want %v", err, expected)
	}
	if len(store.revokeCalls) != 0 {
		t.Fatalf("revoke calls = %+v, want no revoke after status failure", store.revokeCalls)
	}
}

func TestUserLifecycleResetPasswordHashesPasswordRevokesSessionsAndAudits(t *testing.T) {
	userID := userLifecycleTestUUID(131)
	actorID := userLifecycleTestUUID(132)
	store := &fakeUserLifecycleStore{userByID: db.GetUserByIDRow{ID: userID, Username: "operator"}}
	svc := &UserLifecycle{q: store}

	if err := svc.ResetPassword(context.Background(), userID, "newSecret123", actorID); err != nil {
		t.Fatalf("ResetPassword() error = %v", err)
	}
	if len(store.passwordCalls) != 1 || store.passwordCalls[0].ID != userID {
		t.Fatalf("password calls = %+v, want one target update", store.passwordCalls)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(store.passwordCalls[0].PasswordHash), []byte("newSecret123")); err != nil {
		t.Fatalf("stored password is not bcrypt hash of new password: %v", err)
	}
	if len(store.versionCalls) != 1 || store.versionCalls[0] != userID || len(store.revokeCalls) != 1 || store.revokeCalls[0] != userID {
		t.Fatalf("version/revoke calls = %+v/%+v", store.versionCalls, store.revokeCalls)
	}
	if len(store.auditCalls) != 1 || store.auditCalls[0].Action != "USER_PASSWORD_RESET" || store.auditCalls[0].UserID != actorID {
		t.Fatalf("audit calls = %+v, want password reset audit by actor", store.auditCalls)
	}
}

func TestUserLifecycleUpdateProfileLinkValidatesSingleProfileAndAudits(t *testing.T) {
	userID := userLifecycleTestUUID(133)
	actorID := userLifecycleTestUUID(134)
	employeeID := userLifecycleTestUUID(135)
	store := &fakeUserLifecycleStore{}
	svc := &UserLifecycle{q: store}

	if err := svc.UpdateProfileLink(context.Background(), userID, ProfileLink{EmployeeID: employeeID}, actorID); err != nil {
		t.Fatalf("UpdateProfileLink() error = %v", err)
	}
	if len(store.profileCalls) != 1 || store.profileCalls[0].ID != userID || store.profileCalls[0].EmployeeID != employeeID {
		t.Fatalf("profile calls = %+v, want employee link", store.profileCalls)
	}
	if len(store.auditCalls) != 1 || store.auditCalls[0].Action != "USER_PROFILE_LINK_UPDATED" {
		t.Fatalf("audit calls = %+v, want profile link audit", store.auditCalls)
	}
	var metadata map[string]any
	if err := json.Unmarshal(store.auditCalls[0].Metadata, &metadata); err != nil || metadata["employee_id"] == "" {
		t.Fatalf("audit metadata = %s err=%v, want employee_id", string(store.auditCalls[0].Metadata), err)
	}
}

func TestUserLifecycleUpdateProfileLinkRejectsMultipleProfiles(t *testing.T) {
	svc := &UserLifecycle{q: &fakeUserLifecycleStore{}}
	err := svc.UpdateProfileLink(context.Background(), userLifecycleTestUUID(136), ProfileLink{EmployeeID: userLifecycleTestUUID(137), StudentID: userLifecycleTestUUID(138)}, userLifecycleTestUUID(139))
	if err == nil {
		t.Fatal("UpdateProfileLink(multiple profiles) error = nil, want error")
	}
}
