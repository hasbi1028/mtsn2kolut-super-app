-- name: ListEmployees :many
SELECT id, nip, nama, unit_kerja, pusaka_username, pusaka_password, is_active, created_at, updated_at
FROM employees
ORDER BY nama ASC;

-- name: ListActiveEmployees :many
SELECT id, nip, nama, unit_kerja, pusaka_username, pusaka_password, is_active, created_at, updated_at
FROM employees
WHERE is_active = TRUE
ORDER BY nama ASC;

-- name: GetEmployee :one
SELECT id, nip, nama, unit_kerja, pusaka_username, pusaka_password, is_active, created_at, updated_at
FROM employees
WHERE id = $1;

-- name: CreateEmployee :one
INSERT INTO employees (id, nip, nama, unit_kerja, pusaka_username, pusaka_password, is_active)
VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateEmployee :one
UPDATE employees
SET nip             = $2,
    nama            = $3,
    unit_kerja      = $4,
    pusaka_username = $5,
    pusaka_password = $6,
    is_active       = $7,
    updated_at      = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteEmployee :exec
DELETE FROM employees WHERE id = $1;

-- name: CountEmployees :one
SELECT COUNT(*) FROM employees;

-- name: ListEmployeesWithStatus :many
SELECT e.id, e.nip, e.nama, e.unit_kerja, e.pusaka_username, e.is_active, e.created_at,
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
  ) AS has_checkout_schedule
FROM employees e
ORDER BY e.created_at DESC;
