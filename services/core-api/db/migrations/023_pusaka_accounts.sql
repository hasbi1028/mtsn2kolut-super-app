CREATE TABLE IF NOT EXISTS pusaka_accounts (
  id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id      UUID        NOT NULL UNIQUE REFERENCES employees(id) ON DELETE CASCADE,
  pusaka_username  TEXT        NOT NULL DEFAULT '',
  pusaka_password  TEXT        NOT NULL DEFAULT '',
  is_enabled       BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pusaka_accounts_enabled
  ON pusaka_accounts (is_enabled)
  WHERE is_enabled = TRUE;

INSERT INTO pusaka_accounts (employee_id, pusaka_username, pusaka_password, is_enabled)
SELECT e.id, e.pusaka_username, e.pusaka_password, TRUE
FROM employees e
WHERE e.pusaka_username <> '' OR e.pusaka_password <> ''
ON CONFLICT (employee_id) DO UPDATE
SET pusaka_username = EXCLUDED.pusaka_username,
    pusaka_password = EXCLUDED.pusaka_password,
    is_enabled      = TRUE,
    updated_at      = NOW();
