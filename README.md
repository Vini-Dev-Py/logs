# 🚀 Logs — Visual Log Tracing Platform

> Transforme logs em **fluxos visuais interativos**.
> Entenda o que aconteceu no seu sistema em segundos, não minutos.

---

## 🧠 Sobre o Projeto

O **Logs** é uma plataforma de observabilidade que transforma logs tradicionais (texto) em uma **visualização baseada em grafos (nodes + edges)**.

Em vez de ler centenas de linhas de log, você vê **o fluxo completo de execução de uma requisição** — com operações, decisões, chamadas externas e transições entre microserviços.

---

## 🎯 Problema

Ferramentas atuais como Grafana + Loki:

- ❌ Logs em formato texto (difícil leitura)
- ❌ Correlação complexa mesmo com requestId
- ❌ Difícil identificar causa raiz
- ❌ Péssimo para microserviços
- ❌ Debug lento

---

## 💡 Solução

O Logs transforma isso em:

- ✅ **Fluxo visual interativo (React Flow)**
- ✅ Cada operação vira um **node**
- ✅ Conexões mostram o caminho da execução
- ✅ Suporte a **microserviços (ramificações)**
- ✅ Highlight de erros e gargalos
- ✅ Debug muito mais rápido

---

## 🖥️ Preview do Conceito

```
[HTTP /clients]
        ↓
[DB SELECT users]
        ↓
[IF user exists]
     ↙       ↘
[CREATE]   [RETURN ERROR]
        ↓
[CALL PAYMENT SERVICE]
        ↓
[EXTERNAL API PIX]
```

---

## ⚙️ Arquitetura

### 🧩 Serviços

| Serviço     | Responsabilidade                     |
| ----------- | ------------------------------------ |
| logs-bff    | Auth, usuários, empresas, permissões |
| logs-ingest | Recebe logs e envia para fila        |
| logs-worker | Processa fila e salva no banco       |
| logs-query  | Consulta e monta o fluxo             |
| web         | Interface (React)                    |

---

### 🗄️ Bancos

| Banco              | Uso                                 |
| ------------------ | ----------------------------------- |
| PostgreSQL         | usuários, empresas, permissões      |
| Cassandra          | armazenamento de logs (alta escala) |
| RabbitMQ           | fila de processamento               |
| OpenSearch (v2.11) | busca textual por logs              |

---

### 🔁 Fluxo de Dados

```
Client → logs-ingest → RabbitMQ → logs-worker → Cassandra
                                          ↓
                                       logs-query → BFF → Frontend
```

---

## 🧱 Stack

### Backend

- Go (alta performance)
- PostgreSQL (relacional)
- Cassandra (logs massivos)
- RabbitMQ (fila)

### Frontend

- React + Vite + TypeScript
- Tailwind CSS
- Framer Motion
- React Flow

### Infra

- Docker
- Docker Compose
- Traefik (load balancer)

---

## 🧠 Conceitos Importantes

### Trace

Representa uma requisição completa.

### Node

Uma operação dentro do fluxo:

- DB Query
- HTTP call
- IF / LOOP
- Microserviço
- API externa

### Edge

Conexão entre nodes (fluxo de execução).

---

## 🔌 API — Ingestão

### POST `/ingest/v1/log-events`

```json
{
  "traceId": "abc-123",
  "nodeId": "node-1",
  "parentNodeId": null,
  "serviceName": "api-gateway",
  "operation": {
    "type": "HTTP",
    "name": "GET /clients",
    "status": "OK",
    "startAt": "2026-01-01T10:00:00Z",
    "endAt": "2026-01-01T10:00:01Z",
    "durationMs": 1000
  }
}
```

---

## 🔐 Autenticação

Cada empresa possui uma **API Key**:

```
Authorization: Bearer <API_KEY>
```

---

## 🎨 UI / UX

### 🔐 Login

- Centralizado
- Simples
- Foco total no usuário

### 📂 Layout

- Sidebar com animações (Framer Motion)
- Menus com ícones + textos
- Hover states suaves

### 📊 Trace Viewer

- Graph interativo (React Flow)
- Nodes clicáveis
- Painel lateral com detalhes
- Highlight de:
  - erros
  - lentidão

---

## 🚀 Como rodar o projeto

### Pré-requisitos

- Docker
- Docker Compose
- **Nota (Linux/WSL):** O OpenSearch requer aumento de memória virtual configurado no host. Rode: `sudo sysctl -w vm.max_map_count=262144` e adicione em `/etc/sysctl.conf` para persistir.

---

### 🐳 Subir ambiente

```bash
make up
```

ou

```bash
docker-compose up -d
```

---

### 🧱 Rodar migrations

```bash
make migrate
```

---

### 🌱 Seed inicial

```bash
make seed
```

Isso cria:

- Empresa padrão
- Usuário admin
- API Key

---

### 🌐 Acessos

| Serviço    | URL                     |
| ---------- | ----------------------- |
| Frontend   | http://localhost:5173   |
| API BFF    | http://localhost/api    |
| Ingest     | http://localhost/ingest |
| OpenSearch | http://localhost:9200   |

---

## 🧪 Exemplo de uso

Enviar um log:

```bash
curl -X POST http://localhost/ingest/v1/log-events \
  -H "Authorization: Bearer SUA_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{ ... }'
```

---

## 📈 Roadmap

### MVP

- [x] Ingestão de logs
- [x] Visualização de fluxo
- [x] Multi-tenant
- [x] Anotações

### Próximos passos

- [x] Busca textual (OpenSearch)
- [ ] Correlação com métricas (CPU, RAM)
- [ ] Auto-instrumentação (OpenTelemetry)
- [ ] RBAC avançado
- [ ] Persistência de layout do graph

---

## 🧠 Diferencial

Esse projeto não é só mais uma ferramenta de logs.


Ele resolve um problema real:

> **"Entender rapidamente o que aconteceu em uma requisição complexa."**

---

## 🔭 OpenTelemetry (OTLP) Integration

A plataforma aceita traces no formato padrão **OpenTelemetry (OTLP HTTP/JSON)**, permitindo que qualquer aplicação que já use OTEL SDK envie dados sem adapters ou agentes extras.

### Endpoint

```
POST http://<host>/v1/traces
Content-Type: application/json
Authorization: Bearer <sua-api-key>
```

### Configurando sua aplicação (qualquer linguagem)

Basta apontar o seu OTEL SDK para o endpoint acima:

**Variáveis de ambiente (recomendado):**
```bash
OTEL_EXPORTER_OTLP_ENDPOINT=http://<host>
OTEL_EXPORTER_OTLP_HEADERS="Authorization=Bearer <api-key>"
OTEL_SERVICE_NAME=meu-servico
```

**Go (usando nosso shared-otel helper):**
```go
import sharedotel "shared-otel"

shutdown, err := sharedotel.Init(ctx, "meu-servico")
defer shutdown()
```

**Node.js:**
```bash
npm install @opentelemetry/sdk-node @opentelemetry/exporter-trace-otlp-http
```
```js
const { OTLPTraceExporter } = require('@opentelemetry/exporter-trace-otlp-http')
const exporter = new OTLPTraceExporter({
  url: 'http://<host>/v1/traces',
  headers: { Authorization: 'Bearer <api-key>' }
})
```

**Python:**
```bash
pip install opentelemetry-exporter-otlp-proto-http
```
```python
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
exporter = OTLPSpanExporter(
    endpoint="http://<host>/v1/traces",
    headers={"Authorization": "Bearer <api-key>"}
)
```

**Delphi / Pascal (REST Client):**
Como o Delphi pode ser usado em sistemas ERP legados/desktop e pode não possuir um SDK oficial OTEL ativo, basta montar o payload JSON e enviar um POST. Exemplo usando `TRESTClient`:
```pascal
uses
  REST.Client, REST.Types;

var
  Client: TRESTClient;
  Request: TRESTRequest;
  PayloadOTLP: string;
begin
  Client := TRESTClient.Create('http://<host>/v1/traces');
  Request := TRESTRequest.Create(nil);
  try
    Request.Client := Client;
    Request.Method := rmPOST;
    
    // Autenticação usando a sua API Key
    Request.CustomHeaders.AddValue('Authorization', 'Bearer <api-key>');
    
    // Construa o JSON de acordo com a spec OTLP
    PayloadOTLP := '{ "resourceSpans": [ ... ] }';
    Request.AddBody(PayloadOTLP, ctAPPLICATION_JSON);
    
    Request.Execute;
  finally
    Request.Free;
    Client.Free;
  end;
end;
```

---

## 🤝 Contribuição

PRs são bem-vindos 🚀

---

## 📄 Licença

MIT

---

## 👨‍💻 Autor

Desenvolvido por **Vini**
💡 Focado em sistemas distribuídos, performance e observabilidade
