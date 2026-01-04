package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/rodrigo/expense-tracker/internal/domain"
	"github.com/rodrigo/expense-tracker/internal/repository"
)

// NotificationService gerencia o envio de notificações
type NotificationService struct {
	webhookRepo repository.WebhookRepository
	httpClient  *http.Client
}

// NewNotificationService cria uma nova instância
func NewNotificationService(webhookRepo repository.WebhookRepository) *NotificationService {
	return &NotificationService{
		webhookRepo: webhookRepo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// TriggerEvent dispara webhooks para um evento específico
func (s *NotificationService) TriggerEvent(ctx context.Context, userID string, event domain.WebhookEvent, data interface{}) error {
	// Buscar webhooks ativos para este usuário e evento
	webhooks, err := s.webhookRepo.GetActiveByUserIDAndEvent(ctx, userID, event)
	if err != nil {
		return fmt.Errorf("failed to get webhooks: %w", err)
	}

	if len(webhooks) == 0 {
		log.Printf("No active webhooks found for user %s and event %s", userID, event)
		return nil
	}

	// Criar payload
	payload := domain.WebhookPayload{
		Event:     event,
		Timestamp: time.Now().UTC(),
		UserID:    userID,
		Data:      data,
	}

	// Enviar para cada webhook (em goroutines para não bloquear)
	for _, webhook := range webhooks {
		go s.sendWebhook(ctx, webhook, payload)
	}

	return nil
}

// sendWebhook envia uma notificação para um webhook específico
func (s *NotificationService) sendWebhook(ctx context.Context, webhook *domain.Webhook, payload domain.WebhookPayload) {
	var err error

	switch webhook.Type {
	case domain.WebhookTypeSlack:
		err = s.sendToSlack(webhook.URL, payload)
	case domain.WebhookTypeDiscord:
		err = s.sendToDiscord(webhook.URL, payload)
	case domain.WebhookTypeCustom:
		err = s.sendToCustomWebhook(webhook.URL, payload)
	default:
		log.Printf("Unknown webhook type: %s", webhook.Type)
		return
	}

	if err != nil {
		log.Printf("Failed to send webhook %s (%s): %v", webhook.Name, webhook.ID, err)
		return
	}

	// Atualizar timestamp do último disparo
	if err := s.webhookRepo.UpdateLastTrigger(ctx, webhook.ID); err != nil {
		log.Printf("Failed to update last trigger for webhook %s: %v", webhook.ID, err)
	}

	log.Printf("Webhook %s (%s) triggered successfully", webhook.Name, webhook.ID)
}

// sendToSlack envia notificação para Slack
func (s *NotificationService) sendToSlack(webhookURL string, payload domain.WebhookPayload) error {
	// Formatar mensagem para Slack
	message := s.formatSlackMessage(payload)

	slackPayload := map[string]interface{}{
		"text":       message.Text,
		"username":   "Expense Tracker",
		"icon_emoji": ":moneybag:",
	}

	if len(message.Attachments) > 0 {
		slackPayload["attachments"] = message.Attachments
	}

	return s.sendJSONPost(webhookURL, slackPayload)
}

// sendToDiscord envia notificação para Discord
func (s *NotificationService) sendToDiscord(webhookURL string, payload domain.WebhookPayload) error {
	// Formatar mensagem para Discord
	message := s.formatDiscordMessage(payload)

	discordPayload := map[string]interface{}{
		"content":  message.Content,
		"username": "Expense Tracker",
	}

	if len(message.Embeds) > 0 {
		discordPayload["embeds"] = message.Embeds
	}

	return s.sendJSONPost(webhookURL, discordPayload)
}

// sendToCustomWebhook envia notificação para webhook customizado
func (s *NotificationService) sendToCustomWebhook(webhookURL string, payload domain.WebhookPayload) error {
	return s.sendJSONPost(webhookURL, payload)
}

// sendJSONPost envia uma requisição POST com JSON
func (s *NotificationService) sendJSONPost(url string, payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// SlackMessage representa uma mensagem formatada para Slack
type SlackMessage struct {
	Text        string                   `json:"text"`
	Attachments []map[string]interface{} `json:"attachments,omitempty"`
}

// DiscordMessage representa uma mensagem formatada para Discord
type DiscordMessage struct {
	Content string                   `json:"content"`
	Embeds  []map[string]interface{} `json:"embeds,omitempty"`
}

// formatSlackMessage formata o payload para Slack
func (s *NotificationService) formatSlackMessage(payload domain.WebhookPayload) SlackMessage {
	var message SlackMessage

	switch payload.Event {
	case domain.WebhookEventExpenseCreated:
		if data, ok := payload.Data.(domain.ExpenseEventData); ok {
			message.Text = ":money_with_wings: Nova despesa criada"
			message.Attachments = []map[string]interface{}{
				{
					"color": "#36a64f",
					"fields": []map[string]interface{}{
						{"title": "Descrição", "value": data.Expense.Description, "short": true},
						{"title": "Valor", "value": fmt.Sprintf("R$ %.2f", data.Expense.Amount), "short": true},
						{"title": "Categoria", "value": data.Expense.Category, "short": true},
						{"title": "Data", "value": data.Expense.Date.Format("02/01/2006"), "short": true},
					},
				},
			}
		}
	case domain.WebhookEventHighExpense:
		if data, ok := payload.Data.(domain.HighExpenseData); ok {
			message.Text = ":warning: Despesa alta detectada!"
			message.Attachments = []map[string]interface{}{
				{
					"color": "#ff9900",
					"fields": []map[string]interface{}{
						{"title": "Descrição", "value": data.Expense.Description, "short": true},
						{"title": "Valor", "value": fmt.Sprintf("R$ %.2f", data.Expense.Amount), "short": true},
						{"title": "Limite", "value": fmt.Sprintf("R$ %.2f", data.Threshold), "short": true},
						{"title": "Média", "value": fmt.Sprintf("R$ %.2f", data.AverageAmount), "short": true},
					},
				},
			}
		}
	case domain.WebhookEventDailyLimitReached, domain.WebhookEventWeeklyLimitReached, domain.WebhookEventMonthlyLimitReached:
		if data, ok := payload.Data.(domain.LimitReachedData); ok {
			message.Text = fmt.Sprintf(":rotating_light: Limite %s atingido!", data.Period)
			message.Attachments = []map[string]interface{}{
				{
					"color": "#ff0000",
					"fields": []map[string]interface{}{
						{"title": "Limite", "value": fmt.Sprintf("R$ %.2f", data.Limit), "short": true},
						{"title": "Total Atual", "value": fmt.Sprintf("R$ %.2f", data.CurrentTotal), "short": true},
						{"title": "Percentual", "value": fmt.Sprintf("%.1f%%", data.Percentage), "short": true},
					},
				},
			}
		}
	default:
		message.Text = fmt.Sprintf("Evento: %s", payload.Event)
	}

	return message
}

// formatDiscordMessage formata o payload para Discord
func (s *NotificationService) formatDiscordMessage(payload domain.WebhookPayload) DiscordMessage {
	var message DiscordMessage

	switch payload.Event {
	case domain.WebhookEventExpenseCreated:
		if data, ok := payload.Data.(domain.ExpenseEventData); ok {
			message.Content = "💸 Nova despesa criada"
			message.Embeds = []map[string]interface{}{
				{
					"title": data.Expense.Description,
					"color": 3066993, // Verde
					"fields": []map[string]interface{}{
						{"name": "Valor", "value": fmt.Sprintf("R$ %.2f", data.Expense.Amount), "inline": true},
						{"name": "Categoria", "value": data.Expense.Category, "inline": true},
						{"name": "Data", "value": data.Expense.Date.Format("02/01/2006"), "inline": true},
					},
					"timestamp": data.Expense.CreatedAt.Format(time.RFC3339),
				},
			}
		}
	case domain.WebhookEventHighExpense:
		if data, ok := payload.Data.(domain.HighExpenseData); ok {
			message.Content = "⚠️ Despesa alta detectada!"
			message.Embeds = []map[string]interface{}{
				{
					"title":       data.Expense.Description,
					"description": fmt.Sprintf("Valor acima do limite de R$ %.2f", data.Threshold),
					"color":       16753920, // Laranja
					"fields": []map[string]interface{}{
						{"name": "Valor", "value": fmt.Sprintf("R$ %.2f", data.Expense.Amount), "inline": true},
						{"name": "Média", "value": fmt.Sprintf("R$ %.2f", data.AverageAmount), "inline": true},
					},
				},
			}
		}
	case domain.WebhookEventDailyLimitReached, domain.WebhookEventWeeklyLimitReached, domain.WebhookEventMonthlyLimitReached:
		if data, ok := payload.Data.(domain.LimitReachedData); ok {
			message.Content = fmt.Sprintf("🚨 Limite %s atingido!", data.Period)
			message.Embeds = []map[string]interface{}{
				{
					"color": 15158332, // Vermelho
					"fields": []map[string]interface{}{
						{"name": "Limite", "value": fmt.Sprintf("R$ %.2f", data.Limit), "inline": true},
						{"name": "Total Atual", "value": fmt.Sprintf("R$ %.2f", data.CurrentTotal), "inline": true},
						{"name": "Percentual", "value": fmt.Sprintf("%.1f%%", data.Percentage), "inline": true},
					},
				},
			}
		}
	default:
		message.Content = fmt.Sprintf("Evento: %s", payload.Event)
	}

	return message
}
