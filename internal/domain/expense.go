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
	Category    string    `json:"category"`
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

// Validate valida os campos obrigatórios de uma despesa
func (e *Expense) Validate() error {
	if e.Description == "" {
		return ErrEmptyDescription
	}
	if e.Amount <= 0 {
		return ErrInvalidAmount
	}
	if e.Category == "" {
		return ErrEmptyCategory
	}
	return nil
}
