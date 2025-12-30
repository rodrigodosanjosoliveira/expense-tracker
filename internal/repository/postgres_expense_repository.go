package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rodrigo/expense-tracker/internal/domain"
)

// PostgresExpenseRepository implementa ExpenseRepository usando PostgreSQL
// Conceito: pgxpool.Pool - Connection pool para melhor performance
type PostgresExpenseRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresExpenseRepository cria uma nova instância
func NewPostgresExpenseRepository(pool *pgxpool.Pool) *PostgresExpenseRepository {
	return &PostgresExpenseRepository{
		pool: pool,
	}
}

// Create adiciona uma nova despesa no banco
func (r *PostgresExpenseRepository) Create(ctx context.Context, expense *domain.Expense) error {
	query := `
		INSERT INTO expenses (id, description, amount, category, date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		expense.ID,
		expense.Description,
		expense.Amount,
		expense.Category,
		expense.Date,
		expense.CreatedAt,
		expense.UpdatedAt,
	)

	if err != nil {
		// Conceito: Verificar violação de unique constraint (ID duplicado)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAlreadyExists
		}
		return err
	}

	return nil
}

// GetByID retorna uma despesa pelo ID
func (r *PostgresExpenseRepository) GetByID(ctx context.Context, id string) (*domain.Expense, error) {
	query := `
		SELECT id, description, amount, category, date, created_at, updated_at
		FROM expenses
		WHERE id = $1
	`

	var expense domain.Expense
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&expense.ID,
		&expense.Description,
		&expense.Amount,
		&expense.Category,
		&expense.Date,
		&expense.CreatedAt,
		&expense.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &expense, nil
}

// GetAll retorna todas as despesas
func (r *PostgresExpenseRepository) GetAll(ctx context.Context) ([]*domain.Expense, error) {
	query := `
		SELECT id, description, amount, category, date, created_at, updated_at
		FROM expenses
		ORDER BY date DESC, created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []*domain.Expense
	for rows.Next() {
		var expense domain.Expense
		err := rows.Scan(
			&expense.ID,
			&expense.Description,
			&expense.Amount,
			&expense.Category,
			&expense.Date,
			&expense.CreatedAt,
			&expense.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, &expense)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return expenses, nil
}

// Update atualiza uma despesa existente
func (r *PostgresExpenseRepository) Update(ctx context.Context, expense *domain.Expense) error {
	query := `
		UPDATE expenses
		SET description = $2, amount = $3, category = $4, date = $5, updated_at = $6
		WHERE id = $1
	`

	// Conceito: CommandTag retorna informações sobre a query executada
	result, err := r.pool.Exec(
		ctx,
		query,
		expense.ID,
		expense.Description,
		expense.Amount,
		expense.Category,
		expense.Date,
		expense.UpdatedAt,
	)

	if err != nil {
		return err
	}

	// Verificar se alguma linha foi afetada
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// Delete remove uma despesa
func (r *PostgresExpenseRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM expenses WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// Close fecha o pool de conexões
func (r *PostgresExpenseRepository) Close() {
	r.pool.Close()
}

// GetAllWithFilters retorna despesas com filtros, ordenação e paginação
// Conceito: Query dinâmica com parâmetros - protege contra SQL injection
func (r *PostgresExpenseRepository) GetAllWithFilters(ctx context.Context, filters *domain.ExpenseFilters) ([]*domain.Expense, error) {
	query, args := r.buildFilterQuery(filters, false)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []*domain.Expense
	for rows.Next() {
		var expense domain.Expense
		err := rows.Scan(
			&expense.ID,
			&expense.Description,
			&expense.Amount,
			&expense.Category,
			&expense.Date,
			&expense.CreatedAt,
			&expense.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, &expense)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return expenses, nil
}

// Count retorna o total de despesas que correspondem aos filtros
func (r *PostgresExpenseRepository) Count(ctx context.Context, filters *domain.ExpenseFilters) (int, error) {
	query, args := r.buildFilterQuery(filters, true)

	var count int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// buildFilterQuery constrói a query SQL com filtros dinâmicos
// Conceito: Prepared statements - protege contra SQL injection
func (r *PostgresExpenseRepository) buildFilterQuery(filters *domain.ExpenseFilters, isCount bool) (string, []interface{}) {
	var query string
	args := make([]interface{}, 0)
	argCount := 1

	// SELECT
	if isCount {
		query = "SELECT COUNT(*) FROM expenses WHERE 1=1"
	} else {
		query = "SELECT id, description, amount, category, date, created_at, updated_at FROM expenses WHERE 1=1"
	}

	// Filtros WHERE
	if filters.Category != nil {
		query += " AND category = $" + formatArgNum(argCount)
		args = append(args, *filters.Category)
		argCount++
	}

	if filters.MinAmount != nil {
		query += " AND amount >= $" + formatArgNum(argCount)
		args = append(args, *filters.MinAmount)
		argCount++
	}

	if filters.MaxAmount != nil {
		query += " AND amount <= $" + formatArgNum(argCount)
		args = append(args, *filters.MaxAmount)
		argCount++
	}

	if filters.StartDate != nil {
		query += " AND date >= $" + formatArgNum(argCount)
		args = append(args, *filters.StartDate)
		argCount++
	}

	if filters.EndDate != nil {
		query += " AND date <= $" + formatArgNum(argCount)
		args = append(args, *filters.EndDate)
		argCount++
	}

	// ORDER BY e LIMIT/OFFSET apenas para queries de dados
	if !isCount {
		// Ordenação
		sortColumn := filters.SortBy
		if sortColumn == "" {
			sortColumn = "date"
		}

		sortOrder := filters.SortOrder
		if sortOrder != "asc" && sortOrder != "desc" {
			sortOrder = "desc"
		}

		query += " ORDER BY " + sortColumn + " " + sortOrder

		// Paginação
		query += " LIMIT $" + formatArgNum(argCount)
		args = append(args, filters.Limit)
		argCount++

		query += " OFFSET $" + formatArgNum(argCount)
		args = append(args, filters.Offset)
	}

	return query, args
}

// formatArgNum formata o número do argumento para a query
func formatArgNum(num int) string {
	return fmt.Sprintf("%d", num)
}
