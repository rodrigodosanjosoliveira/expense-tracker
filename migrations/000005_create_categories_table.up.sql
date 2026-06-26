-- Migration 000005: Create categories table
-- Categories are scoped per user (user_id required in all queries)

CREATE TABLE IF NOT EXISTS categories (
    id          UUID         PRIMARY KEY,
    user_id     UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);

-- Functional unique index: enforces case-insensitive uniqueness per user
-- Uses LOWER(name) instead of citext to avoid requiring the citext extension
CREATE UNIQUE INDEX idx_categories_user_id_name_lower
    ON categories (user_id, LOWER(name));

-- Index for all queries scoped by user_id (GetAll, GetByName, Delete)
CREATE INDEX idx_categories_user_id ON categories (user_id);

-- Table and column comments
COMMENT ON TABLE categories IS 'User-scoped expense categories';
COMMENT ON COLUMN categories.id IS 'Unique identifier for the category (UUID)';
COMMENT ON COLUMN categories.user_id IS 'References the user who owns this category';
COMMENT ON COLUMN categories.name IS 'Category display name (unique per user, case-insensitive)';
COMMENT ON COLUMN categories.created_at IS 'Timestamp when the category was created';
COMMENT ON COLUMN categories.updated_at IS 'Timestamp when the category was last updated';
