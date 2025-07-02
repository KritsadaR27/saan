package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"integrations/loyverse/internal/redis"
)

// RedisRepository handles Redis operations with enhanced error handling
type RedisRepository struct {
	client  *redis.Client
	monitor *redis.HealthMonitor
}

// NewRedisRepository creates a new Redis repository with enhanced error handling
func NewRedisRepository(config redis.Config) *RedisRepository {
	client := redis.NewClient(config)
	monitor := redis.NewHealthMonitor(client, 30*time.Second)
	
	// Set up health change callback
	monitor.SetHealthChangeCallback(func(healthy bool) {
		if healthy {
			log.Println("Redis connection restored")
		} else {
			log.Println("Redis connection lost")
		}
	})
	
	// Start health monitoring in background
	go monitor.Start(context.Background())
	
	return &RedisRepository{
		client:  client,
		monitor: monitor,
	}
}

// Set stores a value in Redis with expiration and enhanced error handling
func (r *RedisRepository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if !r.monitor.IsHealthy() {
		return fmt.Errorf("Redis is unhealthy, skipping SET operation for key: %s", key)
	}
	
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshaling value for key %s: %w", key, err)
	}
	
	if err := r.client.SafeSet(ctx, key, data, expiration); err != nil {
		return fmt.Errorf("setting key %s: %w", key, err)
	}
	
	log.Printf("Successfully cached data for key: %s", key)
	return nil
}

// Get retrieves a value from Redis with enhanced error handling
func (r *RedisRepository) Get(ctx context.Context, key string, dest interface{}) error {
	if !r.monitor.IsHealthy() {
		return fmt.Errorf("Redis is unhealthy, skipping GET operation for key: %s", key)
	}
	
	val, err := r.client.SafeGet(ctx, key)
	if err != nil {
		return fmt.Errorf("getting key %s: %w", key, err)
	}
	
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("unmarshaling value for key %s: %w", key, err)
	}
	
	return nil
}

// Delete removes a key from Redis with enhanced error handling
func (r *RedisRepository) Delete(ctx context.Context, key string) error {
	if !r.monitor.IsHealthy() {
		log.Printf("Redis is unhealthy, skipping DELETE operation for key: %s", key)
		return nil // Don't fail the operation if Redis is down
	}
	
	if err := r.client.SafeDel(ctx, key); err != nil {
		return fmt.Errorf("deleting key %s: %w", key, err)
	}
	
	return nil
}

// Exists checks if a key exists in Redis with enhanced error handling
func (r *RedisRepository) Exists(ctx context.Context, key string) (bool, error) {
	if !r.monitor.IsHealthy() {
		log.Printf("Redis is unhealthy, assuming key does not exist: %s", key)
		return false, nil // Graceful degradation
	}
	
	result, err := r.client.SafeExists(ctx, key)
	if err != nil {
		return false, fmt.Errorf("checking existence of key %s: %w", key, err)
	}
	
	return result > 0, nil
}

// IsHealthy returns the current health status of Redis
func (r *RedisRepository) IsHealthy() bool {
	return r.monitor.IsHealthy()
}

// GetHealthStats returns health monitoring statistics
func (r *RedisRepository) GetHealthStats() map[string]interface{} {
	return map[string]interface{}{
		"healthy":             r.monitor.IsHealthy(),
		"consecutive_fails":   r.monitor.GetConsecutiveFailures(),
		"last_health_check":   r.monitor.GetLastHealthCheck(),
		"health_history":      r.monitor.GetHealthHistory(),
		"redis_pool_stats":    r.client.GetStats(),
	}
}

// Close closes the Redis connection and stops monitoring
func (r *RedisRepository) Close() error {
	r.monitor.Stop()
	return r.client.Close()
}