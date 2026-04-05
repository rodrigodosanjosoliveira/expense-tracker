package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/rodrigo/expense-tracker/internal/domain"
)

// ParseExpenseFilters extrai filtros dos query parameters da requisição
// Conceito: Parsing de query strings em Go
func ParseExpenseFilters(r *http.Request) *domain.ExpenseFilters {
	filters := domain.NewExpenseFilters()
	query := r.URL.Query()

	// Filtro por categoria (nome, case-insensitive)
	if category := query.Get("category"); category != "" {
		filters.Category = &category
	}

	// Filtro por category_id (UUID exato)
	if categoryID := query.Get("category_id"); categoryID != "" {
		filters.CategoryID = &categoryID
	}

	// Filtro por valor mínimo
	if minAmountStr := query.Get("min_amount"); minAmountStr != "" {
		if minAmount, err := strconv.ParseFloat(minAmountStr, 64); err == nil {
			filters.MinAmount = &minAmount
		}
	}

	// Filtro por valor máximo
	if maxAmountStr := query.Get("max_amount"); maxAmountStr != "" {
		if maxAmount, err := strconv.ParseFloat(maxAmountStr, 64); err == nil {
			filters.MaxAmount = &maxAmount
		}
	}

	// Filtro por data inicial (formato: 2025-12-29)
	if startDateStr := query.Get("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filters.StartDate = &startDate
		}
	}

	// Filtro por data final (formato: 2025-12-29)
	if endDateStr := query.Get("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			// Ajustar para final do dia
			endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			filters.EndDate = &endDate
		}
	}

	// Ordenação
	if sortBy := query.Get("sort_by"); sortBy != "" {
		filters.SortBy = sortBy
	}

	if sortOrder := query.Get("sort_order"); sortOrder != "" {
		filters.SortOrder = sortOrder
	}

	// Paginação
	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filters.Limit = limit
		}
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filters.Offset = offset
		}
	}

	// Também suportar "page" como alternativa ao offset
	if pageStr := query.Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filters.Offset = (page - 1) * filters.Limit
		}
	}

	return filters
}
