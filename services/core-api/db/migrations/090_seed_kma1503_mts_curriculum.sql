WITH active_year AS (
  SELECT id
  FROM academic_years
  WHERE is_active = TRUE
  ORDER BY start_date DESC, name DESC
  LIMIT 1
), profile AS (
  INSERT INTO curriculum_profiles (
    code,
    name,
    regulation_reference,
    education_level,
    effective_academic_year_id,
    status,
    notes
  )
  VALUES (
    'KMA1503-2025-MTS',
    'Kurikulum Merdeka MTs KMA 1503 Tahun 2025',
    'KMA 1503 Tahun 2025',
    'MTs',
    (SELECT id FROM active_year),
    'active',
    'Profil awal untuk verifikasi operator. Alokasi wajib MTs VII/VIII/IX mengikuti review dokumen KMA 1503 Tahun 2025; mapel pilihan dan muatan lokal disiapkan bertahap.'
  )
  ON CONFLICT (code) DO UPDATE
  SET name = EXCLUDED.name,
      regulation_reference = EXCLUDED.regulation_reference,
      education_level = EXCLUDED.education_level,
      effective_academic_year_id = COALESCE(curriculum_profiles.effective_academic_year_id, EXCLUDED.effective_academic_year_id),
      status = CASE WHEN curriculum_profiles.status = 'archived' THEN curriculum_profiles.status ELSE EXCLUDED.status END,
      notes = EXCLUDED.notes,
      updated_at = NOW()
  RETURNING id
), allocation_source(level, subject_code, subject_group, intra_annual_hours, koku_annual_hours, total_annual_hours, intra_weekly_hours, koku_weekly_hours, total_weekly_hours, display_order, notes) AS (
  VALUES
  ('VII','QH','wajib',72,36,108,2.00,1.00,3.00,10,''),
  ('VII','AA','wajib',72,0,72,2.00,0.00,2.00,20,''),
  ('VII','FIKIH','wajib',72,0,72,2.00,0.00,2.00,30,''),
  ('VII','SKI','wajib',72,0,72,2.00,0.00,2.00,40,''),
  ('VII','BARAB','wajib',108,0,108,3.00,0.00,3.00,50,''),
  ('VII','PP','wajib',72,36,108,2.00,1.00,3.00,60,''),
  ('VII','BIND','wajib',180,36,216,5.00,1.00,6.00,70,''),
  ('VII','MTK','wajib',144,0,144,4.00,0.00,4.00,80,''),
  ('VII','IPA','wajib',144,0,144,4.00,0.00,4.00,90,''),
  ('VII','IPS','wajib',108,0,108,3.00,0.00,3.00,100,''),
  ('VII','BING','wajib',108,0,108,3.00,0.00,3.00,110,''),
  ('VII','PJOK','wajib',72,36,108,2.00,1.00,3.00,120,''),
  ('VII','INF','wajib',72,0,72,2.00,0.00,2.00,130,''),
  ('VII','SBD','wajib',72,0,72,2.00,0.00,2.00,140,'Menggunakan master mapel Seni Budaya untuk struktur Seni, Budaya, dan Prakarya.'),

  ('VIII','QH','wajib',72,0,72,2.00,0.00,2.00,10,''),
  ('VIII','AA','wajib',72,0,72,2.00,0.00,2.00,20,''),
  ('VIII','FIKIH','wajib',72,36,108,2.00,1.00,3.00,30,''),
  ('VIII','SKI','wajib',72,0,72,2.00,0.00,2.00,40,''),
  ('VIII','BARAB','wajib',108,0,108,3.00,0.00,3.00,50,''),
  ('VIII','PP','wajib',72,0,72,2.00,0.00,2.00,60,''),
  ('VIII','BIND','wajib',180,0,180,5.00,0.00,5.00,70,''),
  ('VIII','MTK','wajib',144,36,180,4.00,1.00,5.00,80,''),
  ('VIII','IPA','wajib',144,36,180,4.00,1.00,5.00,90,''),
  ('VIII','IPS','wajib',108,0,108,3.00,0.00,3.00,100,''),
  ('VIII','BING','wajib',108,0,108,3.00,0.00,3.00,110,''),
  ('VIII','PJOK','wajib',72,0,72,2.00,0.00,2.00,120,''),
  ('VIII','INF','wajib',72,0,72,2.00,0.00,2.00,130,''),
  ('VIII','SBD','wajib',72,36,108,2.00,1.00,3.00,140,'Menggunakan master mapel Seni Budaya untuk struktur Seni, Budaya, dan Prakarya.'),

  ('IX','QH','wajib',64,0,64,2.00,0.00,2.00,10,''),
  ('IX','AA','wajib',64,32,96,2.00,1.00,3.00,20,'Baris Akidah Akhlak IX pada hasil baca dokumen perlu verifikasi manual; nilai ini menjaga total wajib 1.344 JP/tahun.'),
  ('IX','FIKIH','wajib',64,0,64,2.00,0.00,2.00,30,''),
  ('IX','SKI','wajib',64,0,64,2.00,0.00,2.00,40,''),
  ('IX','BARAB','wajib',96,0,96,3.00,0.00,3.00,50,''),
  ('IX','PP','wajib',64,0,64,2.00,0.00,2.00,60,''),
  ('IX','BIND','wajib',160,0,160,5.00,0.00,5.00,70,''),
  ('IX','MTK','wajib',128,0,128,4.00,0.00,4.00,80,''),
  ('IX','IPA','wajib',128,0,128,4.00,0.00,4.00,90,''),
  ('IX','IPS','wajib',96,32,128,3.00,1.00,4.00,100,''),
  ('IX','BING','wajib',96,32,128,3.00,1.00,4.00,110,''),
  ('IX','PJOK','wajib',64,0,64,2.00,0.00,2.00,120,''),
  ('IX','INF','wajib',64,32,96,2.00,1.00,3.00,130,''),
  ('IX','SBD','wajib',64,0,64,2.00,0.00,2.00,140,'Menggunakan master mapel Seni Budaya untuk struktur Seni, Budaya, dan Prakarya.')
), upsert_allocations AS (
  INSERT INTO curriculum_subject_allocations (
    curriculum_profile_id,
    subject_id,
    level,
    subject_group,
    intra_annual_hours,
    koku_annual_hours,
    total_annual_hours,
    intra_weekly_hours,
    koku_weekly_hours,
    total_weekly_hours,
    lesson_minutes,
    display_order,
    counts_for_schedule,
    counts_for_report,
    counts_for_assessment,
    counts_for_ranking,
    is_required,
    notes
  )
  SELECT
    p.id,
    s.id,
    a.level,
    a.subject_group,
    a.intra_annual_hours,
    a.koku_annual_hours,
    a.total_annual_hours,
    a.intra_weekly_hours,
    a.koku_weekly_hours,
    a.total_weekly_hours,
    40,
    a.display_order,
    TRUE,
    TRUE,
    TRUE,
    TRUE,
    TRUE,
    a.notes
  FROM allocation_source a
  CROSS JOIN profile p
  JOIN subjects s ON s.code = a.subject_code
  ON CONFLICT (curriculum_profile_id, level, subject_id) DO UPDATE
  SET subject_group = EXCLUDED.subject_group,
      intra_annual_hours = EXCLUDED.intra_annual_hours,
      koku_annual_hours = EXCLUDED.koku_annual_hours,
      total_annual_hours = EXCLUDED.total_annual_hours,
      intra_weekly_hours = EXCLUDED.intra_weekly_hours,
      koku_weekly_hours = EXCLUDED.koku_weekly_hours,
      total_weekly_hours = EXCLUDED.total_weekly_hours,
      lesson_minutes = EXCLUDED.lesson_minutes,
      display_order = EXCLUDED.display_order,
      counts_for_schedule = EXCLUDED.counts_for_schedule,
      counts_for_report = EXCLUDED.counts_for_report,
      counts_for_assessment = EXCLUDED.counts_for_assessment,
      counts_for_ranking = EXCLUDED.counts_for_ranking,
      is_required = EXCLUDED.is_required,
      notes = EXCLUDED.notes,
      updated_at = NOW()
  RETURNING id
), class_assignments AS (
  INSERT INTO class_curriculum_assignments (class_id, curriculum_profile_id, is_active, notes)
  SELECT c.id, p.id, TRUE, 'Profil kurikulum awal mengikuti KMA 1503 Tahun 2025.'
  FROM school_classes c
  JOIN active_year ay ON ay.id = c.academic_year_id
  CROSS JOIN profile p
  WHERE c.is_active = TRUE
    AND c.level IN ('VII','VIII','IX')
  ON CONFLICT (class_id, curriculum_profile_id) DO UPDATE
  SET is_active = TRUE,
      notes = EXCLUDED.notes,
      updated_at = NOW()
  RETURNING id
)
INSERT INTO report_settings (academic_year_id, curriculum_profile_id, show_ranking_on_report, notes)
SELECT ay.id, p.id, FALSE, 'Peringkat tidak ditampilkan pada rapor resmi secara bawaan.'
FROM active_year ay
CROSS JOIN profile p
ON CONFLICT (academic_year_id) DO UPDATE
SET curriculum_profile_id = EXCLUDED.curriculum_profile_id,
    show_ranking_on_report = FALSE,
    notes = EXCLUDED.notes,
    updated_at = NOW();
