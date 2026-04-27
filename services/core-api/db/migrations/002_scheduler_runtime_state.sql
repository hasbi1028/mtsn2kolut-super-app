ALTER TABLE schedules
ADD COLUMN IF NOT EXISTS last_enqueued_for_date DATE;
