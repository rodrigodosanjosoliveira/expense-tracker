package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rodrigo/expense-tracker/internal/domain"
)

// PostgresWebhookRepository implementa WebhookRepository usando PostgreSQL
type PostgresWebhookRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresWebhookRepository cria uma nova instância
func NewPostgresWebhookRepository(pool *pgxpool.Pool) *PostgresWebhookRepository {
	return &PostgresWebhookRepository{
		pool: pool,
	}
}

// Create adiciona um novo webhook no banco
func (r *PostgresWebhookRepository) Create(ctx context.Context, webhook *domain.Webhook) error {
	// Converter events para array de strings
	events := make([]string, len(webhook.Events))
	for i, e := range webhook.Events {
		events[i] = string(e)
	}

	query := `
		INSERT INTO webhooks (id, user_id, name, type, url, events, is_active, secret, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		webhook.ID,
		webhook.UserID,
		webhook.Name,
		webhook.Type,
		webhook.URL,
		events, // pgx suporta arrays nativamente
		webhook.IsActive,
		webhook.Secret,
		webhook.CreatedAt,
		webhook.UpdatedAt,
	)

	return err
}

// GetByID retorna um webhook pelo ID
func (r *PostgresWebhookRepository) GetByID(ctx context.Context, id string) (*domain.Webhook, error) {
	query := `
		SELECT id, user_id, name, type, url, events, is_active, secret, created_at, updated_at, last_trigger
		FROM webhooks
		WHERE id = $1
	`

	var webhook domain.Webhook
	var events []string
	var secret *string
	var lastTrigger *time.Time

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&webhook.ID,
		&webhook.UserID,
		&webhook.Name,
		&webhook.Type,
		&webhook.URL,
		&events, // pgx suporta arrays nativamente
		&webhook.IsActive,
		&secret,
		&webhook.CreatedAt,
		&webhook.UpdatedAt,
		&lastTrigger,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrWebhookNotFound
		}
		return nil, err
	}

	// Converter events de []string para []WebhookEvent
	webhook.Events = make([]domain.WebhookEvent, len(events))
	for i, e := range events {
		webhook.Events[i] = domain.WebhookEvent(e)
	}

	if secret != nil {
		webhook.Secret = *secret
	}

	webhook.LastTrigger = lastTrigger

	return &webhook, nil
}

// GetByUserID retorna todos os webhooks de um usuário
func (r *PostgresWebhookRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Webhook, error) {
	query := `
		SELECT id, user_id, name, type, url, events, is_active, secret, created_at, updated_at, last_trigger
		FROM webhooks
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []*domain.Webhook
	for rows.Next() {
		var webhook domain.Webhook
		var events []string
		var secret *string
		var lastTrigger *time.Time

		err := rows.Scan(
			&webhook.ID,
			&webhook.UserID,
			&webhook.Name,
			&webhook.Type,
			&webhook.URL,
			&events, // pgx suporta arrays nativamente
			&webhook.IsActive,
			&secret,
			&webhook.CreatedAt,
			&webhook.UpdatedAt,
			&lastTrigger,
		)
		if err != nil {
			return nil, err
		}

		// Converter events
		webhook.Events = make([]domain.WebhookEvent, len(events))
		for i, e := range events {
			webhook.Events[i] = domain.WebhookEvent(e)
		}

		if secret != nil {
			webhook.Secret = *secret
		}

		webhook.LastTrigger = lastTrigger

		webhooks = append(webhooks, &webhook)
	}

	return webhooks, rows.Err()
}

// GetActiveByUserIDAndEvent retorna webhooks ativos de um usuário para um evento específico
func (r *PostgresWebhookRepository) GetActiveByUserIDAndEvent(ctx context.Context, userID string, event domain.WebhookEvent) ([]*domain.Webhook, error) {
	query := `
		SELECT id, user_id, name, type, url, events, is_active, secret, created_at, updated_at, last_trigger
		FROM webhooks
		WHERE user_id = $1 AND is_active = true AND $2 = ANY(events)
	`

	rows, err := r.pool.Query(ctx, query, userID, string(event))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []*domain.Webhook
	for rows.Next() {
		var webhook domain.Webhook
		var events []string
		var secret *string
		var lastTrigger *time.Time

		err := rows.Scan(
			&webhook.ID,
			&webhook.UserID,
			&webhook.Name,
			&webhook.Type,
			&webhook.URL,
			&events, // pgx suporta arrays nativamente
			&webhook.IsActive,
			&secret,
			&webhook.CreatedAt,
			&webhook.UpdatedAt,
			&lastTrigger,
		)
		if err != nil {
			return nil, err
		}

		// Converter events
		webhook.Events = make([]domain.WebhookEvent, len(events))
		for i, e := range events {
			webhook.Events[i] = domain.WebhookEvent(e)
		}

		if secret != nil {
			webhook.Secret = *secret
		}

		webhook.LastTrigger = lastTrigger

		webhooks = append(webhooks, &webhook)
	}

	return webhooks, rows.Err()
}

// Update atualiza um webhook existente
func (r *PostgresWebhookRepository) Update(ctx context.Context, webhook *domain.Webhook) error {
	// Converter events para array de strings
	events := make([]string, len(webhook.Events))
	for i, e := range webhook.Events {
		events[i] = string(e)
	}

	query := `
		UPDATE webhooks
		SET name = $2, type = $3, url = $4, events = $5, is_active = $6, secret = $7, updated_at = $8
		WHERE id = $1
	`

	result, err := r.pool.Exec(
		ctx,
		query,
		webhook.ID,
		webhook.Name,
		webhook.Type,
		webhook.URL,
		events, // pgx suporta arrays nativamente
		webhook.IsActive,
		webhook.Secret,
		webhook.UpdatedAt,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrWebhookNotFound
	}

	return nil
}

// Delete remove um webhook do banco
func (r *PostgresWebhookRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM webhooks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrWebhookNotFound
	}

	return nil
}

// UpdateLastTrigger atualiza o timestamp do último disparo
func (r *PostgresWebhookRepository) UpdateLastTrigger(ctx context.Context, id string) error {
	query := `
		UPDATE webhooks
		SET last_trigger = $2
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id, time.Now().UTC())
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrWebhookNotFound
	}

	return nil
}
