-- name: ListSchoolRooms :many
SELECT id, code, name, building, floor, room_type, location_note,
       default_capacity, exam_capacity, condition, is_exam_eligible,
       network_ready, power_ready, notes, created_at, updated_at
FROM school_rooms
WHERE (
    sqlc.arg(search)::TEXT = '' OR
    code ILIKE '%' || sqlc.arg(search) || '%' OR
    name ILIKE '%' || sqlc.arg(search) || '%' OR
    building ILIKE '%' || sqlc.arg(search) || '%' OR
    location_note ILIKE '%' || sqlc.arg(search) || '%'
) AND (
    sqlc.arg(room_type)::TEXT = '' OR room_type = sqlc.arg(room_type)
) AND (
    sqlc.arg(condition)::TEXT = '' OR condition = sqlc.arg(condition)
) AND (
    sqlc.arg(exam_eligible)::TEXT = '' OR is_exam_eligible = (sqlc.arg(exam_eligible)::TEXT = 'true')
)
ORDER BY name ASC;

-- name: GetSchoolRoom :one
SELECT id, code, name, building, floor, room_type, location_note,
       default_capacity, exam_capacity, condition, is_exam_eligible,
       network_ready, power_ready, notes, created_at, updated_at
FROM school_rooms
WHERE id = $1;

-- name: CreateSchoolRoom :one
INSERT INTO school_rooms (
    code, name, building, floor, room_type, location_note,
    default_capacity, exam_capacity, condition, is_exam_eligible,
    network_ready, power_ready, notes
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: UpdateSchoolRoom :one
UPDATE school_rooms
SET code             = $2,
    name             = $3,
    building         = $4,
    floor            = $5,
    room_type        = $6,
    location_note    = $7,
    default_capacity = $8,
    exam_capacity    = $9,
    condition        = $10,
    is_exam_eligible = $11,
    network_ready    = $12,
    power_ready      = $13,
    notes            = $14,
    updated_at       = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteSchoolRoom :exec
DELETE FROM school_rooms WHERE id = $1;

-- name: HasSchoolRoomCbtRooms :one
SELECT EXISTS (
    SELECT 1
    FROM cbt_exam_rooms
    WHERE school_room_id = $1
)::boolean;
