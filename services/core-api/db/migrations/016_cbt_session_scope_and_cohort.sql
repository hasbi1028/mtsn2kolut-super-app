ALTER TABLE cbt_exam_sessions
  ADD COLUMN scope_type TEXT NOT NULL DEFAULT 'class',
  ADD COLUMN scope_ref TEXT NOT NULL DEFAULT '',
  ADD COLUMN mix_policy TEXT NOT NULL DEFAULT 'same_grade',
  ADD COLUMN assignment_mode TEXT NOT NULL DEFAULT 'random_balanced',
  ADD COLUMN allow_cross_grade BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN is_special_event BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE cbt_exam_sessions
SET scope_ref = COALESCE(class_id::text, '')
WHERE scope_ref = '' AND class_id IS NOT NULL;

ALTER TABLE cbt_exam_sessions
  ADD CONSTRAINT cbt_exam_sessions_scope_type_chk
    CHECK (scope_type IN ('class', 'grade', 'school', 'custom')),
  ADD CONSTRAINT cbt_exam_sessions_mix_policy_chk
    CHECK (mix_policy IN ('same_class', 'same_grade', 'mixed_scope')),
  ADD CONSTRAINT cbt_exam_sessions_assignment_mode_chk
    CHECK (assignment_mode IN ('manual', 'random_balanced', 'random_by_gender', 'random_by_accommodation')),
  ADD CONSTRAINT cbt_exam_sessions_cross_grade_chk
    CHECK (NOT allow_cross_grade OR is_special_event);

CREATE INDEX idx_cbt_exam_sessions_scope ON cbt_exam_sessions (scope_type, scope_ref, scheduled_start DESC);
