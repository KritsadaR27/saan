package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"integrations/loyverse/internal/connector"
	"integrations/loyverse/internal/events"
	"integrations/loyverse/internal/models"
)

type DiscountSync struct {
	client      *connector.Client
	publisher   *events.Publisher
	redis       *redis.Client
	transformer *events.Transformer
}

func NewDiscountSync(client *connector.Client, publisher *events.Publisher, redis *redis.Client) *DiscountSync {
	return &DiscountSync{
		client:      client,
		publisher:   publisher,
		redis:       redis,
		transformer: events.NewTransformer(),
	}
}

func (s *DiscountSync) Sync(ctx context.Context) error {
	log.Println("Starting discount sync...")
	
	// Get last sync time
	lastSyncKey := "loyverse:sync:discount:last"
	lastSync, _ := s.redis.Get(ctx, lastSyncKey).Result()
	
	// Fetch discounts
	rawData, err := s.client.GetDiscounts(ctx)
	if err != nil {
		return fmt.Errorf("fetching discounts: %w", err)
	}
	
	var domainEvents []events.DomainEvent
	processedCount := 0
	
	for _, raw := range rawData {
		var discount models.Discount
		if err := json.Unmarshal(raw, &discount); err != nil {
			log.Printf("Error unmarshaling discount: %v", err)
			continue
		}
		
		// Skip if not updated since last sync
		if lastSync != "" {
			lastSyncTime, _ := time.Parse(time.RFC3339, lastSync)
			if discount.UpdatedAt.Before(lastSyncTime) {
				continue
			}
		}
		
		// Create domain event
		event := s.createDiscountEvent(discount)
		domainEvents = append(domainEvents, event)
		processedCount++
		
		// Cache discount data
		cacheKey := fmt.Sprintf("loyverse:discount:%s", discount.ID)
		s.redis.Set(ctx, cacheKey, raw, 24*time.Hour)
	}
	
	// Publish events
	if len(domainEvents) > 0 {
		if err := s.publisher.PublishBatch(ctx, domainEvents); err != nil {
			return fmt.Errorf("publishing events: %w", err)
		}
	}
	
	// Update last sync time
	s.redis.Set(ctx, lastSyncKey, time.Now().Format(time.RFC3339), 0)
	
	log.Printf("Discount sync completed. Processed %d discounts", processedCount)
	return nil
}

func (s *DiscountSync) createDiscountEvent(discount models.Discount) events.DomainEvent {
	eventData := map[string]interface{}{
		"discount_id":        discount.ID,
		"name":               discount.Name,
		"type":               discount.Type,
		"value":              discount.Value,
		"is_restricted":      discount.IsRestricted,
		"minimum_amount":     discount.MinimumAmount,
		"valid_since":        discount.ValidSince,
		"valid_until":        discount.ValidUntil,
		"store_ids":          discount.StoreIDs,
		"source":             "loyverse",
	}
	
	data, _ := json.Marshal(eventData)
	
	return events.DomainEvent{
		ID:            fmt.Sprintf("disc-%s-%d", discount.ID, time.Now().Unix()),
		Type:          events.EventDiscountUpdated,
		AggregateID:   discount.ID,
		AggregateType: "discount",
		Timestamp:     time.Now(),
		Version:       1,
		Data:          data,
		Source:        "loyverse-integration",
	}
}
