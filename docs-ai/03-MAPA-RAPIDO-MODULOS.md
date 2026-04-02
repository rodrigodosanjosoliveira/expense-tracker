---
status: VALIDADO
owner: Rodrigo
ultima_validacao: 2026-04-02
criterio_validacao: codigo+teste
fonte_de_verdade: codigo
escalonamento: PM/Rodrigo
---

# Mapa Rapido de Modulos — Expense Tracker

## Taxonomia

```
Expense Tracker (Go REST API)
├── expense      [CRUD de despesas, filtros, paginacao, notificacoes]
│   ├── Handler:    internal/handler/expense_handler.go
│   ├── Service:    internal/service/expense_service.go
│   ├── Repo iface: internal/repository/expense_repository.go
│   ├── Repo pg:    internal/repository/postgres_expense_repository.go
│   └── Repo mem:   internal/repository/memory_expense_repository.go
│
├── auth         [registro, login, geracao e validacao de JWT]
│   ├── Handler:    internal/handler/auth_handler.go
│   ├── Service:    internal/service/auth_service.go
│   ├── Repo iface: internal/repository/user_repository.go
│   └── Repo pg:    internal/repository/postgres_user_repository.go
│
├── webhook      [CRUD de webhooks, trigger automatico por evento de despesa]
│   ├── Handler:    internal/handler/webhook_handler.go
│   ├── Service:    internal/service/webhook_service.go
│   ├── Repo iface: internal/repository/webhook_repository.go
│   └── Repo pg:    internal/repository/postgres_webhook_repository.go
│
├── notification [dispatch de eventos para webhooks registrados]
│   ├── Service:    internal/service/notification_service.go
│   └── Domain:     internal/domain/notification.go
│
├── domain       [entidades, validacao, erros sentinela, filtros]
│   ├── internal/domain/expense.go        (Expense entity + validation)
│   ├── internal/domain/user.go           (User entity + bcrypt)
│   ├── internal/domain/webhook.go        (Webhook entity)
│   ├── internal/domain/notification.go   (Event types)
│   └── internal/domain/filters.go        (ExpenseFilters + pagination)
│
├── middleware   [validacao JWT, extracao de user_id do context]
│   └── internal/middleware/auth.go
│
└── infra        [config, banco, wiring, migrations, Docker]
    ├── cmd/api/main.go                   (wiring de dependencias)
    ├── internal/config/config.go         (env vars, validacao)
    ├── internal/database/postgres.go     (pgx pool configuration)
    ├── migrations/                       (SQL migrations up/down)
    ├── docker-compose.yml
    ├── Dockerfile
    └── Makefile
```

## Pacote minimo de leitura por modulo

| Modulo | Arquivos-chave |
|--------|---------------|
| expense | `internal/handler/expense_handler.go`, `internal/service/expense_service.go`, `internal/repository/postgres_expense_repository.go`, `internal/domain/expense.go`, `internal/domain/filters.go` |
| auth | `internal/handler/auth_handler.go`, `internal/service/auth_service.go`, `internal/repository/postgres_user_repository.go`, `internal/domain/user.go`, `internal/middleware/auth.go` |
| webhook | `internal/handler/webhook_handler.go`, `internal/service/webhook_service.go`, `internal/repository/postgres_webhook_repository.go`, `internal/domain/webhook.go` |
| notification | `internal/service/notification_service.go`, `internal/domain/notification.go` |
| domain | `internal/domain/` (todos os arquivos) |
| infra | `cmd/api/main.go`, `internal/config/config.go`, `internal/database/postgres.go`, `migrations/` |

## Testes por modulo

| Modulo | Arquivos de teste |
|--------|------------------|
| expense handler | `internal/handler/expense_handler_test.go` |
| auth handler | `internal/handler/auth_handler_test.go` |
| webhook handler | `internal/handler/webhook_handler_test.go` |
| domain | `internal/domain/*_test.go` |
| service | `internal/service/*_test.go` |

## Prefixos de DELIVERY_ID

| Modulo | Prefixo |
|--------|---------|
| Expense Management | `EXPENSE` |
| Auth / Users | `AUTH` |
| Webhooks | `WEBHOOK` |
| Notifications | `NOTIF` |
| Domain / Entities | `DOMAIN` |
| Infra (config, DB, migrations, Docker) | `INFRA` |

## Rotas da API por modulo

| Modulo | Rotas |
|--------|-------|
| auth | `POST /auth/register`, `POST /auth/login` |
| expense | `GET /expenses`, `POST /expenses`, `GET /expenses/{id}`, `PUT /expenses/{id}`, `DELETE /expenses/{id}` |
| webhook | `GET /webhooks`, `POST /webhooks`, `GET /webhooks/{id}`, `DELETE /webhooks/{id}` |
| health | `GET /health` (sem auth) |
