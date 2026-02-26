package config

import (
	"os"
	"strings"
)

type Config struct {
	Port           string
	CassandraHosts []string
}

func Load() Config {
	return Config{Port: env("PORT", "8084"), CassandraHosts: strings.Split(env("CASSANDRA_HOSTS", "localhost"), ",")}
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
