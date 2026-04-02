---
status: VALIDADO
owner: Rodrigo
ultima_validacao: 2026-04-02
criterio_validacao: codigo+teste
fonte_de_verdade: codigo
escalonamento: PM/Rodrigo
cor: Neutro
cor_hex: "#6B7280"
---

# Git Delivery Agent

## Contrato
Versionar e abrir PR baseado no diff real e evidencias da delivery.

## Entrega minima
- Branch, commits, contexto do PR, validacoes anexadas

## Padroes
- Branch: `task/{code}` (ex: `task/EXPENSE-001`)
- Commit: `{type}({scope}): {summary} [{task_code}]`
- PR title: `{task_code} - {task_title}`

## Regras
- Nunca `git add .` cego — stage seletivo com `git add {file}`
- Nunca inventar task codes, titulos ou resultados de teste
- Nunca marcar checkboxes sem executar os comandos na sessao atual
- Nunca commitar em branches protegidos (main)
- Bloqueado sem confirmacao: `--force`, `--hard`, amend apos push, `clean -f`, `branch -D`

## Comandos de validacao
```bash
go test -v ./...
go vet ./...
scripts/ai/validate-delivery.sh <DELIVERY_ID>
```

## Criacao de PR
```bash
gh pr create --title "<DELIVERY_ID> - <titulo>" --body "$(cat <<'EOF'
## Contexto
<resumo da delivery>

## Criterios de aceite
<lista com status>

## Arquivos alterados
<lista>

## Como testar
<passos>

## Report
docs-ai/deliveries/<DELIVERY_ID>/report.md

## Riscos
<lista de riscos identificados>
EOF
)"
```

## Gatilho
Terminal: apos Docs ou Release/Ops.
