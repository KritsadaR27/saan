package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache implements the cache interface using Redis
type Cache struct {
	client *redis.Client
	prefix string
}

// NewCache creates a new Redis cache instance
func NewCache(client *redis.Client, prefix string) *Cache {
	return &Cache{
		client: client,
		prefix: prefix,
	}
}

// NewRedisClient creates a new Redis client and Cache instance
func NewRedisClient(redisURL string) (*Cache, error) {
	// Parse Redis URL and create client
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opt)

	// Test connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return NewCache(client, "shipping:"), nil
}

// Get retrieves a value from cache
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	fullKey := c.getFullKey(key)

	val, err := c.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key not found: %s", key)
		}
		return "", fmt.Errorf("failed to get from cache: %w", err)
	}

	return val, nil
}

// Set stores a value in cache with TTL
func (c *Cache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	fullKey := c.getFullKey(key)

	err := c.client.Set(ctx, fullKey, value, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// Delete removes a value from cache
func (c *Cache) Delete(ctx context.Context, key string) error {
	fullKey := c.getFullKey(key)

	err := c.client.Del(ctx, fullKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}

	return nil
}

// SetJSON stores a JSON-serializable object in cache
func (c *Cache) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return c.Set(ctx, key, string(jsonData), ttl)
}

// GetJSON retrieves and deserializes a JSON object from cache
func (c *Cache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	jsonStr, err := c.Get(ctx, key)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(jsonStr), dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

// Exists checks if a key exists in cache
func (c *Cache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := c.getFullKey(key)

	count, err := c.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}

	return count > 0, nil
}

// SetExpire sets expiration for an existing key
func (c *Cache) SetExpire(ctx context.Context, key string, ttl time.Duration) error {
	fullKey := c.getFullKey(key)

	err := c.client.Expire(ctx, fullKey, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set expiration: %w", err)
	}

	return nil
}

// GetTTL gets the remaining time to live for a key
func (c *Cache) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	fullKey := c.getFullKey(key)

	ttl, err := c.client.TTL(ctx, fullKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}

	return ttl, nil
}

// Increment increments the integer value of a key
func (c *Cache) Increment(ctx context.Context, key string) (int64, error) {
	fullKey := c.getFullKey(key)

	val, err := c.client.Incr(ctx, fullKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment: %w", err)
	}

	return val, nil
}

// IncrementBy increments the integer value of a key by the given amount
func (c *Cache) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	fullKey := c.getFullKey(key)

	val, err := c.client.IncrBy(ctx, fullKey, value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment by: %w", err)
	}

	return val, nil
}

// getFullKey returns the full key with prefix
func (c *Cache) getFullKey(key string) string {
	if c.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", c.prefix, key)
}

// FlushPattern deletes all keys matching a pattern
func (c *Cache) FlushPattern(ctx context.Context, pattern string) error {
	fullPattern := c.getFullKey(pattern)

	keys, err := c.client.Keys(ctx, fullPattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys by pattern: %w", err)
	}

	if len(keys) > 0 {
		err = c.client.Del(ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete keys: %w", err)
		}
	}

	return nil
}

// Health checks if Redis connection is healthy
func (c *Cache) Health(ctx context.Context) error {
	_, err := c.client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("redis health check failed: %w", err)
	}

	return nil
}
