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

# Handoff Contract (padrao unico)

Use este formato em qualquer transicao entre agents.

## Template

```md
## Handoff

- Agent origem:
- Agent destino:
- Tipo da demanda: feature | bugfix | melhoria-tecnica
- Modulo principal:
- Modulos secundarios:
- Stack impactado: Go (expense-tracker) | transversal
- Escopo fechado:
- Fora de escopo:
- Criterios de aceite:
- Arquivos provaveis de impacto:
- Invariantes aplicaveis:
- Riscos:
- Assuncoes:
- Evidencias (codigo/teste):
- Proximo passo esperado:
```

## Regras

- Nao enviar handoff com campos vazios criticos (`escopo`, `aceite`, `riscos`).
- Se existir conflito doc x codigo, preencher `Assuncoes` e escalar.
- O destino pode devolver handoff incompleto para ajuste.
