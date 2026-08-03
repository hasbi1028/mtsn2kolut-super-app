-- Izinkan send_mode 'after_rekap' (kirim otomatis oleh scheduler setelah
-- semua job rekap jadwal dengan flag send_telegram_after selesai).
ALTER TABLE pusaka_attendance_telegram_logs
DROP CONSTRAINT ck_pusaka_attendance_telegram_logs_send_mode;

ALTER TABLE pusaka_attendance_telegram_logs
ADD CONSTRAINT ck_pusaka_attendance_telegram_logs_send_mode
CHECK (send_mode = ANY (ARRAY['manual', 'scheduled', 'after_rekap']));
