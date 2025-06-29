package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/saan/order-service/internal/domain"
)

// EventPublisher defines the interface for publishing events
type EventPublisher interface {
	// PublishEvent publishes an event to the message broker
	PublishEvent(ctx context.Context, event *domain.OrderEvent) error
	
	// Close closes the publisher and releases resources
	Close() error
}

// KafkaEventPublisher implements EventPublisher using Apache Kafka
type KafkaEventPublisher struct {
	producer *kafka.Producer
	topic    string
}

// NewKafkaEventPublisher creates a new Kafka event publisher
func NewKafkaEventPublisher(brokers []string, topic string) (*KafkaEventPublisher, error) {
	config := &kafka.ConfigMap{
		"bootstrap.servers": joinStrings(brokers, ","),
		"client.id":         "order-service-publisher",
		"acks":              "all",
		"retries":           10,
		"retry.backoff.ms":  100,
		"delivery.timeout.ms": 300000,
	}

	producer, err := kafka.NewProducer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &KafkaEventPublisher{
		producer: producer,
		topic:    topic,
	}, nil
}

// PublishEvent publishes an event to Kafka
func (p *KafkaEventPublisher) PublishEvent(ctx context.Context, event *domain.OrderEvent) error {
	// Convert event to JSON
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create Kafka message
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &p.topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(event.OrderID.String()),
		Value: eventData,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(event.EventType)},
			{Key: "event_id", Value: []byte(event.ID.String())},
		},
	}

	// Publish with delivery callback
	deliveryChan := make(chan kafka.Event)
	err = p.producer.Produce(message, deliveryChan)
	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	// Wait for delivery confirmation
	select {
	case e := <-deliveryChan:
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				return fmt.Errorf("delivery failed: %w", ev.TopicPartition.Error)
			}
			log.Printf("Successfully delivered event %s to topic %s[%d] at offset %v",
				event.ID, *ev.TopicPartition.Topic, ev.TopicPartition.Partition, ev.TopicPartition.Offset)
		case kafka.Error:
			return fmt.Errorf("delivery error: %w", ev)
		}
	case <-ctx.Done():
		return fmt.Errorf("context cancelled while waiting for delivery: %w", ctx.Err())
	}

	return nil
}

// Close closes the Kafka producer
func (p *KafkaEventPublisher) Close() error {
	p.producer.Close()
	return nil
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}
	
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
