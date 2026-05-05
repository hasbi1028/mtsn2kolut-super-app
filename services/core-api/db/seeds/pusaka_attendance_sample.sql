-- Seed: Sample PUSAKA attendance data for dev/staging
-- Jalankan manual: psql "$DATABASE_URL" -f services/core-api/db/seeds/pusaka_attendance_sample.sql
-- Aman dijalankan berulang (ON CONFLICT DO NOTHING).
-- JANGAN jadikan migration — ini dev/seed data, bukan schema change.

BEGIN;

-- 1. Sample employees (PNS/PPPK eligible untuk PUSAKA)
INSERT INTO employees (id, nip, nama, unit_kerja, employment_type, is_active)
VALUES
  ('a1000000-0000-0000-0000-000000000001', '197501012005011001', 'Ahmad Fauzi, S.Pd.',        'Guru Matematika',       'pns',   TRUE),
  ('a1000000-0000-0000-0000-000000000002', '198003152006042002', 'Siti Rahayu, S.Pd.I.',      'Guru Bahasa Indonesia', 'pns',   TRUE),
  ('a1000000-0000-0000-0000-000000000003', '199205102019031003', 'Muhammad Ilham, S.Pd.',     'Guru IPA',              'pppk',  TRUE),
  ('a1000000-0000-0000-0000-000000000004', '198807202015041004', 'Fitriani, S.Pd.',            'Guru IPS',              'pns',   TRUE),
  ('a1000000-0000-0000-0000-000000000005', '197912052003122005', 'Hasmawati, S.Ag.',           'Guru PAI',              'pns',   TRUE),
  ('a1000000-0000-0000-0000-000000000006', '200001102022031006', 'Rahmat Hidayat, S.Pd.',     'Tata Usaha',            'pppk',  TRUE),
  ('a1000000-0000-0000-0000-000000000007', '198504182009012007', 'Nurjannah, S.Pd.',          'Guru Bahasa Inggris',   'pns',   TRUE),
  ('a1000000-0000-0000-0000-000000000008', '197806092001121008', 'Abdul Karim, S.Pd.I.',      'Guru Fikih',            'pns',   TRUE),
  ('a1000000-0000-0000-0000-000000000009', '199308152020122009', 'Dewi Anggraini, S.Pd.',     'Guru PKn',              'pppk',  TRUE),
  ('a1000000-0000-0000-0000-000000000010', '198111302007011010', 'Syarifuddin, S.Pd.',        'Guru Penjaskes',        'pns',   TRUE)
ON CONFLICT (nip) DO NOTHING;

-- 2. Seed job (source untuk attendance records)
INSERT INTO jobs (id, employee_id, run_type, status, attempts, max_attempts)
VALUES (
  'b2000000-0000-0000-0000-000000000001',
  'a1000000-0000-0000-0000-000000000001',
  'checkin',
  'success',
  1,
  3
)
ON CONFLICT (id) DO NOTHING;

-- 3. Attendance records — 3 hari terakhir, variasi status
-- Gunakan tanggal relatif agar seed tetap relevan saat dijalankan
DO $$
DECLARE
  today    DATE := CURRENT_DATE;
  yday     DATE := CURRENT_DATE - INTERVAL '1 day';
  twodays  DATE := CURRENT_DATE - INTERVAL '2 days';
  seed_job UUID := 'b2000000-0000-0000-0000-000000000001';
  emp      RECORD;
BEGIN
  -- Hari ini: semua pegawai — variasi masuk/pulang
  FOR emp IN
    SELECT id, ROW_NUMBER() OVER (ORDER BY id) AS rn
    FROM employees
    WHERE id IN (
      'a1000000-0000-0000-0000-000000000001',
      'a1000000-0000-0000-0000-000000000002',
      'a1000000-0000-0000-0000-000000000003',
      'a1000000-0000-0000-0000-000000000004',
      'a1000000-0000-0000-0000-000000000005',
      'a1000000-0000-0000-0000-000000000006',
      'a1000000-0000-0000-0000-000000000007',
      'a1000000-0000-0000-0000-000000000008',
      'a1000000-0000-0000-0000-000000000009',
      'a1000000-0000-0000-0000-000000000010'
    )
  LOOP
    INSERT INTO attendance_records (employee_id, tanggal, jam_masuk, jam_pulang, source_job_id)
    VALUES (
      emp.id,
      today,
      CASE WHEN emp.rn <= 9 THEN '07:' || LPAD(((emp.rn - 1) * 3)::TEXT, 2, '0') || ':00 WITA' ELSE '' END,
      CASE WHEN emp.rn <= 7 THEN '14:' || LPAD(((emp.rn - 1) * 4)::TEXT, 2, '0') || ':00 WITA'
           WHEN emp.rn = 8 THEN ''
           ELSE '' END,
      seed_job
    )
    ON CONFLICT (employee_id, tanggal) DO NOTHING;
  END LOOP;

  -- Kemarin: semua hadir lengkap
  FOR emp IN
    SELECT id
    FROM employees
    WHERE id IN (
      'a1000000-0000-0000-0000-000000000001',
      'a1000000-0000-0000-0000-000000000002',
      'a1000000-0000-0000-0000-000000000003',
      'a1000000-0000-0000-0000-000000000004',
      'a1000000-0000-0000-0000-000000000005',
      'a1000000-0000-0000-0000-000000000006',
      'a1000000-0000-0000-0000-000000000007',
      'a1000000-0000-0000-0000-000000000008',
      'a1000000-0000-0000-0000-000000000009',
      'a1000000-0000-0000-0000-000000000010'
    )
  LOOP
    INSERT INTO attendance_records (employee_id, tanggal, jam_masuk, jam_pulang, source_job_id)
    VALUES (emp.id, yday, '07:30:00 WITA', '14:00:00 WITA', seed_job)
    ON CONFLICT (employee_id, tanggal) DO NOTHING;
  END LOOP;

  -- Dua hari lalu: sebagian hadir, sebagian tidak
  FOR emp IN
    SELECT id, ROW_NUMBER() OVER (ORDER BY id) AS rn
    FROM employees
    WHERE id IN (
      'a1000000-0000-0000-0000-000000000001',
      'a1000000-0000-0000-0000-000000000002',
      'a1000000-0000-0000-0000-000000000003',
      'a1000000-0000-0000-0000-000000000004',
      'a1000000-0000-0000-0000-000000000005',
      'a1000000-0000-0000-0000-000000000006',
      'a1000000-0000-0000-0000-000000000007'
    )
  LOOP
    INSERT INTO attendance_records (employee_id, tanggal, jam_masuk, jam_pulang, source_job_id)
    VALUES (
      emp.id,
      twodays,
      '07:35:00 WITA',
      CASE WHEN emp.rn % 3 = 0 THEN '' ELSE '14:05:00 WITA' END,
      seed_job
    )
    ON CONFLICT (employee_id, tanggal) DO NOTHING;
  END LOOP;
END;
$$;

COMMIT;
