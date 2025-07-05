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

// PublishStockUpdated does nothing
func (p *NoopPublisher) PublishStockUpdated(ctx context.Context, productID string, previousLevel, newLevel int, movementType, reason string) error {
	p.logger.Debug("Noop: stock updated event not published")
	return nil
}

// PublishStockLevelLow does nothing
func (p *NoopPublisher) PublishStockLevelLow(ctx context.Context, productID string, currentLevel, threshold int, severity, message string) error {
	p.logger.Debug("Noop: stock level low event not published")
	return nil
}

// PublishProductCreated does nothing
func (p *NoopPublisher) PublishProductCreated(ctx context.Context, productID, loyverseID string) error {
	p.logger.Debug("Noop: product created event not published")
	return nil
}

// PublishProductUpdated does nothing
func (p *NoopPublisher) PublishProductUpdated(ctx context.Context, productID, loyverseID string) error {
	p.logger.Debug("Noop: product updated event not published")
	return nil
}

// PublishProductDeleted does nothing
func (p *NoopPublisher) PublishProductDeleted(ctx context.Context, productID string) error {
	p.logger.Debug("Noop: product deleted event not published")
	return nil
}

// PublishLoyverseSync does nothing
func (p *NoopPublisher) PublishLoyverseSync(ctx context.Context, entityType, entityID, loyverseID, syncStatus, syncDirection string) error {
	p.logger.Debug("Noop: loyverse sync event not published")
	return nil
}

// Close does nothing
func (p *NoopPublisher) Close() error {
	return nil
}
