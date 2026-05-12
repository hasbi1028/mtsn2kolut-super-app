CREATE TABLE IF NOT EXISTS lesson_period_templates (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  academic_year_id UUID NOT NULL REFERENCES academic_years(id) ON DELETE CASCADE,
  day_of_week SMALLINT NOT NULL CHECK (day_of_week BETWEEN 1 AND 6),
  period_number INTEGER NOT NULL,
  start_time TIME NOT NULL,
  end_time TIME NOT NULL,
  activity_type TEXT NOT NULL DEFAULT 'pelajaran',
  label TEXT NOT NULL DEFAULT '',
  is_counted_as_lesson BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT ck_lesson_period_time CHECK (end_time > start_time),
  CONSTRAINT chk_lesson_period_activity CHECK (activity_type IN ('pelajaran','istirahat','upacara','pembiasaan','kokurikuler','ekstrakurikuler','lainnya')),
  CONSTRAINT chk_lesson_period_number_positive CHECK (period_number > 0),
  CONSTRAINT uq_lesson_period_template UNIQUE (academic_year_id, day_of_week, period_number)
);

CREATE INDEX IF NOT EXISTS idx_lesson_period_templates_year_day
  ON lesson_period_templates(academic_year_id, day_of_week, period_number);

ALTER TABLE timetable_slots
  ADD COLUMN IF NOT EXISTS lesson_period_id UUID REFERENCES lesson_period_templates(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS slot_type TEXT NOT NULL DEFAULT 'pelajaran',
  ADD COLUMN IF NOT EXISTS lesson_hours NUMERIC(5,2) NOT NULL DEFAULT 1;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_timetable_slot_type'
  ) THEN
    ALTER TABLE timetable_slots
      ADD CONSTRAINT chk_timetable_slot_type
      CHECK (slot_type IN ('pelajaran','istirahat','upacara','pembiasaan','kokurikuler','ekstrakurikuler','lainnya'));
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_timetable_lesson_hours_range'
  ) THEN
    ALTER TABLE timetable_slots
      ADD CONSTRAINT chk_timetable_lesson_hours_range
      CHECK (lesson_hours >= 0 AND lesson_hours <= 12);
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_timetable_slots_lesson_period
  ON timetable_slots(lesson_period_id);
