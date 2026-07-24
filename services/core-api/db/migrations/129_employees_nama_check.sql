-- Migration: Add CHECK constraint to prevent empty employee names.
-- Cleans up existing empty nama rows with a placeholder before applying the constraint.

UPDATE employees SET nama = '(tanpa nama)' WHERE nama = '';

ALTER TABLE employees ADD CONSTRAINT employees_nama_not_empty CHECK (nama <> '');
