CREATE TABLE IF NOT EXISTS curriculum_profiles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  regulation_reference TEXT NOT NULL DEFAULT '',
  education_level TEXT NOT NULL DEFAULT 'MTs',
  effective_academic_year_id UUID REFERENCES academic_years(id) ON DELETE SET NULL,
  status TEXT NOT NULL DEFAULT 'draft',
  notes TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_curriculum_profiles_status CHECK (status IN ('draft','active','archived'))
);

CREATE TABLE IF NOT EXISTS curriculum_subject_allocations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  curriculum_profile_id UUID NOT NULL REFERENCES curriculum_profiles(id) ON DELETE CASCADE,
  subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE RESTRICT,
  level TEXT NOT NULL,
  subject_group TEXT NOT NULL DEFAULT 'wajib',
  intra_annual_hours INTEGER NOT NULL DEFAULT 0,
  koku_annual_hours INTEGER NOT NULL DEFAULT 0,
  total_annual_hours INTEGER NOT NULL DEFAULT 0,
  intra_weekly_hours NUMERIC(5,2) NOT NULL DEFAULT 0,
  koku_weekly_hours NUMERIC(5,2) NOT NULL DEFAULT 0,
  total_weekly_hours NUMERIC(5,2) NOT NULL DEFAULT 0,
  lesson_minutes INTEGER NOT NULL DEFAULT 40,
  display_order INTEGER NOT NULL DEFAULT 0,
  counts_for_schedule BOOLEAN NOT NULL DEFAULT TRUE,
  counts_for_report BOOLEAN NOT NULL DEFAULT TRUE,
  counts_for_assessment BOOLEAN NOT NULL DEFAULT TRUE,
  counts_for_ranking BOOLEAN NOT NULL DEFAULT TRUE,
  is_required BOOLEAN NOT NULL DEFAULT TRUE,
  notes TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_curriculum_allocation_level CHECK (level IN ('VII','VIII','IX')),
  CONSTRAINT chk_curriculum_allocation_group CHECK (subject_group IN ('wajib','pilihan','muatan_lokal','layanan','kokurikuler','kegiatan')),
  CONSTRAINT uq_curriculum_allocation_profile_level_subject UNIQUE (curriculum_profile_id, level, subject_id)
);

CREATE TABLE IF NOT EXISTS class_curriculum_assignments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  class_id UUID NOT NULL REFERENCES school_classes(id) ON DELETE CASCADE,
  curriculum_profile_id UUID NOT NULL REFERENCES curriculum_profiles(id) ON DELETE RESTRICT,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  notes TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_class_curriculum_profile UNIQUE (class_id, curriculum_profile_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_class_curriculum_active_one
  ON class_curriculum_assignments (class_id)
  WHERE is_active = TRUE;

CREATE TABLE IF NOT EXISTS report_settings (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  academic_year_id UUID NOT NULL REFERENCES academic_years(id) ON DELETE CASCADE,
  curriculum_profile_id UUID REFERENCES curriculum_profiles(id) ON DELETE SET NULL,
  show_ranking_on_report BOOLEAN NOT NULL DEFAULT FALSE,
  ranking_method TEXT NOT NULL DEFAULT 'intrakurikuler_average',
  ranking_tie_policy TEXT NOT NULL DEFAULT 'same_rank',
  notes TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_report_settings_year UNIQUE (academic_year_id),
  CONSTRAINT chk_report_ranking_method CHECK (ranking_method IN ('intrakurikuler_average','weighted_by_jp','report_subject_average')),
  CONSTRAINT chk_report_tie_policy CHECK (ranking_tie_policy IN ('same_rank','dense_rank','ordinal'))
);
