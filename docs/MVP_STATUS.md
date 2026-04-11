# MVP Status

Implementação evoluída para um baseline mais próximo do `logs_mvp_implementation_plan.md`:

## Últimas Atualizações (Abril 2026)

### ✅ Correções de Infraestrutura

- **Volumes persistentes adicionados**: PostgreSQL e Cassandra agora mantêm dados após `make down`
  - `postgres_data` e `cassandra_data` volumes nomeados criados
- **Makefile atualizado**:
  - `make down`: Remove volumes (comportamento antigo, agora com warning)
  - `make down-keep-data`: Mantém dados nos volumes (recomendado para dev)
- **Order de inicialização corrigida**: Cassandra migration deve rodar ANTES dos services

### 🐛 Bugs Resolvidos

1. **logs-query e logs-worker crashando**: Keyspace `logs` não existia no Cassandra
   - Solução: Rodar `make migrate-cassandra` antes de subir os services
2. **Seed retornando 500**: logs-query não estava rodando
   - Solução: Garantir que Cassandra migration roda primeiro
3. **Dados perdidos no `make down`**: Volumes não eram persistentes
   - Solução: Adicionados volumes nomeados no docker-compose.yml

---

## Estrutura e arquitetura

- Monorepo com serviços separados em `apps/`.
- Serviços Go organizados em camadas (`internal/config`, `internal/domain`, `internal/infra`, `internal/ports`, `cmd/*`).
- Pacotes compartilhados em `packages/shared-contracts` e `packages/shared-logger`.

## Backend MVP

- `logs-bff`: autenticação JWT, `/api/me`, list/detail de traces via `logs-query`, CRUD de annotations, integração de busca textual.
- `logs-ingest`: ingestão assíncrona (`202`) com API key + RabbitMQ.
- `logs-worker`: dedupe por `event_dedup` e persistência Cassandra (`nodes`, `edges`, `traces`). Ingestão paralela de texto no OpenSearch.
- `logs-query`: listagem por janela temporal + filtros (`status`, `service`), retorno de grafo e path customizado `/search` validando e buscando texto livre.
- `seed` no BFF para bootstrap de empresa/admin/apiKey e seed de rastreamento com Elasticsearch/Cassandra.

## Frontend MVP

- React + Vite + TypeScript.
- Tailwind CSS (versão atual) integrado ao Vite.
- Login centralizado, layout privado com sidebar animada, filtros de traces e Trace Viewer com React Flow.
- Inclusão de criação de annotation direto na tela do Trace Viewer.
