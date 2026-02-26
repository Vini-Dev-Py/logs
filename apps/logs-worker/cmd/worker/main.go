package main

import (
	"log"

	"logs-worker/internal/config"
	"logs-worker/internal/infra/cassandra"
	"logs-worker/internal/infra/rabbit"

	"github.com/gocql/gocql"
)

func main() {
	cfg := config.Load()
	cluster := gocql.NewCluster(cfg.CassandraHosts...)
	cluster.Keyspace = "logs"
	cluster.Consistency = gocql.Quorum
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	if err := rabbit.Consume(cfg.RabbitMQURL, cassandra.Repo{Session: session}); err != nil {
		log.Fatal(err)
	}
}
