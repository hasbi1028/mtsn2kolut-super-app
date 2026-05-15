CREATE TABLE IF NOT EXISTS pusaka_attendance_telegram_settings (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  settings_key text NOT NULL DEFAULT 'default' UNIQUE,
  is_enabled boolean NOT NULL DEFAULT false,
  send_time time NOT NULL DEFAULT '17:00'::time,
  timezone text NOT NULL DEFAULT 'Asia/Makassar',
  target_chat_id text NOT NULL DEFAULT '',
  include_caption boolean NOT NULL DEFAULT true,
  include_image boolean NOT NULL DEFAULT true,
  report_mode text NOT NULL DEFAULT 'ringkas',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT ck_pusaka_attendance_telegram_settings_singleton CHECK (settings_key = 'default'),
  CONSTRAINT ck_pusaka_attendance_telegram_settings_timezone CHECK (timezone IN ('Asia/Makassar')),
  CONSTRAINT ck_pusaka_attendance_telegram_settings_mode CHECK (report_mode IN ('ringkas')),
  CONSTRAINT ck_pusaka_attendance_telegram_settings_content CHECK (include_caption OR include_image)
);

CREATE TABLE IF NOT EXISTS pusaka_attendance_telegram_logs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  report_date date NOT NULL,
  target_chat_id text NOT NULL,
  send_mode text NOT NULL,
  status text NOT NULL,
  telegram_message_id text,
  error_message text,
  requested_by uuid REFERENCES users(id) ON DELETE SET NULL,
  sent_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT ck_pusaka_attendance_telegram_logs_send_mode CHECK (send_mode IN ('manual', 'scheduled')),
  CONSTRAINT ck_pusaka_attendance_telegram_logs_status CHECK (status IN ('success', 'failed'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_pusaka_attendance_telegram_logs_scheduled_once
  ON pusaka_attendance_telegram_logs (report_date, target_chat_id, send_mode)
  WHERE send_mode = 'scheduled';

CREATE INDEX IF NOT EXISTS idx_pusaka_attendance_telegram_logs_sent_at
  ON pusaka_attendance_telegram_logs (sent_at DESC);
