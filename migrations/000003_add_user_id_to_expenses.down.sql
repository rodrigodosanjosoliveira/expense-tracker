-- Drop index
DROP INDEX IF EXISTS idx_expenses_user_id;

-- Drop foreign key constraint
ALTER TABLE expenses DROP CONSTRAINT IF EXISTS fk_expenses_user_id;

-- Drop user_id column
ALTER TABLE expenses DROP COLUMN IF EXISTS user_id;
