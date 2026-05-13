-- Migration 096: Backup restore planning permission.
-- Adds a non-destructive restore planning permission for validating backup metadata and generating SOP commands.
-- This does NOT add restore execution from the application.

INSERT INTO rbac_permissions (code, module, action, description)
VALUES
  ('backup.restore_plan', 'backup', 'restore_plan', 'Memvalidasi metadata backup dan membuat SOP/perintah restore manual tanpa menjalankan restore production.')
ON CONFLICT (code) DO UPDATE SET
  module = EXCLUDED.module,
  action = EXCLUDED.action,
  description = EXCLUDED.description;

INSERT INTO rbac_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM rbac_roles r
JOIN rbac_permissions p ON p.code = 'backup.restore_plan'
WHERE r.code = 'admin'
ON CONFLICT DO NOTHING;
