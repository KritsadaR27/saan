package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"customer/internal/domain/entity"
)

// KafkaPublisher implements Publisher interface using Kafka
type KafkaPublisher struct {
	writers map[string]*kafka.Writer
	logger  *zap.Logger
	brokers []string
}

// NewKafkaPublisher creates a new Kafka publisher
func NewKafkaPublisher(brokers []string, logger *zap.Logger) Publisher {
	writers := make(map[string]*kafka.Writer)
	
	// Create writers for each topic following SAAN standards
	topics := []string{
		CustomerEventsTopic,
		AnalyticsEventsTopic,
		SyncEventsTopic,
	}
	
	for _, topic := range topics {
		writers[topic] = &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireOne,
			Async:        false, // Synchronous by default
			WriteTimeout: 10 * time.Second,
			ReadTimeout:  10 * time.Second,
		}
	}

	return &KafkaPublisher{
		writers: writers,
		logger:  logger,
		brokers: brokers,
	}
}

// PublishCustomerCreated publishes a customer created event
func (p *KafkaPublisher) PublishCustomerCreated(ctx context.Context, customer *entity.Customer) error {
	event := NewCustomerEvent(CustomerCreated, customer.ID, customer)
	return p.publishEvent(ctx, CustomerEventsTopic, event.CustomerID.String(), event)
}

// PublishCustomerUpdated publishes a customer updated event
func (p *KafkaPublisher) PublishCustomerUpdated(ctx context.Context, customer *entity.Customer) error {
	event := NewCustomerEvent(CustomerUpdated, customer.ID, customer)
	return p.publishEvent(ctx, CustomerEventsTopic, event.CustomerID.String(), event)
}

// PublishCustomerDeleted publishes a customer deleted event
func (p *KafkaPublisher) PublishCustomerDeleted(ctx context.Context, customerID uuid.UUID) error {
	event := NewCustomerEvent(CustomerDeleted, customerID, nil)
	return p.publishEvent(ctx, CustomerEventsTopic, event.CustomerID.String(), event)
}

// PublishCustomerTierUpdated publishes a customer tier updated event (domain interface)
func (p *KafkaPublisher) PublishCustomerTierUpdated(ctx context.Context, customerID uuid.UUID, oldTier, newTier entity.CustomerTier) error {
	return p.PublishCustomerTierUpdatedWithReason(ctx, customerID, oldTier, newTier, "")
}

// PublishCustomerTierUpdatedWithReason publishes a customer tier updated event with reason
func (p *KafkaPublisher) PublishCustomerTierUpdatedWithReason(ctx context.Context, customerID uuid.UUID, oldTier, newTier entity.CustomerTier, reason string) error {
	event := NewCustomerTierEvent(customerID, oldTier, newTier, reason)
	return p.publishEvent(ctx, CustomerEventsTopic, event.CustomerID.String(), event)
}

// PublishLoyverseCustomerSynced publishes a Loyverse customer synced event (domain interface)
func (p *KafkaPublisher) PublishLoyverseCustomerSynced(ctx context.Context, customerID uuid.UUID, loyverseID string) error {
	return p.PublishLoyverseSyncedWithStatus(ctx, customerID, loyverseID, "success")
}

// PublishLoyverseSyncedWithStatus publishes a Loyverse customer synced event with status
func (p *KafkaPublisher) PublishLoyverseSyncedWithStatus(ctx context.Context, customerID uuid.UUID, loyverseID string, syncStatus string) error {
	event := NewSyncEvent("loyverse", "customer", customerID, loyverseID, syncStatus)
	return p.publishEvent(ctx, SyncEventsTopic, event.EntityID.String(), event)
}

// PublishCustomerPointsUpdated publishes a customer points updated event
func (p *KafkaPublisher) PublishCustomerPointsUpdated(ctx context.Context, customerID uuid.UUID, pointsChange, totalPoints int, transactionType, description string) error {
	event := NewCustomerPointsEvent(customerID, pointsChange, totalPoints, transactionType, description)
	return p.publishEvent(ctx, AnalyticsEventsTopic, event.CustomerID.String(), event)
}

// PublishCustomerAddressAdded publishes a customer address added event
func (p *KafkaPublisher) PublishCustomerAddressAdded(ctx context.Context, customerID, addressID uuid.UUID, address *entity.CustomerAddress) error {
	event := NewCustomerAddressEvent(CustomerAddressAdded, customerID, addressID, address)
	return p.publishEvent(ctx, CustomerEventsTopic, event.CustomerID.String(), event)
}

// PublishCustomerAddressUpdated publishes a customer address updated event
func (p *KafkaPublisher) PublishCustomerAddressUpdated(ctx context.Context, customerID, addressID uuid.UUID, address *entity.CustomerAddress) error {
	event := NewCustomerAddressEvent(CustomerAddressUpdated, customerID, addressID, address)
	return p.publishEvent(ctx, CustomerEventsTopic, event.CustomerID.String(), event)
}

// PublishCustomerAddressDeleted publishes a customer address deleted event
func (p *KafkaPublisher) PublishCustomerAddressDeleted(ctx context.Context, customerID, addressID uuid.UUID) error {
	event := NewCustomerAddressEvent(CustomerAddressDeleted, customerID, addressID, nil)
	return p.publishEvent(ctx, CustomerEventsTopic, event.CustomerID.String(), event)
}

// publishEvent is a helper method to publish events to Kafka
func (p *KafkaPublisher) publishEvent(ctx context.Context, topic, key string, payload interface{}) error {
	writer, exists := p.writers[topic]
	if !exists {
		return fmt.Errorf("writer for topic %s not found", topic)
	}

	data, err := json.Marshal(payload)
	if err != nil {
		p.logger.Error("Failed to marshal event payload", zap.Error(err), zap.String("topic", topic))
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(key),
		Value: data,
		Headers: []kafka.Header{
			{
				Key:   "event_type",
				Value: []byte(fmt.Sprintf("%T", payload)),
			},
		},
	}

	err = writer.WriteMessages(ctx, message)
	if err != nil {
		p.logger.Error("Failed to publish event", zap.Error(err), zap.String("topic", topic), zap.String("key", key))
		return fmt.Errorf("failed to publish event: %w", err)
	}

	p.logger.Debug("Event published successfully", zap.String("topic", topic), zap.String("key", key))
	return nil
}

// Close closes all Kafka writers
func (p *KafkaPublisher) Close() error {
	var lastErr error
	for topic, writer := range p.writers {
		if err := writer.Close(); err != nil {
			p.logger.Error("Failed to close writer", zap.Error(err), zap.String("topic", topic))
			lastErr = err
		}
	}
	return lastErr
}
