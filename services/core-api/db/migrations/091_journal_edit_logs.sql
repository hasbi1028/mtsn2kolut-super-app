-- Journal Edit Logs — audit trail for session edits
CREATE TABLE IF NOT EXISTS journal_edit_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id  UUID NOT NULL REFERENCES class_journal_sessions(id) ON DELETE CASCADE,
    edited_by   UUID NOT NULL REFERENCES employees(id),
    edited_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    changes     JSONB NOT NULL DEFAULT '[]'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_journal_edit_logs_session_id ON journal_edit_logs(session_id);
CREATE INDEX IF NOT EXISTS idx_journal_edit_logs_edited_at ON journal_edit_logs(edited_at);
