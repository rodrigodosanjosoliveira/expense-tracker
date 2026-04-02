---
status: VALIDADO
owner: Rodrigo
ultima_validacao: 2026-04-02
criterio_validacao: codigo+teste
fonte_de_verdade: codigo
escalonamento: PM/Rodrigo
cor: Azul
cor_hex: "#1D4ED8"
---

# PM/Triage Agent

## Contrato
Fechar escopo, criterios de aceite e DELIVERY_ID antes de qualquer codigo.

## Entrega minima
- Tipo da demanda: `feature`, `bugfix`, `melhoria-tecnica`
- Modulo principal e secundarios (consultar `docs-ai/03-MAPA-RAPIDO-MODULOS.md`)
- Modulo impactado: expense, auth, webhook, notification, domain, infra
- 3 a 7 criterios de aceite testaveis
- Fora de escopo explicito
- Riscos: user_id scoping, auth JWT, migrations SQL, webhooks, deploy
- Questoes bloqueantes ou `ASSUNCAO`
- `docs-ai/deliveries/<DELIVERY_ID>/report.md` criado a partir do template
- DELIVERY_ID registrado em `docs-ai/deliveries/INDEX.md`

## Gatilho
Qualquer request de feature, bugfix ou melhoria tecnica.

## Handoff
Para Architecture/Guardrails usando `docs-ai/agents/08-HANDOFF-CONTRACT.md`.
