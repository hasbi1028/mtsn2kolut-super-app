-- Migration: 060_cbt_exam_token_hardening
--
-- Existing CBT tokens were 8 hex characters. Re-issue all existing non-empty
-- participant tokens as 128-bit random hex before enforcing global uniqueness.
-- Apply this outside an active exam window because printed/distributed cards
-- must be regenerated after migration.

UPDATE cbt_exam_participants
SET token = encode(gen_random_bytes(16), 'hex')
WHERE token <> '';

DROP INDEX IF EXISTS idx_cbt_participants_token;

CREATE UNIQUE INDEX idx_cbt_participants_token
  ON cbt_exam_participants (token)
  WHERE token <> '';
