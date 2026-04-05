package domain

import (
	"errors"
	"time"
)

// Category representa uma categoria de despesa pertencente a um usuario
type Category struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Erros sentinela de categoria
var (
	ErrCategoryNotFound      = errors.New("category not found")
	ErrCategoryAlreadyExists = errors.New("category already exists")
	ErrCategoryInUse         = errors.New("category is in use by one or more expenses")
	ErrEmptyCategoryName     = errors.New("category name cannot be empty")
	ErrCategoryNameTooLong   = errors.New("category name must be at most 100 characters")
)

// Validate valida os campos obrigatorios de uma categoria
func (c *Category) Validate() error {
	if c.Name == "" {
		return ErrEmptyCategoryName
	}
	if len(c.Name) > 100 {
		return ErrCategoryNameTooLong
	}
	return nil
}
