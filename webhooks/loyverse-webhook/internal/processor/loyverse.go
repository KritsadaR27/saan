// webhooks/loyverse-webhook/internal/processor/loyverse.go
package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"

	"webhooks/loyverse-webhook/internal/validator"
)

// LoyverseProcessor handles processing of Loyverse webhook events
type LoyverseProcessor struct {
	kafkaWriter *kafka.Writer
	redisClient *redis.Client
}

// NewLoyverseProcessor creates a new Loyverse webhook processor
func NewLoyverseProcessor(kafkaWriter *kafka.Writer, redisClient *redis.Client) *LoyverseProcessor {
	return &LoyverseProcessor{
		kafkaWriter: kafkaWriter,
		redisClient: redisClient,
	}
}

// DomainEvent represents a domain event to be published
type DomainEvent struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	Timestamp     time.Time              `json:"timestamp"`
	Version       int                    `json:"version"`
	Data          map[string]interface{} `json:"data"`
	Source        string                 `json:"source"`
}

// ProcessWebhook processes a validated webhook payload
func (p *LoyverseProcessor) ProcessWebhook(ctx context.Context, payload *validator.WebhookPayload) error {
	log.Printf("Processing Loyverse webhook: type=%s, created_at=%v", payload.Type, payload.CreatedAt)

	// Cache the raw webhook for debugging
	if err := p.cacheWebhook(ctx, payload); err != nil {
		log.Printf("Warning: Failed to cache webhook: %v", err)
	}

	// Transform webhook to domain event
	domainEvent, err := p.transformToDomainEvent(payload)
	if err != nil {
		return fmt.Errorf("failed to transform webhook to domain event: %w", err)
	}

	// Publish to Kafka
	if err := p.publishEvent(ctx, domainEvent); err != nil {
		return fmt.Errorf("failed to publish event to Kafka: %w", err)
	}

	log.Printf("Successfully processed webhook %s", domainEvent.ID)
	return nil
}

// cacheWebhook stores the webhook payload in Redis for debugging
func (p *LoyverseProcessor) cacheWebhook(ctx context.Context, payload *validator.WebhookPayload) error {
	cacheKey := fmt.Sprintf("loyverse:webhook:%s:%d", payload.Type, payload.CreatedAt.Unix())
	
	webhookData := map[string]interface{}{
		"type":       payload.Type,
		"created_at": payload.CreatedAt,
		"data":       payload.Data,
		"cached_at":  time.Now(),
	}

	data, err := json.Marshal(webhookData)
	if err != nil {
		return err
	}

	// Cache for 7 days
	return p.redisClient.Set(ctx, cacheKey, data, 7*24*time.Hour).Err()
}

// transformToDomainEvent converts webhook payload to domain event
func (p *LoyverseProcessor) transformToDomainEvent(payload *validator.WebhookPayload) (*DomainEvent, error) {
	// Parse the data field to extract relevant information
	var data map[string]interface{}
	if err := json.Unmarshal(payload.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse webhook data: %w", err)
	}

	// Extract aggregate ID based on webhook type
	aggregateID := p.extractAggregateID(payload.Type, data)
	
	// Create domain event
	domainEvent := &DomainEvent{
		ID:            fmt.Sprintf("loyverse-%s-%d", payload.Type, payload.CreatedAt.Unix()),
		Type:          p.mapWebhookTypeToEventType(payload.Type),
		AggregateID:   aggregateID,
		AggregateType: p.extractAggregateType(payload.Type),
		Timestamp:     payload.CreatedAt,
		Version:       1,
		Data: map[string]interface{}{
			"webhook_type": payload.Type,
			"source":       "loyverse",
			"raw_data":     data,
		},
		Source: "loyverse-webhook",
	}

	return domainEvent, nil
}

// extractAggregateID extracts the aggregate ID from webhook data
func (p *LoyverseProcessor) extractAggregateID(webhookType string, data map[string]interface{}) string {
	// Try to extract ID from common fields
	if id, exists := data["id"]; exists {
		if idStr, ok := id.(string); ok {
			return idStr
		}
	}

	// For receipts, try receipt_number
	if webhookType == "receipt_created" {
		if receiptNumber, exists := data["receipt_number"]; exists {
			if receiptStr, ok := receiptNumber.(string); ok {
				return receiptStr
			}
		}
	}

	// Fallback to webhook type with timestamp
	return fmt.Sprintf("%s-%d", webhookType, time.Now().Unix())
}

// extractAggregateType maps webhook type to aggregate type
func (p *LoyverseProcessor) extractAggregateType(webhookType string) string {
	switch webhookType {
	case "receipt_created", "receipt_updated":
		return "receipt"
	case "customer_created", "customer_updated":
		return "customer"
	case "product_created", "product_updated":
		return "product"
	case "inventory_updated":
		return "inventory"
	default:
		return "loyverse_event"
	}
}

// mapWebhookTypeToEventType maps Loyverse webhook types to domain event types
func (p *LoyverseProcessor) mapWebhookTypeToEventType(webhookType string) string {
	switch webhookType {
	case "receipt_created":
		return "receipt.created"
	case "receipt_updated":
		return "receipt.updated"
	case "customer_created":
		return "customer.created"
	case "customer_updated":
		return "customer.updated"
	case "product_created":
		return "product.created"
	case "product_updated":
		return "product.updated"
	case "inventory_updated":
		return "inventory.updated"
	default:
		return fmt.Sprintf("loyverse.%s", webhookType)
	}
}

// publishEvent publishes the domain event to Kafka
func (p *LoyverseProcessor) publishEvent(ctx context.Context, event *DomainEvent) error {
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(event.AggregateID),
		Value: eventData,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte(event.Type)},
			{Key: "aggregate-type", Value: []byte(event.AggregateType)},
			{Key: "source", Value: []byte(event.Source)},
		},
		Time: event.Timestamp,
	}

	return p.kafkaWriter.WriteMessages(ctx, message)
}
