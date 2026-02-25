package config

import (
	"os"
	"strings"
)

type Config struct {
	RabbitMQURL    string
	CassandraHosts []string
}

func Load() Config {
	return Config{RabbitMQURL: env("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"), CassandraHosts: strings.Split(env("CASSANDRA_HOSTS", "localhost"), ",")}
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
