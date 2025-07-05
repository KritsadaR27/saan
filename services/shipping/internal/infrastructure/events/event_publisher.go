package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// EventPublisher implements the event publisher interface using Kafka
type EventPublisher struct {
	writer   *kafka.Writer
	topic    string
	clientID string
}

// NewEventPublisher creates a new Kafka event publisher
func NewEventPublisher(brokers []string, topic, clientID string) *EventPublisher {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
		BatchTimeout: 10 * time.Millisecond,
		BatchSize:    100,
	}

	return &EventPublisher{
		writer:   writer,
		topic:    topic,
		clientID: clientID,
	}
}

// NewKafkaProducer creates a new Kafka producer with default settings
func NewKafkaProducer(brokers []string, topic, clientID string) (*EventPublisher, error) {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Compression:  kafka.Snappy,
	}
	
	return &EventPublisher{
		writer:   writer,
		topic:    topic,
		clientID: clientID,
	}, nil
}

// Publish publishes an event to Kafka
func (p *EventPublisher) Publish(ctx context.Context, eventType string, data interface{}) error {
	// Create event envelope
	event := map[string]interface{}{
		"event_type":  eventType,
		"data":        data,
		"timestamp":   time.Now().UTC(),
		"source":      "shipping-service",
		"client_id":   p.clientID,
		"version":     "1.0",
	}

	// Serialize event to JSON
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create Kafka message
	message := kafka.Message{
		Key:   []byte(eventType),
		Value: eventData,
		Headers: []kafka.Header{
			{
				Key:   "event-type",
				Value: []byte(eventType),
			},
			{
				Key:   "source",
				Value: []byte("shipping-service"),
			},
			{
				Key:   "timestamp",
				Value: []byte(fmt.Sprintf("%d", time.Now().Unix())),
			},
		},
	}

	// Write message to Kafka
	err = p.writer.WriteMessages(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to publish event to kafka: %w", err)
	}

	return nil
}

// PublishBatch publishes multiple events in a single batch
func (p *EventPublisher) PublishBatch(ctx context.Context, events []map[string]interface{}) error {
	if len(events) == 0 {
		return nil
	}

	messages := make([]kafka.Message, 0, len(events))

	for _, event := range events {
		// Add metadata to event
		event["timestamp"] = time.Now().UTC()
		event["source"] = "shipping-service"
		event["client_id"] = p.clientID
		event["version"] = "1.0"

		// Serialize event
		eventData, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal batch event: %w", err)
		}

		// Get event type
		eventType, ok := event["event_type"].(string)
		if !ok {
			eventType = "unknown"
		}

		// Create message
		message := kafka.Message{
			Key:   []byte(eventType),
			Value: eventData,
			Headers: []kafka.Header{
				{
					Key:   "event-type",
					Value: []byte(eventType),
				},
				{
					Key:   "source",
					Value: []byte("shipping-service"),
				},
			},
		}

		messages = append(messages, message)
	}

	// Write batch to Kafka
	err := p.writer.WriteMessages(ctx, messages...)
	if err != nil {
		return fmt.Errorf("failed to publish batch events to kafka: %w", err)
	}

	return nil
}

// PublishDeliveryEvent publishes delivery-specific events
func (p *EventPublisher) PublishDeliveryEvent(ctx context.Context, deliveryID, eventType string, data interface{}) error {
	event := map[string]interface{}{
		"event_type":   eventType,
		"delivery_id":  deliveryID,
		"data":         data,
		"timestamp":    time.Now().UTC(),
		"source":       "shipping-service",
		"client_id":    p.clientID,
		"version":      "1.0",
	}

	return p.publishEvent(ctx, event, eventType, deliveryID)
}

// PublishVehicleEvent publishes vehicle-specific events
func (p *EventPublisher) PublishVehicleEvent(ctx context.Context, vehicleID, eventType string, data interface{}) error {
	event := map[string]interface{}{
		"event_type":  eventType,
		"vehicle_id":  vehicleID,
		"data":        data,
		"timestamp":   time.Now().UTC(),
		"source":      "shipping-service",
		"client_id":   p.clientID,
		"version":     "1.0",
	}

	return p.publishEvent(ctx, event, eventType, vehicleID)
}

// PublishRouteEvent publishes route-specific events
func (p *EventPublisher) PublishRouteEvent(ctx context.Context, routeID, eventType string, data interface{}) error {
	event := map[string]interface{}{
		"event_type": eventType,
		"route_id":   routeID,
		"data":       data,
		"timestamp":  time.Now().UTC(),
		"source":     "shipping-service",
		"client_id":  p.clientID,
		"version":    "1.0",
	}

	return p.publishEvent(ctx, event, eventType, routeID)
}

// publishEvent is a helper method to publish events
func (p *EventPublisher) publishEvent(ctx context.Context, event map[string]interface{}, eventType, entityID string) error {
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(entityID),
		Value: eventData,
		Headers: []kafka.Header{
			{
				Key:   "event-type",
				Value: []byte(eventType),
			},
			{
				Key:   "entity-id",
				Value: []byte(entityID),
			},
			{
				Key:   "source",
				Value: []byte("shipping-service"),
			},
		},
	}

	err = p.writer.WriteMessages(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to publish %s event: %w", eventType, err)
	}

	return nil
}

// Close closes the Kafka writer
func (p *EventPublisher) Close() error {
	return p.writer.Close()
}

// Health checks if Kafka connection is healthy
func (p *EventPublisher) Health(ctx context.Context) error {
	// Try to create a test connection
	conn, err := kafka.DialContext(ctx, "tcp", p.writer.Addr.String())
	if err != nil {
		return fmt.Errorf("kafka health check failed: %w", err)
	}
	defer conn.Close()

	return nil
}
