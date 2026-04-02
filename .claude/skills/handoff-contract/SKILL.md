---
name: handoff-contract
description: Generate a standardized handoff contract between pipeline agents following the Expense Tracker agent orchestration format. Use when transitioning between agents (e.g., triage→guardrails, dev→code-review, qa→docs).
---

# Handoff Contract Generator

Generates handoff documents between agents in the Expense Tracker delivery pipeline.

## Usage

When invoked, read the handoff contract template at `docs-ai/agents/08-HANDOFF-CONTRACT.md` and fill in all required fields based on the current delivery context.

## Required Fields

1. **Source Agent** — which agent is handing off (e.g., `pm-triage`, `dev-implementation`)
2. **Target Agent** — which agent receives (e.g., `architecture-guardrails`, `code-reviewer`)
3. **DELIVERY_ID** — format `<MODULE>-<NNN>` (e.g., `EXPENSE-001`, `AUTH-002`)
4. **Summary** — what was done in the current phase
5. **Acceptance Criteria Status** — checklist with pass/fail/pending per criterion
6. **Residual Risks** — anything the next agent should watch for
7. **Files Changed** — list of files touched in this phase
8. **Routing Signals** — flags for conditional agents:
   - `SECURITY` if auth, JWT, bcrypt, secrets, or external integrations were touched
   - `RELEASE_OPS` if Docker, docker-compose, Dockerfile, or deploy were touched
   - `MIGRATION` if SQL schema, migrations, or domain entities were changed
   - `QUERY_PERF` if PostgreSQL queries or repository patterns were changed

## Rules

- Never skip required fields — leave explicit `TODO` markers if information is unavailable.
- Use the exact format from `docs-ai/agents/08-HANDOFF-CONTRACT.md`.
- Routing signals must be explicit — never assume the next agent will discover them.
