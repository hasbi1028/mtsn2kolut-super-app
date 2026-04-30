-- name: ListEmployees :many
SELECT e.id, e.nip, e.nama, e.unit_kerja, e.employment_type,
       COALESCE(pa.pusaka_username, '') AS pusaka_username,
       COALESCE(pa.pusaka_password, '') AS pusaka_password,
       COALESCE(pa.is_enabled, FALSE) AS pusaka_is_enabled,
       e.is_active, e.created_at, e.updated_at
FROM employees e
LEFT JOIN pusaka_accounts pa ON pa.employee_id = e.id
ORDER BY nama ASC;

-- name: ListActiveEmployees :many
SELECT e.id, e.nip, e.nama, e.unit_kerja, e.employment_type,
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
SELECT e.id, e.nip, e.nama, e.unit_kerja, e.employment_type,
       COALESCE(pa.pusaka_username, '') AS pusaka_username,
       COALESCE(pa.pusaka_password, '') AS pusaka_password,
       COALESCE(pa.is_enabled, FALSE) AS pusaka_is_enabled,
       e.is_active, e.created_at, e.updated_at
FROM employees e
LEFT JOIN pusaka_accounts pa ON pa.employee_id = e.id
WHERE e.id = $1;

-- name: CreateEmployee :one
INSERT INTO employees (id, nip, nama, unit_kerja, employment_type, is_active)
VALUES (gen_random_uuid(), $1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateEmployee :one
UPDATE employees
SET nip             = $2,
    nama            = $3,
    unit_kerja      = $4,
    employment_type = $5,
    is_active       = $6,
    updated_at      = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteEmployee :exec
DELETE FROM employees WHERE id = $1;

-- name: CountEmployees :one
SELECT COUNT(*) FROM employees;

-- name: ListEmployeesWithStatus :many
SELECT e.id, e.nip, e.nama, e.unit_kerja, e.employment_type, COALESCE(pa.pusaka_username, '') AS pusaka_username, COALESCE(pa.is_enabled, FALSE) AS pusaka_is_enabled, e.is_active, e.created_at,
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
SELECT e.id, e.nip, e.nama, e.unit_kerja, e.employment_type, COALESCE(pa.pusaka_username, '') AS pusaka_username, COALESCE(pa.is_enabled, FALSE) AS pusaka_is_enabled, e.is_active, e.created_at,
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
