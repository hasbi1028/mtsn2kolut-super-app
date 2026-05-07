-- name: ListEmployees :many
SELECT e.id, e.pegawai_uid, COALESCE(e.nip, '')::text AS nip, e.nama, e.unit_kerja, e.employment_type, e.tanggal_lahir,
       COALESCE(e.jenis_kelamin, '')::text AS jenis_kelamin,
       COALESCE(e.tempat_lahir, '')::text AS tempat_lahir,
       COALESCE(pa.pusaka_username, '') AS pusaka_username,
       COALESCE(pa.pusaka_password, '') AS pusaka_password,
       COALESCE(pa.is_enabled, FALSE) AS pusaka_is_enabled,
       e.is_active, e.created_at, e.updated_at
FROM employees e
LEFT JOIN pusaka_accounts pa ON pa.employee_id = e.id
ORDER BY nama ASC;

-- name: ListActiveEmployees :many
SELECT e.id, e.pegawai_uid, COALESCE(e.nip, '')::text AS nip, e.nama, e.unit_kerja, e.employment_type, e.tanggal_lahir,
       COALESCE(e.jenis_kelamin, '')::text AS jenis_kelamin,
       COALESCE(e.tempat_lahir, '')::text AS tempat_lahir,
       COALESCE(pa.pusaka_username, '') AS pusaka_username,
       COALESCE(pa.pusaka_password, '') AS pusaka_password,
       COALESCE(pa.is_enabled, FALSE) AS pusaka_is_enabled,
       e.is_active, e.created_at, e.updated_at
FROM employees e
JOIN pusaka_accounts pa ON pa.employee_id = e.id
WHERE e.is_active = TRUE
  AND pa.is_enabled = TRUE
  AND pa.pusaka_username <> ''
  AND pa.pusaka_password <> ''
ORDER BY nama ASC;

-- name: GetEmployee :one
SELECT e.id, e.pegawai_uid, COALESCE(e.nip, '')::text AS nip, e.nama, e.unit_kerja, e.employment_type, e.tanggal_lahir,
       COALESCE(e.jenis_kelamin, '')::text AS jenis_kelamin,
       COALESCE(e.tempat_lahir, '')::text AS tempat_lahir,
       COALESCE(pa.pusaka_username, '') AS pusaka_username,
       COALESCE(pa.pusaka_password, '') AS pusaka_password,
       COALESCE(pa.is_enabled, FALSE) AS pusaka_is_enabled,
       e.is_active, e.created_at, e.updated_at
FROM employees e
LEFT JOIN pusaka_accounts pa ON pa.employee_id = e.id
WHERE e.id = $1;

-- name: CreateEmployee :one
WITH input AS (
  SELECT
    sqlc.arg(npsn)::text AS npsn,
    NULLIF(btrim(sqlc.arg(nip)::text), '') AS nip,
    btrim(sqlc.arg(nama)::text) AS nama,
    btrim(sqlc.arg(unit_kerja)::text) AS unit_kerja,
    sqlc.arg(employment_type)::text AS employment_type,
    sqlc.arg(tanggal_lahir)::date AS tanggal_lahir,
    NULLIF(btrim(sqlc.arg(jenis_kelamin)::text), '') AS jenis_kelamin,
    NULLIF(btrim(sqlc.arg(tempat_lahir)::text), '') AS tempat_lahir,
    sqlc.arg(is_active)::boolean AS is_active
),
prefix AS (
  SELECT
    input.*,
    (input.npsn || COALESCE(to_char(input.tanggal_lahir, 'YY'), '00'))::text AS uid_prefix
  FROM input
),
next_uid AS (
  SELECT
    prefix.*,
    (prefix.uid_prefix || lpad((COALESCE(max(substring(e.pegawai_uid FROM length(prefix.uid_prefix) + 1 FOR 3)::int), 0) + 1)::text, 3, '0'))::text AS pegawai_uid
  FROM prefix
  LEFT JOIN employees e
    ON e.pegawai_uid ~ ('^' || prefix.uid_prefix || '[0-9]{3}$')
  GROUP BY prefix.npsn, prefix.nip, prefix.nama, prefix.unit_kerja, prefix.employment_type,
           prefix.tanggal_lahir, prefix.jenis_kelamin, prefix.tempat_lahir, prefix.is_active, prefix.uid_prefix
)
INSERT INTO employees (id, pegawai_uid, nip, nama, unit_kerja, employment_type, tanggal_lahir, jenis_kelamin, tempat_lahir, is_active)
SELECT gen_random_uuid(), pegawai_uid, nip, nama, unit_kerja, employment_type, tanggal_lahir, jenis_kelamin, tempat_lahir, is_active
FROM next_uid
RETURNING id;

-- name: UpdateEmployee :one
UPDATE employees
SET nip             = NULLIF(btrim(sqlc.arg(nip)::text), ''),
    nama            = btrim(sqlc.arg(nama)::text),
    unit_kerja      = btrim(sqlc.arg(unit_kerja)::text),
    employment_type = sqlc.arg(employment_type)::text,
    tanggal_lahir   = sqlc.arg(tanggal_lahir)::date,
    jenis_kelamin   = NULLIF(btrim(sqlc.arg(jenis_kelamin)::text), ''),
    tempat_lahir    = NULLIF(btrim(sqlc.arg(tempat_lahir)::text), ''),
    is_active       = sqlc.arg(is_active)::boolean,
    updated_at      = NOW()
WHERE id = sqlc.arg(id)::uuid
RETURNING id;

-- name: DeleteEmployee :exec
DELETE FROM employees WHERE id = $1;

-- name: CountEmployees :one
SELECT COUNT(*) FROM employees;

-- name: ListEmployeesWithStatus :many
SELECT e.id, e.pegawai_uid, COALESCE(e.nip, '')::text AS nip, e.nama, e.unit_kerja, e.employment_type, e.tanggal_lahir,
  COALESCE(e.jenis_kelamin, '')::text AS jenis_kelamin,
  COALESCE(e.tempat_lahir, '')::text AS tempat_lahir,
  COALESCE(pa.pusaka_username, '') AS pusaka_username, COALESCE(pa.is_enabled, FALSE) AS pusaka_is_enabled, e.is_active, e.created_at,
  COALESCE(
    (SELECT j.status::text FROM jobs j
     WHERE j.employee_id = e.id AND (j.status = 'queued' OR j.status = 'running')
     ORDER BY CASE j.status::text WHEN 'running' THEN 0 ELSE 1 END, j.created_at DESC
     LIMIT 1), '') AS active_status,
  COALESCE(
    (SELECT j.run_type::text FROM jobs j
     WHERE j.employee_id = e.id AND (j.status = 'queued' OR j.status = 'running')
     ORDER BY CASE j.status::text WHEN 'running' THEN 0 ELSE 1 END, j.created_at DESC
     LIMIT 1), '') AS active_run_type,
  COALESCE(
    (SELECT j.status::text FROM jobs j
     WHERE j.employee_id = e.id ORDER BY j.created_at DESC LIMIT 1), '') AS last_status,
  COALESCE(
    (SELECT j.run_type::text FROM jobs j
     WHERE j.employee_id = e.id ORDER BY j.created_at DESC LIMIT 1), '') AS last_run_type,
  EXISTS(
    SELECT 1 FROM employee_schedules es
    WHERE es.employee_id = e.id AND es.run_type = 'checkin' AND es.is_enabled = TRUE
  ) AS has_checkin_schedule,
  EXISTS(
    SELECT 1 FROM employee_schedules es
    WHERE es.employee_id = e.id AND es.run_type = 'checkout' AND es.is_enabled = TRUE
  ) AS has_checkout_schedule,
  (e.employment_type IN ('pns', 'pppk')) AS pusaka_eligible,
  COALESCE((pa.employee_id IS NOT NULL AND pa.pusaka_username <> ''), FALSE) AS has_pusaka_account
FROM employees e
LEFT JOIN pusaka_accounts pa ON pa.employee_id = e.id
ORDER BY e.created_at DESC;

-- name: ListPusakaEligibleEmployeesWithStatus :many
SELECT e.id, e.pegawai_uid, COALESCE(e.nip, '')::text AS nip, e.nama, e.unit_kerja, e.employment_type, e.tanggal_lahir,
  COALESCE(e.jenis_kelamin, '')::text AS jenis_kelamin,
  COALESCE(e.tempat_lahir, '')::text AS tempat_lahir,
  COALESCE(pa.pusaka_username, '') AS pusaka_username, COALESCE(pa.is_enabled, FALSE) AS pusaka_is_enabled, e.is_active, e.created_at,
  COALESCE(
    (SELECT j.status::text FROM jobs j
     WHERE j.employee_id = e.id AND (j.status = 'queued' OR j.status = 'running')
     ORDER BY CASE j.status::text WHEN 'running' THEN 0 ELSE 1 END, j.created_at DESC
     LIMIT 1), '') AS active_status,
  COALESCE(
    (SELECT j.run_type::text FROM jobs j
     WHERE j.employee_id = e.id AND (j.status = 'queued' OR j.status = 'running')
     ORDER BY CASE j.status::text WHEN 'running' THEN 0 ELSE 1 END, j.created_at DESC
     LIMIT 1), '') AS active_run_type,
  COALESCE(
    (SELECT j.status::text FROM jobs j
     WHERE j.employee_id = e.id ORDER BY j.created_at DESC LIMIT 1), '') AS last_status,
  COALESCE(
    (SELECT j.run_type::text FROM jobs j
     WHERE j.employee_id = e.id ORDER BY j.created_at DESC LIMIT 1), '') AS last_run_type,
  EXISTS(
    SELECT 1 FROM employee_schedules es
    WHERE es.employee_id = e.id AND es.run_type = 'checkin' AND es.is_enabled = TRUE
  ) AS has_checkin_schedule,
  EXISTS(
    SELECT 1 FROM employee_schedules es
    WHERE es.employee_id = e.id AND es.run_type = 'checkout' AND es.is_enabled = TRUE
  ) AS has_checkout_schedule,
  TRUE AS pusaka_eligible,
  COALESCE((pa.employee_id IS NOT NULL AND pa.pusaka_username <> ''), FALSE) AS has_pusaka_account
FROM employees e
LEFT JOIN pusaka_accounts pa ON pa.employee_id = e.id
WHERE e.employment_type IN ('pns', 'pppk')
ORDER BY e.created_at DESC;
