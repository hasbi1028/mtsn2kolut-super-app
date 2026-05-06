ALTER TABLE users
    ADD COLUMN IF NOT EXISTS display_name text,
    ADD COLUMN IF NOT EXISTS last_login_at timestamptz,
    ADD COLUMN IF NOT EXISTS deleted_at timestamptz;

UPDATE users u
SET display_name = COALESCE(NULLIF(u.display_name, ''), NULLIF(COALESCE(e.nama, s.nama, p.nama, ''), ''), u.username)
FROM users ux
LEFT JOIN employees e ON e.id = ux.employee_id
LEFT JOIN students s ON s.id = ux.student_id
LEFT JOIN parents p ON p.id = ux.parent_id
WHERE u.id = ux.id
  AND (u.display_name IS NULL OR btrim(u.display_name) = '');

CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);
CREATE INDEX IF NOT EXISTS idx_users_last_login_at ON users (last_login_at);
