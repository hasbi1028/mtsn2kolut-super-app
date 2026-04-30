-- Migration 022: Add is_active to users for account suspension

ALTER TABLE users ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT TRUE;
