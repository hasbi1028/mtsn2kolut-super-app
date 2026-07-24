-- Migration: 132_seed_2026_2027
-- Tambah tahun pelajaran 2026/2027 + 2 semester

DO $$
DECLARE
    year_id UUID;
BEGIN
    -- Skip if already exists
    SELECT id INTO year_id FROM academic_years WHERE name = '2026/2027' LIMIT 1;
    
    IF year_id IS NULL THEN
        INSERT INTO academic_years (id, name, start_date, end_date, is_active)
        VALUES (gen_random_uuid(), '2026/2027', DATE '2026-07-01', DATE '2027-06-30', FALSE)
        RETURNING id INTO year_id;

        INSERT INTO semesters (academic_year_id, name, label, start_date, end_date, is_active)
        VALUES
            (year_id, 'Ganjil', 'Semester Ganjil 2026/2027', DATE '2026-07-01', DATE '2026-12-31', FALSE),
            (year_id, 'Genap',  'Semester Genap 2026/2027',  DATE '2027-01-01', DATE '2027-06-30', FALSE);
        
        RAISE NOTICE 'Seed 2026/2027 added';
    ELSE
        RAISE NOTICE 'Seed 2026/2027 already exists, skipped';
    END IF;
END $$;
