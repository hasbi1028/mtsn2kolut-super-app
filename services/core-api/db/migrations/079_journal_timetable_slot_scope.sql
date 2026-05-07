-- Migration 079: Jurnal class session slot linkage and all-scope journal permissions.

ALTER TABLE class_journal_sessions
ADD COLUMN IF NOT EXISTS timetable_slot_id UUID REFERENCES timetable_slots(id) ON DELETE SET NULL;

-- Slot-based journals need their own uniqueness. The legacy unique constraint
-- on (assignment_id, tanggal) would collapse two different slots of the same
-- subject on the same date into one journal session, so keep that rule only for
-- manual/non-slot sessions.
ALTER TABLE class_journal_sessions
DROP CONSTRAINT IF EXISTS uq_journal_session_date;

CREATE UNIQUE INDEX IF NOT EXISTS uq_journal_session_manual_assignment_date
    ON class_journal_sessions (assignment_id, tanggal)
    WHERE timetable_slot_id IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_journal_session_timetable_slot_date
    ON class_journal_sessions (timetable_slot_id, tanggal)
    WHERE timetable_slot_id IS NOT NULL;

WITH permission_seed(code, module, action, description) AS (
    VALUES
        ('journal.read_all', 'journal', 'read_all', 'Melihat seluruh jurnal kelas lintas guru.'),
        ('journal.manage_all', 'journal', 'manage_all', 'Mengelola seluruh jurnal kelas lintas guru.')
)
INSERT INTO rbac_permissions (code, module, action, description)
SELECT code, module, action, description FROM permission_seed
ON CONFLICT (code) DO UPDATE
SET module = EXCLUDED.module,
    action = EXCLUDED.action,
    description = EXCLUDED.description,
    is_active = TRUE,
    updated_at = NOW();

INSERT INTO rbac_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM rbac_roles r
JOIN rbac_permissions p ON p.code IN ('journal.read_all', 'journal.manage_all')
WHERE r.code = 'admin'
ON CONFLICT (role_id, permission_id) DO NOTHING;
