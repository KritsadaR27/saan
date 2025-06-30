// integrations/loyverse/internal/sync/store_sync.go
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

type StoreSync struct {
	client      *connector.Client
	publisher   *events.Publisher
	redis       *redis.Client
	transformer *events.Transformer
}

func NewStoreSync(client *connector.Client, publisher *events.Publisher, redis *redis.Client) *StoreSync {
	return &StoreSync{
		client:      client,
		publisher:   publisher,
		redis:       redis,
		transformer: events.NewTransformer(),
	}
}

func (s *StoreSync) Sync(ctx context.Context) error {
	log.Println("Starting store sync...")
	
	// Get last sync time
	lastSyncKey := "loyverse:sync:store:last"
	lastSync, _ := s.redis.Get(ctx, lastSyncKey).Result()
	
	// Fetch stores
	rawData, err := s.client.GetStores(ctx)
	if err != nil {
		return fmt.Errorf("fetching stores: %w", err)
	}
	
	var domainEvents []events.DomainEvent
	processedCount := 0
	
	for _, raw := range rawData {
		var store models.Store
		if err := json.Unmarshal(raw, &store); err != nil {
			log.Printf("Error unmarshaling store: %v", err)
			continue
		}
		
		// Skip if not updated since last sync
		if lastSync != "" {
			lastSyncTime, _ := time.Parse(time.RFC3339, lastSync)
			if store.UpdatedAt.Before(lastSyncTime) {
				continue
			}
		}
		
		// Create domain event
		event := s.createStoreEvent(store)
		domainEvents = append(domainEvents, event)
		processedCount++
		
		// Cache store data
		cacheKey := fmt.Sprintf("loyverse:store:%s", store.ID)
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
	
	log.Printf("Store sync completed. Processed %d stores", processedCount)
	return nil
}

func (s *StoreSync) createStoreEvent(store models.Store) events.DomainEvent {
	eventData := map[string]interface{}{
		"store_id":       store.ID,
		"name":           store.Name,
		"address":        store.Address,
		"phone":          store.Phone,
		"email":          store.Email,
		"description":    store.Description,
		"receipt_footer": store.ReceiptFooter,
		"tax_number":     store.TaxNumber,
		"source":         "loyverse",
	}
	
	data, _ := json.Marshal(eventData)
	
	return events.DomainEvent{
		ID:            fmt.Sprintf("store-%s-%d", store.ID, time.Now().Unix()),
		Type:          events.EventStoreUpdated,
		AggregateID:   store.ID,
		AggregateType: "store",
		Timestamp:     time.Now(),
		Version:       1,
		Data:          data,
		Source:        "loyverse-integration",
	}
}
