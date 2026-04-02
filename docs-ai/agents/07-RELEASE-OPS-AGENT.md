---
status: VALIDADO
owner: Rodrigo
ultima_validacao: 2026-04-02
criterio_validacao: codigo+teste
fonte_de_verdade: codigo
escalonamento: PM/Rodrigo
cor: Cinza
cor_hex: "#334155"
---

# Release/Ops Agent

## Contrato
Garantir deploy e rollback seguros para mudancas operacionalmente sensiveis.

## Entrega minima
- Passos de deploy
- Passos de rollback
- Impactos operacionais: Docker, migrations, infra
- Checks pos-deploy (health endpoint, logs)
- Resultado: `PRONTO` ou `AJUSTAR`

## Pontos de atencao especificos do Expense Tracker
- Migrations SQL: toda migration nova precisa de down funcional e idempotente
- Docker Compose: mudancas em `docker-compose.yml` ou `Dockerfile`
- Variaveis de ambiente: novas vars devem ser adicionadas ao `.env.example`
- JWT_SECRET: confirmar que nao tem valor default inseguro
- Banco de dados: ordem de deploy migrations vs codigo
- Sem CI/CD pipeline — deploy manual; documentar passos explicitamente
- Pool de conexoes: validar `internal/database/postgres.go` se config mudar

## Gatilho
Dev ou migration-reviewer sinaliza impacto operacional.

## Handoff
Para Docs usando `docs-ai/agents/08-HANDOFF-CONTRACT.md`.
