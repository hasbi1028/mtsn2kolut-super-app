-- day_of_week: 0=Minggu, 1=Senin, 2=Selasa, 3=Rabu, 4=Kamis, 5=Jumat, 6=Sabtu
ALTER TABLE employee_schedules
  ADD COLUMN IF NOT EXISTS day_of_week SMALLINT NOT NULL DEFAULT 1;

ALTER TABLE employee_schedules
  ADD CONSTRAINT chk_day_of_week CHECK (day_of_week BETWEEN 0 AND 6);

ALTER TABLE employee_schedules DROP CONSTRAINT uq_emp_sched_type;

ALTER TABLE employee_schedules
  ADD CONSTRAINT uq_emp_sched_type UNIQUE (employee_id, run_type, day_of_week);
