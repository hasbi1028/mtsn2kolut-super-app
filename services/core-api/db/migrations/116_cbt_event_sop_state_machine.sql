ALTER TABLE cbt_exam_events
    ADD COLUMN IF NOT EXISTS sop_state TEXT NOT NULL DEFAULT 'draft',
    ADD COLUMN IF NOT EXISTS sop_state_updated_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS sop_state_updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS sop_state_note TEXT NOT NULL DEFAULT '';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'cbt_exam_events_sop_state_check'
          AND conrelid = 'cbt_exam_events'::regclass
    ) THEN
        ALTER TABLE cbt_exam_events
            ADD CONSTRAINT cbt_exam_events_sop_state_check
            CHECK (sop_state IN (
                'draft',
                'question_authoring',
                'question_verification',
                'package_ready',
                'participants_rooms_ready',
                'tokens_cards_ready',
                'execution',
                'grading',
                'result_verification',
                'final_archive',
                'cancelled'
            ));
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS cbt_event_sop_transitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES cbt_exam_events(id) ON DELETE CASCADE,
    from_state TEXT NOT NULL CHECK (from_state IN (
        'draft',
        'question_authoring',
        'question_verification',
        'package_ready',
        'participants_rooms_ready',
        'tokens_cards_ready',
        'execution',
        'grading',
        'result_verification',
        'final_archive',
        'cancelled'
    )),
    to_state TEXT NOT NULL CHECK (to_state IN (
        'draft',
        'question_authoring',
        'question_verification',
        'package_ready',
        'participants_rooms_ready',
        'tokens_cards_ready',
        'execution',
        'grading',
        'result_verification',
        'final_archive',
        'cancelled'
    )),
    actor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    note TEXT NOT NULL DEFAULT '',
    gate_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cbt_exam_events_sop_state
    ON cbt_exam_events (sop_state);

CREATE INDEX IF NOT EXISTS idx_cbt_event_sop_transitions_event_created
    ON cbt_event_sop_transitions (event_id, created_at DESC);
