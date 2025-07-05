package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Kafka    KafkaConfig
	External ExternalConfig
	Logging  LoggingConfig
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
	URL      string // For backward compatibility
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
	Brokers       []string
	ConsumerGroup string
	Topics        KafkaTopics
}

// KafkaTopics holds Kafka topic configuration
type KafkaTopics struct {
	LoyverseEvents string
	InventoryEvents string
}

// ExternalConfig holds external service configuration
type ExternalConfig struct {
	OrderServiceURL     string
	ChatServiceURL      string
	LoyverseAPIToken    string
	AdminToken          string
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// Load loads configuration from environment variables
func Load() *Config {
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	maxRetries, _ := strconv.Atoi(getEnv("REDIS_MAX_RETRIES", "3"))
	poolSize, _ := strconv.Atoi(getEnv("REDIS_POOL_SIZE", "10"))
	minIdleConns, _ := strconv.Atoi(getEnv("REDIS_MIN_IDLE_CONNS", "5"))

	return &Config{
		Server: ServerConfig{
			Port:        getEnv("PORT", "8082"),
			Environment: getEnv("GO_ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "saan"),
			Password: getEnv("DB_PASSWORD", "saan_password"),
			Name:     getEnv("DB_NAME", "saan_db"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
			URL:      getEnv("DATABASE_URL", "postgres://saan:saan_password@postgres:5432/saan_db?sslmode=disable"),
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
			Brokers:       []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
			ConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "inventory-service"),
			Topics: KafkaTopics{
				LoyverseEvents:  getEnv("LOYVERSE_EVENT_TOPIC", "loyverse-events"),
				InventoryEvents: getEnv("INVENTORY_EVENT_TOPIC", "inventory-events"),
			},
		},
		External: ExternalConfig{
			OrderServiceURL:  getEnv("ORDER_SERVICE_URL", "http://localhost:8081"),
			ChatServiceURL:   getEnv("CHAT_SERVICE_URL", "http://localhost:8083"),
			LoyverseAPIToken: getEnv("LOYVERSE_API_TOKEN", ""),
			AdminToken:       getEnv("ADMIN_TOKEN", ""),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
