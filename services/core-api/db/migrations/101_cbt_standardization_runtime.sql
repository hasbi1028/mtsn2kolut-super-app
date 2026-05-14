-- CBT standardization Sprint 1-5 additive runtime fields.
-- Existing sessions, printed participant tokens, and package composition stay valid.

ALTER TABLE cbt_packages
  ADD COLUMN IF NOT EXISTS locked_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS locked_by UUID REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS lock_reason TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS snapshot_version INTEGER NOT NULL DEFAULT 0;

ALTER TABLE cbt_packages
  DROP CONSTRAINT IF EXISTS chk_cbt_packages_snapshot_version_nonnegative;

ALTER TABLE cbt_packages
  ADD CONSTRAINT chk_cbt_packages_snapshot_version_nonnegative
  CHECK (snapshot_version >= 0);

CREATE TABLE IF NOT EXISTS cbt_package_question_snapshots (
  id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  package_id       UUID        NOT NULL REFERENCES cbt_packages(id) ON DELETE CASCADE,
  snapshot_version INTEGER     NOT NULL,
  question_id      UUID        NOT NULL REFERENCES cbt_questions(id) ON DELETE RESTRICT,
  position         INTEGER     NOT NULL,
  points           INTEGER     NOT NULL,
  question_code    TEXT        NOT NULL DEFAULT '',
  question_text    TEXT        NOT NULL,
  question_type    TEXT        NOT NULL,
  options          JSONB       NOT NULL DEFAULT '[]'::jsonb,
  option_a         TEXT        NOT NULL DEFAULT '',
  option_b         TEXT        NOT NULL DEFAULT '',
  option_c         TEXT        NOT NULL DEFAULT '',
  option_d         TEXT        NOT NULL DEFAULT '',
  option_e         TEXT        NOT NULL DEFAULT '',
  answer_key       TEXT        NOT NULL DEFAULT '',
  stem_html        TEXT        NOT NULL DEFAULT '',
  stem_latex       TEXT        NOT NULL DEFAULT '',
  stimulus_html    TEXT        NOT NULL DEFAULT '',
  stimulus_latex   TEXT        NOT NULL DEFAULT '',
  rubric_html      TEXT        NOT NULL DEFAULT '',
  explanation_html TEXT        NOT NULL DEFAULT '',
  media_asset_ids  JSONB       NOT NULL DEFAULT '[]'::jsonb,
  metadata         JSONB       NOT NULL DEFAULT '{}'::jsonb,
  snapshot_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (package_id, snapshot_version, question_id),
  UNIQUE (package_id, snapshot_version, position)
);

CREATE INDEX IF NOT EXISTS idx_cbt_package_question_snapshots_package
  ON cbt_package_question_snapshots (package_id, snapshot_version, position);

ALTER TABLE cbt_exam_participants
  ADD COLUMN IF NOT EXISTS token_hash TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS token_hash_version INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS token_generated_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS token_revealed_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS token_revoked_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS option_order JSONB NOT NULL DEFAULT '{}'::jsonb,
  ADD COLUMN IF NOT EXISTS question_draw_log JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE cbt_exam_participants
  DROP CONSTRAINT IF EXISTS chk_cbt_exam_participants_token_hash_version_nonnegative;

ALTER TABLE cbt_exam_participants
  ADD CONSTRAINT chk_cbt_exam_participants_token_hash_version_nonnegative
  CHECK (token_hash_version >= 0);

CREATE UNIQUE INDEX IF NOT EXISTS idx_cbt_exam_participants_token_hash
  ON cbt_exam_participants (token_hash)
  WHERE token_hash <> '' AND token_revoked_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_cbt_exam_participants_overdue
  ON cbt_exam_participants (session_id, submitted_at)
  WHERE submitted_at IS NULL;

ALTER TABLE cbt_exam_rooms
  ADD COLUMN IF NOT EXISTS room_token_hash TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS room_token_hash_version INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS room_token_generated_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS room_token_revealed_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS room_token_revoked_at TIMESTAMPTZ;

ALTER TABLE cbt_exam_rooms
  DROP CONSTRAINT IF EXISTS chk_cbt_exam_rooms_token_hash_version_nonnegative;

ALTER TABLE cbt_exam_rooms
  ADD CONSTRAINT chk_cbt_exam_rooms_token_hash_version_nonnegative
  CHECK (room_token_hash_version >= 0);

CREATE UNIQUE INDEX IF NOT EXISTS idx_cbt_exam_rooms_token_hash
  ON cbt_exam_rooms (room_token_hash)
  WHERE room_token_hash <> '' AND room_token_revoked_at IS NULL;

CREATE TABLE IF NOT EXISTS cbt_result_sync_runs (
  id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id          UUID        NOT NULL REFERENCES cbt_exam_sessions(id) ON DELETE CASCADE,
  grade_assignment_id UUID        REFERENCES class_subject_assignments(id) ON DELETE SET NULL,
  grade_component_id  UUID        REFERENCES grade_components(id) ON DELETE SET NULL,
  status              TEXT        NOT NULL DEFAULT 'preflight',
  candidate_count     INTEGER     NOT NULL DEFAULT 0,
  synced_count        INTEGER     NOT NULL DEFAULT 0,
  skipped_count       INTEGER     NOT NULL DEFAULT 0,
  threshold           NUMERIC(5,2) NOT NULL DEFAULT 75,
  notes               TEXT        NOT NULL DEFAULT '',
  created_by          TEXT        NOT NULL DEFAULT '',
  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_cbt_result_sync_runs_status
    CHECK (status IN ('preflight', 'synced', 'failed', 'cancelled'))
);

CREATE INDEX IF NOT EXISTS idx_cbt_result_sync_runs_session
  ON cbt_result_sync_runs (session_id, created_at DESC);
