# Como executar o Expense Tracker localmente

API REST em Go (1.25) para controle de despesas pessoais, com arquitetura em camadas
(handler → service → repository), autenticação JWT, webhooks e documentação Swagger.

## Pré-requisitos

| Ferramenta | Versão | Obrigatório para |
|------------|--------|------------------|
| Go | 1.25+ | build/run/test |
| Docker + Docker Compose | recente | PostgreSQL e execução em container |
| `migrate` (golang-migrate) | v4 | aplicar migrations manualmente (instalado sob demanda pelo Makefile) |
| `make` | — | usar os atalhos do Makefile (opcional no Windows) |

> **Windows:** o `Makefile` depende de utilitários POSIX (`grep`, `awk`, `sleep`, `test`).
> Sem Git Bash/WSL, prefira os **comandos diretos** listados em cada seção.

---

## Modos de execução

A aplicação decide o repositório em runtime pela variável **`DB_HOST`**
(`internal/config/config.go` → `UsePostgreSQL()`):

- **`DB_HOST` vazio** → repositório **in-memory**. Sobe rápido, mas **auth, webhooks e
  notificações ficam desabilitados** (só existem em PostgreSQL). As rotas de expenses/categories
  ficam registradas, porém **retornam 401** porque os handlers exigem `user_id` no contexto (injetado apenas pelo middleware JWT).
- **`DB_HOST` preenchido** → **PostgreSQL**. Habilita auth JWT, webhooks e notificações.

---

## Opção 1 — Tudo via Docker Compose (mais simples)

Sobe PostgreSQL + API juntos.

```bash
# JWT_SECRET é obrigatório e não pode ser o valor padrão (config.Validate() falha)
# PowerShell:
$env:JWT_SECRET = "um-segredo-bem-grande-e-aleatorio"
docker compose up -d --build

# Bash:
JWT_SECRET="um-segredo-bem-grande-e-aleatorio" docker compose up -d --build
```

- API: http://localhost:8080
- Health: http://localhost:8080/health
- Swagger: http://localhost:8080/swagger/index.html

Parar: `docker compose down` (adicione `-v` para apagar o volume do banco).

> **Atenção (caveat de schema):** o compose monta `./migrations` em
> `/docker-entrypoint-initdb.d`. O Postgres executa **todos** os `*.sql` desse diretório na
> primeira inicialização, incluindo os arquivos `*.down.sql`. Se as tabelas não aparecerem
> como esperado, use a Opção 2 (Postgres no Docker + `migrate` aplicando só os `*.up.sql`),
> ou limpe o diretório de init. Rode `make db-migrate-up` depois para garantir o schema correto.

---

## Opção 2 — API local (Go) + PostgreSQL no Docker (recomendado p/ desenvolvimento)

### 1. Configurar variáveis de ambiente

```bash
cp .env.example .env
```

Edite o `.env` e **troque o `JWT_SECRET`** (o valor padrão é rejeitado na validação):

```env
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
DB_HOST=localhost
DB_PORT=5432
DB_USER=expensetracker
DB_PASSWORD=secret
DB_NAME=expenses
DB_SSLMODE=disable
JWT_SECRET=um-segredo-bem-grande-e-aleatorio
ENV=development
```

> O `main.go` carrega o `.env` via `godotenv.Load()`. As credenciais acima batem com o
> serviço `postgres` do `docker-compose.yml`.

### 2. Subir o PostgreSQL

```bash
docker compose up -d postgres
# ou: make db-up
```

### 3. Aplicar as migrations

```bash
# Via Makefile (instala o migrate se necessário):
make db-migrate-up

# Comando direto (equivalente):
migrate -path migrations \
  -database "postgres://expensetracker:secret@localhost:5432/expenses?sslmode=disable" up
```

Rollback de 1 migration: `make db-migrate-down` · Reset completo: `make db-reset`.

### 4. Rodar a API

```bash
go run cmd/api/main.go
# ou: make run
# ou build: go build -o bin/api cmd/api/main.go  &&  ./bin/api
```

Logs esperados confirmam o modo:
`✓ Connected to PostgreSQL`, `✓ Authentication enabled`, `✓ Webhooks and notifications enabled`.

---

## Opção 3 — Execução rápida em memória (sem banco)

Para experimentar expenses/categories sem subir Postgres. **Sem auth/webhooks.**

```bash
# Deixe DB_HOST vazio e ainda assim defina um JWT_SECRET válido (a validação sempre roda)
# PowerShell:
$env:DB_HOST=""; $env:JWT_SECRET="um-segredo-valido"; go run cmd/api/main.go
```

Você verá `Using in-memory repository` e avisos de que auth/webhooks estão desabilitados.

---

## Endpoints principais

| Rota | Métodos | Auth |
|------|---------|------|
| `/health` | GET | não |
| `/auth/register`, `/auth/login` | POST | não (só com PostgreSQL) |
| `/expenses`, `/expenses/{id}` | GET/POST/PUT/DELETE | JWT (se PostgreSQL) |
| `/categories`, `/categories/{id}` | GET/POST/PUT/DELETE | JWT (se PostgreSQL) |
| `/webhooks`, `/webhooks/{id}` | GET/POST/PUT/DELETE | JWT (só com PostgreSQL) |
| `/swagger/` | GET | não |

Fluxo típico: `POST /auth/register` → `POST /auth/login` (retorna JWT) →
enviar `Authorization: Bearer <token>` nas rotas de dados.

---

## Testes, lint e docs

```bash
go test ./...            # todos os testes   (make test)
go test -cover ./...     # com cobertura     (make test-coverage)
go test -v ./internal/handler/   # pacote específico

go fmt ./...             # formatação        (make fmt)
go vet ./...             # análise estática  (make vet)

make swagger-gen         # regenera docs/ do Swagger
```

Os testes de repositório usam implementação in-memory — **não exigem banco externo**.

---

## Troubleshooting

- **`JWT_SECRET is required and must be changed from default`** → defina um `JWT_SECRET`
  diferente de `your-secret-key-change-me` / `change-me-to-a-secure-secret`.
- **`Failed to connect to database`** → confira se o container `postgres` está de pé
  (`docker compose ps`) e se `DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME` no `.env` batem
  com o compose.
- **Auth/webhooks não funcionam** → `DB_HOST` está vazio; a app está em modo in-memory.
- **Porta 8080/5432 ocupada** → ajuste `SERVER_PORT` no `.env` ou o mapeamento de portas no
  `docker-compose.yml`.
