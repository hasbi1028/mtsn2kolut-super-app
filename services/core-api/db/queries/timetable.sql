-- name: ListTimetableSlots :many
SELECT ts.id, ts.assignment_id, ts.day_of_week, ts.start_time, ts.end_time, ts.room_label, ts.notes,
       ts.created_at, ts.updated_at,
       a.class_id, c.name AS class_name, c.code AS class_code,
       a.subject_id, s.name AS subject_name, s.code AS subject_code,
       a.teacher_employee_id, e.nama AS teacher_name
FROM timetable_slots ts
JOIN class_subject_assignments a ON a.id = ts.assignment_id
JOIN school_classes c ON c.id = a.class_id
JOIN subjects s ON s.id = a.subject_id
JOIN employees e ON e.id = a.teacher_employee_id
ORDER BY ts.day_of_week ASC, ts.start_time ASC, c.name ASC, s.name ASC;

-- name: ListStudentTimetable :many
SELECT ts.id, ts.assignment_id, ts.day_of_week, ts.start_time, ts.end_time, ts.room_label, ts.notes,
       c.id AS class_id, c.name AS class_name, c.code AS class_code,
       s.id AS subject_id, s.name AS subject_name, s.code AS subject_code,
       e.id AS teacher_employee_id, e.nama AS teacher_name
FROM timetable_slots ts
JOIN class_subject_assignments a ON a.id = ts.assignment_id
JOIN school_classes c ON c.id = a.class_id
JOIN subjects s ON s.id = a.subject_id
JOIN employees e ON e.id = a.teacher_employee_id
JOIN students st ON st.class_id = c.id
WHERE st.id = $1
ORDER BY ts.day_of_week ASC, ts.start_time ASC, s.name ASC;

-- name: ListTeacherTimetable :many
SELECT ts.id, ts.assignment_id, ts.day_of_week, ts.start_time, ts.end_time, ts.room_label, ts.notes,
       c.id AS class_id, c.name AS class_name, c.code AS class_code,
       s.id AS subject_id, s.name AS subject_name, s.code AS subject_code,
       e.id AS teacher_employee_id, e.nama AS teacher_name
FROM timetable_slots ts
JOIN class_subject_assignments a ON a.id = ts.assignment_id
JOIN school_classes c ON c.id = a.class_id
JOIN subjects s ON s.id = a.subject_id
JOIN employees e ON e.id = a.teacher_employee_id
WHERE a.teacher_employee_id = $1
ORDER BY ts.day_of_week ASC, ts.start_time ASC, c.name ASC, s.name ASC;

-- name: GetTimetableSlot :one
SELECT id, assignment_id, day_of_week, start_time, end_time, room_label, notes, created_at, updated_at
FROM timetable_slots
WHERE id = $1;

-- name: CreateTimetableSlot :one
INSERT INTO timetable_slots (
    assignment_id, day_of_week, start_time, end_time, room_label, notes
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateTimetableSlot :one
UPDATE timetable_slots
SET assignment_id = $2,
    day_of_week = $3,
    start_time = $4,
    end_time = $5,
    room_label = $6,
    notes = $7,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CountTimetableConflicts :one
SELECT COUNT(*)::int
FROM timetable_slots ts
JOIN class_subject_assignments a ON a.id = ts.assignment_id
WHERE ts.day_of_week = sqlc.arg(day_of_week)
  AND ts.start_time < sqlc.arg(end_time)
  AND ts.end_time > sqlc.arg(start_time)
  AND (
    a.class_id = sqlc.arg(class_id)
    OR a.teacher_employee_id = sqlc.arg(teacher_employee_id)
  )
  AND (
    sqlc.arg(exclude_slot_id)::uuid IS NULL
    OR ts.id <> sqlc.arg(exclude_slot_id)
  );

-- name: CountTimetableRoomConflicts :one
SELECT COUNT(*)::int
FROM timetable_slots ts
WHERE ts.day_of_week = sqlc.arg(day_of_week)
  AND ts.start_time < sqlc.arg(end_time)
  AND ts.end_time > sqlc.arg(start_time)
  AND LOWER(TRIM(ts.room_label)) = LOWER(TRIM(sqlc.arg(room_label)))
  AND TRIM(sqlc.arg(room_label)) <> ''
  AND (
    sqlc.arg(exclude_slot_id)::uuid IS NULL
    OR ts.id <> sqlc.arg(exclude_slot_id)
  );

-- name: DeleteTimetableSlot :exec
DELETE FROM timetable_slots
WHERE id = $1;
