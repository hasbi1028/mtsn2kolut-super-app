-- Drop old UNIQUE(run_type) constraint so we can have multiple times per type
ALTER TABLE schedules DROP CONSTRAINT schedules_run_type_key;
ALTER TABLE schedules ADD CONSTRAINT uq_schedules_type_time UNIQUE (run_type, run_time);

-- Per-employee checkin/checkout schedule table
CREATE TABLE employee_schedules (
  id                     UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id            UUID          NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  run_type               run_type_enum NOT NULL,
  run_time               TEXT          NOT NULL,
  is_enabled             BOOLEAN       NOT NULL DEFAULT TRUE,
  last_enqueued_for_date DATE,
  created_at             TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  updated_at             TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_emp_sched_type CHECK (run_type IN ('checkin', 'checkout')),
  CONSTRAINT uq_emp_sched_type  UNIQUE (employee_id, run_type)
);

CREATE INDEX idx_employee_schedules_employee_id ON employee_schedules (employee_id);
