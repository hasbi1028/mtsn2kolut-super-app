-- Migration: 005_fix_enums_and_cbt_sessions
-- Fixes:
--   1. gender_enum values changed to Indonesian (L/P)
--   2. cbt_question_status_enum adds 'archived'
--   3. cbt_questions.code unique constraint dropped (code is optional)
-- Adds:
--   4. cbt_exam_sessions and related tables

-- 1. Fix gender enum: rename values to L/P
ALTER TYPE gender_enum RENAME VALUE 'male'   TO 'L';
ALTER TYPE gender_enum RENAME VALUE 'female' TO 'P';

-- 2. Add 'archived' to cbt_question_status_enum
ALTER TYPE cbt_question_status_enum ADD VALUE IF NOT EXISTS 'archived';

-- 3. Remove unique constraint on cbt_questions.code
ALTER TABLE cbt_questions DROP CONSTRAINT IF EXISTS cbt_questions_code_key;
-- code may now be empty or duplicated; still required to be non-null

-- ============================================================
-- 4. CBT Exam Sessions
-- ============================================================

CREATE TYPE cbt_session_status_enum AS ENUM ('draft', 'scheduled', 'active', 'finished', 'cancelled');

CREATE TABLE cbt_exam_sessions (
  id               UUID                     PRIMARY KEY DEFAULT gen_random_uuid(),
  package_id       UUID                     NOT NULL REFERENCES cbt_packages(id) ON DELETE RESTRICT,
  class_id         UUID                     NOT NULL REFERENCES school_classes(id) ON DELETE RESTRICT,
  title            TEXT                     NOT NULL,
  scheduled_start  TIMESTAMPTZ              NOT NULL,
  scheduled_end    TIMESTAMPTZ              NOT NULL,
  status           cbt_session_status_enum  NOT NULL DEFAULT 'draft',
  created_at       TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
  UNIQUE (package_id, class_id, scheduled_start)
);

CREATE TABLE cbt_exam_participants (
  id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id     UUID        NOT NULL REFERENCES cbt_exam_sessions(id) ON DELETE CASCADE,
  student_id     UUID        NOT NULL REFERENCES students(id) ON DELETE CASCADE,
  token          TEXT        NOT NULL DEFAULT '',
  joined_at      TIMESTAMPTZ,
  submitted_at   TIMESTAMPTZ,
  score          NUMERIC(5,2),
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (session_id, student_id)
);

CREATE TABLE cbt_student_answers (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  participant_id  UUID        NOT NULL REFERENCES cbt_exam_participants(id) ON DELETE CASCADE,
  question_id     UUID        NOT NULL REFERENCES cbt_questions(id) ON DELETE RESTRICT,
  answer          TEXT        NOT NULL DEFAULT '',
  is_correct      BOOLEAN,
  answered_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (participant_id, question_id)
);

CREATE INDEX idx_cbt_exam_sessions_class   ON cbt_exam_sessions (class_id, scheduled_start DESC);
CREATE INDEX idx_cbt_exam_sessions_package ON cbt_exam_sessions (package_id, status);
CREATE INDEX idx_cbt_participants_session  ON cbt_exam_participants (session_id, student_id);
CREATE INDEX idx_cbt_answers_participant   ON cbt_student_answers (participant_id, question_id);
