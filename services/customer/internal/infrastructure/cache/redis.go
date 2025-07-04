package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/saan-system/services/customer/internal/domain/entity"
	"github.com/saan-system/services/customer/internal/domain/repository"
)

// redisCache implements repository.CacheRepository
type redisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(addr string) (repository.CacheRepository, error) {
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

	return &redisCache{client: client}, nil
}

// GetCustomer retrieves a customer from cache
func (c *redisCache) GetCustomer(ctx context.Context, key string) (*entity.Customer, error) {
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var customer entity.Customer
	if err := json.Unmarshal([]byte(data), &customer); err != nil {
		return nil, err
	}

	return &customer, nil
}

// SetCustomer stores a customer in cache
func (c *redisCache) SetCustomer(ctx context.Context, key string, customer *entity.Customer, ttl int) error {
	data, err := json.Marshal(customer)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, time.Duration(ttl)*time.Second).Err()
}

// DeleteCustomer removes a customer from cache
func (c *redisCache) DeleteCustomer(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// GetThaiAddresses retrieves Thai addresses from cache
func (c *redisCache) GetThaiAddresses(ctx context.Context, key string) ([]entity.ThaiAddress, error) {
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var addresses []entity.ThaiAddress
	if err := json.Unmarshal([]byte(data), &addresses); err != nil {
		return nil, err
	}

	return addresses, nil
}

// SetThaiAddresses stores Thai addresses in cache
func (c *redisCache) SetThaiAddresses(ctx context.Context, key string, addresses []entity.ThaiAddress, ttl int) error {
	data, err := json.Marshal(addresses)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, time.Duration(ttl)*time.Second).Err()
}

// Close closes the Redis connection
func (c *redisCache) Close() error {
	return c.client.Close()
}
