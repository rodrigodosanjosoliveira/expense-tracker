package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rodrigo/expense-tracker/internal/config"
)

// NewPostgresPool cria um novo pool de conexões PostgreSQL
// Conceito: Connection pooling para melhor performance e gerenciamento de recursos
func NewPostgresPool(cfg *config.DatabaseConfig) (*pgxpool.Pool, error) {
	// Configurar pool de conexões
	poolConfig, err := pgxpool.ParseConfig(cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Configurações do pool
	// Conceito: Ajustar de acordo com a carga esperada
	poolConfig.MaxConns = 25                   // Máximo de conexões
	poolConfig.MinConns = 5                    // Mínimo de conexões mantidas
	poolConfig.MaxConnLifetime = time.Hour     // Tempo máximo de vida de uma conexão
	poolConfig.MaxConnIdleTime = 30 * time.Minute // Tempo máximo que uma conexão pode ficar idle
	poolConfig.HealthCheckPeriod = time.Minute    // Intervalo de health check

	// Criar pool com timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Testar conexão
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
