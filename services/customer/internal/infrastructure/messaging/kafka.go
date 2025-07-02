package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"

	"github.com/saan-system/services/customer/internal/domain"
)

// KafkaProducer interface for Kafka operations
type KafkaProducer interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type eventPublisher struct {
	producer KafkaProducer
}

// CustomerCreatedEvent represents a customer created event
type CustomerCreatedEvent struct {
	CustomerID uuid.UUID       `json:"customer_id"`
	Customer   domain.Customer `json:"customer"`
	Timestamp  string          `json:"timestamp"`
}

// CustomerUpdatedEvent represents a customer updated event
type CustomerUpdatedEvent struct {
	CustomerID uuid.UUID       `json:"customer_id"`
	Customer   domain.Customer `json:"customer"`
	Timestamp  string          `json:"timestamp"`
}

// CustomerDeletedEvent represents a customer deleted event
type CustomerDeletedEvent struct {
	CustomerID uuid.UUID `json:"customer_id"`
	Timestamp  string    `json:"timestamp"`
}

// CustomerTierUpdatedEvent represents a customer tier updated event
type CustomerTierUpdatedEvent struct {
	CustomerID uuid.UUID            `json:"customer_id"`
	OldTier    domain.CustomerTier  `json:"old_tier"`
	NewTier    domain.CustomerTier  `json:"new_tier"`
	Timestamp  string               `json:"timestamp"`
}

// LoyverseCustomerSyncedEvent represents a Loyverse customer synced event
type LoyverseCustomerSyncedEvent struct {
	CustomerID  uuid.UUID `json:"customer_id"`
	LoyverseID  string    `json:"loyverse_id"`
	Timestamp   string    `json:"timestamp"`
}

// NewProducer creates a new Kafka producer
func NewProducer() (KafkaProducer, error) {
	brokers := getKafkaBrokers()
	
	producer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Balancer:               &kafka.LeastBytes{},
		RequiredAcks:           kafka.RequireOne,
		AllowAutoTopicCreation: true,
	}

	return producer, nil
}

// NewEventPublisher creates a new event publisher
func NewEventPublisher(producer KafkaProducer) domain.EventPublisher {
	return &eventPublisher{producer: producer}
}

// PublishCustomerCreated publishes a customer created event
func (p *eventPublisher) PublishCustomerCreated(ctx context.Context, customer *domain.Customer) error {
	event := CustomerCreatedEvent{
		CustomerID: customer.ID,
		Customer:   *customer,
		Timestamp:  customer.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return p.publishEvent(ctx, "customer.created", event)
}

// PublishCustomerUpdated publishes a customer updated event
func (p *eventPublisher) PublishCustomerUpdated(ctx context.Context, customer *domain.Customer) error {
	event := CustomerUpdatedEvent{
		CustomerID: customer.ID,
		Customer:   *customer,
		Timestamp:  customer.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return p.publishEvent(ctx, "customer.updated", event)
}

// PublishCustomerDeleted publishes a customer deleted event
func (p *eventPublisher) PublishCustomerDeleted(ctx context.Context, customerID uuid.UUID) error {
	event := CustomerDeletedEvent{
		CustomerID: customerID,
		Timestamp:  fmt.Sprintf("%d", customerID),
	}

	return p.publishEvent(ctx, "customer.deleted", event)
}

// PublishCustomerTierUpdated publishes a customer tier updated event
func (p *eventPublisher) PublishCustomerTierUpdated(ctx context.Context, customerID uuid.UUID, oldTier, newTier domain.CustomerTier) error {
	event := CustomerTierUpdatedEvent{
		CustomerID: customerID,
		OldTier:    oldTier,
		NewTier:    newTier,
		Timestamp:  fmt.Sprintf("%d", customerID),
	}

	return p.publishEvent(ctx, "customer.tier.updated", event)
}

// PublishLoyverseCustomerSynced publishes a Loyverse customer synced event
func (p *eventPublisher) PublishLoyverseCustomerSynced(ctx context.Context, customerID uuid.UUID, loyverseID string) error {
	event := LoyverseCustomerSyncedEvent{
		CustomerID: customerID,
		LoyverseID: loyverseID,
		Timestamp:  fmt.Sprintf("%d", customerID),
	}

	return p.publishEvent(ctx, "customer.loyverse.synced", event)
}

// publishEvent publishes an event to Kafka
func (p *eventPublisher) publishEvent(ctx context.Context, topic string, event interface{}) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	message := kafka.Message{
		Topic: topic,
		Value: data,
	}

	if err := p.producer.WriteMessages(ctx, message); err != nil {
		return fmt.Errorf("failed to publish event to topic %s: %w", topic, err)
	}

	return nil
}

// getKafkaBrokers returns Kafka broker addresses
func getKafkaBrokers() []string {
	brokers := getEnv("KAFKA_BROKERS", "kafka:9092")
	return []string{brokers}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
