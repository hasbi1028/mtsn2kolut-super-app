-- Migration: 060_cbt_exam_token_hardening
--
-- Existing CBT tokens used to be shorter. Do not rotate printed tokens from a
-- normal deploy migration; token re-issue must be an explicit operator action
-- on a draft/scheduled session before cards are printed.
--
-- The migration only enforces that currently stored non-empty tokens are unique.
-- If duplicate legacy tokens exist, stop loudly so the operator can reissue the
-- affected draft/scheduled session tokens through the CBT token action first.

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM cbt_exam_participants
    WHERE token <> ''
    GROUP BY token
    HAVING COUNT(*) > 1
  ) THEN
    RAISE EXCEPTION 'duplicate CBT participant tokens exist; reissue affected draft/scheduled session tokens before applying token uniqueness';
  END IF;
END $$;

DROP INDEX IF EXISTS idx_cbt_participants_token;

CREATE UNIQUE INDEX IF NOT EXISTS idx_cbt_participants_token
  ON cbt_exam_participants (token)
  WHERE token <> '';
