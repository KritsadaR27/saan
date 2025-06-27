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
