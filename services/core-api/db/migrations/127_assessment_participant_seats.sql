-- Add deterministic seat placement for web-first assessment participants.
-- Additive only; existing rows remain valid until room assignment is applied.

ALTER TABLE assessment_participants
  ADD COLUMN IF NOT EXISTS seat_no integer CHECK (seat_no IS NULL OR seat_no > 0);

CREATE UNIQUE INDEX IF NOT EXISTS idx_assessment_participants_room_seat_unique
  ON assessment_participants (session_id, room_id, seat_no)
  WHERE room_id IS NOT NULL AND seat_no IS NOT NULL;
