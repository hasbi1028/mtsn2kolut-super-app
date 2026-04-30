-- Migration 021: Remove redundant role column from users
-- After successfully migrating to user_account_roles many-to-many table.

ALTER TABLE users DROP COLUMN IF EXISTS role;
