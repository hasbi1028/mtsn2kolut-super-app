-- Add performance indexes for frequently queried columns

-- Index for participant lookup by session and student (prevents duplicate enrollments)
CREATE INDEX IF NOT EXISTS idx_cbt_participants_session_student ON cbt_exam_participants(session_id, student_id);

-- Index for student answers lookup by participant (for scoring)
CREATE INDEX IF NOT EXISTS idx_cbt_student_answers_participant ON cbt_student_answers(participant_id);

-- Index for participant events lookup by participant and time (for audit trails)
CREATE INDEX IF NOT EXISTS idx_cbt_participant_events_participant_time ON cbt_participant_events(participant_id, created_at DESC);

-- Index for CBT questions by package (for loading question banks)
CREATE INDEX IF NOT EXISTS idx_cbt_questions_package ON cbt_questions(package_id);

-- Index for CBT packages by subject (for filtering)
CREATE INDEX IF NOT EXISTS idx_cbt_packages_subject ON cbt_packages(subject_code);

-- Index for exam sessions by status and scheduled time (for dashboard views)
CREATE INDEX IF NOT EXISTS idx_cbt_sessions_status_time ON cbt_exam_sessions(status, scheduled_start);

-- Index for exam rooms by session (for room management)
CREATE INDEX IF NOT EXISTS idx_cbt_rooms_session ON cbt_exam_rooms(session_id);

-- Index for proctoring queries (active participants in a session)
CREATE INDEX IF NOT EXISTS idx_cbt_participants_session_active ON cbt_exam_participants(session_id) 
WHERE submitted_at IS NULL AND suspicious_flag = FALSE;