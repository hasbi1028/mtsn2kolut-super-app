-- ============================================
-- IMPORT DATA MTSN 2 KOLAKA UTARA 2026/2027
-- Run in project directory:
--   cd services/core-api && source .env && ./scripts/import-all.sh
-- ============================================

-- Active academic year and semester
\set ay_id 'e3dfb158-76c6-46c2-b6e7-00cd604a0277'
\set smt_id 'ddbe7929-6861-4989-b408-40e27655e8b8'

-- ============ 1. CLASSES ============
UPDATE school_classes SET academic_year_id = :'ay_id' WHERE code IN ('VII.A','VII.B','VII.C','VII.D','VIII.A','VIII.B','VIII.C','IX.A','IX.B','IX.C');

INSERT INTO school_classes (id, academic_year_id, code, name, level, is_active)
VALUES (gen_random_uuid(), :'ay_id', 'VII.E', 'VII.E', 'VII', true)
ON CONFLICT DO NOTHING;

INSERT INTO school_classes (id, academic_year_id, code, name, level, is_active)
VALUES (gen_random_uuid(), :'ay_id', 'VIII.D', 'VIII.D', 'VIII', true)
ON CONFLICT DO NOTHING;

-- ============ 2. DELETE OLD ASSIGNMENTS ============
DELETE FROM class_subject_assignments WHERE semester_id = :'smt_id';

-- ============ 3. ASSIGN GURU ============
INSERT INTO class_subject_assignments (id, class_id, subject_id, teacher_employee_id, semester_id)
SELECT gen_random_uuid(), c.id, s.id, t.id, :'smt_id'::uuid
FROM (VALUES
    ('andi rasnia rasjid', 'al-qur an hadis', 'VII.A'),
    ('andi rasnia rasjid', 'al-qur an hadis', 'VII.B'),
    ('andi rasnia rasjid', 'al-qur an hadis', 'VII.C'),
    ('andi rasnia rasjid', 'al-qur an hadis', 'VII.D'),
    ('andi rasnia rasjid', 'al-qur an hadis', 'VII.E'),
    ('saimang', 'al-qur an hadis', 'VIII.A'),
    ('saimang', 'al-qur an hadis', 'VIII.B'),
    ('saimang', 'al-qur an hadis', 'VIII.C'),
    ('saimang', 'al-qur an hadis', 'VIII.D'),
    ('saimang', 'al-qur an hadis', 'IX.A'),
    ('saimang', 'al-qur an hadis', 'IX.B'),
    ('saimang', 'al-qur an hadis', 'IX.C'),
    ('ratnawati', 'akidah akhlak', 'VIII.A'),
    ('ratnawati', 'akidah akhlak', 'VIII.B'),
    ('ratnawati', 'akidah akhlak', 'IX.A'),
    ('ratnawati', 'akidah akhlak', 'IX.B'),
    ('ratnawati', 'akidah akhlak', 'IX.C'),
    ('sitti rafidah', 'akidah akhlak', 'VII.A'),
    ('sitti rafidah', 'akidah akhlak', 'VII.B'),
    ('sitti rafidah', 'akidah akhlak', 'VII.C'),
    ('sitti rafidah', 'akidah akhlak', 'VII.D'),
    ('sitti rafidah', 'akidah akhlak', 'VII.E'),
    ('sitti rafidah', 'akidah akhlak', 'VIII.C'),
    ('sitti rafidah', 'akidah akhlak', 'VIII.D'),
    ('sitti rafidah', 'fikih', 'VIII.A'),
    ('sitti rafidah', 'fikih', 'VIII.B'),
    ('sitti rafidah', 'fikih', 'VIII.C'),
    ('sitti rafidah', 'fikih', 'VIII.D'),
    ('sitti rahma', 'al-qur an hadis', 'VIII.A'),
    ('sitti rahma', 'al-qur an hadis', 'VIII.B'),
    ('sitti rahma', 'al-qur an hadis', 'VIII.C'),
    ('sitti rahma', 'al-qur an hadis', 'VIII.D'),
    ('sitti rahma', 'fikih', 'VII.E'),
    ('sitti rahma', 'fikih', 'IX.A'),
    ('sitti rahma', 'fikih', 'IX.B'),
    ('sitti rahma', 'fikih', 'IX.C'),
    ('mida', 'fikih', 'VII.A'),
    ('mida', 'fikih', 'VII.B'),
    ('mida', 'fikih', 'VII.C'),
    ('mida', 'fikih', 'VII.D'),
    ('mida', 'ilmu pengetahuan sosial', 'VII.D'),
    ('mida', 'ilmu pengetahuan sosial', 'VII.E'),
    ('susianti', 'sejarah kebudayaan islam', 'VII.A'),
    ('susianti', 'sejarah kebudayaan islam', 'VII.B'),
    ('susianti', 'sejarah kebudayaan islam', 'VII.C'),
    ('susianti', 'sejarah kebudayaan islam', 'VII.D'),
    ('susianti', 'sejarah kebudayaan islam', 'VII.E'),
    ('susianti', 'sejarah kebudayaan islam', 'VIII.A'),
    ('susianti', 'sejarah kebudayaan islam', 'VIII.B'),
    ('susianti', 'sejarah kebudayaan islam', 'VIII.C'),
    ('susianti', 'sejarah kebudayaan islam', 'VIII.D'),
    ('susianti', 'sejarah kebudayaan islam', 'IX.A'),
    ('susianti', 'sejarah kebudayaan islam', 'IX.B'),
    ('susianti', 'sejarah kebudayaan islam', 'IX.C'),
    ('sutra', 'bahasa arab', 'VII.A'),
    ('sutra', 'bahasa arab', 'VII.B'),
    ('sutra', 'bahasa arab', 'VII.C'),
    ('sutra', 'bahasa arab', 'VII.D'),
    ('sutra', 'bahasa arab', 'VII.E'),
    ('sutra', 'bahasa arab', 'IX.A'),
    ('sutra', 'bahasa arab', 'IX.B'),
    ('km muh tang', 'bahasa arab', 'VIII.B'),
    ('nur ilmi', 'bahasa arab', 'VIII.A'),
    ('nur ilmi', 'bahasa arab', 'IX.C'),
    ('nur ilmi', 'muatan lokal keagamaan', 'VIII.C'),
    ('nur ilmi', 'muatan lokal keagamaan', 'VIII.D'),
    ('jayanti jufri', 'bahasa arab', 'VIII.C'),
    ('jayanti jufri', 'bahasa arab', 'VIII.D'),
    ('jayanti jufri', 'muatan lokal keagamaan', 'VIII.A'),
    ('jayanti jufri', 'muatan lokal keagamaan', 'VIII.B'),
    ('herniati', 'bahasa indonesia', 'VIII.A'),
    ('herniati', 'bahasa indonesia', 'IX.A'),
    ('herniati', 'bahasa indonesia', 'IX.B'),
    ('herniati', 'bahasa indonesia', 'IX.C'),
    ('wati zaelani', 'bahasa indonesia', 'VIII.A'),
    ('wati zaelani', 'bahasa indonesia', 'VIII.B'),
    ('wati zaelani', 'bahasa indonesia', 'VIII.C'),
    ('nirmawati', 'bahasa indonesia', 'VII.A'),
    ('andi mirna', 'bahasa indonesia', 'VII.C'),
    ('andi mirna', 'informatika', 'IX.A'),
    ('andi mirna', 'informatika', 'IX.B'),
    ('rustiani', 'bahasa indonesia', 'VII.B'),
    ('ernawati', 'bahasa indonesia', 'VII.D'),
    ('ernawati', 'informatika', 'VII.D'),
    ('ernawati', 'informatika', 'VII.E'),
    ('ernawati', 'informatika', 'VIII.A'),
    ('ernawati', 'informatika', 'VIII.B'),
    ('ernawati', 'informatika', 'VIII.C'),
    ('ernawati', 'informatika', 'VIII.D'),
    ('abdillah', 'ilmu pengetahuan alam', 'VII.D'),
    ('abdillah', 'ilmu pengetahuan alam', 'VII.E'),
    ('abdillah', 'ilmu pengetahuan alam', 'IX.A'),
    ('abdillah', 'ilmu pengetahuan alam', 'IX.B'),
    ('abdillah', 'ilmu pengetahuan alam', 'IX.C'),
    ('kardi', 'ilmu pengetahuan alam', 'VIII.A'),
    ('kardi', 'ilmu pengetahuan alam', 'VIII.B'),
    ('kardi', 'ilmu pengetahuan alam', 'VIII.C'),
    ('kardi', 'ilmu pengetahuan alam', 'VIII.D'),
    ('mirnawati', 'ilmu pengetahuan alam', 'VII.A'),
    ('mirnawati', 'ilmu pengetahuan alam', 'VII.B'),
    ('mirnawati', 'ilmu pengetahuan alam', 'VII.C'),
    ('kaharuddin', 'pendidikan jasmani olahraga dan kesehatan', 'VII.A'),
    ('kaharuddin', 'pendidikan jasmani olahraga dan kesehatan', 'VII.B'),
    ('kaharuddin', 'pendidikan jasmani olahraga dan kesehatan', 'VII.C'),
    ('kaharuddin', 'pendidikan jasmani olahraga dan kesehatan', 'VIII.A'),
    ('kaharuddin', 'pendidikan jasmani olahraga dan kesehatan', 'VIII.B'),
    ('kaharuddin', 'pendidikan jasmani olahraga dan kesehatan', 'VIII.C'),
    ('kaharuddin', 'pendidikan jasmani olahraga dan kesehatan', 'VIII.D'),
    ('kaharuddin', 'pendidikan jasmani olahraga dan kesehatan', 'IX.A'),
    ('kaharuddin', 'pendidikan jasmani olahraga dan kesehatan', 'IX.B'),
    ('kaharuddin', 'pendidikan jasmani olahraga dan kesehatan', 'IX.C'),
    ('sri hardianti', 'pendidikan jasmani olahraga dan kesehatan', 'VII.D'),
    ('sri hardianti', 'pendidikan jasmani olahraga dan kesehatan', 'VII.E'),
    ('irmawati nur', 'matematika', 'VII.A'),
    ('irmawati nur', 'matematika', 'VII.B'),
    ('irmawati nur', 'matematika', 'VII.C'),
    ('irmawati nur', 'matematika', 'VII.D'),
    ('nurul fitri usman', 'matematika', 'VII.E'),
    ('nurul fitri usman', 'matematika', 'VIII.A'),
    ('nurul fitri usman', 'matematika', 'VIII.B'),
    ('nurul fitri usman', 'matematika', 'VIII.C'),
    ('nurul fitri usman', 'matematika', 'VIII.D'),
    ('taufik', 'matematika', 'IX.A'),
    ('taufik', 'matematika', 'IX.B'),
    ('taufik', 'matematika', 'IX.C'),
    ('muh saing', 'pendidikan pancasila', 'VII.A'),
    ('muh saing', 'pendidikan pancasila', 'VII.B'),
    ('muh saing', 'pendidikan pancasila', 'VII.C'),
    ('muh saing', 'pendidikan pancasila', 'VII.D'),
    ('muh saing', 'pendidikan pancasila', 'VII.E'),
    ('muh saing', 'pendidikan pancasila', 'VIII.A'),
    ('muh saing', 'pendidikan pancasila', 'VIII.B'),
    ('muh saing', 'pendidikan pancasila', 'IX.A'),
    ('muh saing', 'pendidikan pancasila', 'IX.B'),
    ('muh saing', 'pendidikan pancasila', 'IX.C'),
    ('mutmainnah', 'ilmu pengetahuan sosial', 'VII.A'),
    ('mutmainnah', 'ilmu pengetahuan sosial', 'VII.B'),
    ('mutmainnah', 'ilmu pengetahuan sosial', 'VII.C'),
    ('mutmainnah', 'ilmu pengetahuan sosial', 'VIII.A'),
    ('mutmainnah', 'ilmu pengetahuan sosial', 'VIII.B'),
    ('mutmainnah', 'ilmu pengetahuan sosial', 'VIII.C'),
    ('mutmainnah', 'ilmu pengetahuan sosial', 'VIII.D'),
    ('mutmainnah', 'pendidikan pancasila', 'VIII.C'),
    ('mutmainnah', 'pendidikan pancasila', 'VIII.D'),
    ('drs h chaeruddin', 'ilmu pengetahuan sosial', 'IX.A'),
    ('drs h chaeruddin', 'ilmu pengetahuan sosial', 'IX.B'),
    ('drs h chaeruddin', 'ilmu pengetahuan sosial', 'IX.C'),
    ('rusnawati', 'bahasa inggris', 'VIII.A'),
    ('rusnawati', 'bahasa inggris', 'VIII.B'),
    ('rusnawati', 'bahasa inggris', 'IX.A'),
    ('rusnawati', 'bahasa inggris', 'IX.B'),
    ('rusnawati', 'bahasa inggris', 'IX.C'),
    ('yurnianti', 'bahasa inggris', 'VII.A'),
    ('yurnianti', 'bahasa inggris', 'VII.B'),
    ('yurnianti', 'bahasa inggris', 'VII.C'),
    ('yurnianti', 'bahasa inggris', 'VII.D'),
    ('yurnianti', 'bahasa inggris', 'VII.E'),
    ('yurnianti', 'informatika', 'IX.C'),
    ('nurunnisa', 'bahasa inggris', 'VIII.C'),
    ('nurunnisa', 'bahasa inggris', 'VIII.D'),
    ('supriadi', 'seni budaya', 'VII.A'),
    ('supriadi', 'seni budaya', 'VII.B'),
    ('supriadi', 'seni budaya', 'VII.C'),
    ('supriadi', 'seni budaya', 'VIII.A'),
    ('supriadi', 'seni budaya', 'VIII.B'),
    ('supriadi', 'seni budaya', 'VIII.C'),
    ('supriadi', 'seni budaya', 'VIII.D'),
    ('supriadi', 'seni budaya', 'IX.A'),
    ('supriadi', 'seni budaya', 'IX.B'),
    ('supriadi', 'seni budaya', 'IX.C'),
    ('asniar', 'seni budaya', 'VII.D'),
    ('asniar', 'seni budaya', 'VII.E'),
    ('km muhammad yani', 'muatan lokal keagamaan', 'VII.A'),
    ('km muhammad yani', 'muatan lokal keagamaan', 'VII.B'),
    ('km muhammad yani', 'muatan lokal keagamaan', 'VII.C'),
    ('km muhammad yani', 'muatan lokal keagamaan', 'VII.D'),
    ('km muhammad yani', 'muatan lokal keagamaan', 'VII.E'),
    ('km muhammad yani', 'muatan lokal keagamaan', 'IX.A'),
    ('km muhammad yani', 'muatan lokal keagamaan', 'IX.B'),
    ('km muhammad yani', 'muatan lokal keagamaan', 'IX.C'),
    ('muh saing', 'informatika', 'VII.A'),
    ('muh saing', 'informatika', 'VII.B'),
    ('muh saing', 'informatika', 'VII.C')
) AS data(teacher_name, subject_name, class_code)
JOIN employees t ON (
    LOWER(REGEXP_REPLACE(t.nama, '[,.\-]', ' ', 'g')) LIKE '%' || LOWER(REGEXP_REPLACE(data.teacher_name, '[,.\-]', ' ', 'g')) || '%'
    OR LOWER(REGEXP_REPLACE(data.teacher_name, '[,.\-]', ' ', 'g')) LIKE '%' || LOWER(REGEXP_REPLACE(t.nama, '[,.\-]', ' ', 'g')) || '%'
)
JOIN subjects s ON (
    LOWER(REGEXP_REPLACE(s.name, '[,.\-]', ' ', 'g')) LIKE '%' || LOWER(REGEXP_REPLACE(data.subject_name, '[,.\-]', ' ', 'g')) || '%'
    OR LOWER(REGEXP_REPLACE(data.subject_name, '[,.\-]', ' ', 'g')) LIKE '%' || LOWER(REGEXP_REPLACE(s.name, '[,.\-]', ' ', 'g')) || '%'
)
JOIN school_classes c ON c.code = data.class_code AND c.academic_year_id = :'ay_id'
ON CONFLICT (class_id, subject_id) DO NOTHING;

-- ============ VERIFY ============
SELECT 'CLASSES' AS step, COUNT(*)::text FROM school_classes WHERE academic_year_id = :'ay_id'
UNION ALL
SELECT 'ASSIGNMENTS', COUNT(*)::text FROM class_subject_assignments WHERE semester_id = :'smt_id'
UNION ALL
SELECT 'ACTIVE_AY', name FROM academic_years WHERE id = :'ay_id';
