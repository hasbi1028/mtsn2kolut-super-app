CREATE TABLE inventory_item_events (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id       UUID        NOT NULL REFERENCES inventory_items(id) ON DELETE CASCADE,
    actor_user_id UUID        NULL REFERENCES users(id) ON DELETE SET NULL,
    action        TEXT        NOT NULL,
    summary       TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_inventory_item_events_action CHECK (action IN ('create', 'update', 'delete'))
);

CREATE INDEX idx_inventory_item_events_item_created ON inventory_item_events(item_id, created_at DESC);
