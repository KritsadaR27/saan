package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"inventory/internal/config"
)

// KafkaPublisher implements the Publisher interface for Kafka
type KafkaPublisher struct {
	writer *kafka.Writer
	logger *logrus.Logger
	config config.KafkaConfig
}

// NewKafkaPublisher creates a new Kafka event publisher
func NewKafkaPublisher(cfg config.KafkaConfig, logger *logrus.Logger) Publisher {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topics.InventoryEvents,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: 1,
		Async:        false,
	}

	return &KafkaPublisher{
		writer: writer,
		logger: logger,
		config: cfg,
	}
}

// Publish publishes an event to Kafka (generic interface)
func (p *KafkaPublisher) publishStockEvent(ctx context.Context, event *StockEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(event.AggregateID),
		Value: data,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte(event.EventType)},
			{Key: "event-source", Value: []byte("inventory-service")},
		},
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		p.logger.WithError(err).Errorf("Failed to publish event %s", event.EventID)
		return fmt.Errorf("failed to publish event: %w", err)
	}

	p.logger.WithFields(logrus.Fields{
		"event_id":   event.EventID,
		"event_type": event.EventType,
		"topic":      p.config.Topics.InventoryEvents,
	}).Debug("Event published successfully")

	return nil
}

// publishProductEvent publishes a product event to Kafka
func (p *KafkaPublisher) publishProductEvent(ctx context.Context, event *ProductEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(event.AggregateID),
		Value: data,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte(event.EventType)},
			{Key: "event-source", Value: []byte("inventory-service")},
		},
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		p.logger.WithError(err).Errorf("Failed to publish event %s", event.EventID)
		return fmt.Errorf("failed to publish event: %w", err)
	}

	p.logger.WithFields(logrus.Fields{
		"event_id":   event.EventID,
		"event_type": event.EventType,
		"topic":      p.config.Topics.InventoryEvents,
	}).Debug("Event published successfully")

	return nil
}

// publishSyncEvent publishes a sync event to Kafka
func (p *KafkaPublisher) publishSyncEvent(ctx context.Context, event *LoyverseSyncEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(event.AggregateID),
		Value: data,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte(event.EventType)},
			{Key: "event-source", Value: []byte("inventory-service")},
		},
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		p.logger.WithError(err).Errorf("Failed to publish event %s", event.EventID)
		return fmt.Errorf("failed to publish event: %w", err)
	}

	p.logger.WithFields(logrus.Fields{
		"event_id":   event.EventID,
		"event_type": event.EventType,
		"topic":      p.config.Topics.InventoryEvents,
	}).Debug("Event published successfully")

	return nil
}

// publishAlertEvent publishes an alert event to Kafka
func (p *KafkaPublisher) publishAlertEvent(ctx context.Context, event *AlertEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(event.AggregateID),
		Value: data,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte(event.EventType)},
			{Key: "event-source", Value: []byte("inventory-service")},
		},
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		p.logger.WithError(err).Errorf("Failed to publish event %s", event.EventID)
		return fmt.Errorf("failed to publish event: %w", err)
	}

	p.logger.WithFields(logrus.Fields{
		"event_id":   event.EventID,
		"event_type": event.EventType,
		"topic":      p.config.Topics.InventoryEvents,
	}).Debug("Event published successfully")

	return nil
}

// PublishStockUpdated publishes a stock updated event
func (p *KafkaPublisher) PublishStockUpdated(ctx context.Context, productID string, previousLevel, newLevel int, movementType, reason string) error {
	event := NewStockEvent(StockUpdated, productID, previousLevel, newLevel, movementType, reason)
	return p.publishStockEvent(ctx, event)
}

// PublishStockLevelLow publishes a stock level low event
func (p *KafkaPublisher) PublishStockLevelLow(ctx context.Context, productID string, currentLevel, threshold int, severity, message string) error {
	event := NewAlertEvent(StockLevelLow, productID, currentLevel, threshold, severity, message)
	return p.publishAlertEvent(ctx, event)
}

// PublishProductCreated publishes a product created event
func (p *KafkaPublisher) PublishProductCreated(ctx context.Context, productID, loyverseID string) error {
	event := NewProductEvent(ProductCreated, productID, loyverseID)
	return p.publishProductEvent(ctx, event)
}

// PublishProductUpdated publishes a product updated event
func (p *KafkaPublisher) PublishProductUpdated(ctx context.Context, productID, loyverseID string) error {
	event := NewProductEvent(ProductUpdated, productID, loyverseID)
	return p.publishProductEvent(ctx, event)
}

// PublishProductDeleted publishes a product deleted event
func (p *KafkaPublisher) PublishProductDeleted(ctx context.Context, productID string) error {
	event := NewProductEvent(ProductDeleted, productID, "")
	return p.publishProductEvent(ctx, event)
}

// PublishLoyverseSync publishes a loyverse sync event
func (p *KafkaPublisher) PublishLoyverseSync(ctx context.Context, entityType, entityID, loyverseID, syncStatus, syncDirection string) error {
	event := NewLoyverseSyncEvent(entityType, entityID, loyverseID, syncStatus, syncDirection)
	return p.publishSyncEvent(ctx, event)
}

// Close closes the Kafka writer
func (p *KafkaPublisher) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}

// Consumer handles Kafka event consumption
type Consumer struct {
	topics   []string
	logger   *logrus.Logger
	handlers map[string]EventHandler
	reader   *kafka.Reader
	brokers  []string
	groupID  string
	ctx      context.Context
	cancel   context.CancelFunc
	ready    chan bool
}

// Message represents a Kafka message
type Message struct {
	Topic     string `json:"topic"`
	Value     []byte `json:"value"`
	Partition int32  `json:"partition"`
	Offset    int64  `json:"offset"`
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(cfg config.KafkaConfig, logger *logrus.Logger) *Consumer {
	ctx, cancel := context.WithCancel(context.Background())
	return &Consumer{
		topics:   []string{cfg.Topics.LoyverseEvents, cfg.Topics.InventoryEvents},
		logger:   logger,
		handlers: make(map[string]EventHandler),
		brokers:  cfg.Brokers,
		groupID:  cfg.ConsumerGroup,
		ctx:      ctx,
		cancel:   cancel,
		ready:    make(chan bool),
	}
}

// RegisterHandler registers an event handler for a specific event type
func (c *Consumer) RegisterHandler(eventType string, handler EventHandler) {
	c.handlers[eventType] = handler
}

// StartConsuming starts consuming events from Kafka
func (c *Consumer) StartConsuming() error {
	// Create Kafka reader
	c.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:        c.brokers,
		Topic:          c.topics[0], // Use first topic for now
		GroupID:        c.groupID,
		StartOffset:    kafka.LastOffset,
		CommitInterval: time.Second,
	})

	// Start consuming in a goroutine
	go func() {
		defer func() {
			if err := c.reader.Close(); err != nil {
				c.logger.Errorf("Error closing reader: %v", err)
			}
		}()

		// Mark as ready
		close(c.ready)

		for {
			// Check if context was cancelled
			if c.ctx.Err() != nil {
				return
			}

			// Read message
			msg, err := c.reader.ReadMessage(c.ctx)
			if err != nil {
				if c.ctx.Err() != nil {
					return // Context cancelled
				}
				c.logger.Errorf("Error reading message: %v", err)
				continue
			}

			// Process the message
			message := &Message{
				Topic:     msg.Topic,
				Value:     msg.Value,
				Partition: int32(msg.Partition),
				Offset:    msg.Offset,
			}

			if err := c.processMessage(message); err != nil {
				c.logger.Errorf("Error processing message: %v", err)
			}
		}
	}()

	// Wait for consumer to be ready
	<-c.ready
	c.logger.Info("Kafka consumer started and ready")
	return nil
}

func (c *Consumer) processMessage(msg *Message) error {
	c.logger.WithFields(logrus.Fields{
		"topic":     msg.Topic,
		"partition": msg.Partition,
		"offset":    msg.Offset,
	}).Debug("Processing Kafka message")

	// Parse message as a generic event structure to get the event type
	var eventMeta struct {
		EventID   string `json:"event_id"`
		EventType string `json:"event_type"`
	}

	if err := json.Unmarshal(msg.Value, &eventMeta); err != nil {
		c.logger.WithError(err).Error("Failed to unmarshal event metadata")
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"event_id":   eventMeta.EventID,
		"event_type": eventMeta.EventType,
	}).Info("Received domain event")

	// Handle different event types using registered handlers
	if handler, exists := c.handlers[eventMeta.EventType]; exists {
		return handler(eventMeta.EventType, msg.Value)
	}

	// Default handling for known event types
	switch eventMeta.EventType {
	case ProductUpdated:
		c.logger.Info("Received product update event")
		// In a real implementation, you'd update local cache or trigger refresh
	case StockUpdated:
		c.logger.Info("Received stock update event")
		// In a real implementation, you'd update stock levels
	case StockMovement:
		c.logger.Info("Received stock movement event")
		// In a real implementation, you'd update stock movements
	case LoyverseSync:
		c.logger.Info("Received loyverse sync event")
		// In a real implementation, you'd handle sync operations
	default:
		c.logger.WithField("event_type", eventMeta.EventType).Debug("Unknown event type")
	}

	return nil
}

// Close closes the Kafka consumer
func (c *Consumer) Close() error {
	c.logger.Info("Closing Kafka consumer...")

	// Cancel the context to stop consuming
	if c.cancel != nil {
		c.cancel()
	}

	// Close the reader
	if c.reader != nil {
		if err := c.reader.Close(); err != nil {
			c.logger.Errorf("Error closing reader: %v", err)
			return err
		}
	}

	c.logger.Info("Kafka consumer closed")
	return nil
}
