package events

import (
	"crypto/rand"
	"fmt"
	"time"
)

// Event types for inventory domain
const (
	StockUpdated      = "inventory.stock_updated"
	StockLevelLow     = "inventory.stock_level_low"
	ProductCreated    = "inventory.product_created"
	ProductUpdated    = "inventory.product_updated"
	ProductDeleted    = "inventory.product_deleted"
	StockMovement     = "inventory.stock_movement"
	StockAdjustment   = "inventory.stock_adjustment"
	LoyverseSync      = "inventory.loyverse_sync"
)

// Topic definitions following SAAN standards
const (
	InventoryEventsTopic = "inventory-events"
	LoyverseEventsTopic  = "loyverse-events"
	AnalyticsEventsTopic = "analytics-events"
)

// Base event structure
type BaseEvent struct {
	EventID     string    `json:"event_id"`
	EventType   string    `json:"event_type"`
	AggregateID string    `json:"aggregate_id"`
	Timestamp   time.Time `json:"timestamp"`
	Version     int       `json:"version"`
}

// StockEvent represents stock-related events
type StockEvent struct {
	BaseEvent
	ProductID       string `json:"product_id"`
	PreviousLevel   int    `json:"previous_level"`
	NewLevel        int    `json:"new_level"`
	MovementType    string `json:"movement_type"` // sale, adjustment, restock, etc.
	Quantity        int    `json:"quantity"`
	Reason          string `json:"reason,omitempty"`
	LocationID      string `json:"location_id,omitempty"`
	TransactionID   string `json:"transaction_id,omitempty"`
}

// ProductEvent represents product-related events
type ProductEvent struct {
	BaseEvent
	ProductID   string                 `json:"product_id"`
	LoyverseID  string                 `json:"loyverse_id,omitempty"`
	ProductData map[string]interface{} `json:"product_data,omitempty"`
	Changes     map[string]interface{} `json:"changes,omitempty"`
}

// LoyverseSyncEvent represents Loyverse synchronization events
type LoyverseSyncEvent struct {
	BaseEvent
	EntityType     string `json:"entity_type"`
	EntityID       string `json:"entity_id"`
	LoyverseID     string `json:"loyverse_id"`
	SyncStatus     string `json:"sync_status"`
	SyncDirection  string `json:"sync_direction"` // from_loyverse, to_loyverse
	ErrorMessage   string `json:"error_message,omitempty"`
}

// AlertEvent represents inventory alert events
type AlertEvent struct {
	BaseEvent
	AlertType    string `json:"alert_type"`
	ProductID    string `json:"product_id"`
	CurrentLevel int    `json:"current_level"`
	Threshold    int    `json:"threshold"`
	Severity     string `json:"severity"` // low, medium, high, critical
	Message      string `json:"message"`
}

// generateID generates a unique ID for events
func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("%x-%x-%x-%x-%x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:])
}

// NewStockEvent creates a new stock event
func NewStockEvent(eventType string, productID string, previousLevel, newLevel int, movementType, reason string) *StockEvent {
	return &StockEvent{
		BaseEvent: BaseEvent{
			EventID:     generateID(),
			EventType:   eventType,
			AggregateID: productID,
			Timestamp:   time.Now(),
			Version:     1,
		},
		ProductID:     productID,
		PreviousLevel: previousLevel,
		NewLevel:      newLevel,
		MovementType:  movementType,
		Quantity:      newLevel - previousLevel,
		Reason:        reason,
	}
}

// NewProductEvent creates a new product event
func NewProductEvent(eventType string, productID string, loyverseID string) *ProductEvent {
	return &ProductEvent{
		BaseEvent: BaseEvent{
			EventID:     generateID(),
			EventType:   eventType,
			AggregateID: productID,
			Timestamp:   time.Now(),
			Version:     1,
		},
		ProductID:  productID,
		LoyverseID: loyverseID,
	}
}

// NewLoyverseSyncEvent creates a new Loyverse sync event
func NewLoyverseSyncEvent(entityType string, entityID string, loyverseID, syncStatus, syncDirection string) *LoyverseSyncEvent {
	return &LoyverseSyncEvent{
		BaseEvent: BaseEvent{
			EventID:     generateID(),
			EventType:   LoyverseSync,
			AggregateID: entityID,
			Timestamp:   time.Now(),
			Version:     1,
		},
		EntityType:    entityType,
		EntityID:      entityID,
		LoyverseID:    loyverseID,
		SyncStatus:    syncStatus,
		SyncDirection: syncDirection,
	}
}

// NewAlertEvent creates a new alert event
func NewAlertEvent(alertType string, productID string, currentLevel, threshold int, severity, message string) *AlertEvent {
	return &AlertEvent{
		BaseEvent: BaseEvent{
			EventID:     generateID(),
			EventType:   StockLevelLow,
			AggregateID: productID,
			Timestamp:   time.Now(),
			Version:     1,
		},
		AlertType:    alertType,
		ProductID:    productID,
		CurrentLevel: currentLevel,
		Threshold:    threshold,
		Severity:     severity,
		Message:      message,
	}
}
