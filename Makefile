up:
	docker compose up -d --build

down:
	docker compose down -v

logs:
	docker compose logs -f --tail=100

migrate:
	docker compose exec -T postgres psql -U logs -d logs -f /workspace/infra/migrations/001_init.sql

seed:
	docker compose exec -T logs-bff /app/seed
