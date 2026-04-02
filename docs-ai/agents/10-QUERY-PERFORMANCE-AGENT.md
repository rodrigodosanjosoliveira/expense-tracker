---
status: VALIDADO
owner: Rodrigo
ultima_validacao: 2026-04-02
criterio_validacao: codigo+teste
fonte_de_verdade: codigo
escalonamento: PM/Rodrigo
cor: Amarelo
cor_hex: "#D97706"
---

# Query Performance Agent

## Contrato
Revisar queries PostgreSQL ineficientes e patterns de acesso a dados nos arquivos tocados.

## Entrega minima
- Arquivos inspecionados
- Achados e fixes aplicados
- Risco residual
- Resultado: `APROVADO` ou `RETORNA_DEV`

## Pontos de atencao especificos do Expense Tracker
- Queries sem filtro `user_id` (bug de seguranca E performance)
- Queries sem filtro de `category`, `date_from`, `date_to` quando disponiveis em `filters.go`
- Endpoints de listagem sem paginacao (retornar colecao unbounded)
- N+1 queries em loops (ex: buscar webhook e depois disparar em loop separado)
- Queries sem uso de indices disponiveis (ver migrations para indices criados)
- `pgx.Rows` nao fechado com `defer rows.Close()` (connection leak)
- Falta de `context.Context` propagation em queries longas
- Indices existentes: category, date, created_at em expenses; user_id em webhooks

## Gatilho
Apos Code-Reviewer, antes de QA, quando tocar queries SQL, repository patterns, ou endpoints com volume.

## Handoff
Se `APROVADO` → QA. Se `RETORNA_DEV` → Dev.
