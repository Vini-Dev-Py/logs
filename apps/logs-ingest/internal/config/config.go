package config

import "os"

type Config struct{ Port, DatabaseURL, RabbitMQURL string }

func Load() Config {
	return Config{Port: env("PORT", "8082"), DatabaseURL: env("DATABASE_URL", ""), RabbitMQURL: env("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")}
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
