# Expense Tracker API

API REST para controle de despesas pessoais desenvolvida em Go usando TDD (Test-Driven Development).

## Arquitetura

O projeto segue uma arquitetura em camadas (layered architecture):

```
expense-tracker/
├── cmd/api/              # Ponto de entrada da aplicação
├── internal/
│   ├── domain/          # Entidades e regras de negócio
│   ├── repository/      # Camada de persistência
│   ├── service/         # Lógica de negócio
│   └── handler/         # HTTP handlers (controllers)
└── go.mod
```

## Conceitos Go Aprendidos

### 1. **Estrutura de Projeto**
- `internal/` - Pacotes privados (não podem ser importados por outros módulos)
- `cmd/` - Pontos de entrada (main packages)
- Convention over configuration

### 2. **Testing**
- **Table-driven tests** - Padrão para testar múltiplos cenários
- **Subtests com t.Run()** - Organizar e executar testes isoladamente
- **httptest** - Testar handlers HTTP sem servidor real
- **Mocking manual** - Criar mocks implementando interfaces

### 3. **Interfaces**
- Satisfação implícita (duck typing)
- Dependency injection via interfaces
- Facilita testes e desacoplamento

### 4. **Concorrência**
- **sync.RWMutex** - Proteção para acesso concorrente
- **Goroutines** - Funções executadas concorrentemente
- **Channels** - Comunicação entre goroutines
- **defer** - Executar função ao final do escopo

### 5. **HTTP e JSON**
- **net/http** - Biblioteca padrão para HTTP
- **encoding/json** - Encoder/Decoder para JSON
- **http.ServeMux** - Router padrão
- **http.ResponseWriter** e **http.Request** - Parâmetros dos handlers

### 6. **Context**
- Propagar timeouts, cancelamento e valores
- Padrão em todas as operações de I/O

### 7. **Graceful Shutdown**
- Desligar servidor graciosamente
- Aguardar requisições em andamento finalizarem

### 8. **Documentação com Swagger**
- **swaggo/swag** - Geração automática de documentação OpenAPI
- **Anotações** - Documentar API direto no código
- **Swagger UI** - Interface interativa para testar endpoints

## Documentação Interativa (Swagger)

A API possui documentação interativa gerada automaticamente com Swagger/OpenAPI.

### Acessar a Documentação

1. Inicie o servidor:
```bash
go run cmd/api/main.go
# ou
make run
```

2. Abra no navegador:
```
http://localhost:8080/swagger/index.html
```

### Gerar Documentação

Sempre que modificar os handlers ou adicionar novos endpoints, regenere a documentação:

```bash
# Gerar documentação Swagger
make swagger-gen

# Ou manualmente
swag init -g cmd/api/main.go -o docs
```

### Estrutura de Anotações

As anotações Swagger são adicionadas como comentários acima dos handlers:

```go
// @Summary Criar nova despesa
// @Description Cria uma nova despesa pessoal
// @Tags Expenses
// @Accept json
// @Produce json
// @Param expense body domain.Expense true "Dados da despesa"
// @Success 201 {object} domain.Expense
// @Failure 400 {string} string "Dados inválidos"
// @Router /expenses [post]
func (h *ExpenseHandler) CreateExpense(w http.ResponseWriter, r *http.Request) {
    // implementação...
}
```

## Endpoints da API

### Criar Despesa
```bash
POST /expenses
Content-Type: application/json

{
  "description": "Almoço",
  "amount": 25.50,
  "category": "Alimentação",
  "date": "2025-12-29T12:00:00Z"
}
```

### Listar Despesas
```bash
GET /expenses
```

### Buscar Despesa
```bash
GET /expenses/{id}
```

### Atualizar Despesa
```bash
PUT /expenses/{id}
Content-Type: application/json

{
  "description": "Jantar",
  "amount": 50.00,
  "category": "Alimentação",
  "date": "2025-12-29T19:00:00Z"
}
```

### Deletar Despesa
```bash
DELETE /expenses/{id}
```

### Health Check
```bash
GET /health
```

## Como Executar

### Rodar a aplicação
```bash
go run cmd/api/main.go
```

O servidor iniciará na porta 8080.

### Rodar todos os testes
```bash
go test -v ./...
```

### Rodar testes com coverage
```bash
go test -cover ./...
```

### Rodar testes de um pacote específico
```bash
go test -v ./internal/domain/
go test -v ./internal/repository/
go test -v ./internal/service/
go test -v ./internal/handler/
```

### Build da aplicação
```bash
go build -o bin/api cmd/api/main.go
```

Executar o binário:
```bash
./bin/api
```

## Próximos Passos

### 1. Melhorias na API
- [ ] Adicionar paginação em ListExpenses
- [ ] Filtros por categoria e data
- [ ] Ordenação customizável
- [ ] Middleware de logging
- [ ] Middleware de autenticação

### 2. Persistência
- [ ] Implementar repository com PostgreSQL
- [ ] Migrations com golang-migrate
- [ ] Connection pooling

### 3. Testes
- [ ] Aumentar cobertura de testes
- [ ] Testes de integração
- [ ] Testes de carga com vegeta ou k6

### 4. Deploy
- [ ] Dockerfile
- [ ] Docker Compose
- [ ] CI/CD com GitHub Actions
- [ ] Deploy em cloud (Fly.io, Railway, etc)

### 5. Observabilidade
- [ ] Structured logging com slog
- [ ] Métricas com Prometheus
- [ ] Tracing com OpenTelemetry

### 6. Recursos Avançados
- [ ] GraphQL API
- [ ] gRPC API
- [ ] Websockets para atualizações em tempo real
- [ ] Cache com Redis

## Dependências

- `github.com/google/uuid` - Geração de UUIDs

## Ecossistema Go

### Bibliotecas Populares

**Web Frameworks**
- `gin-gonic/gin` - Framework web rápido e minimalista
- `gofiber/fiber` - Framework inspirado no Express.js
- `labstack/echo` - Framework web de alta performance
- `gorilla/mux` - Router HTTP poderoso

**Banco de Dados**
- `database/sql` - Interface padrão para SQL
- `lib/pq` - Driver PostgreSQL
- `go-sql-driver/mysql` - Driver MySQL
- `jmoiron/sqlx` - Extensão do database/sql
- `gorm.io/gorm` - ORM completo

**Validação**
- `go-playground/validator` - Validação baseada em tags

**Logging**
- `log/slog` - Structured logging (Go 1.21+)
- `sirupsen/logrus` - Logger estruturado
- `uber-go/zap` - Logger de alta performance

**Testing**
- `stretchr/testify` - Assertions e mocks
- `golang/mock` - Framework de mocking

**Configuração**
- `spf13/viper` - Gerenciamento de configuração
- `joho/godotenv` - Carregar variáveis de ambiente

**CLI**
- `spf13/cobra` - Framework para CLIs
- `urfave/cli` - Framework CLI simples

## Comandos Go Úteis

```bash
# Inicializar módulo
go mod init github.com/usuario/projeto

# Adicionar dependência
go get github.com/pacote/nome

# Atualizar dependências
go mod tidy

# Verificar código
go vet ./...

# Formatar código
go fmt ./...

# Análise estática
go install golang.org/x/tools/cmd/staticcheck@latest
staticcheck ./...

# Ver documentação
go doc package.Function
```

## Padrões e Boas Práticas

1. **Nomes de pacotes** - Curtos, minúsculos, sem underscores
2. **Interfaces pequenas** - Preferir interfaces com 1-3 métodos
3. **Erros como valores** - Retornar erros explicitamente
4. **Constructor functions** - Usar `NewXxx()` para inicializar
5. **Defer para cleanup** - Usar defer para unlock, close, etc
6. **Accept interfaces, return structs** - Flexibilidade nas entradas
7. **Zero values úteis** - Structs devem ser utilizáveis sem inicialização

## Recursos para Aprender Mais

- [Tour of Go](https://go.dev/tour/) - Tutorial interativo oficial
- [Effective Go](https://go.dev/doc/effective_go) - Guia oficial de boas práticas
- [Go by Example](https://gobyexample.com/) - Exemplos práticos
- [Awesome Go](https://awesome-go.com/) - Lista curada de bibliotecas
- [Go Blog](https://go.dev/blog/) - Blog oficial do Go
