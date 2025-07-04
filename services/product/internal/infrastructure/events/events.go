package events

import (
	"time"

	"github.com/google/uuid"
)

// Event types following PROJECT_RULES.md patterns
const (
	// Product Events
	ProductCreatedEvent = "product.created"
	ProductUpdatedEvent = "product.updated"
	ProductDeletedEvent = "product.deleted"
	ProductActivatedEvent = "product.activated"
	ProductDeactivatedEvent = "product.deactivated"
	
	// Category Events
	CategoryCreatedEvent = "category.created"
	CategoryUpdatedEvent = "category.updated"
	CategoryDeletedEvent = "category.deleted"
	
	// Pricing Events
	PriceUpdatedEvent = "price.updated"
	PricingTierCreatedEvent = "pricing.tier.created"
	VIPPricingUpdatedEvent = "vip.pricing.updated"
	
	// Inventory Events
	StockUpdatedEvent = "stock.updated"
	AvailabilityChangedEvent = "availability.changed"
	
	// Sync Events
	LoyverseSyncCompletedEvent = "loyverse.sync.completed"
	SyncFailedEvent = "sync.failed"
)

// BaseEvent represents the base structure for all events
type BaseEvent struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Source    string    `json:"source"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// ProductEvent represents product-related events
type ProductEvent struct {
	BaseEvent
	ProductID uuid.UUID   `json:"product_id"`
	SKU       string      `json:"sku"`
	Name      string      `json:"name"`
	Action    string      `json:"action"`
	Changes   interface{} `json:"changes,omitempty"`
}

// CategoryEvent represents category-related events
type CategoryEvent struct {
	BaseEvent
	CategoryID uuid.UUID   `json:"category_id"`
	Name       string      `json:"name"`
	ParentID   *uuid.UUID  `json:"parent_id,omitempty"`
	Action     string      `json:"action"`
	Changes    interface{} `json:"changes,omitempty"`
}

// PricingEvent represents pricing-related events
type PricingEvent struct {
	BaseEvent
	ProductID    uuid.UUID `json:"product_id"`
	CustomerType string    `json:"customer_type,omitempty"`
	OldPrice     float64   `json:"old_price"`
	NewPrice     float64   `json:"new_price"`
	Currency     string    `json:"currency"`
	Action       string    `json:"action"`
}

// InventoryEvent represents inventory-related events
type InventoryEvent struct {
	BaseEvent
	ProductID     uuid.UUID `json:"product_id"`
	LocationID    *uuid.UUID `json:"location_id,omitempty"`
	OldQuantity   int       `json:"old_quantity"`
	NewQuantity   int       `json:"new_quantity"`
	ChangeReason  string    `json:"change_reason"`
	Action        string    `json:"action"`
}

// SyncEvent represents sync-related events
type SyncEvent struct {
	BaseEvent
	SyncType      string    `json:"sync_type"`
	SourceSystem  string    `json:"source_system"`
	RecordsCount  int       `json:"records_count"`
	SuccessCount  int       `json:"success_count"`
	FailureCount  int       `json:"failure_count"`
	Status        string    `json:"status"`
	ErrorMessage  string    `json:"error_message,omitempty"`
}

// NewBaseEvent creates a new base event
func NewBaseEvent(eventType, source string) BaseEvent {
	return BaseEvent{
		ID:        uuid.New().String(),
		Type:      eventType,
		Source:    source,
		Timestamp: time.Now(),
		Version:   "1.0",
	}
}

// NewProductEvent creates a new product event
func NewProductEvent(eventType string, productID uuid.UUID, sku, name, action string, changes interface{}) *ProductEvent {
	return &ProductEvent{
		BaseEvent: NewBaseEvent(eventType, "product-service"),
		ProductID: productID,
		SKU:       sku,
		Name:      name,
		Action:    action,
		Changes:   changes,
	}
}

// NewCategoryEvent creates a new category event
func NewCategoryEvent(eventType string, categoryID uuid.UUID, name string, parentID *uuid.UUID, action string, changes interface{}) *CategoryEvent {
	return &CategoryEvent{
		BaseEvent:  NewBaseEvent(eventType, "product-service"),
		CategoryID: categoryID,
		Name:       name,
		ParentID:   parentID,
		Action:     action,
		Changes:    changes,
	}
}

// NewPricingEvent creates a new pricing event
func NewPricingEvent(eventType string, productID uuid.UUID, customerType string, oldPrice, newPrice float64, action string) *PricingEvent {
	return &PricingEvent{
		BaseEvent:    NewBaseEvent(eventType, "product-service"),
		ProductID:    productID,
		CustomerType: customerType,
		OldPrice:     oldPrice,
		NewPrice:     newPrice,
		Currency:     "THB",
		Action:       action,
	}
}

// NewInventoryEvent creates a new inventory event
func NewInventoryEvent(eventType string, productID uuid.UUID, locationID *uuid.UUID, oldQty, newQty int, reason, action string) *InventoryEvent {
	return &InventoryEvent{
		BaseEvent:    NewBaseEvent(eventType, "product-service"),
		ProductID:    productID,
		LocationID:   locationID,
		OldQuantity:  oldQty,
		NewQuantity:  newQty,
		ChangeReason: reason,
		Action:       action,
	}
}

// NewSyncEvent creates a new sync event
func NewSyncEvent(eventType, syncType, sourceSystem string, recordsCount, successCount, failureCount int, status, errorMessage string) *SyncEvent {
	return &SyncEvent{
		BaseEvent:    NewBaseEvent(eventType, "product-service"),
		SyncType:     syncType,
		SourceSystem: sourceSystem,
		RecordsCount: recordsCount,
		SuccessCount: successCount,
		FailureCount: failureCount,
		Status:       status,
		ErrorMessage: errorMessage,
	}
}
