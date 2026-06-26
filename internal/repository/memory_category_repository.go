package repository

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/rodrigo/expense-tracker/internal/domain"
)

// MemoryCategoryRepository implements CategoryRepository using in-memory storage.
// It is intended exclusively for tests — no external database required.
type MemoryCategoryRepository struct {
	mu         sync.RWMutex
	categories map[string]*domain.Category
	// inUse contains category IDs that should return domain.ErrCategoryInUse on Delete.
	// Use MarkCategoryInUse to populate this set in tests.
	inUse map[string]bool
}

// NewMemoryCategoryRepository creates a new empty in-memory category repository.
func NewMemoryCategoryRepository() *MemoryCategoryRepository {
	return &MemoryCategoryRepository{
		categories: make(map[string]*domain.Category),
		inUse:      make(map[string]bool),
	}
}

// Create persists a new category. Returns domain.ErrCategoryAlreadyExists when a
// category with the same name (case-insensitive) already exists for the same user_id.
// Returns ErrAlreadyExists when a category with the same ID already exists.
func (r *MemoryCategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.categories[category.ID]; exists {
		return ErrAlreadyExists
	}

	for _, c := range r.categories {
		if c.UserID == category.UserID && strings.EqualFold(c.Name, category.Name) {
			return domain.ErrCategoryAlreadyExists
		}
	}

	now := time.Now().UTC()
	saved := *category
	saved.CreatedAt = now
	saved.UpdatedAt = now
	r.categories[saved.ID] = &saved
	return nil
}

// GetByID returns a category by its ID scoped to userID.
// Returns ErrNotFound when no matching record exists for that user.
func (r *MemoryCategoryRepository) GetByID(ctx context.Context, id string, userID string) (*domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.categories[id]
	if !ok || c.UserID != userID {
		return nil, ErrNotFound
	}

	result := *c
	return &result, nil
}

// GetByName returns a category by name (case-insensitive) scoped to userID.
// Returns ErrNotFound when no matching record exists.
func (r *MemoryCategoryRepository) GetByName(ctx context.Context, name string, userID string) (*domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, c := range r.categories {
		if c.UserID == userID && strings.EqualFold(c.Name, name) {
			result := *c
			return &result, nil
		}
	}
	return nil, ErrNotFound
}

// GetAll returns all categories owned by userID.
func (r *MemoryCategoryRepository) GetAll(ctx context.Context, userID string) ([]*domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*domain.Category, 0)
	for _, c := range r.categories {
		if c.UserID == userID {
			copy := *c
			result = append(result, &copy)
		}
	}
	return result, nil
}

// Update persists changes to an existing category.
// Returns ErrNotFound when the category does not exist for the given userID.
// Returns domain.ErrCategoryAlreadyExists on name collision with another category.
func (r *MemoryCategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.categories[category.ID]
	if !ok || existing.UserID != category.UserID {
		return ErrNotFound
	}

	// Check name collision with other categories of the same user
	for _, c := range r.categories {
		if c.ID != category.ID && c.UserID == category.UserID && strings.EqualFold(c.Name, category.Name) {
			return domain.ErrCategoryAlreadyExists
		}
	}

	saved := *category
	saved.CreatedAt = existing.CreatedAt
	saved.UpdatedAt = time.Now().UTC()
	r.categories[saved.ID] = &saved
	return nil
}

// Delete removes a category by ID scoped to userID.
// Returns ErrNotFound when the category does not exist for the given userID.
// Returns domain.ErrCategoryInUse when the category is marked as in-use (test helper).
func (r *MemoryCategoryRepository) Delete(ctx context.Context, id string, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, ok := r.categories[id]
	if !ok || c.UserID != userID {
		return ErrNotFound
	}

	if r.inUse[id] {
		return domain.ErrCategoryInUse
	}

	delete(r.categories, id)
	return nil
}

// MarkCategoryInUse marks the given category ID so that the next Delete call
// returns domain.ErrCategoryInUse. This is a test-only helper that simulates
// the FK-violation path without needing a real expense record.
func (r *MemoryCategoryRepository) MarkCategoryInUse(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.inUse[id] = true
}
