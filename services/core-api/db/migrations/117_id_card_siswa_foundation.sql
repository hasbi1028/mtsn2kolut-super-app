-- Wave 1: ID Card Siswa Terpadu foundation (additive only).

CREATE TABLE IF NOT EXISTS student_id_cards (
  id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  student_id          UUID        NOT NULL REFERENCES students(id) ON DELETE CASCADE,
  card_no             TEXT        NOT NULL UNIQUE,
  token_hash          TEXT        NOT NULL UNIQUE,
  token_hint          TEXT        NOT NULL DEFAULT '',
  status              TEXT        NOT NULL DEFAULT 'active',
  issued_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  printed_at          TIMESTAMPTZ,
  revoked_at          TIMESTAMPTZ,
  revoked_reason      TEXT        NOT NULL DEFAULT '',
  reissued_from_id    UUID        REFERENCES student_id_cards(id) ON DELETE SET NULL,
  created_by_user_id  UUID        REFERENCES users(id) ON DELETE SET NULL,
  updated_by_user_id  UUID        REFERENCES users(id) ON DELETE SET NULL,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_student_id_cards_status CHECK (status IN ('active','inactive','lost','damaged','revoked','replaced','expired','suspended'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_student_id_cards_active_student
  ON student_id_cards(student_id)
  WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_student_id_cards_student ON student_id_cards(student_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_student_id_cards_status ON student_id_cards(status, created_at DESC);

CREATE TABLE IF NOT EXISTS student_id_card_events (
  id                 UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  card_id            UUID        NOT NULL REFERENCES student_id_cards(id) ON DELETE CASCADE,
  student_id         UUID        NOT NULL REFERENCES students(id) ON DELETE CASCADE,
  event_type         TEXT        NOT NULL,
  actor_user_id      UUID        REFERENCES users(id) ON DELETE SET NULL,
  source             TEXT        NOT NULL DEFAULT 'backend',
  metadata           JSONB       NOT NULL DEFAULT '{}'::jsonb,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_student_id_card_events_card ON student_id_card_events(card_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_student_id_card_events_student ON student_id_card_events(student_id, created_at DESC);

CREATE TABLE IF NOT EXISTS student_portal_pin_credentials (
  student_id         UUID        PRIMARY KEY REFERENCES students(id) ON DELETE CASCADE,
  pin_hash           TEXT        NOT NULL,
  is_enabled         BOOLEAN     NOT NULL DEFAULT TRUE,
  failed_attempts    INT         NOT NULL DEFAULT 0,
  locked_until       TIMESTAMPTZ,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS student_card_portal_login_attempts (
  id                 UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  card_id            UUID        REFERENCES student_id_cards(id) ON DELETE SET NULL,
  student_id         UUID        REFERENCES students(id) ON DELETE SET NULL,
  challenge          TEXT        NOT NULL UNIQUE,
  status             TEXT        NOT NULL DEFAULT 'started',
  ip_address         TEXT        NOT NULL DEFAULT '',
  user_agent         TEXT        NOT NULL DEFAULT '',
  expires_at         TIMESTAMPTZ NOT NULL,
  completed_at       TIMESTAMPTZ,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_student_card_portal_login_status CHECK (status IN ('started','completed','failed','expired'))
);

CREATE INDEX IF NOT EXISTS idx_student_card_portal_login_attempts_student ON student_card_portal_login_attempts(student_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_student_card_portal_login_attempts_expires ON student_card_portal_login_attempts(expires_at);
