-- Migration: 064_seed_academic_2025_2026
-- Seed tahun pelajaran, kelas, dan mata pelajaran MTs Kurikulum Merdeka.

UPDATE academic_years
SET is_active = FALSE,
    updated_at = NOW()
WHERE is_active = TRUE
  AND name <> '2025/2026';

INSERT INTO academic_years (name, start_date, end_date, is_active)
VALUES ('2025/2026', DATE '2025-07-01', DATE '2026-06-30', TRUE)
ON CONFLICT (name) DO UPDATE
SET start_date = EXCLUDED.start_date,
    end_date = EXCLUDED.end_date,
    is_active = TRUE,
    updated_at = NOW();

WITH target_year AS (
  SELECT id
  FROM academic_years
  WHERE name = '2025/2026'
)
INSERT INTO school_classes (academic_year_id, code, name, level, is_active)
SELECT target_year.id, seed.code, seed.name, seed.level, TRUE
FROM target_year
CROSS JOIN (
  VALUES
    ('VII.A', 'VII.A', 'VII'),
    ('VII.B', 'VII.B', 'VII'),
    ('VII.C', 'VII.C', 'VII'),
    ('VII.D', 'VII.D', 'VII'),
    ('VIII.A', 'VIII.A', 'VIII'),
    ('VIII.B', 'VIII.B', 'VIII'),
    ('VIII.C', 'VIII.C', 'VIII'),
    ('IX.A', 'IX.A', 'IX'),
    ('IX.B', 'IX.B', 'IX'),
    ('IX.C', 'IX.C', 'IX')
) AS seed(code, name, level)
ON CONFLICT (academic_year_id, code) DO UPDATE
SET name = EXCLUDED.name,
    level = EXCLUDED.level,
    is_active = TRUE,
    updated_at = NOW();

INSERT INTO subjects (code, name, is_active)
VALUES
  ('QH', 'Al-Qur''an Hadis', TRUE),
  ('AA', 'Akidah Akhlak', TRUE),
  ('FIKIH', 'Fikih', TRUE),
  ('SKI', 'Sejarah Kebudayaan Islam', TRUE),
  ('PP', 'Pendidikan Pancasila', TRUE),
  ('BIND', 'Bahasa Indonesia', TRUE),
  ('BARAB', 'Bahasa Arab', TRUE),
  ('BING', 'Bahasa Inggris', TRUE),
  ('MTK', 'Matematika', TRUE),
  ('IPA', 'Ilmu Pengetahuan Alam', TRUE),
  ('IPS', 'Ilmu Pengetahuan Sosial', TRUE),
  ('PJOK', 'Pendidikan Jasmani, Olahraga, dan Kesehatan', TRUE),
  ('INF', 'Informatika', TRUE),
  ('SBD', 'Seni Budaya', TRUE),
  ('PRA', 'Prakarya', TRUE),
  ('BK', 'Bimbingan Konseling', TRUE),
  ('MULOK', 'Muatan Lokal', TRUE),
  ('P5RA', 'Projek Penguatan Profil Pelajar Pancasila dan Profil Pelajar Rahmatan lil Alamin', TRUE)
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    is_active = TRUE,
    updated_at = NOW();
