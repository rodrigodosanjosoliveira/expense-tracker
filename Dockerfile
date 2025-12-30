# Dockerfile multi-stage para otimização de tamanho

# Stage 1: Build
FROM golang:1.25-alpine AS builder

# Instalar dependências necessárias
RUN apk add --no-cache git

# Configurar diretório de trabalho
WORKDIR /app

# Copiar arquivos de dependências
COPY go.mod go.sum ./

# Download de dependências
RUN go mod download

# Copiar código fonte
COPY . .

# Build da aplicação
# CGO_ENABLED=0 para binário estático
# -ldflags="-w -s" para reduzir tamanho do binário
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/bin/api \
    cmd/api/main.go

# Stage 2: Runtime
FROM alpine:latest

# Instalar certificados CA para HTTPS
RUN apk --no-cache add ca-certificates

# Criar usuário não-root
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Configurar diretório de trabalho
WORKDIR /home/appuser

# Copiar binário do stage de build
COPY --from=builder /app/bin/api .

# Mudar ownership para usuário não-root
RUN chown -R appuser:appgroup /home/appuser

# Trocar para usuário não-root
USER appuser

# Expor porta
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Comando de execução
CMD ["./api"]
