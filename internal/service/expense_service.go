package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rodrigo/expense-tracker/internal/domain"
	"github.com/rodrigo/expense-tracker/internal/repository"
)

// ExpenseService contém a lógica de negócio
type ExpenseService struct {
	repo                repository.ExpenseRepository
	idGenerator         IDGenerator
	notificationService *NotificationService  // Opcional: para notificações
	categoryService     CategoryServiceInterface // Opcional: para validação de categoria
}

// IDGenerator é uma interface para gerar IDs únicos
// Conceito: Facilita testes com mocks
type IDGenerator interface {
	Generate() string
}

// NewExpenseService cria uma nova instância do service
func NewExpenseService(repo repository.ExpenseRepository, idGen IDGenerator) *ExpenseService {
	return &ExpenseService{
		repo:        repo,
		idGenerator: idGen,
	}
}

// SetNotificationService define o serviço de notificações (opcional)
func (s *ExpenseService) SetNotificationService(notificationService *NotificationService) {
	s.notificationService = notificationService
}

// SetCategoryService injeta o servico de categorias para validacao de category_id
func (s *ExpenseService) SetCategoryService(categorySvc CategoryServiceInterface) {
	s.categoryService = categorySvc
}

// CreateExpense cria uma nova despesa
func (s *ExpenseService) CreateExpense(ctx context.Context, expense *domain.Expense) error {
	// Validar
	if err := expense.Validate(); err != nil {
		return err
	}

	// Validar/resolver categoria via CategoryService (se injetado)
	if s.categoryService != nil {
		if err := s.resolveCategoryForExpense(ctx, expense); err != nil {
			return err
		}
	}

	// Gerar ID se não fornecido
	if expense.ID == "" {
		expense.ID = s.idGenerator.Generate()
	}

	// Setar timestamps
	now := time.Now().UTC()
	expense.CreatedAt = now
	expense.UpdatedAt = now

	// Persistir; em caso de colisão de ID, gerar um UUID globalmente único e tentar novamente
	if err := s.repo.Create(ctx, expense); err != nil {
		if err == repository.ErrAlreadyExists {
			expense.ID = uuid.New().String()
			if err2 := s.repo.Create(ctx, expense); err2 != nil {
				return err2
			}
		} else {
			return err
		}
	}

	// Disparar notificações (se configurado)
	if s.notificationService != nil && expense.UserID != "" {
		// Notificar criação de despesa
		eventData := domain.ExpenseEventData{
			Expense: expense,
			Action:  "created",
		}

		// Disparar em goroutine para não bloquear a resposta
		go s.notificationService.TriggerEvent(context.Background(), expense.UserID, domain.WebhookEventExpenseCreated, eventData)
	}

	return nil
}

// GetExpense retorna uma despesa pelo ID
func (s *ExpenseService) GetExpense(ctx context.Context, id string) (*domain.Expense, error) {
	return s.repo.GetByID(ctx, id)
}

// ListExpenses retorna todas as despesas
func (s *ExpenseService) ListExpenses(ctx context.Context) ([]*domain.Expense, error) {
	return s.repo.GetAll(ctx)
}

// UpdateExpense atualiza uma despesa existente
func (s *ExpenseService) UpdateExpense(ctx context.Context, expense *domain.Expense) error {
	// Validar
	if err := expense.Validate(); err != nil {
		return err
	}

	// Validar/resolver categoria via CategoryService (se injetado)
	if s.categoryService != nil {
		if err := s.resolveCategoryForExpense(ctx, expense); err != nil {
			return err
		}
	}

	// Verificar se existe
	existing, err := s.repo.GetByID(ctx, expense.ID)
	if err != nil {
		return err
	}

	// Preservar CreatedAt e atualizar UpdatedAt
	expense.CreatedAt = existing.CreatedAt
	expense.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, expense); err != nil {
		return err
	}

	// Disparar notificações (se configurado)
	if s.notificationService != nil && expense.UserID != "" {
		eventData := domain.ExpenseEventData{
			Expense: expense,
			Action:  "updated",
		}
		go s.notificationService.TriggerEvent(context.Background(), expense.UserID, domain.WebhookEventExpenseUpdated, eventData)
	}

	return nil
}

// DeleteExpense remove uma despesa
func (s *ExpenseService) DeleteExpense(ctx context.Context, id string) error {
	// Buscar despesa antes de deletar (para notificações)
	expense, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Disparar notificações (se configurado)
	if s.notificationService != nil && expense.UserID != "" {
		eventData := domain.ExpenseEventData{
			Expense: expense,
			Action:  "deleted",
		}
		go s.notificationService.TriggerEvent(context.Background(), expense.UserID, domain.WebhookEventExpenseDeleted, eventData)
	}

	return nil
}

// ListExpensesWithFilters retorna despesas com filtros, ordenação e paginação
func (s *ExpenseService) ListExpensesWithFilters(ctx context.Context, filters *domain.ExpenseFilters) (*domain.ExpenseListResponse, error) {
	// Validar filtros
	if err := filters.Validate(); err != nil {
		return nil, err
	}

	// Buscar despesas
	expenses, err := s.repo.GetAllWithFilters(ctx, filters)
	if err != nil {
		return nil, err
	}

	// Contar total
	total, err := s.repo.Count(ctx, filters)
	if err != nil {
		return nil, err
	}

	// Calcular informações de paginação
	pagination := &domain.PaginationInfo{
		Total:      total,
		Limit:      filters.Limit,
		Offset:     filters.Offset,
		HasNext:    filters.Offset+filters.Limit < total,
		HasPrev:    filters.Offset > 0,
		TotalPages: (total + filters.Limit - 1) / filters.Limit, // Arredondar para cima
	}

	return &domain.ExpenseListResponse{
		Data:       expenses,
		Pagination: pagination,
	}, nil
}

// resolveCategoryForExpense valida ou resolve a categoria de uma despesa.
// Prioridade: category_id (validar ownership) > category string (lookup por nome).
// Se nenhum dos dois estiver presente, o comportamento legado e mantido (Validate ja aceitou).
func (s *ExpenseService) resolveCategoryForExpense(ctx context.Context, expense *domain.Expense) error {
	if expense.CategoryID != nil {
		if *expense.CategoryID == "" {
			return domain.ErrCategoryNotFound
		}
		// Validar que a categoria pertence ao usuario
		_, err := s.categoryService.GetCategory(ctx, *expense.CategoryID, expense.UserID)
		if err != nil {
			return domain.ErrCategoryNotFound
		}
		return nil
	}

	if expense.Category != "" {
		// Resolver nome para category_id
		cat, err := s.categoryService.LookupByName(ctx, expense.Category, expense.UserID)
		if err != nil {
			return domain.ErrCategoryNotFound
		}
		expense.CategoryID = &cat.ID
		expense.Category = cat.Name
		return nil
	}

	// Nenhum dos dois: comportamento legado — sem validacao
	return nil
}
