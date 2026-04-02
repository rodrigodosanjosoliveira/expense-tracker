---
status: VALIDADO
owner: Rodrigo
ultima_validacao: 2026-04-02
criterio_validacao: codigo+teste
fonte_de_verdade: codigo
escalonamento: PM/Rodrigo
cor: Vermelho
cor_hex: "#B91C1C"
---

# Security Agent

## Contrato
Revisar auth, JWT, bcrypt, user_id scoping, segredos e integracoes externas.

## Entrega minima
- Superficie revisada, achados, risco residual
- Resultado: `OK` ou `BLOQUEIA`

## Pontos de atencao especificos do Expense Tracker
- JWT: validacao no middleware `internal/middleware/auth.go`, claims corretos
- user_id: toda query usa `filters.UserID` — nunca retornar dados de outro usuario
- bcrypt: senhas nunca em texto plano, custo adequado
- JWT_SECRET: nao pode ser valor default inseguro em producao
- Webhooks: scoped por user_id, URLs nao expostas a outros usuarios
- Secrets em config/env/docs: nenhum segredo hardcoded
- Input validation: todos os endpoints validam entrada antes de processar

## Gatilho
Dev sinaliza impacto de seguranca no handoff.

## Handoff
Se `BLOQUEIA` → Dev. Se `OK` → QA ou Release/Ops (proximo no fluxo).
