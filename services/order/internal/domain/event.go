package domain
package domain

import (
	"time"

	"github.com/google/uuid"
)

// EventType represents the type of order event
type EventType string

const (
	EventOrderCreated   EventType = "OrderCreated"
	EventOrderConfirmed EventType = "OrderConfirmed"
	EventOrderReady     EventType = "OrderReady"
	EventOrderDelivered EventType = "OrderDelivered"
	EventOrderPaid      EventType = "OrderPaid"
	EventOrderCancelled EventType = "OrderCancelled"
)

// EventStatus represents the status of an event
type EventStatus string

const (
	EventStatusPending EventStatus = "pending"
	EventStatusSent    EventStatus = "sent"
	EventStatusFailed  EventStatus = "failed"
)

// OrderEvent represents an order event in the outbox pattern
type OrderEvent struct {
	ID         uuid.UUID              `json:"id" db:"id"`
	OrderID    uuid.UUID              `json:"order_id" db:"order_id"`
	EventType  EventType              `json:"event_type" db:"event_type"`
	Payload    map[string]interface{} `json:"payload" db:"payload"`
	Status     EventStatus            `json:"status" db:"status"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
	SentAt     *time.Time             `json:"sent_at" db:"sent_at"`
	RetryCount int                    `json:"retry_count" db:"retry_count"`
}

// NewOrderEvent creates a new order event
func NewOrderEvent(orderID uuid.UUID, eventType EventType, payload map[string]interface{}) *OrderEvent {
	return &OrderEvent{
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
func (e *OrderEvent) MarkAsSent() {
	now := time.Now()
	e.Status = EventStatusSent
	e.SentAt = &now
}

// MarkAsFailed marks the event as failed and increments retry count
func (e *OrderEvent) MarkAsFailed() {
	e.Status = EventStatusFailed
	e.RetryCount++
}

// ShouldRetry checks if the event should be retried based on retry count
func (e *OrderEvent) ShouldRetry(maxRetries int) bool {
	return e.Status == EventStatusFailed && e.RetryCount < maxRetries
}

// EventRepository defines the interface for event operations
type EventRepository interface {
	// Create creates a new event
	Create(ctx context.Context, event *OrderEvent) error
	
	// GetPending retrieves pending events for processing
	GetPending(ctx context.Context, limit int) ([]*OrderEvent, error)
	
	// UpdateStatus updates the status of an event
	UpdateStatus(ctx context.Context, id uuid.UUID, status EventStatus) error
	
	// IncrementRetry increments the retry count for a failed event
	IncrementRetry(ctx context.Context, id uuid.UUID) error
}
