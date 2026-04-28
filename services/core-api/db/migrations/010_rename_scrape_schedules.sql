-- Unify morning/afternoon schedule labels to single "Rekap" concept.
-- The run_type values (morning/afternoon) are kept for backward compatibility;
-- only the display labels are updated.
UPDATE schedules SET label = 'Rekap' WHERE run_type IN ('morning', 'afternoon');
