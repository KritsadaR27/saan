package events

import (
	"context"
)

// Publisher interface for event publishing
type Publisher interface {
	// Basic publishing
	Publish(ctx context.Context, topic string, event interface{}) error
	PublishAsync(ctx context.Context, topic string, event interface{}) error
	
	// Product events
	PublishProductEvent(ctx context.Context, event *ProductEvent) error
	PublishCategoryEvent(ctx context.Context, event *CategoryEvent) error
	PublishPricingEvent(ctx context.Context, event *PricingEvent) error
	PublishInventoryEvent(ctx context.Context, event *InventoryEvent) error
	PublishSyncEvent(ctx context.Context, event *SyncEvent) error
	
	// Health and status
	Close() error
	IsHealthy() bool
}

// Topics following PROJECT_RULES.md patterns
const (
	ProductEventsTopic   = "product-events"
	CategoryEventsTopic  = "category-events"
	PricingEventsTopic   = "pricing-events"
	InventoryEventsTopic = "inventory-events"
	SyncEventsTopic      = "sync-events"
)
