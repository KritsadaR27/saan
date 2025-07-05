package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds the application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Kafka    KafkaConfig
	External ExternalConfig
	Logging  LoggingConfig
	JWT      JWTConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port        string
	Host        string
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
	OrderEvents     string
	PaymentEvents   string
	InventoryEvents string
	NotificationEvents string
}

// ExternalConfig holds external service configuration
type ExternalConfig struct {
	InventoryServiceURL   string
	CustomerServiceURL    string
	PaymentServiceURL     string
	NotificationServiceURL string
	UserServiceURL        string
	DeliveryServiceURL    string
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	maxRetries, _ := strconv.Atoi(getEnv("REDIS_MAX_RETRIES", "3"))
	poolSize, _ := strconv.Atoi(getEnv("REDIS_POOL_SIZE", "10"))
	minIdleConns, _ := strconv.Atoi(getEnv("REDIS_MIN_IDLE_CONNS", "5"))

	return &Config{
		Server: ServerConfig{
			Port:        getEnv("PORT", "8081"),
			Host:        getEnv("SERVER_HOST", "0.0.0.0"),
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
			Brokers:       strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ","),
			ConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "order-service"),
			Topics: KafkaTopics{
				OrderEvents:        getEnv("ORDER_EVENT_TOPIC", "order-events"),
				PaymentEvents:      getEnv("PAYMENT_EVENT_TOPIC", "payment-events"),
				InventoryEvents:    getEnv("INVENTORY_EVENT_TOPIC", "inventory-events"),
				NotificationEvents: getEnv("NOTIFICATION_EVENT_TOPIC", "notification-events"),
			},
		},
		External: ExternalConfig{
			InventoryServiceURL:    getEnv("INVENTORY_SERVICE_URL", "http://inventory-service:8082"),
			CustomerServiceURL:     getEnv("CUSTOMER_SERVICE_URL", "http://customer-service:8084"),
			PaymentServiceURL:      getEnv("PAYMENT_SERVICE_URL", "http://payment-service:8085"),
			NotificationServiceURL: getEnv("NOTIFICATION_SERVICE_URL", "http://notification-service:8092"),
			UserServiceURL:         getEnv("USER_SERVICE_URL", "http://user-service:8088"),
			DeliveryServiceURL:     getEnv("DELIVERY_SERVICE_URL", "http://delivery-service:8089"),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "default-secret-key-for-development"),
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
