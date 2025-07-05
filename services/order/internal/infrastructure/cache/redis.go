package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"order/internal/infrastructure/config"
)

// Cache key patterns following PROJECT_RULES.md
const (
	OrderKey         = "order:order:%s"         // order:order:{order_id}
	CustomerOrderKey = "order:customer:%s"      // order:customer:{customer_id}
	OrderStatsKey    = "order:stats:%s"         // order:stats:{date}
	OrderCacheKey    = "order:cache:%s"         // order:cache:{cache_key}
)

// RedisClient implements Redis caching functionality for orders
type RedisClient struct {
	client *redis.Client
	logger *logrus.Logger
}

// NewRedisClient creates a new Redis client with configuration
func NewRedisClient(cfg config.RedisConfig, logger *logrus.Logger) (*RedisClient, error) {
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

	return &RedisClient{
		client: client,
		logger: logger,
	}, nil
}

// Close closes the Redis connection
func (c *RedisClient) Close() error {
	return c.client.Close()
}

// Ping checks the Redis connection health
func (c *RedisClient) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Health checks the Redis connection health (alias for Ping)
func (c *RedisClient) Health(ctx context.Context) error {
	return c.Ping(ctx)
}

// GetClient returns the underlying Redis client
func (c *RedisClient) GetClient() *redis.Client {
	return c.client
}

// ===== ORDER OPERATIONS =====

// GetOrder retrieves a single order from Redis cache
func (c *RedisClient) GetOrder(ctx context.Context, orderID string) ([]byte, error) {
	key := fmt.Sprintf(OrderKey, orderID)
	
	result, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		c.logger.WithError(err).WithField("key", key).Error("Redis GET failed")
		return nil, fmt.Errorf("failed to get order from cache: %w", err)
	}

	return []byte(result), nil
}

// SetOrder stores a single order in Redis cache with TTL
func (c *RedisClient) SetOrder(ctx context.Context, orderID string, data interface{}, ttl time.Duration) error {
	key := fmt.Sprintf(OrderKey, orderID)
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		c.logger.WithError(err).WithField("order_id", orderID).Error("Failed to marshal order")
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	err = c.client.Set(ctx, key, jsonData, ttl).Err()
	if err != nil {
		c.logger.WithError(err).WithField("key", key).Error("Redis SET failed")
		return fmt.Errorf("failed to set order in cache: %w", err)
	}

	return nil
}

// DeleteOrder removes an order from cache
func (c *RedisClient) DeleteOrder(ctx context.Context, orderID string) error {
	key := fmt.Sprintf(OrderKey, orderID)
	
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		c.logger.WithError(err).WithField("key", key).Error("Redis DELETE failed")
		return fmt.Errorf("failed to delete order from cache: %w", err)
	}
	
	return nil
}

// ===== CUSTOMER ORDER OPERATIONS =====

// GetCustomerOrders retrieves customer orders from cache
func (c *RedisClient) GetCustomerOrders(ctx context.Context, customerID string) ([]string, error) {
	key := fmt.Sprintf(CustomerOrderKey, customerID)
	
	result, err := c.client.SMembers(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return []string{}, nil // Cache miss
		}
		c.logger.WithError(err).WithField("key", key).Error("Redis SMEMBERS failed")
		return nil, fmt.Errorf("failed to get customer orders from cache: %w", err)
	}

	return result, nil
}

// AddCustomerOrder adds an order to customer's order set
func (c *RedisClient) AddCustomerOrder(ctx context.Context, customerID, orderID string, ttl time.Duration) error {
	key := fmt.Sprintf(CustomerOrderKey, customerID)
	
	err := c.client.SAdd(ctx, key, orderID).Err()
	if err != nil {
		c.logger.WithError(err).WithField("key", key).Error("Redis SADD failed")
		return fmt.Errorf("failed to add customer order to cache: %w", err)
	}

	// Set TTL if provided
	if ttl > 0 {
		c.client.Expire(ctx, key, ttl)
	}

	return nil
}

// ===== GENERIC CACHE OPERATIONS =====

// Set stores data with TTL
func (c *RedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

// Get retrieves data
func (c *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Delete removes data
func (c *RedisClient) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Exists checks if key exists
func (c *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.client.Exists(ctx, key).Result()
	return result > 0, err
}
