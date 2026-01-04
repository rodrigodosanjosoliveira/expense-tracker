package repository

import (
	"context"
	"errors"

	"github.com/rodrigo/expense-tracker/internal/domain"
)

// Erros específicos do user repository
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrDuplicateUsername = errors.New("username already taken")
	ErrDuplicateEmail    = errors.New("email already registered")
)

// UserRepository define as operações de persistência para usuários
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
}
