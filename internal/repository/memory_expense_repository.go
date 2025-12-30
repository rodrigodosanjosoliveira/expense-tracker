package repository

import (
	"context"
	"sort"
	"sync"

	"github.com/rodrigo/expense-tracker/internal/domain"
)

// MemoryExpenseRepository implementa ExpenseRepository usando memória
// Conceito: sync.RWMutex protege acesso concorrente (goroutines)
type MemoryExpenseRepository struct {
	mu       sync.RWMutex
	expenses map[string]*domain.Expense
}

// NewMemoryExpenseRepository cria uma nova instância
// Conceito: Constructor function - padrão comum em Go
func NewMemoryExpenseRepository() *MemoryExpenseRepository {
	return &MemoryExpenseRepository{
		expenses: make(map[string]*domain.Expense),
	}
}

// Create adiciona uma nova despesa
func (r *MemoryExpenseRepository) Create(ctx context.Context, expense *domain.Expense) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.expenses[expense.ID]; exists {
		return ErrAlreadyExists
	}

	r.expenses[expense.ID] = expense
	return nil
}

// GetByID retorna uma despesa pelo ID
func (r *MemoryExpenseRepository) GetByID(ctx context.Context, id string) (*domain.Expense, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	expense, exists := r.expenses[id]
	if !exists {
		return nil, ErrNotFound
	}

	return expense, nil
}

// GetAll retorna todas as despesas
func (r *MemoryExpenseRepository) GetAll(ctx context.Context) ([]*domain.Expense, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*domain.Expense, 0, len(r.expenses))
	for _, expense := range r.expenses {
		result = append(result, expense)
	}

	return result, nil
}

// Update atualiza uma despesa existente
func (r *MemoryExpenseRepository) Update(ctx context.Context, expense *domain.Expense) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.expenses[expense.ID]; !exists {
		return ErrNotFound
	}

	r.expenses[expense.ID] = expense
	return nil
}

// Delete remove uma despesa
func (r *MemoryExpenseRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.expenses[id]; !exists {
		return ErrNotFound
	}

	delete(r.expenses, id)
	return nil
}

// GetAllWithFilters retorna despesas com filtros, ordenação e paginação
func (r *MemoryExpenseRepository) GetAllWithFilters(ctx context.Context, filters *domain.ExpenseFilters) ([]*domain.Expense, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Coletar todas as despesas
	result := make([]*domain.Expense, 0, len(r.expenses))
	for _, expense := range r.expenses {
		// Aplicar filtros
		if !r.matchesFilters(expense, filters) {
			continue
		}
		result = append(result, expense)
	}

	// Ordenar
	r.sortExpenses(result, filters)

	// Aplicar paginação
	start := filters.Offset
	end := filters.Offset + filters.Limit

	if start > len(result) {
		return []*domain.Expense{}, nil
	}
	if end > len(result) {
		end = len(result)
	}

	return result[start:end], nil
}

// Count retorna o total de despesas que correspondem aos filtros
func (r *MemoryExpenseRepository) Count(ctx context.Context, filters *domain.ExpenseFilters) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, expense := range r.expenses {
		if r.matchesFilters(expense, filters) {
			count++
		}
	}

	return count, nil
}

// matchesFilters verifica se uma despesa corresponde aos filtros
func (r *MemoryExpenseRepository) matchesFilters(expense *domain.Expense, filters *domain.ExpenseFilters) bool {
	// Filtro por categoria
	if filters.Category != nil && expense.Category != *filters.Category {
		return false
	}

	// Filtro por valor mínimo
	if filters.MinAmount != nil && expense.Amount < *filters.MinAmount {
		return false
	}

	// Filtro por valor máximo
	if filters.MaxAmount != nil && expense.Amount > *filters.MaxAmount {
		return false
	}

	// Filtro por data inicial
	if filters.StartDate != nil && expense.Date.Before(*filters.StartDate) {
		return false
	}

	// Filtro por data final
	if filters.EndDate != nil && expense.Date.After(*filters.EndDate) {
		return false
	}

	return true
}

// sortExpenses ordena as despesas de acordo com os filtros
func (r *MemoryExpenseRepository) sortExpenses(expenses []*domain.Expense, filters *domain.ExpenseFilters) {
	sort.Slice(expenses, func(i, j int) bool {
		var less bool
		switch filters.SortBy {
		case "amount":
			less = expenses[i].Amount < expenses[j].Amount
		case "category":
			less = expenses[i].Category < expenses[j].Category
		case "created_at":
			less = expenses[i].CreatedAt.Before(expenses[j].CreatedAt)
		default: // "date"
			less = expenses[i].Date.Before(expenses[j].Date)
		}

		if filters.SortOrder == "desc" {
			return !less
		}
		return less
	})
}
