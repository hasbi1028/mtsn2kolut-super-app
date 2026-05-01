CREATE TABLE IF NOT EXISTS user_ui_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    sidebar_pinned JSONB NOT NULL DEFAULT '[]'::jsonb,
    sidebar_recent JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_ui_preferences_updated_at
    ON user_ui_preferences (updated_at DESC);
