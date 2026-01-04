package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config armazena todas as configurações da aplicação
type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	JWTSecret string
	Env       string
}

// ServerConfig configurações do servidor HTTP
type ServerConfig struct {
	Port string
	Host string
}

// DatabaseConfig configurações do banco de dados
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

// Load carrega as configurações a partir de variáveis de ambiente
// Conceito: Twelve-Factor App - Configuração via variáveis de ambiente
func Load() (*Config, error) {
	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("error parsing DB_PORT: %w", err)
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", ""),
			Port:     dbPort,
			User:     getEnv("DB_USER", ""),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", ""),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWTSecret: getEnv("JWT_SECRET", "your-secret-key-change-me"),
		Env:       getEnv("ENV", "development"),
	}, nil
}

// ConnectionString retorna a string de conexão PostgreSQL
func (c *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.Name,
		c.SSLMode,
	)
}

// getEnv retorna o valor de uma variável de ambiente ou um valor padrão
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) Validate() error {
	if c.Database.Host != "" {
		if c.Database.User == "" {
			return fmt.Errorf("database user is required")
		}
		if c.Database.Name == "" {
			return fmt.Errorf("database name is required")
		}
	}

	if c.JWTSecret == "" || c.JWTSecret == "your-secret-key-change-me" {
		return fmt.Errorf("JWT_SECRET is required and must be changed from default")
	}

	return nil
}

func (c *Config) UsePostgreSQL() bool {
	return c.Database.Host != ""
}
