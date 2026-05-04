-- Migration: 061_cbt_participant_seat_invariants
--
-- Seat numbers are optional while a participant is unseated, but once present
-- they must be positive and unique within an assigned CBT room. Stop loudly on
-- legacy bad data so operators can repair the affected seating plan first.

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM cbt_exam_participants
    WHERE seat_no IS NOT NULL
      AND seat_no <= 0
  ) THEN
    RAISE EXCEPTION 'invalid CBT participant seat numbers exist; clear or repair non-positive seat_no values before applying seat invariants';
  END IF;

  IF EXISTS (
    SELECT 1
    FROM cbt_exam_participants
    WHERE room_id IS NOT NULL
      AND seat_no IS NOT NULL
    GROUP BY room_id, seat_no
    HAVING COUNT(*) > 1
  ) THEN
    RAISE EXCEPTION 'duplicate CBT participant seats exist; each (room_id, seat_no) must be unique before applying seat invariants';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'chk_cbt_participants_seat_no_positive'
      AND conrelid = 'cbt_exam_participants'::regclass
  ) THEN
    ALTER TABLE cbt_exam_participants
      ADD CONSTRAINT chk_cbt_participants_seat_no_positive
      CHECK (seat_no IS NULL OR seat_no > 0) NOT VALID;
  END IF;
END $$;

ALTER TABLE cbt_exam_participants
  VALIDATE CONSTRAINT chk_cbt_participants_seat_no_positive;

CREATE UNIQUE INDEX IF NOT EXISTS idx_cbt_participants_room_seat_unique
  ON cbt_exam_participants (room_id, seat_no)
  WHERE room_id IS NOT NULL
    AND seat_no IS NOT NULL;
