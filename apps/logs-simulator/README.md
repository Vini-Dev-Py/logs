# Logs Simulator Server

Simulador de carga **on-demand** para testes de ingestão de logs e traces.

## Como funciona

O simulador roda como um **servidor HTTP contínuo** que aguarda requisições para iniciar simulações. Ele **NÃO** inicia automaticamente - você precisa chamá-lo via API.

## Inicialização

```bash
# Iniciar o servidor do simulador (perfil simulation)
make simulate

# Ou diretamente
docker compose --profile simulation up -d --build logs-simulator
```

## Endpoints HTTP

### Iniciar simulação

```bash
# Com configurações padrão
curl -X POST http://localhost/simulator/simulate

# Com configurações customizadas
curl -X POST http://localhost/simulator/simulate \
  -H "Content-Type: application/json" \
  -d '{
    "bursts": 10,
    "burstSize": 25,
    "mode": "native",
    "companyId": "my-company-id"
  }'
```

### Verificar status

```bash
curl http://localhost/simulator/simulate
```

Resposta:

```json
{
  "id": "abc123...",
  "status": "running", // ou "completed", "failed", "stopped"
  "startedAt": "2026-04-12T10:00:00Z",
  "config": {
    "mode": "native",
    "burstSize": 50,
    "bursts": 20,
    "companyId": "c152767f-..."
  },
  "totalSent": 500,
  "totalOk": 498,
  "totalFail": 2
}
```

### Parar simulação em andamento

```bash
curl -X POST http://localhost/simulator/simulate/stop
```

### Health check

```bash
curl http://localhost/simulator/health
```

## Comandos Makefile

| Comando                | Descrição                                         |
| ---------------------- | ------------------------------------------------- |
| `make simulate`        | Inicia o servidor e mostra logs                   |
| `make simulate-run`    | Dispara uma simulação via HTTP                    |
| `make simulate-status` | Verifica status da simulação                      |
| `make simulate-stop`   | Para simulação em andamento                       |
| `make simulate-light`  | Inicia com carga leve (5 bursts x 20 traces)      |
| `make simulate-heavy`  | Inicia com carga pesada (100 bursts x 100 traces) |
| `make simulate-otlp`   | Inicia usando formato OTLP                        |
| `make simulator-down`  | Para o servidor do simulador                      |

## Configuração via Variáveis de Ambiente

| Variável       | Default                                        | Descrição                  |
| -------------- | ---------------------------------------------- | -------------------------- |
| `PORT`         | `8085`                                         | Porta do servidor          |
| `TARGET_URL`   | `http://logs-ingest:8082/ingest/v1/log-events` | URL de ingestão nativa     |
| `OTLP_URL`     | `http://logs-ingest:8082/v1/traces`            | URL de ingestão OTLP       |
| `API_KEY`      | `logs_dev_api_key`                             | Chave de autenticação      |
| `MODE`         | `native`                                       | Modo: `native` ou `otlp`   |
| `WORKERS`      | `10`                                           | Workers concorrentes       |
| `BURST_SIZE`   | `50`                                           | Traces por burst (default) |
| `BURSTS`       | `20`                                           | Número de bursts (default) |
| `BURST_DELAY`  | `2s`                                           | Intervalo entre bursts     |
| `SEARCH_URL`   | `http://logs-bff:8081/api/search`              | URL para validação         |
| `SEARCH_QUERY` | `SELECT`                                       | Query de busca padrão      |
| `COMPANY_ID`   | `c152767f-...`                                 | ID da empresa nos eventos  |

## Notas

- **Apenas para desenvolvimento**: O simulador usa o profile `simulation` e **nunca** sobe em produção
- **Traefik**: Disponível em `/simulator/*` via reverse proxy
- **Cenários**: Gera traces realistas de e-commerce, pagamentos, autenticação, relatórios, webhooks e sync de dados
