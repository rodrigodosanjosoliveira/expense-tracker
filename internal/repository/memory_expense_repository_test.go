package repository

import (
	"context"
	"testing"
	"time"

	"github.com/rodrigo/expense-tracker/internal/domain"
)

func TestMemoryExpenseRepositoryCreate(t *testing.T) {
	repo := NewMemoryExpenseRepository()
	ctx := context.Background()

	expense := &domain.Expense{
		ID:          "1",
		Description: "Café",
		Amount:      5.50,
		Category:    "Alimentação",
		Date:        time.Now(),
	}

	err := repo.Create(ctx, expense)
	if err != nil {
		t.Errorf("Create() error = %v, want nil", err)
	}

	// Tentar criar novamente deve retornar erro
	err = repo.Create(ctx, expense)
	if err != ErrAlreadyExists {
		t.Errorf("Create() error = %v, want %v", err, ErrAlreadyExists)
	}
}

func TestMemoryExpenseRepositoryGetByID(t *testing.T) {
	repo := NewMemoryExpenseRepository()
	ctx := context.Background()

	expense := &domain.Expense{
		ID:          "1",
		Description: "Transporte",
		Amount:      10.00,
		Category:    "Transporte",
		Date:        time.Now(),
	}

	// Buscar ID inexistente
	_, err := repo.GetByID(ctx, "999")
	if err != ErrNotFound {
		t.Errorf("GetByID() error = %v, want %v", err, ErrNotFound)
	}

	// Criar e buscar
	repo.Create(ctx, expense)
	found, err := repo.GetByID(ctx, "1")
	if err != nil {
		t.Errorf("GetByID() error = %v, want nil", err)
	}
	if found.ID != expense.ID {
		t.Errorf("GetByID() got ID = %v, want %v", found.ID, expense.ID)
	}
}

func TestMemoryExpenseRepositoryGetAll(t *testing.T) {
	repo := NewMemoryExpenseRepository()
	ctx := context.Background()

	// Lista vazia
	expenses, err := repo.GetAll(ctx)
	if err != nil {
		t.Errorf("GetAll() error = %v, want nil", err)
	}
	if len(expenses) != 0 {
		t.Errorf("GetAll() got %d expenses, want 0", len(expenses))
	}

	// Adicionar algumas despesas
	repo.Create(ctx, &domain.Expense{ID: "1", Description: "A", Amount: 10, Category: "Cat1"})
	repo.Create(ctx, &domain.Expense{ID: "2", Description: "B", Amount: 20, Category: "Cat2"})

	expenses, err = repo.GetAll(ctx)
	if err != nil {
		t.Errorf("GetAll() error = %v, want nil", err)
	}
	if len(expenses) != 2 {
		t.Errorf("GetAll() got %d expenses, want 2", len(expenses))
	}
}

func TestMemoryExpenseRepositoryUpdate(t *testing.T) {
	repo := NewMemoryExpenseRepository()
	ctx := context.Background()

	expense := &domain.Expense{
		ID:          "1",
		Description: "Original",
		Amount:      10.00,
		Category:    "Cat1",
		Date:        time.Now(),
	}

	// Update de despesa inexistente
	err := repo.Update(ctx, expense)
	if err != ErrNotFound {
		t.Errorf("Update() error = %v, want %v", err, ErrNotFound)
	}

	// Criar e atualizar
	repo.Create(ctx, expense)
	expense.Description = "Atualizado"
	err = repo.Update(ctx, expense)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	// Verificar atualização
	updated, _ := repo.GetByID(ctx, "1")
	if updated.Description != "Atualizado" {
		t.Errorf("Update() description = %v, want Atualizado", updated.Description)
	}
}

func TestMemoryExpenseRepositoryDelete(t *testing.T) {
	repo := NewMemoryExpenseRepository()
	ctx := context.Background()

	// Delete de despesa inexistente
	err := repo.Delete(ctx, "999")
	if err != ErrNotFound {
		t.Errorf("Delete() error = %v, want %v", err, ErrNotFound)
	}

	// Criar e deletar
	expense := &domain.Expense{ID: "1", Description: "A", Amount: 10, Category: "Cat1"}
	repo.Create(ctx, expense)
	err = repo.Delete(ctx, "1")
	if err != nil {
		t.Errorf("Delete() error = %v, want nil", err)
	}

	// Verificar que foi deletado
	_, err = repo.GetByID(ctx, "1")
	if err != ErrNotFound {
		t.Errorf("GetByID() after delete error = %v, want %v", err, ErrNotFound)
	}
}
