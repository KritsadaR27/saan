// integrations/loyverse/internal/sync/supplier_sync.go
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

	"github.com/go-redis/redis/v8"
)

type SupplierSync struct {
	client      *connector.Client
	publisher   *events.Publisher
	redis       *redis.Client
	transformer *events.Transformer
}

func NewSupplierSync(client *connector.Client, publisher *events.Publisher, redis *redis.Client) *SupplierSync {
	return &SupplierSync{
		client:      client,
		publisher:   publisher,
		redis:       redis,
		transformer: events.NewTransformer(),
	}
}

func (s *SupplierSync) Sync(ctx context.Context) error {
	log.Println("Starting supplier sync...")

	// Get last sync time
	lastSyncKey := "loyverse:sync:supplier:last"
	lastSync, _ := s.redis.Get(ctx, lastSyncKey).Result()

	// Fetch suppliers
	rawData, err := s.client.GetSuppliers(ctx)
	if err != nil {
		return fmt.Errorf("fetching suppliers: %w", err)
	}

	var domainEvents []events.DomainEvent
	processedCount := 0
	activeSuppliers := make([]string, 0)

	for _, raw := range rawData {
		var supplier models.Supplier
		if err := json.Unmarshal(raw, &supplier); err != nil {
			log.Printf("Error unmarshaling supplier: %v", err)
			continue
		}

		// Skip if not updated since last sync
		if lastSync != "" {
			lastSyncTime, _ := time.Parse(time.RFC3339, lastSync)
			if supplier.UpdatedAt.Before(lastSyncTime) {
				continue
			}
		}

		// Track active suppliers
		if supplier.DeletedAt == nil {
			activeSuppliers = append(activeSuppliers, supplier.ID)
		}

		// Create domain event
		event := s.createSupplierEvent(supplier)
		domainEvents = append(domainEvents, event)
		processedCount++

		// Cache supplier data
		cacheKey := fmt.Sprintf("loyverse:supplier:%s", supplier.ID)
		s.redis.Set(ctx, cacheKey, raw, 24*time.Hour)
	}

	// Cache active suppliers list
	activeData, _ := json.Marshal(activeSuppliers)
	s.redis.Set(ctx, "loyverse:suppliers:active", activeData, 24*time.Hour)

	// Update supplier count
	countKey := "loyverse:sync:count:suppliers"
	s.redis.Set(ctx, countKey, processedCount, 0)

	// Publish events
	if len(domainEvents) > 0 {
		if err := s.publisher.PublishBatch(ctx, domainEvents); err != nil {
			return fmt.Errorf("publishing events: %w", err)
		}
	}

	// Update last sync time
	s.redis.Set(ctx, lastSyncKey, time.Now().Format(time.RFC3339), 0)

	log.Printf("Supplier sync completed. Processed %d suppliers (%d active)", processedCount, len(activeSuppliers))
	return nil
}

func (s *SupplierSync) createSupplierEvent(supplier models.Supplier) events.DomainEvent {
	eventData := map[string]interface{}{
		"supplier_id":  supplier.ID,
		"name":         supplier.Name,
		"contact_name": supplier.ContactName,
		"phone":        supplier.Phone,
		"email":        supplier.Email,
		"address":      supplier.Address,
		"note":         supplier.Note,
		"is_active":    supplier.DeletedAt == nil,
		"source":       "loyverse",
	}

	data, _ := json.Marshal(eventData)

	return events.DomainEvent{
		ID:            fmt.Sprintf("supplier-%s-%d", supplier.ID, time.Now().Unix()),
		Type:          events.EventSupplierUpdated,
		AggregateID:   supplier.ID,
		AggregateType: "supplier",
		Timestamp:     time.Now(),
		Version:       1,
		Data:          data,
		Source:        "loyverse-integration",
	}
}
