package domain

import (
	"errors"
	"time"
)

// Expense representa uma despesa pessoal
type Expense struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	// Category is the legacy free-text field. In responses it is populated with
	// the resolved category name via JOIN/lookup. In requests it is accepted as
	// a lookup key when category_id is not provided. Kept for backward compat;
	// will be deprecated in a future delivery.
	Category   string  `json:"category"`
	// CategoryID references the normalized categories table. Nullable: legacy
	// expenses created before migration 000006 will have nil here.
	CategoryID *string `json:"category_id"`
	Date        time.Time `json:"date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Erros de validação
var (
	ErrEmptyDescription = errors.New("description cannot be empty")
	ErrInvalidAmount    = errors.New("amount must be greater than zero")
	ErrEmptyCategory    = errors.New("category cannot be empty")
)

// Validate valida os campos obrigatórios de uma despesa.
// A despesa e valida quanto a categoria se ao menos um de category (string) ou
// category_id (UUID) estiver presente. Ambos ausentes retorna ErrEmptyCategory.
func (e *Expense) Validate() error {
	if e.Description == "" {
		return ErrEmptyDescription
	}
	if e.Amount <= 0 {
		return ErrInvalidAmount
	}
	if e.Category == "" && (e.CategoryID == nil || *e.CategoryID == "") {
		return ErrEmptyCategory
	}
	return nil
}
