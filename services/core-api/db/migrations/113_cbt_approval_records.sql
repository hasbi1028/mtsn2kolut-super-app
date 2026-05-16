CREATE TABLE IF NOT EXISTS cbt_approval_records (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  entity_type TEXT NOT NULL,
  entity_id UUID NOT NULL,
  approval_type TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'approved',
  approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
  approved_at TIMESTAMPTZ,
  revoked_by UUID REFERENCES users(id) ON DELETE SET NULL,
  revoked_at TIMESTAMPTZ,
  notes TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_cbt_approval_records_entity_type
    CHECK (entity_type = ANY (ARRAY['event', 'session', 'package', 'result']::TEXT[])),
  CONSTRAINT chk_cbt_approval_records_status
    CHECK (status = ANY (ARRAY['approved', 'revoked']::TEXT[])),
  CONSTRAINT chk_cbt_approval_records_approval_type
    CHECK (approval_type = ANY (ARRAY[
      'package_ready',
      'participants_rooms_ready',
      'tokens_cards_ready',
      'results_verified',
      'final_archive',
      'session_minutes',
      'room_handover'
    ]::TEXT[]))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_cbt_approval_records_active
  ON cbt_approval_records (entity_type, entity_id, approval_type)
  WHERE status = 'approved';

CREATE INDEX IF NOT EXISTS idx_cbt_approval_records_entity
  ON cbt_approval_records (entity_type, entity_id, approval_type, status);

CREATE INDEX IF NOT EXISTS idx_cbt_approval_records_approved_at
  ON cbt_approval_records (approved_at DESC);
