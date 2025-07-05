package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Redis    RedisConfig    `json:"redis"`
	Kafka    KafkaConfig    `json:"kafka"`
	Logging  LoggingConfig  `json:"logging"`
	External ExternalConfig `json:"external"`
	Loyverse LoyverseConfig `json:"loyverse"`
}

type ServerConfig struct {
	Host        string `json:"host"`
	Port        string `json:"port"`
	Environment string `json:"environment"`
}

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
	SSLMode  string `json:"ssl_mode"`
}

type RedisConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

type KafkaConfig struct {
	Brokers []string `json:"brokers"`
	GroupID string   `json:"group_id"`
	Topics  TopicsConfig `json:"topics"`
}

type TopicsConfig struct {
	PaymentEvents string `json:"payment_events"`
	OrderEvents   string `json:"order_events"`
	DeliveryEvents string `json:"delivery_events"`
}

type LoggingConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

type ExternalConfig struct {
	OrderServiceURL        string `json:"order_service_url"`
	CustomerServiceURL     string `json:"customer_service_url"`
	ShippingServiceURL     string `json:"shipping_service_url"`
	NotificationServiceURL string `json:"notification_service_url"`
}

type LoyverseConfig struct {
	APIKey    string `json:"api_key"`
	BaseURL   string `json:"base_url"`
	Timeout   int    `json:"timeout"`
	RetryCount int   `json:"retry_count"`
}

func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:        getEnv("SERVER_HOST", "0.0.0.0"),
			Port:        getEnv("SERVER_PORT", "8005"),
			Environment: getEnv("ENVIRONMENT", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "payment_service"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Kafka: KafkaConfig{
			Brokers: getEnvAsSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
			GroupID: getEnv("KAFKA_GROUP_ID", "payment-service"),
			Topics: TopicsConfig{
				PaymentEvents:  getEnv("KAFKA_TOPIC_PAYMENT_EVENTS", "payment.events"),
				OrderEvents:    getEnv("KAFKA_TOPIC_ORDER_EVENTS", "order.events"),
				DeliveryEvents: getEnv("KAFKA_TOPIC_DELIVERY_EVENTS", "delivery.events"),
			},
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		External: ExternalConfig{
			OrderServiceURL:        getEnv("ORDER_SERVICE_URL", "http://localhost:8001"),
			CustomerServiceURL:     getEnv("CUSTOMER_SERVICE_URL", "http://localhost:8002"),
			ShippingServiceURL:     getEnv("SHIPPING_SERVICE_URL", "http://localhost:8006"),
			NotificationServiceURL: getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8007"),
		},
		Loyverse: LoyverseConfig{
			APIKey:     getEnv("LOYVERSE_API_KEY", ""),
			BaseURL:    getEnv("LOYVERSE_BASE_URL", "https://api.loyverse.com/v1.0"),
			Timeout:    getEnvAsInt("LOYVERSE_TIMEOUT", 30),
			RetryCount: getEnvAsInt("LOYVERSE_RETRY_COUNT", 3),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
