package service

import (
	"context"
	"time"

	"github.com/rodrigo/expense-tracker/internal/domain"
	"github.com/rodrigo/expense-tracker/internal/repository"
)

// WebhookService gerencia webhooks
type WebhookService struct {
	webhookRepo repository.WebhookRepository
	idGenerator IDGenerator
}

// NewWebhookService cria uma nova instância
func NewWebhookService(webhookRepo repository.WebhookRepository, idGen IDGenerator) *WebhookService {
	return &WebhookService{
		webhookRepo: webhookRepo,
		idGenerator: idGen,
	}
}

// CreateWebhook cria um novo webhook
func (s *WebhookService) CreateWebhook(ctx context.Context, webhook *domain.Webhook) error {
	// Validar
	if err := webhook.Validate(); err != nil {
		return err
	}

	// Gerar ID se não fornecido
	if webhook.ID == "" {
		webhook.ID = s.idGenerator.Generate()
	}

	// Setar timestamps
	now := time.Now().UTC()
	webhook.CreatedAt = now
	webhook.UpdatedAt = now

	// Garantir que IsActive seja true por padrão
	if !webhook.IsActive {
		webhook.IsActive = true
	}

	// Persistir
	return s.webhookRepo.Create(ctx, webhook)
}

// GetWebhook retorna um webhook pelo ID
func (s *WebhookService) GetWebhook(ctx context.Context, id string) (*domain.Webhook, error) {
	return s.webhookRepo.GetByID(ctx, id)
}

// GetUserWebhooks retorna todos os webhooks de um usuário
func (s *WebhookService) GetUserWebhooks(ctx context.Context, userID string) ([]*domain.Webhook, error) {
	return s.webhookRepo.GetByUserID(ctx, userID)
}

// UpdateWebhook atualiza um webhook existente
func (s *WebhookService) UpdateWebhook(ctx context.Context, webhook *domain.Webhook) error {
	// Validar
	if err := webhook.Validate(); err != nil {
		return err
	}

	// Verificar se existe
	existing, err := s.webhookRepo.GetByID(ctx, webhook.ID)
	if err != nil {
		return err
	}

	// Preservar CreatedAt e atualizar UpdatedAt
	webhook.CreatedAt = existing.CreatedAt
	webhook.UpdatedAt = time.Now().UTC()

	return s.webhookRepo.Update(ctx, webhook)
}

// DeleteWebhook remove um webhook
func (s *WebhookService) DeleteWebhook(ctx context.Context, id string) error {
	return s.webhookRepo.Delete(ctx, id)
}

// ToggleWebhook ativa/desativa um webhook
func (s *WebhookService) ToggleWebhook(ctx context.Context, id string, isActive bool) error {
	webhook, err := s.webhookRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	webhook.IsActive = isActive
	webhook.UpdatedAt = time.Now().UTC()

	return s.webhookRepo.Update(ctx, webhook)
}
