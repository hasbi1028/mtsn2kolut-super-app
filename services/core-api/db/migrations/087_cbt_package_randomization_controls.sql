ALTER TABLE cbt_packages
  ADD COLUMN IF NOT EXISTS source_mode TEXT NOT NULL DEFAULT 'teacher_class' CHECK (source_mode IN ('teacher_class', 'level_subject_teachers', 'event_pool')),
  ADD COLUMN IF NOT EXISTS randomize_options BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS draw_pg_count INTEGER NOT NULL DEFAULT 0 CHECK (draw_pg_count >= 0),
  ADD COLUMN IF NOT EXISTS draw_essay_count INTEGER NOT NULL DEFAULT 0 CHECK (draw_essay_count >= 0),
  ADD COLUMN IF NOT EXISTS random_seed TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS composition_log JSONB NOT NULL DEFAULT '{}'::jsonb;

CREATE INDEX IF NOT EXISTS idx_cbt_packages_event_source_mode
  ON cbt_packages (event_id, source_mode, subject_id);
