-- Sprint 97: user-owned employee contact profile fields.
-- Official employee identity remains in existing master fields; these columns are
-- limited to self-service contact data from /api/auth/account/contact.

ALTER TABLE employees
    ADD COLUMN IF NOT EXISTS phone TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS email TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS address TEXT NOT NULL DEFAULT '';
