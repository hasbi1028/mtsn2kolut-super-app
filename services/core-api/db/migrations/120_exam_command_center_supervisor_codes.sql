-- Command Center Ujian: stable operational code for photo-based supervisor roster.
-- This is additive and audit-safe: it does not change existing auth or employment semantics.

ALTER TABLE employees
	ADD COLUMN IF NOT EXISTS exam_supervisor_code integer,
	ADD COLUMN IF NOT EXISTS display_name_source text NOT NULL DEFAULT 'manual',
	ADD COLUMN IF NOT EXISTS name_verified_at timestamptz,
	ADD COLUMN IF NOT EXISTS name_verified_by uuid REFERENCES users(id) ON DELETE SET NULL;

DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1
		FROM pg_constraint
		WHERE conname = 'chk_employees_exam_supervisor_code'
		  AND conrelid = 'employees'::regclass
	) THEN
		ALTER TABLE employees
			ADD CONSTRAINT chk_employees_exam_supervisor_code
			CHECK (exam_supervisor_code IS NULL OR exam_supervisor_code BETWEEN 1 AND 27);
	END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS uq_employees_exam_supervisor_code
	ON employees (exam_supervisor_code)
	WHERE exam_supervisor_code IS NOT NULL;
