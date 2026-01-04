-- Add user_id column to expenses table
ALTER TABLE expenses ADD COLUMN user_id UUID;

-- Create foreign key constraint to users table
ALTER TABLE expenses ADD CONSTRAINT fk_expenses_user_id
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Create index for better query performance
CREATE INDEX idx_expenses_user_id ON expenses(user_id);

-- Add comment
COMMENT ON COLUMN expenses.user_id IS 'References the user who owns this expense';
