---
status: VALIDADO
owner: Rodrigo
ultima_validacao: 2026-04-02
criterio_validacao: codigo+teste
fonte_de_verdade: codigo
escalonamento: PM/Rodrigo
---

# Start Here — Expense Tracker

## Objetivo
Reduzir tokens e tempo do prompt ao codigo com contexto confiavel.

## Ordem de leitura curta
1. Este arquivo
2. `01-INVARIANTES-GLOBAIS.md`
3. `03-MAPA-RAPIDO-MODULOS.md`
4. `agents/00-ORQUESTRACAO-AGENTS.md`

## Regra de demanda
- Toda demanda inicia pelo agent PM/Triage (`/triage`).
- Nenhuma linha de codigo e escrita antes do triage fechar escopo.
- DELIVERY_ID formato: `<MODULO>-<NNN>` (ex: `EXPENSE-001`, `AUTH-002`).

## Resolucao de conflito
- Codigo e fonte de verdade.
- Se doc diverge de codigo: marcar `ASSUNCAO` e escalar.
- Nao inventar regra de negocio.

## Fluxo de validacao de entrega
1. Preencher `docs-ai/deliveries/<ID>/report.md` a partir do template.
2. Rodar `scripts/ai/validate-delivery.sh <ID>`.
3. Se falhar, corrigir antes de abrir PR.

## Camadas canonicas
- `CLAUDE.md` — governanca raiz
- `.claude/agents/` — prompts executaveis
- `.claude/commands/` — slash commands
- `.claude/skills/` — skills reutilizaveis
- `docs-ai/` — docs operacionais + contratos + deliveries

## Stack do projeto
- **Go 1.25** — REST API com `net/http` stdlib
- **PostgreSQL 16** — banco de dados via pgx/v5
- **JWT** — autenticacao (golang-jwt/jwt/v5, 24h expiry)
- **bcrypt** — hash de senhas (golang.org/x/crypto)
- **Docker Compose** — ambiente local (sem CI/CD pipeline)
- **Arquitetura**: Layered (Handler -> Service -> Repository -> Domain)
