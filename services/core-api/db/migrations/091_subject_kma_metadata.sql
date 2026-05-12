ALTER TABLE subjects
  ADD COLUMN IF NOT EXISTS counts_for_ranking BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN IF NOT EXISTS is_local_content BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS is_choice_subject BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE subjects
SET category = 'muatan_lokal',
    is_local_content = TRUE,
    counts_for_ranking = COALESCE(counts_for_ranking, TRUE),
    updated_at = NOW()
WHERE is_active = TRUE
  AND (
    LOWER(name) LIKE '%muatan lokal%'
    OR LOWER(name) LIKE '%mulok%'
  );

UPDATE subjects
SET category = 'layanan',
    is_assessment_subject = FALSE,
    is_report_subject = FALSE,
    is_schedule_activity = TRUE,
    counts_for_ranking = FALSE,
    updated_at = NOW()
WHERE is_active = TRUE
  AND (
    LOWER(code) IN ('bk', 'bp')
    OR LOWER(name) LIKE '%bimbingan konseling%'
    OR LOWER(name) LIKE '%bimbingan dan konseling%'
  );

UPDATE subjects
SET category = 'kokurikuler',
    is_assessment_subject = FALSE,
    is_report_subject = TRUE,
    is_schedule_activity = TRUE,
    counts_for_ranking = FALSE,
    updated_at = NOW()
WHERE is_active = TRUE
  AND (
    LOWER(code) IN ('p5', 'p5ra')
    OR LOWER(name) LIKE '%p5ra%'
    OR LOWER(name) LIKE '%projek penguatan%'
    OR LOWER(name) LIKE '%kokurikuler%'
  );
