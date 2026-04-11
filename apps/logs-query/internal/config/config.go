package config

import (
	"os"
	"strings"
)

type Config struct {
	Port            string
	CassandraHosts  []string
	OpenSearchHosts []string
}

func Load() Config {
	return Config{
		Port:            env("PORT", "8084"),
		CassandraHosts:  strings.Split(env("CASSANDRA_HOSTS", "localhost"), ","),
		OpenSearchHosts: strings.Split(env("OPENSEARCH_HOSTS", "http://localhost:9200"), ","),
	}
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
