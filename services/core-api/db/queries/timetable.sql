-- name: ListTimetableSlots :many
SELECT ts.id, ts.assignment_id, ts.day_of_week, ts.lesson_period_id, ts.slot_type, ts.lesson_hours::float8 AS lesson_hours,
       ts.start_time, ts.end_time, ts.room_label, ts.notes,
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

-- name: ListWeeklyTimetableClasses :many
SELECT id, code, name, level
FROM school_classes
WHERE academic_year_id = $1
  AND is_active = TRUE
ORDER BY level ASC, name ASC;

-- name: ListWeeklyTimetableTeachers :many
SELECT
    id,
    COALESCE(nip, '')::text AS nip,
    nama,
    unit_kerja
FROM employees
WHERE is_active = TRUE
ORDER BY nama ASC;

-- name: ListWeeklyTimetableSubjects :many
SELECT
    id,
    code,
    name,
    category,
    is_assessment_subject,
    is_report_subject,
    is_schedule_activity,
    default_weekly_hours,
    display_order
FROM subjects
WHERE is_active = TRUE
ORDER BY display_order ASC, name ASC;

-- name: ListWeeklyTimetableAssignments :many
WITH active_classes AS (
    SELECT id, code, name, level
    FROM school_classes
    WHERE academic_year_id = $1
      AND is_active = TRUE
),
active_class_curriculum AS (
    SELECT DISTINCT ON (cca.class_id)
        cca.class_id,
        cca.curriculum_profile_id
    FROM class_curriculum_assignments cca
    JOIN active_classes c ON c.id = cca.class_id
    WHERE cca.is_active = TRUE
    ORDER BY cca.class_id, cca.updated_at DESC, cca.created_at DESC
)
SELECT
    a.id,
    a.class_id,
    c.code AS class_code,
    c.name AS class_name,
    c.level AS class_level,
    a.subject_id,
    s.code AS subject_code,
    s.name AS subject_name,
    s.category AS subject_category,
    s.is_schedule_activity,
    s.default_weekly_hours,
    a.teacher_employee_id,
    e.nama AS teacher_name,
    COALESCE(e.nip, '')::text AS teacher_nip,
    COALESCE(ov.intra_weekly_hours, alloc.intra_weekly_hours, CASE WHEN s.is_schedule_activity THEN 0 ELSE s.default_weekly_hours::numeric END, 0)::float8 AS expected_intra_weekly_hours,
    COALESCE(ov.koku_weekly_hours, alloc.koku_weekly_hours, CASE WHEN s.is_schedule_activity THEN s.default_weekly_hours::numeric ELSE 0 END, 0)::float8 AS expected_koku_weekly_hours,
    COALESCE(ov.additional_weekly_hours, 0)::float8 AS expected_additional_weekly_hours,
    COALESCE(ov.total_weekly_hours, alloc.total_weekly_hours, s.default_weekly_hours::numeric, 0)::float8 AS expected_weekly_hours
FROM class_subject_assignments a
JOIN active_classes c ON c.id = a.class_id
JOIN subjects s ON s.id = a.subject_id
JOIN employees e ON e.id = a.teacher_employee_id
LEFT JOIN active_class_curriculum acc ON acc.class_id = c.id
LEFT JOIN curriculum_subject_allocations alloc
  ON alloc.curriculum_profile_id = acc.curriculum_profile_id
 AND alloc.level = c.level
 AND alloc.subject_id = a.subject_id
LEFT JOIN class_subject_allocation_overrides ov ON ov.assignment_id = a.id
WHERE s.is_active = TRUE
  AND e.is_active = TRUE
ORDER BY c.level ASC, c.name ASC, s.display_order ASC, s.name ASC, e.nama ASC;

-- name: ListWeeklyTimetableSlots :many
SELECT
    ts.id,
    ts.assignment_id,
    ts.day_of_week,
    ts.lesson_period_id,
    COALESCE(lpt.period_number, 0)::int AS period_number,
    ts.slot_type,
    ts.lesson_hours::float8 AS lesson_hours,
    COALESCE(lpt.label, '')::text AS lesson_period_label,
    COALESCE(lpt.activity_type, ts.slot_type)::text AS lesson_period_activity_type,
    COALESCE(lpt.is_counted_as_lesson, ts.lesson_hours > 0)::boolean AS is_counted_as_lesson,
    ts.start_time,
    ts.end_time,
    ts.room_label,
    ts.notes,
    ts.created_at,
    ts.updated_at,
    a.class_id,
    c.name AS class_name,
    c.code AS class_code,
    c.level AS class_level,
    a.subject_id,
    s.name AS subject_name,
    s.code AS subject_code,
    s.category AS subject_category,
    s.is_schedule_activity,
    a.teacher_employee_id,
    e.nama AS teacher_name
FROM timetable_slots ts
JOIN class_subject_assignments a ON a.id = ts.assignment_id
JOIN school_classes c ON c.id = a.class_id
JOIN subjects s ON s.id = a.subject_id
JOIN employees e ON e.id = a.teacher_employee_id
LEFT JOIN lesson_period_templates lpt ON lpt.id = ts.lesson_period_id
WHERE c.academic_year_id = $1
  AND c.is_active = TRUE
ORDER BY ts.day_of_week ASC, ts.start_time ASC, c.level ASC, c.name ASC, s.name ASC;

-- name: ListLessonPeriodTemplates :many
SELECT lpt.id, lpt.academic_year_id, ay.name AS academic_year_name,
       lpt.day_of_week, lpt.period_number, lpt.start_time, lpt.end_time,
       lpt.activity_type, lpt.label, lpt.is_counted_as_lesson,
       lpt.created_at, lpt.updated_at
FROM lesson_period_templates lpt
JOIN academic_years ay ON ay.id = lpt.academic_year_id
WHERE lpt.academic_year_id = $1
ORDER BY lpt.day_of_week ASC, lpt.period_number ASC, lpt.start_time ASC;

-- name: ListLessonPeriodTemplatesForDay :many
SELECT id, academic_year_id, day_of_week, period_number, start_time, end_time,
       activity_type, label, is_counted_as_lesson, created_at, updated_at
FROM lesson_period_templates
WHERE academic_year_id = sqlc.arg(academic_year_id)
  AND day_of_week = sqlc.arg(day_of_week)
ORDER BY period_number ASC, start_time ASC;

-- name: GetLessonPeriodTemplate :one
SELECT id, academic_year_id, day_of_week, period_number, start_time, end_time,
       activity_type, label, is_counted_as_lesson, created_at, updated_at
FROM lesson_period_templates
WHERE id = $1;

-- name: CountLessonPeriodTemplatesForDay :one
SELECT COUNT(*)::int
FROM lesson_period_templates
WHERE academic_year_id = sqlc.arg(academic_year_id)
  AND day_of_week = sqlc.arg(day_of_week);

-- name: CreateLessonPeriodTemplate :one
INSERT INTO lesson_period_templates (
    academic_year_id, day_of_week, period_number, start_time, end_time,
    activity_type, label, is_counted_as_lesson
)
VALUES (
    sqlc.arg(academic_year_id), sqlc.arg(day_of_week), sqlc.arg(period_number),
    sqlc.arg(start_time), sqlc.arg(end_time), sqlc.arg(activity_type),
    sqlc.arg(label), sqlc.arg(is_counted_as_lesson)
)
RETURNING *;

-- name: UpdateLessonPeriodTemplate :one
UPDATE lesson_period_templates
SET day_of_week = sqlc.arg(day_of_week),
    period_number = sqlc.arg(period_number),
    start_time = sqlc.arg(start_time),
    end_time = sqlc.arg(end_time),
    activity_type = sqlc.arg(activity_type),
    label = sqlc.arg(label),
    is_counted_as_lesson = sqlc.arg(is_counted_as_lesson),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteLessonPeriodTemplate :exec
DELETE FROM lesson_period_templates
WHERE id = $1;

-- name: ListTimetableConflicts :many
WITH scoped_slots AS (
    SELECT
        ts.id,
        ts.assignment_id,
        ts.day_of_week,
        ts.start_time,
        ts.end_time,
        ts.room_label,
        a.class_id,
        c.name AS class_name,
        c.code AS class_code,
        c.level AS class_level,
        a.subject_id,
        s.name AS subject_name,
        s.code AS subject_code,
        a.teacher_employee_id,
        e.nama AS teacher_name,
        LOWER(TRIM(ts.room_label)) AS room_key
    FROM timetable_slots ts
    JOIN class_subject_assignments a ON a.id = ts.assignment_id
    JOIN school_classes c ON c.id = a.class_id
    JOIN subjects s ON s.id = a.subject_id
    JOIN employees e ON e.id = a.teacher_employee_id
    WHERE c.academic_year_id = $1
      AND c.is_active = TRUE
),
conflicts AS (
    SELECT
        s.id AS slot_id,
        NULL::uuid AS related_slot_id,
        'invalid_time_range'::text AS conflict_type,
        'Rentang waktu slot tidak valid'::text AS message
    FROM scoped_slots s
    WHERE s.start_time >= s.end_time

    UNION ALL

    SELECT
        a.id AS slot_id,
        b.id AS related_slot_id,
        'same_class'::text AS conflict_type,
        ('Rombel ' || a.class_code || ' memiliki slot tumpang tindih')::text AS message
    FROM scoped_slots a
    JOIN scoped_slots b ON a.id::text < b.id::text
     AND a.day_of_week = b.day_of_week
     AND a.start_time < b.end_time
     AND a.end_time > b.start_time
     AND a.class_id = b.class_id

    UNION ALL

    SELECT
        a.id AS slot_id,
        b.id AS related_slot_id,
        'same_teacher'::text AS conflict_type,
        ('Guru ' || a.teacher_name || ' mengajar pada slot yang tumpang tindih')::text AS message
    FROM scoped_slots a
    JOIN scoped_slots b ON a.id::text < b.id::text
     AND a.day_of_week = b.day_of_week
     AND a.start_time < b.end_time
     AND a.end_time > b.start_time
     AND a.teacher_employee_id = b.teacher_employee_id

    UNION ALL

    SELECT
        a.id AS slot_id,
        b.id AS related_slot_id,
        'same_room'::text AS conflict_type,
        ('Ruang ' || a.room_label || ' dipakai pada slot yang tumpang tindih')::text AS message
    FROM scoped_slots a
    JOIN scoped_slots b ON a.id::text < b.id::text
     AND a.day_of_week = b.day_of_week
     AND a.start_time < b.end_time
     AND a.end_time > b.start_time
     AND a.room_key <> ''
     AND a.room_key = b.room_key
)
SELECT
    c.conflict_type,
    c.message,
    s.id AS slot_id,
    s.assignment_id,
    s.day_of_week,
    s.start_time,
    s.end_time,
    s.room_label,
    s.class_id,
    s.class_code,
    s.class_name,
    s.class_level,
    s.subject_id,
    s.subject_code,
    s.subject_name,
    s.teacher_employee_id,
    s.teacher_name,
    c.related_slot_id,
    COALESCE(r.assignment_id, '00000000-0000-0000-0000-000000000000'::uuid) AS related_assignment_id,
    COALESCE(r.day_of_week, 0)::smallint AS related_day_of_week,
    COALESCE(r.start_time, '00:00'::time) AS related_start_time,
    COALESCE(r.end_time, '00:00'::time) AS related_end_time,
    COALESCE(r.room_label, '')::text AS related_room_label,
    COALESCE(r.class_id, '00000000-0000-0000-0000-000000000000'::uuid) AS related_class_id,
    COALESCE(r.class_code, '')::text AS related_class_code,
    COALESCE(r.class_name, '')::text AS related_class_name,
    COALESCE(r.class_level, '')::text AS related_class_level,
    COALESCE(r.subject_id, '00000000-0000-0000-0000-000000000000'::uuid) AS related_subject_id,
    COALESCE(r.subject_code, '')::text AS related_subject_code,
    COALESCE(r.subject_name, '')::text AS related_subject_name,
    COALESCE(r.teacher_employee_id, '00000000-0000-0000-0000-000000000000'::uuid) AS related_teacher_employee_id,
    COALESCE(r.teacher_name, '')::text AS related_teacher_name
FROM conflicts c
JOIN scoped_slots s ON s.id = c.slot_id
LEFT JOIN scoped_slots r ON r.id = c.related_slot_id
ORDER BY s.day_of_week ASC, s.start_time ASC, c.conflict_type ASC, s.class_name ASC, s.subject_name ASC;

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

-- name: ListParentChildrenTimetable :many
SELECT st.id AS student_id, st.nama AS student_name,
       ts.id, ts.assignment_id, ts.day_of_week, ts.start_time, ts.end_time, ts.room_label, ts.notes,
       c.id AS class_id, c.name AS class_name, c.code AS class_code,
       s.id AS subject_id, s.name AS subject_name, s.code AS subject_code,
       e.id AS teacher_employee_id, e.nama AS teacher_name
FROM parent_students ps
JOIN students st ON st.id = ps.student_id
JOIN school_classes c ON c.id = st.class_id
JOIN class_subject_assignments a ON a.class_id = c.id
JOIN timetable_slots ts ON ts.assignment_id = a.id
JOIN subjects s ON s.id = a.subject_id
JOIN employees e ON e.id = a.teacher_employee_id
WHERE ps.parent_id = $1
ORDER BY st.nama ASC, ts.day_of_week ASC, ts.start_time ASC, s.name ASC;

-- name: GetTimetableSlot :one
SELECT id, assignment_id, day_of_week, start_time, end_time, room_label, notes,
       created_at, updated_at, lesson_period_id, slot_type, lesson_hours
FROM timetable_slots
WHERE id = $1;

-- name: CreateTimetableSlot :one
INSERT INTO timetable_slots (
    assignment_id, day_of_week, start_time, end_time, room_label, notes,
    lesson_period_id, slot_type, lesson_hours
)
VALUES (
    sqlc.arg(assignment_id), sqlc.arg(day_of_week), sqlc.arg(start_time), sqlc.arg(end_time),
    sqlc.arg(room_label), sqlc.arg(notes), sqlc.narg(lesson_period_id), sqlc.arg(slot_type),
    sqlc.arg(lesson_hours)
)
RETURNING *;

-- name: UpdateTimetableSlot :one
UPDATE timetable_slots
SET assignment_id = sqlc.arg(assignment_id),
    day_of_week = sqlc.arg(day_of_week),
    start_time = sqlc.arg(start_time),
    end_time = sqlc.arg(end_time),
    room_label = sqlc.arg(room_label),
    notes = sqlc.arg(notes),
    lesson_period_id = sqlc.narg(lesson_period_id),
    slot_type = sqlc.arg(slot_type),
    lesson_hours = sqlc.arg(lesson_hours),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
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

-- name: LockTimetableMutationScope :one
SELECT 1::bigint
FROM pg_advisory_xact_lock(hashtextextended(sqlc.arg(lock_key)::text, 0));

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
