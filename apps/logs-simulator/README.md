# Logs Simulator

Simulador de carga para a plataforma **Logs**. Gera traces distribuídos realistas e os envia em massa para testar escalabilidade, busca textual e ingestão.

## ⚙️ Cenários Simulados

O simulator gera traces que representam fluxos reais de microserviços:

| Cenário | Descrição | Nodes |
|---------|-----------|-------|
| **E-Commerce** | Catálogo → Cache → Recommendations → ML Engine | 7 nodes |
| **Payment Processing** | Checkout → Order → Stripe → Rollback (com 5% erro) | 5-6 nodes |
| **User Authentication** | Login → DB → JWT → Redis Session → Audit | 5 nodes |
| **Report Generation** | Heavy SQL → Cache Miss → PDF → S3 Upload | 5 nodes |
| **Webhook Delivery** | Dispatch → Queue → Worker → External Delivery (10% erro) | 4-5 nodes |
| **Data Sync** | CDC → Transform → Partner API → Mark Synced | 5 nodes |

## 🚀 Como Usar

### Via Docker (recomendado)

```bash
# Carga leve: 100 traces, 10 workers
make simulate

# Carga pesada: 5000 traces
BURSTS=100 BURST_SIZE=50 WORKERS=20 make simulate

# Modo OTLP (OpenTelemetry)
MODE=otlp make simulate
```

### Via Variáveis de Ambiente

```bash
# Configuração completa
TARGET_URL=http://localhost/ingest/v1/log-events \
OTLP_URL=http://localhost/v1/traces \
API_KEY=logs_dev_api_key \
MODE=native \
WORKERS=10 \
BURST_SIZE=50 \
BURSTS=20 \
BURST_DELAY=2s \
SEARCH_URL=http://localhost/api/search \
SEARCH_QUERY=SELECT \
make simulate
```

### Via Go (local)

```bash
cd apps/logs-simulator
go mod tidy
go run . 

# Com parâmetros
BURST_SIZE=100 BURSTS=50 go run .
```

## 📊 Métricas Reportadas

Ao final da execução, o simulator exibe:

- **Total Events Sent**: Número total de nodes/events enviados
- **Successful (2xx)**: Eventos aceitos pela API
- **Failed**: Eventos que retornaram erro
- **Search Queries OK/FAIL**: Validações de busca textual
- **Throughput**: Events/segundo
- **Success Rate**: % de sucesso

## 🔍 Validação de Busca Textual

A cada 5 bursts, o simulator executa buscas automáticas:

1. `SELECT` — busca queries SQL
2. `INSERT` — busca inserts
3. Query customizada via `SEARCH_QUERY`

Isso valida que o **OpenSearch** está indexando corretamente os dados.

## 🎯 Casos de Teste

### Teste de Escalabilidade
```bash
BURSTS=200 BURST_SIZE=100 WORKERS=50 make simulate
```
→ 20.000 traces, ~100.000 events

### Teste de Busca Textual
```bash
BURSTS=10 BURST_SIZE=50 SEARCH_QUERY="stripe" make simulate
```
→ Gera dados e valida a busca

### Teste OTLP
```bash
MODE=otlp BURSTS=50 BURST_SIZE=20 make simulate
```
→ Envia via OpenTelemetry HTTP/JSON

### Teste de Resiliência (erros injetados)
```bash
# 5% payment errors, 10% webhook errors
BURSTS=30 BURST_SIZE=50 make simulate
```
