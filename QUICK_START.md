# Guia Rápido - Expense Tracker API

## Início Rápido (5 minutos)

### 1. Rodar a aplicação

```bash
# Opção 1: Rodar diretamente
go run cmd/api/main.go

# Opção 2: Compilar e executar
make build
./bin/api

# Opção 3: Docker
docker-compose up
```

A API estará disponível em `http://localhost:8080`

**Documentação Swagger:** `http://localhost:8080/swagger/index.html`

### 2. Testar a API

**Opção 1: Usar o Swagger UI (Recomendado)**

Abra no navegador: `http://localhost:8080/swagger/index.html`

- Interface interativa para testar todos os endpoints
- Documentação completa com exemplos
- Teste diretamente pelo navegador

**Opção 2: Usar curl**

**Criar uma despesa:**
```bash
curl -X POST http://localhost:8080/expenses \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Almoço",
    "amount": 25.50,
    "category": "Alimentação",
    "date": "2025-12-29T12:00:00Z"
  }'
```

**Listar todas as despesas:**
```bash
curl http://localhost:8080/expenses
```

**Buscar uma despesa específica:**
```bash
curl http://localhost:8080/expenses/{id}
```

**Atualizar uma despesa:**
```bash
curl -X PUT http://localhost:8080/expenses/{id} \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Jantar",
    "amount": 50.00,
    "category": "Alimentação",
    "date": "2025-12-29T19:00:00Z"
  }'
```

**Deletar uma despesa:**
```bash
curl -X DELETE http://localhost:8080/expenses/{id}
```

### 3. Rodar testes

```bash
# Todos os testes
make test

# Testes com cobertura
make test-coverage

# Testes detalhados
make test-verbose
```

### 4. Script de teste automático

```bash
# Execute em um terminal
go run cmd/api/main.go

# Em outro terminal, rode o script de teste
./test-api.sh
```

## Comandos Make Úteis

```bash
make help            # Ver todos os comandos
make build           # Compilar
make run             # Executar
make test            # Rodar testes
make test-coverage   # Testes com cobertura
make clean           # Limpar arquivos gerados
make fmt             # Formatar código
make vet             # Análise estática
make dev             # Modo desenvolvimento (hot reload)
```

## Desenvolvimento com Hot Reload

```bash
# Instalar air (primeira vez)
make install-tools

# Rodar com hot reload
make dev
```

Agora qualquer mudança no código reiniciará automaticamente o servidor!

## Estrutura do Projeto

```
expense-tracker/
├── cmd/api/              # Ponto de entrada
│   └── main.go
├── internal/
│   ├── domain/          # Entidades (Expense)
│   ├── repository/      # Persistência (Memory)
│   ├── service/         # Lógica de negócio
│   └── handler/         # HTTP handlers
├── bin/                 # Binários compilados
├── Makefile            # Comandos úteis
├── Dockerfile          # Container Docker
└── docker-compose.yml  # Orquestração
```

## Testando com Postman/Insomnia

Importe esta coleção básica:

**Base URL:** `http://localhost:8080`

**Endpoints:**
- `POST /expenses` - Criar despesa
- `GET /expenses` - Listar todas
- `GET /expenses/{id}` - Buscar por ID
- `PUT /expenses/{id}` - Atualizar
- `DELETE /expenses/{id}` - Deletar
- `GET /health` - Health check

**Body de exemplo (POST/PUT):**
```json
{
  "description": "Café da manhã",
  "amount": 15.00,
  "category": "Alimentação",
  "date": "2025-12-29T08:00:00Z"
}
```

## Próximos Passos

1. **Adicionar banco de dados:**
   - Implementar PostgreSQL repository
   - Migrations com golang-migrate

2. **Melhorar a API:**
   - Paginação
   - Filtros (categoria, data)
   - Ordenação

3. **Adicionar autenticação:**
   - JWT
   - Middleware de auth

4. **Observabilidade:**
   - Logging estruturado
   - Métricas
   - Tracing

## Dicas TDD

1. **Red-Green-Refactor:**
   - Red: Escreva o teste (que falha)
   - Green: Implemente o código mínimo
   - Refactor: Melhore o código

2. **Execute testes frequentemente:**
   ```bash
   # Terminal com watch
   watch -n 1 go test ./...
   ```

3. **Testes primeiro, sempre:**
   - Nunca implemente sem antes ter o teste
   - Use table-driven tests
   - Mock de dependências

## Problemas Comuns

**Porta 8080 ocupada:**
```bash
# Linux/Mac
lsof -ti:8080 | xargs kill -9

# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F
```

**Dependências não encontradas:**
```bash
go mod tidy
go mod download
```

**Testes falhando:**
```bash
go clean -testcache
go test -v ./...
```

## Recursos de Aprendizado

- [A Tour of Go](https://go.dev/tour/) - Tutorial interativo
- [Effective Go](https://go.dev/doc/effective_go) - Boas práticas
- [Go by Example](https://gobyexample.com/) - Exemplos práticos
- [Learn Go with Tests](https://quii.gitbook.io/learn-go-with-tests/) - TDD em Go

## Contribuindo

1. Fork o projeto
2. Crie uma branch (`git checkout -b feature/nova-funcionalidade`)
3. Commit suas mudanças (`git commit -am 'Adiciona nova funcionalidade'`)
4. Push para a branch (`git push origin feature/nova-funcionalidade`)
5. Abra um Pull Request

## Licença

MIT
