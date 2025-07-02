// integrations/loyverse/internal/sync/payment_type_sync.go
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

type PaymentTypeSync struct {
	client      *connector.Client
	publisher   *events.Publisher
	redis       *redis.Client
	transformer *events.Transformer
}

func NewPaymentTypeSync(client *connector.Client, publisher *events.Publisher, redis *redis.Client) *PaymentTypeSync {
	return &PaymentTypeSync{
		client:      client,
		publisher:   publisher,
		redis:       redis,
		transformer: events.NewTransformer(),
	}
}

func (s *PaymentTypeSync) Sync(ctx context.Context) error {
	log.Println("Starting payment type sync...")

	// Fetch payment types
	rawData, err := s.client.GetPaymentTypes(ctx)
	if err != nil {
		return fmt.Errorf("fetching payment types: %w", err)
	}

	var domainEvents []events.DomainEvent
	processedCount := 0

	// Group payment types by type
	paymentTypesByType := make(map[string][]models.PaymentType)

	for _, raw := range rawData {
		var paymentType models.PaymentType
		if err := json.Unmarshal(raw, &paymentType); err != nil {
			log.Printf("Error unmarshaling payment type: %v", err)
			continue
		}

		// Create domain event
		event := s.createPaymentTypeEvent(paymentType)
		domainEvents = append(domainEvents, event)
		processedCount++

		// Cache payment type data with enhanced error handling
		cacheKey := fmt.Sprintf("loyverse:payment_type:%s", paymentType.ID)
		if err := s.redis.SafeSet(ctx, cacheKey, raw, 24*time.Hour); err != nil {
			log.Printf("Error caching payment type %s: %v", paymentType.ID, err)
		}

		// Group by type
		paymentTypesByType[paymentType.Type] = append(paymentTypesByType[paymentType.Type], paymentType)
	}

	// Cache payment types by type for quick lookup with enhanced error handling
	for typeKey, types := range paymentTypesByType {
		groupKey := fmt.Sprintf("loyverse:payment_types:by_type:%s", typeKey)
		groupData, _ := json.Marshal(types)
		if err := s.redis.SafeSet(ctx, groupKey, groupData, 24*time.Hour); err != nil {
			log.Printf("Error caching payment types by type %s: %v", typeKey, err)
		}
	}

	// Cache all payment types for quick access with enhanced error handling
	allKey := "loyverse:payment_types:all"
	allData, _ := json.Marshal(rawData)
	if err := s.redis.SafeSet(ctx, allKey, allData, 24*time.Hour); err != nil {
		log.Printf("Error caching all payment types: %v", err)
	}

	// Update payment type count with enhanced error handling
	countKey := "loyverse:sync:count:payment_types"
	if err := s.redis.SafeSet(ctx, countKey, processedCount, 0); err != nil {
		log.Printf("Error updating payment type count: %v", err)
	}

	// Publish events
	if len(domainEvents) > 0 {
		if err := s.publisher.PublishBatch(ctx, domainEvents); err != nil {
			return fmt.Errorf("publishing events: %w", err)
		}
	}

	log.Printf("Payment type sync completed. Processed %d payment types", processedCount)
	return nil
}

func (s *PaymentTypeSync) createPaymentTypeEvent(paymentType models.PaymentType) events.DomainEvent {
	eventData := map[string]interface{}{
		"payment_type_id":  paymentType.ID,
		"name":             paymentType.Name,
		"type":             paymentType.Type,
		"show_in_pos":      paymentType.ShowInPOS,
		"show_in_receipts": paymentType.ShowInReceipts,
		"order_index":      paymentType.OrderIndex,
		"is_active":        paymentType.DeletedAt == nil,
		"source":           "loyverse",
	}

	data, _ := json.Marshal(eventData)

	return events.DomainEvent{
		ID:            fmt.Sprintf("payment_type-%s-%d", paymentType.ID, time.Now().Unix()),
		Type:          events.EventPaymentTypeUpdated,
		AggregateID:   paymentType.ID,
		AggregateType: "payment_type",
		Timestamp:     time.Now(),
		Version:       1,
		Data:          data,
		Source:        "loyverse-integration",
	}
}
