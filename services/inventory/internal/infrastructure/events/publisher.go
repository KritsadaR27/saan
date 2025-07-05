package events

import (
	"context"
)

// Publisher defines the interface for event publishing
type Publisher interface {
	// Stock events
	PublishStockUpdated(ctx context.Context, productID string, previousLevel, newLevel int, movementType, reason string) error
	PublishStockLevelLow(ctx context.Context, productID string, currentLevel, threshold int, severity, message string) error
	
	// Product events
	PublishProductCreated(ctx context.Context, productID, loyverseID string) error
	PublishProductUpdated(ctx context.Context, productID, loyverseID string) error
	PublishProductDeleted(ctx context.Context, productID string) error
	
	// Sync events
	PublishLoyverseSync(ctx context.Context, entityType, entityID, loyverseID, syncStatus, syncDirection string) error
	
	// Lifecycle
	Close() error
}

// EventHandler defines the interface for event handling
type EventHandler func(eventType string, data []byte) error
