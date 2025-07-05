package config

import (
	"os"
)

type Config struct {
	ServerPort      string
	DatabaseURL     string
	RedisURL        string
	KafkaBrokers    []string
	KafkaTopic      string
	ServiceName     string
}

func Load() *Config {
	return &Config{
		ServerPort:   getEnv("SERVER_PORT", "8086"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://localhost/shipping?sslmode=disable"),
		RedisURL:     getEnv("REDIS_URL", "redis://localhost:6379"),
		KafkaBrokers: []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		KafkaTopic:   getEnv("KAFKA_TOPIC", "shipping.events"),
		ServiceName:  getEnv("SERVICE_NAME", "shipping-service"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
