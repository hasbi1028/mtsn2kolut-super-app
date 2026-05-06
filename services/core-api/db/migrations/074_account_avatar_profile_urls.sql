-- Sprint 97.5: user-owned profile/avatar photo URLs.
-- Students already own photo_url from the kesiswaan foundation; employees and
-- parents need the same nullable-by-empty-string URL slot for account avatars.

ALTER TABLE employees
    ADD COLUMN IF NOT EXISTS photo_url TEXT NOT NULL DEFAULT '';

ALTER TABLE parents
    ADD COLUMN IF NOT EXISTS photo_url TEXT NOT NULL DEFAULT '';
