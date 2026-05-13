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

WITH ay AS (
  SELECT id FROM academic_years WHERE name = '2025/2026' LIMIT 1
), event_row AS (
  SELECT e.id
  FROM cbt_exam_events e
  JOIN ay ON ay.id = e.academic_year_id
  WHERE e.title = 'UTS Genap'
  LIMIT 1
), subject_list AS (
  SELECT s.id, s.code, s.name
  FROM subjects s
  WHERE s.code = ANY(ARRAY['PP','BIND','PJOK','IPA','IPS','MLOKAG','FIKIH','AA','INF','MTK','MLOKWIRA','QH','BARAB','SBD','BING','SKI'])
), classes AS (
  SELECT sc.id, sc.code, sc.name, sc.level
  FROM school_classes sc
  JOIN ay ON ay.id = sc.academic_year_id
  WHERE sc.is_active = TRUE AND sc.level IN ('VII','VIII')
), package_rows AS (
  SELECT p.id, p.title, p.subject_id, c.id AS class_id, c.code AS class_code, c.level, s.name AS subject_name, event_row.id AS event_id
  FROM cbt_packages p
  JOIN event_row ON p.event_id = event_row.id
  JOIN subject_list s ON s.id = p.subject_id
  JOIN classes c ON p.title = 'UTS Genap - ' || c.code || ' - ' || s.name
), sessions_insert AS (
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
), rooms_insert AS (
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
  (SELECT COUNT(*) FROM package_rows) AS package_rows,
  (SELECT COUNT(*) FROM sessions_insert) AS newly_inserted_sessions,
  (SELECT COUNT(*) FROM session_rows) AS total_sessions,
  (SELECT COUNT(*) FROM rooms_insert) AS newly_inserted_rooms,
  (SELECT COUNT(*) FROM cbt_exam_rooms r JOIN session_rows sr ON sr.id = r.session_id) AS total_rooms,
  (SELECT COUNT(*) FROM participant_insert) AS newly_inserted_participants,
  (SELECT COUNT(*) FROM cbt_exam_participants ep JOIN session_rows sr ON sr.id = ep.session_id) AS total_participants;

COMMIT;
