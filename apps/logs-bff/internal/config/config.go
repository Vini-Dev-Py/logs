package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	QueryURL    string
}

func Load() Config {
	return Config{
		Port:        env("PORT", "8081"),
		DatabaseURL: env("DATABASE_URL", "postgres://logs:logs@localhost:5432/logs?sslmode=disable"),
		JWTSecret:   env("JWT_SECRET", "secret"),
		QueryURL:    env("QUERY_URL", "http://localhost:8084"),
	}
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
