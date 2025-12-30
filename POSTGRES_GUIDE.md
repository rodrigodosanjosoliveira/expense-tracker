# Guia PostgreSQL - Expense Tracker

## Visão Geral

A API agora suporta **dois modos de persistência**:

1. **In-Memory** (padrão) - Para desenvolvimento rápido e testes
2. **PostgreSQL** - Para produção e desenvolvimento com persistência

## Modo In-Memory (Padrão)

Por padrão, a aplicação usa armazenamento em memória:

```bash
go run cmd/api/main.go
# Output: Using in-memory repository (set DB_HOST to use PostgreSQL)
```

**Vantagens:**
- Zero configuração
- Perfeito para testes rápidos
- Não precisa de banco de dados

**Desvantagens:**
- Dados são perdidos quando o servidor é reiniciado
- Não é adequado para produção

## Modo PostgreSQL

### Setup Rápido com Docker

```bash
# 1. Iniciar PostgreSQL
make db-up

# 2. Executar migrations (criar tabelas)
make db-migrate-up

# 3. Rodar aplicação com PostgreSQL
DB_HOST=localhost go run cmd/api/main.go
# Output: Initializing PostgreSQL repository...
#         ✓ Connected to PostgreSQL
```

### Setup Manual

#### 1. Instalar PostgreSQL

**Ubuntu/Debian:**
```bash
sudo apt-get install postgresql postgresql-contrib
```

**macOS:**
```bash
brew install postgresql@16
brew services start postgresql@16
```

**Windows:**
Download do [site oficial](https://www.postgresql.org/download/windows/)

#### 2. Criar Database e Usuário

```bash
# Conectar ao PostgreSQL
psql postgres

# Criar usuário
CREATE USER expensetracker WITH PASSWORD 'secret';

# Criar database
CREATE DATABASE expenses OWNER expensetracker;

# Dar permissões
GRANT ALL PRIVILEGES ON DATABASE expenses TO expensetracker;

# Sair
\q
```

#### 3. Executar Migrations

```bash
make db-migrate-up
```

#### 4. Configurar Variáveis de Ambiente

Crie um arquivo `.env` (ou exporte as variáveis):

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=expensetracker
export DB_PASSWORD=secret
export DB_NAME=expenses
export DB_SSLMODE=disable
```

#### 5. Rodar Aplicação

```bash
# Carregar .env e rodar
source .env
go run cmd/api/main.go
```

## Comandos Make Úteis

```bash
# Banco de Dados
make db-up              # Iniciar PostgreSQL (Docker)
make db-down            # Parar PostgreSQL
make db-migrate-up      # Executar migrations
make db-migrate-down    # Reverter última migration
make db-reset           # Resetar banco (down + up + migrate)

# Criar nova migration
make db-create-migration NAME=add_users_table

# Aplicação
make help               # Ver todos os comandos disponíveis
```

## Estrutura do Banco de Dados

### Tabela: expenses

```sql
CREATE TABLE expenses (
    id UUID PRIMARY KEY,
    description VARCHAR(255) NOT NULL,
    amount DECIMAL(10, 2) NOT NULL CHECK (amount > 0),
    category VARCHAR(100) NOT NULL,
    date TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Índices
CREATE INDEX idx_expenses_category ON expenses(category);
CREATE INDEX idx_expenses_date ON expenses(date);
CREATE INDEX idx_expenses_created_at ON expenses(created_at);
```

## Migrations

### O que são Migrations?

Migrations são scripts SQL versionados que modificam o schema do banco de dados. Permitem:

- **Versionamento** do schema do banco
- **Rollback** de mudanças
- **Consistência** entre ambientes
- **Histórico** de alterações

### Estrutura das Migrations

```
migrations/
├── 000001_create_expenses_table.up.sql    # Criar tabela
├── 000001_create_expenses_table.down.sql  # Reverter criação
├── 000002_add_user_column.up.sql          # Adicionar coluna
└── 000002_add_user_column.down.sql        # Remover coluna
```

### Criar Nova Migration

```bash
make db-create-migration NAME=add_category_index
```

Isso cria dois arquivos:
- `000002_add_category_index.up.sql` - Aplicar mudança
- `000002_add_category_index.down.sql` - Reverter mudança

**Exemplo de UP:**
```sql
CREATE INDEX idx_expenses_category ON expenses(category);
```

**Exemplo de DOWN:**
```sql
DROP INDEX IF EXISTS idx_expenses_category;
```

### Executar Migrations

```bash
# Aplicar todas as migrations pendentes
make db-migrate-up

# Reverter última migration
make db-migrate-down

# Ver status das migrations
migrate -path migrations -database "$(DB_URL)" version
```

## Docker Compose

O projeto inclui PostgreSQL no `docker-compose.yml`:

```yaml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: expensetracker
      POSTGRES_PASSWORD: secret
      POSTGRES_DB: expenses
    ports:
      - "5432:5432"
```

### Rodar com Docker Compose

```bash
# Iniciar todos os serviços (API + PostgreSQL)
docker-compose up

# Apenas PostgreSQL
docker-compose up -d postgres

# Ver logs
docker-compose logs -f postgres

# Parar tudo
docker-compose down

# Parar e remover volumes (apaga dados!)
docker-compose down -v
```

## Variáveis de Ambiente

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `DB_HOST` | localhost | Host do PostgreSQL |
| `DB_PORT` | 5432 | Porta do PostgreSQL |
| `DB_USER` | expensetracker | Usuário do banco |
| `DB_PASSWORD` | secret | Senha do banco |
| `DB_NAME` | expenses | Nome do database |
| `DB_SSLMODE` | disable | Modo SSL (disable/require) |
| `SERVER_PORT` | 8080 | Porta do servidor HTTP |

## Desenvolvimento

### Workflow Recomendado

1. **Desenvolvimento Local:**
   ```bash
   # Usar in-memory (rápido, sem setup)
   go run cmd/api/main.go
   ```

2. **Desenvolvimento com Persistência:**
   ```bash
   # Usar PostgreSQL local
   make db-up
   make db-migrate-up
   DB_HOST=localhost go run cmd/api/main.go
   ```

3. **Testes de Integração:**
   ```bash
   # Resetar banco antes de testar
   make db-reset
   DB_HOST=localhost go test ./...
   ```

## Connection Pooling

O projeto usa **pgxpool** para gerenciar conexões:

```go
// Configurações do pool
MaxConns: 25                    // Máximo de conexões
MinConns: 5                     // Mínimo mantido
MaxConnLifetime: 1 hora         // Vida máxima da conexão
MaxConnIdleTime: 30 minutos     // Tempo máximo idle
HealthCheckPeriod: 1 minuto     // Intervalo de health check
```

### Benefícios:
- **Performance** - Reutiliza conexões
- **Escalabilidade** - Gerencia recursos automaticamente
- **Confiabilidade** - Reconexão automática em caso de falha

## Troubleshooting

### Erro: "connection refused"

```bash
# Verificar se PostgreSQL está rodando
docker-compose ps

# Ver logs do PostgreSQL
docker-compose logs postgres

# Reiniciar PostgreSQL
make db-down && make db-up
```

### Erro: "database does not exist"

```bash
# Criar database manualmente
docker-compose exec postgres psql -U expensetracker -c "CREATE DATABASE expenses;"
```

### Erro: "relation does not exist"

```bash
# Executar migrations
make db-migrate-up
```

### Resetar Completamente

```bash
# Parar tudo e remover volumes
docker-compose down -v

# Subir novamente
make db-up
make db-migrate-up
```

## Produção

### Recomendações

1. **Use SSL:**
   ```bash
   DB_SSLMODE=require
   ```

2. **Use Variáveis de Ambiente Seguras:**
   - Nunca commite senhas no código
   - Use secrets management (AWS Secrets Manager, Vault, etc)

3. **Configure Connection Pool:**
   ```go
   MaxConns: ajustar baseado na carga
   MinConns: manter conexões mínimas para baixa latência
   ```

4. **Monitore:**
   - Número de conexões ativas
   - Tempo de query
   - Erros de conexão

5. **Backup Regular:**
   ```bash
   pg_dump -U expensetracker expenses > backup.sql
   ```

## Próximos Passos

- [ ] Adicionar testes de integração com PostgreSQL
- [ ] Implementar migrations automáticas no startup (opcional)
- [ ] Adicionar índices para queries específicas
- [ ] Implementar soft delete (ao invés de DELETE físico)
- [ ] Adicionar auditoria (quem criou/atualizou)
- [ ] Implementar paginação em GetAll()

## Recursos

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [pgx - PostgreSQL Driver](https://github.com/jackc/pgx)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [Docker PostgreSQL](https://hub.docker.com/_/postgres)
