-- Sprint 21: Kesiswaan foundation.
-- Student affairs extends the existing student master instead of creating a second student table.

ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'kesiswaan';

ALTER TABLE students
    ADD COLUMN nik TEXT NOT NULL DEFAULT '',
    ADD COLUMN tempat_lahir TEXT NOT NULL DEFAULT '',
    ADD COLUMN tanggal_lahir DATE,
    ADD COLUMN alamat TEXT NOT NULL DEFAULT '',
    ADD COLUMN agama TEXT NOT NULL DEFAULT '',
    ADD COLUMN anak_ke INTEGER,
    ADD COLUMN phone TEXT NOT NULL DEFAULT '',
    ADD COLUMN photo_url TEXT NOT NULL DEFAULT '',
    ADD COLUMN total_violation_points INTEGER NOT NULL DEFAULT 0,
    ADD CONSTRAINT chk_students_anak_ke CHECK (anak_ke IS NULL OR anak_ke > 0),
    ADD CONSTRAINT chk_students_violation_points CHECK (total_violation_points >= 0);

CREATE UNIQUE INDEX uq_students_nik_not_empty ON students(nik) WHERE nik <> '';

CREATE TABLE violation_categories (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    code        TEXT        NOT NULL UNIQUE,
    name        TEXT        NOT NULL,
    point       INTEGER     NOT NULL DEFAULT 0,
    severity    TEXT        NOT NULL DEFAULT 'ringan',
    description TEXT        NOT NULL DEFAULT '',
    is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_violation_categories_code CHECK (btrim(code) <> ''),
    CONSTRAINT chk_violation_categories_name CHECK (btrim(name) <> ''),
    CONSTRAINT chk_violation_categories_point CHECK (point >= 0),
    CONSTRAINT chk_violation_categories_severity CHECK (severity IN ('ringan', 'sedang', 'berat'))
);

CREATE TABLE student_violations (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id              UUID        NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    category_id             UUID        REFERENCES violation_categories(id) ON DELETE SET NULL,
    incident_date           DATE        NOT NULL,
    points                  INTEGER     NOT NULL DEFAULT 0,
    description             TEXT        NOT NULL DEFAULT '',
    action_taken            TEXT        NOT NULL DEFAULT '',
    status                  TEXT        NOT NULL DEFAULT 'open',
    reported_by_employee_id UUID        REFERENCES employees(id) ON DELETE SET NULL,
    recorded_by_user_id     UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_student_violations_points CHECK (points >= 0),
    CONSTRAINT chk_student_violations_status CHECK (status IN ('open', 'resolved', 'canceled'))
);

CREATE TABLE student_achievements (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id          UUID        NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    achievement_date    DATE        NOT NULL,
    title               TEXT        NOT NULL,
    level               TEXT        NOT NULL DEFAULT 'school',
    category            TEXT        NOT NULL DEFAULT '',
    organizer           TEXT        NOT NULL DEFAULT '',
    description         TEXT        NOT NULL DEFAULT '',
    document_url        TEXT        NOT NULL DEFAULT '',
    recorded_by_user_id UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_student_achievements_title CHECK (btrim(title) <> ''),
    CONSTRAINT chk_student_achievements_level CHECK (level IN ('school', 'district', 'province', 'national', 'international'))
);

CREATE INDEX idx_student_violations_student_date ON student_violations(student_id, incident_date DESC);
CREATE INDEX idx_student_violations_status ON student_violations(status, incident_date DESC);
CREATE INDEX idx_student_violations_category ON student_violations(category_id) WHERE category_id IS NOT NULL;
CREATE INDEX idx_student_achievements_student_date ON student_achievements(student_id, achievement_date DESC);
CREATE INDEX idx_student_achievements_level ON student_achievements(level, achievement_date DESC);

CREATE OR REPLACE FUNCTION refresh_student_violation_points(p_student_id UUID)
RETURNS VOID AS $$
BEGIN
    UPDATE students
    SET total_violation_points = COALESCE((
            SELECT SUM(points)
            FROM student_violations
            WHERE student_id = p_student_id
              AND status <> 'canceled'
        ), 0),
        updated_at = NOW()
    WHERE id = p_student_id;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION trg_refresh_student_violation_points()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        PERFORM refresh_student_violation_points(OLD.student_id);
        RETURN OLD;
    END IF;

    PERFORM refresh_student_violation_points(NEW.student_id);
    IF TG_OP = 'UPDATE' AND OLD.student_id <> NEW.student_id THEN
        PERFORM refresh_student_violation_points(OLD.student_id);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER student_violations_refresh_points
AFTER INSERT OR UPDATE OR DELETE ON student_violations
FOR EACH ROW EXECUTE FUNCTION trg_refresh_student_violation_points();
