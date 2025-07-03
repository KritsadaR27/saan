package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Environment string
	Port        string
	Database    DatabaseConfig
	Redis       RedisConfig
	Kafka       KafkaConfig
	Cache       CacheConfig
	External    ExternalConfig
	Security    SecurityConfig
	Logging     LoggingConfig
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
	MigrationsPath  string
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
	Topics  KafkaTopics
}

// KafkaTopics defines all Kafka topics
type KafkaTopics struct {
	ProductCreated string
	ProductUpdated string
	ProductDeleted string
	ProductSynced  string
	PriceChanged   string
	StockChanged   string
	InventoryLow   string
	InventoryAlert string
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	ProductTTL   int // seconds
	PriceTTL     int // seconds
	InventoryTTL int // seconds
	CategoryTTL  int // seconds
	SearchTTL    int // seconds
	StatsTTL     int // seconds
}

// ExternalConfig holds external service configuration
type ExternalConfig struct {
	LoyverseService     string
	OrderService        string
	InventoryService    string
	CustomerService     string
	LocationService     string
	NotificationService string
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	JWTSecret      string
	InternalAPIKey string
	RateLimitRPM   int
	AllowedOrigins []string
	EnableCORS     bool
	TrustedProxies []string
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
	File   string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// .env file is optional, so we don't return error
		fmt.Println("No .env file found, using environment variables")
	}

	config := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Port:        getEnv("PORT", "8083"),

		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "postgres"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "saan"),
			Password:        getEnv("DB_PASSWORD", "saan_password"),
			Database:        getEnv("DB_NAME", "saan_db"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvInt("DB_CONN_MAX_LIFETIME", 300),
			MigrationsPath:  getEnv("DB_MIGRATIONS_PATH", "file://migrations"),
		},

		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "redis"),
			Port:         getEnv("REDIS_PORT", "6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			Database:     getEnvInt("REDIS_DB", 0),
			MaxRetries:   getEnvInt("REDIS_MAX_RETRIES", 3),
			PoolSize:     getEnvInt("REDIS_POOL_SIZE", 10),
			MinIdleConns: getEnvInt("REDIS_MIN_IDLE_CONNS", 2),
		},

		Kafka: KafkaConfig{
			Brokers: strings.Split(getEnv("KAFKA_BROKERS", "kafka:9092"), ","),
			Topics: KafkaTopics{
				ProductCreated: getEnv("KAFKA_TOPIC_PRODUCT_CREATED", "product.created"),
				ProductUpdated: getEnv("KAFKA_TOPIC_PRODUCT_UPDATED", "product.updated"),
				ProductDeleted: getEnv("KAFKA_TOPIC_PRODUCT_DELETED", "product.deleted"),
				ProductSynced:  getEnv("KAFKA_TOPIC_PRODUCT_SYNCED", "product.synced"),
				PriceChanged:   getEnv("KAFKA_TOPIC_PRICE_CHANGED", "price.changed"),
				StockChanged:   getEnv("KAFKA_TOPIC_STOCK_CHANGED", "stock.changed"),
				InventoryLow:   getEnv("KAFKA_TOPIC_INVENTORY_LOW", "inventory.low"),
				InventoryAlert: getEnv("KAFKA_TOPIC_INVENTORY_ALERT", "inventory.alert"),
			},
		},

		Cache: CacheConfig{
			ProductTTL:   getEnvInt("CACHE_PRODUCT_TTL", 3600),  // 1 hour
			PriceTTL:     getEnvInt("CACHE_PRICE_TTL", 1800),    // 30 minutes
			InventoryTTL: getEnvInt("CACHE_INVENTORY_TTL", 300), // 5 minutes
			CategoryTTL:  getEnvInt("CACHE_CATEGORY_TTL", 7200), // 2 hours
			SearchTTL:    getEnvInt("CACHE_SEARCH_TTL", 600),    // 10 minutes
			StatsTTL:     getEnvInt("CACHE_STATS_TTL", 900),     // 15 minutes
		},

		External: ExternalConfig{
			LoyverseService:     getEnv("LOYVERSE_SERVICE_URL", "http://loyverse:8100"),
			OrderService:        getEnv("ORDER_SERVICE_URL", "http://order:8081"),
			InventoryService:    getEnv("INVENTORY_SERVICE_URL", "http://inventory:8082"),
			CustomerService:     getEnv("CUSTOMER_SERVICE_URL", "http://customer:8110"),
			LocationService:     getEnv("LOCATION_SERVICE_URL", "http://location:8090"),
			NotificationService: getEnv("NOTIFICATION_SERVICE_URL", "http://notification:8091"),
		},

		Security: SecurityConfig{
			JWTSecret:      getEnv("JWT_SECRET", "your-secret-key"),
			InternalAPIKey: getEnv("INTERNAL_API_KEY", "internal-api-key"),
			RateLimitRPM:   getEnvInt("RATE_LIMIT_RPM", 1000),
			AllowedOrigins: strings.Split(getEnv("ALLOWED_ORIGINS", "*"), ","),
			EnableCORS:     getEnvBool("ENABLE_CORS", true),
			TrustedProxies: strings.Split(getEnv("TRUSTED_PROXIES", ""), ","),
		},

		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
			File:   getEnv("LOG_FILE", ""),
		},
	}

	// Validate required configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// validateConfig validates that all required configuration is present
func validateConfig(config *Config) error {
	required := map[string]string{
		"Database Host":     config.Database.Host,
		"Database User":     config.Database.User,
		"Database Password": config.Database.Password,
		"Database Name":     config.Database.Database,
		"Redis Host":        config.Redis.Host,
		"JWT Secret":        config.Security.JWTSecret,
		"Internal API Key":  config.Security.InternalAPIKey,
	}

	for field, value := range required {
		if value == "" {
			return fmt.Errorf("%s is required", field)
		}
	}

	return nil
}

// GetDatabaseURL returns the database connection URL
func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Database,
		c.Database.SSLMode,
	)
}

// GetRedisURL returns the Redis connection URL
func (c *Config) GetRedisURL() string {
	if c.Redis.Password == "" {
		return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
	}
	return fmt.Sprintf("%s:%s@%s:%s", c.Redis.Password, c.Redis.Database, c.Redis.Host, c.Redis.Port)
}

// Helper functions for environment variable parsing
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

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
