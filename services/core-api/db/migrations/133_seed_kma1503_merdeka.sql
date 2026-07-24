-- Migration: 133_seed_kma1503_merdeka
-- Seed kurikulum KMA 1503 (Kurikulum Merdeka MTs) + alokasi mapel per tingkat

DO $$
DECLARE
    cur_id UUID;
    sub RECORD;
BEGIN
    -- Cek sudah ada
    SELECT id INTO cur_id FROM curriculum_profiles WHERE code = 'KMA-1503-2025' LIMIT 1;
    
    IF cur_id IS NOT NULL THEN
        RAISE NOTICE 'Kurikulum KMA-1503-2025 already exists, skipping';
        RETURN;
    END IF;

    -- Buat curriculum profile
    INSERT INTO curriculum_profiles (id, code, name, regulation_reference, education_level, effective_academic_year_id, status, notes)
    VALUES (gen_random_uuid(), 'KMA-1503-2025', 'Kurikulum Merdeka MTs', 'KMA 1503 Tahun 2025', 'MTs',
            (SELECT id FROM academic_years WHERE is_active = TRUE LIMIT 1),
            'active', 'Kurikulum Merdeka untuk MTs sesuai KMA 1503/2025')
    RETURNING id INTO cur_id;

    -- Alokasi mapel per tingkat
    -- Format: (subject_code LIKE pattern, level, group, intra_hours, koku_hours, required)

    -- VII
    FOR sub IN
        SELECT * FROM (VALUES
            ('QH',      'VII', 'wajib',      2, 0),
            ('AA',      'VII', 'wajib',      2, 0),
            ('FIKIH',   'VII', 'wajib',      2, 0),
            ('SKI',     'VII', 'wajib',      2, 0),
            ('BARAB',   'VII', 'wajib',      3, 0),
            ('PP',      'VII', 'wajib',      3, 0),
            ('BIND',    'VII', 'wajib',      5, 0),
            ('MTK',     'VII', 'wajib',      5, 0),
            ('IPA',     'VII', 'wajib',      5, 0),
            ('IPS',     'VII', 'wajib',      4, 0),
            ('BING',    'VII', 'wajib',      3, 0),
            ('PJOK',    'VII', 'wajib',      3, 0),
            ('INF',     'VII', 'wajib',      3, 0),
            ('SBD',     'VII', 'wajib',      3, 0),
            ('PRA',     'VII', 'wajib',      2, 0),
            ('MLOKAG',  'VII', 'muatan_lokal', 2, 0),
            ('MLOKWIRA','VII', 'muatan_lokal', 2, 0),
            ('BK',      'VII', 'layanan',    1, 0),
            ('P5RA',    'VII', 'kokurikuler', 0, 4)
        ) AS t(code, lvl, grp, intra, koku)
    LOOP
        INSERT INTO curriculum_subject_allocations (
            curriculum_profile_id, subject_id, level, subject_group,
            intra_weekly_hours, koku_weekly_hours, total_weekly_hours,
            is_required, counts_for_schedule, counts_for_report, counts_for_assessment, counts_for_ranking,
            lesson_minutes, display_order, notes
        )
        SELECT cur_id, s.id, sub.lvl, sub.grp,
            sub.intra, sub.koku, sub.intra + sub.koku,
            true, true, true, true, true,
            40, 0, ''
        FROM subjects s
        WHERE s.code = sub.code
          AND s.is_active = TRUE;
    END LOOP;

    -- VIII
    FOR sub IN
        SELECT * FROM (VALUES
            ('QH',      'VIII', 'wajib',      2, 0),
            ('AA',      'VIII', 'wajib',      2, 0),
            ('FIKIH',   'VIII', 'wajib',      2, 0),
            ('SKI',     'VIII', 'wajib',      2, 0),
            ('BARAB',   'VIII', 'wajib',      3, 0),
            ('PP',      'VIII', 'wajib',      3, 0),
            ('BIND',    'VIII', 'wajib',      5, 0),
            ('MTK',     'VIII', 'wajib',      5, 0),
            ('IPA',     'VIII', 'wajib',      5, 0),
            ('IPS',     'VIII', 'wajib',      4, 0),
            ('BING',    'VIII', 'wajib',      3, 0),
            ('PJOK',    'VIII', 'wajib',      3, 0),
            ('INF',     'VIII', 'wajib',      3, 0),
            ('SBD',     'VIII', 'wajib',      3, 0),
            ('PRA',     'VIII', 'wajib',      2, 0),
            ('MLOKAG',  'VIII', 'muatan_lokal', 2, 0),
            ('MLOKWIRA','VIII', 'muatan_lokal', 2, 0),
            ('BK',      'VIII', 'layanan',    1, 0),
            ('P5RA',    'VIII', 'kokurikuler', 0, 4)
        ) AS t(code, lvl, grp, intra, koku)
    LOOP
        INSERT INTO curriculum_subject_allocations (
            curriculum_profile_id, subject_id, level, subject_group,
            intra_weekly_hours, koku_weekly_hours, total_weekly_hours,
            is_required, counts_for_schedule, counts_for_report, counts_for_assessment, counts_for_ranking,
            lesson_minutes, display_order, notes
        )
        SELECT cur_id, s.id, sub.lvl, sub.grp,
            sub.intra, sub.koku, sub.intra + sub.koku,
            true, true, true, true, true,
            40, 0, ''
        FROM subjects s
        WHERE s.code = sub.code
          AND s.is_active = TRUE;
    END LOOP;

    -- IX
    FOR sub IN
        SELECT * FROM (VALUES
            ('QH',      'IX', 'wajib',      2, 0),
            ('AA',      'IX', 'wajib',      2, 0),
            ('FIKIH',   'IX', 'wajib',      2, 0),
            ('SKI',     'IX', 'wajib',      2, 0),
            ('BARAB',   'IX', 'wajib',      3, 0),
            ('PP',      'IX', 'wajib',      3, 0),
            ('BIND',    'IX', 'wajib',      5, 0),
            ('MTK',     'IX', 'wajib',      5, 0),
            ('IPA',     'IX', 'wajib',      5, 0),
            ('IPS',     'IX', 'wajib',      4, 0),
            ('BING',    'IX', 'wajib',      3, 0),
            ('PJOK',    'IX', 'wajib',      3, 0),
            ('INF',     'IX', 'wajib',      3, 0),
            ('SBD',     'IX', 'wajib',      3, 0),
            ('PRA',     'IX', 'wajib',      2, 0),
            ('MLOKAG',  'IX', 'muatan_lokal', 2, 0),
            ('MLOKWIRA','IX', 'muatan_lokal', 2, 0),
            ('BK',      'IX', 'layanan',    1, 0),
            ('P5RA',    'IX', 'kokurikuler', 0, 4)
        ) AS t(code, lvl, grp, intra, koku)
    LOOP
        INSERT INTO curriculum_subject_allocations (
            curriculum_profile_id, subject_id, level, subject_group,
            intra_weekly_hours, koku_weekly_hours, total_weekly_hours,
            is_required, counts_for_schedule, counts_for_report, counts_for_assessment, counts_for_ranking,
            lesson_minutes, display_order, notes
        )
        SELECT cur_id, s.id, sub.lvl, sub.grp,
            sub.intra, sub.koku, sub.intra + sub.koku,
            true, true, true, true, true,
            40, 0, ''
        FROM subjects s
        WHERE s.code = sub.code
          AND s.is_active = TRUE;
    END LOOP;

    RAISE NOTICE 'Curriculum KMA-1503-2025 + allocations seeded successfully';
END $$;
