---
name: git-delivery
description: Create branch, atomic commits, and Pull Request following the Expense Tracker versioning standards. Use at the end of the delivery pipeline to package and ship the work.
---

# Git Delivery

Handles branching, committing, and PR creation for Expense Tracker deliveries.

## Branch Naming

Format: `task/<delivery-id>`

Examples:
- `task/EXPENSE-001`
- `task/AUTH-002`
- `task/INFRA-003`

## Commit Convention

Format: `<type>(<scope>): <description> [<DELIVERY_ID>]`

Scopes by module:
- Expense: `expense`
- Auth: `auth`
- Webhook: `webhook`
- Notification: `notification`
- Domain: `domain`
- Infrastructure: `infra`, `config`, `migrations`
- Tests: `test`
- Docs: `docs`

Examples:
- `feat(expense): adiciona filtro por categoria [EXPENSE-001]`
- `fix(auth): corrige validacao de JWT expirado [AUTH-002]`
- `feat(migrations): adiciona tabela de categorias [INFRA-003]`

## Pull Request

Use `gh` (GitHub CLI) to create PRs:

```bash
gh pr create --title "<type>(<scope>): <description>" --body "$(cat <<'EOF'
## Contexto
<resumo da delivery>

## Criterios de aceite
<lista com status>

## Arquivos alterados
<lista>

## Como testar
go test -v ./...

## Report
docs-ai/deliveries/<DELIVERY_ID>/report.md

## Riscos
<lista>
EOF
)"
```

## Rules

- One PR per delivery (atomic).
- All commits in the branch should relate to the same DELIVERY_ID.
- Run `scripts/ai/validate-delivery.sh <DELIVERY_ID>` before creating the PR.
- Never force-push to main.
