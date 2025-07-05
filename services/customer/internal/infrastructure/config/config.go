package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	Kafka     KafkaConfig
	External  ExternalConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port        string
	Environment string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host         string
	Port         string
	Password     string
	Database     int
	MaxRetries   int
	PoolSize     int
	MinIdleConns int
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers []string
	Topic   string
}

// ExternalConfig holds external service configuration
type ExternalConfig struct {
	LoyverseAPIToken string
	LoyverseBaseURL  string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	maxRetries, _ := strconv.Atoi(getEnv("REDIS_MAX_RETRIES", "3"))
	poolSize, _ := strconv.Atoi(getEnv("REDIS_POOL_SIZE", "10"))
	minIdleConns, _ := strconv.Atoi(getEnv("REDIS_MIN_IDLE_CONNS", "5"))

	return &Config{
		Server: ServerConfig{
			Port:        getEnv("PORT", "8084"),
			Environment: getEnv("ENVIRONMENT", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "saan_customers"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnv("REDIS_PORT", "6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			Database:     redisDB,
			MaxRetries:   maxRetries,
			PoolSize:     poolSize,
			MinIdleConns: minIdleConns,
		},
		Kafka: KafkaConfig{
			Brokers: []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
			Topic:   getEnv("KAFKA_TOPIC", "customer-events"),
		},
		External: ExternalConfig{
			LoyverseAPIToken: getEnv("LOYVERSE_API_TOKEN", ""),
			LoyverseBaseURL:  getEnv("LOYVERSE_BASE_URL", "https://api.loyverse.com/v1.0"),
		},
	}, nil
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
