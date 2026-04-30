-- Migration 020: Rename user_roles to avoid sqlc name conflict with UserRole enum

ALTER TABLE user_roles RENAME TO user_account_roles;
