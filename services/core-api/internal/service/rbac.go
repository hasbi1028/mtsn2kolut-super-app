package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type rbacStore interface {
	ListRbacRoles(ctx context.Context) ([]db.RbacRole, error)
	ListRbacPermissions(ctx context.Context) ([]db.RbacPermission, error)
	ListRbacRolePermissions(ctx context.Context) ([]db.ListRbacRolePermissionsRow, error)
	ListRbacUserRoles(ctx context.Context) ([]db.ListRbacUserRolesRow, error)
	GetUserPermissionCodes(ctx context.Context, userID pgtype.UUID) ([]string, error)
	GetUserRoleCodesFromRbac(ctx context.Context, userID pgtype.UUID) ([]string, error)
	DeleteRolePermissions(ctx context.Context, code string) error
	AddRolePermissionByCode(ctx context.Context, arg db.AddRolePermissionByCodeParams) error
	DeleteUserRbacRoles(ctx context.Context, userID pgtype.UUID) error
	AddUserRbacRoleByCode(ctx context.Context, arg db.AddUserRbacRoleByCodeParams) error
	CountActiveAdminsByRbac(ctx context.Context) (int64, error)
	UserHasRbacRole(ctx context.Context, arg db.UserHasRbacRoleParams) (bool, error)
	CreateAuditLog(ctx context.Context, arg db.CreateAuditLogParams) (db.AuditLog, error)
}

type rbacTxStarter interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type RBAC struct {
	q  rbacStore
	tx rbacTxStarter
}

type RBACMatrix struct {
	Roles           []db.RbacRole                   `json:"roles"`
	Permissions     []db.RbacPermission             `json:"permissions"`
	RolePermissions []db.ListRbacRolePermissionsRow `json:"role_permissions"`
	UserRoles       []db.ListRbacUserRolesRow       `json:"user_roles"`
}

func NewRBAC(q *db.Queries) *RBAC { return &RBAC{q: q} }

func NewRBACWithPool(pool *pgxpool.Pool) *RBAC { return &RBAC{q: db.New(pool), tx: pool} }

func (s *RBAC) ListMatrix(ctx context.Context) (RBACMatrix, error) {
	roles, err := s.q.ListRbacRoles(ctx)
	if err != nil {
		return RBACMatrix{}, err
	}
	permissions, err := s.q.ListRbacPermissions(ctx)
	if err != nil {
		return RBACMatrix{}, err
	}
	rolePermissions, err := s.q.ListRbacRolePermissions(ctx)
	if err != nil {
		return RBACMatrix{}, err
	}
	userRoles, err := s.q.ListRbacUserRoles(ctx)
	if err != nil {
		return RBACMatrix{}, err
	}
	return RBACMatrix{Roles: roles, Permissions: permissions, RolePermissions: rolePermissions, UserRoles: userRoles}, nil
}

func (s *RBAC) GetUserPermissions(ctx context.Context, userID pgtype.UUID) ([]string, error) {
	return s.q.GetUserPermissionCodes(ctx, userID)
}

func (s *RBAC) GetUserRoles(ctx context.Context, userID pgtype.UUID) ([]string, error) {
	return s.q.GetUserRoleCodesFromRbac(ctx, userID)
}

func (s *RBAC) ReplaceRolePermissions(ctx context.Context, roleCode string, permissionCodes []string, actorID pgtype.UUID) error {
	roleCode = normalizeRBACCode(roleCode)
	permissionCodes = normalizeRBACCodes(permissionCodes)
	if roleCode == "" {
		return fmt.Errorf("role wajib diisi")
	}
	if len(permissionCodes) == 0 {
		return fmt.Errorf("minimal satu permission wajib dipilih")
	}

	roles, permissions, err := s.roleAndPermissionIndexes(ctx)
	if err != nil {
		return err
	}
	role, ok := roles[roleCode]
	if !ok || !role.IsActive {
		return fmt.Errorf("role tidak dikenal atau tidak aktif")
	}
	for _, code := range permissionCodes {
		perm, ok := permissions[code]
		if !ok || !perm.IsActive {
			return fmt.Errorf("permission %s tidak dikenal atau tidak aktif", code)
		}
	}
	if roleCode == "admin" && !containsAll(permissionCodes, []string{"roles.manage", "users.manage_roles"}) {
		admins, err := s.q.CountActiveAdminsByRbac(ctx)
		if err != nil {
			return err
		}
		if admins <= 1 {
			return fmt.Errorf("permission kritikal admin terakhir tidak boleh dicabut")
		}
	}

	return s.withRBACStore(ctx, func(store rbacStore) error {
		if err := store.DeleteRolePermissions(ctx, roleCode); err != nil {
			return err
		}
		for _, code := range permissionCodes {
			if err := store.AddRolePermissionByCode(ctx, db.AddRolePermissionByCodeParams{Code: roleCode, Code_2: code}); err != nil {
				return err
			}
		}
		return s.audit(ctx, store, actorID, "RBAC_ROLE_PERMISSIONS_UPDATED", "rbac_role", roleCode, map[string]any{"role": roleCode, "permissions": permissionCodes})
	})
}

func (s *RBAC) ReplaceUserRoles(ctx context.Context, userID pgtype.UUID, roleCodes []string, actorID pgtype.UUID) error {
	roleCodes = normalizeRBACCodes(roleCodes)
	if len(roleCodes) == 0 {
		return fmt.Errorf("minimal satu role wajib dipilih")
	}
	if err := s.EnsureCanChangeUserRoles(ctx, userID, roleCodes, actorID); err != nil {
		return err
	}
	return s.withRBACStore(ctx, func(store rbacStore) error {
		if err := store.DeleteUserRbacRoles(ctx, userID); err != nil {
			return err
		}
		for _, code := range roleCodes {
			if err := store.AddUserRbacRoleByCode(ctx, db.AddUserRbacRoleByCodeParams{UserID: userID, Code: code}); err != nil {
				return err
			}
		}
		return s.audit(ctx, store, actorID, "USER_ROLES_UPDATED", "user", uuidEntityID(userID), map[string]any{"user_id": uuidEntityID(userID), "roles": roleCodes})
	})
}

func (s *RBAC) EnsureCanChangeUserRoles(ctx context.Context, targetUserID pgtype.UUID, nextRoleCodes []string, actorID pgtype.UUID) error {
	roles, err := s.roleIndex(ctx)
	if err != nil {
		return err
	}
	for _, code := range normalizeRBACCodes(nextRoleCodes) {
		role, ok := roles[code]
		if !ok || !role.IsActive {
			return fmt.Errorf("role %s tidak dikenal atau tidak aktif", code)
		}
	}
	targetIsAdmin, err := s.q.UserHasRbacRole(ctx, db.UserHasRbacRoleParams{UserID: targetUserID, Code: "admin"})
	if err != nil {
		return err
	}
	willBeAdmin := containsString(nextRoleCodes, "admin")
	if targetIsAdmin && !willBeAdmin {
		admins, err := s.q.CountActiveAdminsByRbac(ctx)
		if err != nil {
			return err
		}
		if admins <= 1 {
			return fmt.Errorf("admin aktif terakhir tidak boleh kehilangan role admin")
		}
	}
	if uuidEqual(targetUserID, actorID) && targetIsAdmin && !willBeAdmin {
		return fmt.Errorf("tidak boleh mencabut akses admin diri sendiri")
	}
	return nil
}

func (s *RBAC) roleAndPermissionIndexes(ctx context.Context) (map[string]db.RbacRole, map[string]db.RbacPermission, error) {
	roles, err := s.roleIndex(ctx)
	if err != nil {
		return nil, nil, err
	}
	permissionsRows, err := s.q.ListRbacPermissions(ctx)
	if err != nil {
		return nil, nil, err
	}
	permissions := make(map[string]db.RbacPermission, len(permissionsRows))
	for _, permission := range permissionsRows {
		permissions[permission.Code] = permission
	}
	return roles, permissions, nil
}

func (s *RBAC) roleIndex(ctx context.Context) (map[string]db.RbacRole, error) {
	rows, err := s.q.ListRbacRoles(ctx)
	if err != nil {
		return nil, err
	}
	roles := make(map[string]db.RbacRole, len(rows))
	for _, role := range rows {
		roles[role.Code] = role
	}
	return roles, nil
}

func (s *RBAC) withRBACStore(ctx context.Context, fn func(rbacStore) error) error {
	if s.tx == nil {
		return fn(s.q)
	}
	tx, err := s.tx.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := fn(db.New(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *RBAC) audit(ctx context.Context, store rbacStore, actorID pgtype.UUID, action, entityType, entityID string, metadata map[string]any) error {
	payload, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	_, err = store.CreateAuditLog(ctx, db.CreateAuditLogParams{UserID: actorID, Action: action, EntityType: entityType, EntityID: entityID, Metadata: payload})
	return err
}

func normalizeRBACCodes(codes []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(codes))
	for _, raw := range codes {
		code := normalizeRBACCode(raw)
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		out = append(out, code)
	}
	sort.Strings(out)
	return out
}

func normalizeRBACCode(code string) string {
	return strings.ToLower(strings.TrimSpace(code))
}

func containsAll(values []string, required []string) bool {
	for _, code := range required {
		if !containsString(values, code) {
			return false
		}
	}
	return true
}

func containsString(values []string, target string) bool {
	target = normalizeRBACCode(target)
	for _, value := range values {
		if normalizeRBACCode(value) == target {
			return true
		}
	}
	return false
}

func uuidEntityID(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	return id.String()
}

func uuidEqual(a, b pgtype.UUID) bool {
	return a.Valid && b.Valid && a.Bytes == b.Bytes
}
