ALTER TABLE cbt_exam_events
  ADD COLUMN target_levels TEXT[] NOT NULL DEFAULT '{}';

CREATE INDEX idx_cbt_events_target_levels
  ON cbt_exam_events USING GIN (target_levels);
