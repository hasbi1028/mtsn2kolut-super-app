-- Migration 019: Academic RBAC Expansion

-- 1. Expansion of user roles
ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'siswa';
ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'staf';
ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'ortu';

-- 2. Student Status for Lifecycle Management
CREATE TYPE student_status_enum AS ENUM ('prospective', 'active', 'alumni', 'mutated');

ALTER TABLE students ADD COLUMN status student_status_enum NOT NULL DEFAULT 'active';

-- 3. Parent Profiles
CREATE TABLE parents (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    nama       TEXT        NOT NULL,
    phone      TEXT        NOT NULL DEFAULT '',
    address    TEXT        NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 4. Parent-Student Relationship
CREATE TABLE parent_students (
    parent_id  UUID NOT NULL REFERENCES parents(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    PRIMARY KEY (parent_id, student_id)
);

-- 5. Many-to-Many Roles (RBAC)
CREATE TABLE user_roles (
    user_id UUID      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role    user_role NOT NULL,
    PRIMARY KEY (user_id, role)
);

-- 6. Migrate existing single-role users to many-to-many
INSERT INTO user_roles (user_id, role)
SELECT id, role FROM users
ON CONFLICT (user_id, role) DO NOTHING;

-- 7. Link Users to Student or Parent profiles
ALTER TABLE users ADD COLUMN student_id UUID REFERENCES students(id) ON DELETE SET NULL;
ALTER TABLE users ADD COLUMN parent_id  UUID REFERENCES parents(id)  ON DELETE SET NULL;

-- 8. Add Indexes
CREATE INDEX idx_parent_students_student ON parent_students(student_id);
CREATE INDEX idx_users_student_id ON users(student_id) WHERE student_id IS NOT NULL;
CREATE INDEX idx_users_parent_id ON users(parent_id) WHERE parent_id IS NOT NULL;
