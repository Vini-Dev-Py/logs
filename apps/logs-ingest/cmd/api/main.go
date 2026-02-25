package main

import (
	"context"
	"log"
	"net/http"

	"logs-ingest/internal/config"
	httpx "logs-ingest/internal/infra/http"
	"logs-ingest/internal/infra/postgres"
	"logs-ingest/internal/infra/rabbit"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()
	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	pub, err := rabbit.New(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal(err)
	}
	s := httpx.New(postgres.CompanyRepo{DB: db}, pub)
	log.Printf("logs-ingest listening on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, s.Handler()))
}
