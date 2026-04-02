---
status: DRAFT
owner: Rodrigo
ultima_validacao: 2026-04-02
criterio_validacao: codigo+teste
fonte_de_verdade: codigo
escalonamento: PM/Rodrigo
delivery_id: "<DELIVERY_ID>"
tipo_demanda: "feature|bugfix|melhoria-tecnica"
modulo_principal: "<MODULO>"
stack_impactado: "Go (expense-tracker)|transversal"
require_security: false
require_release_ops: false
require_front_handoff: false
arquitetura_change: false
rodrigo_approval: "N/A"
---

# Delivery Report - <DELIVERY_ID>

## PM/Triage

- Escopo fechado:
- Fora de escopo:
- Criterios de aceite:
- Riscos:

## Architecture/Guardrails

- Invariantes aplicaveis:
- Arquivos provaveis de impacto:
- Assuncoes:
- Mudanca de arquitetura?: nao
- Rodrigo aprovou?: N/A

## Data Model

- Resultado: APROVADO|RETORNA_GUARDRAILS|N/A
- Schema diff (domain entities / migrations adicionadas, alteradas ou removidas):
- Boundary layered (Handler->Service->Repository->Domain): ok|violacao
- user_id scoping: ok|risco

## Tests

- Resultado: escritos|N/A
- Testes criados:
- Criterios de aceite cobertos:
- Gaps ou cenarios nao testaveis:

## Dev

- Arquivos alterados:
- Resumo tecnico:
- Testes criados/ajustados:
- Evidencias (codigo/teste):

## Code Review

- Resultado: APROVADO|RETORNA_DEV
- Achados BLOQUEANTE:
- Achados SUGESTAO:

## Query Performance

- Resultado: APROVADO|RETORNA_DEV|N/A
- Queries sem user_id filter encontradas:
- Queries unbounded (sem paginacao) encontradas:
- N+1 patterns encontrados:
- ASSUNCOES (fora do escopo):

## QA

- Resultado: APROVA|AJUSTA
- Cenarios executados:
- Regressao validada:

## Docs

- Docs atualizados:
- Status dos docs:
- Divergencias doc x codigo:

## Security

- Resultado: OK|BLOQUEIA|N/A
- Itens revisados:

## Schema Evolution Review

- Resultado: LIBERADO|BLOQUEIA|N/A
- Migrations auditadas:
- Backward compatibility: ok|risco
- Down migration testada: sim|nao|N/A
- Risco: BAIXO|MEDIO|ALTO

## Release/Ops

- Resultado: PRONTO|AJUSTAR|N/A
- Impacto operacional:
- Plano de rollback:

## Final Checklist

- [ ] user_id scoping validado em toda query SQL modificada.
- [ ] Middleware JWT aplicado em todos os novos endpoints de dados.
- [ ] Teste criado/ajustado para a mudanca.
- [ ] Sem segredo novo em arquivo de config/env/doc.
- [ ] Doc do modulo atualizado no mesmo PR.
- [ ] Status de doc coerente (`DRAFT`, `VALIDADO`, `LEGADO`, `ASSUNCAO`).
- [ ] Divergencias com codigo marcadas e escaladas.
- [ ] Boundary de camada validada (Handler nao acessa Repository).
- [ ] Domain nao importa nenhum pacote interno.
- [ ] Migration nova tem arquivo `up` e `down` (se aplicavel).
