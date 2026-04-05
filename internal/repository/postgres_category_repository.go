package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rodrigo/expense-tracker/internal/domain"
)

// PostgresCategoryRepository implementa CategoryRepository usando PostgreSQL
type PostgresCategoryRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresCategoryRepository cria uma nova instancia
func NewPostgresCategoryRepository(pool *pgxpool.Pool) *PostgresCategoryRepository {
	return &PostgresCategoryRepository{pool: pool}
}

// Create persiste uma nova categoria
// Detecta violacao de unique constraint (23505) e retorna domain.ErrCategoryAlreadyExists
func (r *PostgresCategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	query := `
		INSERT INTO categories (id, user_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.pool.Exec(ctx, query,
		category.ID,
		category.UserID,
		category.Name,
		category.CreatedAt,
		category.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrCategoryAlreadyExists
		}
		return err
	}

	return nil
}

// GetByID retorna uma categoria pelo ID scoped ao userID
// Retorna ErrNotFound quando nao existe para o usuario
func (r *PostgresCategoryRepository) GetByID(ctx context.Context, id string, userID string) (*domain.Category, error) {
	query := `
		SELECT id, user_id, name, created_at, updated_at
		FROM categories
		WHERE id = $1 AND user_id = $2
	`

	var cat domain.Category
	err := r.pool.QueryRow(ctx, query, id, userID).Scan(
		&cat.ID,
		&cat.UserID,
		&cat.Name,
		&cat.CreatedAt,
		&cat.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &cat, nil
}

// GetByName retorna uma categoria pelo nome (case-insensitive) scoped ao userID
// Retorna ErrNotFound quando nao existe para o usuario
func (r *PostgresCategoryRepository) GetByName(ctx context.Context, name string, userID string) (*domain.Category, error) {
	query := `
		SELECT id, user_id, name, created_at, updated_at
		FROM categories
		WHERE LOWER(name) = LOWER($1) AND user_id = $2
	`

	var cat domain.Category
	err := r.pool.QueryRow(ctx, query, name, userID).Scan(
		&cat.ID,
		&cat.UserID,
		&cat.Name,
		&cat.CreatedAt,
		&cat.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &cat, nil
}

// GetAll retorna todas as categorias do userID
func (r *PostgresCategoryRepository) GetAll(ctx context.Context, userID string) ([]*domain.Category, error) {
	query := `
		SELECT id, user_id, name, created_at, updated_at
		FROM categories
		WHERE user_id = $1
		ORDER BY name ASC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*domain.Category
	for rows.Next() {
		var cat domain.Category
		if err := rows.Scan(&cat.ID, &cat.UserID, &cat.Name, &cat.CreatedAt, &cat.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, &cat)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

// Update persiste alteracoes em uma categoria existente
// Retorna ErrNotFound quando nao existe para o usuario
// Retorna domain.ErrCategoryAlreadyExists em colisao de nome
func (r *PostgresCategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	query := `
		UPDATE categories
		SET name = $1, updated_at = $2
		WHERE id = $3 AND user_id = $4
	`

	result, err := r.pool.Exec(ctx, query,
		category.Name,
		category.UpdatedAt,
		category.ID,
		category.UserID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrCategoryAlreadyExists
		}
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// Delete remove uma categoria pelo ID scoped ao userID
// Retorna ErrNotFound quando nao existe para o usuario
// Retorna domain.ErrCategoryInUse em violacao de FK (expenses referenciando a categoria)
func (r *PostgresCategoryRepository) Delete(ctx context.Context, id string, userID string) error {
	// Verificar existencia e ownership antes de deletar
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1 AND user_id = $2)`
	if err := r.pool.QueryRow(ctx, checkQuery, id, userID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}

	query := `DELETE FROM categories WHERE id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, query, id, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return domain.ErrCategoryInUse
		}
		return err
	}

	return nil
}
