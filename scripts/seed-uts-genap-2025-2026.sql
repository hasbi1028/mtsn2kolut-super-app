\set ON_ERROR_STOP on
\if :{?confirm_uts_genap_seed}
\else
\echo 'Refusing direct seed execution. Use scripts/run-uts-genap-seed-guarded.sh --apply after backup and approval.'
\quit 3
\endif
\if :confirm_uts_genap_seed
\else
\echo 'Refusing seed execution: confirm_uts_genap_seed must be 1.'
\quit 3
\endif
BEGIN;

-- Seed resmi UTS Genap 2025/2026 untuk tingkat VII dan VIII.
-- Diset idempotent-ish: jika event sudah ada, data pendukung akan direuse/dilengkapi.

WITH upsert_subjects(code, name) AS (
  VALUES
    ('PP', 'Pendidikan Pancasila'),
    ('BIND', 'Bahasa Indonesia'),
    ('PJOK', 'Pendidikan Jasmani, Olahraga, dan Kesehatan'),
    ('IPA', 'Ilmu Pengetahuan Alam'),
    ('IPS', 'Ilmu Pengetahuan Sosial'),
    ('MLOKAG', 'Muatan Lokal Keagamaan'),
    ('FIKIH', 'Fikih'),
    ('AA', 'Akidah Akhlak'),
    ('INF', 'Informatika'),
    ('MTK', 'Matematika'),
    ('MLOKWIRA', 'Mulok Kewirausahaan'),
    ('QH', 'Al-Qur''an Hadis'),
    ('BARAB', 'Bahasa Arab'),
    ('SBD', 'Seni Budaya'),
    ('BING', 'Bahasa Inggris'),
    ('SKI', 'Sejarah Kebudayaan Islam')
)
INSERT INTO subjects (code, name, is_active)
SELECT code, name, TRUE FROM upsert_subjects
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    is_active = TRUE,
    updated_at = NOW();

WITH ay AS (
  SELECT id FROM academic_years WHERE name = '2025/2026' LIMIT 1
), ev AS (
  INSERT INTO cbt_exam_events (title, exam_type, scope, target_levels, academic_year_id, status)
  SELECT 'UTS Genap', 'uts'::cbt_exam_type, 'grade', ARRAY['VII','VIII']::text[], ay.id, 'draft'
  FROM ay
  WHERE NOT EXISTS (
    SELECT 1 FROM cbt_exam_events e
    WHERE e.title = 'UTS Genap' AND e.academic_year_id = ay.id
  )
  RETURNING id
), event_row AS (
  SELECT id FROM ev
  UNION ALL
  SELECT e.id
  FROM cbt_exam_events e
  JOIN ay ON ay.id = e.academic_year_id
  WHERE e.title = 'UTS Genap'
  LIMIT 1
), subject_list AS (
  SELECT s.id, s.code, s.name
  FROM subjects s
  WHERE s.code = ANY(ARRAY['PP','BIND','PJOK','IPA','IPS','MLOKAG','FIKIH','AA','INF','MTK','MLOKWIRA','QH','BARAB','SBD','BING','SKI'])
), target_insert AS (
  INSERT INTO cbt_event_subject_targets (event_id, subject_id, target_questions)
  SELECT event_row.id, subject_list.id, 25
  FROM event_row CROSS JOIN subject_list
  ON CONFLICT (event_id, subject_id) DO UPDATE
  SET target_questions = EXCLUDED.target_questions,
      updated_at = NOW()
  RETURNING id
), target_default AS (
  INSERT INTO cbt_event_question_requirements (event_id, scope_mode, level, class_id, subject_id, target_pg, target_essay, status_filter)
  SELECT event_row.id, 'per_rombel', NULL, NULL, NULL, 20, 5, 'published_only'
  FROM event_row
  ON CONFLICT DO NOTHING
  RETURNING id
), classes AS (
  SELECT sc.id, sc.code, sc.name, sc.level
  FROM school_classes sc
  JOIN ay ON ay.id = sc.academic_year_id
  WHERE sc.is_active = TRUE AND sc.level IN ('VII','VIII')
), req_insert AS (
  INSERT INTO cbt_event_question_requirements (event_id, scope_mode, level, class_id, subject_id, target_pg, target_essay, status_filter)
  SELECT event_row.id, 'per_rombel', classes.level, classes.id, subject_list.id, 20, 5, 'published_only'
  FROM event_row CROSS JOIN classes CROSS JOIN subject_list
  WHERE NOT EXISTS (
    SELECT 1 FROM cbt_event_question_requirements r
    WHERE r.event_id = event_row.id
      AND r.scope_mode = 'per_rombel'
      AND r.class_id = classes.id
      AND r.subject_id = subject_list.id
  )
  RETURNING id
), packages AS (
  INSERT INTO cbt_packages (
    event_id, subject_id, title, description, duration_minutes,
    randomize_questions, randomize_options, draw_pg_count, draw_essay_count,
    source_mode, random_seed, composition_log, is_active
  )
  SELECT
    event_row.id,
    subject_list.id,
    'UTS Genap - ' || classes.code || ' - ' || subject_list.name,
    'Paket draft UTS Genap 2025/2026; tingkat ' || classes.level || '; rombel ' || classes.code || '; target 20 PG + 5 essay.',
    60,
    TRUE,
    TRUE,
    20,
    5,
    'teacher_class',
    encode(gen_random_bytes(8), 'hex'),
    jsonb_build_object('seeded_by','seed-uts-genap-2025-2026.sql','scope','per_rombel','level',classes.level,'class_code',classes.code,'target_pg',20,'target_essay',5),
    TRUE
  FROM event_row CROSS JOIN classes CROSS JOIN subject_list
  WHERE NOT EXISTS (
    SELECT 1 FROM cbt_packages p
    WHERE p.event_id = event_row.id
      AND p.subject_id = subject_list.id
      AND p.title = 'UTS Genap - ' || classes.code || ' - ' || subject_list.name
  )
  RETURNING id, title
), package_rows AS (
  SELECT p.id, p.title, p.subject_id, c.id AS class_id, c.code AS class_code, c.level, s.name AS subject_name, event_row.id AS event_id
  FROM cbt_packages p
  JOIN event_row ON p.event_id = event_row.id
  JOIN subject_list s ON s.id = p.subject_id
  JOIN classes c ON p.title = 'UTS Genap - ' || c.code || ' - ' || s.name
), sessions AS (
  INSERT INTO cbt_exam_sessions (
    package_id, class_id, title, scheduled_start, scheduled_end, status, event_id,
    scope_type, scope_ref, mix_policy, assignment_mode, allow_cross_grade, is_special_event
  )
  SELECT
    pr.id,
    pr.class_id,
    'Sesi ' || pr.title,
    TIMESTAMPTZ '2026-06-04 07:30:00+08',
    TIMESTAMPTZ '2026-06-04 08:30:00+08',
    'draft'::cbt_session_status_enum,
    pr.event_id,
    'class',
    pr.class_id::text,
    'same_grade',
    'random_balanced',
    FALSE,
    FALSE
  FROM package_rows pr
  WHERE NOT EXISTS (
    SELECT 1 FROM cbt_exam_sessions es
    WHERE es.package_id = pr.id AND es.class_id = pr.class_id
  )
  RETURNING id, package_id, class_id
), session_rows AS (
  SELECT es.id, es.package_id, es.class_id
  FROM cbt_exam_sessions es
  JOIN package_rows pr ON pr.id = es.package_id AND pr.class_id = es.class_id
), room_nums AS (
  SELECT generate_series(1, 7) AS room_no
), rooms AS (
  INSERT INTO cbt_exam_rooms (
    session_id, room_name, capacity, school_room_id, room_name_snapshot,
    capacity_override, room_token, status, is_locked
  )
  SELECT
    sr.id,
    'Ruang Ujian ' || lpad(room_nums.room_no::text, 2, '0'),
    20,
    NULL,
    'Ruang Ujian ' || lpad(room_nums.room_no::text, 2, '0'),
    20,
    encode(gen_random_bytes(16), 'hex'),
    'draft',
    FALSE
  FROM session_rows sr CROSS JOIN room_nums
  WHERE NOT EXISTS (
    SELECT 1 FROM cbt_exam_rooms r
    WHERE r.session_id = sr.id
      AND r.room_name = 'Ruang Ujian ' || lpad(room_nums.room_no::text, 2, '0')
  )
  RETURNING id, session_id, room_name
), all_rooms AS (
  SELECT r.id, r.session_id, r.room_name,
         row_number() OVER (PARTITION BY r.session_id ORDER BY r.room_name) AS room_idx
  FROM cbt_exam_rooms r
  JOIN session_rows sr ON sr.id = r.session_id
), candidate_participants AS (
  SELECT
    sr.id AS session_id,
    st.id AS student_id,
    row_number() OVER (PARTITION BY sr.id ORDER BY random(), st.nama) AS rn
  FROM session_rows sr
  JOIN students st ON st.class_id = sr.class_id
  WHERE st.is_active = TRUE
), participant_seed AS (
  SELECT
    cp.session_id,
    cp.student_id,
    ar.id AS room_id,
    ((cp.rn - 1) / 7) + 1 AS seat_no
  FROM candidate_participants cp
  JOIN all_rooms ar ON ar.session_id = cp.session_id AND ar.room_idx = (((cp.rn - 1) % 7) + 1)
), participant_insert AS (
  INSERT INTO cbt_exam_participants (session_id, student_id, token, room_id, seat_no)
  SELECT
    ps.session_id,
    ps.student_id,
    encode(gen_random_bytes(16), 'hex'),
    ps.room_id,
    ps.seat_no
  FROM participant_seed ps
  WHERE NOT EXISTS (
    SELECT 1 FROM cbt_exam_participants ep
    WHERE ep.session_id = ps.session_id AND ep.student_id = ps.student_id
  )
  RETURNING id
)
SELECT
  (SELECT id FROM event_row) AS event_id,
  (SELECT COUNT(*) FROM subject_list) AS subject_count,
  (SELECT COUNT(*) FROM classes) AS class_count,
  (SELECT COUNT(*) FROM package_rows) AS total_packages,
  (SELECT COUNT(*) FROM session_rows) AS total_sessions,
  (SELECT COUNT(*) FROM cbt_exam_rooms r JOIN session_rows sr ON sr.id = r.session_id) AS total_rooms,
  (SELECT COUNT(*) FROM cbt_exam_participants ep JOIN session_rows sr ON sr.id = ep.session_id) AS total_participants,
  (SELECT COUNT(*) FROM participant_insert) AS newly_inserted_participants;

COMMIT;
