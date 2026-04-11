package main

import (
	"log"
	"time"

	"logs-worker/internal/config"
	"logs-worker/internal/infra/cassandra"
	"logs-worker/internal/infra/rabbit"
	search "shared-search"

	"github.com/gocql/gocql"
)

func main() {
	cfg := config.Load()
	cluster := gocql.NewCluster(cfg.CassandraHosts...)
	cluster.Keyspace = "logs"
	cluster.Consistency = gocql.Quorum
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

	searchClient, err := search.NewClient(cfg.OpenSearchHosts)
	if err != nil {
		log.Printf("failed to connect to opensearch: %v", err)
	}

	repo := cassandra.Repo{
		Session: session,
		Search:  searchClient,
	}

	if err := rabbit.Consume(cfg.RabbitMQURL, repo); err != nil {
		log.Fatal(err)
	}
}
