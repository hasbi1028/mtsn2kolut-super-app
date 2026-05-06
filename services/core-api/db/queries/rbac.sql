-- name: ListRbacRoles :many
SELECT id, code, name, description, is_system, is_active, created_at, updated_at
FROM rbac_roles
ORDER BY is_system DESC, code;

-- name: ListRbacPermissions :many
SELECT id, code, module, action, description, is_active, created_at, updated_at
FROM rbac_permissions
ORDER BY module, code;

-- name: ListRbacRolePermissions :many
SELECT
    r.code AS role_code,
    p.code AS permission_code,
    p.module AS permission_module,
    p.action AS permission_action
FROM rbac_role_permissions rp
JOIN rbac_roles r ON r.id = rp.role_id
JOIN rbac_permissions p ON p.id = rp.permission_id
ORDER BY r.code, p.module, p.code;

-- name: ListRbacUserRoles :many
SELECT
    ur.user_id,
    u.username,
    r.code AS role_code,
    r.name AS role_name,
    ur.created_at
FROM rbac_user_roles ur
JOIN users u ON u.id = ur.user_id
JOIN rbac_roles r ON r.id = ur.role_id
ORDER BY u.username, r.code;

-- name: GetUserPermissionCodes :many
SELECT DISTINCT p.code
FROM rbac_user_roles ur
JOIN rbac_roles r ON r.id = ur.role_id
JOIN rbac_role_permissions rp ON rp.role_id = r.id
JOIN rbac_permissions p ON p.id = rp.permission_id
WHERE ur.user_id = $1
  AND r.is_active = TRUE
  AND p.is_active = TRUE
ORDER BY p.code;

-- name: GetUserRoleCodesFromRbac :many
SELECT r.code
FROM rbac_user_roles ur
JOIN rbac_roles r ON r.id = ur.role_id
WHERE ur.user_id = $1
  AND r.is_active = TRUE
ORDER BY r.code;

-- name: DeleteRolePermissions :exec
DELETE FROM rbac_role_permissions
WHERE role_id = (SELECT id FROM rbac_roles WHERE code = $1);

-- name: AddRolePermissionByCode :exec
INSERT INTO rbac_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM rbac_roles r
JOIN rbac_permissions p ON p.code = $2
WHERE r.code = $1
  AND r.is_active = TRUE
  AND p.is_active = TRUE
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- name: DeleteUserRbacRoles :exec
DELETE FROM rbac_user_roles
WHERE user_id = $1;

-- name: AddUserRbacRoleByCode :exec
INSERT INTO rbac_user_roles (user_id, role_id)
SELECT $1, r.id
FROM rbac_roles r
WHERE r.code = $2
  AND r.is_active = TRUE
ON CONFLICT (user_id, role_id) DO NOTHING;

-- name: CountActiveAdminsByRbac :one
SELECT count(*)::bigint
FROM users u
JOIN rbac_user_roles ur ON ur.user_id = u.id
JOIN rbac_roles r ON r.id = ur.role_id
WHERE u.is_active = TRUE
  AND r.code = 'admin'
  AND r.is_active = TRUE;

-- name: UserHasRbacRole :one
SELECT EXISTS (
    SELECT 1
    FROM rbac_user_roles ur
    JOIN rbac_roles r ON r.id = ur.role_id
    WHERE ur.user_id = $1
      AND r.code = $2
      AND r.is_active = TRUE
)::boolean;

-- name: ListUserIDsByRoleCode :many
SELECT ur.user_id
FROM rbac_user_roles ur
JOIN rbac_roles r ON r.id = ur.role_id
JOIN users u ON u.id = ur.user_id
WHERE r.code = $1
  AND r.is_active = TRUE
  AND u.is_active = TRUE
ORDER BY ur.user_id;
