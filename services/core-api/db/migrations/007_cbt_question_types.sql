-- Migration 007: CBT Question Types + Dynamic Options + Essay Grading

-- question_type: 'multiple_choice' | 'true_false' | 'short_answer' | 'essay' | 'multiple_answer'
-- options: JSONB array — [{"label":"A","text":"..."}, ...]
--   empty array for short_answer and essay
-- answer_key for multiple_answer: comma-separated labels e.g. "A,C"
-- answer_key for essay: empty string — scored via manual_score
ALTER TABLE cbt_questions
  ADD COLUMN question_type TEXT  NOT NULL DEFAULT 'multiple_choice',
  ADD COLUMN options        JSONB NOT NULL DEFAULT '[]';

-- Essay and short_answer grading by guru
ALTER TABLE cbt_student_answers
  ADD COLUMN manual_score NUMERIC(5,2),
  ADD COLUMN graded_by    TEXT,
  ADD COLUMN graded_at    TIMESTAMPTZ;

CREATE INDEX idx_cbt_questions_type ON cbt_questions (question_type);
