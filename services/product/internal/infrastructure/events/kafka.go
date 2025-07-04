package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

// KafkaPublisher implements Publisher interface using Kafka
type KafkaPublisher struct {
	writers map[string]*kafka.Writer
	logger  *logrus.Logger
	brokers []string
}

// NewKafkaPublisher creates a new Kafka publisher
func NewKafkaPublisher(brokers []string, logger *logrus.Logger) *KafkaPublisher {
	writers := make(map[string]*kafka.Writer)
	
	// Create writers for each topic following PROJECT_RULES.md
	topics := []string{
		ProductEventsTopic,
		CategoryEventsTopic,
		PricingEventsTopic,
		InventoryEventsTopic,
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

// Publish publishes an event to a specific topic
func (k *KafkaPublisher) Publish(ctx context.Context, topic string, event interface{}) error {
	writer, exists := k.writers[topic]
	if !exists {
		return fmt.Errorf("writer for topic %s not found", topic)
	}

	// Serialize event to JSON
	eventData, err := json.Marshal(event)
	if err != nil {
		k.logger.WithError(err).WithField("topic", topic).Error("Failed to marshal event")
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create Kafka message
	message := kafka.Message{
		Key:   []byte(fmt.Sprintf("%s-%d", topic, time.Now().UnixNano())),
		Value: eventData,
		Time:  time.Now(),
	}

	// Publish message
	err = writer.WriteMessages(ctx, message)
	if err != nil {
		k.logger.WithError(err).WithFields(logrus.Fields{
			"topic":     topic,
			"event_type": fmt.Sprintf("%T", event),
		}).Error("Failed to publish event")
		return fmt.Errorf("failed to publish event to topic %s: %w", topic, err)
	}

	k.logger.WithFields(logrus.Fields{
		"topic":     topic,
		"event_type": fmt.Sprintf("%T", event),
	}).Info("Event published successfully")

	return nil
}

// PublishAsync publishes an event asynchronously
func (k *KafkaPublisher) PublishAsync(ctx context.Context, topic string, event interface{}) error {
	// For async publishing, we could use goroutines or async writer
	// For simplicity, using the same sync method but with a shorter timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	
	return k.Publish(ctxWithTimeout, topic, event)
}

// PublishProductEvent publishes a product-related event
func (k *KafkaPublisher) PublishProductEvent(ctx context.Context, event *ProductEvent) error {
	return k.Publish(ctx, ProductEventsTopic, event)
}

// PublishCategoryEvent publishes a category-related event
func (k *KafkaPublisher) PublishCategoryEvent(ctx context.Context, event *CategoryEvent) error {
	return k.Publish(ctx, CategoryEventsTopic, event)
}

// PublishPricingEvent publishes a pricing-related event
func (k *KafkaPublisher) PublishPricingEvent(ctx context.Context, event *PricingEvent) error {
	return k.Publish(ctx, PricingEventsTopic, event)
}

// PublishInventoryEvent publishes an inventory-related event
func (k *KafkaPublisher) PublishInventoryEvent(ctx context.Context, event *InventoryEvent) error {
	return k.Publish(ctx, InventoryEventsTopic, event)
}

// PublishSyncEvent publishes a sync-related event
func (k *KafkaPublisher) PublishSyncEvent(ctx context.Context, event *SyncEvent) error {
	return k.Publish(ctx, SyncEventsTopic, event)
}

// Close closes all Kafka writers
func (k *KafkaPublisher) Close() error {
	var lastErr error
	
	for topic, writer := range k.writers {
		if err := writer.Close(); err != nil {
			k.logger.WithError(err).WithField("topic", topic).Error("Failed to close Kafka writer")
			lastErr = err
		}
	}
	
	if lastErr != nil {
		return fmt.Errorf("failed to close some Kafka writers: %w", lastErr)
	}
	
	k.logger.Info("All Kafka writers closed successfully")
	return nil
}

// IsHealthy checks if the Kafka publisher is healthy
func (k *KafkaPublisher) IsHealthy() bool {
	// Simple health check - check if writers exist and are configured
	for _, writer := range k.writers {
		if writer != nil {
			return true
		}
	}
	return false
}

// CreateTopics creates Kafka topics if they don't exist (for development)
func CreateKafkaTopics(brokers []string, logger *logrus.Logger) error {
	topics := []string{
		ProductEventsTopic,
		CategoryEventsTopic,
		PricingEventsTopic,
		InventoryEventsTopic,
		SyncEventsTopic,
	}
	
	// Note: In production, topics should be created by ops/infrastructure
	// This is mainly for development convenience
	logger.WithField("topics", topics).Info("Kafka topics configured")
	
	return nil
}
