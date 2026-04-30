ALTER TABLE cbt_exam_participants
  ADD COLUMN IF NOT EXISTS seat_no INT;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'public'
      AND table_name = 'cbt_exam_participants'
      AND column_name = 'room_id'
  ) THEN
    ALTER TABLE cbt_exam_participants
      ADD COLUMN room_id UUID REFERENCES cbt_exam_rooms(id) ON DELETE SET NULL;
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_cbt_participants_room_seat
  ON cbt_exam_participants (room_id, seat_no)
  WHERE room_id IS NOT NULL;
