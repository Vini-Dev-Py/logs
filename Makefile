up:
	docker compose up -d --build

down:
	docker compose down -v
	@echo "⚠️  Volumes removidos. Para manter dados, use: docker compose down (sem -v)"

down-keep-data:
	docker compose down
	@echo "✅ Containers parados. Dados preservados nos volumes."

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
	docker compose up --build -d logs-simulator
	docker compose logs -f logs-simulator

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
	docker compose up --build -d logs-simulator
	docker compose logs -f logs-simulator

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
	docker compose up --build -d logs-simulator
	docker compose logs -f logs-simulator

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
	docker compose up --build -d logs-simulator
	docker compose logs -f logs-simulator
