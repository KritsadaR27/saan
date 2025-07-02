// integrations/loyverse/internal/sync/product_sync.go
package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"integrations/loyverse/internal/connector"
	"integrations/loyverse/internal/events"
	"integrations/loyverse/internal/models"
	"integrations/loyverse/internal/redis"
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
	
	// Test Redis connection first with enhanced error handling
	if !s.redis.IsHealthy() {
		log.Println("WARNING: Redis is currently unhealthy, continuing with degraded functionality")
	} else {
		log.Println("Redis connection verified and healthy")
	}
	
	// Get last sync time (for future use)
	lastSyncKey := "loyverse:sync:product:last"
	// lastSync, _ := s.redis.SafeGet(ctx, lastSyncKey) // Disabled for testing
	
	// Fetch products
	rawData, err := s.client.GetProducts(ctx)
	if err != nil {
		return fmt.Errorf("fetching products: %w", err)
	}
	
	var domainEvents []events.DomainEvent
	processedCount := 0
	
	for i, raw := range rawData {
		// Debug: Log the first few raw items to understand structure
		if i < 3 {
			log.Printf("DEBUG: Raw item %d: %s", i, string(raw))
		}
		
		var item models.Item
		if err := json.Unmarshal(raw, &item); err != nil {
			log.Printf("Error unmarshaling item %d: %v", i, err)
			if i < 3 {
				log.Printf("DEBUG: Failed raw item %d: %s", i, string(raw))
			}
			continue
		}
		
	// For testing: Sync all products (skip time filter)
	// TODO: Re-enable time filter for production
	/*
	// Skip if not updated since last sync
	if lastSync != "" {
		lastSyncTime, _ := time.Parse(time.RFC3339, lastSync)
		if item.UpdatedAt.Before(lastSyncTime) {
			continue
		}
	}
	*/
		
		// Create domain event
		event := s.createProductEvent(item)
		domainEvents = append(domainEvents, event)
		processedCount++
		
		// Cache product data with enhanced error handling
		cacheKey := fmt.Sprintf("loyverse:product:%s", item.ID)
		if err := s.redis.SafeSet(ctx, cacheKey, raw, 24*time.Hour); err != nil {
			log.Printf("Error caching product %s: %v", item.ID, err)
		} else {
			log.Printf("Successfully cached product %s with key %s", item.ID, cacheKey)
		}
		
		// Cache variants with enhanced error handling
		for _, variant := range item.Variants {
			variantKey := fmt.Sprintf("loyverse:variant:%s", variant.ID)
			variantData, _ := json.Marshal(variant)
			if err := s.redis.SafeSet(ctx, variantKey, variantData, 24*time.Hour); err != nil {
				log.Printf("Error caching variant %s: %v", variant.ID, err)
			} else {
				log.Printf("Successfully cached variant %s with key %s", variant.ID, variantKey)
			}
		}
	}
	
	// Update product count with enhanced error handling
	countKey := "loyverse:sync:count:products"
	if err := s.redis.SafeSet(ctx, countKey, processedCount, 0); err != nil {
		log.Printf("Error updating product count: %v", err)
	} else {
		log.Printf("Successfully updated product count: %d", processedCount)
	}
	
	// Publish events
	if len(domainEvents) > 0 {
		if err := s.publisher.PublishBatch(ctx, domainEvents); err != nil {
			return fmt.Errorf("publishing events: %w", err)
		}
	}
	
	// Update last sync time with enhanced error handling
	if err := s.redis.SafeSet(ctx, lastSyncKey, time.Now().Format(time.RFC3339), 0); err != nil {
		log.Printf("Error updating last sync time: %v", err)
	} else {
		log.Printf("Successfully updated last sync time")
	}
	
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
