-- Sprint 55: Document-cycle audit filters, verification queue, and PIC reminders.

ALTER TABLE document_cycle_events
    DROP CONSTRAINT chk_document_cycle_event_type,
    ADD CONSTRAINT chk_document_cycle_event_type
        CHECK (event_type IN ('created', 'updated', 'status_changed', 'generated', 'monitoring_note', 'reminder_sent'));

CREATE INDEX idx_document_cycle_obligations_verifier
    ON document_cycle_obligations(verifier_employee_id, period_year DESC, status, due_date)
    WHERE verifier_employee_id IS NOT NULL;

CREATE INDEX idx_document_cycle_events_type_actor
    ON document_cycle_events(obligation_id, event_type, actor_user_id, created_at DESC);

CREATE TABLE app_notifications (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category    TEXT        NOT NULL,
    title       TEXT        NOT NULL,
    body        TEXT        NOT NULL,
    entity_type TEXT        NOT NULL DEFAULT '',
    entity_id   TEXT        NOT NULL DEFAULT '',
    link_path   TEXT        NOT NULL DEFAULT '',
    dedupe_key  TEXT        NOT NULL,
    read_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_app_notifications_category
        CHECK (category IN ('document_cycle_reminder', 'document_cycle_late', 'system'))
);

CREATE UNIQUE INDEX ux_app_notifications_user_dedupe
    ON app_notifications(user_id, dedupe_key);

CREATE INDEX idx_app_notifications_user_created
    ON app_notifications(user_id, created_at DESC);

CREATE INDEX idx_app_notifications_user_unread
    ON app_notifications(user_id, created_at DESC)
    WHERE read_at IS NULL;
