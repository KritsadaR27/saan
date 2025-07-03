package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

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

type EventHandler func(eventType string, data []byte) error

type Message struct {
	Topic     string `json:"topic"`
	Value     []byte `json:"value"`
	Partition int32  `json:"partition"`
	Offset    int64  `json:"offset"`
}

func NewConsumer(brokers, groupID string) *Consumer {
	ctx, cancel := context.WithCancel(context.Background())
	return &Consumer{
		topics:   []string{"loyverse-events"},
		logger:   logrus.New(),
		handlers: make(map[string]EventHandler),
		brokers:  strings.Split(brokers, ","),
		groupID:  groupID,
		ctx:      ctx,
		cancel:   cancel,
		ready:    make(chan bool),
	}
}

func (c *Consumer) RegisterHandler(eventType string, handler EventHandler) {
	c.handlers[eventType] = handler
}

func (c *Consumer) StartConsuming() error {
	// Create Kafka reader
	c.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:        c.brokers,
		Topic:          "loyverse-events",
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

	// Parse message as DomainEvent from Loyverse
	var event struct {
		ID            string          `json:"id"`
		Type          string          `json:"type"`
		AggregateID   string          `json:"aggregate_id"`
		AggregateType string          `json:"aggregate_type"`
		Timestamp     string          `json:"timestamp"`
		Version       int             `json:"version"`
		Data          json.RawMessage `json:"data"`
		Source        string          `json:"source"`
	}

	if err := json.Unmarshal(msg.Value, &event); err != nil {
		c.logger.WithError(err).Error("Failed to unmarshal DomainEvent")
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"event_id":   event.ID,
		"event_type": event.Type,
		"source":     event.Source,
	}).Info("Received domain event")

	// Handle different event types using registered handlers
	if handler, exists := c.handlers[event.Type]; exists {
		return handler(event.Type, event.Data)
	}

	// Default handling for known event types
	switch event.Type {
	case "product.updated":
		c.logger.WithField("product_id", event.AggregateID).Info("Received product update event")
		// In a real implementation, you'd update local cache or trigger refresh
	case "inventory.updated":
		c.logger.WithField("variant_id", event.AggregateID).Info("Received inventory update event")
		// In a real implementation, you'd update stock levels
	case "receipt.created":
		c.logger.WithField("receipt_id", event.AggregateID).Info("Received receipt created event")
		// In a real implementation, you'd update stock movements
	default:
		c.logger.WithField("event_type", event.Type).Debug("Unknown event type")
	}

	return nil
}

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
