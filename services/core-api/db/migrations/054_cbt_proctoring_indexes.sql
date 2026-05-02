CREATE INDEX IF NOT EXISTS idx_cbt_participants_session_id
ON cbt_exam_participants (session_id, id);
