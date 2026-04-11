package main

import (
	"context"
	"log"
	"net/http"

	"logs-bff/internal/config"
	httpx "logs-bff/internal/infra/http"
	"logs-bff/internal/infra/postgres"
	sharedotel "shared-otel"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	// Init OTEL — no-op if LOGS_API_KEY not set
	if shutdown, err := sharedotel.Init(context.Background(), "logs-bff"); err == nil {
		defer shutdown()
	}

	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	server := httpx.New(cfg, postgres.Repositories{DB: db})
	log.Printf("logs-bff listening on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, server.Handler()))
}
