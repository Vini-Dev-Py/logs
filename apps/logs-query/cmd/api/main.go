package main

import (
	"log"
	"net/http"
	"time"

	"logs-query/internal/config"
	"logs-query/internal/infra/cassandra"
	httpx "logs-query/internal/infra/http"

	"github.com/gocql/gocql"
)

func main() {
	cfg := config.Load()
	cluster := gocql.NewCluster(cfg.CassandraHosts...)
	cluster.Keyspace = "logs"
	var session *gocql.Session
	var err error

	// Retry connection up to 15 times, waiting 2 seconds between attempts
	for i := 0; i < 15; i++ {
		session, err = cluster.CreateSession()
		if err == nil {
			break
		}
		log.Printf("cassandra: waiting for cassandra... %v", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("failed to connect to cassandra after retries: %v", err)
	}
	defer session.Close()
	srv := httpx.New(cassandra.Repo{Session: session})
	log.Printf("logs-query listening on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, srv.Handler()))
}
