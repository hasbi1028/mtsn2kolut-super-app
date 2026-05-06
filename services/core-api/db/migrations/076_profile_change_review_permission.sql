-- Tahap 5: dedicated reviewer permission for official profile change requests.

INSERT INTO rbac_permissions (code, module, action, description)
VALUES ('profile_changes.review', 'profile_changes', 'review', 'Meninjau dan menyetujui permintaan perubahan data resmi profil.')
ON CONFLICT (code) DO UPDATE
SET module = EXCLUDED.module,
    action = EXCLUDED.action,
    description = EXCLUDED.description,
    is_active = TRUE,
    updated_at = NOW();

INSERT INTO rbac_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM rbac_roles r
JOIN rbac_permissions p ON p.code = 'profile_changes.review'
WHERE r.code = 'admin'
ON CONFLICT (role_id, permission_id) DO NOTHING;
