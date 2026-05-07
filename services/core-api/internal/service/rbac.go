package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "mtsn2kolut-super-app/backend/internal/repository/postgres"
)

type rbacStore interface {
	ListRbacRoles(ctx context.Context) ([]db.RbacRole, error)
	GetRbacRoleByCode(ctx context.Context, code string) (db.RbacRole, error)
	CreateRbacRole(ctx context.Context, arg db.CreateRbacRoleParams) (db.RbacRole, error)
	UpdateRbacRole(ctx context.Context, arg db.UpdateRbacRoleParams) (db.RbacRole, error)
	SetRbacRoleActive(ctx context.Context, arg db.SetRbacRoleActiveParams) (db.RbacRole, error)
	ListRbacPermissions(ctx context.Context) ([]db.RbacPermission, error)
	GetRbacPermissionByCode(ctx context.Context, code string) (db.RbacPermission, error)
	CreateRbacPermission(ctx context.Context, arg db.CreateRbacPermissionParams) (db.RbacPermission, error)
	UpdateRbacPermission(ctx context.Context, arg db.UpdateRbacPermissionParams) (db.RbacPermission, error)
	SetRbacPermissionActive(ctx context.Context, arg db.SetRbacPermissionActiveParams) (db.RbacPermission, error)
	ListRbacRolePermissions(ctx context.Context) ([]db.ListRbacRolePermissionsRow, error)
	ListRbacUserRoles(ctx context.Context) ([]db.ListRbacUserRolesRow, error)
	GetUserPermissionCodes(ctx context.Context, userID pgtype.UUID) ([]string, error)
	GetUserRoleCodesFromRbac(ctx context.Context, userID pgtype.UUID) ([]string, error)
	DeleteRolePermissions(ctx context.Context, code string) error
	AddRolePermissionByCode(ctx context.Context, arg db.AddRolePermissionByCodeParams) error
	DeleteUserRbacRoles(ctx context.Context, userID pgtype.UUID) error
	AddUserRbacRoleByCode(ctx context.Context, arg db.AddUserRbacRoleByCodeParams) error
	RemoveAllUserRoles(ctx context.Context, userID pgtype.UUID) error
	SyncLegacyUserRolesFromRbac(ctx context.Context, userID pgtype.UUID) error
	IncrementUserAuthVersion(ctx context.Context, userID pgtype.UUID) (int32, error)
	RevokeAllAuthSessionsForUser(ctx context.Context, userID pgtype.UUID) (int64, error)
	CountActiveAdminsByRbac(ctx context.Context) (int64, error)
	CountActiveUsersByRbacRole(ctx context.Context, code string) (int64, error)
	UserHasRbacRole(ctx context.Context, arg db.UserHasRbacRoleParams) (bool, error)
	ListUserIDsByRoleCode(ctx context.Context, code string) ([]pgtype.UUID, error)
	ListActiveUserIDsByRoleCodeAnyStatus(ctx context.Context, code string) ([]pgtype.UUID, error)
	ListActiveUserIDsByPermissionCode(ctx context.Context, code string) ([]pgtype.UUID, error)
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

type RBACRoleInput struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RBACPermissionInput struct {
	Code        string `json:"code"`
	Module      string `json:"module"`
	Action      string `json:"action"`
	Description string `json:"description"`
}

var (
	rbacRoleCodePattern       = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	rbacPermissionCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$`)
)

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

func (s *RBAC) CreateRole(ctx context.Context, input RBACRoleInput, actorID pgtype.UUID) (db.RbacRole, error) {
	input = normalizeRoleInput(input, true)
	if err := validateRoleInput(input, true); err != nil {
		return db.RbacRole{}, err
	}
	var role db.RbacRole
	err := s.withRBACStore(ctx, func(store rbacStore) error {
		created, err := store.CreateRbacRole(ctx, db.CreateRbacRoleParams{Code: input.Code, Name: input.Name, Description: input.Description})
		if err != nil {
			return err
		}
		role = created
		return s.audit(ctx, store, actorID, "RBAC_ROLE_CREATED", "rbac_role", role.Code, map[string]any{"code": role.Code, "name": role.Name})
	})
	return role, err
}

func (s *RBAC) UpdateRole(ctx context.Context, roleCode string, input RBACRoleInput, actorID pgtype.UUID) (db.RbacRole, error) {
	roleCode = normalizeRBACCode(roleCode)
	input = normalizeRoleInput(input, false)
	if err := validateRoleInput(input, false); err != nil {
		return db.RbacRole{}, err
	}
	current, err := s.q.GetRbacRoleByCode(ctx, roleCode)
	if err != nil {
		return db.RbacRole{}, err
	}
	if current.IsSystem {
		return db.RbacRole{}, fmt.Errorf("role sistem tidak boleh diubah")
	}
	var role db.RbacRole
	err = s.withRBACStore(ctx, func(store rbacStore) error {
		updated, err := store.UpdateRbacRole(ctx, db.UpdateRbacRoleParams{Code: roleCode, Name: input.Name, Description: input.Description})
		if err != nil {
			return err
		}
		role = updated
		return s.audit(ctx, store, actorID, "RBAC_ROLE_UPDATED", "rbac_role", role.Code, map[string]any{"code": role.Code, "name": role.Name})
	})
	return role, err
}

func (s *RBAC) SetRoleActive(ctx context.Context, roleCode string, active bool, actorID pgtype.UUID) error {
	roleCode = normalizeRBACCode(roleCode)
	role, err := s.q.GetRbacRoleByCode(ctx, roleCode)
	if err != nil {
		return err
	}
	if role.IsSystem && !active {
		return fmt.Errorf("role sistem tidak boleh dinonaktifkan")
	}
	if !active {
		count, err := s.q.CountActiveUsersByRbacRole(ctx, roleCode)
		if err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("role masih dipakai user aktif")
		}
	}
	return s.withRBACStore(ctx, func(store rbacStore) error {
		affectedUsers, err := store.ListActiveUserIDsByRoleCodeAnyStatus(ctx, roleCode)
		if err != nil {
			return err
		}
		updated, err := store.SetRbacRoleActive(ctx, db.SetRbacRoleActiveParams{Code: roleCode, IsActive: active})
		if err != nil {
			return err
		}
		if err := invalidateUsersForRBACChange(ctx, store, affectedUsers); err != nil {
			return err
		}
		return s.audit(ctx, store, actorID, "RBAC_ROLE_STATUS_UPDATED", "rbac_role", updated.Code, map[string]any{"code": updated.Code, "is_active": updated.IsActive})
	})
}

func (s *RBAC) CreatePermission(ctx context.Context, input RBACPermissionInput, actorID pgtype.UUID) (db.RbacPermission, error) {
	input = normalizePermissionInput(input, true)
	if err := validatePermissionInput(input, true); err != nil {
		return db.RbacPermission{}, err
	}
	var permission db.RbacPermission
	err := s.withRBACStore(ctx, func(store rbacStore) error {
		created, err := store.CreateRbacPermission(ctx, db.CreateRbacPermissionParams{Code: input.Code, Module: input.Module, Action: input.Action, Description: input.Description})
		if err != nil {
			return err
		}
		permission = created
		return s.audit(ctx, store, actorID, "RBAC_PERMISSION_CREATED", "rbac_permission", permission.Code, map[string]any{"code": permission.Code, "module": permission.Module, "action": permission.Action})
	})
	return permission, err
}

func (s *RBAC) UpdatePermission(ctx context.Context, code string, input RBACPermissionInput, actorID pgtype.UUID) (db.RbacPermission, error) {
	code = normalizeRBACCode(code)
	input = normalizePermissionInput(input, false)
	if err := validatePermissionInput(input, false); err != nil {
		return db.RbacPermission{}, err
	}
	if _, err := s.q.GetRbacPermissionByCode(ctx, code); err != nil {
		return db.RbacPermission{}, err
	}
	var permission db.RbacPermission
	err := s.withRBACStore(ctx, func(store rbacStore) error {
		updated, err := store.UpdateRbacPermission(ctx, db.UpdateRbacPermissionParams{Code: code, Module: input.Module, Action: input.Action, Description: input.Description})
		if err != nil {
			return err
		}
		permission = updated
		return s.audit(ctx, store, actorID, "RBAC_PERMISSION_UPDATED", "rbac_permission", permission.Code, map[string]any{"code": permission.Code, "module": permission.Module, "action": permission.Action})
	})
	return permission, err
}

func (s *RBAC) SetPermissionActive(ctx context.Context, code string, active bool, actorID pgtype.UUID) error {
	code = normalizeRBACCode(code)
	permission, err := s.q.GetRbacPermissionByCode(ctx, code)
	if err != nil {
		return err
	}
	if !active && isCriticalRBACPermission(permission.Code) {
		admins, err := s.q.CountActiveAdminsByRbac(ctx)
		if err != nil {
			return err
		}
		if admins <= 1 {
			return fmt.Errorf("permission kritikal admin terakhir tidak boleh dinonaktifkan")
		}
	}
	return s.withRBACStore(ctx, func(store rbacStore) error {
		affectedUsers, err := store.ListActiveUserIDsByPermissionCode(ctx, code)
		if err != nil {
			return err
		}
		updated, err := store.SetRbacPermissionActive(ctx, db.SetRbacPermissionActiveParams{Code: code, IsActive: active})
		if err != nil {
			return err
		}
		if err := invalidateUsersForRBACChange(ctx, store, affectedUsers); err != nil {
			return err
		}
		return s.audit(ctx, store, actorID, "RBAC_PERMISSION_STATUS_UPDATED", "rbac_permission", updated.Code, map[string]any{"code": updated.Code, "is_active": updated.IsActive})
	})
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
		affectedUsers, err := store.ListUserIDsByRoleCode(ctx, roleCode)
		if err != nil {
			return err
		}
		if err := store.DeleteRolePermissions(ctx, roleCode); err != nil {
			return err
		}
		for _, code := range permissionCodes {
			if err := store.AddRolePermissionByCode(ctx, db.AddRolePermissionByCodeParams{Code: roleCode, Code_2: code}); err != nil {
				return err
			}
		}
		if err := invalidateUsersForRBACChange(ctx, store, affectedUsers); err != nil {
			return err
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
		if err := store.RemoveAllUserRoles(ctx, userID); err != nil {
			return err
		}
		if err := store.SyncLegacyUserRolesFromRbac(ctx, userID); err != nil {
			return err
		}
		if err := invalidateUsersForRBACChange(ctx, store, []pgtype.UUID{userID}); err != nil {
			return err
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

func invalidateUsersForRBACChange(ctx context.Context, store rbacStore, userIDs []pgtype.UUID) error {
	seen := map[string]struct{}{}
	for _, userID := range userIDs {
		key := uuidEntityID(userID)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		if _, err := store.IncrementUserAuthVersion(ctx, userID); err != nil {
			return err
		}
		if _, err := store.RevokeAllAuthSessionsForUser(ctx, userID); err != nil {
			return err
		}
	}
	return nil
}

func normalizeRoleInput(input RBACRoleInput, includeCode bool) RBACRoleInput {
	if includeCode {
		input.Code = normalizeRBACCode(input.Code)
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	return input
}

func validateRoleInput(input RBACRoleInput, includeCode bool) error {
	if includeCode && !rbacRoleCodePattern.MatchString(input.Code) {
		return fmt.Errorf("kode role tidak valid")
	}
	if input.Name == "" {
		return fmt.Errorf("nama role wajib diisi")
	}
	return nil
}

func normalizePermissionInput(input RBACPermissionInput, includeCode bool) RBACPermissionInput {
	if includeCode {
		input.Code = normalizeRBACCode(input.Code)
	}
	input.Module = normalizeRBACCode(input.Module)
	input.Action = normalizeRBACCode(input.Action)
	input.Description = strings.TrimSpace(input.Description)
	return input
}

func validatePermissionInput(input RBACPermissionInput, includeCode bool) error {
	if includeCode && !rbacPermissionCodePattern.MatchString(input.Code) {
		return fmt.Errorf("kode permission tidak valid")
	}
	if !rbacRoleCodePattern.MatchString(input.Module) {
		return fmt.Errorf("module permission tidak valid")
	}
	if !rbacRoleCodePattern.MatchString(input.Action) {
		return fmt.Errorf("action permission tidak valid")
	}
	return nil
}

func isCriticalRBACPermission(code string) bool {
	return containsString([]string{"roles.manage", "users.manage_roles"}, code)
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
