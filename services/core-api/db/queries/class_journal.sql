-- name: CountJournalSessionsForAssignment :one
SELECT COUNT(*)::int FROM class_journal_sessions
WHERE assignment_id = $1;

-- name: LockJournalAssignmentForUpdate :one
SELECT id
FROM class_subject_assignments
WHERE id = $1
FOR UPDATE;

-- name: NextJournalMeetingNumber :one
SELECT (COALESCE(MAX(pertemuan_ke), 0) + 1)::int
FROM class_journal_sessions
WHERE assignment_id = $1;

-- name: CreateJournalSession :one
INSERT INTO class_journal_sessions
    (assignment_id, timetable_slot_id, tanggal, pertemuan_ke, materi, kegiatan, catatan, guru_hadir)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, assignment_id, tanggal, pertemuan_ke, materi, kegiatan, catatan, guru_hadir, created_at, updated_at, timetable_slot_id;

-- name: GetJournalSession :one
SELECT
    s.id, s.assignment_id, s.timetable_slot_id, s.tanggal, s.pertemuan_ke,
    s.materi, s.kegiatan, s.catatan, s.guru_hadir,
    s.created_at, s.updated_at,
    csa.class_id, c.name AS class_name, c.code AS class_code,
    csa.subject_id, subj.name AS subject_name,
    csa.teacher_employee_id, e.nama AS teacher_name
FROM class_journal_sessions s
JOIN class_subject_assignments csa ON csa.id = s.assignment_id
JOIN school_classes c ON c.id = csa.class_id
JOIN subjects subj ON subj.id = csa.subject_id
JOIN employees e ON e.id = csa.teacher_employee_id
WHERE s.id = $1;

-- name: GetJournalSessionIDByAssignmentDate :one
SELECT id
FROM class_journal_sessions
WHERE assignment_id = $1
  AND tanggal = $2
  AND timetable_slot_id IS NULL;

-- name: GetJournalSessionIDByTimetableSlotDate :one
SELECT id
FROM class_journal_sessions
WHERE timetable_slot_id = $1
  AND tanggal = $2;

-- name: ListJournalSessions :many
SELECT
    s.id, s.assignment_id, s.timetable_slot_id, s.tanggal, s.pertemuan_ke,
    s.materi, s.kegiatan, s.catatan, s.guru_hadir,
    s.created_at, s.updated_at,
    c.name AS class_name, c.code AS class_code,
    subj.name AS subject_name,
    e.nama AS teacher_name
FROM class_journal_sessions s
JOIN class_subject_assignments csa ON csa.id = s.assignment_id
JOIN school_classes c ON c.id = csa.class_id
JOIN subjects subj ON subj.id = csa.subject_id
JOIN employees e ON e.id = csa.teacher_employee_id
WHERE s.assignment_id = $1
ORDER BY s.tanggal DESC, s.pertemuan_ke DESC;

-- name: UpdateJournalSession :one
UPDATE class_journal_sessions
SET materi = $2, kegiatan = $3, catatan = $4, guru_hadir = $5, updated_at = NOW()
WHERE id = $1
RETURNING id, assignment_id, tanggal, pertemuan_ke, materi, kegiatan, catatan, guru_hadir, created_at, updated_at, timetable_slot_id;

-- name: DeleteJournalSession :exec
DELETE FROM class_journal_sessions WHERE id = $1;

-- name: InsertJournalEditLog :exec
INSERT INTO journal_edit_logs (session_id, edited_by, changes)
VALUES ($1, $2, $3);

-- name: ListJournalEditLogs :many
SELECT
    l.id, l.session_id, l.edited_by, l.edited_at, l.changes,
    e.nama AS edited_by_name
FROM journal_edit_logs l
JOIN employees e ON e.id = l.edited_by
WHERE l.session_id = $1
ORDER BY l.edited_at DESC;

-- name: UpsertJournalAttendance :one
INSERT INTO class_journal_attendances (session_id, student_id, status, catatan)
SELECT
    sqlc.arg(session_id)::uuid,
    st.id,
    sqlc.arg(status)::journal_attendance_status,
    sqlc.arg(catatan)::text
FROM class_journal_sessions s
JOIN class_subject_assignments csa ON csa.id = s.assignment_id
JOIN students st ON st.id = sqlc.arg(student_id)::uuid
    AND st.class_id = csa.class_id
    AND st.is_active = TRUE
WHERE s.id = sqlc.arg(session_id)::uuid
ON CONFLICT (session_id, student_id) DO UPDATE
    SET status     = EXCLUDED.status,
        catatan    = EXCLUDED.catatan,
        updated_at = NOW()
RETURNING id, session_id, student_id, status, catatan, created_at, updated_at;

-- name: ListJournalAttendances :many
SELECT
    a.id, a.session_id, a.student_id, a.status, a.catatan,
    a.created_at, a.updated_at,
    st.nis, st.nisn, st.nama, st.gender
FROM class_journal_attendances a
JOIN students st ON st.id = a.student_id
WHERE a.session_id = $1
ORDER BY st.nama ASC;

-- name: ListJournalAttendanceSummary :many
SELECT
    st.id          AS student_id,
    st.nis,
    st.nisn,
    st.nama,
    COUNT(js.id)::int                                    AS total_pertemuan,
    COUNT(*) FILTER (WHERE a.status = 'hadir')::int      AS hadir,
    COUNT(*) FILTER (WHERE a.status = 'sakit')::int      AS sakit,
    COUNT(*) FILTER (WHERE a.status = 'izin')::int       AS izin,
    COUNT(*) FILTER (WHERE a.status = 'alpha')::int      AS alpha
FROM class_subject_assignments csa
JOIN students st ON st.class_id = csa.class_id AND st.is_active = TRUE
LEFT JOIN class_journal_sessions js   ON js.assignment_id = csa.id
LEFT JOIN class_journal_attendances a ON a.session_id = js.id AND a.student_id = st.id
WHERE csa.id = $1
GROUP BY st.id, st.nis, st.nisn, st.nama
ORDER BY st.nama ASC;
