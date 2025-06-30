// integrations/loyverse/internal/sync/category_sync.go
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

type CategorySync struct {
	client      *connector.Client
	publisher   *events.Publisher
	redis       *redis.Client
	transformer *events.Transformer
}

func NewCategorySync(client *connector.Client, publisher *events.Publisher, redis *redis.Client) *CategorySync {
	return &CategorySync{
		client:      client,
		publisher:   publisher,
		redis:       redis,
		transformer: events.NewTransformer(),
	}
}

func (s *CategorySync) Sync(ctx context.Context) error {
	log.Println("Starting category sync...")
	
	// Get last sync time
	lastSyncKey := "loyverse:sync:category:last"
	lastSync, _ := s.redis.Get(ctx, lastSyncKey).Result()
	
	// Fetch categories
	rawData, err := s.client.GetCategories(ctx)
	if err != nil {
		return fmt.Errorf("fetching categories: %w", err)
	}
	
	var domainEvents []events.DomainEvent
	processedCount := 0
	
	for _, raw := range rawData {
		var category models.Category
		if err := json.Unmarshal(raw, &category); err != nil {
			log.Printf("Error unmarshaling category: %v", err)
			continue
		}
		
		// Skip if not updated since last sync
		if lastSync != "" {
			lastSyncTime, _ := time.Parse(time.RFC3339, lastSync)
			if category.UpdatedAt.Before(lastSyncTime) {
				continue
			}
		}
		
		// Create domain event
		event := s.createCategoryEvent(category)
		domainEvents = append(domainEvents, event)
		processedCount++
		
		// Cache category data
		cacheKey := fmt.Sprintf("loyverse:category:%s", category.ID)
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
	
	log.Printf("Category sync completed. Processed %d categories", processedCount)
	return nil
}

func (s *CategorySync) createCategoryEvent(category models.Category) events.DomainEvent {
	eventData := map[string]interface{}{
		"category_id": category.ID,
		"name":        category.Name,
		"color":       category.Color,
		"parent_id":   category.ParentID,
		"source":      "loyverse",
	}
	
	data, _ := json.Marshal(eventData)
	
	return events.DomainEvent{
		ID:            fmt.Sprintf("cat-%s-%d", category.ID, time.Now().Unix()),
		Type:          events.EventCategoryUpdated,
		AggregateID:   category.ID,
		AggregateType: "category",
		Timestamp:     time.Now(),
		Version:       1,
		Data:          data,
		Source:        "loyverse-integration",
	}
}
