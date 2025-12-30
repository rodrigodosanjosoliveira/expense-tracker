#!/bin/bash

# Script para testar a API de Despesas

BASE_URL="http://localhost:8080"

echo "======================================"
echo "Testando API de Controle de Despesas"
echo "======================================"
echo ""

# 1. Health Check
echo "1. Health Check"
curl -s "$BASE_URL/health"
echo -e "\n"

# 2. Criar primeira despesa
echo "2. Criar despesa - Almoço"
EXPENSE1=$(curl -s -X POST "$BASE_URL/expenses" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Almoço no restaurante",
    "amount": 45.50,
    "category": "Alimentação",
    "date": "2025-12-29T12:00:00Z"
  }')
echo "$EXPENSE1" | jq '.'
ID1=$(echo "$EXPENSE1" | jq -r '.id')
echo "ID criado: $ID1"
echo ""

# 3. Criar segunda despesa
echo "3. Criar despesa - Uber"
EXPENSE2=$(curl -s -X POST "$BASE_URL/expenses" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Corrida de Uber",
    "amount": 25.00,
    "category": "Transporte",
    "date": "2025-12-29T08:30:00Z"
  }')
echo "$EXPENSE2" | jq '.'
ID2=$(echo "$EXPENSE2" | jq -r '.id')
echo "ID criado: $ID2"
echo ""

# 4. Criar terceira despesa
echo "4. Criar despesa - Cinema"
EXPENSE3=$(curl -s -X POST "$BASE_URL/expenses" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Ingresso de cinema",
    "amount": 35.00,
    "category": "Entretenimento",
    "date": "2025-12-29T20:00:00Z"
  }')
echo "$EXPENSE3" | jq '.'
ID3=$(echo "$EXPENSE3" | jq -r '.id')
echo "ID criado: $ID3"
echo ""

# 5. Listar todas as despesas
echo "5. Listar todas as despesas"
curl -s "$BASE_URL/expenses" | jq '.'
echo ""

# 6. Buscar despesa específica
echo "6. Buscar despesa específica (ID: $ID1)"
curl -s "$BASE_URL/expenses/$ID1" | jq '.'
echo ""

# 7. Atualizar despesa
echo "7. Atualizar despesa (ID: $ID1) - aumentar valor"
curl -s -X PUT "$BASE_URL/expenses/$ID1" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Almoço no restaurante (com sobremesa)",
    "amount": 55.00,
    "category": "Alimentação",
    "date": "2025-12-29T12:00:00Z"
  }' | jq '.'
echo ""

# 8. Verificar atualização
echo "8. Verificar atualização"
curl -s "$BASE_URL/expenses/$ID1" | jq '.'
echo ""

# 9. Deletar despesa
echo "9. Deletar despesa (ID: $ID2)"
curl -s -X DELETE "$BASE_URL/expenses/$ID2" -w "\nStatus Code: %{http_code}\n"
echo ""

# 10. Listar despesas após deleção
echo "10. Listar despesas após deleção"
curl -s "$BASE_URL/expenses" | jq '.'
echo ""

# 11. Tentar buscar despesa deletada
echo "11. Tentar buscar despesa deletada (deve retornar 404)"
curl -s "$BASE_URL/expenses/$ID2" -w "\nStatus Code: %{http_code}\n"
echo ""

# 12. Teste de validação - descrição vazia
echo "12. Teste de validação - descrição vazia (deve falhar)"
curl -s -X POST "$BASE_URL/expenses" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "",
    "amount": 10.00,
    "category": "Test"
  }' -w "\nStatus Code: %{http_code}\n"
echo ""

# 13. Teste de validação - valor zero
echo "13. Teste de validação - valor zero (deve falhar)"
curl -s -X POST "$BASE_URL/expenses" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Test",
    "amount": 0,
    "category": "Test"
  }' -w "\nStatus Code: %{http_code}\n"
echo ""

echo "======================================"
echo "Testes concluídos!"
echo "======================================"
