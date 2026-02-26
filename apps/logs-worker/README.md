# Logs Worker

Este é um serviço background runner (worker) dedicado ao processamento assíncrono e persistência definitiva dos logs na infraestrutura principal.

## ⚙️ Papel na Arquitetura

O `logs-worker` consome a fila do **RabbitMQ** populada pelo `logs-ingest`, e efetua as seguintes tarefas:

1. Deserializa a mensagem da fila.
2. Faz o **Parse** para a estrutura de dados (DTO) de inserção.
3. Conduz a persistência dos dados brutos e massivos no **Cassandra** (`nodes`, `traces`, `edges`).
4. (Recente) Realiza a persistência textual para indexação livre no **OpenSearch**, de modo a permitir a busca Full-Text (como Ctrl+F em traces).

Para evitar que mensagens enviadas em duplicidade sujem o banco, este serviço implementa **deduplicação (Idempotency)** por ID de evento/mensagem antes da inserção final.

## 🚀 Testes Locais

Diferente de APIs web, o worker não abre portas HTTP para requests. Ele só precisa do Rabbit e do Cassandra (e OpenSearch) online.

Para testá-lo em isolamento, os logs nos containers bastam:
\`\`\`bash
docker compose logs -f logs-worker
\`\`\`
_(É esperado observar logs informando qual node/id e trace/id foram sincronizados nas DBs e a volumetria processada)._
