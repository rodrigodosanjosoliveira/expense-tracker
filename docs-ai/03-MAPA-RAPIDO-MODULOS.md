---
status: VALIDADO
owner: Rodrigo
ultima_validacao: 2026-04-02
criterio_validacao: codigo+teste
fonte_de_verdade: codigo
escalonamento: PM/Rodrigo
ultima_atualizacao: EXPENSE-001
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
├── category     [CRUD de categorias normalizadas com FK, scoped por user_id]
│   ├── Handler:    internal/handler/category_handler.go
│   ├── Service:    internal/service/category_service.go
│   ├── Repo iface: internal/repository/category_repository.go
│   ├── Repo pg:    internal/repository/postgres_category_repository.go
│   └── Repo mem:   internal/repository/memory_category_repository.go
│
├── notification [dispatch de eventos para webhooks registrados]
│   ├── Service:    internal/service/notification_service.go
│   └── Domain:     internal/domain/notification.go
│
├── domain       [entidades, validacao, erros sentinela, filtros]
│   ├── internal/domain/expense.go        (Expense entity + validation; CategoryID *string)
│   ├── internal/domain/category.go       (Category entity + validation + sentinel errors)
│   ├── internal/domain/user.go           (User entity + bcrypt)
│   ├── internal/domain/webhook.go        (Webhook entity)
│   ├── internal/domain/notification.go   (Event types)
│   └── internal/domain/filters.go        (ExpenseFilters + pagination; CategoryID *string)
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
| category | `internal/handler/category_handler.go`, `internal/service/category_service.go`, `internal/repository/category_repository.go`, `internal/repository/postgres_category_repository.go`, `internal/domain/category.go` |
| auth | `internal/handler/auth_handler.go`, `internal/service/auth_service.go`, `internal/repository/postgres_user_repository.go`, `internal/domain/user.go`, `internal/middleware/auth.go` |
| webhook | `internal/handler/webhook_handler.go`, `internal/service/webhook_service.go`, `internal/repository/postgres_webhook_repository.go`, `internal/domain/webhook.go` |
| notification | `internal/service/notification_service.go`, `internal/domain/notification.go` |
| domain | `internal/domain/` (todos os arquivos) |
| infra | `cmd/api/main.go`, `internal/config/config.go`, `internal/database/postgres.go`, `migrations/` |

## Testes por modulo

| Modulo | Arquivos de teste |
|--------|------------------|
| expense handler | `internal/handler/expense_handler_test.go` |
| category handler | `internal/handler/category_handler_test.go` |
| category service | `internal/service/category_service_test.go` |
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
| category | `GET /categories`, `POST /categories`, `GET /categories/{id}`, `PUT /categories/{id}`, `DELETE /categories/{id}` |
| webhook | `GET /webhooks`, `POST /webhooks`, `GET /webhooks/{id}`, `DELETE /webhooks/{id}` |
| health | `GET /health` (sem auth) |

## Dependencias entre modulos

| Modulo | Depende de | Observacao |
|--------|-----------|-----------|
| expense | domain, auth (middleware JWT), category (via SetCategoryService) | `ExpenseService` injeta `CategoryServiceInterface` para resolver `category_id` |
| category | domain, auth (middleware JWT) | Sem dependencia de outros modulos de negocio |
| webhook | domain, auth (middleware JWT), notification | Disparo automatico a cada evento de despesa |
| notification | domain, webhook | Fire-and-forget, nao bloqueia response |

## Notas de backward compatibility

- `domain.Expense.Category` (string) permanece como campo de OUTPUT populado via JOIN/lookup. Aceito como INPUT para lookup por nome (sem auto-create de categoria).
- `domain.Expense.CategoryID *string` adicionado em EXPENSE-001 — nullable para suportar despesas legadas (pre-migration 000006).
- Despesas legadas sem `category_id` retornam `"category_id": null` e `"category": ""` no JSON.
