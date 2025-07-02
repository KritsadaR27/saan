package config

import (
	"os"
)

type Config struct {
	// Server Configuration
	Port        string
	Environment string

	// Database Configuration
	DatabaseURL string

	// Redis Configuration
	RedisAddr     string
	RedisPassword string

	// Kafka Configuration
	KafkaBrokers        string
	KafkaConsumerGroup  string
	LoyverseEventTopic  string

	// External Service URLs
	OrderServiceURL string
	ChatServiceURL  string

	// API Keys
	AdminToken string

	// Loyverse Integration
	LoyverseAPIToken string

	// Logging
	LogLevel  string
	LogFormat string
}

func Load() *Config {
	return &Config{
		// Server
		Port:        getEnv("PORT", "8082"),
		Environment: getEnv("GO_ENV", "development"),

		// Database
		DatabaseURL: getEnv("DATABASE_URL", "postgres://saan:saan_password@postgres:5432/saan_db?sslmode=disable"),

		// Redis
		RedisAddr:     getEnv("REDIS_ADDR", "redis:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		// Kafka
		KafkaBrokers:       getEnv("KAFKA_BROKERS", "kafka:9092"),
		KafkaConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "inventory-service"),
		LoyverseEventTopic: getEnv("LOYVERSE_EVENT_TOPIC", "loyverse-events"),

		// External Services
		OrderServiceURL: getEnv("ORDER_SERVICE_URL", "http://order:8081"),
		ChatServiceURL:  getEnv("CHAT_SERVICE_URL", "http://chatbot:8090"),

		// API Keys
		AdminToken:       getEnv("ADMIN_TOKEN", "saan-dev-admin-2024-secure"),
		LoyverseAPIToken: getEnv("LOYVERSE_API_TOKEN", ""),

		// Logging
		LogLevel:  getEnv("LOG_LEVEL", "debug"),
		LogFormat: getEnv("LOG_FORMAT", "json"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
