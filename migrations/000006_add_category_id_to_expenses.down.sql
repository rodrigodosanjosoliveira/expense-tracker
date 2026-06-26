-- Migration 000006 rollback: Remove category_id from expenses table

DROP INDEX IF EXISTS idx_expenses_category_id;

ALTER TABLE expenses DROP CONSTRAINT IF EXISTS fk_expenses_category_id;

ALTER TABLE expenses DROP COLUMN IF EXISTS category_id;
