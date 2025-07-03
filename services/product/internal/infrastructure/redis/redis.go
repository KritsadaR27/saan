package redis

import (
	"context"
	"fmt"
	"time"

	"product-service/internal/infrastructure/config"

	"github.com/redis/go-redis/v9"
)

// Client is the global Redis client
var Client *redis.Client

// Initialize initializes the Redis connection
func Initialize(cfg config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
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

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	Client = rdb
	return rdb, nil
}

// GetClient returns the Redis client
func GetClient() *redis.Client {
	return Client
}

// Health checks the Redis connection health
func Health(ctx context.Context, client *redis.Client) error {
	return client.Ping(ctx).Err()
}
