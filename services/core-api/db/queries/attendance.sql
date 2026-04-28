-- name: ListAttendance :many
SELECT ar.id, ar.employee_id, e.nama AS employee_nama, e.nip AS employee_nip,
       ar.tanggal, ar.jam_masuk, ar.jam_pulang, ar.source_job_id,
       ar.created_at, ar.updated_at
FROM attendance_records ar
JOIN employees e ON e.id = ar.employee_id
ORDER BY ar.tanggal DESC, e.nama ASC
LIMIT $1 OFFSET $2;

-- name: ListAttendanceInRange :many
SELECT ar.id, ar.employee_id, e.nama AS employee_nama, e.nip AS employee_nip,
       ar.tanggal, ar.jam_masuk, ar.jam_pulang, ar.source_job_id,
       ar.created_at, ar.updated_at
FROM attendance_records ar
JOIN employees e ON e.id = ar.employee_id
WHERE ar.tanggal >= $1 AND ar.tanggal <= $2
ORDER BY ar.tanggal DESC, e.nama ASC;

-- name: CountAttendanceInRange :one
SELECT COUNT(*) FROM attendance_records
WHERE tanggal >= $1 AND tanggal <= $2;

-- name: ListAttendanceByDate :many
SELECT ar.id, ar.employee_id, e.nama AS employee_nama, e.nip AS employee_nip,
       ar.tanggal, ar.jam_masuk, ar.jam_pulang, ar.source_job_id,
       ar.created_at, ar.updated_at
FROM attendance_records ar
JOIN employees e ON e.id = ar.employee_id
WHERE ar.tanggal = $1
ORDER BY e.nama ASC;

-- name: ListAttendanceByEmployee :many
SELECT ar.id, ar.employee_id, e.nama AS employee_nama, e.nip AS employee_nip,
       ar.tanggal, ar.jam_masuk, ar.jam_pulang, ar.source_job_id,
       ar.created_at, ar.updated_at
FROM attendance_records ar
JOIN employees e ON e.id = ar.employee_id
WHERE ar.employee_id = $1
ORDER BY ar.tanggal DESC
LIMIT $2 OFFSET $3;

-- name: GetMonthlyAttendanceSummary :many
SELECT 
    e.id AS employee_id,
    e.nama AS employee_nama,
    e.nip AS employee_nip,
    COUNT(ar.id)::int AS total_days,
    COUNT(CASE WHEN ar.jam_masuk != '' AND ar.jam_pulang != '' THEN 1 END)::int AS complete_days,
    COUNT(CASE WHEN ar.jam_masuk != '' AND ar.jam_pulang = '' THEN 1 END)::int AS missing_checkout,
    COUNT(CASE WHEN ar.jam_masuk = '' AND ar.jam_pulang != '' THEN 1 END)::int AS missing_checkin
FROM employees e
LEFT JOIN attendance_records ar ON ar.employee_id = e.id 
    AND ar.tanggal >= $1 AND ar.tanggal <= $2
WHERE e.is_active = TRUE
GROUP BY e.id, e.nama, e.nip
ORDER BY e.nama ASC;

-- name: UpsertAttendance :one
INSERT INTO attendance_records (id, employee_id, tanggal, jam_masuk, jam_pulang, source_job_id)
VALUES (gen_random_uuid(), $1, $2, $3, $4, $5)
ON CONFLICT (employee_id, tanggal) DO UPDATE
  SET jam_masuk     = CASE WHEN EXCLUDED.jam_masuk  != '' THEN EXCLUDED.jam_masuk  ELSE attendance_records.jam_masuk  END,
      jam_pulang    = CASE WHEN EXCLUDED.jam_pulang != '' THEN EXCLUDED.jam_pulang ELSE attendance_records.jam_pulang END,
      source_job_id = EXCLUDED.source_job_id,
      updated_at    = NOW()
RETURNING *;

-- name: CountAttendance :one
SELECT COUNT(*) FROM attendance_records;
