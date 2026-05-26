CREATE TABLE IF NOT EXISTS cbt_exam_access_cards (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  card_type TEXT NOT NULL,
  event_id UUID NOT NULL REFERENCES cbt_exam_events(id) ON DELETE CASCADE,
  session_id UUID NOT NULL REFERENCES cbt_exam_sessions(id) ON DELETE CASCADE,
  room_id UUID REFERENCES cbt_exam_rooms(id) ON DELETE CASCADE,
  participant_id UUID REFERENCES cbt_exam_participants(id) ON DELETE CASCADE,
  room_proctor_id UUID REFERENCES cbt_room_proctors(id) ON DELETE CASCADE,
  assigned_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  token_hash TEXT NOT NULL,
  pin_hash TEXT NOT NULL,
  token_hash_version INTEGER NOT NULL DEFAULT 1,
  pin_hash_version INTEGER NOT NULL DEFAULT 1,
  status TEXT NOT NULL DEFAULT 'active',
  failed_attempts INTEGER NOT NULL DEFAULT 0,
  max_failed_attempts INTEGER NOT NULL DEFAULT 5,
  last_failed_at TIMESTAMPTZ,
  verified_at TIMESTAMPTZ,
  expires_at TIMESTAMPTZ,
  revoked_at TIMESTAMPTZ,
  generated_by UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT chk_cbt_exam_access_cards_type CHECK (card_type IN ('participant', 'proctor')),
  CONSTRAINT chk_cbt_exam_access_cards_status CHECK (status IN ('active', 'locked', 'revoked', 'expired')),
  CONSTRAINT chk_cbt_exam_access_cards_target CHECK (
    (card_type = 'participant' AND participant_id IS NOT NULL AND room_proctor_id IS NULL)
    OR (card_type = 'proctor' AND room_proctor_id IS NOT NULL AND participant_id IS NULL AND room_id IS NOT NULL)
  ),
  CONSTRAINT chk_cbt_exam_access_cards_attempts CHECK (failed_attempts >= 0 AND max_failed_attempts > 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_cbt_exam_access_cards_token_active
  ON cbt_exam_access_cards (token_hash)
  WHERE revoked_at IS NULL AND status IN ('active', 'locked');

CREATE UNIQUE INDEX IF NOT EXISTS idx_cbt_exam_access_cards_participant_active
  ON cbt_exam_access_cards (participant_id)
  WHERE card_type = 'participant' AND revoked_at IS NULL AND status = 'active';

CREATE UNIQUE INDEX IF NOT EXISTS idx_cbt_exam_access_cards_proctor_active
  ON cbt_exam_access_cards (room_proctor_id)
  WHERE card_type = 'proctor' AND revoked_at IS NULL AND status = 'active';

CREATE INDEX IF NOT EXISTS idx_cbt_exam_access_cards_event_type
  ON cbt_exam_access_cards (event_id, card_type, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_cbt_exam_access_cards_session_room
  ON cbt_exam_access_cards (session_id, room_id, card_type);
