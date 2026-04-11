# Logs Ingest API

Este é o serviço responsável por receber todos os eventos de logs enviados pelas aplicações clientes (através de bibliotecas, curl, etc). Ele atua como uma porta de entrada rápida e assíncrona para não onerar o cliente que está gerando o log.

## ⚙️ Papel na Arquitetura

O `logs-ingest` é otimizado para **alta vazão (throughput)**.
Ao receber um payload HTTP com os dados do log, ele:

1. Valida de forma básica o payload (Schema, IDs obrigatórios).
2. Valida a **API Key** enviada no Header (`Authorization: Bearer <KEY>`) consultando o banco PostgreSQL para garantir que pertence a uma empresa ativa.
3. Envia o evento (Node) para a fila do **RabbitMQ**.
4. Retorna um `202 Accepted` imediatamente para o cliente.

Como ele joga a carga pesada de escrita no Cassandra para a fila, o cliente nunca fica travado esperando a persistência lenta acontecer.

## 🚀 Como testar (Métricas e Ingestão)

Com a API Key gerada no banco de dados (o seed inicial cria a key `logs_dev_api_key` padrão para testes), você pode simular a subida de um log para a plataforma.

### Exemplo via cURL

\`\`\`bash
curl -X POST http://localhost/ingest/v1/log-events \
 -H "Authorization: Bearer logs_dev_api_key" \
 -H "Content-Type: application/json" \
 -d '{
"traceId": "trace-xyz-123",
"nodeId": "node-xyz-456",
"parentNodeId": null,
"serviceName": "payment-api",
"operation": {
"type": "HTTP",
"name": "POST /payments",
"status": "OK",
"startAt": "2026-02-25T10:00:00Z",
"endAt": "2026-02-25T10:00:05Z",
"durationMs": 5000
}
}'
\`\`\`

A resposta esperada é um HTTP 202 sem corpo, indicando sucesso no enfileiramento.
