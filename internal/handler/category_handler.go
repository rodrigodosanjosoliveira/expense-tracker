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

// CategoryHandler gerencia as requisicoes HTTP para categorias
type CategoryHandler struct {
	service service.CategoryServiceInterface
}

// NewCategoryHandler cria uma nova instancia de CategoryHandler
func NewCategoryHandler(svc service.CategoryServiceInterface) *CategoryHandler {
	return &CategoryHandler{service: svc}
}

// CreateCategory handler para POST /categories
func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	cat := &domain.Category{
		UserID: userID,
		Name:   req.Name,
	}

	if err := h.service.CreateCategory(r.Context(), cat); err != nil {
		switch err {
		case domain.ErrEmptyCategoryName, domain.ErrCategoryNameTooLong:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case domain.ErrCategoryAlreadyExists:
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cat)
}

// ListCategories handler para GET /categories
func (h *CategoryHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	categories, err := h.service.ListCategories(r.Context(), userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

// GetCategory handler para GET /categories/{id}
func (h *CategoryHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/categories/")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	cat, err := h.service.GetCategory(r.Context(), id, userID)
	if err != nil {
		if err == repository.ErrNotFound || err == domain.ErrCategoryNotFound {
			http.Error(w, "Category not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cat)
}

// UpdateCategory handler para PUT /categories/{id}
func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/categories/")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	cat := &domain.Category{
		ID:     id,
		UserID: userID,
		Name:   req.Name,
	}

	if err := h.service.UpdateCategory(r.Context(), cat); err != nil {
		switch err {
		case domain.ErrEmptyCategoryName, domain.ErrCategoryNameTooLong:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case domain.ErrCategoryAlreadyExists:
			http.Error(w, err.Error(), http.StatusConflict)
		case repository.ErrNotFound, domain.ErrCategoryNotFound:
			http.Error(w, "Category not found", http.StatusNotFound)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	updated, err := h.service.GetCategory(r.Context(), id, userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// DeleteCategory handler para DELETE /categories/{id}
func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/categories/")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteCategory(r.Context(), id, userID); err != nil {
		switch err {
		case repository.ErrNotFound, domain.ErrCategoryNotFound:
			http.Error(w, "Category not found", http.StatusNotFound)
		case domain.ErrCategoryInUse:
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
