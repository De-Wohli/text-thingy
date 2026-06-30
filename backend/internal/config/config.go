// Package config centralizes environment-driven configuration so both the
// gateway and worker binaries read settings the same way.
package config

import "os"

type Config struct {
	Port        string
	PostgresURL string
	RedisAddr   string
	RabbitMQURL string
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8080"),
		PostgresURL: getEnv("DATABASE_URL", "postgres://dnd:dnd@localhost:5432/dnd5e?sslmode=disable"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
