-- Student and parent account portal foundation.
-- Additive migration: keeps existing roles/users compatible and only seeds new portal permissions.

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS password_changed_at TIMESTAMPTZ;

WITH permission_seed(code, module, action, description) AS (
  VALUES
    ('student_portal.read', 'student_portal', 'read', 'Mengakses portal siswa.'),
    ('student_portal.profile_read', 'student_portal', 'profile_read', 'Melihat profil siswa sendiri.'),
    ('student_portal.schedule_read', 'student_portal', 'schedule_read', 'Melihat jadwal siswa sendiri.'),
    ('student_portal.grades_read', 'student_portal', 'grades_read', 'Melihat hasil/nilai siswa sendiri yang sudah dirilis.'),
    ('student_portal.assessment_take', 'student_portal', 'assessment_take', 'Mengikuti asesmen sebagai siswa.'),
    ('student_portal.profile_change_request', 'student_portal', 'profile_change_request', 'Mengajukan perubahan data resmi siswa sendiri.'),
    ('parent_portal.read', 'parent_portal', 'read', 'Mengakses portal orang tua/wali.'),
    ('parent_portal.children_read', 'parent_portal', 'children_read', 'Melihat daftar anak yang terhubung.'),
    ('parent_portal.child_profile_read', 'parent_portal', 'child_profile_read', 'Melihat profil anak yang terhubung.'),
    ('parent_portal.child_schedule_read', 'parent_portal', 'child_schedule_read', 'Melihat jadwal anak yang terhubung.'),
    ('parent_portal.child_attendance_read', 'parent_portal', 'child_attendance_read', 'Melihat kehadiran anak yang terhubung.'),
    ('parent_portal.child_grades_read', 'parent_portal', 'child_grades_read', 'Melihat hasil/nilai anak yang sudah dirilis.'),
    ('parent_portal.profile_change_request', 'parent_portal', 'profile_change_request', 'Mengajukan perubahan data orang tua/anak melalui approval.'),
    ('student_accounts.manage', 'student_accounts', 'manage', 'Membuat, reset, dan menonaktifkan akun siswa.'),
    ('parent_accounts.manage', 'parent_accounts', 'manage', 'Membuat, reset, dan menonaktifkan akun orang tua/wali.')
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
  SELECT 'admin', code
  FROM rbac_permissions
  WHERE code IN (
    'student_portal.read',
    'student_portal.profile_read',
    'student_portal.schedule_read',
    'student_portal.grades_read',
    'student_portal.assessment_take',
    'student_portal.profile_change_request',
    'parent_portal.read',
    'parent_portal.children_read',
    'parent_portal.child_profile_read',
    'parent_portal.child_schedule_read',
    'parent_portal.child_attendance_read',
    'parent_portal.child_grades_read',
    'parent_portal.profile_change_request',
    'student_accounts.manage',
    'parent_accounts.manage'
  )
  UNION ALL
  VALUES
    ('siswa', 'dashboard.read'),
    ('siswa', 'notifications.read'),
    ('siswa', 'settings.account'),
    ('siswa', 'student_portal.read'),
    ('siswa', 'student_portal.profile_read'),
    ('siswa', 'student_portal.schedule_read'),
    ('siswa', 'student_portal.grades_read'),
    ('siswa', 'student_portal.assessment_take'),
    ('siswa', 'student_portal.profile_change_request'),
    ('ortu', 'dashboard.read'),
    ('ortu', 'notifications.read'),
    ('ortu', 'settings.account'),
    ('ortu', 'parent_portal.read'),
    ('ortu', 'parent_portal.children_read'),
    ('ortu', 'parent_portal.child_profile_read'),
    ('ortu', 'parent_portal.child_schedule_read'),
    ('ortu', 'parent_portal.child_attendance_read'),
    ('ortu', 'parent_portal.child_grades_read'),
    ('ortu', 'parent_portal.profile_change_request'),
    ('kesiswaan', 'student_accounts.manage'),
    ('kesiswaan', 'parent_accounts.manage')
)
INSERT INTO rbac_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM role_permission_seed seed
JOIN rbac_roles r ON r.code = seed.role_code
JOIN rbac_permissions p ON p.code = seed.permission_code
ON CONFLICT (role_id, permission_id) DO NOTHING;
