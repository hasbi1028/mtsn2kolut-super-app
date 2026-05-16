-- Sprint 1 Bank Soal role workflow foundation.
-- Additive only: seed new permissions and reviewer scopes without revoking
-- existing guru permissions or enforcing scope-based question visibility.

WITH permission_seed(code, module, action, description) AS (
    VALUES
        ('bank_soal.update_own', 'bank_soal', 'update_own', 'Mengubah soal Bank Soal milik sendiri.'),
        ('bank_soal.submit', 'bank_soal', 'submit', 'Mengirim soal Bank Soal ke alur review.'),
        ('bank_soal.approve', 'bank_soal', 'approve', 'Menyetujui soal Bank Soal secara final.'),
        ('bank_soal.assign_reviewer', 'bank_soal', 'assign_reviewer', 'Mengelola scope reviewer dan approver Bank Soal.'),
        ('bank_soal.read_all', 'bank_soal', 'read_all', 'Melihat seluruh soal Bank Soal lintas pemilik dan scope.'),
        ('bank_soal.use_in_package', 'bank_soal', 'use_in_package', 'Menggunakan soal Bank Soal pada paket asesmen.'),
        ('bank_soal.audit', 'bank_soal', 'audit', 'Melihat audit dan riwayat workflow Bank Soal.')
)
INSERT INTO rbac_permissions (code, module, action, description, is_active)
SELECT code, module, action, description, TRUE
FROM permission_seed
ON CONFLICT (code) DO UPDATE
SET module = EXCLUDED.module,
    action = EXCLUDED.action,
    description = EXCLUDED.description,
    is_active = TRUE,
    updated_at = NOW();

WITH role_permission_seed(role_code, permission_code) AS (
    SELECT r.code, p.code
    FROM rbac_roles r
    JOIN rbac_permissions p ON p.module = 'bank_soal'
    WHERE r.code IN ('admin', 'superadmin')
    UNION ALL
    VALUES
        ('guru', 'bank_soal.read'),
        ('guru', 'bank_soal.create'),
        ('guru', 'bank_soal.update_own'),
        ('guru', 'bank_soal.submit')
)
INSERT INTO rbac_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM role_permission_seed seed
JOIN rbac_roles r ON r.code = seed.role_code
JOIN rbac_permissions p ON p.code = seed.permission_code
ON CONFLICT (role_id, permission_id) DO NOTHING;

CREATE TABLE IF NOT EXISTS bank_soal_reviewer_scopes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject_id UUID REFERENCES subjects(id) ON DELETE CASCADE,
    grade_level SMALLINT,
    can_review BOOLEAN NOT NULL DEFAULT TRUE,
    can_approve BOOLEAN NOT NULL DEFAULT FALSE,
    assigned_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, subject_id, grade_level),
    CHECK (grade_level IS NULL OR grade_level BETWEEN 7 AND 9)
);

-- The table-level UNIQUE keeps the proposed schema, while this expression
-- index makes "all subjects" / "all grades" scopes unique under PostgreSQL's
-- normal NULL semantics.
CREATE UNIQUE INDEX IF NOT EXISTS idx_bank_soal_reviewer_scopes_unique_nullsafe
    ON bank_soal_reviewer_scopes (
        user_id,
        COALESCE(subject_id, '00000000-0000-0000-0000-000000000000'::uuid),
        COALESCE(grade_level, -1)
    );

CREATE INDEX IF NOT EXISTS idx_bank_soal_reviewer_scopes_user
    ON bank_soal_reviewer_scopes(user_id, can_review, can_approve);

CREATE INDEX IF NOT EXISTS idx_bank_soal_reviewer_scopes_subject_grade
    ON bank_soal_reviewer_scopes(subject_id, grade_level);
