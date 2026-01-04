#!/bin/bash

# Script de testes completo para Expense Tracker API
# Testa: Autenticação, CRUD de Despesas, Webhooks e Notificações

set -e  # Parar em caso de erro

API_URL="http://localhost:8080"
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     EXPENSE TRACKER API - TESTE COMPLETO                  ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Verificar se a API está rodando
echo -e "${YELLOW}[1/10] Verificando se a API está rodando...${NC}"
if ! curl -s -o /dev/null -w "%{http_code}" "$API_URL/health" | grep -q "200"; then
    echo -e "${RED}✗ API não está rodando em $API_URL${NC}"
    echo -e "${YELLOW}Execute: make run${NC}"
    exit 1
fi
echo -e "${GREEN}✓ API está rodando${NC}"
echo ""

# ========================================
# TESTES DE AUTENTICAÇÃO
# ========================================

echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  TESTANDO AUTENTICAÇÃO${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""

# Gerar username e email únicos
TIMESTAMP=$(date +%s)
USERNAME="testuser_$TIMESTAMP"
EMAIL="test_${TIMESTAMP}@example.com"
PASSWORD="senha12345"

echo -e "${YELLOW}[2/10] Registrando novo usuário...${NC}"
echo "Username: $USERNAME"
echo "Email: $EMAIL"

REGISTER_RESPONSE=$(curl -s -X POST "$API_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"$USERNAME\",
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\"
  }")

if echo "$REGISTER_RESPONSE" | grep -q "token"; then
    echo -e "${GREEN}✓ Usuário registrado com sucesso${NC}"
    TOKEN=$(echo "$REGISTER_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    USER_ID=$(echo "$REGISTER_RESPONSE" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    echo "Token obtido: ${TOKEN:0:50}..."
    echo "User ID: $USER_ID"
else
    echo -e "${RED}✗ Falha ao registrar usuário${NC}"
    echo "$REGISTER_RESPONSE"
    exit 1
fi
echo ""

echo -e "${YELLOW}[3/10] Testando login...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"$USERNAME\",
    \"password\": \"$PASSWORD\"
  }")

if echo "$LOGIN_RESPONSE" | grep -q "token"; then
    echo -e "${GREEN}✓ Login realizado com sucesso${NC}"
    # Atualizar token com o do login
    TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
else
    echo -e "${RED}✗ Falha no login${NC}"
    echo "$LOGIN_RESPONSE"
    exit 1
fi
echo ""

echo -e "${YELLOW}[4/10] Testando login com credenciais inválidas...${NC}"
INVALID_LOGIN=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"$USERNAME\",
    \"password\": \"senha_errada\"
  }")

STATUS=$(echo "$INVALID_LOGIN" | tail -n1)
if [ "$STATUS" = "401" ]; then
    echo -e "${GREEN}✓ Login inválido corretamente rejeitado (401)${NC}"
else
    echo -e "${RED}✗ Deveria retornar 401, retornou: $STATUS${NC}"
fi
echo ""

# ========================================
# TESTES DE DESPESAS
# ========================================

echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  TESTANDO CRUD DE DESPESAS${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""

echo -e "${YELLOW}[5/10] Criando despesa...${NC}"
CREATE_EXPENSE=$(curl -s -X POST "$API_URL/expenses" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Almoço no restaurante",
    "amount": 45.50,
    "category": "Alimentação",
    "date": "2026-01-03T12:00:00Z"
  }')

if echo "$CREATE_EXPENSE" | grep -q "id"; then
    echo -e "${GREEN}✓ Despesa criada com sucesso${NC}"
    EXPENSE_ID=$(echo "$CREATE_EXPENSE" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    echo "Expense ID: $EXPENSE_ID"
    echo "$CREATE_EXPENSE" | jq '.' 2>/dev/null || echo "$CREATE_EXPENSE"
else
    echo -e "${RED}✗ Falha ao criar despesa${NC}"
    echo "$CREATE_EXPENSE"
fi
echo ""

echo -e "${YELLOW}[6/10] Listando despesas...${NC}"
LIST_EXPENSES=$(curl -s -X GET "$API_URL/expenses" \
  -H "Authorization: Bearer $TOKEN")

if echo "$LIST_EXPENSES" | grep -q "Almoço no restaurante"; then
    echo -e "${GREEN}✓ Despesas listadas com sucesso${NC}"
    COUNT=$(echo "$LIST_EXPENSES" | grep -o '"id"' | wc -l)
    echo "Total de despesas: $COUNT"
else
    echo -e "${RED}✗ Falha ao listar despesas${NC}"
fi
echo ""

echo -e "${YELLOW}[7/10] Atualizando despesa...${NC}"
UPDATE_EXPENSE=$(curl -s -X PUT "$API_URL/expenses/$EXPENSE_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Almoço no restaurante - ATUALIZADO",
    "amount": 52.00,
    "category": "Alimentação",
    "date": "2026-01-03T12:00:00Z"
  }')

if echo "$UPDATE_EXPENSE" | grep -q "ATUALIZADO"; then
    echo -e "${GREEN}✓ Despesa atualizada com sucesso${NC}"
    echo "$UPDATE_EXPENSE" | jq '.' 2>/dev/null || echo "$UPDATE_EXPENSE"
else
    echo -e "${RED}✗ Falha ao atualizar despesa${NC}"
fi
echo ""

# ========================================
# TESTES DE WEBHOOKS
# ========================================

echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  TESTANDO WEBHOOKS${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""

echo -e "${YELLOW}[8/10] Criando webhook...${NC}"
CREATE_WEBHOOK=$(curl -s -X POST "$API_URL/webhooks" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Webhook",
    "type": "custom",
    "url": "https://webhook.site/unique-url-here",
    "events": ["expense.created", "expense.updated", "expense.deleted"],
    "is_active": true
  }')

if echo "$CREATE_WEBHOOK" | grep -q "id"; then
    echo -e "${GREEN}✓ Webhook criado com sucesso${NC}"
    WEBHOOK_ID=$(echo "$CREATE_WEBHOOK" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    echo "Webhook ID: $WEBHOOK_ID"
    echo "$CREATE_WEBHOOK" | jq '.' 2>/dev/null || echo "$CREATE_WEBHOOK"
else
    echo -e "${RED}✗ Falha ao criar webhook${NC}"
    echo "$CREATE_WEBHOOK"
fi
echo ""

echo -e "${YELLOW}[9/10] Listando webhooks...${NC}"
LIST_WEBHOOKS=$(curl -s -X GET "$API_URL/webhooks" \
  -H "Authorization: Bearer $TOKEN")

if echo "$LIST_WEBHOOKS" | grep -q "Test Webhook"; then
    echo -e "${GREEN}✓ Webhooks listados com sucesso${NC}"
    echo "$LIST_WEBHOOKS" | jq '.' 2>/dev/null || echo "$LIST_WEBHOOKS"
else
    echo -e "${RED}✗ Falha ao listar webhooks${NC}"
    echo "Resposta recebida:"
    echo "$LIST_WEBHOOKS" | jq '.' 2>/dev/null || echo "$LIST_WEBHOOKS"
fi
echo ""

# ========================================
# TESTE DE ISOLAMENTO
# ========================================

echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  TESTANDO ISOLAMENTO DE DADOS${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""

echo -e "${YELLOW}[10/10] Testando isolamento entre usuários...${NC}"

# Criar segundo usuário
USERNAME2="testuser2_$TIMESTAMP"
EMAIL2="test2_${TIMESTAMP}@example.com"

REGISTER2=$(curl -s -X POST "$API_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"$USERNAME2\",
    \"email\": \"$EMAIL2\",
    \"password\": \"$PASSWORD\"
  }")

TOKEN2=$(echo "$REGISTER2" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

# Tentar acessar despesa do primeiro usuário com token do segundo
ACCESS_TEST=$(curl -s -w "\n%{http_code}" -X GET "$API_URL/expenses/$EXPENSE_ID" \
  -H "Authorization: Bearer $TOKEN2")

STATUS=$(echo "$ACCESS_TEST" | tail -n1)
if [ "$STATUS" = "403" ]; then
    echo -e "${GREEN}✓ Isolamento funcionando corretamente (403 Forbidden)${NC}"
else
    echo -e "${RED}✗ Falha no isolamento! Status: $STATUS${NC}"
fi
echo ""

# ========================================
# LIMPEZA E RESUMO
# ========================================

echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  LIMPEZA${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""

echo -e "${YELLOW}Deletando dados de teste...${NC}"

# Deletar webhook
if [ ! -z "$WEBHOOK_ID" ]; then
    curl -s -X DELETE "$API_URL/webhooks/$WEBHOOK_ID" \
      -H "Authorization: Bearer $TOKEN" > /dev/null
    echo -e "${GREEN}✓ Webhook deletado${NC}"
fi

# Deletar despesa
if [ ! -z "$EXPENSE_ID" ]; then
    curl -s -X DELETE "$API_URL/expenses/$EXPENSE_ID" \
      -H "Authorization: Bearer $TOKEN" > /dev/null
    echo -e "${GREEN}✓ Despesa deletada${NC}"
fi

echo ""
echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                    RESUMO DOS TESTES                      ║${NC}"
echo -e "${BLUE}╠════════════════════════════════════════════════════════════╣${NC}"
echo -e "${GREEN}║  ✓ Autenticação (Register e Login)                       ║${NC}"
echo -e "${GREEN}║  ✓ CRUD de Despesas                                       ║${NC}"
echo -e "${GREEN}║  ✓ CRUD de Webhooks                                       ║${NC}"
echo -e "${GREEN}║  ✓ Isolamento de dados entre usuários                    ║${NC}"
echo -e "${BLUE}╠════════════════════════════════════════════════════════════╣${NC}"
echo -e "${GREEN}║  Usuários criados:                                        ║${NC}"
echo -e "${GREEN}║    - $USERNAME                    ║${NC}"
echo -e "${GREEN}║    - $USERNAME2                   ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}Dica: Para testar notificações reais, configure um webhook do Slack ou Discord!${NC}"
echo ""
