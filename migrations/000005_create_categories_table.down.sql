-- Migration 000005 rollback: Drop categories table
-- NOTE: migration 000006 (which adds category_id FK to expenses) must be rolled back first

DROP INDEX IF EXISTS idx_categories_user_id;
DROP INDEX IF EXISTS idx_categories_user_id_name_lower;
DROP TABLE IF EXISTS categories;
