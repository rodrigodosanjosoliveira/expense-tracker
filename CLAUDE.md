# CLAUDE.md - Expense Tracker

## Missao
Executar tasks no projeto Expense Tracker (Go REST API) com baixo custo de token e baixa ambiguidade.

- Fonte de verdade: codigo.
- Se doc divergir do codigo, registrar `ASSUNCAO` e escalar para PM/Rodrigo.
- Nao inventar regra de negocio, contrato ou comportamento.
- Mudanca de arquitetura exige aprovacao explicita do Rodrigo.

## Leitura minima obrigatoria
1. [docs-ai/00-START-HERE.md](docs-ai/00-START-HERE.md)
2. [docs-ai/01-INVARIANTES-GLOBAIS.md](docs-ai/01-INVARIANTES-GLOBAIS.md)
3. [docs-ai/03-MAPA-RAPIDO-MODULOS.md](docs-ai/03-MAPA-RAPIDO-MODULOS.md)
4. [docs-ai/agents/00-ORQUESTRACAO-AGENTS.md](docs-ai/agents/00-ORQUESTRACAO-AGENTS.md)

## Regra de contexto minimo
1. Identifique o modulo impactado (consulte `docs-ai/03-MAPA-RAPIDO-MODULOS.md`).
2. Leia apenas os arquivos relevantes ao modulo:
   - expense: `internal/handler/expense_handler.go`, `internal/service/expense_service.go`, `internal/repository/postgres_expense_repository.go`
   - category: `internal/handler/category_handler.go`, `internal/service/category_service.go`, `internal/repository/category_repository.go`, `internal/repository/postgres_category_repository.go`, `internal/domain/category.go`
   - auth: `internal/handler/auth_handler.go`, `internal/service/auth_service.go`, `internal/middleware/auth.go`
   - webhook: `internal/handler/webhook_handler.go`, `internal/service/webhook_service.go`
   - notification: `internal/service/notification_service.go`, `internal/domain/notification.go`
   - domain: `internal/domain/`
   - infra: `cmd/api/main.go`, `internal/config/`, `internal/database/`, `migrations/`
3. So complemente com docs globais quando houver risco transversal:
   - auth/seguranca: `internal/middleware/auth.go`, `internal/service/auth_service.go`
   - banco/migrations: `migrations/`, `internal/database/postgres.go`
   - config/deploy: `docker-compose.yml`, `Dockerfile`, `.env.example`
   - testes: `internal/handler/*_test.go`, `internal/service/*_test.go`
4. Nao ler repositorio inteiro por padrao.

## Invariantes operacionais
- **Layered Architecture**: Handler nao acessa Repository diretamente — sempre via Service.
- **user_id scoping**: Toda query e scoped por `user_id`. Ausencia e bug de seguranca critico.
- **Auth**: Nenhum endpoint de dados sem middleware JWT. `user_id` extraido via `middleware.GetUserIDFromContext(r.Context())`.
- **Domain puro**: `internal/domain/` nao importa nenhum outro pacote interno do projeto.
- **Erros sentinela**: Usar `domain.ErrXxx` e `repository.ErrNotFound` — nunca strings literais de erro.
- **Migrations**: Toda migration tem arquivo `up` e `down`. Numbering sequencial `000N_nome.{up,down}.sql`.
- **Construtores**: Padrao `NewXxx(...)` com injecao de dependencia via interfaces.
- Mudou codigo, atualize docs relacionadas no mesmo PR.

## Fluxo de agents

> **REGRA BLOQUEANTE — sem excecao.**
> Para qualquer pedido de implementacao, nova feature, bug fix ou mudanca de comportamento:
> - NUNCA escrever codigo diretamente, mesmo que o prompt ja descreva a solucao completa.
> - SEMPRE iniciar pelo agent `pm-triage` (`/triage`) antes de qualquer linha de codigo.
> - SEMPRE aguardar cada gate antes de avancar para o proximo.
> - Ignorar esta regra — inclusive por "eficiencia" ou "o prompt ja estava detalhado" — e um erro critico.

Sequencia obrigatoria: `PM/Triage -> Architecture/Guardrails -> [Data-Model se tocar domain/entity/migration] -> [Test-Writer] -> Dev -> [Code-Reviewer] -> QA -> Docs`.

- `Security` quando tocar auth, JWT, senhas, segredos, middleware, integracao externa.
- `Release/Ops` quando tocar Docker, docker-compose, Dockerfile, migrations de deploy.
- `Migration-Reviewer` obrigatorio junto com `Release/Ops` quando houver mudanca de schema SQL.
- `Data-Model` obrigatorio quando tocar domain entities, repository interfaces ou schema de banco.
- `Test-Writer` obrigatorio antes ou junto com `Dev` em toda delivery com criterio de aceite testavel.
- `Code-Reviewer` obrigatorio apos `Dev` e antes de `QA` em toda delivery.
- Subagents do Claude em `.claude/agents/`:
  - `pm-triage`
  - `architecture-guardrails`
  - `dev-implementation`
  - `qa-validation`
  - `docs-update`
  - `security-review`
  - `release-ops`
  - `code-reviewer`
  - `data-model`
  - `context-loader`
  - `migration-reviewer`
  - `query-performance`
  - `test-writer`
  - `git-delivery`
  - `agent-auditor`
- Comandos do Claude em `.claude/commands/`:
  - `triage`
  - `guardrails`
  - `dev`
  - `qa`
  - `docs`
  - `security`
  - `release-ops`
  - `handoff`
  - `report`
  - `validate-delivery`
  - `code-review`
  - `data-model`
  - `context-loader`
  - `migration-review`
  - `query-performance`
  - `test-writer`
  - `git-delivery`
  - `agent-audit`
- Handoff: [docs-ai/agents/08-HANDOFF-CONTRACT.md](docs-ai/agents/08-HANDOFF-CONTRACT.md)
- Report obrigatorio: `docs-ai/deliveries/<DELIVERY_ID>/report.md`
- Template: [docs-ai/deliveries/_template/report.md](docs-ai/deliveries/_template/report.md)

## Governanca de camadas

- Regras persistentes do Claude: `CLAUDE.md`
- Prompts executaveis: `.claude/agents/`
- Skills reutilizaveis: `.claude/skills/`
- Docs operacionais e artefatos de delivery: `docs-ai/`
- Hook repo-level: `scripts/ai/block-direct-edit.sh`
- Bootstrap do hook: `scripts/ai/install-claude-hooks.sh`

## Comandos essenciais

### Testes
```bash
go test ./...                              # todos os testes
go test -v ./...                           # verbose
go test -cover ./...                       # com cobertura
go test -v ./internal/handler/            # pacote especifico
make test
make test-verbose
make test-coverage
```

### Build
```bash
make build                                 # bin/api
go build -o bin/api ./cmd/api/main.go
```

### Lint / analise
```bash
go fmt ./...
go vet ./...
make lint
```

### Servidor local
```bash
make run                                   # go run cmd/api/main.go
docker compose up -d --build              # via Docker
```

### Banco de dados (requer Docker)
```bash
make db-up                                 # sobe PostgreSQL
make db-migrate-up                         # aplica migrations
make db-migrate-down                       # rollback 1 migration
make db-reset                              # down + up + migrate
make db-create-migration NAME=nome         # cria nova migration
```

### Docs
```bash
make swagger-gen                           # regenera Swagger/OpenAPI
```

### Validacao de delivery
```bash
scripts/ai/validate-delivery.sh <DELIVERY_ID>
```

## Regra de execucao de testes
- Testes unitarios, de handler e de servico: `go test -v ./internal/...`
- Foque no pacote impactado — nao rode tudo sem necessidade.
- Testes de repositorio usam o in-memory repository — nao requerem banco de dados externo.

## Hotspots do repo
- Handlers: `internal/handler/expense_handler.go`, `internal/handler/auth_handler.go`, `internal/handler/webhook_handler.go`, `internal/handler/category_handler.go`
- Services: `internal/service/expense_service.go`, `internal/service/auth_service.go`, `internal/service/notification_service.go`, `internal/service/category_service.go`
- Repositories: `internal/repository/postgres_expense_repository.go`, `internal/repository/memory_expense_repository.go`, `internal/repository/postgres_category_repository.go`, `internal/repository/memory_category_repository.go`
- Domain: `internal/domain/expense.go`, `internal/domain/user.go`, `internal/domain/filters.go`, `internal/domain/category.go`
- Auth middleware: `internal/middleware/auth.go`
- Wiring: `cmd/api/main.go`
- Migrations: `migrations/`
- Config: `internal/config/config.go`
- Testes: `internal/handler/*_test.go`, `internal/service/*_test.go`

## Modulos — referencias rapidas
Ver taxonomia completa em: [docs-ai/03-MAPA-RAPIDO-MODULOS.md](docs-ai/03-MAPA-RAPIDO-MODULOS.md)

## Regra final
Use [docs-ai/00-START-HERE.md](docs-ai/00-START-HERE.md) como ponto de entrada, nao como leitura inicial completa do repo.
