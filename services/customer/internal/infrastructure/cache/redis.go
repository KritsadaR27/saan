package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/saan-system/services/customer/internal/domain"
)

// RedisClient interface for Redis operations
type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Close() error
}

type cacheRepository struct {
	client RedisClient
}

// New creates a new Redis client
func New() (RedisClient, error) {
	addr := getEnv("REDIS_ADDR", "redis:6379")
	password := getEnv("REDIS_PASSWORD", "")
	db := 0

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}

// NewCacheRepository creates a new cache repository
func NewCacheRepository(client RedisClient) domain.CacheRepository {
	return &cacheRepository{client: client}
}

// GetCustomer retrieves a customer from cache
func (r *cacheRepository) GetCustomer(ctx context.Context, key string) (*domain.Customer, error) {
	result := r.client.Get(ctx, key)
	if result.Err() != nil {
		if result.Err() == redis.Nil {
			return nil, fmt.Errorf("customer not found in cache")
		}
		return nil, result.Err()
	}

	var customer domain.Customer
	if err := json.Unmarshal([]byte(result.Val()), &customer); err != nil {
		return nil, fmt.Errorf("failed to unmarshal customer: %w", err)
	}

	return &customer, nil
}

// SetCustomer stores a customer in cache
func (r *cacheRepository) SetCustomer(ctx context.Context, key string, customer *domain.Customer, ttl int) error {
	data, err := json.Marshal(customer)
	if err != nil {
		return fmt.Errorf("failed to marshal customer: %w", err)
	}

	expiration := time.Duration(ttl) * time.Second
	if err := r.client.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set customer in cache: %w", err)
	}

	return nil
}

// DeleteCustomer removes a customer from cache
func (r *cacheRepository) DeleteCustomer(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete customer from cache: %w", err)
	}
	return nil
}

// GetThaiAddresses retrieves Thai addresses from cache
func (r *cacheRepository) GetThaiAddresses(ctx context.Context, key string) ([]domain.ThaiAddress, error) {
	result := r.client.Get(ctx, key)
	if result.Err() != nil {
		if result.Err() == redis.Nil {
			return nil, fmt.Errorf("Thai addresses not found in cache")
		}
		return nil, result.Err()
	}

	var addresses []domain.ThaiAddress
	if err := json.Unmarshal([]byte(result.Val()), &addresses); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Thai addresses: %w", err)
	}

	return addresses, nil
}

// SetThaiAddresses stores Thai addresses in cache
func (r *cacheRepository) SetThaiAddresses(ctx context.Context, key string, addresses []domain.ThaiAddress, ttl int) error {
	data, err := json.Marshal(addresses)
	if err != nil {
		return fmt.Errorf("failed to marshal Thai addresses: %w", err)
	}

	expiration := time.Duration(ttl) * time.Second
	if err := r.client.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set Thai addresses in cache: %w", err)
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
