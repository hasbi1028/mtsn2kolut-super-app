-- Migration 084: Internal analytics dashboard permissions.
-- Adds explicit RBAC permissions for aggregate analytics reads and aggregate-only
-- CSV export. No public collector or raw event access is introduced by this
-- migration.

WITH permission_seed(code, module, action, description) AS (
  VALUES
    ('analytics.read', 'analytics', 'read', 'Melihat dashboard analytics internal yang hanya berisi agregat.'),
    ('analytics.export', 'analytics', 'export', 'Mengekspor laporan analytics agregat tanpa event mentah atau metadata sensitif.'),
    ('analytics.security_read', 'analytics', 'security_read', 'Melihat sinyal keamanan analytics yang sudah diagregasi.')
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
JOIN rbac_permissions p ON p.code IN ('analytics.read', 'analytics.export', 'analytics.security_read')
WHERE r.code = 'admin'
ON CONFLICT (role_id, permission_id) DO NOTHING;
