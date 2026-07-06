# Postman — Expense Tracker API

Arquivos para testar a API no Postman.

- `expense-tracker.postman_collection.json` — endpoints (Auth, Expenses, Categories, Webhooks, Health)
- `expense-tracker.postman_environment.json` — environment `Expense Tracker - Local` (`base_url`, `jwt_token`, `user_id`)

## Como usar

1. No Postman: **Import** → selecione os dois arquivos.
2. Selecione o environment **"Expense Tracker - Local"** (canto superior direito).
3. Suba a API com PostgreSQL — auth/webhooks exigem banco (veja `../RUNNING.md`).
4. Rode **Authentication → Register** (ou **Login**). O token JWT é salvo
   automaticamente em `{{jwt_token}}` pelo script de teste.
5. Use as demais pastas. A collection já envia `Authorization: Bearer {{jwt_token}}`
   por herança (auth no nível da collection).

## Variáveis preenchidas automaticamente

- `jwt_token`, `user_id` — no Register/Login (environment)
- `expense_id`, `category_id`, `webhook_id` — ao criar cada recurso (collection)

Assim, os requests de Get/Update/Delete usam o último recurso criado sem edição manual.

## Rodando via WSL (banco/API dentro do WSL)

Se você subir o PostgreSQL e/ou a API **dentro do WSL** (ex.: Docker instalado só no
WSL, não no Windows Docker Desktop), a rede entre o Windows e a VM do WSL costuma
bloquear o acesso por `localhost`. Duas situações:

- **API rodando no Windows, Postgres em container no WSL:** o encaminhamento de porta do
  WSL2 pode não espelhar a porta do container para o Windows. Aponte o `DB_HOST` da API
  para o IP da distro (`wsl hostname -I`) em vez de `localhost`. Se o firewall bloquear,
  rode a API também dentro do WSL.
- **API rodando dentro do WSL:** o Postman (app Windows) pode não alcançar
  `http://localhost:8080`. Ajuste `base_url` no environment para o IP da distro, ex.:
  `http://172.28.250.104:8080` (rode `wsl hostname -I` para descobrir o IP atual — ele
  muda a cada reinício do WSL). Alternativamente, teste de dentro do WSL com `curl`.

> A collection foi validada de ponta a ponta contra a API real (Go + PostgreSQL):
> **24/24** requests OK, cobrindo register → login → JWT → CRUD de categories/expenses/
> webhooks e os casos de erro (401 sem token, 409 duplicada, 400 payload inválido).

## Notas

- `base_url` padrão: `http://localhost:8080`. Ajuste se mudar `SERVER_PORT` ou usar WSL
  (ver seção acima).
- Em modo in-memory (`DB_HOST` vazio) `/auth/*` e `/webhooks` ficam indisponíveis, e as
  rotas de expenses/categories retornam **401** — os handlers extraem `user_id` do
  contexto, que só é preenchido pelo middleware JWT (ativo apenas com PostgreSQL). Ou
  seja, para exercitar a collection use **PostgreSQL**.
- Nos filtros de **List Expenses**, os parâmetros opcionais vêm desabilitados —
  habilite-os na aba *Params* conforme necessário.
