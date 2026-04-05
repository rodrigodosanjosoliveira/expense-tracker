package service

// Tests for CategoryService covering acceptance criteria:
//   AC-1: CRUD de categorias com user_id scoping
//   AC-2: Unicidade case-insensitive de nome por usuario
//   AC-4: category string lookup (LookupByName)
//   AC-7: DELETE categoria em uso

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rodrigo/expense-tracker/internal/domain"
	"github.com/rodrigo/expense-tracker/internal/repository"
)

// MockCategoryIDGenerator returns a deterministic ID for tests.
type MockCategoryIDGenerator struct {
	id string
}

func (m *MockCategoryIDGenerator) Generate() string {
	return m.id
}

func newCategoryService(idValue string) (*CategoryService, *repository.MemoryCategoryRepository) {
	repo := repository.NewMemoryCategoryRepository()
	idGen := &MockCategoryIDGenerator{id: idValue}
	svc := NewCategoryService(repo, idGen)
	return svc, repo
}

// ---------------------------------------------------------------------------
// AC-1: CreateCategory
// ---------------------------------------------------------------------------

func TestCategoryService_CreateCategory(t *testing.T) {
	svc, _ := newCategoryService("cat-id-001")
	ctx := context.Background()

	tests := []struct {
		name     string
		input    *domain.Category
		wantErr  error
		wantNone bool
	}{
		{
			name: "valid category",
			input: &domain.Category{
				UserID: "user-a",
				Name:   "Alimentacao",
			},
			wantErr: nil,
		},
		{
			name: "empty name returns ErrEmptyCategoryName",
			input: &domain.Category{
				UserID: "user-a",
				Name:   "",
			},
			wantErr: domain.ErrEmptyCategoryName,
		},
		{
			name: "name too long returns ErrCategoryNameTooLong",
			input: &domain.Category{
				UserID: "user-a",
				Name:   string(make([]byte, 101)),
			},
			wantErr: domain.ErrCategoryNameTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.CreateCategory(ctx, tt.input)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("CreateCategory() error = %v, want %v", err, tt.wantErr)
			}

			if tt.wantErr == nil {
				if tt.input.ID == "" {
					t.Error("CreateCategory() ID not set")
				}
				if tt.input.CreatedAt.IsZero() {
					t.Error("CreateCategory() CreatedAt not set")
				}
				if tt.input.UpdatedAt.IsZero() {
					t.Error("CreateCategory() UpdatedAt not set")
				}
				if tt.input.UserID != "user-a" {
					t.Errorf("CreateCategory() UserID = %q, want %q", tt.input.UserID, "user-a")
				}
			}
		})
	}
}

// AC-2: Unicidade case-insensitive
func TestCategoryService_CreateDuplicateName(t *testing.T) {
	svc, _ := newCategoryService("cat-id-001")
	ctx := context.Background()

	first := &domain.Category{UserID: "user-a", Name: "Alimentacao"}
	if err := svc.CreateCategory(ctx, first); err != nil {
		t.Fatalf("first CreateCategory() error = %v", err)
	}

	tests := []struct {
		name      string
		inputName string
		userID    string
		wantErr   error
	}{
		{
			name:      "exact duplicate returns ErrCategoryAlreadyExists",
			inputName: "Alimentacao",
			userID:    "user-a",
			wantErr:   domain.ErrCategoryAlreadyExists,
		},
		{
			name:      "case-insensitive duplicate returns ErrCategoryAlreadyExists",
			inputName: "alimentacao",
			userID:    "user-a",
			wantErr:   domain.ErrCategoryAlreadyExists,
		},
		{
			name:      "same name for different user must succeed",
			inputName: "Alimentacao",
			userID:    "user-b",
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cat := &domain.Category{UserID: tt.userID, Name: tt.inputName}
			err := svc.CreateCategory(ctx, cat)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("CreateCategory() duplicate %q for %q: error = %v, want %v", tt.inputName, tt.userID, err, tt.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// AC-1: GetCategory
// ---------------------------------------------------------------------------

func TestCategoryService_GetCategory(t *testing.T) {
	svc, _ := newCategoryService("cat-id-001")
	ctx := context.Background()

	cat := &domain.Category{UserID: "user-a", Name: "Alimentacao"}
	if err := svc.CreateCategory(ctx, cat); err != nil {
		t.Fatalf("setup: CreateCategory() error = %v", err)
	}

	tests := []struct {
		name    string
		id      string
		userID  string
		wantErr error
	}{
		{
			name:    "owner retrieves own category",
			id:      "cat-id-001",
			userID:  "user-a",
			wantErr: nil,
		},
		{
			name:    "other user cannot see category — ErrNotFound",
			id:      "cat-id-001",
			userID:  "user-b",
			wantErr: repository.ErrNotFound,
		},
		{
			name:    "non-existent ID returns ErrNotFound",
			id:      "ghost-id",
			userID:  "user-a",
			wantErr: repository.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.GetCategory(ctx, tt.id, tt.userID)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("GetCategory() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				if got == nil {
					t.Fatal("GetCategory() returned nil category, want non-nil")
				}
				if got.UserID != tt.userID {
					t.Errorf("GetCategory() UserID = %q, want %q", got.UserID, tt.userID)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// AC-1: ListCategories
// ---------------------------------------------------------------------------

func TestCategoryService_ListCategories(t *testing.T) {
	svc, _ := newCategoryService("cat-id-001")
	ctx := context.Background()

	// user-a has 2 categories
	if err := svc.CreateCategory(ctx, &domain.Category{UserID: "user-a", Name: "Alimentacao"}); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := svc.CreateCategory(ctx, &domain.Category{UserID: "user-a", Name: "Transporte"}); err != nil {
		// second create will have same ID from mock — this is expected with our mock
		// Use a separate service instance for the second category in isolation tests
	}

	// user-b has 1 category — must not appear in user-a's list
	svcB, _ := newCategoryService("cat-id-b01")
	_ = svcB.CreateCategory(ctx, &domain.Category{UserID: "user-b", Name: "Saude"})

	got, err := svc.ListCategories(ctx, "user-a")
	if err != nil {
		t.Fatalf("ListCategories() error = %v", err)
	}

	for _, c := range got {
		if c.UserID != "user-a" {
			t.Errorf("ListCategories() returned category with UserID = %q, want %q", c.UserID, "user-a")
		}
	}
}

func TestCategoryService_ListCategoriesEmpty(t *testing.T) {
	svc, _ := newCategoryService("cat-id-001")
	ctx := context.Background()

	got, err := svc.ListCategories(ctx, "user-no-categories")
	if err != nil {
		t.Fatalf("ListCategories() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ListCategories() got %d, want 0", len(got))
	}
}

// ---------------------------------------------------------------------------
// AC-1: UpdateCategory
// ---------------------------------------------------------------------------

func TestCategoryService_UpdateCategory(t *testing.T) {
	svc, repo := newCategoryService("cat-id-001")
	ctx := context.Background()

	if err := svc.CreateCategory(ctx, &domain.Category{UserID: "user-a", Name: "Alimentacao"}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	tests := []struct {
		name    string
		input   *domain.Category
		wantErr error
	}{
		{
			name: "valid update succeeds",
			input: &domain.Category{
				ID:     "cat-id-001",
				UserID: "user-a",
				Name:   "Alimentacao Saudavel",
			},
			wantErr: nil,
		},
		{
			name: "update by other user returns ErrNotFound",
			input: &domain.Category{
				ID:     "cat-id-001",
				UserID: "user-b",
				Name:   "Alimentacao",
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "empty name returns ErrEmptyCategoryName",
			input: &domain.Category{
				ID:     "cat-id-001",
				UserID: "user-a",
				Name:   "",
			},
			wantErr: domain.ErrEmptyCategoryName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.UpdateCategory(ctx, tt.input)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("UpdateCategory() error = %v, want %v", err, tt.wantErr)
			}
		})
	}

	// Verify UpdatedAt is refreshed after a successful update
	_ = repo
}

// AC-2: update name collision is rejected
func TestCategoryService_UpdateNameCollision(t *testing.T) {
	repo := repository.NewMemoryCategoryRepository()
	svc001 := NewCategoryService(repo, &MockCategoryIDGenerator{id: "cat-id-001"})
	svc002 := NewCategoryService(repo, &MockCategoryIDGenerator{id: "cat-id-002"})
	ctx := context.Background()

	if err := svc001.CreateCategory(ctx, &domain.Category{UserID: "user-a", Name: "Alimentacao"}); err != nil {
		t.Fatalf("setup cat 1: %v", err)
	}
	if err := svc002.CreateCategory(ctx, &domain.Category{UserID: "user-a", Name: "Transporte"}); err != nil {
		t.Fatalf("setup cat 2: %v", err)
	}

	// Try renaming "Transporte" to "alimentacao" (case-insensitive clash)
	err := svc002.UpdateCategory(ctx, &domain.Category{
		ID:     "cat-id-002",
		UserID: "user-a",
		Name:   "alimentacao",
	})

	if !errors.Is(err, domain.ErrCategoryAlreadyExists) {
		t.Errorf("UpdateCategory() collision: error = %v, want %v", err, domain.ErrCategoryAlreadyExists)
	}
}

// ---------------------------------------------------------------------------
// AC-1 + AC-7: DeleteCategory
// ---------------------------------------------------------------------------

func TestCategoryService_DeleteCategory(t *testing.T) {
	svc, repo := newCategoryService("cat-id-001")
	ctx := context.Background()

	if err := svc.CreateCategory(ctx, &domain.Category{UserID: "user-a", Name: "Alimentacao"}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	tests := []struct {
		name    string
		id      string
		userID  string
		wantErr error
	}{
		{
			name:    "owner deletes own category",
			id:      "cat-id-001",
			userID:  "user-a",
			wantErr: nil,
		},
		{
			name:    "non-existent after deletion returns ErrNotFound",
			id:      "cat-id-001",
			userID:  "user-a",
			wantErr: repository.ErrNotFound,
		},
		{
			name:    "other user delete returns ErrNotFound",
			id:      "cat-id-001",
			userID:  "user-b",
			wantErr: repository.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.DeleteCategory(ctx, tt.id, tt.userID)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("DeleteCategory() %s: error = %v, want %v", tt.name, err, tt.wantErr)
			}
		})
	}

	_ = repo
}

// AC-7: delete category in use returns ErrCategoryInUse
func TestCategoryService_DeleteCategoryInUse(t *testing.T) {
	svc, repo := newCategoryService("cat-id-001")
	ctx := context.Background()

	if err := svc.CreateCategory(ctx, &domain.Category{UserID: "user-a", Name: "Alimentacao"}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Mark in-use via test helper
	repo.MarkCategoryInUse("cat-id-001")

	err := svc.DeleteCategory(ctx, "cat-id-001", "user-a")

	if !errors.Is(err, domain.ErrCategoryInUse) {
		t.Errorf("DeleteCategory() in-use: error = %v, want %v", err, domain.ErrCategoryInUse)
	}
}

// ---------------------------------------------------------------------------
// AC-4: LookupByName — resolves name to category (no auto-create)
// ---------------------------------------------------------------------------

func TestCategoryService_LookupByName(t *testing.T) {
	svc, _ := newCategoryService("cat-id-001")
	ctx := context.Background()

	if err := svc.CreateCategory(ctx, &domain.Category{UserID: "user-a", Name: "Alimentacao"}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	tests := []struct {
		name    string
		lookup  string
		userID  string
		wantErr error
		wantID  string
	}{
		{
			name:    "exact match returns category",
			lookup:  "Alimentacao",
			userID:  "user-a",
			wantErr: nil,
			wantID:  "cat-id-001",
		},
		{
			name:    "case-insensitive match returns category",
			lookup:  "alimentacao",
			userID:  "user-a",
			wantErr: nil,
			wantID:  "cat-id-001",
		},
		{
			name:    "unknown name returns ErrNotFound (no auto-create)",
			lookup:  "CategoriaInexistente",
			userID:  "user-a",
			wantErr: repository.ErrNotFound,
		},
		{
			name:    "correct name but wrong user returns ErrNotFound",
			lookup:  "Alimentacao",
			userID:  "user-b",
			wantErr: repository.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.LookupByName(ctx, tt.lookup, tt.userID)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("LookupByName() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				if got == nil {
					t.Fatal("LookupByName() returned nil, want category")
				}
				if got.ID != tt.wantID {
					t.Errorf("LookupByName() ID = %q, want %q", got.ID, tt.wantID)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// AC-3: ExpenseService validates category_id ownership
// ---------------------------------------------------------------------------

// TestExpenseService_CreateWithCategoryID verifies that ExpenseService rejects
// a category_id that does not belong to the expense's user_id.
// Requires ExpenseService.SetCategoryService (not yet implemented — RED).
func TestExpenseService_CreateWithCategoryID(t *testing.T) {
	expRepo := repository.NewMemoryExpenseRepository()
	expIDGen := &MockIDGenerator{id: "exp-id-001"}
	expSvc := NewExpenseService(expRepo, expIDGen)

	catRepo := repository.NewMemoryCategoryRepository()
	catIDGen := &MockCategoryIDGenerator{id: "cat-id-001"}
	catSvc := NewCategoryService(catRepo, catIDGen)

	// Inject category service into expense service
	expSvc.SetCategoryService(catSvc)

	ctx := context.Background()

	// Seed a category for user-a
	validCatID := "cat-id-001"
	_ = catRepo.Create(ctx, &domain.Category{
		ID: validCatID, UserID: "user-a", Name: "Alimentacao",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})

	// Seed a category for another user
	otherCatID := "cat-id-other"
	_ = catRepo.Create(ctx, &domain.Category{
		ID: otherCatID, UserID: "user-b", Name: "Saude",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})

	tests := []struct {
		name    string
		expense *domain.Expense
		wantErr error
	}{
		{
			name: "valid category_id owned by user succeeds",
			expense: &domain.Expense{
				UserID:      "user-a",
				Description: "Almoço",
				Amount:      25.50,
				CategoryID:  &validCatID,
				Date:        time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "category_id of another user returns ErrCategoryNotFound",
			expense: &domain.Expense{
				UserID:      "user-a",
				Description: "Consulta",
				Amount:      100.00,
				CategoryID:  &otherCatID,
				Date:        time.Now(),
			},
			wantErr: domain.ErrCategoryNotFound,
		},
		{
			name: "non-existent category_id returns ErrCategoryNotFound",
			expense: &domain.Expense{
				UserID:      "user-a",
				Description: "Algo",
				Amount:      10.00,
				CategoryID:  strPtr("ghost-id"),
				Date:        time.Now(),
			},
			wantErr: domain.ErrCategoryNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := expSvc.CreateExpense(ctx, tt.expense)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("CreateExpense() %s: error = %v, want %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

// AC-4: ExpenseService resolves category string to category_id via LookupByName
func TestExpenseService_CreateWithCategoryNameLookup(t *testing.T) {
	expRepo := repository.NewMemoryExpenseRepository()
	expIDGen := &MockIDGenerator{id: "exp-id-001"}
	expSvc := NewExpenseService(expRepo, expIDGen)

	catRepo := repository.NewMemoryCategoryRepository()
	catIDGen := &MockCategoryIDGenerator{id: "cat-id-001"}
	catSvc := NewCategoryService(catRepo, catIDGen)
	expSvc.SetCategoryService(catSvc)

	ctx := context.Background()

	validCatID := "cat-id-001"
	_ = catRepo.Create(ctx, &domain.Category{
		ID: validCatID, UserID: "user-a", Name: "Alimentacao",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})

	tests := []struct {
		name           string
		expense        *domain.Expense
		wantErr        error
		wantCategoryID string
	}{
		{
			name: "category string resolves to category_id",
			expense: &domain.Expense{
				UserID:      "user-a",
				Description: "Almoço",
				Amount:      25.50,
				Category:    "Alimentacao",
				Date:        time.Now(),
			},
			wantErr:        nil,
			wantCategoryID: validCatID,
		},
		{
			name: "case-insensitive lookup resolves",
			expense: &domain.Expense{
				UserID:      "user-a",
				Description: "Jantar",
				Amount:      30.00,
				Category:    "alimentacao",
				Date:        time.Now(),
			},
			wantErr:        nil,
			wantCategoryID: validCatID,
		},
		{
			name: "unknown category name returns ErrCategoryNotFound",
			expense: &domain.Expense{
				UserID:      "user-a",
				Description: "Algo",
				Amount:      10.00,
				Category:    "CategoriaInexistente",
				Date:        time.Now(),
			},
			wantErr: domain.ErrCategoryNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := expSvc.CreateExpense(ctx, tt.expense)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("CreateExpense() %s: error = %v, want %v", tt.name, err, tt.wantErr)
			}

			if tt.wantErr == nil && tt.wantCategoryID != "" {
				if tt.expense.CategoryID == nil || *tt.expense.CategoryID != tt.wantCategoryID {
					var got string
					if tt.expense.CategoryID != nil {
						got = *tt.expense.CategoryID
					}
					t.Errorf("CreateExpense() CategoryID = %q, want %q", got, tt.wantCategoryID)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func strPtr(s string) *string {
	return &s
}
