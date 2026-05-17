-- CBT BYOD proctoring policy/action hardening.
-- Additive migration: extend the existing canonical participant event table
-- instead of introducing a second live event source.

ALTER TABLE cbt_participant_events
  ADD COLUMN IF NOT EXISTS severity TEXT NOT NULL DEFAULT 'info',
  ADD COLUMN IF NOT EXISTS category TEXT NOT NULL DEFAULT 'timeline',
  ADD COLUMN IF NOT EXISTS risk_delta INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS dedup_key TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS correlation_id TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS original_event_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS requires_note BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS acknowledged_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS acknowledged_by UUID REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS acknowledge_note TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS actor_username_snapshot TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS actor_employee_id UUID REFERENCES employees(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS request_id TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS source_ip TEXT NOT NULL DEFAULT '';

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_cbt_participant_events_severity'
  ) THEN
    ALTER TABLE cbt_participant_events
      ADD CONSTRAINT chk_cbt_participant_events_severity
      CHECK (severity IN ('info', 'warning', 'medium', 'critical', 'technical')) NOT VALID;
  END IF;
END $$;

ALTER TABLE cbt_participant_events
  VALIDATE CONSTRAINT chk_cbt_participant_events_severity;

ALTER TABLE cbt_exam_participants
  ADD COLUMN IF NOT EXISTS last_local_save_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS last_synced_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS pending_answer_count INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS sync_state TEXT NOT NULL DEFAULT 'unknown';

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_cbt_exam_participants_sync_state'
  ) THEN
    ALTER TABLE cbt_exam_participants
      ADD CONSTRAINT chk_cbt_exam_participants_sync_state
      CHECK (sync_state IN ('synced', 'pending', 'failed', 'unknown')) NOT VALID;
  END IF;
END $$;

ALTER TABLE cbt_exam_participants
  VALIDATE CONSTRAINT chk_cbt_exam_participants_sync_state;

CREATE TABLE IF NOT EXISTS cbt_proctor_actions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id UUID NOT NULL REFERENCES cbt_exam_sessions(id) ON DELETE CASCADE,
  room_id UUID REFERENCES cbt_exam_rooms(id) ON DELETE SET NULL,
  participant_id UUID REFERENCES cbt_exam_participants(id) ON DELETE SET NULL,
  event_id UUID REFERENCES cbt_participant_events(id) ON DELETE SET NULL,
  action_type TEXT NOT NULL,
  reason TEXT NOT NULL,
  notes TEXT NOT NULL DEFAULT '',
  actor_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  actor_username_snapshot TEXT NOT NULL DEFAULT '',
  actor_employee_id UUID REFERENCES employees(id) ON DELETE SET NULL,
  request_id TEXT NOT NULL DEFAULT '',
  source_ip TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_cbt_proctor_actions_action_type'
  ) THEN
    ALTER TABLE cbt_proctor_actions
      ADD CONSTRAINT chk_cbt_proctor_actions_action_type
      CHECK (action_type IN (
        'warn_student',
        'hold_access',
        'unlock_access',
        'reset_device_binding',
        'force_submit',
        'mark_technical_issue',
        'mark_incident',
        'escalate_to_committee',
        'clear_after_check'
      )) NOT VALID;
  END IF;
END $$;

ALTER TABLE cbt_proctor_actions
  VALIDATE CONSTRAINT chk_cbt_proctor_actions_action_type;

CREATE INDEX IF NOT EXISTS idx_cbt_participant_events_session_created
  ON cbt_participant_events ((event_data->>'session_id'), created_at DESC);

CREATE INDEX IF NOT EXISTS idx_cbt_participant_events_ack
  ON cbt_participant_events (severity, acknowledged_at, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_cbt_participant_events_dedup
  ON cbt_participant_events (participant_id, dedup_key, created_at DESC)
  WHERE dedup_key <> '';

CREATE INDEX IF NOT EXISTS idx_cbt_exam_participants_sync
  ON cbt_exam_participants (session_id, sync_state, pending_answer_count);

CREATE INDEX IF NOT EXISTS idx_cbt_proctor_actions_session_created
  ON cbt_proctor_actions (session_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_cbt_proctor_actions_room_created
  ON cbt_proctor_actions (room_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_cbt_proctor_actions_participant_created
  ON cbt_proctor_actions (participant_id, created_at DESC);
