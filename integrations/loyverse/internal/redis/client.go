package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

// Config holds Redis client configuration
type Config struct {
	Addr            string
	Password        string
	DB              int
	MaxRetries      int
	MinRetryBackoff time.Duration
	MaxRetryBackoff time.Duration
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	PoolSize        int
	PoolTimeout     time.Duration
	IdleTimeout     time.Duration
}

// DefaultConfig returns default Redis configuration
func DefaultConfig() Config {
	return Config{
		Addr:            "localhost:6379",
		Password:        "",
		DB:              0,
		MaxRetries:      3,
		MinRetryBackoff: 100 * time.Millisecond,
		MaxRetryBackoff: 3 * time.Second,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		PoolSize:        10,
		PoolTimeout:     4 * time.Second,
		IdleTimeout:     5 * time.Minute,
	}
}

// Client wraps redis.Client with enhanced error handling
type Client struct {
	*redis.Client
	config       Config
	healthStatus bool
}

// NewClient creates a new Redis client with enhanced error handling
func NewClient(config Config) *Client {
	client := redis.NewClient(&redis.Options{
		Addr:            config.Addr,
		Password:        config.Password,
		DB:              config.DB,
		MaxRetries:      config.MaxRetries,
		MinRetryBackoff: config.MinRetryBackoff,
		MaxRetryBackoff: config.MaxRetryBackoff,
		DialTimeout:     config.DialTimeout,
		ReadTimeout:     config.ReadTimeout,
		WriteTimeout:    config.WriteTimeout,
		PoolSize:        config.PoolSize,
		PoolTimeout:     config.PoolTimeout,
		IdleTimeout:     config.IdleTimeout,
	})

	c := &Client{
		Client:       client,
		config:       config,
		healthStatus: true,
	}

	// Test initial connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Ping(ctx).Err(); err != nil {
		log.Printf("WARNING: Initial Redis connection failed: %v", err)
		c.healthStatus = false
	}

	return c
}

// IsHealthy returns the current health status of the Redis connection
func (c *Client) IsHealthy() bool {
	return c.healthStatus
}

// CheckHealth performs a health check and updates status
func (c *Client) CheckHealth(ctx context.Context) error {
	err := c.Ping(ctx).Err()
	c.healthStatus = err == nil
	return err
}

// SafeSet performs a Set operation with enhanced error handling
func (c *Client) SafeSet(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.withRetry(ctx, "SET", func() error {
		return c.Set(ctx, key, value, expiration).Err()
	})
}

// SafeGet performs a Get operation with enhanced error handling
func (c *Client) SafeGet(ctx context.Context, key string) (string, error) {
	var result string
	var err error
	
	retryErr := c.withRetry(ctx, "GET", func() error {
		result, err = c.Get(ctx, key).Result()
		return err
	})
	
	return result, retryErr
}

// SafeDel performs a Del operation with enhanced error handling
func (c *Client) SafeDel(ctx context.Context, keys ...string) error {
	return c.withRetry(ctx, "DEL", func() error {
		return c.Del(ctx, keys...).Err()
	})
}

// SafeExists performs an Exists operation with enhanced error handling
func (c *Client) SafeExists(ctx context.Context, keys ...string) (int64, error) {
	var result int64
	var err error
	
	retryErr := c.withRetry(ctx, "EXISTS", func() error {
		result, err = c.Exists(ctx, keys...).Result()
		return err
	})
	
	return result, retryErr
}

// withRetry executes a Redis operation with retry logic
func (c *Client) withRetry(ctx context.Context, operation string, fn func() error) error {
	var lastErr error
	
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate backoff duration
			backoff := c.config.MinRetryBackoff * time.Duration(1<<uint(attempt-1))
			if backoff > c.config.MaxRetryBackoff {
				backoff = c.config.MaxRetryBackoff
			}
			
			log.Printf("Redis %s operation failed, retrying in %v (attempt %d/%d): %v", 
				operation, backoff, attempt, c.config.MaxRetries, lastErr)
			
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during retry for %s operation: %w", operation, ctx.Err())
			}
		}
		
		lastErr = fn()
		if lastErr == nil {
			// Success - update health status
			c.healthStatus = true
			return nil
		}
		
		// Update health status on error
		c.healthStatus = false
		
		// Check if we should stop retrying
		if !c.shouldRetry(lastErr) {
			break
		}
	}
	
	return fmt.Errorf("Redis %s operation failed after %d attempts: %w", 
		operation, c.config.MaxRetries+1, lastErr)
}

// shouldRetry determines if an error is retryable
func (c *Client) shouldRetry(err error) bool {
	if err == nil {
		return false
	}
	
	// Don't retry on context cancellation
	if err == context.Canceled || err == context.DeadlineExceeded {
		return false
	}
	
	// Don't retry on NOAUTH errors
	if err.Error() == "NOAUTH Authentication required." {
		return false
	}
	
	// Don't retry on key not found (redis.Nil)
	if err == redis.Nil {
		return false
	}
	
	// Retry on network errors and other temporary failures
	return true
}

// GetStats returns Redis client statistics
func (c *Client) GetStats() *redis.PoolStats {
	return c.PoolStats()
}

// LogStats logs current Redis client statistics
func (c *Client) LogStats() {
	stats := c.GetStats()
	log.Printf("Redis Stats - Hits: %d, Misses: %d, Timeouts: %d, TotalConns: %d, IdleConns: %d, StaleConns: %d",
		stats.Hits, stats.Misses, stats.Timeouts, stats.TotalConns, stats.IdleConns, stats.StaleConns)
}

// Close closes the Redis client connection
func (c *Client) Close() error {
	c.healthStatus = false
	return c.Client.Close()
}
