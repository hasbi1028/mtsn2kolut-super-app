-- Sprint Fondasi Rombel + Orang Tua
-- Additive normalization over the existing school_classes/students/parents model.

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'parent_relationship_enum') THEN
        CREATE TYPE parent_relationship_enum AS ENUM ('ayah', 'ibu', 'wali', 'lainnya');
    END IF;
END $$;

ALTER TABLE parents
    ADD COLUMN IF NOT EXISTS occupation TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS income_band TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS nik TEXT NOT NULL DEFAULT '';

CREATE UNIQUE INDEX IF NOT EXISTS uq_parents_nik_not_empty
    ON parents(nik)
    WHERE nik <> '';

ALTER TABLE parent_students
    ADD COLUMN IF NOT EXISTS relationship parent_relationship_enum NOT NULL DEFAULT 'wali',
    ADD COLUMN IF NOT EXISTS is_primary_contact BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS notes TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_parent_students_relationship
    ON parent_students(student_id, relationship);

CREATE UNIQUE INDEX IF NOT EXISTS uq_parent_students_primary_contact
    ON parent_students(student_id)
    WHERE is_primary_contact = TRUE;

CREATE TABLE IF NOT EXISTS class_homeroom_assignments (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    class_id         UUID        NOT NULL REFERENCES school_classes(id) ON DELETE CASCADE,
    employee_id      UUID        NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    academic_year_id UUID        REFERENCES academic_years(id) ON DELETE SET NULL,
    start_date       DATE        NOT NULL DEFAULT CURRENT_DATE,
    end_date         DATE,
    is_active        BOOLEAN     NOT NULL DEFAULT TRUE,
    notes            TEXT        NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT ck_class_homeroom_date_range CHECK (end_date IS NULL OR end_date >= start_date)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_class_homeroom_active_class
    ON class_homeroom_assignments(class_id)
    WHERE is_active = TRUE;

CREATE INDEX IF NOT EXISTS idx_class_homeroom_class_dates
    ON class_homeroom_assignments(class_id, start_date DESC);

CREATE INDEX IF NOT EXISTS idx_class_homeroom_employee_active
    ON class_homeroom_assignments(employee_id, is_active);
