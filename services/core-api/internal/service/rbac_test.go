package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type fakeRBACStore struct {
	roles           []db.RbacRole
	permissions     []db.RbacPermission
	rolePermissions []db.ListRbacRolePermissionsRow
	userRoles       []db.ListRbacUserRolesRow
	userPerms       []string
	userRoleCodes   []string
	activeAdmins    int64
	userHasRole     bool

	listRolesErr           error
	listPermissionsErr     error
	listRolePermissionsErr error
	listUserRolesErr       error
	userPermsErr           error
	userRolesErr           error
	deleteRolePermErr      error
	addRolePermErr         error
	deleteUserRolesErr     error
	addUserRoleErr         error
	countAdminsErr         error
	userHasRoleErr         error

	deletedRolePerms []string
	addedRolePerms   []db.AddRolePermissionByCodeParams
	deletedUserRoles []pgtype.UUID
	addedUserRoles   []db.AddUserRbacRoleByCodeParams
	auditActions     []string
}

func rbacTestUUID(seed byte) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte{seed, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, seed}, Valid: true}
}

func (f *fakeRBACStore) ListRbacRoles(ctx context.Context) ([]db.RbacRole, error) {
	return f.roles, f.listRolesErr
}
func (f *fakeRBACStore) ListRbacPermissions(ctx context.Context) ([]db.RbacPermission, error) {
	return f.permissions, f.listPermissionsErr
}
func (f *fakeRBACStore) ListRbacRolePermissions(ctx context.Context) ([]db.ListRbacRolePermissionsRow, error) {
	return f.rolePermissions, f.listRolePermissionsErr
}
func (f *fakeRBACStore) ListRbacUserRoles(ctx context.Context) ([]db.ListRbacUserRolesRow, error) {
	return f.userRoles, f.listUserRolesErr
}
func (f *fakeRBACStore) GetUserPermissionCodes(ctx context.Context, userID pgtype.UUID) ([]string, error) {
	return f.userPerms, f.userPermsErr
}
func (f *fakeRBACStore) GetUserRoleCodesFromRbac(ctx context.Context, userID pgtype.UUID) ([]string, error) {
	return f.userRoleCodes, f.userRolesErr
}
func (f *fakeRBACStore) DeleteRolePermissions(ctx context.Context, code string) error {
	f.deletedRolePerms = append(f.deletedRolePerms, code)
	return f.deleteRolePermErr
}
func (f *fakeRBACStore) AddRolePermissionByCode(ctx context.Context, arg db.AddRolePermissionByCodeParams) error {
	f.addedRolePerms = append(f.addedRolePerms, arg)
	return f.addRolePermErr
}
func (f *fakeRBACStore) DeleteUserRbacRoles(ctx context.Context, userID pgtype.UUID) error {
	f.deletedUserRoles = append(f.deletedUserRoles, userID)
	return f.deleteUserRolesErr
}
func (f *fakeRBACStore) AddUserRbacRoleByCode(ctx context.Context, arg db.AddUserRbacRoleByCodeParams) error {
	f.addedUserRoles = append(f.addedUserRoles, arg)
	return f.addUserRoleErr
}
func (f *fakeRBACStore) CountActiveAdminsByRbac(ctx context.Context) (int64, error) {
	return f.activeAdmins, f.countAdminsErr
}
func (f *fakeRBACStore) UserHasRbacRole(ctx context.Context, arg db.UserHasRbacRoleParams) (bool, error) {
	return f.userHasRole, f.userHasRoleErr
}
func (f *fakeRBACStore) CreateAuditLog(ctx context.Context, arg db.CreateAuditLogParams) (db.AuditLog, error) {
	f.auditActions = append(f.auditActions, arg.Action)
	return db.AuditLog{}, nil
}

func TestRBACListMatrixAggregatesRolesPermissionsAndMappings(t *testing.T) {
	store := &fakeRBACStore{
		roles:       []db.RbacRole{{Code: "admin", Name: "Administrator"}, {Code: "guru", Name: "Guru"}},
		permissions: []db.RbacPermission{{Code: "users.read", Module: "users"}, {Code: "bank_soal.read", Module: "bank_soal"}},
		rolePermissions: []db.ListRbacRolePermissionsRow{
			{RoleCode: "admin", PermissionCode: "users.read"},
			{RoleCode: "guru", PermissionCode: "bank_soal.read"},
		},
		userRoles: []db.ListRbacUserRolesRow{{Username: "admin", RoleCode: "admin"}},
	}
	svc := &RBAC{q: store}

	matrix, err := svc.ListMatrix(context.Background())
	if err != nil {
		t.Fatalf("ListMatrix() error = %v", err)
	}
	if len(matrix.Roles) != 2 || len(matrix.Permissions) != 2 || len(matrix.RolePermissions) != 2 || len(matrix.UserRoles) != 1 {
		t.Fatalf("ListMatrix() = %+v, want all RBAC slices", matrix)
	}
}

func TestRBACGetUserCapabilitiesForwardStoreResults(t *testing.T) {
	userID := rbacTestUUID(10)
	store := &fakeRBACStore{userPerms: []string{"bank_soal.read", "users.read"}, userRoleCodes: []string{"admin"}}
	svc := &RBAC{q: store}

	perms, err := svc.GetUserPermissions(context.Background(), userID)
	if err != nil || !reflect.DeepEqual(perms, store.userPerms) {
		t.Fatalf("GetUserPermissions() = %v, %v; want %v", perms, err, store.userPerms)
	}
	roles, err := svc.GetUserRoles(context.Background(), userID)
	if err != nil || !reflect.DeepEqual(roles, store.userRoleCodes) {
		t.Fatalf("GetUserRoles() = %v, %v; want %v", roles, err, store.userRoleCodes)
	}
}

func TestRBACReplaceUserRolesValidatesUnknownAndLastAdmin(t *testing.T) {
	adminID := rbacTestUUID(11)
	actorID := rbacTestUUID(12)
	store := &fakeRBACStore{roles: []db.RbacRole{{Code: "admin", IsActive: true}, {Code: "guru", IsActive: true}}, activeAdmins: 1, userHasRole: true}
	svc := &RBAC{q: store}

	if err := svc.ReplaceUserRoles(context.Background(), adminID, []string{"unknown"}, actorID); err == nil {
		t.Fatal("ReplaceUserRoles(unknown) error = nil, want validation error")
	}
	if err := svc.ReplaceUserRoles(context.Background(), adminID, []string{"guru"}, actorID); err == nil {
		t.Fatal("ReplaceUserRoles(remove last admin) error = nil, want safety error")
	}
	if len(store.deletedUserRoles) != 0 || len(store.addedUserRoles) != 0 {
		t.Fatalf("mutations after failed validation = deleted %v added %v, want none", store.deletedUserRoles, store.addedUserRoles)
	}
}

func TestRBACReplaceUserRolesSucceedsAndAudits(t *testing.T) {
	userID := rbacTestUUID(13)
	actorID := rbacTestUUID(14)
	store := &fakeRBACStore{roles: []db.RbacRole{{Code: "guru", IsActive: true}, {Code: "staf", IsActive: true}}, activeAdmins: 2, userHasRole: false}
	svc := &RBAC{q: store}

	if err := svc.ReplaceUserRoles(context.Background(), userID, []string{"guru", "staf", "guru"}, actorID); err != nil {
		t.Fatalf("ReplaceUserRoles() error = %v", err)
	}
	if len(store.deletedUserRoles) != 1 || store.deletedUserRoles[0] != userID {
		t.Fatalf("deleted user roles = %v, want target user", store.deletedUserRoles)
	}
	if got := len(store.addedUserRoles); got != 2 {
		t.Fatalf("added user roles len = %d, want de-duplicated 2: %+v", got, store.addedUserRoles)
	}
	if !reflect.DeepEqual(store.auditActions, []string{"USER_ROLES_UPDATED"}) {
		t.Fatalf("audit actions = %v, want USER_ROLES_UPDATED", store.auditActions)
	}
}

func TestRBACReplaceRolePermissionsProtectsCriticalAdminPermissions(t *testing.T) {
	actorID := rbacTestUUID(15)
	store := &fakeRBACStore{
		roles:        []db.RbacRole{{Code: "admin", IsActive: true}, {Code: "guru", IsActive: true}},
		permissions:  []db.RbacPermission{{Code: "roles.manage", IsActive: true}, {Code: "users.manage_roles", IsActive: true}, {Code: "users.read", IsActive: true}},
		activeAdmins: 1,
	}
	svc := &RBAC{q: store}

	if err := svc.ReplaceRolePermissions(context.Background(), "admin", []string{"users.read"}, actorID); err == nil {
		t.Fatal("ReplaceRolePermissions(admin loses critical permissions) error = nil, want safety error")
	}
	if err := svc.ReplaceRolePermissions(context.Background(), "guru", []string{"users.read", "users.read"}, actorID); err != nil {
		t.Fatalf("ReplaceRolePermissions(guru) error = %v", err)
	}
	if len(store.deletedRolePerms) != 1 || store.deletedRolePerms[0] != "guru" || len(store.addedRolePerms) != 1 {
		t.Fatalf("role permission mutations = deleted %v added %v", store.deletedRolePerms, store.addedRolePerms)
	}
}

func TestRBACStoreErrorsPropagate(t *testing.T) {
	expected := errors.New("store down")
	svc := &RBAC{q: &fakeRBACStore{listRolesErr: expected}}
	if _, err := svc.ListMatrix(context.Background()); !errors.Is(err, expected) {
		t.Fatalf("ListMatrix() error = %v, want %v", err, expected)
	}
}
