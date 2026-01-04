package domain

import (
	"errors"
	"time"
)

// WebhookType representa o tipo de webhook
type WebhookType string

const (
	WebhookTypeSlack   WebhookType = "slack"
	WebhookTypeDiscord WebhookType = "discord"
	WebhookTypeCustom  WebhookType = "custom"
)

// WebhookEvent representa os eventos que disparam webhooks
type WebhookEvent string

const (
	WebhookEventExpenseCreated      WebhookEvent = "expense.created"
	WebhookEventExpenseUpdated      WebhookEvent = "expense.updated"
	WebhookEventExpenseDeleted      WebhookEvent = "expense.deleted"
	WebhookEventDailyLimitReached   WebhookEvent = "daily_limit.reached"
	WebhookEventWeeklyLimitReached  WebhookEvent = "weekly_limit.reached"
	WebhookEventMonthlyLimitReached WebhookEvent = "monthly_limit.reached"
	WebhookEventHighExpense         WebhookEvent = "high_expense.detected"
)

// Webhook representa uma configuração de webhook
type Webhook struct {
	ID          string         `json:"id"`
	UserID      string         `json:"user_id"`
	Name        string         `json:"name"`
	Type        WebhookType    `json:"type"`
	URL         string         `json:"url"`
	Events      []WebhookEvent `json:"events"`
	IsActive    bool           `json:"is_active"`
	Secret      string         `json:"secret,omitempty"` // Para validação de webhooks custom
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	LastTrigger *time.Time     `json:"last_trigger,omitempty"`
}

// WebhookPayload representa o payload enviado para webhooks
type WebhookPayload struct {
	Event     WebhookEvent `json:"event"`
	Timestamp time.Time    `json:"timestamp"`
	UserID    string       `json:"user_id"`
	Data      interface{}  `json:"data"`
}

// ExpenseEventData representa dados de eventos relacionados a despesas
type ExpenseEventData struct {
	Expense *Expense `json:"expense"`
	Action  string   `json:"action"` // created, updated, deleted
}

// LimitReachedData representa dados de eventos de limite atingido
type LimitReachedData struct {
	Period       string  `json:"period"` // daily, weekly, monthly
	Limit        float64 `json:"limit"`
	CurrentTotal float64 `json:"current_total"`
	Percentage   float64 `json:"percentage"`
}

// HighExpenseData representa dados de evento de despesa alta
type HighExpenseData struct {
	Expense       *Expense `json:"expense"`
	Threshold     float64  `json:"threshold"`
	AverageAmount float64  `json:"average_amount"`
}

// Erros de validação de webhook
var (
	ErrEmptyWebhookName   = errors.New("webhook name cannot be empty")
	ErrInvalidWebhookType = errors.New("invalid webhook type")
	ErrEmptyWebhookURL    = errors.New("webhook URL cannot be empty")
	ErrNoEventsSelected   = errors.New("at least one event must be selected")
	ErrInvalidWebhookURL  = errors.New("invalid webhook URL format")
)

// Validate valida os campos obrigatórios de um webhook
func (w *Webhook) Validate() error {
	if w.Name == "" {
		return ErrEmptyWebhookName
	}

	if w.Type != WebhookTypeSlack && w.Type != WebhookTypeDiscord && w.Type != WebhookTypeCustom {
		return ErrInvalidWebhookType
	}

	if w.URL == "" {
		return ErrEmptyWebhookURL
	}

	// Validação básica de URL
	if len(w.URL) < 10 || (w.URL[:4] != "http" && w.URL[:5] != "https") {
		return ErrInvalidWebhookURL
	}

	if len(w.Events) == 0 {
		return ErrNoEventsSelected
	}

	return nil
}

// IsEventEnabled verifica se um evento específico está habilitado
func (w *Webhook) IsEventEnabled(event WebhookEvent) bool {
	if !w.IsActive {
		return false
	}

	for _, e := range w.Events {
		if e == event {
			return true
		}
	}
	return false
}

// UpdateLastTrigger atualiza o timestamp do último disparo
func (w *Webhook) UpdateLastTrigger() {
	now := time.Now().UTC()
	w.LastTrigger = &now
}
