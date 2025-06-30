package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

// Processor processes webhook events with deduplication
type Processor struct {
	redis *redis.Client
}

// NewProcessor creates a new webhook processor
func NewProcessor(redis *redis.Client) *Processor {
	return &Processor{
		redis: redis,
	}
}

// ProcessEvent processes a webhook event with deduplication
func (p *Processor) ProcessEvent(ctx context.Context, eventID string, eventType string, data json.RawMessage) error {
	// Check if event already processed
	key := fmt.Sprintf("loyverse:webhook:processed:%s", eventID)
	exists, err := p.redis.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("checking event existence: %w", err)
	}

	if exists > 0 {
		log.Printf("Event %s already processed, skipping", eventID)
		return nil
	}

	// Mark event as processed
	if err := p.redis.Set(ctx, key, time.Now().Format(time.RFC3339), 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("marking event as processed: %w", err)
	}

	// Store event data for debugging
	eventKey := fmt.Sprintf("loyverse:webhook:event:%s", eventID)
	eventData := map[string]interface{}{
		"type":        eventType,
		"data":        string(data),
		"processed_at": time.Now().Format(time.RFC3339),
	}
	
	if err := p.redis.HMSet(ctx, eventKey, eventData).Err(); err != nil {
		log.Printf("Error storing event data: %v", err)
	}
	p.redis.Expire(ctx, eventKey, 7*24*time.Hour)

	return nil
}

// StoreLatestReceipt stores the latest receipt data in Redis
func (p *Processor) StoreLatestReceipt(ctx context.Context, receiptData json.RawMessage) error {
	key := "loyverse:latest_receipt"
	
	receiptInfo := map[string]interface{}{
		"data":       string(receiptData),
		"updated_at": time.Now().Format(time.RFC3339),
	}
	
	if err := p.redis.HMSet(ctx, key, receiptInfo).Err(); err != nil {
		return fmt.Errorf("storing latest receipt: %w", err)
	}
	
	// Set expiration to 30 days
	p.redis.Expire(ctx, key, 30*24*time.Hour)
	
	return nil
}

// GetLatestReceipt retrieves the latest receipt data from Redis
func (p *Processor) GetLatestReceipt(ctx context.Context) (json.RawMessage, time.Time, error) {
	key := "loyverse:latest_receipt"
	
	result, err := p.redis.HMGet(ctx, key, "data", "updated_at").Result()
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("getting latest receipt: %w", err)
	}
	
	if result[0] == nil {
		return nil, time.Time{}, fmt.Errorf("no receipt found")
	}
	
	dataStr, ok := result[0].(string)
	if !ok {
		return nil, time.Time{}, fmt.Errorf("invalid data format")
	}
	
	var updatedAt time.Time
	if result[1] != nil {
		if updatedAtStr, ok := result[1].(string); ok {
			updatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		}
	}
	
	return json.RawMessage(dataStr), updatedAt, nil
}

// StoreCategories stores category data in Redis
func (p *Processor) StoreCategories(ctx context.Context, categories []json.RawMessage) error {
	for i, categoryData := range categories {
		key := fmt.Sprintf("loyverse:category:%d", i)
		
		categoryInfo := map[string]interface{}{
			"data":       string(categoryData),
			"updated_at": time.Now().Format(time.RFC3339),
		}
		
		if err := p.redis.HMSet(ctx, key, categoryInfo).Err(); err != nil {
			return fmt.Errorf("storing category %d: %w", i, err)
		}
		
		// Set expiration to 30 days
		p.redis.Expire(ctx, key, 30*24*time.Hour)
	}
	
	// Store the count of categories
	countKey := "loyverse:categories:count"
	if err := p.redis.Set(ctx, countKey, len(categories), 30*24*time.Hour).Err(); err != nil {
		return fmt.Errorf("storing categories count: %w", err)
	}
	
	return nil
}

// GetAllCategories retrieves all categories from Redis
func (p *Processor) GetAllCategories(ctx context.Context) ([]json.RawMessage, error) {
	// Get count of categories
	countKey := "loyverse:categories:count"
	count, err := p.redis.Get(ctx, countKey).Int()
	if err != nil {
		if err == redis.Nil {
			return []json.RawMessage{}, nil
		}
		return nil, fmt.Errorf("getting categories count: %w", err)
	}
	
	var categories []json.RawMessage
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("loyverse:category:%d", i)
		
		result, err := p.redis.HMGet(ctx, key, "data").Result()
		if err != nil {
			log.Printf("Error getting category %d: %v", i, err)
			continue
		}
		
		if result[0] != nil {
			if dataStr, ok := result[0].(string); ok {
				categories = append(categories, json.RawMessage(dataStr))
			}
		}
	}
	
	return categories, nil
}
