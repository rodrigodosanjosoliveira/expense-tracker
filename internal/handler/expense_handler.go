package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/rodrigo/expense-tracker/internal/domain"
	"github.com/rodrigo/expense-tracker/internal/repository"
	"github.com/rodrigo/expense-tracker/internal/service"
)

// ExpenseHandler gerencia as requisições HTTP
type ExpenseHandler struct {
	service *service.ExpenseService
}

// NewExpenseHandler cria uma nova instância
func NewExpenseHandler(service *service.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{
		service: service,
	}
}

// CreateExpense handler para POST /expenses
// @Summary Criar nova despesa
// @Description Cria uma nova despesa pessoal
// @Tags Expenses
// @Accept json
// @Produce json
// @Param expense body domain.Expense true "Dados da despesa"
// @Success 201 {object} domain.Expense
// @Failure 400 {string} string "Dados inválidos"
// @Failure 500 {string} string "Erro interno"
// @Router /expenses [post]
func (h *ExpenseHandler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	// Conceito: http.ResponseWriter e *http.Request são os parâmetros padrão
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var expense domain.Expense
	// Conceito: Decoder para ler JSON do request body
	if err := json.NewDecoder(r.Body).Decode(&expense); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.service.CreateExpense(r.Context(), &expense); err != nil {
		switch err {
		case domain.ErrEmptyDescription, domain.ErrInvalidAmount, domain.ErrEmptyCategory:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Conceito: Setar header e status code
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(expense)
}

// GetExpense handler para GET /expenses/{id}
// @Summary Buscar despesa por ID
// @Description Retorna uma despesa específica pelo ID
// @Tags Expenses
// @Produce json
// @Param id path string true "ID da despesa"
// @Success 200 {object} domain.Expense
// @Failure 400 {string} string "ID inválido"
// @Failure 404 {string} string "Despesa não encontrada"
// @Failure 500 {string} string "Erro interno"
// @Router /expenses/{id} [get]
func (h *ExpenseHandler) GetExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extrair ID do path (simplificado - em produção use um router)
	id := strings.TrimPrefix(r.URL.Path, "/expenses/")
	if id == "" || id == "/expenses/" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	expense, err := h.service.GetExpense(r.Context(), id)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "Expense not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expense)
}

// ListExpenses handler para GET /expenses
// @Summary Listar despesas com filtros e paginação
// @Description Retorna despesas com suporte a filtros, ordenação e paginação
// @Tags Expenses
// @Produce json
// @Param category query string false "Filtrar por categoria"
// @Param min_amount query number false "Valor mínimo"
// @Param max_amount query number false "Valor máximo"
// @Param start_date query string false "Data inicial (formato: 2025-12-29)"
// @Param end_date query string false "Data final (formato: 2025-12-29)"
// @Param sort_by query string false "Ordenar por (date, amount, category, created_at)" default(date)
// @Param sort_order query string false "Ordem (asc, desc)" default(desc)
// @Param limit query int false "Itens por página" default(50)
// @Param offset query int false "Offset para paginação" default(0)
// @Param page query int false "Página (alternativa ao offset)" default(1)
// @Success 200 {object} domain.ExpenseListResponse
// @Failure 500 {string} string "Erro interno"
// @Router /expenses [get]
func (h *ExpenseHandler) ListExpenses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parsear filtros dos query parameters
	filters := ParseExpenseFilters(r)

	// Buscar com filtros
	response, err := h.service.ListExpensesWithFilters(r.Context(), filters)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateExpense handler para PUT /expenses/{id}
// @Summary Atualizar despesa
// @Description Atualiza uma despesa existente
// @Tags Expenses
// @Accept json
// @Produce json
// @Param id path string true "ID da despesa"
// @Param expense body domain.Expense true "Dados atualizados da despesa"
// @Success 200 {object} domain.Expense
// @Failure 400 {string} string "Dados inválidos"
// @Failure 404 {string} string "Despesa não encontrada"
// @Failure 500 {string} string "Erro interno"
// @Router /expenses/{id} [put]
func (h *ExpenseHandler) UpdateExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/expenses/")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	var expense domain.Expense
	if err := json.NewDecoder(r.Body).Decode(&expense); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	expense.ID = id

	if err := h.service.UpdateExpense(r.Context(), &expense); err != nil {
		switch err {
		case repository.ErrNotFound:
			http.Error(w, "Expense not found", http.StatusNotFound)
		case domain.ErrEmptyDescription, domain.ErrInvalidAmount, domain.ErrEmptyCategory:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expense)
}

// DeleteExpense handler para DELETE /expenses/{id}
// @Summary Deletar despesa
// @Description Remove uma despesa pelo ID
// @Tags Expenses
// @Param id path string true "ID da despesa"
// @Success 204 "Despesa deletada com sucesso"
// @Failure 400 {string} string "ID inválido"
// @Failure 404 {string} string "Despesa não encontrada"
// @Failure 500 {string} string "Erro interno"
// @Router /expenses/{id} [delete]
func (h *ExpenseHandler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/expenses/")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteExpense(r.Context(), id); err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "Expense not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
