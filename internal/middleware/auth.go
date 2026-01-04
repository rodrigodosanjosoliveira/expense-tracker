package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/rodrigo/expense-tracker/internal/service"
)

// contextKey é um tipo customizado para chaves do context
type contextKey string

const (
	// UserIDKey é a chave para armazenar o user ID no context
	UserIDKey contextKey = "user_id"
	// UsernameKey é a chave para armazenar o username no context
	UsernameKey contextKey = "username"
)

// AuthMiddleware é um middleware para autenticação JWT
type AuthMiddleware struct {
	authService *service.AuthService
}

// NewAuthMiddleware cria uma nova instância do middleware
func NewAuthMiddleware(authService *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// Authenticate é o middleware que valida o JWT
func (m *AuthMiddleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extrair token do header Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Verificar formato "Bearer {token}"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format. Expected: Bearer {token}", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Validar token
		claims, err := m.authService.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Adicionar user_id e username ao context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UsernameKey, claims.Username)

		// Chamar próximo handler com context atualizado
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// GetUserIDFromContext extrai o user ID do context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// GetUsernameFromContext extrai o username do context
func GetUsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(UsernameKey).(string)
	return username, ok
}
