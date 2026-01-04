package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rodrigo/expense-tracker/internal/domain"
	"github.com/rodrigo/expense-tracker/internal/repository"
)

var (
	ErrInvalidToken = errors.New("invalid or expired token")
)

// AuthService contém a lógica de autenticação
type AuthService struct {
	userRepo    repository.UserRepository
	idGenerator IDGenerator
	jwtSecret   string
	// Token expiration time (default: 24 hours)
	tokenExpiration time.Duration
}

// JWTClaims representa os claims do JWT
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// AuthResponse representa a resposta de autenticação
type AuthResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      *domain.User `json:"user"`
}

// NewAuthService cria uma nova instância do auth service
func NewAuthService(userRepo repository.UserRepository, idGen IDGenerator, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		idGenerator:     idGen,
		jwtSecret:       jwtSecret,
		tokenExpiration: 24 * time.Hour, // 24 horas por padrão
	}
}

// SetTokenExpiration define o tempo de expiração do token
func (s *AuthService) SetTokenExpiration(duration time.Duration) {
	s.tokenExpiration = duration
}

// Register registra um novo usuário
func (s *AuthService) Register(ctx context.Context, registration *domain.UserRegistration) (*AuthResponse, error) {
	// Validar username
	if registration.Username == "" {
		return nil, domain.ErrEmptyUsername
	}
	if len(registration.Username) < 3 {
		return nil, domain.ErrUsernameTooShort
	}
	if len(registration.Username) > 50 {
		return nil, domain.ErrUsernameTooLong
	}

	// Validar email
	if registration.Email == "" {
		return nil, domain.ErrEmptyEmail
	}

	// Validar e hash da senha
	passwordHash, err := domain.HashPassword(registration.Password)
	if err != nil {
		return nil, err
	}

	// Criar usuário
	now := time.Now().UTC()
	user := &domain.User{
		ID:           s.idGenerator.Generate(),
		Username:     registration.Username,
		Email:        registration.Email,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Validar entidade
	if err := user.Validate(); err != nil {
		return nil, err
	}

	// Persistir
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Gerar token
	token, expiresAt, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user.Sanitize(),
	}, nil
}

// Login autentica um usuário e retorna um JWT
func (s *AuthService) Login(ctx context.Context, credentials *domain.UserCredentials) (*AuthResponse, error) {
	// Buscar usuário pelo username
	user, err := s.userRepo.GetByUsername(ctx, credentials.Username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	// Verificar senha
	if err := user.CheckPassword(credentials.Password); err != nil {
		return nil, err
	}

	// Gerar token
	token, expiresAt, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user.Sanitize(),
	}, nil
}

// ValidateToken valida um JWT e retorna os claims
func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verificar método de assinatura
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GetUserFromToken extrai e valida o usuário de um token
func (s *AuthService) GetUserFromToken(ctx context.Context, tokenString string) (*domain.User, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}

	return user, nil
}

// generateToken gera um novo JWT para o usuário
func (s *AuthService) generateToken(user *domain.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.tokenExpiration)

	claims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "expense-tracker",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}
