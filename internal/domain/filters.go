package domain

import "time"

// ExpenseFilters define os filtros disponíveis para listar despesas
type ExpenseFilters struct {
	// Filtro de usuário (não exposto no JSON - definido pelo middleware)
	UserID *string `json:"-"`

	// Filtros
	Category  *string    `json:"category,omitempty"`
	MinAmount *float64   `json:"min_amount,omitempty"`
	MaxAmount *float64   `json:"max_amount,omitempty"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`

	// Ordenação
	SortBy    string `json:"sort_by,omitempty"`    // date, amount, category, created_at
	SortOrder string `json:"sort_order,omitempty"` // asc, desc

	// Paginação
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// NewExpenseFilters cria filtros com valores padrão
func NewExpenseFilters() *ExpenseFilters {
	return &ExpenseFilters{
		SortBy:    "date",
		SortOrder: "desc",
		Limit:     50,
		Offset:    0,
	}
}

// Validate valida os filtros
func (f *ExpenseFilters) Validate() error {
	// Validar sort_by
	validSortBy := map[string]bool{
		"date":       true,
		"amount":     true,
		"category":   true,
		"created_at": true,
	}
	if f.SortBy != "" && !validSortBy[f.SortBy] {
		f.SortBy = "date" // Default
	}

	// Validar sort_order
	if f.SortOrder != "asc" && f.SortOrder != "desc" {
		f.SortOrder = "desc" // Default
	}

	// Validar limit
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Limit > 100 {
		f.Limit = 100 // Máximo
	}

	// Validar offset
	if f.Offset < 0 {
		f.Offset = 0
	}

	// Validar valores
	if f.MinAmount != nil && *f.MinAmount < 0 {
		minAmount := 0.0
		f.MinAmount = &minAmount
	}

	if f.MaxAmount != nil && *f.MaxAmount < 0 {
		f.MaxAmount = nil
	}

	// Validar datas
	if f.StartDate != nil && f.EndDate != nil {
		if f.StartDate.After(*f.EndDate) {
			// Trocar se estiverem invertidas
			f.StartDate, f.EndDate = f.EndDate, f.StartDate
		}
	}

	return nil
}

// PaginationInfo informações sobre a paginação
type PaginationInfo struct {
	Total      int  `json:"total"`
	Limit      int  `json:"limit"`
	Offset     int  `json:"offset"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
	TotalPages int  `json:"total_pages"`
}

// ExpenseListResponse resposta com paginação
type ExpenseListResponse struct {
	Data       []*Expense      `json:"data"`
	Pagination *PaginationInfo `json:"pagination"`
}
