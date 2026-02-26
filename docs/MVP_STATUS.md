# MVP Status

Implementação evoluída para um baseline mais próximo do `logs_mvp_implementation_plan.md`:

## Estrutura e arquitetura
- Monorepo com serviços separados em `apps/`.
- Serviços Go organizados em camadas (`internal/config`, `internal/domain`, `internal/infra`, `internal/ports`, `cmd/*`).
- Pacotes compartilhados em `packages/shared-contracts` e `packages/shared-logger`.

## Backend MVP
- `logs-bff`: autenticação JWT, `/api/me`, list/detail de traces via `logs-query`, CRUD de annotations.
- `logs-ingest`: ingestão assíncrona (`202`) com API key + RabbitMQ.
- `logs-worker`: dedupe por `event_dedup` e persistência Cassandra (`nodes`, `edges`, `traces`).
- `logs-query`: listagem por janela temporal + filtros (`status`, `service`) e retorno de grafo.
- `seed` no BFF para bootstrap de empresa/admin/apiKey.

## Frontend MVP
- React + Vite + TypeScript.
- Tailwind CSS (versão atual) integrado ao Vite.
- Login centralizado, layout privado com sidebar animada, filtros de traces e Trace Viewer com React Flow.
- Inclusão de criação de annotation direto na tela do Trace Viewer.
