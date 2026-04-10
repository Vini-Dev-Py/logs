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
