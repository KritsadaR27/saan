package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"customer/internal/domain/entity"
	"customer/internal/domain/repository"
	"customer/internal/infrastructure/config"
)

// Cache key patterns following PROJECT_RULES.md
const (
	CustomerKey        = "customer:hot:%s"           // customer:hot:{customer_id}
	CustomerListKey    = "customer:list:%s"          // customer:list:{filter_hash}
	CustomerSearchKey  = "customer:search:%s"        // customer:search:{query_hash}
	ThaiAddressKey     = "address:thai:%s"           // address:thai:{province_code}
	DeliveryRouteKey   = "delivery:route:%s"         // delivery:route:{route_id}
	CustomerTierKey    = "customer:tier:%s"          // customer:tier:{customer_id}
	CustomerPointsKey  = "customer:points:%s"        // customer:points:{customer_id}
	AnalyticsKey       = "analytics:%s"               // analytics:{metric_key}
)

// RedisCache implements repository.CacheRepository
type RedisCache struct {
	client *redis.Client
	logger *zap.Logger
}

// NewRedisCache creates a new Redis cache instance with configuration
func NewRedisCache(cfg config.RedisConfig, logger *zap.Logger) (repository.CacheRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.Database,
		MaxRetries:   cfg.MaxRetries,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: client,
		logger: logger,
	}, nil
}

// NewRedisCacheSimple creates a new Redis cache instance with simple config (for backward compatibility)
func NewRedisCacheSimple(addr string) (repository.CacheRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password for development
		DB:       0,  // default DB
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	logger, _ := zap.NewProduction()
	return &RedisCache{
		client: client,
		logger: logger,
	}, nil
}

// GetCustomer retrieves a customer from cache
func (c *RedisCache) GetCustomer(ctx context.Context, key string) (*entity.Customer, error) {
	cacheKey := fmt.Sprintf(CustomerKey, key)
	data, err := c.client.Get(ctx, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		c.logger.Error("Redis GET failed", zap.Error(err), zap.String("key", cacheKey))
		return nil, fmt.Errorf("failed to get customer from cache: %w", err)
	}

	var customer entity.Customer
	if err := json.Unmarshal([]byte(data), &customer); err != nil {
		c.logger.Error("Failed to unmarshal customer", zap.Error(err), zap.String("key", cacheKey))
		return nil, fmt.Errorf("failed to unmarshal customer: %w", err)
	}

	return &customer, nil
}

// SetCustomer stores a customer in cache
func (c *RedisCache) SetCustomer(ctx context.Context, key string, customer *entity.Customer, ttl int) error {
	cacheKey := fmt.Sprintf(CustomerKey, key)
	data, err := json.Marshal(customer)
	if err != nil {
		c.logger.Error("Failed to marshal customer", zap.Error(err), zap.String("key", cacheKey))
		return fmt.Errorf("failed to marshal customer: %w", err)
	}

	err = c.client.Set(ctx, cacheKey, data, time.Duration(ttl)*time.Second).Err()
	if err != nil {
		c.logger.Error("Redis SET failed", zap.Error(err), zap.String("key", cacheKey))
		return fmt.Errorf("failed to set customer in cache: %w", err)
	}

	return nil
}

// DeleteCustomer removes a customer from cache
func (c *RedisCache) DeleteCustomer(ctx context.Context, key string) error {
	cacheKey := fmt.Sprintf(CustomerKey, key)
	err := c.client.Del(ctx, cacheKey).Err()
	if err != nil {
		c.logger.Error("Redis DELETE failed", zap.Error(err), zap.String("key", cacheKey))
		return fmt.Errorf("failed to delete customer from cache: %w", err)
	}
	return nil
}

// GetThaiAddresses retrieves Thai addresses from cache
func (c *RedisCache) GetThaiAddresses(ctx context.Context, key string) ([]entity.ThaiAddress, error) {
	cacheKey := fmt.Sprintf(ThaiAddressKey, key)
	data, err := c.client.Get(ctx, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		c.logger.Error("Redis GET failed", zap.Error(err), zap.String("key", cacheKey))
		return nil, fmt.Errorf("failed to get Thai addresses from cache: %w", err)
	}

	var addresses []entity.ThaiAddress
	if err := json.Unmarshal([]byte(data), &addresses); err != nil {
		c.logger.Error("Failed to unmarshal Thai addresses", zap.Error(err), zap.String("key", cacheKey))
		return nil, fmt.Errorf("failed to unmarshal Thai addresses: %w", err)
	}

	return addresses, nil
}

// SetThaiAddresses stores Thai addresses in cache
func (c *RedisCache) SetThaiAddresses(ctx context.Context, key string, addresses []entity.ThaiAddress, ttl int) error {
	cacheKey := fmt.Sprintf(ThaiAddressKey, key)
	data, err := json.Marshal(addresses)
	if err != nil {
		c.logger.Error("Failed to marshal Thai addresses", zap.Error(err), zap.String("key", cacheKey))
		return fmt.Errorf("failed to marshal Thai addresses: %w", err)
	}

	err = c.client.Set(ctx, cacheKey, data, time.Duration(ttl)*time.Second).Err()
	if err != nil {
		c.logger.Error("Redis SET failed", zap.Error(err), zap.String("key", cacheKey))
		return fmt.Errorf("failed to set Thai addresses in cache: %w", err)
	}

	return nil
}

// Health checks the Redis connection health
func (c *RedisCache) Health(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// GetClient returns the underlying Redis client (for advanced operations)
func (c *RedisCache) GetClient() *redis.Client {
	return c.client
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}
