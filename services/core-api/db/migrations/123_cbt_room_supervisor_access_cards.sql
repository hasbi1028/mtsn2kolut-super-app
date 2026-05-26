-- Supervisor CBT access cards are issued per exam room, not per individual proctor.
-- Keep one active supervisor sheet credential for each room so substitute supervisors can use it.

WITH ranked_active_proctor_cards AS (
  SELECT
    id,
    ROW_NUMBER() OVER (PARTITION BY room_id ORDER BY created_at DESC, id DESC) AS rn
  FROM cbt_exam_access_cards
  WHERE card_type = 'proctor'
    AND room_id IS NOT NULL
    AND revoked_at IS NULL
    AND status = 'active'
)
UPDATE cbt_exam_access_cards ac
SET status = 'revoked', revoked_at = NOW(), updated_at = NOW()
FROM ranked_active_proctor_cards ranked
WHERE ac.id = ranked.id
  AND ranked.rn > 1;

ALTER TABLE cbt_exam_access_cards
  DROP CONSTRAINT IF EXISTS chk_cbt_exam_access_cards_target;

ALTER TABLE cbt_exam_access_cards
  ADD CONSTRAINT chk_cbt_exam_access_cards_target CHECK (
    (card_type = 'participant' AND participant_id IS NOT NULL AND room_proctor_id IS NULL)
    OR (card_type = 'proctor' AND room_id IS NOT NULL AND participant_id IS NULL)
  );

DROP INDEX IF EXISTS idx_cbt_exam_access_cards_proctor_active;

CREATE UNIQUE INDEX IF NOT EXISTS idx_cbt_exam_access_cards_proctor_room_active
  ON cbt_exam_access_cards (room_id)
  WHERE card_type = 'proctor' AND revoked_at IS NULL AND status = 'active';
