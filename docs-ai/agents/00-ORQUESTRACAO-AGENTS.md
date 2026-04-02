---
status: VALIDADO
owner: Rodrigo
ultima_validacao: 2026-04-02
criterio_validacao: codigo+teste
fonte_de_verdade: codigo
escalonamento: PM/Rodrigo
cor: Preto
cor_hex: "#111827"
---

# Orquestracao de Agents — Expense Tracker

## Canonicidade

- Prompts executaveis vivem em `.claude/agents/`.
- Este diretorio `docs-ai/agents/` guarda contrato, gatilho, orquestracao e validacao.
- Commands devem ficar curtos e apontar para agents ou skills.

Objetivo: transformar pedido em entrega com menos retrabalho e menos token.

## Sequencia recomendada

1. [PM/Triage Agent](01-PM-TRIAGE-AGENT.md) — **obrigatorio**
2. [Architecture/Guardrails Agent](02-ARCHITECTURE-GUARDRAILS-AGENT.md) — **obrigatorio**
3. [Data-Model Agent] — **obrigatorio quando tocar domain entities, repository interfaces ou migrations SQL**
4. [Test-Writer Agent] — **obrigatorio antes ou junto com Dev em toda delivery testavel**
5. [Dev Agent](03-DEV-AGENT.md) — **obrigatorio**
6. [Code-Reviewer Agent] — **obrigatorio apos Dev, antes do QA**
7. [Query-Performance Agent](10-QUERY-PERFORMANCE-AGENT.md) — **condicional: apos Code-Reviewer, antes do QA, quando tocar queries SQL, repository patterns, ou endpoints com volume de dados**
8. [QA Agent](04-QA-AGENT.md) — **obrigatorio**
9. [Docs Agent](05-DOCS-AGENT.md) — **obrigatorio**
10. [Security Agent](06-SECURITY-AGENT.md) — obrigatorio quando tocar auth, JWT, bcrypt, middleware, segredos, integracao externa
11. [Release/Ops Agent](07-RELEASE-OPS-AGENT.md) — obrigatorio quando tocar Docker, docker-compose, Dockerfile, migrations de deploy
12. [Migration-Reviewer Agent] — obrigatorio junto com Release/Ops quando houver mudanca de schema SQL
13. [Git Delivery Agent](11-GIT-DELIVERY-AGENT.md) — **terminal: apos Docs ou Release/Ops, para versionar e abrir o PR**

## Paleta oficial dos agents

- PM/Triage Agent: Azul (`#1D4ED8`)
- Architecture/Guardrails Agent: Ciano (`#0E7490`)
- Data-Model Agent: Teal (`#0F766E`)
- Test-Writer Agent: Roxo (`#7C3AED`)
- Dev Agent: Verde (`#15803D`)
- Code-Reviewer Agent: Lima (`#65A30D`)
- Query-Performance Agent: Amarelo (`#D97706`)
- QA Agent: Laranja (`#C2410C`)
- Docs Agent: Indigo (`#4338CA`)
- Security Agent: Vermelho (`#B91C1C`)
- Release/Ops Agent: Cinza (`#334155`)
- Migration-Reviewer Agent: Cinza Escuro (`#1E293B`)
- Git Delivery Agent: Neutro (`#6B7280`)

## Regra de acionamento

- Use sempre PM -> Architecture -> Dev -> Code-Reviewer -> QA -> Docs.
- Acione Data-Model quando tocar entity, document, domain model, partition key ou schema.
- Acione Test-Writer antes ou junto com Dev em toda delivery com criterio de aceite testavel.
- Acione Code-Reviewer apos Dev e antes de QA em toda delivery.
- Acione Query-Performance apos Code-Reviewer e antes de QA quando tocar queries SQL, repository patterns, ou endpoints com volume de dados.
- Acione Security quando tocar auth, JWT, bcrypt, middleware, segredos, integracao externa.
- Acione Release/Ops quando tocar Docker, docker-compose, Dockerfile, migrations de deploy, infra.
- Acione Migration-Reviewer junto com Release/Ops quando houver mudanca de schema SQL.
- Acione Git Delivery ao final do ciclo para versionar e abrir o MR.

## Contrato unico de handoff

Todos os agents devem usar o formato de [Handoff Contract](08-HANDOFF-CONTRACT.md).
Validacao automatica da entrega: [Validacao de Entrega](09-VALIDACAO-ENTREGA.md).

## Fonte de verdade

- Em conflito entre doc e implementacao, prevalece codigo.
- Quando isso acontecer, o agente deve marcar `ASSUNCAO` e escalar para PM/Rodrigo.
