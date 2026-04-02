---
status: VALIDADO
owner: Rodrigo
ultima_validacao: 2026-04-02
criterio_validacao: codigo+teste
fonte_de_verdade: codigo
escalonamento: PM/Rodrigo
---

# Checklist Antes do PR — Expense Tracker

Antes de declarar uma delivery pronta para PR, verificar:

- [ ] `user_id` scoping validado em toda query SQL modificada.
- [ ] Middleware JWT aplicado em todos os novos endpoints de dados.
- [ ] Teste criado/ajustado para a mudanca (httptest + in-memory repository).
- [ ] Sem segredo novo em arquivo de config/env/doc/comentario.
- [ ] `JWT_SECRET` nao hardcoded em nenhum arquivo.
- [ ] Doc do modulo atualizado no mesmo PR.
- [ ] Status de doc coerente (`DRAFT`, `VALIDADO`, `LEGADO`, `ASSUNCAO`).
- [ ] Divergencias doc x codigo marcadas como `ASSUNCAO` e escaladas.
- [ ] Boundary de camada validada: Handler nao acessa Repository diretamente.
- [ ] Domain nao importa nenhum pacote interno.
- [ ] Erros sentinela usados — sem string literal de erro.
- [ ] Migration nova tem arquivo `up` e `down`.

## Validacao automatica

```bash
scripts/ai/validate-delivery.sh <DELIVERY_ID>
```
