-- Migration: 004_cbt_foundation

CREATE TYPE gender_enum AS ENUM ('male', 'female');
CREATE TYPE cbt_question_status_enum AS ENUM ('draft', 'published');
CREATE TYPE cbt_question_difficulty_enum AS ENUM ('easy', 'medium', 'hard');

CREATE TABLE academic_years (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  name       TEXT        NOT NULL UNIQUE,
  start_date DATE        NOT NULL,
  end_date   DATE        NOT NULL,
  is_active  BOOLEAN     NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_academic_years_active_true
  ON academic_years (is_active)
  WHERE is_active = TRUE;

CREATE TABLE school_classes (
  id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  academic_year_id UUID        NOT NULL REFERENCES academic_years(id) ON DELETE CASCADE,
  code             TEXT        NOT NULL,
  name             TEXT        NOT NULL,
  level            TEXT        NOT NULL,
  is_active        BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (academic_year_id, code)
);

CREATE TABLE subjects (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  code       TEXT        NOT NULL UNIQUE,
  name       TEXT        NOT NULL,
  is_active  BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE class_subject_assignments (
  id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  class_id            UUID        NOT NULL REFERENCES school_classes(id) ON DELETE CASCADE,
  subject_id          UUID        NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
  teacher_employee_id UUID        NOT NULL REFERENCES employees(id),
  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (class_id, subject_id)
);

CREATE TABLE students (
  id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  nis          TEXT        NOT NULL UNIQUE,
  nisn         TEXT        NOT NULL DEFAULT '',
  nama         TEXT        NOT NULL,
  gender       gender_enum NOT NULL,
  parent_name  TEXT        NOT NULL DEFAULT '',
  parent_phone TEXT        NOT NULL DEFAULT '',
  class_id     UUID        REFERENCES school_classes(id) ON DELETE SET NULL,
  is_active    BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE cbt_questions (
  id           UUID                         PRIMARY KEY DEFAULT gen_random_uuid(),
  subject_id   UUID                         NOT NULL REFERENCES subjects(id) ON DELETE RESTRICT,
  code         TEXT                         NOT NULL UNIQUE,
  question_text TEXT                        NOT NULL,
  option_a     TEXT                         NOT NULL,
  option_b     TEXT                         NOT NULL,
  option_c     TEXT                         NOT NULL,
  option_d     TEXT                         NOT NULL,
  option_e     TEXT                         NOT NULL DEFAULT '',
  answer_key   TEXT                         NOT NULL CHECK (answer_key IN ('A', 'B', 'C', 'D', 'E')),
  explanation  TEXT                         NOT NULL DEFAULT '',
  difficulty   cbt_question_difficulty_enum NOT NULL DEFAULT 'medium',
  status       cbt_question_status_enum     NOT NULL DEFAULT 'draft',
  created_at   TIMESTAMPTZ                  NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ                  NOT NULL DEFAULT NOW()
);

CREATE TABLE cbt_packages (
  id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  subject_id          UUID        NOT NULL REFERENCES subjects(id) ON DELETE RESTRICT,
  title               TEXT        NOT NULL,
  description         TEXT        NOT NULL DEFAULT '',
  duration_minutes    INTEGER     NOT NULL DEFAULT 60,
  randomize_questions BOOLEAN     NOT NULL DEFAULT FALSE,
  is_active           BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE cbt_package_questions (
  package_id  UUID        NOT NULL REFERENCES cbt_packages(id) ON DELETE CASCADE,
  question_id UUID        NOT NULL REFERENCES cbt_questions(id) ON DELETE RESTRICT,
  position    INTEGER     NOT NULL DEFAULT 1,
  points      INTEGER     NOT NULL DEFAULT 1,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (package_id, question_id)
);

CREATE INDEX idx_school_classes_academic_year  ON school_classes (academic_year_id, level, name);
CREATE INDEX idx_students_class                ON students (class_id, nama);
CREATE INDEX idx_cbt_questions_subject_status  ON cbt_questions (subject_id, status, created_at DESC);
CREATE INDEX idx_cbt_packages_subject          ON cbt_packages (subject_id, created_at DESC);
CREATE INDEX idx_cbt_package_questions_package ON cbt_package_questions (package_id, position);
