up:
	docker compose up -d --build

down:
	docker compose down -v
	@echo "⚠️  Volumes removidos. Para manter dados, use: docker compose down (sem -v)"

down-keep-data:
	docker compose down
	@echo "✅ Containers parados. Dados preservados nos volumes."

# Rebuild apenas serviços que mudaram (Go/Node detecta mudanças no Dockerfile/context)
restart:
	@echo "🔄 Reconstruindo e reiniciando serviços de aplicação..."
	docker compose up -d --build logs-bff logs-query logs-ingest logs-worker web
	@echo "✅ Aplicação reiniciada. Infraestrutura (postgres, cassandra, etc.) mantida."

# Rebuild apenas um serviço específico (ex: make restart-service SERVICE=logs-query)
restart-service:
	@echo "🔄 Reconstruindo $(SERVICE)..."
	docker compose up -d --build $(SERVICE)
	@echo "✅ $(SERVICE) reiniciado."

# Apenas reinicia sem rebuild (mais rápido, usa imagem já construída)
restart-quick:
	@echo "🔄 Reiniciando serviços sem rebuild..."
	docker compose restart logs-bff logs-query logs-ingest logs-worker web
	@echo "✅ Serviços reiniciados."

# Rebuild do frontend apenas
restart-web:
	@echo "🔄 Reconstruindo frontend..."
	docker compose up -d --build web
	@echo "✅ Frontend reiniciado."

# Rebuild do backend apenas
restart-backend:
	@echo "🔄 Reconstruindo serviços backend..."
	docker compose up -d --build logs-bff logs-query logs-ingest logs-worker
	@echo "✅ Backend reiniciado."

logs:
	docker compose logs -f --tail=100

migrate: migrate-postgres migrate-cassandra

migrate-postgres:
	docker compose exec -T postgres psql -U logs -d logs -f /workspace/infra/migrations/001_init.sql
	docker compose exec -T postgres psql -U logs -d logs -f /workspace/infra/migrations/002_rbac_init.sql

migrate-cassandra:
	docker compose exec -T cassandra cqlsh -u cassandra -p cassandra -f /workspace/infra/cassandra/schema.cql

seed:
	docker compose exec -T logs-bff /app/seed

simulate:
	@echo "🎭 Starting simulator server..."
	docker compose --profile simulation up -d --build logs-simulator
	@echo "✅ Simulator running. Use endpoints below to trigger simulations:"
	@echo ""
	@echo "   POST /simulator/simulate     - Start simulation (default config)"
	@echo "   GET  /simulator/simulate     - Check simulation status"
	@echo "   POST /simulator/simulate/stop - Stop running simulation"
	@echo "   POST /simulator/seed         - Alias for /simulate (backwards compat)"
	@echo "   GET  /simulator/health       - Health check"
	@echo ""
	@echo "Examples:"
	@echo "   curl -X POST http://localhost/simulator/simulate"
	@echo "   curl -X POST http://localhost/simulator/simulate -H 'Content-Type: application/json' -d '{\"bursts\":10,\"burstSize\":20}'"
	@echo "   curl http://localhost/simulator/simulate"
	docker compose --profile simulation logs -f logs-simulator

simulate-light:
	TARGET_URL=http://logs-ingest:8082/ingest/v1/log-events \
	OTLP_URL=http://logs-ingest:8082/v1/traces \
	API_KEY=logs_dev_api_key \
	MODE=native \
	WORKERS=5 \
	BURST_SIZE=20 \
	BURSTS=5 \
	BURST_DELAY=1s \
	SEARCH_URL=http://logs-bff:8081/api/search \
	docker compose --profile simulation up -d --build logs-simulator
	docker compose --profile simulation logs -f logs-simulator

simulate-heavy:
	TARGET_URL=http://logs-ingest:8082/ingest/v1/log-events \
	OTLP_URL=http://logs-ingest:8082/v1/traces \
	API_KEY=logs_dev_api_key \
	MODE=native \
	WORKERS=50 \
	BURST_SIZE=100 \
	BURSTS=100 \
	BURST_DELAY=500ms \
	SEARCH_URL=http://logs-bff:8081/api/search \
	docker compose --profile simulation up -d --build logs-simulator
	docker compose --profile simulation logs -f logs-simulator

simulate-otlp:
	TARGET_URL=http://logs-ingest:8082/ingest/v1/log-events \
	OTLP_URL=http://logs-ingest:8082/v1/traces \
	API_KEY=logs_dev_api_key \
	MODE=otlp \
	WORKERS=20 \
	BURST_SIZE=50 \
	BURSTS=20 \
	BURST_DELAY=1s \
	SEARCH_URL=http://logs-bff:8081/api/search \
	docker compose --profile simulation up -d --build logs-simulator
	docker compose --profile simulation logs -f logs-simulator

# Run a single simulation on-demand via HTTP
simulate-run:
	@echo "🚀 Triggering simulation..."
	curl -X POST http://localhost/simulator/simulate \
		-H "Content-Type: application/json" \
		-d '{"bursts":${BURSTS:-20},"burstSize":${BURST_SIZE:-50}}'
	@echo ""
	@echo "📊 Check status with: curl http://localhost/simulator/simulate"

simulate-status:
	@echo "📊 Simulation status:"
	curl -s http://localhost/simulator/simulate | python3 -m json.tool 2>/dev/null || curl -s http://localhost/simulator/simulate

simulate-stop:
	@echo "⏹️  Stopping simulation..."
	curl -X POST http://localhost/simulator/simulate/stop
	@echo ""

simulator-down:
	@echo "🛑 Stopping simulator server..."
	docker compose --profile simulation down logs-simulator
	@echo "✅ Simulator stopped."
