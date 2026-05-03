-- Migration 055: CBT dynamic answer keys
--
-- The original CBT foundation table limited answer_key to A-E. The question
-- bank now stores dynamic options in JSONB, supports up to six composer options,
-- and also supports non-label answer keys for short-answer items.
ALTER TABLE cbt_questions
  DROP CONSTRAINT IF EXISTS cbt_questions_answer_key_check;
