-- Add employee birth date for deterministic account generation.
ALTER TABLE employees
    ADD COLUMN IF NOT EXISTS tanggal_lahir date;

CREATE INDEX IF NOT EXISTS idx_employees_tanggal_lahir ON employees (tanggal_lahir);
