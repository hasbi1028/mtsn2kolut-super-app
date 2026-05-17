package db

import (
	"context"
	"crypto/rand"
	"slices"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestIntegrationUsersAuthRbacRepository(t *testing.T) {
	t.Parallel()

	tdb := setupIntegrationTestDB(t)
	ctx := context.Background()
	q := tdb.Q
	suffix := integrationSuffix()

	user, err := q.CreateUserWithMustChangePassword(ctx, CreateUserWithMustChangePasswordParams{
		Username:           "it-user-" + suffix,
		PasswordHash:       "hash:v1:" + suffix,
		DisplayName:        pgtype.Text{String: "Integration User " + suffix, Valid: true},
		IsActive:           true,
		MustChangePassword: true,
	})
	if err != nil {
		t.Fatalf("create user with must-change password: %v", err)
	}
	if !user.ID.Valid {
		t.Fatalf("created user has invalid id")
	}
	if !user.MustChangePassword || user.PasswordChangedAt.Valid {
		t.Fatalf("created user password flags = must_change:%v changed_at:%v, want must_change true and null changed_at", user.MustChangePassword, user.PasswordChangedAt.Valid)
	}

	gotByUsername, err := q.GetUserByUsername(ctx, user.Username)
	if err != nil {
		t.Fatalf("get user by username: %v", err)
	}
	if gotByUsername.ID != user.ID || gotByUsername.PasswordHash != "hash:v1:"+suffix || !gotByUsername.IsActive {
		t.Fatalf("got user by username = %#v, want id %v active with original hash", gotByUsername, user.ID)
	}

	if err := q.AddUserRole(ctx, AddUserRoleParams{UserID: user.ID, Role: UserRoleGuru}); err != nil {
		t.Fatalf("add legacy user role: %v", err)
	}
	legacyRoles, err := q.GetUserRoles(ctx, user.ID)
	if err != nil {
		t.Fatalf("get legacy user roles: %v", err)
	}
	if !slices.Contains(legacyRoles, UserRoleGuru) {
		t.Fatalf("legacy roles = %v, want guru", legacyRoles)
	}

	if err := q.RemoveUserRole(ctx, RemoveUserRoleParams{UserID: user.ID, Role: UserRoleGuru}); err != nil {
		t.Fatalf("remove legacy user role: %v", err)
	}
	legacyRoles, err = q.GetUserRoles(ctx, user.ID)
	if err != nil {
		t.Fatalf("get legacy user roles after remove: %v", err)
	}
	if slices.Contains(legacyRoles, UserRoleGuru) {
		t.Fatalf("legacy roles after remove = %v, did not want guru", legacyRoles)
	}

	initialAdminCount, err := q.CountActiveAdminsByRbac(ctx)
	if err != nil {
		t.Fatalf("count initial active admins by rbac: %v", err)
	}
	adminRole, err := q.GetRbacRoleByCode(ctx, "admin")
	if err != nil {
		t.Fatalf("get admin rbac role: %v", err)
	}
	if !adminRole.IsSystem || !adminRole.IsActive {
		t.Fatalf("admin role = %#v, want active system role", adminRole)
	}
	usersReadPermission, err := q.GetRbacPermissionByCode(ctx, "users.read")
	if err != nil {
		t.Fatalf("get users.read permission: %v", err)
	}
	if usersReadPermission.Module != "users" || usersReadPermission.Action != "read" || !usersReadPermission.IsActive {
		t.Fatalf("users.read permission = %#v", usersReadPermission)
	}

	if err := q.AddUserRbacRoleByCode(ctx, AddUserRbacRoleByCodeParams{UserID: user.ID, Code: "admin"}); err != nil {
		t.Fatalf("add admin rbac role to user: %v", err)
	}
	if err := q.SyncLegacyUserRolesFromRbac(ctx, user.ID); err != nil {
		t.Fatalf("sync legacy user roles from rbac: %v", err)
	}
	legacyRoles, err = q.GetUserRoles(ctx, user.ID)
	if err != nil {
		t.Fatalf("get legacy user roles after rbac sync: %v", err)
	}
	if !slices.Contains(legacyRoles, UserRoleAdmin) {
		t.Fatalf("legacy roles after rbac sync = %v, want admin", legacyRoles)
	}

	rbacRoles, err := q.GetUserRoleCodesFromRbac(ctx, user.ID)
	if err != nil {
		t.Fatalf("get user role codes from rbac: %v", err)
	}
	if !slices.Contains(rbacRoles, "admin") {
		t.Fatalf("rbac role codes = %v, want admin", rbacRoles)
	}
	permissions, err := q.GetUserPermissionCodes(ctx, user.ID)
	if err != nil {
		t.Fatalf("get user permission codes: %v", err)
	}
	if !slices.Contains(permissions, "users.read") || !slices.Contains(permissions, "roles.read") {
		t.Fatalf("admin permissions = %v, want users.read and roles.read", permissions)
	}
	adminCount, err := q.CountActiveAdminsByRbac(ctx)
	if err != nil {
		t.Fatalf("count active admins by rbac: %v", err)
	}
	if adminCount != initialAdminCount+1 {
		t.Fatalf("active admin count = %d, want %d", adminCount, initialAdminCount+1)
	}
	adminIDs, err := q.ListUserIDsByRoleCode(ctx, "admin")
	if err != nil {
		t.Fatalf("list user ids by admin role: %v", err)
	}
	if !uuidSliceContains(adminIDs, user.ID) {
		t.Fatalf("admin user ids = %v, want created user %v", adminIDs, user.ID)
	}
	usersReadIDs, err := q.ListActiveUserIDsByPermissionCode(ctx, "users.read")
	if err != nil {
		t.Fatalf("list active user ids by users.read permission: %v", err)
	}
	if !uuidSliceContains(usersReadIDs, user.ID) {
		t.Fatalf("users.read user ids = %v, want created user %v", usersReadIDs, user.ID)
	}
	hasAdminRole, err := q.UserHasRbacRole(ctx, UserHasRbacRoleParams{UserID: user.ID, Code: "admin"})
	if err != nil {
		t.Fatalf("user has admin rbac role: %v", err)
	}
	if !hasAdminRole {
		t.Fatalf("user has admin rbac role = false, want true")
	}

	customRoleCode := "it_role_" + suffix
	customPermissionCode := "it_module_" + suffix + ".read"
	customRole, err := q.CreateRbacRole(ctx, CreateRbacRoleParams{Code: customRoleCode, Name: "Integration Role", Description: "created by integration test"})
	if err != nil {
		t.Fatalf("create custom rbac role: %v", err)
	}
	if customRole.IsSystem || !customRole.IsActive {
		t.Fatalf("custom role = %#v, want active non-system role", customRole)
	}
	customPermission, err := q.CreateRbacPermission(ctx, CreateRbacPermissionParams{Code: customPermissionCode, Module: "it_module_" + suffix, Action: "read", Description: "created by integration test"})
	if err != nil {
		t.Fatalf("create custom rbac permission: %v", err)
	}
	if !customPermission.IsActive {
		t.Fatalf("custom permission inactive: %#v", customPermission)
	}
	updatedRole, err := q.UpdateRbacRole(ctx, UpdateRbacRoleParams{Code: customRoleCode, Name: "Integration Role Updated", Description: "updated by integration test"})
	if err != nil {
		t.Fatalf("update custom rbac role: %v", err)
	}
	if updatedRole.Name != "Integration Role Updated" {
		t.Fatalf("updated role name = %q", updatedRole.Name)
	}
	updatedPermission, err := q.UpdateRbacPermission(ctx, UpdateRbacPermissionParams{Code: customPermissionCode, Module: "it_module_" + suffix, Action: "manage", Description: "updated by integration test"})
	if err != nil {
		t.Fatalf("update custom rbac permission: %v", err)
	}
	if updatedPermission.Action != "manage" {
		t.Fatalf("updated permission action = %q", updatedPermission.Action)
	}
	if err := q.AddRolePermissionByCode(ctx, AddRolePermissionByCodeParams{Code: customRoleCode, Code_2: customPermissionCode}); err != nil {
		t.Fatalf("add custom permission to custom role: %v", err)
	}
	if err := q.AddUserRbacRoleByCode(ctx, AddUserRbacRoleByCodeParams{UserID: user.ID, Code: customRoleCode}); err != nil {
		t.Fatalf("add custom rbac role to user: %v", err)
	}
	permissions, err = q.GetUserPermissionCodes(ctx, user.ID)
	if err != nil {
		t.Fatalf("get user permission codes after custom role: %v", err)
	}
	if !slices.Contains(permissions, customPermissionCode) {
		t.Fatalf("permissions after custom role = %v, want %q", permissions, customPermissionCode)
	}
	if _, err := q.SetRbacPermissionActive(ctx, SetRbacPermissionActiveParams{Code: customPermissionCode, IsActive: false}); err != nil {
		t.Fatalf("deactivate custom permission: %v", err)
	}
	permissions, err = q.GetUserPermissionCodes(ctx, user.ID)
	if err != nil {
		t.Fatalf("get user permission codes after permission deactivate: %v", err)
	}
	if slices.Contains(permissions, customPermissionCode) {
		t.Fatalf("permissions after permission deactivate = %v, did not want %q", permissions, customPermissionCode)
	}
	if err := q.DeleteRolePermissions(ctx, customRoleCode); err != nil {
		t.Fatalf("delete custom role permissions: %v", err)
	}
	if _, err := q.SetRbacRoleActive(ctx, SetRbacRoleActiveParams{Code: customRoleCode, IsActive: false}); err != nil {
		t.Fatalf("deactivate custom role: %v", err)
	}
	hasCustomRole, err := q.UserHasRbacRole(ctx, UserHasRbacRoleParams{UserID: user.ID, Code: customRoleCode})
	if err != nil {
		t.Fatalf("user has inactive custom rbac role: %v", err)
	}
	if hasCustomRole {
		t.Fatalf("user has inactive custom rbac role = true, want false")
	}
	anyStatusIDs, err := q.ListActiveUserIDsByRoleCodeAnyStatus(ctx, customRoleCode)
	if err != nil {
		t.Fatalf("list active user ids by role code any status: %v", err)
	}
	if !uuidSliceContains(anyStatusIDs, user.ID) {
		t.Fatalf("any-status custom role ids = %v, want created user %v", anyStatusIDs, user.ID)
	}

	sessionID := randomTestUUID(t)
	session, err := q.CreateAuthSession(ctx, CreateAuthSessionParams{
		ID:               sessionID,
		UserID:           user.ID,
		RefreshTokenHash: "refresh:v1:" + suffix,
		ExpiresAt:        pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
		IpAddress:        "127.0.0.1",
		UserAgent:        "integration-test",
		DeviceLabel:      "Original device",
	})
	if err != nil {
		t.Fatalf("create auth session: %v", err)
	}
	if session.ID != sessionID || session.UserID != user.ID || session.RevokedAt.Valid {
		t.Fatalf("created auth session = %#v", session)
	}
	activeSessions, err := q.ListActiveAuthSessionsByUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("list active auth sessions by user: %v", err)
	}
	if !authSessionListed(activeSessions, sessionID) {
		t.Fatalf("active sessions = %v, want session %v", activeSessions, sessionID)
	}
	updatedRows, err := q.UpdateOwnedAuthSessionLabel(ctx, UpdateOwnedAuthSessionLabelParams{UserID: user.ID, ID: sessionID, DeviceLabel: "Renamed device"})
	if err != nil {
		t.Fatalf("update owned auth session label: %v", err)
	}
	if updatedRows != 1 {
		t.Fatalf("updated auth session labels = %d, want 1", updatedRows)
	}
	gotSession, err := q.GetAuthSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("get auth session: %v", err)
	}
	if gotSession.DeviceLabel != "Renamed device" {
		t.Fatalf("auth session device label = %q, want Renamed device", gotSession.DeviceLabel)
	}
	wrongHashRows, err := q.RevokeLiveAuthSessionByHash(ctx, RevokeLiveAuthSessionByHashParams{ID: sessionID, RefreshTokenHash: "wrong-hash"})
	if err != nil {
		t.Fatalf("revoke live auth session by wrong hash: %v", err)
	}
	if wrongHashRows != 0 {
		t.Fatalf("wrong-hash revoke rows = %d, want 0", wrongHashRows)
	}
	revokedRows, err := q.RevokeLiveAuthSessionByHash(ctx, RevokeLiveAuthSessionByHashParams{ID: sessionID, RefreshTokenHash: "refresh:v1:" + suffix})
	if err != nil {
		t.Fatalf("revoke live auth session by hash: %v", err)
	}
	if revokedRows != 1 {
		t.Fatalf("revoke live auth session rows = %d, want 1", revokedRows)
	}
	gotSession, err = q.GetAuthSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("get revoked auth session: %v", err)
	}
	if !gotSession.RevokedAt.Valid {
		t.Fatalf("revoked auth session has null revoked_at")
	}

	passwordSessionID := randomTestUUID(t)
	if _, err := q.CreateAuthSession(ctx, CreateAuthSessionParams{
		ID:               passwordSessionID,
		UserID:           user.ID,
		RefreshTokenHash: "refresh:v2:" + suffix,
		ExpiresAt:        pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
		IpAddress:        "127.0.0.1",
		UserAgent:        "integration-test",
		DeviceLabel:      "Password session",
	}); err != nil {
		t.Fatalf("create auth session for password invalidation: %v", err)
	}
	newAuthVersion, err := q.ChangeUserPasswordAndInvalidate(ctx, ChangeUserPasswordAndInvalidateParams{ID: user.ID, PasswordHash: "hash:v2:" + suffix})
	if err != nil {
		t.Fatalf("change user password and invalidate sessions: %v", err)
	}
	if newAuthVersion != user.AuthVersion+1 {
		t.Fatalf("new auth version = %d, want %d", newAuthVersion, user.AuthVersion+1)
	}
	gotSession, err = q.GetAuthSession(ctx, passwordSessionID)
	if err != nil {
		t.Fatalf("get password-invalidated auth session: %v", err)
	}
	if !gotSession.RevokedAt.Valid {
		t.Fatalf("password-invalidated auth session has null revoked_at")
	}
	gotByID, err := q.GetUserByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("get user by id after password change: %v", err)
	}
	if gotByID.PasswordHash != "hash:v2:"+suffix || gotByID.MustChangePassword || !gotByID.PasswordChangedAt.Valid {
		t.Fatalf("user after password change = hash:%q must_change:%v changed_at:%v", gotByID.PasswordHash, gotByID.MustChangePassword, gotByID.PasswordChangedAt.Valid)
	}

	if err := q.DeleteUserRbacRoles(ctx, user.ID); err != nil {
		t.Fatalf("delete user rbac roles: %v", err)
	}
	rbacRoles, err = q.GetUserRoleCodesFromRbac(ctx, user.ID)
	if err != nil {
		t.Fatalf("get user role codes after delete: %v", err)
	}
	if len(rbacRoles) != 0 {
		t.Fatalf("rbac role codes after delete = %v, want empty", rbacRoles)
	}
}

func randomTestUUID(t *testing.T) pgtype.UUID {
	t.Helper()
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		t.Fatalf("generate random uuid: %v", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return pgtype.UUID{Bytes: b, Valid: true}
}

func uuidSliceContains(ids []pgtype.UUID, id pgtype.UUID) bool {
	return slices.ContainsFunc(ids, func(candidate pgtype.UUID) bool {
		return candidate == id
	})
}

func authSessionListed(sessions []AuthSession, id pgtype.UUID) bool {
	return slices.ContainsFunc(sessions, func(session AuthSession) bool {
		return session.ID == id
	})
}
