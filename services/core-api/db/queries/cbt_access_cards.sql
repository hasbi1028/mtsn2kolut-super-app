-- name: CreateCbtExamAccessCard :one
INSERT INTO cbt_exam_access_cards (
  card_type, event_id, session_id, room_id, participant_id, room_proctor_id,
  assigned_user_id, token_hash, pin_hash, expires_at, generated_by
)
VALUES (
  $1, $2, $3, $4, $5, $6,
  $7, $8, $9, $10, $11
)
RETURNING *;

-- name: RevokeActiveParticipantAccessCards :exec
UPDATE cbt_exam_access_cards
SET status = 'revoked', revoked_at = NOW(), updated_at = NOW()
WHERE card_type = 'participant'
  AND participant_id = $1
  AND revoked_at IS NULL
  AND status = 'active';

-- name: RevokeActiveProctorAccessCards :exec
UPDATE cbt_exam_access_cards
SET status = 'revoked', revoked_at = NOW(), updated_at = NOW()
WHERE card_type = 'proctor'
  AND room_id = $1
  AND revoked_at IS NULL
  AND status = 'active';

-- name: ListCbtParticipantAccessCardTargetsByEvent :many
SELECT
  ep.id AS participant_id,
  ep.session_id,
  cs.event_id,
  ep.student_id,
  st.nis,
  st.nisn,
  st.nama AS student_name,
  COALESCE(sc.name, '') AS class_name,
  COALESCE(sc.code, '') AS class_code,
  ep.room_id,
  COALESCE(r.room_name, '') AS room_name,
  ep.seat_no,
  cs.title AS session_title,
  cs.scheduled_start,
  cs.scheduled_end,
  p.title AS package_title,
  ac.id AS card_id,
  COALESCE(ac.status, '') AS card_status,
  COALESCE(ac.failed_attempts, 0)::int AS failed_attempts,
  ac.created_at AS card_created_at,
  ac.expires_at AS card_expires_at,
  ac.revoked_at AS card_revoked_at,
  ac.verified_at AS card_verified_at,
  (CASE WHEN ac.id IS NOT NULL
          AND btrim(ep.token) <> ''
          AND ac.token_hash = encode(digest(btrim(ep.token), 'sha256'), 'hex')
        THEN btrim(ep.token) ELSE '' END)::text AS print_token,
  (CASE WHEN ac.id IS NOT NULL
          AND btrim(ep.token) <> ''
          AND ac.pin_hash = encode(digest(upper(right(btrim(ep.token), 4)), 'sha256'), 'hex')
        THEN upper(right(btrim(ep.token), 4)) ELSE '' END)::text AS print_pin
FROM cbt_exam_participants ep
JOIN students st ON st.id = ep.student_id
LEFT JOIN school_classes sc ON sc.id = st.class_id
JOIN cbt_exam_sessions cs ON cs.id = ep.session_id
JOIN cbt_packages p ON p.id = cs.package_id
LEFT JOIN cbt_exam_rooms r ON r.id = ep.room_id
LEFT JOIN LATERAL (
  SELECT id, status, failed_attempts, created_at, expires_at, revoked_at, verified_at, token_hash, pin_hash
  FROM cbt_exam_access_cards ac
  WHERE ac.card_type = 'participant'
    AND ac.participant_id = ep.id
    AND ac.revoked_at IS NULL
    AND ac.status = 'active'
  ORDER BY ac.created_at DESC
  LIMIT 1
) ac ON TRUE
WHERE cs.event_id = $1
  AND (sqlc.narg('session_id')::uuid IS NULL OR ep.session_id = sqlc.narg('session_id')::uuid)
  AND (sqlc.narg('room_id')::uuid IS NULL OR ep.room_id = sqlc.narg('room_id')::uuid)
ORDER BY cs.scheduled_start ASC, r.room_name ASC NULLS LAST, ep.seat_no ASC NULLS LAST, st.nama ASC;

-- name: ListCbtProctorAccessCardTargetsByEvent :many
SELECT
  NULL::uuid AS room_proctor_id,
  r.id AS room_id,
  r.session_id,
  cs.event_id,
  NULL::uuid AS employee_id,
  COALESCE(pa.employee_names, 'Lembar Pengawas Ruang')::text AS employee_name,
  ''::text AS nip,
  COALESCE(pa.proctor_roles, 'pengawas_ruang')::text AS proctor_role,
  r.room_name,
  cs.title AS session_title,
  cs.scheduled_start,
  cs.scheduled_end,
  p.title AS package_title,
  ac.id AS card_id,
  COALESCE(ac.status, '') AS card_status,
  COALESCE(ac.failed_attempts, 0)::int AS failed_attempts,
  ac.created_at AS card_created_at,
  ac.expires_at AS card_expires_at,
  ac.revoked_at AS card_revoked_at,
  ac.verified_at AS card_verified_at,
  (CASE WHEN ac.id IS NOT NULL
          AND btrim(r.room_token) <> ''
          AND ac.token_hash = encode(digest(btrim(r.room_token), 'sha256'), 'hex')
        THEN btrim(r.room_token) ELSE '' END)::text AS print_token,
  (CASE WHEN ac.id IS NOT NULL
          AND btrim(r.room_token) <> ''
          AND ac.pin_hash = encode(digest(upper(right(regexp_replace(btrim(r.room_token), '[^A-Za-z0-9]', '', 'g'), 4)), 'sha256'), 'hex')
        THEN upper(right(regexp_replace(btrim(r.room_token), '[^A-Za-z0-9]', '', 'g'), 4)) ELSE '' END)::text AS print_pin
FROM cbt_exam_rooms r
JOIN cbt_exam_sessions cs ON cs.id = r.session_id
JOIN cbt_packages p ON p.id = cs.package_id
LEFT JOIN LATERAL (
  SELECT
    string_agg(emp.nama, ', ' ORDER BY emp.nama) FILTER (WHERE emp.id IS NOT NULL) AS employee_names,
    string_agg(DISTINCT rp.role::text, ', ' ORDER BY rp.role::text) FILTER (WHERE rp.role IS NOT NULL AND rp.role::text <> '') AS proctor_roles
  FROM cbt_room_proctors rp
  JOIN employees emp ON emp.id = rp.employee_id
  WHERE rp.exam_room_id = r.id
) pa ON TRUE
LEFT JOIN LATERAL (
  SELECT id, status, failed_attempts, created_at, expires_at, revoked_at, verified_at, token_hash, pin_hash
  FROM cbt_exam_access_cards ac
  WHERE ac.card_type = 'proctor'
    AND ac.room_id = r.id
    AND ac.revoked_at IS NULL
    AND ac.status = 'active'
  ORDER BY ac.created_at DESC
  LIMIT 1
) ac ON TRUE
WHERE cs.event_id = $1
  AND (sqlc.narg('session_id')::uuid IS NULL OR r.session_id = sqlc.narg('session_id')::uuid)
  AND (sqlc.narg('room_id')::uuid IS NULL OR r.id = sqlc.narg('room_id')::uuid)
ORDER BY cs.scheduled_start ASC, r.room_name ASC;

-- name: GetCbtAccessCardByTokenHash :one
SELECT
  ac.id,
  ac.card_type,
  ac.event_id,
  ev.title AS event_title,
  ac.session_id,
  cs.title AS session_title,
  cs.status AS session_status,
  cs.scheduled_start,
  cs.scheduled_end,
  ac.room_id,
  COALESCE(r.room_name, '') AS room_name,
  ac.participant_id,
  ep.student_id,
  COALESCE(st.nis, '') AS nis,
  COALESCE(st.nisn, '') AS nisn,
  COALESCE(st.nama, '') AS student_name,
  COALESCE(sc.name, '') AS class_name,
  ep.seat_no,
  ac.room_proctor_id,
  rp.employee_id AS proctor_employee_id,
  (CASE
    WHEN ac.card_type = 'proctor' THEN COALESCE(emp.nama, pa.employee_names, 'Lembar Pengawas Ruang')
    ELSE COALESCE(emp.nama, '')
  END)::text AS proctor_name,
  (CASE
    WHEN ac.card_type = 'proctor' THEN COALESCE(rp.role::text, pa.proctor_roles, 'pengawas_ruang')
    ELSE COALESCE(rp.role::text, '')
  END)::text AS proctor_role,
  p.title AS package_title,
  p.duration_minutes,
  ac.pin_hash,
  ac.status AS card_status,
  ac.failed_attempts,
  ac.max_failed_attempts,
  ac.expires_at,
  ac.revoked_at,
  ac.created_at
FROM cbt_exam_access_cards ac
JOIN cbt_exam_events ev ON ev.id = ac.event_id
JOIN cbt_exam_sessions cs ON cs.id = ac.session_id
JOIN cbt_packages p ON p.id = cs.package_id
LEFT JOIN cbt_exam_rooms r ON r.id = ac.room_id
LEFT JOIN cbt_exam_participants ep ON ep.id = ac.participant_id
LEFT JOIN students st ON st.id = ep.student_id
LEFT JOIN school_classes sc ON sc.id = st.class_id
LEFT JOIN cbt_room_proctors rp ON rp.id = ac.room_proctor_id
LEFT JOIN employees emp ON emp.id = rp.employee_id
LEFT JOIN LATERAL (
  SELECT
    string_agg(e.nama, ', ' ORDER BY e.nama) FILTER (WHERE e.id IS NOT NULL) AS employee_names,
    string_agg(DISTINCT crp.role::text, ', ' ORDER BY crp.role::text) FILTER (WHERE crp.role IS NOT NULL AND crp.role::text <> '') AS proctor_roles
  FROM cbt_room_proctors crp
  JOIN employees e ON e.id = crp.employee_id
  WHERE crp.exam_room_id = ac.room_id
) pa ON TRUE
WHERE ac.token_hash = $1;

-- name: MarkCbtAccessCardVerified :exec
UPDATE cbt_exam_access_cards
SET failed_attempts = 0, last_failed_at = NULL, verified_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: RecordCbtAccessCardFailedAttempt :one
UPDATE cbt_exam_access_cards
SET failed_attempts = failed_attempts + 1,
    last_failed_at = NOW(),
    status = CASE WHEN failed_attempts + 1 >= max_failed_attempts THEN 'locked' ELSE status END,
    updated_at = NOW()
WHERE id = $1
RETURNING failed_attempts, status;
