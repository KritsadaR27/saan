package events

import (
	"context"

	"github.com/sirupsen/logrus"
)

// NoopPublisher is a no-operation implementation of the Publisher interface
// Useful for testing or when events are disabled
type NoopPublisher struct {
	logger *logrus.Logger
}

// NewNoopPublisher creates a new noop event publisher
func NewNoopPublisher(logger *logrus.Logger) Publisher {
	return &NoopPublisher{
		logger: logger,
	}
}

// PublishOrderCreated does nothing
func (p *NoopPublisher) PublishOrderCreated(ctx context.Context, orderID, customerID string, orderData map[string]interface{}) error {
	p.logger.Debug("Noop: order created event not published")
	return nil
}

// PublishOrderUpdated does nothing
func (p *NoopPublisher) PublishOrderUpdated(ctx context.Context, orderID, customerID string, changes map[string]interface{}) error {
	p.logger.Debug("Noop: order updated event not published")
	return nil
}

// PublishOrderCancelled does nothing
func (p *NoopPublisher) PublishOrderCancelled(ctx context.Context, orderID, customerID string, reason string) error {
	p.logger.Debug("Noop: order cancelled event not published")
	return nil
}

// PublishOrderCompleted does nothing
func (p *NoopPublisher) PublishOrderCompleted(ctx context.Context, orderID, customerID string) error {
	p.logger.Debug("Noop: order completed event not published")
	return nil
}

// PublishPaymentProcessed does nothing
func (p *NoopPublisher) PublishPaymentProcessed(ctx context.Context, orderID, paymentID string, amount float64, currency string) error {
	p.logger.Debug("Noop: payment processed event not published")
	return nil
}

// PublishPaymentFailed does nothing
func (p *NoopPublisher) PublishPaymentFailed(ctx context.Context, orderID, paymentID string, amount float64, currency string, reason string) error {
	p.logger.Debug("Noop: payment failed event not published")
	return nil
}

// PublishInventoryReserved does nothing
func (p *NoopPublisher) PublishInventoryReserved(ctx context.Context, orderID, productID string, quantity int, reservationID string) error {
	p.logger.Debug("Noop: inventory reserved event not published")
	return nil
}

// PublishInventoryReleased does nothing
func (p *NoopPublisher) PublishInventoryReleased(ctx context.Context, orderID, productID string, quantity int, reservationID string) error {
	p.logger.Debug("Noop: inventory released event not published")
	return nil
}

// PublishNotificationSent does nothing
func (p *NoopPublisher) PublishNotificationSent(ctx context.Context, orderID, customerID string, notificationType, channel string, data map[string]interface{}) error {
	p.logger.Debug("Noop: notification sent event not published")
	return nil
}

// Close does nothing
func (p *NoopPublisher) Close() error {
	return nil
}
