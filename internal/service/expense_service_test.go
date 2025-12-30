package service

import (
	"context"
	"testing"
	"time"

	"github.com/rodrigo/expense-tracker/internal/domain"
	"github.com/rodrigo/expense-tracker/internal/repository"
)

// MockIDGenerator é um mock para testes
// Conceito: Mocking manual - simples e eficaz em Go
type MockIDGenerator struct {
	id string
}

func (m *MockIDGenerator) Generate() string {
	return m.id
}

func TestExpenseService_CreateExpense(t *testing.T) {
	repo := repository.NewMemoryExpenseRepository()
	idGen := &MockIDGenerator{id: "test-123"}
	service := NewExpenseService(repo, idGen)
	ctx := context.Background()

	tests := []struct {
		name    string
		expense *domain.Expense
		wantErr bool
	}{
		{
			name: "valid expense",
			expense: &domain.Expense{
				Description: "Jantar",
				Amount:      50.00,
				Category:    "Alimentação",
				Date:        time.Now(),
			},
			wantErr: false,
		},
		{
			name: "invalid expense - empty description",
			expense: &domain.Expense{
				Description: "",
				Amount:      50.00,
				Category:    "Alimentação",
			},
			wantErr: true,
		},
		{
			name: "invalid expense - zero amount",
			expense: &domain.Expense{
				Description: "Jantar",
				Amount:      0,
				Category:    "Alimentação",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CreateExpense(ctx, tt.expense)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateExpense() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verificar que ID e timestamps foram setados
			if !tt.wantErr {
				if tt.expense.ID != "test-123" {
					t.Errorf("CreateExpense() ID = %v, want test-123", tt.expense.ID)
				}
				if tt.expense.CreatedAt.IsZero() {
					t.Error("CreateExpense() CreatedAt not set")
				}
				if tt.expense.UpdatedAt.IsZero() {
					t.Error("CreateExpense() UpdatedAt not set")
				}
			}
		})
	}
}

func TestExpenseService_GetExpense(t *testing.T) {
	repo := repository.NewMemoryExpenseRepository()
	idGen := &MockIDGenerator{id: "1"}
	service := NewExpenseService(repo, idGen)
	ctx := context.Background()

	// Criar uma despesa
	expense := &domain.Expense{
		Description: "Teste",
		Amount:      10.00,
		Category:    "Cat1",
		Date:        time.Now(),
	}
	service.CreateExpense(ctx, expense)

	// Buscar
	found, err := service.GetExpense(ctx, "1")
	if err != nil {
		t.Errorf("GetExpense() error = %v, want nil", err)
	}
	if found.Description != "Teste" {
		t.Errorf("GetExpense() description = %v, want Teste", found.Description)
	}
}

func TestExpenseService_UpdateExpense(t *testing.T) {
	repo := repository.NewMemoryExpenseRepository()
	idGen := &MockIDGenerator{id: "1"}
	service := NewExpenseService(repo, idGen)
	ctx := context.Background()

	// Criar
	expense := &domain.Expense{
		Description: "Original",
		Amount:      10.00,
		Category:    "Cat1",
		Date:        time.Now(),
	}
	service.CreateExpense(ctx, expense)
	createdAt := expense.CreatedAt

	// Aguardar um pouco para garantir timestamps diferentes
	time.Sleep(10 * time.Millisecond)

	// Atualizar
	expense.Description = "Atualizado"
	expense.Amount = 20.00
	err := service.UpdateExpense(ctx, expense)
	if err != nil {
		t.Errorf("UpdateExpense() error = %v, want nil", err)
	}

	// Verificar que CreatedAt foi preservado e UpdatedAt foi alterado
	if expense.CreatedAt != createdAt {
		t.Error("UpdateExpense() CreatedAt should be preserved")
	}
	if expense.UpdatedAt.Equal(createdAt) {
		t.Error("UpdateExpense() UpdatedAt should be updated")
	}

	// Buscar e verificar
	updated, _ := service.GetExpense(ctx, "1")
	if updated.Description != "Atualizado" {
		t.Errorf("UpdateExpense() description = %v, want Atualizado", updated.Description)
	}
}

func TestExpenseService_DeleteExpense(t *testing.T) {
	repo := repository.NewMemoryExpenseRepository()
	idGen := &MockIDGenerator{id: "1"}
	service := NewExpenseService(repo, idGen)
	ctx := context.Background()

	// Criar
	expense := &domain.Expense{
		Description: "Teste",
		Amount:      10.00,
		Category:    "Cat1",
		Date:        time.Now(),
	}
	service.CreateExpense(ctx, expense)

	// Deletar
	err := service.DeleteExpense(ctx, "1")
	if err != nil {
		t.Errorf("DeleteExpense() error = %v, want nil", err)
	}

	// Verificar que foi deletado
	_, err = service.GetExpense(ctx, "1")
	if err != repository.ErrNotFound {
		t.Errorf("GetExpense() after delete error = %v, want %v", err, repository.ErrNotFound)
	}
}

func TestExpenseService_ListExpenses(t *testing.T) {
	repo := repository.NewMemoryExpenseRepository()
	idGen := &MockIDGenerator{id: "1"}
	service := NewExpenseService(repo, idGen)
	ctx := context.Background()

	// Lista vazia
	expenses, err := service.ListExpenses(ctx)
	if err != nil {
		t.Errorf("ListExpenses() error = %v, want nil", err)
	}
	if len(expenses) != 0 {
		t.Errorf("ListExpenses() got %d expenses, want 0", len(expenses))
	}

	// Adicionar despesas
	service.CreateExpense(ctx, &domain.Expense{
		Description: "A",
		Amount:      10,
		Category:    "Cat1",
		Date:        time.Now(),
	})

	idGen.id = "2"
	service.CreateExpense(ctx, &domain.Expense{
		Description: "B",
		Amount:      20,
		Category:    "Cat2",
		Date:        time.Now(),
	})

	// Listar
	expenses, err = service.ListExpenses(ctx)
	if err != nil {
		t.Errorf("ListExpenses() error = %v, want nil", err)
	}
	if len(expenses) != 2 {
		t.Errorf("ListExpenses() got %d expenses, want 2", len(expenses))
	}
}
