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

-- name: CreateTimetableSlot :one
INSERT INTO timetable_slots (
    assignment_id, day_of_week, start_time, end_time, room_label, notes
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: DeleteTimetableSlot :exec
DELETE FROM timetable_slots
WHERE id = $1;
