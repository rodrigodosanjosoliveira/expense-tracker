# Copilot / AI agent instructions — Expense Tracker

Short, actionable guidance for editing and extending this repository.

1) Big picture
- Layered Go service: entrypoint `cmd/api/main.go` initializes dependencies and wires: `internal/handler` -> `internal/service` -> `internal/repository` -> `internal/domain`.
- Auth, webhooks and notifications are optional and enabled only when PostgreSQL is configured (see `internal/config/config.go` and `cmd/api/main.go`).

2) Running & developer workflows
- Run server: `make run` or `go run cmd/api/main.go`.
- Tests: `make test` or `go test ./...`. Use `make test-verbose` for detailed output.
- Swagger: update docs after changing handlers/comments with `make swagger-gen` (uses `swag init -g cmd/api/main.go -o docs`).
- DB: use Docker Compose targets in `Makefile`: `make db-up`, then `make db-migrate-up`. Create migrations: `make db-create-migration NAME=your_name`.

3) Important runtime configuration
- The project reads config from env vars (see `internal/config/config.go`).
- To enable Postgres (authentication + webhooks): set `DB_HOST`/`DB_USER`/`DB_NAME` etc. If DB host is empty the app uses an in-memory expense repository.
- `JWT_SECRET` is required and validated on startup; set it in your env or `.env` (otherwise `Config.Validate()` fails).

4) Project-specific patterns & conventions
- Dependency injection by interface: repositories and services are passed into constructors (e.g. `NewExpenseService(repo, idGen)`). Favor this when adding features.
- ID generation is abstracted by `service.IDGenerator` (see `internal/service/uuid_generator.go`) so tests can inject deterministic IDs.
- Constructors follow `NewXxx(...)` naming; prefer them for initialization and tests.
- Errors use package-level sentinel errors: domain errors (e.g. `domain.ErrEmptyDescription`) and repository errors (e.g. `repository.ErrNotFound`, `repository.ErrAlreadyExists`) — handle them explicitly in handlers/services.

5) Security & data isolation (critical)
- Always ensure queries are scoped by `user_id`. The Postgres repository builds queries around `filters.UserID` in `internal/repository/postgres_expense_repository.go`. Missing the `user_id` filter is a security bug.
- Handlers expect the authenticated `user_id` in context (see `internal/middleware` and how handlers use `middleware.GetUserIDFromContext`). If you change auth flow, update these call sites.

6) Database vs in-memory behavior
- Tests and quick dev runs use `NewMemoryExpenseRepository()` (in-memory). Some features (users, webhooks) require Postgres; the code logs warnings when those are disabled (see `cmd/api/main.go`).
- When adding DB-backed features, add SQL files under `migrations/` and document migration steps in `Makefile`.

7) Code changes checklist (pre-PR)
- Run `go fmt ./...` and `go vet ./...`.
- Run `make test-coverage` or at minimum `go test ./...`.
- If modifying handlers or adding endpoints, run `make swagger-gen` to regenerate `docs/`.
- If changes affect DB schema, add migration files in `migrations/` and include `make db-migrate-up` instructions in PR description.

8) Tests & examples
- Handlers are tested with `httptest` and often use the in-memory repository (see `internal/handler/expense_handler_test.go`). Follow the same pattern for new handler tests.
- Service and repository tests use table-driven style and explicit setup of repos (see `internal/service/expense_service_test.go` and `internal/repository/memory_expense_repository_test.go`).

9) Useful files to inspect when working on a change
- Entrypoint and wiring: `cmd/api/main.go`
- Config: `internal/config/config.go`
- Handlers: `internal/handler/` (e.g. `expense_handler.go`, `auth_handler.go`)
- Services: `internal/service/` (e.g. `expense_service.go`, `auth_service.go`)
- Repositories: `internal/repository/` (e.g. `postgres_expense_repository.go`, `memory_expense_repository.go`)
- Domain models and validation: `internal/domain/` (e.g. `expense.go`)
- Migrations: `migrations/`

If anything above is unclear or you want more examples (unit test templates, common PR checklist, or a curated list of TODOs), tell me which section to expand. 
