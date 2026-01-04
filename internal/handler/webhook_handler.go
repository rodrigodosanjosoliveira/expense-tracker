package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/rodrigo/expense-tracker/internal/domain"
	"github.com/rodrigo/expense-tracker/internal/middleware"
	"github.com/rodrigo/expense-tracker/internal/repository"
	"github.com/rodrigo/expense-tracker/internal/service"
)

// WebhookHandler gerencia as requisições HTTP de webhooks
type WebhookHandler struct {
	service *service.WebhookService
}

// NewWebhookHandler cria uma nova instância
func NewWebhookHandler(service *service.WebhookService) *WebhookHandler {
	return &WebhookHandler{
		service: service,
	}
}

// CreateWebhook handler para POST /webhooks
// @Summary Criar novo webhook
// @Description Cria um novo webhook para receber notificações
// @Tags Webhooks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param webhook body domain.Webhook true "Dados do webhook"
// @Success 201 {object} domain.Webhook
// @Failure 400 {string} string "Dados inválidos"
// @Failure 401 {string} string "Não autorizado"
// @Failure 500 {string} string "Erro interno"
// @Router /webhooks [post]
func (h *WebhookHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extrair user_id do context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var webhook domain.Webhook
	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Associar webhook ao usuário autenticado
	webhook.UserID = userID

	if err := h.service.CreateWebhook(r.Context(), &webhook); err != nil {
		switch err {
		case domain.ErrEmptyWebhookName, domain.ErrInvalidWebhookType,
			domain.ErrEmptyWebhookURL, domain.ErrNoEventsSelected, domain.ErrInvalidWebhookURL:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(webhook)
}

// ListWebhooks handler para GET /webhooks
// @Summary Listar webhooks
// @Description Retorna todos os webhooks do usuário autenticado
// @Tags Webhooks
// @Produce json
// @Security BearerAuth
// @Success 200 {array} domain.Webhook
// @Failure 401 {string} string "Não autorizado"
// @Failure 500 {string} string "Erro interno"
// @Router /webhooks [get]
func (h *WebhookHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extrair user_id do context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	webhooks, err := h.service.GetUserWebhooks(r.Context(), userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webhooks)
}

// GetWebhook handler para GET /webhooks/{id}
// @Summary Buscar webhook por ID
// @Description Retorna um webhook específico pelo ID
// @Tags Webhooks
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do webhook"
// @Success 200 {object} domain.Webhook
// @Failure 401 {string} string "Não autorizado"
// @Failure 403 {string} string "Sem permissão"
// @Failure 404 {string} string "Webhook não encontrado"
// @Failure 500 {string} string "Erro interno"
// @Router /webhooks/{id} [get]
func (h *WebhookHandler) GetWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extrair user_id do context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extrair ID do path
	id := strings.TrimPrefix(r.URL.Path, "/webhooks/")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	webhook, err := h.service.GetWebhook(r.Context(), id)
	if err != nil {
		if err == repository.ErrWebhookNotFound {
			http.Error(w, "Webhook not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Verificar se o webhook pertence ao usuário autenticado
	if webhook.UserID != userID {
		http.Error(w, "Forbidden: this webhook belongs to another user", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webhook)
}

// UpdateWebhook handler para PUT /webhooks/{id}
// @Summary Atualizar webhook
// @Description Atualiza um webhook existente
// @Tags Webhooks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do webhook"
// @Param webhook body domain.Webhook true "Dados atualizados do webhook"
// @Success 200 {object} domain.Webhook
// @Failure 400 {string} string "Dados inválidos"
// @Failure 401 {string} string "Não autorizado"
// @Failure 403 {string} string "Sem permissão"
// @Failure 404 {string} string "Webhook não encontrado"
// @Failure 500 {string} string "Erro interno"
// @Router /webhooks/{id} [put]
func (h *WebhookHandler) UpdateWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extrair user_id do context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/webhooks/")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	// Verificar se o webhook existe e pertence ao usuário
	existing, err := h.service.GetWebhook(r.Context(), id)
	if err != nil {
		if err == repository.ErrWebhookNotFound {
			http.Error(w, "Webhook not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	if existing.UserID != userID {
		http.Error(w, "Forbidden: this webhook belongs to another user", http.StatusForbidden)
		return
	}

	var webhook domain.Webhook
	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	webhook.ID = id
	webhook.UserID = userID

	if err := h.service.UpdateWebhook(r.Context(), &webhook); err != nil {
		switch err {
		case repository.ErrWebhookNotFound:
			http.Error(w, "Webhook not found", http.StatusNotFound)
		case domain.ErrEmptyWebhookName, domain.ErrInvalidWebhookType,
			domain.ErrEmptyWebhookURL, domain.ErrNoEventsSelected, domain.ErrInvalidWebhookURL:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webhook)
}

// DeleteWebhook handler para DELETE /webhooks/{id}
// @Summary Deletar webhook
// @Description Remove um webhook pelo ID
// @Tags Webhooks
// @Security BearerAuth
// @Param id path string true "ID do webhook"
// @Success 204 "Webhook deletado com sucesso"
// @Failure 401 {string} string "Não autorizado"
// @Failure 403 {string} string "Sem permissão"
// @Failure 404 {string} string "Webhook não encontrado"
// @Failure 500 {string} string "Erro interno"
// @Router /webhooks/{id} [delete]
func (h *WebhookHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extrair user_id do context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/webhooks/")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	// Verificar se o webhook existe e pertence ao usuário
	existing, err := h.service.GetWebhook(r.Context(), id)
	if err != nil {
		if err == repository.ErrWebhookNotFound {
			http.Error(w, "Webhook not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	if existing.UserID != userID {
		http.Error(w, "Forbidden: this webhook belongs to another user", http.StatusForbidden)
		return
	}

	if err := h.service.DeleteWebhook(r.Context(), id); err != nil {
		if err == repository.ErrWebhookNotFound {
			http.Error(w, "Webhook not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
