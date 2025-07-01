// integrations/loyverse/internal/sync/receipt_sync.go
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

// ReceiptSync handles receipt synchronization
type ReceiptSync struct {
	client      *connector.Client
	publisher   *events.Publisher
	redis       *redis.Client
	transformer *events.Transformer
}

// NewReceiptSync creates a new receipt sync instance
func NewReceiptSync(client *connector.Client, publisher *events.Publisher, redis *redis.Client) *ReceiptSync {
	return &ReceiptSync{
		client:      client,
		publisher:   publisher,
		redis:       redis,
		transformer: events.NewTransformer(),
	}
}

// Sync performs receipt synchronization
func (s *ReceiptSync) Sync(ctx context.Context) error {
	log.Println("Starting receipt sync...")
	
	// Get last sync time
	lastSyncKey := "loyverse:sync:receipt:last"
	lastSync, _ := s.redis.Get(ctx, lastSyncKey).Result()
	
	// Fetch receipts
	rawData, err := s.client.GetReceipts(ctx)
	if err != nil {
		return fmt.Errorf("fetching receipts: %w", err)
	}
	
	var domainEvents []events.DomainEvent
	processedCount := 0
	
	for _, raw := range rawData {
		var receipt models.Receipt
		if err := json.Unmarshal(raw, &receipt); err != nil {
			log.Printf("Error unmarshaling receipt: %v", err)
			continue
		}
		
		// Skip if not updated since last sync
		if lastSync != "" {
			lastSyncTime, _ := time.Parse(time.RFC3339, lastSync)
			if receipt.UpdatedAt.Before(lastSyncTime) {
				continue
			}
		}
		
		// Create domain event
		event := s.createReceiptEvent(receipt)
		domainEvents = append(domainEvents, event)
		processedCount++
		
		// Cache receipt data
		cacheKey := fmt.Sprintf("loyverse:receipt:%s", receipt.ID)
		s.redis.Set(ctx, cacheKey, raw, 24*time.Hour)
	}
	
	// Update receipt count
	countKey := "loyverse:sync:count:receipts"
	s.redis.Set(ctx, countKey, processedCount, 0)
	
	// Publish events
	if len(domainEvents) > 0 {
		if err := s.publisher.PublishBatch(ctx, domainEvents); err != nil {
			return fmt.Errorf("publishing events: %w", err)
		}
	}
	
	// Update last sync time
	s.redis.Set(ctx, lastSyncKey, time.Now().Format(time.RFC3339), 0)
	
	log.Printf("Receipt sync completed. Processed %d receipts", processedCount)
	return nil
}

func (s *ReceiptSync) createReceiptEvent(receipt models.Receipt) events.DomainEvent {
	eventData := map[string]interface{}{
		"receipt_id":       receipt.ID,
		"receipt_number":   receipt.Number,
		"note":             receipt.Note,
		"receipt_type":     receipt.ReceiptType,
		"refund_for":       receipt.RefundFor,
		"order":            receipt.Order,
		"receipt_date":     receipt.ReceiptDate,
		"source":           receipt.Source,
		"total_money":      receipt.TotalMoney,
		"total_tax":        receipt.TotalTax,
		"points_earned":    receipt.PointsEarned,
		"points_deducted":  receipt.PointsDeducted,
		"points_balance":   receipt.PointsBalance,
		"cancelled_at":     receipt.CancelledAt,
		"loyverse_source":  "loyverse",
	}
	
	data, _ := json.Marshal(eventData)
	
	return events.DomainEvent{
		ID:            fmt.Sprintf("receipt-%s-%d", receipt.ID, time.Now().Unix()),
		Type:          events.EventReceiptCreated,
		AggregateID:   receipt.ID,
		AggregateType: "receipt",
		Timestamp:     time.Now(),
		Version:       1,
		Data:          data,
		Source:        "loyverse-integration",
	}
}
