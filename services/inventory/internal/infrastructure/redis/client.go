package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"services/inventory/internal/domain"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type Client struct {
	client *redis.Client
	logger *logrus.Logger
}

func NewClient(addr, password string) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	return &Client{
		client: rdb,
		logger: logrus.New(),
	}
}

func (c *Client) Close() error {
	return c.client.Close()
}

// Health check
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// ===== PRODUCT OPERATIONS =====

// GetProduct retrieves a single product from Redis cache
func (c *Client) GetProduct(ctx context.Context, productID string) (*domain.Product, error) {
	key := fmt.Sprintf("loyverse:product:%s", productID)
	
	result, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("product not found: %s", productID)
		}
		return nil, fmt.Errorf("redis get error: %w", err)
	}

	var product domain.Product
	if err := json.Unmarshal([]byte(result), &product); err != nil {
		return nil, fmt.Errorf("json unmarshal error: %w", err)
	}

	return &product, nil
}

// GetAllProducts retrieves all products from Redis cache
func (c *Client) GetAllProducts(ctx context.Context) ([]domain.Product, error) {
	pattern := "loyverse:product:*"
	
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("redis keys error: %w", err)
	}

	if len(keys) == 0 {
		return []domain.Product{}, nil
	}

	// Use pipeline for better performance
	pipe := c.client.Pipeline()
	
	// Queue all GET commands
	cmds := make([]*redis.StringCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}
	
	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("redis pipeline error: %w", err)
	}

	// Parse results
	products := make([]domain.Product, 0, len(keys))
	for _, cmd := range cmds {
		result, err := cmd.Result()
		if err != nil {
			c.logger.WithError(err).Warn("Failed to get product from pipeline result")
			continue
		}

		var product domain.Product
		if err := json.Unmarshal([]byte(result), &product); err != nil {
			c.logger.WithError(err).Warn("Failed to unmarshal product")
			continue
		}

		products = append(products, product)
	}

	return products, nil
}

// SearchProducts searches for products by name or SKU
func (c *Client) SearchProducts(ctx context.Context, query string) ([]domain.Product, error) {
	// Get all products first (in a real implementation, you might want to use Redis Search)
	products, err := c.GetAllProducts(ctx)
	if err != nil {
		return nil, err
	}

	// Filter products by query
	var filteredProducts []domain.Product
	for _, product := range products {
		if containsIgnoreCase(product.Name, query) || 
		   containsIgnoreCase(product.SKU, query) ||
		   containsIgnoreCase(product.Barcode, query) {
			filteredProducts = append(filteredProducts, product)
		}
	}

	return filteredProducts, nil
}

// ===== STOCK OPERATIONS =====

// GetProductStock retrieves stock levels for a specific product
func (c *Client) GetProductStock(ctx context.Context, productID string) ([]domain.StockLevel, error) {
	key := fmt.Sprintf("loyverse:stock:product:%s", productID)
	
	result, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return []domain.StockLevel{}, nil
		}
		return nil, fmt.Errorf("redis get error: %w", err)
	}

	var stockLevels []domain.StockLevel
	if err := json.Unmarshal([]byte(result), &stockLevels); err != nil {
		return nil, fmt.Errorf("json unmarshal error: %w", err)
	}

	return stockLevels, nil
}

// GetStoreStock retrieves all stock levels for a specific store
func (c *Client) GetStoreStock(ctx context.Context, storeID string) ([]domain.StockLevel, error) {
	key := fmt.Sprintf("loyverse:stock:store:%s", storeID)
	
	result, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return []domain.StockLevel{}, nil
		}
		return nil, fmt.Errorf("redis get error: %w", err)
	}

	var stockLevels []domain.StockLevel
	if err := json.Unmarshal([]byte(result), &stockLevels); err != nil {
		return nil, fmt.Errorf("json unmarshal error: %w", err)
	}

	return stockLevels, nil
}

// GetLowStockItems retrieves products with low stock levels
func (c *Client) GetLowStockItems(ctx context.Context) ([]domain.StockLevel, error) {
	key := "loyverse:stock:low_stock"
	
	result, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return []domain.StockLevel{}, nil
		}
		return nil, fmt.Errorf("redis get error: %w", err)
	}

	var lowStockItems []domain.StockLevel
	if err := json.Unmarshal([]byte(result), &lowStockItems); err != nil {
		return nil, fmt.Errorf("json unmarshal error: %w", err)
	}

	return lowStockItems, nil
}

// ===== STORE & CATEGORY OPERATIONS =====

// GetAllStores retrieves all stores from Redis cache
func (c *Client) GetAllStores(ctx context.Context) ([]domain.Store, error) {
	pattern := "loyverse:store:*"
	
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("redis keys error: %w", err)
	}

	if len(keys) == 0 {
		return []domain.Store{}, nil
	}

	// Use pipeline for better performance
	pipe := c.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("redis pipeline error: %w", err)
	}

	stores := make([]domain.Store, 0, len(keys))
	for _, cmd := range cmds {
		result, err := cmd.Result()
		if err != nil {
			continue
		}

		var store domain.Store
		if err := json.Unmarshal([]byte(result), &store); err != nil {
			continue
		}

		stores = append(stores, store)
	}

	return stores, nil
}

// GetAllCategories retrieves all categories from Redis cache
func (c *Client) GetAllCategories(ctx context.Context) ([]domain.Category, error) {
	pattern := "loyverse:category:*"
	
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("redis keys error: %w", err)
	}

	if len(keys) == 0 {
		return []domain.Category{}, nil
	}

	// Use pipeline for better performance
	pipe := c.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("redis pipeline error: %w", err)
	}

	categories := make([]domain.Category, 0, len(keys))
	for _, cmd := range cmds {
		result, err := cmd.Result()
		if err != nil {
			continue
		}

		var category domain.Category
		if err := json.Unmarshal([]byte(result), &category); err != nil {
			continue
		}

		categories = append(categories, category)
	}

	return categories, nil
}

// ===== CACHE OPERATIONS =====

// SetCache sets a value in Redis with TTL
func (c *Client) SetCache(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

// GetCache gets a value from Redis cache
func (c *Client) GetCache(ctx context.Context, key string, dest interface{}) error {
	result, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("cache miss for key: %s", key)
		}
		return fmt.Errorf("redis get error: %w", err)
	}

	return json.Unmarshal([]byte(result), dest)
}

// DeleteCache deletes a key from Redis
func (c *Client) DeleteCache(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// ===== HELPER FUNCTIONS =====

func containsIgnoreCase(s, substr string) bool {
	// Simple case-insensitive contains check
	// In production, you might want to use a more sophisticated search
	return len(s) >= len(substr) && 
		   (s == substr || 
		    stringContains(stringToLower(s), stringToLower(substr)))
}

func stringToLower(s string) string {
	// Simple ASCII lowercase conversion
	result := make([]byte, len(s))
	for i, b := range []byte(s) {
		if b >= 'A' && b <= 'Z' {
			result[i] = b + 32
		} else {
			result[i] = b
		}
	}
	return string(result)
}

func stringContains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
