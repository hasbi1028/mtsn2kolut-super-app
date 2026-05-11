CREATE TABLE IF NOT EXISTS cbt_event_question_requirements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES cbt_exam_events(id) ON DELETE CASCADE,
    scope_mode TEXT NOT NULL DEFAULT 'per_rombel' CHECK (scope_mode IN ('per_rombel', 'per_level', 'pool_level_subject')),
    level TEXT,
    class_id UUID REFERENCES school_classes(id) ON DELETE CASCADE,
    subject_id UUID REFERENCES subjects(id) ON DELETE CASCADE,
    target_pg INTEGER NOT NULL DEFAULT 20 CHECK (target_pg >= 0),
    target_essay INTEGER NOT NULL DEFAULT 5 CHECK (target_essay >= 0),
    status_filter TEXT NOT NULL DEFAULT 'published_only' CHECK (status_filter IN ('published_only', 'all_progress')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_cbt_event_question_requirements_event_default
    ON cbt_event_question_requirements (event_id)
    WHERE level IS NULL AND class_id IS NULL AND subject_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_cbt_event_question_requirements_event
    ON cbt_event_question_requirements (event_id);
