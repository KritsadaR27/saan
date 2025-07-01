// integrations/loyverse/internal/sync/inventory_sync.go
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

// InventorySync handles inventory synchronization
type InventorySync struct {
	client      *connector.Client
	publisher   *events.Publisher
	redis       *redis.Client
	transformer *events.Transformer
}

// NewInventorySync creates a new inventory sync instance
func NewInventorySync(client *connector.Client, publisher *events.Publisher, redis *redis.Client) *InventorySync {
	return &InventorySync{
		client:      client,
		publisher:   publisher,
		redis:       redis,
		transformer: events.NewTransformer(),
	}
}

// Sync performs inventory synchronization
func (s *InventorySync) Sync(ctx context.Context) error {
	log.Println("Starting inventory sync...")
	
	// Get last sync time
	lastSyncKey := "loyverse:sync:inventory:last"
	lastSync, _ := s.redis.Get(ctx, lastSyncKey).Result()
	
	// Fetch inventory levels
	rawData, err := s.client.GetInventoryLevels(ctx)
	if err != nil {
		return fmt.Errorf("fetching inventory levels: %w", err)
	}
	
	var domainEvents []events.DomainEvent
	processedCount := 0
	
	for _, raw := range rawData {
		var inventory models.InventoryLevel
		if err := json.Unmarshal(raw, &inventory); err != nil {
			log.Printf("Error unmarshaling inventory: %v", err)
			continue
		}
		
		// Skip if not updated since last sync
		if lastSync != "" {
			lastSyncTime, _ := time.Parse(time.RFC3339, lastSync)
			if inventory.UpdatedAt.Before(lastSyncTime) {
				continue
			}
		}
		
		// Create domain event
		event := s.createInventoryEvent(inventory)
		domainEvents = append(domainEvents, event)
		processedCount++
		
		// Cache inventory data
		cacheKey := fmt.Sprintf("loyverse:inventory:%s:%s", inventory.VariantID, inventory.StoreID)
		s.redis.Set(ctx, cacheKey, raw, 24*time.Hour)
	}
	
	// Update inventory count
	countKey := "loyverse:sync:count:inventory"
	s.redis.Set(ctx, countKey, processedCount, 0)
	
	// Publish events
	if len(domainEvents) > 0 {
		if err := s.publisher.PublishBatch(ctx, domainEvents); err != nil {
			return fmt.Errorf("publishing events: %w", err)
		}
	}
	
	// Update last sync time
	s.redis.Set(ctx, lastSyncKey, time.Now().Format(time.RFC3339), 0)
	
	log.Printf("Inventory sync completed. Processed %d inventory levels", processedCount)
	return nil
}

func (s *InventorySync) createInventoryEvent(inventory models.InventoryLevel) events.DomainEvent {
	eventData := map[string]interface{}{
		"variant_id":        inventory.VariantID,
		"store_id":          inventory.StoreID,
		"in_stock":          inventory.InStock,
		"updated_at":        inventory.UpdatedAt,
		"source":            "loyverse",
	}
	
	data, _ := json.Marshal(eventData)
	
	return events.DomainEvent{
		ID:            fmt.Sprintf("inventory-%s-%s-%d", inventory.VariantID, inventory.StoreID, time.Now().Unix()),
		Type:          events.EventInventoryUpdated,
		AggregateID:   fmt.Sprintf("%s-%s", inventory.VariantID, inventory.StoreID),
		AggregateType: "inventory",
		Timestamp:     time.Now(),
		Version:       1,
		Data:          data,
		Source:        "loyverse-integration",
	}
}
