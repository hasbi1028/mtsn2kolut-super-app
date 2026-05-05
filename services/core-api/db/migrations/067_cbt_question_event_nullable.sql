-- Bank Soal must support reusable questions that are not tied to a CBT event.
-- Migration 062 creates this column nullable for new databases; this defensive
-- migration also fixes any database where the column existed with NOT NULL.
ALTER TABLE cbt_questions
  ALTER COLUMN event_id DROP NOT NULL;
