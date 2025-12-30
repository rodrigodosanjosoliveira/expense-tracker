-- Migration: Criar tabela expenses
-- Conceito: Migrations permitem versionar o schema do banco de dados

CREATE TABLE IF NOT EXISTS expenses (
    id UUID PRIMARY KEY,
    description VARCHAR(255) NOT NULL,
    amount DECIMAL(10, 2) NOT NULL CHECK (amount > 0),
    category VARCHAR(100) NOT NULL,
    date TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Índices para melhorar performance de queries
CREATE INDEX idx_expenses_category ON expenses(category);
CREATE INDEX idx_expenses_date ON expenses(date);
CREATE INDEX idx_expenses_created_at ON expenses(created_at);

-- Comentários na tabela
COMMENT ON TABLE expenses IS 'Tabela de despesas pessoais';
COMMENT ON COLUMN expenses.id IS 'UUID único da despesa';
COMMENT ON COLUMN expenses.description IS 'Descrição da despesa';
COMMENT ON COLUMN expenses.amount IS 'Valor da despesa em reais';
COMMENT ON COLUMN expenses.category IS 'Categoria da despesa (Alimentação, Transporte, etc)';
COMMENT ON COLUMN expenses.date IS 'Data em que a despesa ocorreu';
COMMENT ON COLUMN expenses.created_at IS 'Data de criação do registro';
COMMENT ON COLUMN expenses.updated_at IS 'Data da última atualização';
