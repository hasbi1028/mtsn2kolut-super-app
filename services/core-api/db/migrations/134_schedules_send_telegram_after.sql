-- Auto-send Telegram report after this schedule's recap jobs all finish.
-- Per-schedule toggle: admin memilih jadwal mana yang setelah rekap selesai
-- langsung mengirim laporan Telegram otomatis.
ALTER TABLE schedules
ADD COLUMN send_telegram_after BOOLEAN NOT NULL DEFAULT FALSE;
