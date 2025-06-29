package domain

import (
	"context"
	
	"github.com/google/uuid"
)

// OrderRepository defines the interface for order data operations
type OrderRepository interface {
	// Create creates a new order
	Create(ctx context.Context, order *Order) error
	
	// GetByID retrieves an order by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*Order, error)
	
	// GetByCustomerID retrieves all orders for a customer
	GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*Order, error)
	
	// Update updates an existing order
	Update(ctx context.Context, order *Order) error
	
	// Delete deletes an order by ID
	Delete(ctx context.Context, id uuid.UUID) error
	
	// List retrieves orders with pagination
	List(ctx context.Context, limit, offset int) ([]*Order, error)
	
	// GetByStatus retrieves orders by status
	GetByStatus(ctx context.Context, status OrderStatus) ([]*Order, error)
}

// OrderItemRepository defines the interface for order item data operations
type OrderItemRepository interface {
	// Create creates a new order item
	Create(ctx context.Context, item *OrderItem) error
	
	// GetByOrderID retrieves all items for an order
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*OrderItem, error)
	
	// Update updates an existing order item
	Update(ctx context.Context, item *OrderItem) error
	
	// Delete deletes an order item by ID
	Delete(ctx context.Context, id uuid.UUID) error
}

// OrderAuditRepository defines the interface for order audit log operations
type OrderAuditRepository interface {
	// Create creates a new audit log entry
	Create(ctx context.Context, auditLog *OrderAuditLog) error
	
	// GetByOrderID retrieves all audit logs for an order
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*OrderAuditLog, error)
	
	// GetByUserID retrieves audit logs by user ID
	GetByUserID(ctx context.Context, userID string) ([]*OrderAuditLog, error)
	
	// GetByAction retrieves audit logs by action type
	GetByAction(ctx context.Context, action AuditAction) ([]*OrderAuditLog, error)
	
	// List retrieves audit logs with pagination
	List(ctx context.Context, limit, offset int) ([]*OrderAuditLog, error)
}

// OrderEventRepository defines the interface for order events outbox operations
type OrderEventRepository interface {
	// Create creates a new event in the outbox
	Create(ctx context.Context, event *OrderEventOutbox) error
	
	// GetPendingEvents retrieves all pending events for processing
	GetPendingEvents(ctx context.Context, limit int) ([]*OrderEventOutbox, error)
	
	// GetFailedEvents retrieves failed events that can be retried
	GetFailedEvents(ctx context.Context, maxRetries int, limit int) ([]*OrderEventOutbox, error)
	
	// UpdateStatus updates the status of an event
	UpdateStatus(ctx context.Context, eventID uuid.UUID, status EventStatus) error
	
	// MarkAsSent marks an event as successfully sent
	MarkAsSent(ctx context.Context, eventID uuid.UUID) error
	
	// MarkAsFailed marks an event as failed and increments retry count
	MarkAsFailed(ctx context.Context, eventID uuid.UUID) error
	
	// GetByOrderID retrieves all events for an order
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*OrderEventOutbox, error)
	
	// Delete removes old processed events (for cleanup)
	Delete(ctx context.Context, eventID uuid.UUID) error
}

// EventRepository is an alias for OrderEventRepository to match the task specification
type EventRepository = OrderEventRepository

// AuditRepository is an alias for OrderAuditRepository to match the task specification
type AuditRepository = OrderAuditRepository
