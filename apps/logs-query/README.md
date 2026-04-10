# Logs Query API

O `logs-query` é o motor de leitura (Read-Heavy) da aplicação. Ele isola a complexidade de ir até o Cassandra e o OpenSearch para retornar e formatar visões completas dos fluxos e rastros gerados.

## ⚙️ Papel na Arquitetura

Este microserviço apenas **lê** os dados do Cassandra e OpenSearch, nunca faz inserts (CQRS Pattern). Ele constrói a visão gráfica de Árvore/Grafo (nodes + edges) antes de devolver ao BFF e posteriormente ao Front-end.

Suas responsabilidades incluem:

1. Recuperar todos os `Traces` consolidados dentro de uma janela de tempo (ex: últimos 30 min) para uma Company Específica.
2. Recuperar todos os metadados brutos (Nodes, Edges, e Annotations) de um `Trace` em foco.
3. Buscar texto livre cruzando `node ID`, `name` (operações) e `dbQuery` utilizando o **OpenSearch**.
4. Cruzar e extrair dados temporais do **Prometheus** da plataforma (seu TSDB de métricas externas) combinando o `startAt` e `endAt` dos nós baseados em seus `serviceNames`, permitindo que o React exiba gráficos de consumo lado a lado com a rastreabilidade.

## 🚀 Uso e Exemplos de Requisição

A comunicação acontece via requests HTTP. Ele exige o header internamente `X-Company-Id`. Normalmente quem passa esse Header é o BFF.

### Exemplo: Consultar detalhes de um Trace

\`\`\`bash
curl -X GET "http://localhost:8084/query/v1/traces/trace-xyz-123" \
 -H "X-Company-Id: id-da-companhia-do-banco-pg"
\`\`\`

_(A API deve devolver um JSON contando `nodes`, `edges` e `annotations` pertinentes ao Request original)._

### Exemplo: Buscar em Textos Abertos

\`\`\`bash
curl -X GET "http://localhost:8084/query/v1/search?query=SELECT" \
 -H "X-Company-Id: id-da-companhia-do-banco-pg"
\`\`\`
