package domain

import (
	"errors"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User representa um usuário do sistema
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Nunca expor password no JSON
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserCredentials representa credenciais de login
type UserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserRegistration representa dados de registro
type UserRegistration struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Erros de validação de usuário
var (
	ErrEmptyUsername      = errors.New("username cannot be empty")
	ErrUsernameTooShort   = errors.New("username must be at least 3 characters")
	ErrUsernameTooLong    = errors.New("username must be at most 50 characters")
	ErrInvalidUsername    = errors.New("username can only contain letters, numbers, and underscores")
	ErrEmptyEmail         = errors.New("email cannot be empty")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrEmptyPassword      = errors.New("password cannot be empty")
	ErrPasswordTooShort   = errors.New("password must be at least 8 characters")
	ErrInvalidCredentials = errors.New("invalid username or password")
)

var (
	// Regex para validar username (letras, números e underscores)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	// Regex simples para validar email
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

// Validate valida os campos obrigatórios de um usuário
func (u *User) Validate() error {
	// Validar username
	if u.Username == "" {
		return ErrEmptyUsername
	}
	if len(u.Username) < 3 {
		return ErrUsernameTooShort
	}
	if len(u.Username) > 50 {
		return ErrUsernameTooLong
	}
	if !usernameRegex.MatchString(u.Username) {
		return ErrInvalidUsername
	}

	// Validar email
	if u.Email == "" {
		return ErrEmptyEmail
	}
	if !emailRegex.MatchString(u.Email) {
		return ErrInvalidEmail
	}

	return nil
}

// ValidatePassword valida a senha (antes de hash)
func ValidatePassword(password string) error {
	if password == "" {
		return ErrEmptyPassword
	}
	if len(password) < 8 {
		return ErrPasswordTooShort
	}
	return nil
}

// HashPassword gera um hash bcrypt da senha
func HashPassword(password string) (string, error) {
	if err := ValidatePassword(password); err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// CheckPassword verifica se a senha corresponde ao hash
func (u *User) CheckPassword(password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	if err != nil {
		return ErrInvalidCredentials
	}
	return nil
}

// Sanitize remove informações sensíveis antes de retornar ao cliente
func (u *User) Sanitize() *User {
	return &User{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		// PasswordHash is omitted (already excluded in JSON with json:"-")
	}
}
