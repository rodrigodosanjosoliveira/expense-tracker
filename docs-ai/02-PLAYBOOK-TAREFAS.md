---
status: VALIDADO
owner: Rodrigo
ultima_validacao: 2026-04-02
criterio_validacao: codigo+teste
fonte_de_verdade: codigo
escalonamento: PM/Rodrigo
---

# Playbook de Tarefas — Expense Tracker

## Feature

1. Identificar modulo impactado (consultar `03-MAPA-RAPIDO-MODULOS.md`).
2. Definir arquivos provaveis, riscos e invariantes aplicaveis.
3. Implementar com testes cobrindo aceite.
4. Atualizar docs do modulo no mesmo PR.

## Bugfix

1. Reproduzir o problema (teste ou evidencia manual).
2. Aplicar fix minimo no codigo.
3. Criar teste de regressao.
4. Documentar causa raiz no report da delivery.

## Melhoria tecnica

1. Definir objetivo e metricas de sucesso.
2. Garantir equivalencia funcional (nao quebrar comportamento existente).
3. Atualizar docs se a melhoria muda convencoes ou patterns.

## Quando criar testes de integracao

- Quando o fluxo cruza camadas (handler -> service -> repository -> PostgreSQL).
- Quando o fluxo envolve autenticacao JWT (middleware, claims, user_id extraction).
- Quando o fluxo envolve webhooks e notificacoes (dispatch de eventos).
- Quando o fluxo envolve migrations de banco (schema evolution).

## Regras de implementacao

- Nunca pular o fluxo de agents — iniciar sempre pelo `/triage`.
- Toda mudanca em `internal/domain/` e considerada de alto risco.
- Toda mudanca em `internal/middleware/auth.go` dispara `Security` review.
- Toda migration nova dispara `Migration-Reviewer` e `Release/Ops`.
- `user_id` scoping deve ser validado em toda mudanca de repository.
