-- Migration 095: Backup Center permissions.
-- Adds explicit RBAC permissions for backup observability, safe download, and manual backup creation.
-- Restore execution remains intentionally absent from this migration.

WITH permission_seed(code, module, action, description) AS (
  VALUES
    ('backup.read', 'backup', 'read', 'Melihat status dan daftar backup PostgreSQL.'),
    ('backup.download', 'backup', 'download', 'Mengunduh file backup PostgreSQL yang tersedia.'),
    ('backup.create', 'backup', 'create', 'Menjalankan backup PostgreSQL manual dari Backup Center.')
)
INSERT INTO rbac_permissions (code, module, action, description)
SELECT code, module, action, description FROM permission_seed
ON CONFLICT (code) DO UPDATE
SET module = EXCLUDED.module,
    action = EXCLUDED.action,
    description = EXCLUDED.description,
    is_active = TRUE,
    updated_at = NOW();

INSERT INTO rbac_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM rbac_roles r
JOIN rbac_permissions p ON p.code IN ('backup.read', 'backup.download', 'backup.create')
WHERE r.code = 'admin'
ON CONFLICT (role_id, permission_id) DO NOTHING;
