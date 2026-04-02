---
status: VALIDADO
owner: Rodrigo
ultima_validacao: 2026-04-02
criterio_validacao: codigo+teste
fonte_de_verdade: codigo
escalonamento: PM/Rodrigo
---

# Invariantes Globais — Expense Tracker

Aplique a qualquer feature, melhoria ou bugfix. Sem excecao.

## Arquitetura Layered

- **Handler -> Service -> Repository -> Domain**: camada superior nunca pula camada inferior.
- Handler nao acessa Repository diretamente — sempre via Service interface.
- Service nao acessa banco diretamente — sempre via Repository interface.
- `internal/domain/` nao importa nenhum outro pacote interno do projeto.
- Construtores seguem padrao `NewXxx(...)` com injecao de dependencia via interfaces.
- Mudanca de arquitetura exige aprovacao explicita do Rodrigo.

## Auth e Seguranca

- Nenhum endpoint de dados sem middleware JWT (`internal/middleware/auth.go`).
- `user_id` extraido SEMPRE via `middleware.GetUserIDFromContext(r.Context())` — nunca de query param ou body.
- `JWT_SECRET` nunca pode ter valor default inseguro em producao — validado em `internal/config/config.go`.
- Senhas armazenadas APENAS como hash bcrypt (`domain.User.PasswordHash`) — nunca em texto plano.
- Token JWT tem expiracao de 24h — nao aumentar sem justificativa explicita.

## Isolamento de Dados (user_id scoping)

- **TODA query SQL e scoped por `user_id`** — ausencia e bug de seguranca critico.
- Repository recebe `filters.UserID` obrigatoriamente em `internal/repository/postgres_expense_repository.go`.
- InMemory repository deve respeitar o mesmo isolamento por `user_id`.
- Nunca retornar dados de um usuario em request autenticado de outro usuario.

## Banco de Dados (PostgreSQL + pgx)

- Toda migration tem arquivo `up` e `down` obrigatorios.
- Migrations numeradas sequencialmente: `000N_nome_descritivo.{up,down}.sql`.
- Migrations sao idempotentes e nao causam perda de dados no rollback.
- Sem ORM — queries raw via pgx. Sem mapeamento automatico magico.
- Pool de conexoes: max 25, min 5 (definido em `internal/database/postgres.go`).
- Modo desenvolvimento: `DB_HOST` ausente ativa in-memory repository (sem auth).

## Erros Sentinela

- Usar `domain.ErrXxx` para erros de negocio (ex: `ErrEmptyDescription`, `ErrInvalidAmount`).
- Usar `repository.ErrNotFound`, `repository.ErrAlreadyExists` para erros de persistencia.
- Nunca comparar erros por string literal — usar `errors.Is()`.
- Handlers mapeiam erros sentinela para HTTP status codes apropriados.

## Testes

- Handlers testados com `httptest` + in-memory repository (sem banco real).
- Services e repositories usam table-driven tests com `t.Run()`.
- `MockIDGenerator` injetado para IDs deterministicos em testes.
- Nunca mockar banco em testes de integracao — usar PostgreSQL real.
- Testes unitarios nao dependem de variaveis de ambiente externas.
- Cobertura minima: todo criterio de aceite deve ter pelo menos um teste.

## Webhooks e Notificacoes

- Webhooks sao scoped por `user_id` — nunca disparar webhook de outro usuario.
- Notificacoes sao fire-and-forget (nao bloqueiam resposta da API).
- Eventos validos definidos em `internal/domain/notification.go`.

## Deploy e Rollback

- Rollback de migration via `make db-migrate-down` sem perda de dados.
- Sem CI/CD pipeline — deploy manual via Docker Compose.
- Toda mudanca operacional (Docker, migrations) precisa nota de Release/Ops no report.
- `JWT_SECRET` e variaveis de banco nao podem ter defaults inseguros em producao.

## Doc e Confianca

- Docs atualizadas no mesmo PR do codigo.
- Status de doc: `DRAFT`, `VALIDADO`, `ASSUNCAO`, `LEGADO`.
- `VALIDADO` so com codigo + teste como evidencia.
- Conflito doc x codigo -> `ASSUNCAO` + escalar para Rodrigo.
