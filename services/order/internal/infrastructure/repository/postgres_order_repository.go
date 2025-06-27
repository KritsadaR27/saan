package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/saan/order-service/internal/domain"
)

// PostgresOrderRepository implements the OrderRepository interface
type PostgresOrderRepository struct {
	db *sqlx.DB
}

// NewPostgresOrderRepository creates a new PostgreSQL order repository
func NewPostgresOrderRepository(db *sqlx.DB) domain.OrderRepository {
	return &PostgresOrderRepository{db: db}
}

// Create creates a new order
func (r *PostgresOrderRepository) Create(ctx context.Context, order *domain.Order) error {
	query := `
		INSERT INTO orders (id, customer_id, status, total_amount, shipping_address, billing_address, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	
	_, err := r.db.ExecContext(ctx, query,
		order.ID, order.CustomerID, order.Status, order.TotalAmount,
		order.ShippingAddress, order.BillingAddress, order.Notes,
		order.CreatedAt, order.UpdatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}
	
	return nil
}

// GetByID retrieves an order by its ID
func (r *PostgresOrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	query := `
		SELECT id, customer_id, status, total_amount, shipping_address, billing_address, notes, created_at, updated_at
		FROM orders
		WHERE id = $1
	`
	
	var order domain.Order
	err := r.db.GetContext(ctx, &order, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order by ID: %w", err)
	}
	
	return &order, nil
}

// GetByCustomerID retrieves all orders for a customer
func (r *PostgresOrderRepository) GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*domain.Order, error) {
	query := `
		SELECT id, customer_id, status, total_amount, shipping_address, billing_address, notes, created_at, updated_at
		FROM orders
		WHERE customer_id = $1
		ORDER BY created_at DESC
	`
	
	var orders []*domain.Order
	err := r.db.SelectContext(ctx, &orders, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by customer ID: %w", err)
	}
	
	return orders, nil
}

// Update updates an existing order
func (r *PostgresOrderRepository) Update(ctx context.Context, order *domain.Order) error {
	query := `
		UPDATE orders
		SET customer_id = $2, status = $3, total_amount = $4, shipping_address = $5, 
		    billing_address = $6, notes = $7, updated_at = $8
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(ctx, query,
		order.ID, order.CustomerID, order.Status, order.TotalAmount,
		order.ShippingAddress, order.BillingAddress, order.Notes, order.UpdatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return domain.ErrOrderNotFound
	}
	
	return nil
}

// Delete deletes an order by ID
func (r *PostgresOrderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM orders WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return domain.ErrOrderNotFound
	}
	
	return nil
}

// List retrieves orders with pagination
func (r *PostgresOrderRepository) List(ctx context.Context, limit, offset int) ([]*domain.Order, error) {
	query := `
		SELECT id, customer_id, status, total_amount, shipping_address, billing_address, notes, created_at, updated_at
		FROM orders
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	
	var orders []*domain.Order
	err := r.db.SelectContext(ctx, &orders, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}
	
	return orders, nil
}

// GetByStatus retrieves orders by status
func (r *PostgresOrderRepository) GetByStatus(ctx context.Context, status domain.OrderStatus) ([]*domain.Order, error) {
	query := `
		SELECT id, customer_id, status, total_amount, shipping_address, billing_address, notes, created_at, updated_at
		FROM orders
		WHERE status = $1
		ORDER BY created_at DESC
	`
	
	var orders []*domain.Order
	err := r.db.SelectContext(ctx, &orders, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by status: %w", err)
	}
	
	return orders, nil
}

// PostgresOrderItemRepository implements the OrderItemRepository interface
type PostgresOrderItemRepository struct {
	db *sqlx.DB
}

// NewPostgresOrderItemRepository creates a new PostgreSQL order item repository
func NewPostgresOrderItemRepository(db *sqlx.DB) domain.OrderItemRepository {
	return &PostgresOrderItemRepository{db: db}
}

// Create creates a new order item
func (r *PostgresOrderItemRepository) Create(ctx context.Context, item *domain.OrderItem) error {
	query := `
		INSERT INTO order_items (id, order_id, product_id, quantity, unit_price, total_price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	
	_, err := r.db.ExecContext(ctx, query,
		item.ID, item.OrderID, item.ProductID, item.Quantity,
		item.UnitPrice, item.TotalPrice, item.CreatedAt, item.UpdatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to create order item: %w", err)
	}
	
	return nil
}

// GetByOrderID retrieves all items for an order
func (r *PostgresOrderItemRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.OrderItem, error) {
	query := `
		SELECT id, order_id, product_id, quantity, unit_price, total_price, created_at, updated_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY created_at
	`
	
	var items []*domain.OrderItem
	err := r.db.SelectContext(ctx, &items, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	
	return items, nil
}

// Update updates an existing order item
func (r *PostgresOrderItemRepository) Update(ctx context.Context, item *domain.OrderItem) error {
	query := `
		UPDATE order_items
		SET order_id = $2, product_id = $3, quantity = $4, unit_price = $5, total_price = $6, updated_at = $7
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(ctx, query,
		item.ID, item.OrderID, item.ProductID, item.Quantity,
		item.UnitPrice, item.TotalPrice, item.UpdatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update order item: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return domain.ErrOrderItemNotFound
	}
	
	return nil
}

// Delete deletes an order item by ID
func (r *PostgresOrderItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM order_items WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete order item: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return domain.ErrOrderItemNotFound
	}
	
	return nil
}