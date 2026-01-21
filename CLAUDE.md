# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Expense Tracker is a REST API for personal expense management built in Go using TDD. It uses a layered architecture with dependency injection.

## Common Commands

```bash
# Run the server
make run                    # or: go run cmd/api/main.go

# Tests
make test                   # Run all tests
make test-verbose           # Run tests with detailed output
make test-coverage          # Run tests with coverage report
go test -v ./internal/handler/  # Run tests for a specific package

# Code quality
go fmt ./...               # Format code
go vet ./...               # Analyze code

# Swagger documentation (regenerate after modifying handlers)
make swagger-gen           # or: swag init -g cmd/api/main.go -o docs

# Database (requires Docker)
make db-up                 # Start PostgreSQL
make db-migrate-up         # Run migrations
make db-create-migration NAME=migration_name  # Create new migration
make db-reset              # Reset database (down + up + migrate)

# Build
make build                 # Outputs binary to bin/api
```

## Architecture

```
cmd/api/main.go          → Entrypoint, wires all dependencies
internal/
  config/                → Environment configuration (Config struct)
  handler/               → HTTP handlers (controllers)
  service/               → Business logic layer
  repository/            → Data persistence layer (interfaces + implementations)
  domain/                → Entities, validation rules, and sentinel errors
  middleware/            → HTTP middleware (auth)
  database/              → Database connection management
migrations/              → SQL migration files
docs/                    → Auto-generated Swagger documentation
```

**Request flow:** Handler → Service → Repository → Domain

## Key Patterns

- **Dependency injection via interfaces:** Repositories and services are passed into constructors (e.g., `NewExpenseService(repo, idGen)`)
- **ID generation abstraction:** `service.IDGenerator` interface allows deterministic IDs in tests
- **Sentinel errors:** Domain errors (`domain.ErrEmptyDescription`, `domain.ErrInvalidAmount`) and repository errors (`repository.ErrNotFound`, `repository.ErrAlreadyExists`) — handle these explicitly
- **Constructor naming:** Use `NewXxx(...)` pattern for initialization

## Configuration

The app reads config from environment variables (see `internal/config/config.go`):

- `JWT_SECRET` — Required, must be changed from default
- `DB_HOST`, `DB_USER`, `DB_NAME`, `DB_PASSWORD`, `DB_PORT` — When `DB_HOST` is set, PostgreSQL is used; otherwise in-memory repository
- `SERVER_PORT`, `SERVER_HOST` — Server binding (defaults: 8080, 0.0.0.0)

**Important:** Authentication and webhooks require PostgreSQL. Without `DB_HOST`, the app runs in development mode with in-memory storage and no auth.

## Security Requirement

**Always scope queries by `user_id`.** The Postgres repository builds queries using `filters.UserID` in `internal/repository/postgres_expense_repository.go`. Missing the `user_id` filter is a security bug.

Handlers retrieve the authenticated user via `middleware.GetUserIDFromContext(r.Context())`.

## Testing Patterns

- Handlers use `httptest` with in-memory repository (see `internal/handler/expense_handler_test.go`)
- Services and repositories use table-driven tests with `t.Run()` subtests
- Inject `MockIDGenerator` for deterministic IDs in tests
