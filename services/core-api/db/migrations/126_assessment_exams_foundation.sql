-- Web-first Asesmen/CBT rebuild foundation.
-- Additive only: new assessment_* tables do not replace legacy CBT tables.

CREATE TABLE IF NOT EXISTS assessment_exams (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  title text NOT NULL,
  subject_id uuid REFERENCES subjects(id) ON DELETE SET NULL,
  grade_level smallint CHECK (grade_level IS NULL OR grade_level BETWEEN 7 AND 9),
  status text NOT NULL DEFAULT 'draft'
    CHECK (status IN ('draft', 'ready', 'running', 'finished', 'archived')),
  starts_at timestamptz,
  ends_at timestamptz,
  created_by uuid REFERENCES users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT assessment_exams_time_chk
    CHECK (starts_at IS NULL OR ends_at IS NULL OR ends_at > starts_at)
);

CREATE INDEX IF NOT EXISTS idx_assessment_exams_status_created_at
  ON assessment_exams (status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_assessment_exams_starts_at
  ON assessment_exams (starts_at);

CREATE TABLE IF NOT EXISTS assessment_sessions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  exam_id uuid NOT NULL REFERENCES assessment_exams(id) ON DELETE CASCADE,
  title text NOT NULL,
  starts_at timestamptz,
  ends_at timestamptz,
  status text NOT NULL DEFAULT 'draft'
    CHECK (status IN ('draft', 'ready', 'running', 'finished', 'archived')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT assessment_sessions_time_chk
    CHECK (starts_at IS NULL OR ends_at IS NULL OR ends_at > starts_at)
);

CREATE INDEX IF NOT EXISTS idx_assessment_sessions_exam
  ON assessment_sessions (exam_id, created_at);

CREATE TABLE IF NOT EXISTS assessment_rooms (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id uuid NOT NULL REFERENCES assessment_sessions(id) ON DELETE CASCADE,
  code text NOT NULL,
  name text NOT NULL,
  capacity integer NOT NULL DEFAULT 30 CHECK (capacity > 0),
  status text NOT NULL DEFAULT 'draft'
    CHECK (status IN ('draft', 'ready', 'running', 'finished', 'needs_check')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (session_id, code)
);

CREATE INDEX IF NOT EXISTS idx_assessment_rooms_session
  ON assessment_rooms (session_id, code);

CREATE TABLE IF NOT EXISTS assessment_participants (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id uuid NOT NULL REFERENCES assessment_sessions(id) ON DELETE CASCADE,
  room_id uuid REFERENCES assessment_rooms(id) ON DELETE SET NULL,
  student_id uuid NOT NULL REFERENCES students(id) ON DELETE RESTRICT,
  status text NOT NULL DEFAULT 'registered'
    CHECK (status IN ('registered', 'waiting', 'active', 'submitted', 'absent', 'blocked')),
  started_at timestamptz,
  submitted_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (session_id, student_id),
  CONSTRAINT assessment_participants_submit_time_chk
    CHECK (started_at IS NULL OR submitted_at IS NULL OR submitted_at >= started_at)
);

CREATE INDEX IF NOT EXISTS idx_assessment_participants_session_room
  ON assessment_participants (session_id, room_id, status);

CREATE TABLE IF NOT EXISTS assessment_access_cards (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  card_type text NOT NULL CHECK (card_type IN ('participant', 'proctor')),
  session_id uuid NOT NULL REFERENCES assessment_sessions(id) ON DELETE CASCADE,
  room_id uuid REFERENCES assessment_rooms(id) ON DELETE CASCADE,
  participant_id uuid REFERENCES assessment_participants(id) ON DELETE CASCADE,
  token_hash text NOT NULL,
  pin_hash text NOT NULL,
  status text NOT NULL DEFAULT 'active'
    CHECK (status IN ('active', 'revoked', 'expired')),
  failed_attempts integer NOT NULL DEFAULT 0 CHECK (failed_attempts >= 0),
  expires_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CHECK (
    (card_type = 'participant' AND participant_id IS NOT NULL)
    OR (card_type = 'proctor' AND room_id IS NOT NULL AND participant_id IS NULL)
  ),
  UNIQUE (participant_id),
  UNIQUE (room_id, card_type)
);

CREATE INDEX IF NOT EXISTS idx_assessment_access_cards_session
  ON assessment_access_cards (session_id, card_type, status);

WITH permission_seed(code, module, action, description) AS (
  VALUES
    ('asesmen.read', 'asesmen', 'read', 'Melihat Asesmen Ujian web-first.'),
    ('asesmen.manage', 'asesmen', 'manage', 'Membuat dan mengatur draft Asesmen Ujian web-first.'),
    ('asesmen.cards_issue', 'asesmen', 'cards_issue', 'Menerbitkan Kartu Peserta Ujian dan Lembar Pengawas Ruang.'),
    ('asesmen.proctor_admin', 'asesmen', 'proctor_admin', 'Mengelola pelaksanaan dan pengawasan teknis Asesmen Ujian.'),
    ('asesmen.results_read', 'asesmen', 'results_read', 'Melihat hasil Asesmen Ujian.'),
    ('asesmen.results_manage', 'asesmen', 'results_manage', 'Mengelola hasil Asesmen Ujian.')
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
  VALUES
    ('admin', 'asesmen.read'),
    ('admin', 'asesmen.manage'),
    ('admin', 'asesmen.cards_issue'),
    ('admin', 'asesmen.proctor_admin'),
    ('admin', 'asesmen.results_read'),
    ('admin', 'asesmen.results_manage'),
    ('guru', 'asesmen.read'),
    ('guru', 'asesmen.manage'),
    ('guru', 'asesmen.results_read'),
    ('staf', 'asesmen.read'),
    ('staf', 'asesmen.proctor_admin')
)
INSERT INTO rbac_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM role_permission_seed seed
JOIN rbac_roles r ON r.code = seed.role_code
JOIN rbac_permissions p ON p.code = seed.permission_code
ON CONFLICT (role_id, permission_id) DO NOTHING;
