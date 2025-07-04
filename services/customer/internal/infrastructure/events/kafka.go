package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
	"github.com/google/uuid"
	"github.com/saan-system/services/customer/internal/domain/entity"
	"github.com/saan-system/services/customer/internal/domain/repository"
)

// kafkaEventPublisher implements repository.EventPublisher
type kafkaEventPublisher struct {
	writer *kafka.Writer
}

// NewKafkaEventPublisher creates a new Kafka event publisher
func NewKafkaEventPublisher(brokers []string, topic string) repository.EventPublisher {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &kafkaEventPublisher{
		writer: writer,
	}
}

// PublishCustomerCreated publishes a customer created event
func (p *kafkaEventPublisher) PublishCustomerCreated(ctx context.Context, customer *entity.Customer) error {
	event := map[string]interface{}{
		"event_type":  "customer.created",
		"customer_id": customer.ID,
		"customer":    customer,
		"timestamp":   customer.CreatedAt,
	}

	return p.publishEvent(ctx, "customer.created", customer.ID.String(), event)
}

// PublishCustomerUpdated publishes a customer updated event
func (p *kafkaEventPublisher) PublishCustomerUpdated(ctx context.Context, customer *entity.Customer) error {
	event := map[string]interface{}{
		"event_type":  "customer.updated",
		"customer_id": customer.ID,
		"customer":    customer,
		"timestamp":   customer.UpdatedAt,
	}

	return p.publishEvent(ctx, "customer.updated", customer.ID.String(), event)
}

// PublishCustomerDeleted publishes a customer deleted event
func (p *kafkaEventPublisher) PublishCustomerDeleted(ctx context.Context, customerID uuid.UUID) error {
	event := map[string]interface{}{
		"event_type":  "customer.deleted",
		"customer_id": customerID,
		"timestamp":   nil, // Will be set by publishEvent
	}

	return p.publishEvent(ctx, "customer.deleted", customerID.String(), event)
}

// PublishCustomerTierUpdated publishes a customer tier updated event
func (p *kafkaEventPublisher) PublishCustomerTierUpdated(ctx context.Context, customerID uuid.UUID, oldTier, newTier entity.CustomerTier) error {
	event := map[string]interface{}{
		"event_type":  "customer.tier_updated",
		"customer_id": customerID,
		"old_tier":    oldTier,
		"new_tier":    newTier,
		"timestamp":   nil, // Will be set by publishEvent
	}

	return p.publishEvent(ctx, "customer.tier_updated", customerID.String(), event)
}

// PublishLoyverseCustomerSynced publishes a Loyverse customer synced event
func (p *kafkaEventPublisher) PublishLoyverseCustomerSynced(ctx context.Context, customerID uuid.UUID, loyverseID string) error {
	event := map[string]interface{}{
		"event_type":   "customer.loyverse_synced",
		"customer_id":  customerID,
		"loyverse_id":  loyverseID,
		"timestamp":    nil, // Will be set by publishEvent
	}

	return p.publishEvent(ctx, "customer.loyverse_synced", customerID.String(), event)
}

// publishEvent is a helper method to publish events to Kafka
func (p *kafkaEventPublisher) publishEvent(ctx context.Context, eventType, key string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(key),
		Value: data,
		Headers: []kafka.Header{
			{
				Key:   "event_type",
				Value: []byte(eventType),
			},
		},
	}

	err = p.writer.WriteMessages(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

// Close closes the Kafka writer
func (p *kafkaEventPublisher) Close() error {
	return p.writer.Close()
}
