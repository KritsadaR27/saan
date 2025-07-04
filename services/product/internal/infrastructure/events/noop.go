package events

import "context"

// NoOpPublisher is a no-operation publisher for development/testing
type NoOpPublisher struct{}

// NewNoOpPublisher creates a new no-op publisher
func NewNoOpPublisher() *NoOpPublisher {
	return &NoOpPublisher{}
}

// Publish does nothing
func (n *NoOpPublisher) Publish(ctx context.Context, topic string, event interface{}) error {
	return nil
}

// PublishAsync does nothing
func (n *NoOpPublisher) PublishAsync(ctx context.Context, topic string, event interface{}) error {
	return nil
}

// PublishProductEvent does nothing
func (n *NoOpPublisher) PublishProductEvent(ctx context.Context, event *ProductEvent) error {
	return nil
}

// PublishCategoryEvent does nothing
func (n *NoOpPublisher) PublishCategoryEvent(ctx context.Context, event *CategoryEvent) error {
	return nil
}

// PublishPricingEvent does nothing
func (n *NoOpPublisher) PublishPricingEvent(ctx context.Context, event *PricingEvent) error {
	return nil
}

// PublishInventoryEvent does nothing
func (n *NoOpPublisher) PublishInventoryEvent(ctx context.Context, event *InventoryEvent) error {
	return nil
}

// PublishSyncEvent does nothing
func (n *NoOpPublisher) PublishSyncEvent(ctx context.Context, event *SyncEvent) error {
	return nil
}

// Close does nothing
func (n *NoOpPublisher) Close() error {
	return nil
}

// IsHealthy always returns true
func (n *NoOpPublisher) IsHealthy() bool {
	return true
}
