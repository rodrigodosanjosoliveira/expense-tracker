package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rodrigo/expense-tracker/internal/domain"
	"github.com/rodrigo/expense-tracker/internal/middleware"
	"github.com/rodrigo/expense-tracker/internal/repository"
	"github.com/rodrigo/expense-tracker/internal/service"
)

const (
	ExpensesPath    = "/expenses"
	ExpenseByIDPath = "/expenses/123"
	testUserID      = "test-user-id"
)

// setupHandlerWithCategories creates an ExpenseHandler whose underlying
// ExpenseService is aware of the CategoryService backed by the provided
// MemoryCategoryRepository. This setup is required for AC-3, AC-4, AC-5,
// AC-6 tests which involve category_id resolution and validation.
//
// CategoryService and the SetCategoryService method on ExpenseService are not
// yet implemented — these tests will fail to compile until Dev delivers them.
func setupHandlerWithCategories(catRepo *repository.MemoryCategoryRepository) (*ExpenseHandler, *repository.MemoryExpenseRepository) {
	expRepo := repository.NewMemoryExpenseRepository()
	idGen := &mockIDGen{}

	catIDGen := &mockCategoryIDGen{}
	catSvc := service.NewCategoryService(catRepo, catIDGen)

	svc := service.NewExpenseService(expRepo, idGen)
	svc.SetCategoryService(catSvc)

	return NewExpenseHandler(svc), expRepo
}

type mockIDGen struct{}

func (m *mockIDGen) Generate() string {
	return "test-id-123"
}

func setupHandler() *ExpenseHandler {
	repo := repository.NewMemoryExpenseRepository()
	idGen := &mockIDGen{}
	svc := service.NewExpenseService(repo, idGen)
	return NewExpenseHandler(svc)
}

func withUserID(r *http.Request) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.UserIDKey, testUserID)
	return r.WithContext(ctx)
}

func TestExpenseHandlerCreateExpense(t *testing.T) {
	handler := setupHandler()

	tests := []struct {
		name       string
		body       interface{}
		wantStatus int
		addUserID  bool
	}{
		{
			name: "valid expense",
			body: domain.Expense{
				Description: "Almoço",
				Amount:      25.50,
				Category:    "Alimentação",
				Date:        time.Now(),
			},
			wantStatus: http.StatusCreated,
			addUserID:  true,
		},
		{
			name: "invalid - empty description",
			body: domain.Expense{
				Description: "",
				Amount:      25.50,
				Category:    "Alimentação",
			},
			wantStatus: http.StatusBadRequest,
			addUserID:  true,
		},
		{
			name:       "invalid JSON",
			body:       "invalid json",
			wantStatus: http.StatusBadRequest,
			addUserID:  true,
		},
		{
			name: "unauthorized - missing user ID",
			body: domain.Expense{
				Description: "Jantar",
				Amount:      30.00,
				Category:    "Alimentação",
				Date:        time.Now(),
			},
			wantStatus: http.StatusUnauthorized,
			addUserID:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Conceito: Marshal converte struct para JSON
			body, _ := json.Marshal(tt.body)
			// Conceito: httptest.NewRequest cria uma request de teste
			req := httptest.NewRequest(http.MethodPost, ExpensesPath, bytes.NewReader(body))
			if tt.addUserID {
				req = withUserID(req)
			}
			// Conceito: httptest.NewRecorder grava a resposta
			w := httptest.NewRecorder()

			handler.CreateExpense(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("CreateExpense() status = %d, want %d", w.Code, tt.wantStatus)
			}

			// Verificar que retornou JSON válido em caso de sucesso
			if tt.wantStatus == http.StatusCreated {
				var result domain.Expense
				if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
					t.Errorf("CreateExpense() failed to decode response: %v", err)
				}
				if result.ID != "test-id-123" {
					t.Errorf("CreateExpense() ID = %v, want test-id-123", result.ID)
				}
			}
		})
	}
}

func TestExpenseHandlerGetExpense(t *testing.T) {
	handler := setupHandler()

	// Criar uma despesa primeiro
	expense := &domain.Expense{
		ID:          "123",
		Description: "Teste",
		Amount:      10.00,
		Category:    "Cat1",
		Date:        time.Now(),
		UserID:      testUserID,
	}
	handler.service.CreateExpense(context.TODO(), expense)

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{
			name:       "existing expense",
			path:       ExpenseByIDPath,
			wantStatus: http.StatusOK,
		},
		{
			name:       "non-existing expense",
			path:       ExpensesPath + "/999",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "empty ID",
			path:       ExpensesPath + "/",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := withUserID(httptest.NewRequest(http.MethodGet, tt.path, nil))
			w := httptest.NewRecorder()

			handler.GetExpense(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("GetExpense() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestExpenseHandlerListExpenses(t *testing.T) {
	handler := setupHandler()

	// Criar algumas despesas
	handler.service.CreateExpense(context.TODO(), &domain.Expense{
		Description: "A",
		Amount:      10,
		Category:    "Cat1",
		Date:        time.Now(),
		UserID:      testUserID,
	})

	req := withUserID(httptest.NewRequest(http.MethodGet, ExpensesPath, nil))
	w := httptest.NewRecorder()

	handler.ListExpenses(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ListExpenses() status = %d, want %d", w.Code, http.StatusOK)
	}

	var response domain.ExpenseListResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Errorf("ListExpenses() failed to decode response: %v", err)
	}

	if len(response.Data) != 1 {
		t.Errorf("ListExpenses() got %d expenses, want 1", len(response.Data))
	}

	if response.Pagination == nil {
		t.Error("ListExpenses() pagination is nil")
	}

	if response.Pagination.Total != 1 {
		t.Errorf("ListExpenses() total = %d, want 1", response.Pagination.Total)
	}
}

func TestExpenseHandlerUpdateExpense(t *testing.T) {
	handler := setupHandler()

	// Criar uma despesa
	expense := &domain.Expense{
		ID:          "123",
		Description: "Original",
		Amount:      10.00,
		Category:    "Cat1",
		Date:        time.Now(),
		UserID:      testUserID,
	}
	handler.service.CreateExpense(context.TODO(), expense)

	updatedExpense := domain.Expense{
		Description: "Atualizado",
		Amount:      20.00,
		Category:    "Cat1",
		Date:        time.Now(),
	}

	body, _ := json.Marshal(updatedExpense)
	req := withUserID(httptest.NewRequest(http.MethodPut, ExpenseByIDPath, bytes.NewReader(body)))
	w := httptest.NewRecorder()

	handler.UpdateExpense(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("UpdateExpense() status = %d, want %d", w.Code, http.StatusOK)
	}

	var result domain.Expense
	json.NewDecoder(w.Body).Decode(&result)
	if result.Description != "Atualizado" {
		t.Errorf("UpdateExpense() description = %v, want Atualizado", result.Description)
	}
}

func TestExpenseHandlerDeleteExpense(t *testing.T) {
	handler := setupHandler()

	// Criar uma despesa
	expense := &domain.Expense{
		ID:          "123",
		Description: "Teste",
		Amount:      10.00,
		Category:    "Cat1",
		Date:        time.Now(),
		UserID:      testUserID,
	}
	handler.service.CreateExpense(context.TODO(), expense)

	req := withUserID(httptest.NewRequest(http.MethodDelete, ExpenseByIDPath, nil))
	w := httptest.NewRecorder()

	handler.DeleteExpense(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("DeleteExpense() status = %d, want %d", w.Code, http.StatusNoContent)
	}

	// Verificar que foi deletado
	req2 := withUserID(httptest.NewRequest(http.MethodGet, "/expenses/123", nil))
	w2 := httptest.NewRecorder()
	handler.GetExpense(w2, req2)

	if w2.Code != http.StatusNotFound {
		t.Errorf("GetExpense() after delete status = %d, want %d", w2.Code, http.StatusNotFound)
	}
}

func TestExpenseHandlerUnauthorized(t *testing.T) {
	handler := setupHandler()

	tests := []struct {
		name    string
		handler http.HandlerFunc
		method  string
		path    string
		body    interface{}
	}{
		{
			name:    "CreateExpense without auth",
			handler: handler.CreateExpense,
			method:  http.MethodPost,
			path:    ExpensesPath,
			body:    domain.Expense{Description: "Test", Amount: 10, Category: "Cat"},
		},
		{
			name:    "ListExpenses without auth",
			handler: handler.ListExpenses,
			method:  http.MethodGet,
			path:    ExpensesPath,
		},
		{
			name:    "GetExpense without auth",
			handler: handler.GetExpense,
			method:  http.MethodGet,
			path:    ExpenseByIDPath,
		},
		{
			name:    "UpdateExpense without auth",
			handler: handler.UpdateExpense,
			method:  http.MethodPut,
			path:    ExpenseByIDPath,
			body:    domain.Expense{Description: "Updated", Amount: 20, Category: "Cat"},
		},
		{
			name:    "DeleteExpense without auth",
			handler: handler.DeleteExpense,
			method:  http.MethodDelete,
			path:    ExpenseByIDPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyBytes []byte
			if tt.body != nil {
				bodyBytes, _ = json.Marshal(tt.body)
			}
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			tt.handler(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("%s: got status %d, want %d (Unauthorized)", tt.name, w.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestExpenseHandlerMethodNotAllowed(t *testing.T) {
	handler := setupHandler()

	// Testar método inválido em cada endpoint
	tests := []struct {
		name       string
		method     string
		path       string
		handler    http.HandlerFunc
		wantStatus int
	}{
		{
			name:       "GET on create",
			method:     http.MethodGet,
			path:       ExpensesPath,
			handler:    handler.CreateExpense,
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "POST on get",
			method:     http.MethodPost,
			path:       ExpenseByIDPath,
			handler:    handler.GetExpense,
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			tt.handler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// AC-3: Expense references category by ID — category_id validation
// ---------------------------------------------------------------------------

// TestExpenseCreateWithCategoryID verifies that POST /expenses accepts category_id
// and that a category_id belonging to another user returns HTTP 422.
func TestExpenseCreateWithCategoryID(t *testing.T) {
	catRepo := repository.NewMemoryCategoryRepository()
	h, _ := setupHandlerWithCategories(catRepo)

	// Seed a category for testUserID
	catID := "valid-cat-id"
	seedCategoryViaRepo(catRepo, catID, "Alimentacao", testUserID)

	// Seed a category for another user — must not be usable by testUserID
	otherCatID := "other-cat-id"
	seedCategoryViaRepo(catRepo, otherCatID, "Saude", "other-user")

	tests := []struct {
		name       string
		body       map[string]interface{}
		wantStatus int
	}{
		{
			name: "valid category_id returns 201",
			body: map[string]interface{}{
				"description": "Almoço",
				"amount":      25.50,
				"category_id": catID,
				"date":        time.Now(),
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "category_id of another user returns 422",
			body: map[string]interface{}{
				"description": "Consulta",
				"amount":      100.00,
				"category_id": otherCatID,
				"date":        time.Now(),
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "non-existent category_id returns 422",
			body: map[string]interface{}{
				"description": "Algo",
				"amount":      50.00,
				"category_id": "does-not-exist",
				"date":        time.Now(),
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := withUserID(httptest.NewRequest(http.MethodPost, ExpensesPath, bytes.NewReader(body)))
			w := httptest.NewRecorder()
			h.CreateExpense(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("CreateExpense() %s: status = %d, want %d (body: %s)", tt.name, w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// AC-3: PUT /expenses/{id} — same category_id validation on update
func TestExpenseUpdateWithCategoryID(t *testing.T) {
	catRepo := repository.NewMemoryCategoryRepository()
	h, expRepo := setupHandlerWithCategories(catRepo)

	catID := "valid-cat-id"
	seedCategoryViaRepo(catRepo, catID, "Alimentacao", testUserID)

	// Pre-create an expense to update
	exp := &domain.Expense{
		ID:          "123",
		UserID:      testUserID,
		Description: "Original",
		Amount:      10.00,
		Category:    "Alimentacao",
		CategoryID:  &catID,
		Date:        time.Now(),
	}
	_ = expRepo.Create(context.Background(), exp)

	tests := []struct {
		name       string
		body       map[string]interface{}
		wantStatus int
	}{
		{
			name: "valid category_id on update returns 200",
			body: map[string]interface{}{
				"description": "Atualizado",
				"amount":      30.00,
				"category_id": catID,
				"date":        time.Now(),
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "non-existent category_id on update returns 422",
			body: map[string]interface{}{
				"description": "Atualizado",
				"amount":      30.00,
				"category_id": "ghost-cat-id",
				"date":        time.Now(),
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := withUserID(httptest.NewRequest(http.MethodPut, ExpenseByIDPath, bytes.NewReader(body)))
			w := httptest.NewRecorder()
			h.UpdateExpense(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("UpdateExpense() %s: status = %d, want %d (body: %s)", tt.name, w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// AC-4: category string lookup — resolves name to category_id
// ---------------------------------------------------------------------------

// TestExpenseCreateWithCategoryNameLookup verifies that POST /expenses accepts a
// category (string) field, resolves it via name lookup, and sets category_id.
// If the category name does not exist, it must return HTTP 422 (no auto-create).
func TestExpenseCreateWithCategoryNameLookup(t *testing.T) {
	catRepo := repository.NewMemoryCategoryRepository()
	h, _ := setupHandlerWithCategories(catRepo)

	// Seed a category for testUserID
	seedCategoryViaRepo(catRepo, "cat-alim-id", "Alimentacao", testUserID)

	tests := []struct {
		name           string
		body           map[string]interface{}
		wantStatus     int
		wantCategoryID string
	}{
		{
			name: "category string resolves to existing category_id",
			body: map[string]interface{}{
				"description": "Almoço",
				"amount":      25.50,
				"category":    "Alimentacao",
				"date":        time.Now(),
			},
			wantStatus:     http.StatusCreated,
			wantCategoryID: "cat-alim-id",
		},
		{
			name: "case-insensitive category string resolves",
			body: map[string]interface{}{
				"description": "Jantar",
				"amount":      30.00,
				"category":    "alimentacao",
				"date":        time.Now(),
			},
			wantStatus:     http.StatusCreated,
			wantCategoryID: "cat-alim-id",
		},
		{
			name: "unknown category string returns 422 (no auto-create)",
			body: map[string]interface{}{
				"description": "Algo",
				"amount":      10.00,
				"category":    "CategoriaInexistente",
				"date":        time.Now(),
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := withUserID(httptest.NewRequest(http.MethodPost, ExpensesPath, bytes.NewReader(body)))
			w := httptest.NewRecorder()
			h.CreateExpense(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("CreateExpense() %s: status = %d, want %d (body: %s)", tt.name, w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantStatus == http.StatusCreated && tt.wantCategoryID != "" {
				var result domain.Expense
				if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
					t.Fatalf("decode failed: %v", err)
				}
				if result.CategoryID == nil || *result.CategoryID != tt.wantCategoryID {
					var got string
					if result.CategoryID != nil {
						got = *result.CategoryID
					}
					t.Errorf("CreateExpense() CategoryID = %q, want %q", got, tt.wantCategoryID)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// AC-5: GET /expenses and GET /expenses/{id} return category and category_id
// ---------------------------------------------------------------------------

func TestExpenseResponseIncludesCategoryFields(t *testing.T) {
	catRepo := repository.NewMemoryCategoryRepository()
	h, expRepo := setupHandlerWithCategories(catRepo)

	catID := "cat-resp-id"
	seedCategoryViaRepo(catRepo, catID, "Alimentacao", testUserID)

	exp := &domain.Expense{
		ID:          "123",
		UserID:      testUserID,
		Description: "Almoço",
		Amount:      25.50,
		Category:    "Alimentacao",
		CategoryID:  &catID,
		Date:        time.Now(),
	}
	_ = expRepo.Create(context.Background(), exp)

	t.Run("GET /expenses returns category and category_id", func(t *testing.T) {
		req := withUserID(httptest.NewRequest(http.MethodGet, ExpensesPath, nil))
		w := httptest.NewRecorder()
		h.ListExpenses(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("ListExpenses() status = %d, want %d", w.Code, http.StatusOK)
		}

		var response domain.ExpenseListResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if len(response.Data) == 0 {
			t.Fatal("ListExpenses() returned empty data")
		}

		got := response.Data[0]
		if got.CategoryID == nil || *got.CategoryID != catID {
			t.Errorf("ListExpenses() CategoryID = %v, want %q", got.CategoryID, catID)
		}
		if got.Category == "" {
			t.Error("ListExpenses() Category name is empty, expected resolved name")
		}
	})

	t.Run("GET /expenses/{id} returns category and category_id", func(t *testing.T) {
		req := withUserID(httptest.NewRequest(http.MethodGet, ExpenseByIDPath, nil))
		w := httptest.NewRecorder()
		h.GetExpense(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("GetExpense() status = %d, want %d", w.Code, http.StatusOK)
		}

		var result domain.Expense
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if result.CategoryID == nil || *result.CategoryID != catID {
			t.Errorf("GetExpense() CategoryID = %v, want %q", result.CategoryID, catID)
		}
		if result.Category == "" {
			t.Error("GetExpense() Category name is empty, expected resolved name")
		}
	})
}

// ---------------------------------------------------------------------------
// AC-6: Filter by category name and category_id — both scoped by user_id
// ---------------------------------------------------------------------------

func TestExpenseListFilterByCategory(t *testing.T) {
	catRepo := repository.NewMemoryCategoryRepository()
	h, expRepo := setupHandlerWithCategories(catRepo)

	catAlimID := "cat-alim"
	catTranspID := "cat-transp"
	seedCategoryViaRepo(catRepo, catAlimID, "Alimentacao", testUserID)
	seedCategoryViaRepo(catRepo, catTranspID, "Transporte", testUserID)

	// Seed expenses for testUserID
	_ = expRepo.Create(context.Background(), &domain.Expense{
		ID: "exp-1", UserID: testUserID,
		Description: "Almoço", Amount: 25, Category: "Alimentacao", CategoryID: &catAlimID,
		Date: time.Now(),
	})
	_ = expRepo.Create(context.Background(), &domain.Expense{
		ID: "exp-2", UserID: testUserID,
		Description: "Uber", Amount: 15, Category: "Transporte", CategoryID: &catTranspID,
		Date: time.Now(),
	})
	// Seed expense for another user — must never appear
	otherCatID := "other-cat"
	seedCategoryViaRepo(catRepo, otherCatID, "Alimentacao", "other-user")
	_ = expRepo.Create(context.Background(), &domain.Expense{
		ID: "exp-3", UserID: "other-user",
		Description: "Outro", Amount: 99, Category: "Alimentacao", CategoryID: &otherCatID,
		Date: time.Now(),
	})

	t.Run("filter by category name (case-insensitive) returns only user's matching expenses", func(t *testing.T) {
		req := withUserID(httptest.NewRequest(http.MethodGet, ExpensesPath+"?category=alimentacao", nil))
		w := httptest.NewRecorder()
		h.ListExpenses(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("ListExpenses() status = %d, want %d", w.Code, http.StatusOK)
		}

		var resp domain.ExpenseListResponse
		_ = json.NewDecoder(w.Body).Decode(&resp)

		if len(resp.Data) != 1 {
			t.Errorf("filter by category name: got %d expenses, want 1", len(resp.Data))
		}
		if len(resp.Data) > 0 && resp.Data[0].UserID != testUserID {
			t.Errorf("filter by category name: returned expense from other user")
		}
	})

	t.Run("filter by category_id returns only user's matching expenses", func(t *testing.T) {
		req := withUserID(httptest.NewRequest(http.MethodGet, ExpensesPath+"?category_id="+catAlimID, nil))
		w := httptest.NewRecorder()
		h.ListExpenses(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("ListExpenses() status = %d, want %d", w.Code, http.StatusOK)
		}

		var resp domain.ExpenseListResponse
		_ = json.NewDecoder(w.Body).Decode(&resp)

		if len(resp.Data) != 1 {
			t.Errorf("filter by category_id: got %d expenses, want 1", len(resp.Data))
		}
		if len(resp.Data) > 0 {
			got := resp.Data[0]
			if got.CategoryID == nil || *got.CategoryID != catAlimID {
				t.Errorf("filter by category_id: returned expense with wrong category_id")
			}
			if got.UserID != testUserID {
				t.Error("filter by category_id: returned expense from other user")
			}
		}
	})
}

// ---------------------------------------------------------------------------
// AC-8: Legacy expenses without category_id return null category_id and empty category
// ---------------------------------------------------------------------------

func TestExpenseLegacyNoCategoryID(t *testing.T) {
	catRepo := repository.NewMemoryCategoryRepository()
	h, expRepo := setupHandlerWithCategories(catRepo)

	// Legacy expense: CategoryID is nil, Category is empty string
	_ = expRepo.Create(context.Background(), &domain.Expense{
		ID:          "legacy-exp",
		UserID:      testUserID,
		Description: "Compra antiga",
		Amount:      50.00,
		Category:    "",
		CategoryID:  nil,
		Date:        time.Now(),
	})

	t.Run("GET /expenses/{id} for legacy expense returns null category_id and empty category", func(t *testing.T) {
		req := withUserID(httptest.NewRequest(http.MethodGet, "/expenses/legacy-exp", nil))
		w := httptest.NewRecorder()
		h.GetExpense(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("GetExpense() legacy: status = %d, want %d (body: %s)", w.Code, http.StatusOK, w.Body.String())
		}

		// Decode as raw map to check null vs absent
		var raw map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&raw); err != nil {
			t.Fatalf("decode failed: %v", err)
		}

		// category_id must be present in JSON and null
		catIDVal, exists := raw["category_id"]
		if !exists {
			t.Error("GetExpense() legacy: JSON response missing 'category_id' key")
		}
		if catIDVal != nil {
			t.Errorf("GetExpense() legacy: category_id = %v, want null", catIDVal)
		}

		// category must be present and empty string (not absent)
		catVal, exists := raw["category"]
		if !exists {
			t.Error("GetExpense() legacy: JSON response missing 'category' key")
		}
		if catStr, ok := catVal.(string); ok && catStr != "" {
			t.Errorf("GetExpense() legacy: category = %q, want empty string", catStr)
		}
	})

	t.Run("GET /expenses list includes legacy expense with null category_id", func(t *testing.T) {
		req := withUserID(httptest.NewRequest(http.MethodGet, ExpensesPath, nil))
		w := httptest.NewRecorder()
		h.ListExpenses(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("ListExpenses() status = %d, want %d", w.Code, http.StatusOK)
		}

		var response domain.ExpenseListResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("decode failed: %v", err)
		}

		found := false
		for _, exp := range response.Data {
			if exp.ID == "legacy-exp" {
				found = true
				if exp.CategoryID != nil {
					t.Errorf("legacy expense CategoryID = %v, want nil", exp.CategoryID)
				}
				// Category string should be empty (not panic / not other value)
				if exp.Category != "" {
					t.Errorf("legacy expense Category = %q, want empty string", exp.Category)
				}
			}
		}
		if !found {
			t.Error("ListExpenses() did not include legacy expense")
		}
	})
}
