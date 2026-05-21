-- Wave 1: ID Card Siswa RBAC seed.

WITH permission_seed(code, module, action, description) AS (
  VALUES
    ('id_cards.read', 'id_cards', 'read', 'Melihat kartu identitas siswa.'),
    ('id_cards.manage', 'id_cards', 'manage', 'Membuat, menerbitkan ulang, mencetak, dan mengubah status kartu identitas siswa.'),
    ('id_cards.scan', 'id_cards', 'scan', 'Memindai kartu identitas siswa untuk layanan operasional.'),
    ('id_cards.audit', 'id_cards', 'audit', 'Melihat audit kartu identitas siswa.')
)
INSERT INTO rbac_permissions (code, module, action, description)
SELECT code, module, action, description FROM permission_seed
ON CONFLICT (code) DO UPDATE
SET module = EXCLUDED.module,
    action = EXCLUDED.action,
    description = EXCLUDED.description,
    is_active = TRUE,
    updated_at = NOW();

WITH role_permission_seed(role_code, permission_code) AS (
  VALUES
    ('admin', 'id_cards.read'),
    ('admin', 'id_cards.manage'),
    ('admin', 'id_cards.scan'),
    ('admin', 'id_cards.audit'),
    ('kesiswaan', 'id_cards.read'),
    ('kesiswaan', 'id_cards.manage'),
    ('kesiswaan', 'id_cards.scan'),
    ('kesiswaan', 'id_cards.audit'),
    ('staf', 'id_cards.scan'),
    ('guru', 'id_cards.scan')
)
INSERT INTO rbac_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM role_permission_seed seed
JOIN rbac_roles r ON r.code = seed.role_code
JOIN rbac_permissions p ON p.code = seed.permission_code
ON CONFLICT (role_id, permission_id) DO NOTHING;
