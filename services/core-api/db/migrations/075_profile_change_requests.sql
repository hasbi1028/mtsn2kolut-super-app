-- Sprint 97.6: official profile data change requests.
-- Self-service account edits may update contact/avatar directly, but official
-- identity fields are only changed after an admin/reviewer approval.

CREATE TYPE profile_change_request_status AS ENUM (
    'pending',
    'approved',
    'rejected',
    'cancelled'
);

CREATE TABLE profile_change_requests (
    id                 UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    requester_user_id  UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    profile_type       TEXT        NOT NULL,
    target_employee_id UUID        REFERENCES employees(id) ON DELETE SET NULL,
    target_student_id  UUID        REFERENCES students(id) ON DELETE SET NULL,
    target_parent_id   UUID        REFERENCES parents(id) ON DELETE SET NULL,
    field_key          TEXT        NOT NULL,
    current_value      TEXT        NOT NULL DEFAULT '',
    requested_value    TEXT        NOT NULL DEFAULT '',
    reason             TEXT        NOT NULL DEFAULT '',
    status             profile_change_request_status NOT NULL DEFAULT 'pending',
    reviewer_user_id   UUID        REFERENCES users(id) ON DELETE SET NULL,
    review_note        TEXT        NOT NULL DEFAULT '',
    reviewed_at        TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_profile_change_requests_profile_type
        CHECK (profile_type IN ('employee', 'student', 'parent')),
    CONSTRAINT chk_profile_change_requests_field_key
        CHECK (btrim(field_key) <> ''),
    CONSTRAINT chk_profile_change_requests_status_review
        CHECK (
            (status = 'pending' AND reviewer_user_id IS NULL AND reviewed_at IS NULL)
            OR status <> 'pending'
        ),
    CONSTRAINT chk_profile_change_requests_target
        CHECK (
            (profile_type = 'employee' AND target_employee_id IS NOT NULL AND target_student_id IS NULL AND target_parent_id IS NULL)
            OR (profile_type = 'student' AND target_employee_id IS NULL AND target_student_id IS NOT NULL AND target_parent_id IS NULL)
            OR (profile_type = 'parent' AND target_employee_id IS NULL AND target_student_id IS NULL AND target_parent_id IS NOT NULL)
        )
);

CREATE INDEX idx_profile_change_requests_requester_created
    ON profile_change_requests(requester_user_id, created_at DESC);

CREATE INDEX idx_profile_change_requests_status_created
    ON profile_change_requests(status, created_at DESC);

CREATE INDEX idx_profile_change_requests_reviewer_reviewed
    ON profile_change_requests(reviewer_user_id, reviewed_at DESC)
    WHERE reviewer_user_id IS NOT NULL;
