-- Add dynamic RBAC permissions for employee master-data access.
-- Admin keeps the default grant; other roles only gain access if explicitly assigned.

WITH permission_seed(code, module, action, description) AS (
  VALUES
    ('employees.read', 'employees', 'read', 'Melihat data master pegawai.'),
    ('employees.manage', 'employees', 'manage', 'Mengelola data master pegawai.')
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
JOIN rbac_permissions p ON p.code IN ('employees.read', 'employees.manage')
WHERE r.code = 'admin'
ON CONFLICT (role_id, permission_id) DO NOTHING;
