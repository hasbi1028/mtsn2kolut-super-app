-- Dynamic RBAC foundation: roles, permissions, role-permission matrix, and user-role assignments.
-- This migration is additive and keeps the existing user_account_roles enum table as compatibility source.

CREATE TABLE IF NOT EXISTS rbac_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (code ~ '^[a-z][a-z0-9_]*$')
);

CREATE TABLE IF NOT EXISTS rbac_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL UNIQUE,
    module TEXT NOT NULL,
    action TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (code ~ '^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$')
);

CREATE TABLE IF NOT EXISTS rbac_role_permissions (
    role_id UUID NOT NULL REFERENCES rbac_roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES rbac_permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS rbac_user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES rbac_roles(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_rbac_role_permissions_permission ON rbac_role_permissions(permission_id);
CREATE INDEX IF NOT EXISTS idx_rbac_user_roles_role ON rbac_user_roles(role_id);

WITH role_seed(code, name, description, is_system, is_active) AS (
    VALUES
        ('admin', 'Administrator', 'Akses penuh sistem dan pengaturan keamanan.', TRUE, TRUE),
        ('guru', 'Guru', 'Akses guru untuk pembelajaran, Bank Soal, nilai, dan Asesmen terkait.', TRUE, TRUE),
        ('staf', 'Staf', 'Akses staf operasional madrasah/TU.', TRUE, TRUE),
        ('kesiswaan', 'Kesiswaan', 'Akses pengelolaan kesiswaan.', TRUE, TRUE),
        ('siswa', 'Siswa', 'Akses portal siswa.', TRUE, TRUE),
        ('ortu', 'Orang Tua/Wali', 'Akses portal orang tua/wali.', TRUE, TRUE)
)
INSERT INTO rbac_roles (code, name, description, is_system, is_active)
SELECT code, name, description, is_system, is_active FROM role_seed
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    description = EXCLUDED.description,
    is_system = EXCLUDED.is_system,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

WITH permission_seed(code, module, action, description) AS (
    VALUES
        ('dashboard.read', 'dashboard', 'read', 'Melihat dashboard utama.'),
        ('notifications.read', 'notifications', 'read', 'Melihat notifikasi.'),
        ('settings.account', 'settings', 'account', 'Mengelola pengaturan akun sendiri.'),
        ('settings.school_profile', 'settings', 'school_profile', 'Mengelola profil madrasah.'),
        ('audit.read', 'audit', 'read', 'Melihat audit log sistem.'),
        ('users.read', 'users', 'read', 'Melihat daftar dan detail user.'),
        ('users.create', 'users', 'create', 'Membuat user baru.'),
        ('users.update', 'users', 'update', 'Mengubah data user.'),
        ('users.deactivate', 'users', 'deactivate', 'Menonaktifkan atau mengaktifkan user.'),
        ('users.reset_password', 'users', 'reset_password', 'Reset password user.'),
        ('users.manage_roles', 'users', 'manage_roles', 'Mengatur role user.'),
        ('roles.read', 'roles', 'read', 'Melihat role dan matriks permission.'),
        ('roles.manage', 'roles', 'manage', 'Mengelola role dan permission.'),
        ('academic.read', 'academic', 'read', 'Melihat data akademik.'),
        ('academic.manage', 'academic', 'manage', 'Mengelola data akademik.'),
        ('students.read', 'students', 'read', 'Melihat data siswa.'),
        ('students.manage', 'students', 'manage', 'Mengelola data siswa.'),
        ('parents.read', 'parents', 'read', 'Melihat data orang tua/wali.'),
        ('parents.manage', 'parents', 'manage', 'Mengelola data orang tua/wali.'),
        ('grades.read', 'grades', 'read', 'Melihat nilai/rapor.'),
        ('grades.manage', 'grades', 'manage', 'Mengelola nilai/rapor.'),
        ('journal.read', 'journal', 'read', 'Melihat jurnal kelas.'),
        ('journal.manage', 'journal', 'manage', 'Mengelola jurnal kelas.'),
        ('bank_soal.read', 'bank_soal', 'read', 'Melihat Bank Soal.'),
        ('bank_soal.create', 'bank_soal', 'create', 'Membuat soal Bank Soal.'),
        ('bank_soal.update', 'bank_soal', 'update', 'Mengubah soal Bank Soal.'),
        ('bank_soal.review', 'bank_soal', 'review', 'Melakukan review/verifikasi soal.'),
        ('bank_soal.publish', 'bank_soal', 'publish', 'Mempublikasikan soal.'),
        ('bank_soal.import', 'bank_soal', 'import', 'Impor soal.'),
        ('bank_soal.delete', 'bank_soal', 'delete', 'Menghapus soal.'),
        ('bank_soal.analytics', 'bank_soal', 'analytics', 'Melihat analisis Bank Soal.'),
        ('bank_soal.settings', 'bank_soal', 'settings', 'Melihat/mengelola pengaturan Bank Soal.'),
        ('asesmen.read', 'asesmen', 'read', 'Melihat modul Asesmen.'),
        ('asesmen.package_manage', 'asesmen', 'package_manage', 'Mengelola paket Asesmen.'),
        ('asesmen.event_manage', 'asesmen', 'event_manage', 'Mengelola kegiatan Asesmen.'),
        ('asesmen.session_manage', 'asesmen', 'session_manage', 'Mengelola sesi Asesmen.'),
        ('asesmen.participant_manage', 'asesmen', 'participant_manage', 'Mengelola peserta Asesmen.'),
        ('asesmen.proctor', 'asesmen', 'proctor', 'Mengawasi pelaksanaan Asesmen.'),
        ('asesmen.score', 'asesmen', 'score', 'Mengoreksi/menilai Asesmen.'),
        ('asesmen.result_read', 'asesmen', 'result_read', 'Melihat hasil Asesmen.'),
        ('asesmen.result_manage', 'asesmen', 'result_manage', 'Mengelola hasil Asesmen.'),
        ('kesiswaan.read', 'kesiswaan', 'read', 'Melihat modul kesiswaan.'),
        ('kesiswaan.manage', 'kesiswaan', 'manage', 'Mengelola modul kesiswaan.'),
        ('letters.read', 'letters', 'read', 'Melihat persuratan.'),
        ('letters.manage', 'letters', 'manage', 'Mengelola persuratan.'),
        ('archives.read', 'archives', 'read', 'Melihat arsip.'),
        ('archives.manage', 'archives', 'manage', 'Mengelola arsip.'),
        ('inventory.read', 'inventory', 'read', 'Melihat inventaris.'),
        ('inventory.manage', 'inventory', 'manage', 'Mengelola inventaris.'),
        ('library.read', 'library', 'read', 'Melihat perpustakaan.'),
        ('library.manage', 'library', 'manage', 'Mengelola perpustakaan.'),
        ('governance.read', 'governance', 'read', 'Melihat tata kelola.'),
        ('governance.manage', 'governance', 'manage', 'Mengelola tata kelola.'),
        ('document_cycles.read', 'document_cycles', 'read', 'Melihat siklus dokumen.'),
        ('document_cycles.manage', 'document_cycles', 'manage', 'Mengelola siklus dokumen.'),
        ('pusaka.read', 'pusaka', 'read', 'Melihat modul PUSAKA.'),
        ('pusaka.manage', 'pusaka', 'manage', 'Mengelola modul PUSAKA.'),
        ('pusaka.sync', 'pusaka', 'sync', 'Menjalankan sinkronisasi PUSAKA.'),
        ('pusaka.credentials_manage', 'pusaka', 'credentials_manage', 'Mengelola kredensial PUSAKA.'),
        ('website.read', 'website', 'read', 'Melihat konten website.'),
        ('website.manage', 'website', 'manage', 'Mengelola konten website.')
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
    SELECT 'admin', code FROM rbac_permissions
    UNION ALL
    VALUES
        ('guru', 'dashboard.read'),
        ('guru', 'notifications.read'),
        ('guru', 'settings.account'),
        ('guru', 'students.read'),
        ('guru', 'grades.read'),
        ('guru', 'grades.manage'),
        ('guru', 'journal.read'),
        ('guru', 'journal.manage'),
        ('guru', 'bank_soal.read'),
        ('guru', 'bank_soal.create'),
        ('guru', 'bank_soal.update'),
        ('guru', 'bank_soal.review'),
        ('guru', 'bank_soal.import'),
        ('guru', 'bank_soal.analytics'),
        ('guru', 'bank_soal.settings'),
        ('guru', 'asesmen.read'),
        ('guru', 'asesmen.score'),
        ('guru', 'asesmen.result_read'),
        ('staf', 'dashboard.read'),
        ('staf', 'notifications.read'),
        ('staf', 'settings.account'),
        ('staf', 'letters.read'),
        ('staf', 'letters.manage'),
        ('staf', 'archives.read'),
        ('staf', 'archives.manage'),
        ('staf', 'inventory.read'),
        ('staf', 'inventory.manage'),
        ('staf', 'library.read'),
        ('staf', 'library.manage'),
        ('staf', 'governance.read'),
        ('staf', 'governance.manage'),
        ('staf', 'document_cycles.read'),
        ('staf', 'document_cycles.manage'),
        ('staf', 'asesmen.read'),
        ('staf', 'asesmen.proctor'),
        ('kesiswaan', 'dashboard.read'),
        ('kesiswaan', 'notifications.read'),
        ('kesiswaan', 'settings.account'),
        ('kesiswaan', 'students.read'),
        ('kesiswaan', 'students.manage'),
        ('kesiswaan', 'parents.read'),
        ('kesiswaan', 'kesiswaan.read'),
        ('kesiswaan', 'kesiswaan.manage'),
        ('siswa', 'dashboard.read'),
        ('siswa', 'notifications.read'),
        ('siswa', 'settings.account'),
        ('ortu', 'dashboard.read'),
        ('ortu', 'notifications.read'),
        ('ortu', 'settings.account')
)
INSERT INTO rbac_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM role_permission_seed seed
JOIN rbac_roles r ON r.code = seed.role_code
JOIN rbac_permissions p ON p.code = seed.permission_code
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO rbac_user_roles (user_id, role_id)
SELECT uar.user_id, rr.id
FROM user_account_roles uar
JOIN rbac_roles rr ON rr.code = uar.role::text
ON CONFLICT (user_id, role_id) DO NOTHING;
