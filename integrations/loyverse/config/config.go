// integrations/loyverse/config/config.go
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds application configuration
type Config struct {
	// Loyverse API
	LoyverseAPIToken string

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Kafka
	KafkaBrokers []string
	KafkaTopic   string

	// Sync intervals
	ProductSyncInterval   string
	InventorySyncInterval string
	ReceiptSyncInterval   string
	CustomerSyncInterval  string
	
	// Extended sync intervals
	EmployeeSyncInterval    string
	CategorySyncInterval    string
	SupplierSyncInterval    string
	PaymentTypeSyncInterval string
	StoreSyncInterval       string

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

		// Redis
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),

		// Kafka
		KafkaBrokers: strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ","),
		KafkaTopic:   getEnv("KAFKA_TOPIC", "loyverse-events"),

		// Sync intervals (cron format)
		ProductSyncInterval:   getEnv("PRODUCT_SYNC_INTERVAL", "*/30 * * * *"),
		InventorySyncInterval: getEnv("INVENTORY_SYNC_INTERVAL", "*/15 * * * *"),
		ReceiptSyncInterval:   getEnv("RECEIPT_SYNC_INTERVAL", "*/5 * * * *"),
		CustomerSyncInterval:  getEnv("CUSTOMER_SYNC_INTERVAL", "0 * * * *"),
		
		// Extended sync intervals (from environment variables)
		EmployeeSyncInterval:    getEnv("EMPLOYEE_SYNC_INTERVAL", "0 */12 * * *"),
		CategorySyncInterval:    getEnv("CATEGORY_SYNC_INTERVAL", "0 */12 * * *"),
		SupplierSyncInterval:    getEnv("SUPPLIER_SYNC_INTERVAL", "0 */12 * * *"),
		PaymentTypeSyncInterval: getEnv("PAYMENT_TYPE_SYNC_INTERVAL", "0 */12 * * *"),
		StoreSyncInterval:       getEnv("STORE_SYNC_INTERVAL", "0 */12 * * *"),

		// Server (note: webhook service now uses port 8093)
		Port:       getEnvInt("PORT", 8084),
		AdminToken: getEnv("ADMIN_TOKEN", ""),
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
