package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"product-service/internal/domain/entity"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// RedisCache implements the cache repository interface
type RedisCache struct {
	client *redis.Client
	logger *logrus.Logger
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(host, port, password string, database int, logger *logrus.Logger) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       database,
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

// Cache Keys following PROJECT_RULES.md patterns
const (
	// Product caching patterns
	ProductKey             = "product:hot:%s"                    // product:hot:{product_id}
	ProductListKey         = "product:list:%s"                   // product:list:{filter_hash}
	ProductFeaturedKey     = "product:featured"                  // featured products
	ProductCategoryKey     = "product:category:%s"               // product:category:{category_id}
	ProductSearchKey       = "product:search:%s"                 // product:search:{query_hash}

	// Pricing cache patterns
	PricingCalculationKey  = "pricing:calculation:%s:%s"         // pricing:calculation:{customer_id}:{product_id}
	PricingTiersKey        = "pricing:tiers:%s"                  // pricing:tiers:{product_id}
	VIPBenefitsKey         = "vip:benefits:%s"                   // vip:benefits:{vip_level}

	// Inventory cache patterns
	InventoryLevelsKey     = "inventory:levels:%s"               // inventory:levels:{product_id}
	AvailabilityKey        = "product:availability:%s"           // product:availability:{product_id}

	// Analytics & Metrics
	MetricsDailyOrdersKey  = "metrics:daily:orders:%s"           // metrics:daily:orders:{date}
	MetricsDailyRevenueKey = "metrics:daily:revenue:%s"          // metrics:daily:revenue:{date}
	AnalyticsTrendsKey     = "analytics:trends:%s"               // analytics:trends:{category}
	DashboardStatsKey      = "dashboard:stats:%s"                // dashboard:stats:{date}

	// Session & Authentication (for future use)
	UserSessionKey         = "user:session:%s"                   // user:session:{session_id}
	APIRateLimitKey        = "api:rate_limit:%s"                 // api:rate_limit:{user_id}
)

// Basic cache operations
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	err = r.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		r.logger.WithError(err).WithField("key", key).Error("Redis SET failed")
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

func (r *RedisCache) Get(ctx context.Context, key string) (interface{}, error) {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		r.logger.WithError(err).WithField("key", key).Error("Redis GET failed")
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}

	var result interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached value: %w", err)
	}

	return result, nil
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		r.logger.WithError(err).WithField("key", key).Error("Redis DELETE failed")
		return fmt.Errorf("failed to delete cache: %w", err)
	}
	return nil
}

func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check cache existence: %w", err)
	}
	return count > 0, nil
}

// Batch operations
func (r *RedisCache) SetBatch(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	pipe := r.client.Pipeline()
	defer pipe.Close()

	for key, value := range items {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
		}
		pipe.Set(ctx, key, data, ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		r.logger.WithError(err).Error("Redis batch SET failed")
		return fmt.Errorf("failed to execute batch set: %w", err)
	}

	return nil
}

func (r *RedisCache) GetBatch(ctx context.Context, keys []string) (map[string]interface{}, error) {
	pipe := r.client.Pipeline()
	defer pipe.Close()

	for _, key := range keys {
		pipe.Get(ctx, key)
	}

	cmders, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		r.logger.WithError(err).Error("Redis batch GET failed")
		return nil, fmt.Errorf("failed to execute batch get: %w", err)
	}

	result := make(map[string]interface{})
	for i, cmder := range cmders {
		cmd := cmder.(*redis.StringCmd)
		data, err := cmd.Result()
		if err == redis.Nil {
			continue // Skip cache misses
		}
		if err != nil {
			continue // Skip errors
		}

		var value interface{}
		if err := json.Unmarshal([]byte(data), &value); err == nil {
			result[keys[i]] = value
		}
	}

	return result, nil
}

func (r *RedisCache) DeleteBatch(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	err := r.client.Del(ctx, keys...).Err()
	if err != nil {
		r.logger.WithError(err).Error("Redis batch DELETE failed")
		return fmt.Errorf("failed to delete batch: %w", err)
	}
	return nil
}

// Pattern operations
func (r *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys for pattern %s: %w", pattern, err)
	}

	if len(keys) > 0 {
		return r.DeleteBatch(ctx, keys)
	}

	return nil
}

func (r *RedisCache) GetKeys(ctx context.Context, pattern string) ([]string, error) {
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys for pattern %s: %w", pattern, err)
	}
	return keys, nil
}

// Product-specific cache operations following PROJECT_RULES.md
func (r *RedisCache) SetProduct(ctx context.Context, productID uuid.UUID, product *entity.Product, ttl time.Duration) error {
	key := fmt.Sprintf(ProductKey, productID.String())
	return r.Set(ctx, key, product, ttl)
}

func (r *RedisCache) GetProduct(ctx context.Context, productID uuid.UUID) (*entity.Product, error) {
	key := fmt.Sprintf(ProductKey, productID.String())
	data, err := r.Get(ctx, key)
	if err != nil || data == nil {
		return nil, err
	}

	// Re-marshal and unmarshal to get proper type
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var product entity.Product
	if err := json.Unmarshal(jsonData, &product); err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *RedisCache) SetProductList(ctx context.Context, key string, products []*entity.Product, ttl time.Duration) error {
	cacheKey := fmt.Sprintf(ProductListKey, key)
	return r.Set(ctx, cacheKey, products, ttl)
}

func (r *RedisCache) GetProductList(ctx context.Context, key string) ([]*entity.Product, error) {
	cacheKey := fmt.Sprintf(ProductListKey, key)
	data, err := r.Get(ctx, cacheKey)
	if err != nil || data == nil {
		return nil, err
	}

	// Re-marshal and unmarshal to get proper type
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var products []*entity.Product
	if err := json.Unmarshal(jsonData, &products); err != nil {
		return nil, err
	}

	return products, nil
}

// Price cache operations
func (r *RedisCache) SetPrice(ctx context.Context, productID uuid.UUID, priceCalc *entity.PriceCalculation, ttl time.Duration) error {
	key := fmt.Sprintf(PricingCalculationKey, priceCalc.ProductID.String(), productID.String())
	return r.Set(ctx, key, priceCalc, ttl)
}

func (r *RedisCache) GetPrice(ctx context.Context, productID uuid.UUID) (*entity.PriceCalculation, error) {
	// For simple lookup, use product ID for both parameters
	key := fmt.Sprintf(PricingCalculationKey, productID.String(), productID.String())
	data, err := r.Get(ctx, key)
	if err != nil || data == nil {
		return nil, err
	}

	// Re-marshal and unmarshal to get proper type
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var priceCalc entity.PriceCalculation
	if err := json.Unmarshal(jsonData, &priceCalc); err != nil {
		return nil, err
	}

	return &priceCalc, nil
}

// Inventory cache operations
func (r *RedisCache) SetAvailability(ctx context.Context, productID uuid.UUID, availability *entity.ProductAvailability, ttl time.Duration) error {
	key := fmt.Sprintf(AvailabilityKey, productID.String())
	return r.Set(ctx, key, availability, ttl)
}

func (r *RedisCache) GetAvailability(ctx context.Context, productID uuid.UUID) (*entity.ProductAvailability, error) {
	key := fmt.Sprintf(AvailabilityKey, productID.String())
	data, err := r.Get(ctx, key)
	if err != nil || data == nil {
		return nil, err
	}

	// Re-marshal and unmarshal to get proper type
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var availability entity.ProductAvailability
	if err := json.Unmarshal(jsonData, &availability); err != nil {
		return nil, err
	}

	return &availability, nil
}

// Statistics cache
func (r *RedisCache) SetStats(ctx context.Context, key string, stats interface{}, ttl time.Duration) error {
	cacheKey := fmt.Sprintf(DashboardStatsKey, key)
	return r.Set(ctx, cacheKey, stats, ttl)
}

func (r *RedisCache) GetStats(ctx context.Context, key string) (interface{}, error) {
	cacheKey := fmt.Sprintf(DashboardStatsKey, key)
	return r.Get(ctx, cacheKey)
}

// Metrics operations for real-time counters
func (r *RedisCache) IncrementMetric(ctx context.Context, key string, increment int64) error {
	err := r.client.IncrBy(ctx, key, increment).Err()
	if err != nil {
		r.logger.WithError(err).WithField("key", key).Error("Redis INCR failed")
		return fmt.Errorf("failed to increment metric: %w", err)
	}
	return nil
}

func (r *RedisCache) GetMetric(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get metric: %w", err)
	}
	return val, nil
}

// Cache invalidation helpers
func (r *RedisCache) InvalidateProduct(ctx context.Context, productID uuid.UUID) error {
	patterns := []string{
		fmt.Sprintf(ProductKey, productID.String()),
		fmt.Sprintf(PricingCalculationKey, "*", productID.String()),
		fmt.Sprintf(PricingTiersKey, productID.String()),
		fmt.Sprintf(AvailabilityKey, productID.String()),
		fmt.Sprintf(InventoryLevelsKey, productID.String()),
	}

	for _, pattern := range patterns {
		if err := r.DeletePattern(ctx, pattern); err != nil {
			r.logger.WithError(err).WithField("pattern", pattern).Warn("Failed to invalidate cache pattern")
		}
	}

	return nil
}

func (r *RedisCache) InvalidateProductList(ctx context.Context) error {
	patterns := []string{
		fmt.Sprintf(ProductListKey, "*"),
		ProductFeaturedKey,
		fmt.Sprintf(ProductCategoryKey, "*"),
		fmt.Sprintf(ProductSearchKey, "*"),
	}

	for _, pattern := range patterns {
		if err := r.DeletePattern(ctx, pattern); err != nil {
			r.logger.WithError(err).WithField("pattern", pattern).Warn("Failed to invalidate cache pattern")
		}
	}

	return nil
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}
