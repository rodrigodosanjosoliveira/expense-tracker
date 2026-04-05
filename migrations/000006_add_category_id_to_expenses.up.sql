-- Migration 000006: Add category_id to expenses table
-- Strategy: nullable column + NOT VALID FK for zero-downtime (no table-level lock on existing rows)
-- Enforcement of non-null for NEW expenses is handled by the service layer, not the DB constraint

ALTER TABLE expenses
    ADD COLUMN IF NOT EXISTS category_id UUID;

-- Add FK with NOT VALID to skip validation of existing rows (all NULL, which is valid for nullable FK)
-- A separate VALIDATE CONSTRAINT step can be run online after backfill if needed
ALTER TABLE expenses
    ADD CONSTRAINT fk_expenses_category_id
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE RESTRICT
    NOT VALID;

-- Index to support JOIN queries (expense -> category) and filter by category_id
CREATE INDEX idx_expenses_category_id ON expenses (category_id);

COMMENT ON COLUMN expenses.category_id IS 'References the normalized category; NULL for legacy expenses created before migration 000006';
