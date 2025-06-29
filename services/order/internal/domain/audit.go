package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AuditAction represents the type of action performed on an order
type AuditAction string

const (
	AuditActionCreate        AuditAction = "CREATE"
	AuditActionUpdate        AuditAction = "UPDATE"
	AuditActionStatusChange  AuditAction = "CHANGE_STATUS"
	AuditActionOverrideStock AuditAction = "OVERRIDE_STOCK"
	AuditActionCancel        AuditAction = "CANCEL"
)

// OrderAuditLog represents an audit log entry for order changes
type OrderAuditLog struct {
	ID        uuid.UUID              `json:"id" db:"id"`
	OrderID   uuid.UUID              `json:"order_id" db:"order_id"`
	UserID    *string                `json:"user_id,omitempty" db:"user_id"`
	Action    AuditAction            `json:"action" db:"action"`
	Details   map[string]interface{} `json:"details,omitempty" db:"details"`
	Timestamp time.Time              `json:"timestamp" db:"timestamp"`
}

// NewAuditLog creates a new audit log entry
func NewAuditLog(orderID uuid.UUID, userID *string, action AuditAction, details map[string]interface{}) *OrderAuditLog {
	return &OrderAuditLog{
		ID:        uuid.New(),
		OrderID:   orderID,
		UserID:    userID,
		Action:    action,
		Details:   details,
		Timestamp: time.Now(),
	}
}

// AuditLog is an alias for OrderAuditLog to match the task specification
type AuditLog = OrderAuditLog

// EventStatus represents the status of an event in the outbox
type EventStatus string

const (
	EventStatusPending   EventStatus = "pending"
	EventStatusSent      EventStatus = "sent"
	EventStatusFailed    EventStatus = "failed"
	EventStatusCancelled EventStatus = "cancelled"
)

// EventType represents the type of event being published
type EventType string

const (
	EventTypeOrderCreated      EventType = "OrderCreated"
	EventTypeOrderConfirmed    EventType = "OrderConfirmed"
	EventTypeOrderReady        EventType = "OrderReady"
	EventTypeOrderDelivered    EventType = "OrderDelivered"
	EventTypeOrderPaid         EventType = "OrderPaid"
	EventTypeOrderCancelled    EventType = "OrderCancelled"
	EventTypeOrderUpdated      EventType = "order_updated"
	EventTypeOrderShipped      EventType = "order_shipped"
	EventTypeOrderRefunded     EventType = "order_refunded"
	EventTypePaymentUpdated    EventType = "payment_updated"
	EventTypeInventoryReserved EventType = "inventory_reserved"
	EventTypeInventoryReleased EventType = "inventory_released"
)

// OrderEventOutbox represents an event in the outbox pattern for reliable event publishing
type OrderEventOutbox struct {
	ID         uuid.UUID              `json:"id" db:"id"`
	OrderID    uuid.UUID              `json:"order_id" db:"order_id"`
	EventType  EventType              `json:"event_type" db:"event_type"`
	Payload    map[string]interface{} `json:"payload" db:"payload"`
	Status     EventStatus            `json:"status" db:"status"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
	SentAt     *time.Time             `json:"sent_at,omitempty" db:"sent_at"`
	RetryCount int                    `json:"retry_count" db:"retry_count"`
}

// NewOrderEvent creates a new order event for the outbox
func NewOrderEvent(orderID uuid.UUID, eventType EventType, payload map[string]interface{}) *OrderEventOutbox {
	return &OrderEventOutbox{
		ID:         uuid.New(),
		OrderID:    orderID,
		EventType:  eventType,
		Payload:    payload,
		Status:     EventStatusPending,
		CreatedAt:  time.Now(),
		RetryCount: 0,
	}
}

// MarkAsSent marks the event as successfully sent
func (e *OrderEventOutbox) MarkAsSent() {
	now := time.Now()
	e.Status = EventStatusSent
	e.SentAt = &now
}

// MarkAsFailed marks the event as failed and increments retry count
func (e *OrderEventOutbox) MarkAsFailed() {
	e.Status = EventStatusFailed
	e.RetryCount++
}

// MarkAsCancelled marks the event as cancelled
func (e *OrderEventOutbox) MarkAsCancelled() {
	e.Status = EventStatusCancelled
}

// ShouldRetry determines if the event should be retried based on retry count
func (e *OrderEventOutbox) ShouldRetry(maxRetries int) bool {
	return e.Status == EventStatusFailed && e.RetryCount < maxRetries
}

// GetPayloadAsJSON returns the payload as JSON bytes
func (e *OrderEventOutbox) GetPayloadAsJSON() ([]byte, error) {
	return json.Marshal(e.Payload)
}

// SetPayloadFromJSON sets the payload from JSON bytes
func (e *OrderEventOutbox) SetPayloadFromJSON(data []byte) error {
	return json.Unmarshal(data, &e.Payload)
}

// OrderEvent is an alias for OrderEventOutbox to match the task specification
type OrderEvent = OrderEventOutbox

// EventPublisher defines the interface for publishing events
type EventPublisher interface {
	PublishEvent(ctx context.Context, event *OrderEvent) error
}
