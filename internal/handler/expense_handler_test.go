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
		},
		{
			name: "invalid - empty description",
			body: domain.Expense{
				Description: "",
				Amount:      25.50,
				Category:    "Alimentação",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid JSON",
			body:       "invalid json",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Conceito: Marshal converte struct para JSON
			body, _ := json.Marshal(tt.body)
			// Conceito: httptest.NewRequest cria uma request de teste
			req := withUserID(httptest.NewRequest(http.MethodPost, ExpensesPath, bytes.NewReader(body)))
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
