package repository

import (
	"context"

	"github.com/rodrigo/expense-tracker/internal/domain"
)

// CategoryRepository define as operacoes de persistencia para categorias
type CategoryRepository interface {
	// Create persists a new category. Returns domain.ErrCategoryAlreadyExists
	// (mapped from pgx unique-violation code 23505) if a category with the same
	// name already exists for the given user_id (case-insensitive).
	Create(ctx context.Context, category *domain.Category) error

	// GetByID returns a category by its ID scoped to userID.
	// Returns ErrNotFound when no matching record exists.
	GetByID(ctx context.Context, id string, userID string) (*domain.Category, error)

	// GetByName returns a category by its name (case-insensitive) scoped to userID.
	// Used by ExpenseService to resolve a category string to a category_id.
	// Returns ErrNotFound when no matching record exists.
	GetByName(ctx context.Context, name string, userID string) (*domain.Category, error)

	// GetAll returns all categories owned by userID.
	GetAll(ctx context.Context, userID string) ([]*domain.Category, error)

	// Update persists changes to an existing category.
	// Returns ErrNotFound when the category does not exist for the given userID.
	// Returns domain.ErrCategoryAlreadyExists on name collision.
	Update(ctx context.Context, category *domain.Category) error

	// Delete removes a category by ID scoped to userID.
	// Returns ErrNotFound when the category does not exist for the given userID.
	// Returns domain.ErrCategoryInUse when at least one expense still references
	// the category (mapped from pgx FK-violation code 23503).
	Delete(ctx context.Context, id string, userID string) error
}
