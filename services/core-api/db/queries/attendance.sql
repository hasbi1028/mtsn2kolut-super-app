-- name: ListAttendance :many
SELECT ar.id, ar.employee_id, e.nama AS employee_nama, e.nip AS employee_nip,
       ar.tanggal, ar.jam_masuk, ar.jam_pulang, ar.source_job_id,
       ar.created_at, ar.updated_at
FROM attendance_records ar
JOIN employees e ON e.id = ar.employee_id
ORDER BY ar.tanggal DESC, e.nama ASC
LIMIT $1 OFFSET $2;

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
