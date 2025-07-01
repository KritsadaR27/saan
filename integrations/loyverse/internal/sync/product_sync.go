// integrations/loyverse/internal/sync/product_sync.go
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

type ProductSync struct {
	client      *connector.Client
	publisher   *events.Publisher
	redis       *redis.Client
	transformer *events.Transformer
}

func NewProductSync(client *connector.Client, publisher *events.Publisher, redis *redis.Client) *ProductSync {
	return &ProductSync{
		client:      client,
		publisher:   publisher,
		redis:       redis,
		transformer: events.NewTransformer(),
	}
}

func (s *ProductSync) Sync(ctx context.Context) error {
	log.Println("Starting product sync...")
	
	// Get last sync time
	lastSyncKey := "loyverse:sync:product:last"
	lastSync, _ := s.redis.Get(ctx, lastSyncKey).Result()
	
	// Fetch products
	rawData, err := s.client.GetProducts(ctx)
	if err != nil {
		return fmt.Errorf("fetching products: %w", err)
	}
	
	var domainEvents []events.DomainEvent
	processedCount := 0
	
	for _, raw := range rawData {
		var item models.Item
		if err := json.Unmarshal(raw, &item); err != nil {
			log.Printf("Error unmarshaling item: %v", err)
			continue
		}
		
		// Skip if not updated since last sync
		if lastSync != "" {
			lastSyncTime, _ := time.Parse(time.RFC3339, lastSync)
			if item.UpdatedAt.Before(lastSyncTime) {
				continue
			}
		}
		
		// Create domain event
		event := s.createProductEvent(item)
		domainEvents = append(domainEvents, event)
		processedCount++
		
		// Cache product data
		cacheKey := fmt.Sprintf("loyverse:product:%s", item.ID)
		s.redis.Set(ctx, cacheKey, raw, 24*time.Hour)
		
		// Cache variants
		for _, variant := range item.Variants {
			variantKey := fmt.Sprintf("loyverse:variant:%s", variant.ID)
			variantData, _ := json.Marshal(variant)
			s.redis.Set(ctx, variantKey, variantData, 24*time.Hour)
		}
	}
	
	// Update product count
	countKey := "loyverse:sync:count:products"
	s.redis.Set(ctx, countKey, processedCount, 0)
	
	// Publish events
	if len(domainEvents) > 0 {
		if err := s.publisher.PublishBatch(ctx, domainEvents); err != nil {
			return fmt.Errorf("publishing events: %w", err)
		}
	}
	
	// Update last sync time
	s.redis.Set(ctx, lastSyncKey, time.Now().Format(time.RFC3339), 0)
	
	log.Printf("Product sync completed. Processed %d products", processedCount)
	return nil
}

func (s *ProductSync) createProductEvent(item models.Item) events.DomainEvent {
	eventData := map[string]interface{}{
		"product_id":          item.ID,
		"name":                item.Name,
		"description":         item.Description,
		"category_id":         item.CategoryID,
		"primary_supplier_id": item.PrimarySupplierID,
		"sku":                 item.SKU,
		"barcode":             item.Barcode,
		"track_stock":         item.TrackStock,
		"sold_by_weight":      item.SoldByWeight,
		"is_composite":        item.IsComposite,
		"use_production":      item.UseProduction,
		"variants":            item.Variants,
		"source":              "loyverse",
	}
	
	data, _ := json.Marshal(eventData)
	
	return events.DomainEvent{
		ID:            fmt.Sprintf("product-%s-%d", item.ID, time.Now().Unix()),
		Type:          events.EventProductUpdated,
		AggregateID:   item.ID,
		AggregateType: "product",
		Timestamp:     time.Now(),
		Version:       1,
		Data:          data,
		Source:        "loyverse-integration",
	}
}
