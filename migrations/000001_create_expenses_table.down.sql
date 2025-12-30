-- Migration: Reverter criação da tabela expenses

DROP INDEX IF EXISTS idx_expenses_created_at;
DROP INDEX IF EXISTS idx_expenses_date;
DROP INDEX IF EXISTS idx_expenses_category;
DROP TABLE IF EXISTS expenses;
