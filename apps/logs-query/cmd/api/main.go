package main

import (
	"log"
	"net/http"

	"logs-query/internal/config"
	"logs-query/internal/infra/cassandra"
	httpx "logs-query/internal/infra/http"

	"github.com/gocql/gocql"
)

func main() {
	cfg := config.Load()
	cluster := gocql.NewCluster(cfg.CassandraHosts...)
	cluster.Keyspace = "logs"
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	srv := httpx.New(cassandra.Repo{Session: session})
	log.Printf("logs-query listening on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, srv.Handler()))
}
