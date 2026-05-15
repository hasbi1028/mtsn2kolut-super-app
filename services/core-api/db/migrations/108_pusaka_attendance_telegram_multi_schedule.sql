ALTER TABLE pusaka_attendance_telegram_settings
  ADD COLUMN IF NOT EXISTS send_times text[] NOT NULL DEFAULT ARRAY['17:00'];

ALTER TABLE pusaka_attendance_telegram_settings
  ADD COLUMN IF NOT EXISTS send_days int[] NOT NULL DEFAULT ARRAY[1,2,3,4,5,6];

UPDATE pusaka_attendance_telegram_settings
SET send_times = ARRAY[to_char(send_time, 'HH24:MI')]
WHERE send_times IS NULL
   OR array_length(send_times, 1) IS NULL
   OR (send_times = ARRAY['17:00'] AND send_time <> '17:00'::time);

UPDATE pusaka_attendance_telegram_settings
SET send_days = ARRAY[1,2,3,4,5,6]
WHERE send_days IS NULL OR array_length(send_days, 1) IS NULL;

ALTER TABLE pusaka_attendance_telegram_settings
  DROP CONSTRAINT IF EXISTS ck_pusaka_attendance_telegram_settings_send_days;

ALTER TABLE pusaka_attendance_telegram_settings
  ADD CONSTRAINT ck_pusaka_attendance_telegram_settings_send_days
  CHECK (send_days <@ ARRAY[0,1,2,3,4,5,6] AND cardinality(send_days) >= 1);

ALTER TABLE pusaka_attendance_telegram_logs
  ADD COLUMN IF NOT EXISTS schedule_time text NOT NULL DEFAULT '';

DROP INDEX IF EXISTS uq_pusaka_attendance_telegram_logs_scheduled_once;

CREATE UNIQUE INDEX IF NOT EXISTS uq_pusaka_attendance_telegram_logs_scheduled_once
  ON pusaka_attendance_telegram_logs (report_date, target_chat_id, send_mode, schedule_time)
  WHERE send_mode = 'scheduled';
