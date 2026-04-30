ALTER TABLE employees
  ADD COLUMN IF NOT EXISTS employment_type TEXT NOT NULL DEFAULT 'lainnya';

UPDATE employees e
SET employment_type = CASE
  WHEN EXISTS (
    SELECT 1
    FROM pusaka_accounts pa
    WHERE pa.employee_id = e.id
  ) THEN 'pns'
  ELSE 'lainnya'
END
WHERE employment_type IS NULL
   OR employment_type = '';

ALTER TABLE employees
  DROP CONSTRAINT IF EXISTS employees_employment_type_check;

ALTER TABLE employees
  ADD CONSTRAINT employees_employment_type_check
  CHECK (employment_type IN ('pns', 'pppk', 'honorer', 'lainnya'));
