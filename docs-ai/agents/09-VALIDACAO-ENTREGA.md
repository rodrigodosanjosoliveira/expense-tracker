---
status: VALIDADO
owner: Rodrigo
ultima_validacao: 2026-04-02
criterio_validacao: codigo+teste
fonte_de_verdade: codigo
escalonamento: PM/Rodrigo
---

# Validacao de Entrega

## Artefato obrigatorio

Todo delivery deve produzir `docs-ai/deliveries/<DELIVERY_ID>/report.md` a partir do template em `docs-ai/deliveries/_template/report.md`.

## Comando de validacao

```bash
scripts/ai/validate-delivery.sh <DELIVERY_ID>
```

## O que o script valida

1. **Secoes obrigatorias**: PM/Triage, Architecture/Guardrails, Dev, QA, Docs, Final Checklist.
2. **Campos criticos preenchidos**: nenhum campo com `TODO`, `TBD`, `pendente` ou vazio.
3. **Final Checklist completo**: todos os itens marcados.
4. **Flags condicionais**: se `require_security=true`, secao Security deve existir com resultado. Idem para `require_front_handoff`, `require_release_ops`, `arquitetura_change`.
5. **Aprovacao de arquitetura**: se `arquitetura_change=true`, `rodrigo_approval` deve estar preenchido.

## Regra bloqueante

Se o comando falhar, a delivery NAO esta pronta para PR. Corrigir os erros apontados antes de prosseguir.
