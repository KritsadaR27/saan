package events

import (
	"time"

	"github.com/google/uuid"
	"customer/internal/domain/entity"
)

// Event types
const (
	CustomerCreated         = "customer.created"
	CustomerUpdated         = "customer.updated"
	CustomerDeleted         = "customer.deleted"
	CustomerTierUpdated     = "customer.tier_updated"
	CustomerLoyverseSynced  = "customer.loyverse_synced"
	CustomerPointsUpdated   = "customer.points_updated"
	CustomerAddressAdded    = "customer.address_added"
	CustomerAddressUpdated  = "customer.address_updated"
	CustomerAddressDeleted  = "customer.address_deleted"
)

// Topic definitions following SAAN standards
const (
	CustomerEventsTopic = "customer-events"
	AnalyticsEventsTopic = "analytics-events"
	SyncEventsTopic = "sync-events"
)

// Base event structure
type BaseEvent struct {
	EventID     uuid.UUID `json:"event_id"`
	EventType   string    `json:"event_type"`
	AggregateID uuid.UUID `json:"aggregate_id"`
	Timestamp   time.Time `json:"timestamp"`
	Version     int       `json:"version"`
}

// CustomerEvent represents customer-related events
type CustomerEvent struct {
	BaseEvent
	CustomerID uuid.UUID        `json:"customer_id"`
	Customer   *entity.Customer `json:"customer,omitempty"`
	Changes    map[string]interface{} `json:"changes,omitempty"`
}

// CustomerTierEvent represents customer tier change events
type CustomerTierEvent struct {
	BaseEvent
	CustomerID uuid.UUID             `json:"customer_id"`
	OldTier    entity.CustomerTier    `json:"old_tier"`
	NewTier    entity.CustomerTier    `json:"new_tier"`
	Reason     string                 `json:"reason,omitempty"`
}

// CustomerPointsEvent represents customer points events
type CustomerPointsEvent struct {
	BaseEvent
	CustomerID       uuid.UUID `json:"customer_id"`
	PointsChange     int       `json:"points_change"`
	TotalPoints      int       `json:"total_points"`
	TransactionType  string    `json:"transaction_type"`
	ReferenceID      *uuid.UUID `json:"reference_id,omitempty"`
	Description      string    `json:"description,omitempty"`
}

// CustomerAddressEvent represents customer address events
type CustomerAddressEvent struct {
	BaseEvent
	CustomerID uuid.UUID               `json:"customer_id"`
	AddressID  uuid.UUID               `json:"address_id"`
	Address    *entity.CustomerAddress `json:"address,omitempty"`
	Changes    map[string]interface{}  `json:"changes,omitempty"`
}

// SyncEvent represents external system sync events
type SyncEvent struct {
	BaseEvent
	SourceSystem   string    `json:"source_system"`
	EntityType     string    `json:"entity_type"`
	EntityID       uuid.UUID `json:"entity_id"`
	ExternalID     string    `json:"external_id"`
	SyncStatus     string    `json:"sync_status"`
	ErrorMessage   string    `json:"error_message,omitempty"`
}

// NewCustomerEvent creates a new customer event
func NewCustomerEvent(eventType string, customerID uuid.UUID, customer *entity.Customer) *CustomerEvent {
	return &CustomerEvent{
		BaseEvent: BaseEvent{
			EventID:     uuid.New(),
			EventType:   eventType,
			AggregateID: customerID,
			Timestamp:   time.Now(),
			Version:     1,
		},
		CustomerID: customerID,
		Customer:   customer,
	}
}

// NewCustomerTierEvent creates a new customer tier event
func NewCustomerTierEvent(customerID uuid.UUID, oldTier, newTier entity.CustomerTier, reason string) *CustomerTierEvent {
	return &CustomerTierEvent{
		BaseEvent: BaseEvent{
			EventID:     uuid.New(),
			EventType:   CustomerTierUpdated,
			AggregateID: customerID,
			Timestamp:   time.Now(),
			Version:     1,
		},
		CustomerID: customerID,
		OldTier:    oldTier,
		NewTier:    newTier,
		Reason:     reason,
	}
}

// NewCustomerPointsEvent creates a new customer points event
func NewCustomerPointsEvent(customerID uuid.UUID, pointsChange, totalPoints int, transactionType, description string) *CustomerPointsEvent {
	return &CustomerPointsEvent{
		BaseEvent: BaseEvent{
			EventID:     uuid.New(),
			EventType:   CustomerPointsUpdated,
			AggregateID: customerID,
			Timestamp:   time.Now(),
			Version:     1,
		},
		CustomerID:      customerID,
		PointsChange:    pointsChange,
		TotalPoints:     totalPoints,
		TransactionType: transactionType,
		Description:     description,
	}
}

// NewCustomerAddressEvent creates a new customer address event
func NewCustomerAddressEvent(eventType string, customerID, addressID uuid.UUID, address *entity.CustomerAddress) *CustomerAddressEvent {
	return &CustomerAddressEvent{
		BaseEvent: BaseEvent{
			EventID:     uuid.New(),
			EventType:   eventType,
			AggregateID: customerID,
			Timestamp:   time.Now(),
			Version:     1,
		},
		CustomerID: customerID,
		AddressID:  addressID,
		Address:    address,
	}
}

// NewSyncEvent creates a new sync event
func NewSyncEvent(sourceSystem, entityType string, entityID uuid.UUID, externalID, syncStatus string) *SyncEvent {
	return &SyncEvent{
		BaseEvent: BaseEvent{
			EventID:     uuid.New(),
			EventType:   CustomerLoyverseSynced,
			AggregateID: entityID,
			Timestamp:   time.Now(),
			Version:     1,
		},
		SourceSystem: sourceSystem,
		EntityType:   entityType,
		EntityID:     entityID,
		ExternalID:   externalID,
		SyncStatus:   syncStatus,
	}
}
