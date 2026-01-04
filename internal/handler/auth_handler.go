package handler

import (
	"encoding/json"
	"net/http"

	"github.com/rodrigo/expense-tracker/internal/domain"
	"github.com/rodrigo/expense-tracker/internal/repository"
	"github.com/rodrigo/expense-tracker/internal/service"
)

// AuthHandler gerencia as requisições de autenticação
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler cria uma nova instância
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handler para POST /auth/register
// @Summary Registrar novo usuário
// @Description Cria uma nova conta de usuário
// @Tags Authentication
// @Accept json
// @Produce json
// @Param registration body domain.UserRegistration true "Dados de registro"
// @Success 201 {object} service.AuthResponse
// @Failure 400 {string} string "Dados inválidos"
// @Failure 409 {string} string "Usuário já existe"
// @Failure 500 {string} string "Erro interno"
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var registration domain.UserRegistration
	if err := json.NewDecoder(r.Body).Decode(&registration); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	response, err := h.authService.Register(r.Context(), &registration)
	if err != nil {
		switch err {
		case domain.ErrEmptyUsername, domain.ErrUsernameTooShort, domain.ErrUsernameTooLong,
			domain.ErrInvalidUsername, domain.ErrEmptyEmail, domain.ErrInvalidEmail,
			domain.ErrEmptyPassword, domain.ErrPasswordTooShort:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case repository.ErrDuplicateUsername:
			http.Error(w, "Username already taken", http.StatusConflict)
		case repository.ErrDuplicateEmail:
			http.Error(w, "Email already registered", http.StatusConflict)
		case repository.ErrUserAlreadyExists:
			http.Error(w, "User already exists", http.StatusConflict)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Login handler para POST /auth/login
// @Summary Login de usuário
// @Description Autentica um usuário e retorna um JWT
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body domain.UserCredentials true "Credenciais de login"
// @Success 200 {object} service.AuthResponse
// @Failure 400 {string} string "Dados inválidos"
// @Failure 401 {string} string "Credenciais inválidas"
// @Failure 500 {string} string "Erro interno"
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var credentials domain.UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	response, err := h.authService.Login(r.Context(), &credentials)
	if err != nil {
		switch err {
		case domain.ErrInvalidCredentials:
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
