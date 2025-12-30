.PHONY: help build run test test-verbose test-coverage clean dev swagger-gen swagger-install db-up db-down db-migrate-up db-migrate-down db-create-migration

# Variáveis
APP_NAME=expense-tracker
BINARY=bin/api
MAIN=cmd/api/main.go
GOBIN=$(shell go env GOPATH)/bin
AIR=$(GOBIN)/air
SWAG=$(GOBIN)/swag
MIGRATE=$(GOBIN)/migrate

# Database
DB_URL=postgres://expensetracker:secret@localhost:5432/expenses?sslmode=disable
MIGRATIONS_PATH=migrations

help: ## Exibir esta mensagem de ajuda
	@echo "Comandos disponíveis:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Compilar a aplicação
	@echo "Compilando..."
	@go build -o $(BINARY) $(MAIN)
	@echo "Binário gerado em $(BINARY)"

run: ## Executar a aplicação
	@go run $(MAIN)

dev: ## Executar em modo desenvolvimento (rebuild automático)
	@test -f $(AIR) || (echo "Instalando air..." && go install github.com/air-verse/air@latest)
	@$(AIR)

test: ## Rodar todos os testes
	@go test ./...

test-verbose: ## Rodar testes com saída detalhada
	@go test -v ./...

test-coverage: ## Rodar testes com cobertura
	@go test -cover ./...

test-coverage-html: ## Gerar relatório HTML de cobertura
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Relatório gerado em coverage.html"

test-api: build ## Testar API com script
	@./test-api.sh

clean: ## Limpar arquivos gerados
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Limpeza concluída"

fmt: ## Formatar código
	@go fmt ./...

vet: ## Analisar código
	@go vet ./...

lint: ## Executar linter
	@which golangci-lint > /dev/null || (echo "Instale golangci-lint: https://golangci-lint.run/welcome/install/" && exit 1)
	@golangci-lint run

tidy: ## Limpar dependências
	@go mod tidy

install-tools: ## Instalar ferramentas de desenvolvimento
	@echo "Instalando ferramentas..."
	@go install github.com/air-verse/air@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "air instalado (hot reload)"
	@echo "swag instalado (documentação Swagger)"

deps: ## Baixar dependências
	@go mod download

update-deps: ## Atualizar dependências
	@go get -u ./...
	@go mod tidy

docker-build: ## Construir imagem Docker
	@docker build -t $(APP_NAME):latest .

docker-run: ## Executar container Docker
	@docker run -p 8080:8080 $(APP_NAME):latest

swagger-install: ## Instalar Swagger CLI
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Swagger instalado com sucesso!"

swagger-gen: ## Gerar documentação Swagger
	@test -f $(SWAG) || (echo "Instalando swag..." && go install github.com/swaggo/swag/cmd/swag@latest)
	@echo "Gerando documentação Swagger..."
	@$(SWAG) init -g cmd/api/main.go -o docs
	@echo "Documentação gerada em docs/"
	@echo "Acesse: http://localhost:8080/swagger/index.html"

# Comandos de Banco de Dados
db-up: ## Iniciar PostgreSQL com Docker Compose
	@docker-compose up -d postgres
	@echo "Aguardando PostgreSQL iniciar..."
	@sleep 3
	@echo "✓ PostgreSQL rodando em localhost:5432"

db-down: ## Parar PostgreSQL
	@docker-compose down
	@echo "✓ PostgreSQL parado"

db-migrate-up: ## Executar migrations (criar tabelas)
	@test -f $(MIGRATE) || (echo "Instalando migrate..." && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
	@echo "Executando migrations..."
	@$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up
	@echo "✓ Migrations executadas com sucesso"

db-migrate-down: ## Reverter última migration
	@test -f $(MIGRATE) || (echo "Instalando migrate..." && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
	@echo "Revertendo migration..."
	@$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down 1
	@echo "✓ Migration revertida"

db-create-migration: ## Criar nova migration (uso: make db-create-migration NAME=nome_da_migration)
	@test -f $(MIGRATE) || (echo "Instalando migrate..." && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
	@if [ -z "$(NAME)" ]; then echo "Erro: Especifique NAME=nome_da_migration"; exit 1; fi
	@$(MIGRATE) create -ext sql -dir $(MIGRATIONS_PATH) -seq $(NAME)
	@echo "✓ Migration criada em $(MIGRATIONS_PATH)/"

db-reset: db-down db-up db-migrate-up ## Resetar banco de dados (down + up + migrate)

all: clean test build ## Limpar, testar e compilar
