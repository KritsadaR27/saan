package config

import (
	"os"
)

// Config holds the application configuration
type Config struct {
	// Server
	Port        string
	Environment string

	// Database
	DatabaseURL string

	// Redis
	RedisAddr     string
	RedisPassword string

	// Kafka
	KafkaBrokers []string

	// External APIs
	LineChannelSecret      string
	LineChannelAccessToken string
	FacebookAppSecret      string
	FacebookAppID          string
	FacebookPageAccessToken string

	// Logging
	LogLevel  string
	LogFormat string

	// Service URLs
	OrderServiceURL     string
	InventoryServiceURL string

	// Authentication
	AdminToken string
}

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		// Server
		Port:        getEnv("PORT", "8090"),
		Environment: getEnv("GO_ENV", "development"),

		// Database
		DatabaseURL: getEnv("DATABASE_URL", "postgres://saan:saan_password@postgres:5432/saan_db?sslmode=disable"),

		// Redis
		RedisAddr:     getEnv("REDIS_ADDR", "redis:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		// Kafka
		KafkaBrokers: []string{getEnv("KAFKA_BROKERS", "kafka:9092")},

		// External APIs
		LineChannelSecret:       getEnv("LINE_CHANNEL_SECRET", ""),
		LineChannelAccessToken:  getEnv("LINE_CHANNEL_ACCESS_TOKEN", ""),
		FacebookAppSecret:       getEnv("FACEBOOK_APP_SECRET", ""),
		FacebookAppID:           getEnv("FACEBOOK_APP_ID", ""),
		FacebookPageAccessToken: getEnv("FACEBOOK_PAGE_ACCESS_TOKEN", ""),

		// Logging
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "text"),

		// Service URLs
		OrderServiceURL:     getEnv("ORDER_SERVICE_URL", "http://order:8081"),
		InventoryServiceURL: getEnv("INVENTORY_SERVICE_URL", "http://inventory:8082"),

		// Authentication
		AdminToken: getEnv("ADMIN_TOKEN", "saan-dev-admin-2024-secure"),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
