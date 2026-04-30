-- not_before: worker will not claim job until this time passes (used for randomized schedule delay)
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS not_before TIMESTAMPTZ;

-- random_window_minutes: scheduler randomizes job execution by 0..N minutes after scheduled time
ALTER TABLE employee_schedules ADD COLUMN IF NOT EXISTS random_window_minutes SMALLINT NOT NULL DEFAULT 0;
