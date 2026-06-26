package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rodrigo/expense-tracker/internal/domain"
	"github.com/rodrigo/expense-tracker/internal/repository"
)

// CategoryIDGenerator gera IDs unicos para categorias
type CategoryIDGenerator interface {
	Generate() string
}

// CategoryServiceInterface define as operacoes de negocio para categorias
type CategoryServiceInterface interface {
	CreateCategory(ctx context.Context, category *domain.Category) error
	GetCategory(ctx context.Context, id, userID string) (*domain.Category, error)
	ListCategories(ctx context.Context, userID string) ([]*domain.Category, error)
	UpdateCategory(ctx context.Context, category *domain.Category) error
	DeleteCategory(ctx context.Context, id, userID string) error
	LookupByName(ctx context.Context, name, userID string) (*domain.Category, error)
}

// CategoryService contem a logica de negocio para categorias
type CategoryService struct {
	repo        repository.CategoryRepository
	idGenerator CategoryIDGenerator
}

// NewCategoryService cria uma nova instancia do CategoryService
func NewCategoryService(repo repository.CategoryRepository, idGen CategoryIDGenerator) *CategoryService {
	return &CategoryService{
		repo:        repo,
		idGenerator: idGen,
	}
}

// CreateCategory cria uma nova categoria para o usuario
func (s *CategoryService) CreateCategory(ctx context.Context, category *domain.Category) error {
	if err := category.Validate(); err != nil {
		return err
	}

	if category.ID == "" {
		category.ID = s.idGenerator.Generate()
	}

	now := time.Now().UTC()
	category.CreatedAt = now
	category.UpdatedAt = now

	// Em caso de colisão de ID, gerar um UUID globalmente único e tentar novamente
	if err := s.repo.Create(ctx, category); err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			category.ID = uuid.New().String()
			return s.repo.Create(ctx, category)
		}
		return err
	}
	return nil
}

// GetCategory retorna uma categoria pelo ID, scoped por userID
// Retorna repository.ErrNotFound se nao existir ou nao pertencer ao usuario
func (s *CategoryService) GetCategory(ctx context.Context, id, userID string) (*domain.Category, error) {
	return s.repo.GetByID(ctx, id, userID)
}

// ListCategories retorna todas as categorias do usuario
func (s *CategoryService) ListCategories(ctx context.Context, userID string) ([]*domain.Category, error) {
	return s.repo.GetAll(ctx, userID)
}

// UpdateCategory atualiza uma categoria existente
// Valida ownership e unicidade de nome
func (s *CategoryService) UpdateCategory(ctx context.Context, category *domain.Category) error {
	if err := category.Validate(); err != nil {
		return err
	}

	category.UpdatedAt = time.Now().UTC()

	return s.repo.Update(ctx, category)
}

// DeleteCategory remove uma categoria pelo ID, scoped por userID
// Retorna domain.ErrCategoryInUse se a categoria estiver em uso
func (s *CategoryService) DeleteCategory(ctx context.Context, id, userID string) error {
	return s.repo.Delete(ctx, id, userID)
}

// LookupByName busca uma categoria pelo nome (case-insensitive), scoped por userID
// Retorna repository.ErrNotFound se nao existir — sem auto-criacao
func (s *CategoryService) LookupByName(ctx context.Context, name, userID string) (*domain.Category, error) {
	return s.repo.GetByName(ctx, name, userID)
}
