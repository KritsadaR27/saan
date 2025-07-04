package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds application configuration
type Config struct {
	// Loyverse API
	LoyverseAPIToken string
	WebhookSecret    string

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Kafka
	KafkaBrokers []string
	KafkaTopic   string

	// Sync intervals (duration)
	ProductSyncInterval   time.Duration
	InventorySyncInterval time.Duration
	ReceiptSyncInterval   time.Duration
	CustomerSyncInterval  time.Duration

	// Server
	Port       int
	AdminToken string
	TimeZone   string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		// Loyverse
		LoyverseAPIToken: getEnv("LOYVERSE_API_TOKEN", ""),
		WebhookSecret:    getEnv("LOYVERSE_WEBHOOK_SECRET", ""),

		// Redis
		RedisAddr:     getEnv("REDIS_ADDR", "redis:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),

		// Kafka
		KafkaBrokers: strings.Split(getEnv("KAFKA_BROKERS", "kafka:9092"), ","),
		KafkaTopic:   getEnv("KAFKA_TOPIC", "loyverse-events"),

		// Sync intervals (using time.Duration)
		ProductSyncInterval:   getEnvDuration("PRODUCT_SYNC_INTERVAL", 30*time.Minute),
		InventorySyncInterval: getEnvDuration("INVENTORY_SYNC_INTERVAL", 15*time.Minute),
		ReceiptSyncInterval:   getEnvDuration("RECEIPT_SYNC_INTERVAL", 5*time.Minute),
		CustomerSyncInterval:  getEnvDuration("CUSTOMER_SYNC_INTERVAL", 60*time.Minute),

		// Server
		Port:       getEnvInt("PORT", 8083),
		AdminToken: getEnv("ADMIN_TOKEN", "loyverse-admin-token-dev"),
		TimeZone:   getEnv("TZ", "Asia/Bangkok"),
	}

	// Validate required fields
	if cfg.LoyverseAPIToken == "" {
		return nil, fmt.Errorf("LOYVERSE_API_TOKEN is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
