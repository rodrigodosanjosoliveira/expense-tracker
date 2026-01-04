package repository

import (
	"context"
	"errors"

	"github.com/rodrigo/expense-tracker/internal/domain"
)

// Erros específicos do webhook repository
var (
	ErrWebhookNotFound      = errors.New("webhook not found")
	ErrWebhookAlreadyExists = errors.New("webhook already exists")
)

// WebhookRepository define as operações de persistência para webhooks
type WebhookRepository interface {
	Create(ctx context.Context, webhook *domain.Webhook) error
	GetByID(ctx context.Context, id string) (*domain.Webhook, error)
	GetByUserID(ctx context.Context, userID string) ([]*domain.Webhook, error)
	GetActiveByUserIDAndEvent(ctx context.Context, userID string, event domain.WebhookEvent) ([]*domain.Webhook, error)
	Update(ctx context.Context, webhook *domain.Webhook) error
	Delete(ctx context.Context, id string) error
	UpdateLastTrigger(ctx context.Context, id string) error
}
