---
status: VALIDADO
owner: Rodrigo
ultima_validacao: 2026-04-02
criterio_validacao: codigo+teste
fonte_de_verdade: codigo
escalonamento: PM/Rodrigo
cor: Laranja
cor_hex: "#C2410C"
---

# QA Agent

## Contrato
Validar aceite, regressao e pontos criticos do diff.

## Entrega minima
- Happy path, edge case principal, regressao validada
- Risco residual
- Resultado: `APROVA` ou `AJUSTA`

## Comandos de teste
```bash
go test -v ./...
go test -v ./internal/handler/
go test -v ./internal/service/
```

## Gatilho
Code review aprovado, implementacao pronta para QA.

## Handoff
Se aprovado → Docs. Se bloqueado → Dev.
