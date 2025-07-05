package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"order/internal/infrastructure/config"
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
		Topic:        cfg.Topics.OrderEvents,
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

// publishOrderEvent publishes an order event to Kafka
func (p *KafkaPublisher) publishOrderEvent(ctx context.Context, event *OrderEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(event.AggregateID),
		Value: data,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte(event.EventType)},
			{Key: "event-source", Value: []byte("order-service")},
		},
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		p.logger.WithError(err).Errorf("Failed to publish event %s", event.EventID)
		return fmt.Errorf("failed to publish event: %w", err)
	}

	p.logger.WithFields(logrus.Fields{
		"event_id":   event.EventID,
		"event_type": event.EventType,
		"topic":      p.config.Topics.OrderEvents,
	}).Debug("Event published successfully")

	return nil
}

// publishPaymentEvent publishes a payment event to Kafka
func (p *KafkaPublisher) publishPaymentEvent(ctx context.Context, event *PaymentEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(event.AggregateID),
		Value: data,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte(event.EventType)},
			{Key: "event-source", Value: []byte("order-service")},
		},
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		p.logger.WithError(err).Errorf("Failed to publish event %s", event.EventID)
		return fmt.Errorf("failed to publish event: %w", err)
	}

	p.logger.WithFields(logrus.Fields{
		"event_id":   event.EventID,
		"event_type": event.EventType,
		"topic":      p.config.Topics.PaymentEvents,
	}).Debug("Payment event published successfully")

	return nil
}

// publishInventoryEvent publishes an inventory event to Kafka
func (p *KafkaPublisher) publishInventoryEvent(ctx context.Context, event *InventoryEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(event.AggregateID),
		Value: data,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte(event.EventType)},
			{Key: "event-source", Value: []byte("order-service")},
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
	}).Debug("Inventory event published successfully")

	return nil
}

// publishNotificationEvent publishes a notification event to Kafka
func (p *KafkaPublisher) publishNotificationEvent(ctx context.Context, event *NotificationEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(event.AggregateID),
		Value: data,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte(event.EventType)},
			{Key: "event-source", Value: []byte("order-service")},
		},
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		p.logger.WithError(err).Errorf("Failed to publish event %s", event.EventID)
		return fmt.Errorf("failed to publish event: %w", err)
	}

	p.logger.WithFields(logrus.Fields{
		"event_id":   event.EventID,
		"event_type": event.EventType,
		"topic":      p.config.Topics.NotificationEvents,
	}).Debug("Notification event published successfully")

	return nil
}

// PublishOrderCreated publishes an order created event
func (p *KafkaPublisher) PublishOrderCreated(ctx context.Context, orderID, customerID string, orderData map[string]interface{}) error {
	event := NewOrderEvent(OrderCreated, orderID, customerID)
	event.OrderData = orderData
	return p.publishOrderEvent(ctx, event)
}

// PublishOrderUpdated publishes an order updated event
func (p *KafkaPublisher) PublishOrderUpdated(ctx context.Context, orderID, customerID string, changes map[string]interface{}) error {
	event := NewOrderEvent(OrderUpdated, orderID, customerID)
	event.Changes = changes
	return p.publishOrderEvent(ctx, event)
}

// PublishOrderCancelled publishes an order cancelled event
func (p *KafkaPublisher) PublishOrderCancelled(ctx context.Context, orderID, customerID string, reason string) error {
	event := NewOrderEvent(OrderCancelled, orderID, customerID)
	event.Changes = map[string]interface{}{"reason": reason}
	return p.publishOrderEvent(ctx, event)
}

// PublishOrderCompleted publishes an order completed event
func (p *KafkaPublisher) PublishOrderCompleted(ctx context.Context, orderID, customerID string) error {
	event := NewOrderEvent(OrderCompleted, orderID, customerID)
	return p.publishOrderEvent(ctx, event)
}

// PublishPaymentProcessed publishes a payment processed event
func (p *KafkaPublisher) PublishPaymentProcessed(ctx context.Context, orderID, paymentID string, amount float64, currency string) error {
	event := NewPaymentEvent(PaymentProcessed, orderID, paymentID, amount, currency)
	event.Status = "success"
	return p.publishPaymentEvent(ctx, event)
}

// PublishPaymentFailed publishes a payment failed event
func (p *KafkaPublisher) PublishPaymentFailed(ctx context.Context, orderID, paymentID string, amount float64, currency string, reason string) error {
	event := NewPaymentEvent(PaymentFailed, orderID, paymentID, amount, currency)
	event.Status = "failed"
	event.ProviderData = map[string]interface{}{"reason": reason}
	return p.publishPaymentEvent(ctx, event)
}

// PublishInventoryReserved publishes an inventory reserved event
func (p *KafkaPublisher) PublishInventoryReserved(ctx context.Context, orderID, productID string, quantity int, reservationID string) error {
	event := NewInventoryEvent(InventoryReserved, orderID, productID, quantity, "reserve")
	event.ReservationID = reservationID
	return p.publishInventoryEvent(ctx, event)
}

// PublishInventoryReleased publishes an inventory released event
func (p *KafkaPublisher) PublishInventoryReleased(ctx context.Context, orderID, productID string, quantity int, reservationID string) error {
	event := NewInventoryEvent(InventoryReleased, orderID, productID, quantity, "release")
	event.ReservationID = reservationID
	return p.publishInventoryEvent(ctx, event)
}

// PublishNotificationSent publishes a notification sent event
func (p *KafkaPublisher) PublishNotificationSent(ctx context.Context, orderID, customerID string, notificationType, channel string, data map[string]interface{}) error {
	event := NewNotificationEvent(NotificationSent, orderID, customerID, notificationType, channel)
	event.Data = data
	return p.publishNotificationEvent(ctx, event)
}

// Close closes the Kafka writer
func (p *KafkaPublisher) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}
