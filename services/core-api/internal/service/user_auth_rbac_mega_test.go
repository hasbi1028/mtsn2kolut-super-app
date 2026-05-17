package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"mtsn2kolut-super-app/backend/internal/domain"
	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

func megaAuthRBACUUID(seed byte) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte{seed, seed, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, seed, seed}, Valid: true}
}

func TestMegaAuthChangePasswordLifecycleEdges(t *testing.T) {
	ctx := context.Background()
	userID := megaAuthRBACUUID(1)
	oldHash, err := bcrypt.GenerateFromPassword([]byte("oldSecret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword(old) error = %v", err)
	}
	store := newFakeStore()
	store.users["operator"] = db.User{ID: userID, Username: "operator", PasswordHash: string(oldHash), IsActive: true, MustChangePassword: true}
	svc := &Auth{q: store, jwtSecret: []byte("secret")}

	if err := svc.ChangePassword(ctx, "operator", "wrongSecret123", "newSecret123"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("ChangePassword(wrong old) = %v, want unauthorized", err)
	}
	if store.users["operator"].MustChangePassword != true {
		t.Fatalf("wrong old password changed must-change flag; user = %+v", store.users["operator"])
	}

	if err := svc.ChangePassword(ctx, "operator", "oldSecret123", "operator"); !errors.Is(err, domain.ErrWeakPassword) {
		t.Fatalf("ChangePassword(username as password) = %v, want weak password", err)
	}
	if store.users["operator"].MustChangePassword != true || store.users["operator"].AuthVersion != 0 {
		t.Fatalf("weak new password mutated user = %+v, want unchanged must-change/version", store.users["operator"])
	}

	sessionID := megaAuthRBACUUID(2)
	expires := pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true}
	store.authSessions[sessionID] = db.AuthSession{ID: sessionID, UserID: userID, ExpiresAt: expires, RefreshTokenHash: "legacy"}
	if err := svc.ChangePassword(ctx, "operator", "oldSecret123", "newSecret123"); err != nil {
		t.Fatalf("ChangePassword(valid) error = %v", err)
	}
	updated := store.users["operator"]
	if updated.MustChangePassword || !updated.PasswordChangedAt.Valid || updated.AuthVersion != 1 {
		t.Fatalf("updated user = %+v, want must-change cleared, timestamp set, version incremented", updated)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(updated.PasswordHash), []byte("newSecret123")); err != nil {
		t.Fatalf("stored password is not new bcrypt hash: %v", err)
	}
	if !store.authSessions[sessionID].RevokedAt.Valid {
		t.Fatalf("session after password change = %+v, want revoked", store.authSessions[sessionID])
	}
}

func TestMegaAuthSessionVersionAndAccessValidationEdges(t *testing.T) {
	ctx := context.Background()
	userID := megaAuthRBACUUID(3)
	otherUserID := megaAuthRBACUUID(4)
	sessionID := megaAuthRBACUUID(5)
	store := newFakeStore()
	store.users["student"] = db.User{ID: userID, Username: "student", IsActive: true, AuthVersion: 7}
	svc := &Auth{q: store, jwtSecret: []byte("secret")}

	if ok := svc.validAuthVersion(ctx, mapClaims("", int64(7))); ok {
		t.Fatal("validAuthVersion(blank subject) = true, want false")
	}
	if ok := svc.validAuthVersion(ctx, mapClaims(userID.String(), "not-a-number")); ok {
		t.Fatal("validAuthVersion(non-numeric version) = true, want false")
	}
	if ok := svc.validAuthVersion(ctx, mapClaims(userID.String(), "7")); !ok {
		t.Fatal("validAuthVersion(string version matching current) = false, want true")
	}

	expires := pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true}
	store.authSessions[sessionID] = db.AuthSession{ID: sessionID, UserID: otherUserID, ExpiresAt: expires}
	ok, err := svc.ValidateAccessSession(ctx, userID.String(), sessionID.String())
	if err != nil || ok {
		t.Fatalf("ValidateAccessSession(wrong session owner) = %v/%v, want false nil", ok, err)
	}

	store.authSessions[sessionID] = db.AuthSession{ID: sessionID, UserID: userID, ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(-time.Minute), Valid: true}}
	ok, err = svc.ValidateAccessSession(ctx, userID.String(), sessionID.String())
	if err != nil || ok {
		t.Fatalf("ValidateAccessSession(expired session) = %v/%v, want false nil", ok, err)
	}

	store.authSessions[sessionID] = db.AuthSession{ID: sessionID, UserID: userID, ExpiresAt: expires}
	ok, err = svc.ValidateAccessSession(ctx, userID.String(), sessionID.String())
	if err != nil || !ok {
		t.Fatalf("ValidateAccessSession(valid session) = %v/%v, want true nil", ok, err)
	}
	if !store.authSessions[sessionID].LastUsedAt.Valid {
		t.Fatalf("valid access session = %+v, want last-used touch", store.authSessions[sessionID])
	}
}

func mapClaims(subject string, version any) map[string]any {
	return map[string]any{"sub": subject, "ver": version}
}

func TestMegaUserLifecycleGuardAndAuditEdges(t *testing.T) {
	ctx := context.Background()
	userID := userLifecycleTestUUID(201)
	actorID := userLifecycleTestUUID(202)
	expected := errors.New("rbac lookup failed")
	store := &fakeUserLifecycleStore{roleErr: expected}
	svc := &UserLifecycle{q: store}

	if err := svc.UpdateStatus(ctx, userID, false, actorID); !errors.Is(err, expected) {
		t.Fatalf("UpdateStatus(deactivate role error) = %v, want %v", err, expected)
	}
	if len(store.statusCalls) != 0 || len(store.revokeCalls) != 0 || len(store.auditCalls) != 0 {
		t.Fatalf("guard error calls = status %v revoke %v audit %v, want none", store.statusCalls, store.revokeCalls, store.auditCalls)
	}

	store = &fakeUserLifecycleStore{userByID: db.GetUserByIDRow{ID: userID, Username: "operator"}, auditErr: errors.New("audit down")}
	svc = &UserLifecycle{q: store}
	if err := svc.ForcePasswordChange(ctx, userID, actorID); err == nil || !strings.Contains(err.Error(), "audit down") {
		t.Fatalf("ForcePasswordChange(audit failure) = %v, want audit error", err)
	}
	if len(store.mustChangePasswordIDs) != 1 || len(store.versionCalls) != 1 || len(store.revokeCalls) != 1 || len(store.auditCalls) != 1 {
		t.Fatalf("force-password calls = mustChange %v version %v revoke %v audit %v, want all pre-audit mutations attempted", store.mustChangePasswordIDs, store.versionCalls, store.revokeCalls, store.auditCalls)
	}
}

func TestMegaRBACSelfAdminRemovalAndInactiveRoleEdges(t *testing.T) {
	ctx := context.Background()
	adminID := rbacTestUUID(201)
	store := &fakeRBACStore{
		roles:        []db.RbacRole{{Code: "admin", IsActive: true}, {Code: "guru", IsActive: true}, {Code: "disabled", IsActive: false}},
		activeAdmins: 2,
		userHasRole:  true,
	}
	svc := &RBAC{q: store}

	if err := svc.ReplaceUserRoles(ctx, adminID, []string{"guru"}, adminID); err == nil || !strings.Contains(err.Error(), "diri sendiri") {
		t.Fatalf("ReplaceUserRoles(self admin removal) = %v, want self-removal guard", err)
	}
	if len(store.deletedUserRoles) != 0 || len(store.addedUserRoles) != 0 || len(store.versionedUsers) != 0 {
		t.Fatalf("self-removal calls = deleted %v added %v versioned %v, want none", store.deletedUserRoles, store.addedUserRoles, store.versionedUsers)
	}

	if err := svc.ReplaceUserRoles(ctx, adminID, []string{"disabled"}, rbacTestUUID(202)); err == nil || !strings.Contains(err.Error(), "tidak dikenal atau tidak aktif") {
		t.Fatalf("ReplaceUserRoles(inactive role) = %v, want inactive-role validation", err)
	}
	if len(store.deletedUserRoles) != 0 || len(store.addedUserRoles) != 0 || len(store.versionedUsers) != 0 {
		t.Fatalf("inactive-role calls = deleted %v added %v versioned %v, want none", store.deletedUserRoles, store.addedUserRoles, store.versionedUsers)
	}
}

func TestMegaRBACPermissionMutationStopsBeforeAuditWhenInvalidationFails(t *testing.T) {
	ctx := context.Background()
	actorID := rbacTestUUID(203)
	userID := rbacTestUUID(204)
	expected := errors.New("auth version failed")
	store := &fakeRBACStore{
		permissions:         []db.RbacPermission{{Code: "reports.view", IsActive: true}},
		usersByPermission:   map[string][]pgtype.UUID{"reports.view": {userID}},
		incrementVersionErr: expected,
	}
	svc := &RBAC{q: store}

	err := svc.SetPermissionActive(ctx, " reports.view ", false, actorID)
	if !errors.Is(err, expected) {
		t.Fatalf("SetPermissionActive(invalidation failure) = %v, want %v", err, expected)
	}
	if len(store.setPermissionActives) != 1 || store.setPermissionActives[0].Code != "reports.view" || store.setPermissionActives[0].IsActive {
		t.Fatalf("set permission calls = %+v, want normalized inactive update before invalidation", store.setPermissionActives)
	}
	if len(store.revokedUsers) != 0 || len(store.auditActions) != 0 {
		t.Fatalf("post-invalidation-failure calls = revoked %v audit %v, want none", store.revokedUsers, store.auditActions)
	}
}

func TestMegaEmployeeAccountGenerationStopsOnRoleAssignmentFailure(t *testing.T) {
	ctx := context.Background()
	expected := errors.New("legacy role insert failed")
	store := &fakeEmployeeAccountGenerationStore{
		roleErr: expected,
		rows: []db.ListEmployeeAccountGenerationCandidatesRow{{
			EmployeeID: testGenerationUUID(201), Nip: "1", Nama: "Guru Role Error", TanggalLahir: validDate(1990, 1, 1), NomorUrut: 1, GeneratedUsername: "4040603190001",
		}},
	}
	generator := &EmployeeAccountGenerator{q: store, npsn: DefaultEmployeeAccountNPSN, role: DefaultEmployeeAccountRole}

	result, err := generator.Generate(ctx, testGenerationUUID(202))
	if !errors.Is(err, expected) {
		t.Fatalf("Generate(role failure) error = %v, want %v", err, expected)
	}
	if result.Created != 0 || result.Ready != 1 || result.Items[0].Status != "ready" {
		t.Fatalf("Generate(role failure) result = %+v, want fatal error before item marked created", result)
	}
	if len(store.createCalls) != 1 || len(store.roleCalls) != 1 || len(store.rbacRoleCalls) != 0 || len(store.auditCalls) != 0 {
		t.Fatalf("generation calls = create %v legacy-role %v rbac-role %v audit %v, want stop at legacy role failure", store.createCalls, store.roleCalls, store.rbacRoleCalls, store.auditCalls)
	}
}

func TestMegaAuthCurrentVersionMapsMissingAndSuspendedUsers(t *testing.T) {
	ctx := context.Background()
	store := newFakeStore()
	svc := &Auth{q: store, jwtSecret: []byte("secret")}

	if _, err := svc.CurrentAuthVersion(ctx, "not-a-uuid"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("CurrentAuthVersion(malformed) = %v, want unauthorized", err)
	}
	if _, err := svc.CurrentAuthVersion(ctx, megaAuthRBACUUID(9).String()); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("CurrentAuthVersion(missing) = %v, want unauthorized", err)
	}
	inactiveID := megaAuthRBACUUID(10)
	store.users["inactive"] = db.User{ID: inactiveID, Username: "inactive", IsActive: false, AuthVersion: 3}
	if _, err := svc.CurrentAuthVersion(ctx, inactiveID.String()); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("CurrentAuthVersion(inactive) = %v, want unauthorized", err)
	}

	store.getUserByIDErr = pgx.ErrTxClosed
	if _, err := svc.CurrentAuthVersion(ctx, inactiveID.String()); !errors.Is(err, pgx.ErrTxClosed) {
		t.Fatalf("CurrentAuthVersion(store error) = %v, want raw store error", err)
	}
}
