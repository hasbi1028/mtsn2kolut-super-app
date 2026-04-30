-- Keep existing schedules deterministic unless an operator explicitly enables a delay window.
ALTER TABLE employee_schedules
  ALTER COLUMN random_window_minutes SET DEFAULT 0;

-- Repair rows that may have inherited the old rollout default unintentionally.
UPDATE employee_schedules
SET random_window_minutes = 0
WHERE random_window_minutes = 15;
