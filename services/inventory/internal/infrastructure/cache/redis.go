package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"inventory/internal/config"
	"inventory/internal/domain"
)

// Cache key patterns following PROJECT_RULES.md
const (
	ProductKey         = "inventory:product:%s"        // inventory:product:{product_id}
	ProductListKey     = "inventory:products:%s"       // inventory:products:{filter_hash}
	StockLevelKey      = "inventory:stock:%s"          // inventory:stock:{product_id}
	LoyverseProductKey = "loyverse:product:%s"         // loyverse:product:{loyverse_id}
	AnalyticsKey       = "inventory:analytics:%s"      // inventory:analytics:{metric_key}
	CacheStatsKey      = "inventory:cache_stats:%s"    // inventory:cache_stats:{date}
)

// RedisClient implements enhanced Redis caching functionality
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

// NewRedisClientSimple creates a new Redis client with simple config (for backward compatibility)
func NewRedisClientSimple(addr, password string) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{
		client: client,
		logger: logrus.New(),
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

// ===== PRODUCT OPERATIONS =====

// GetProduct retrieves a single product from Redis cache
func (c *RedisClient) GetProduct(ctx context.Context, productID string) (*domain.Product, error) {
	key := fmt.Sprintf(ProductKey, productID)
	
	result, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		c.logger.WithError(err).WithField("key", key).Error("Redis GET failed")
		return nil, fmt.Errorf("failed to get product from cache: %w", err)
	}

	var product domain.Product
	if err := json.Unmarshal([]byte(result), &product); err != nil {
		c.logger.WithError(err).WithField("key", key).Error("Failed to unmarshal product")
		return nil, fmt.Errorf("failed to unmarshal product: %w", err)
	}

	return &product, nil
}

// SetProduct stores a single product in Redis cache with TTL
func (c *RedisClient) SetProduct(ctx context.Context, productID string, product *domain.Product, ttl time.Duration) error {
	key := fmt.Sprintf(ProductKey, productID)
	
	data, err := json.Marshal(product)
	if err != nil {
		c.logger.WithError(err).WithField("product_id", productID).Error("Failed to marshal product")
		return fmt.Errorf("failed to marshal product: %w", err)
	}

	err = c.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		c.logger.WithError(err).WithField("key", key).Error("Redis SET failed")
		return fmt.Errorf("failed to set product in cache: %w", err)
	}

	return nil
}

// DeleteProduct removes a product from cache
func (c *RedisClient) DeleteProduct(ctx context.Context, productID string) error {
	key := fmt.Sprintf(ProductKey, productID)
	
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		c.logger.WithError(err).WithField("key", key).Error("Redis DELETE failed")
		return fmt.Errorf("failed to delete product from cache: %w", err)
	}
	
	return nil
}

// ===== STOCK LEVEL OPERATIONS =====

// GetStockLevel retrieves stock level for a product
func (c *RedisClient) GetStockLevel(ctx context.Context, productID string) (int, error) {
	key := fmt.Sprintf(StockLevelKey, productID)
	
	result, err := c.client.Get(ctx, key).Int()
	if err != nil {
		if err == redis.Nil {
			return 0, nil // Default to 0 if not found
		}
		c.logger.WithError(err).WithField("key", key).Error("Redis GET failed")
		return 0, fmt.Errorf("failed to get stock level from cache: %w", err)
	}

	return result, nil
}

// SetStockLevel stores stock level for a product
func (c *RedisClient) SetStockLevel(ctx context.Context, productID string, level int, ttl time.Duration) error {
	key := fmt.Sprintf(StockLevelKey, productID)
	
	err := c.client.Set(ctx, key, level, ttl).Err()
	if err != nil {
		c.logger.WithError(err).WithField("key", key).Error("Redis SET failed")
		return fmt.Errorf("failed to set stock level in cache: %w", err)
	}

	return nil
}

// IncrementStock increments stock level for a product
func (c *RedisClient) IncrementStock(ctx context.Context, productID string, increment int) (int, error) {
	key := fmt.Sprintf(StockLevelKey, productID)
	
	result, err := c.client.IncrBy(ctx, key, int64(increment)).Result()
	if err != nil {
		c.logger.WithError(err).WithField("key", key).Error("Redis INCR failed")
		return 0, fmt.Errorf("failed to increment stock level: %w", err)
	}

	return int(result), nil
}

// ===== LOYVERSE OPERATIONS =====

// GetLoyverseProduct retrieves a Loyverse product from cache
func (c *RedisClient) GetLoyverseProduct(ctx context.Context, loyverseID string) (*domain.Product, error) {
	key := fmt.Sprintf(LoyverseProductKey, loyverseID)
	
	result, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		c.logger.WithError(err).WithField("key", key).Error("Redis GET failed")
		return nil, fmt.Errorf("failed to get Loyverse product from cache: %w", err)
	}

	var product domain.Product
	if err := json.Unmarshal([]byte(result), &product); err != nil {
		c.logger.WithError(err).WithField("key", key).Error("Failed to unmarshal Loyverse product")
		return nil, fmt.Errorf("failed to unmarshal Loyverse product: %w", err)
	}

	return &product, nil
}

// SetLoyverseProduct stores a Loyverse product in cache
func (c *RedisClient) SetLoyverseProduct(ctx context.Context, loyverseID string, product *domain.Product, ttl time.Duration) error {
	key := fmt.Sprintf(LoyverseProductKey, loyverseID)
	
	data, err := json.Marshal(product)
	if err != nil {
		c.logger.WithError(err).WithField("loyverse_id", loyverseID).Error("Failed to marshal Loyverse product")
		return fmt.Errorf("failed to marshal Loyverse product: %w", err)
	}

	err = c.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		c.logger.WithError(err).WithField("key", key).Error("Redis SET failed")
		return fmt.Errorf("failed to set Loyverse product in cache: %w", err)
	}

	return nil
}

// ===== GENERIC OPERATIONS =====

// Set stores a value with TTL
func (c *RedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	err = c.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		c.logger.WithError(err).WithField("key", key).Error("Redis SET failed")
		return fmt.Errorf("failed to set value in cache: %w", err)
	}

	return nil
}

// Get retrieves a value from cache
func (c *RedisClient) Get(ctx context.Context, key string) (string, error) {
	result, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // Cache miss
		}
		c.logger.WithError(err).WithField("key", key).Error("Redis GET failed")
		return "", fmt.Errorf("failed to get value from cache: %w", err)
	}

	return result, nil
}

// Delete removes a key from cache
func (c *RedisClient) Delete(ctx context.Context, key string) error {
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		c.logger.WithError(err).WithField("key", key).Error("Redis DELETE failed")
		return fmt.Errorf("failed to delete key from cache: %w", err)
	}

	return nil
}

// ===== PATTERN OPERATIONS =====

// DeletePattern removes all keys matching a pattern
func (c *RedisClient) DeletePattern(ctx context.Context, pattern string) error {
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys for pattern %s: %w", pattern, err)
	}

	if len(keys) > 0 {
		err = c.client.Del(ctx, keys...).Err()
		if err != nil {
			c.logger.WithError(err).WithField("pattern", pattern).Error("Redis batch DELETE failed")
			return fmt.Errorf("failed to delete keys with pattern %s: %w", pattern, err)
		}
	}

	return nil
}

// InvalidateProduct removes all cached data for a product
func (c *RedisClient) InvalidateProduct(ctx context.Context, productID string) error {
	patterns := []string{
		fmt.Sprintf(ProductKey, productID),
		fmt.Sprintf(StockLevelKey, productID),
		fmt.Sprintf(ProductListKey, "*"), // Invalidate all product lists
	}

	for _, pattern := range patterns {
		if err := c.DeletePattern(ctx, pattern); err != nil {
			c.logger.WithError(err).WithField("pattern", pattern).Warn("Failed to invalidate cache pattern")
		}
	}

	return nil
}
