package repository

import (
	"context"
	"errors"

	"github.com/rodrigo/expense-tracker/internal/domain"
)

// Erros comuns do repository
var (
	ErrNotFound      = errors.New("expense not found")
	ErrAlreadyExists = errors.New("expense already exists")
)

// ExpenseRepository define as operações de persistência
// Conceito: Interfaces em Go são satisfeitas implicitamente
type ExpenseRepository interface {
	Create(ctx context.Context, expense *domain.Expense) error
	GetByID(ctx context.Context, id string) (*domain.Expense, error)
	GetAll(ctx context.Context) ([]*domain.Expense, error)
	GetAllWithFilters(ctx context.Context, filters *domain.ExpenseFilters) ([]*domain.Expense, error)
	Count(ctx context.Context, filters *domain.ExpenseFilters) (int, error)
	Update(ctx context.Context, expense *domain.Expense) error
	Delete(ctx context.Context, id string) error
}
