package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"order/internal/domain"
	"order/internal/infrastructure/database"
)

// OrderRepository implements the OrderRepository interface using PostgreSQL
type OrderRepository struct {
	conn *database.Connection
}

// NewOrderRepository creates a new PostgreSQL order repository
func NewOrderRepository(conn *database.Connection) domain.OrderRepository {
	return &OrderRepository{conn: conn}
}

// Create creates a new order
func (r *OrderRepository) Create(ctx context.Context, order *domain.Order) error {
	query := `
		INSERT INTO orders (
			id, customer_id, code, status, source, paid_status, total_amount, 
			discount, shipping_fee, tax, tax_enabled, shipping_address, billing_address, 
			payment_method, promo_code, notes, confirmed_at, cancelled_at, cancelled_reason,
			created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21
		)
	`
	
	_, err := r.conn.DB.ExecContext(ctx, query,
		order.ID, order.CustomerID, order.Code, order.Status, order.Source, order.PaidStatus,
		order.TotalAmount, order.Discount, order.ShippingFee, order.Tax, order.TaxEnabled,
		order.ShippingAddress, order.BillingAddress, order.PaymentMethod, order.PromoCode,
		order.Notes, order.ConfirmedAt, order.CancelledAt, order.CancelledReason,
		order.CreatedAt, order.UpdatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}
	
	return nil
}

// GetByID retrieves an order by its ID
func (r *OrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	query := `
		SELECT id, customer_id, code, status, source, paid_status, total_amount, 
			   discount, shipping_fee, tax, tax_enabled, shipping_address, billing_address, 
			   payment_method, promo_code, notes, confirmed_at, cancelled_at, cancelled_reason,
			   created_at, updated_at
		FROM orders
		WHERE id = $1
	`
	
	order := &domain.Order{}
	err := r.conn.DB.GetContext(ctx, order, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	
	return order, nil
}

// GetByCustomerID retrieves all orders for a customer
func (r *OrderRepository) GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*domain.Order, error) {
	query := `
		SELECT id, customer_id, code, status, source, paid_status, total_amount, 
			   discount, shipping_fee, tax, tax_enabled, shipping_address, billing_address, 
			   payment_method, promo_code, notes, confirmed_at, cancelled_at, cancelled_reason,
			   created_at, updated_at
		FROM orders
		WHERE customer_id = $1
		ORDER BY created_at DESC
	`
	
	var orders []*domain.Order
	err := r.conn.DB.SelectContext(ctx, &orders, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by customer ID: %w", err)
	}
	
	return orders, nil
}

// Update updates an existing order
func (r *OrderRepository) Update(ctx context.Context, order *domain.Order) error {
	query := `
		UPDATE orders SET
			customer_id = $2, code = $3, status = $4, source = $5, paid_status = $6,
			total_amount = $7, discount = $8, shipping_fee = $9, tax = $10, tax_enabled = $11,
			shipping_address = $12, billing_address = $13, payment_method = $14, promo_code = $15,
			notes = $16, confirmed_at = $17, cancelled_at = $18, cancelled_reason = $19, updated_at = $20
		WHERE id = $1
	`
	
	result, err := r.conn.DB.ExecContext(ctx, query,
		order.ID, order.CustomerID, order.Code, order.Status, order.Source, order.PaidStatus,
		order.TotalAmount, order.Discount, order.ShippingFee, order.Tax, order.TaxEnabled,
		order.ShippingAddress, order.BillingAddress, order.PaymentMethod, order.PromoCode,
		order.Notes, order.ConfirmedAt, order.CancelledAt, order.CancelledReason, order.UpdatedAt,
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
func (r *OrderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM orders WHERE id = $1`
	
	result, err := r.conn.DB.ExecContext(ctx, query, id)
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
func (r *OrderRepository) List(ctx context.Context, limit, offset int) ([]*domain.Order, error) {
	query := `
		SELECT id, customer_id, code, status, source, paid_status, total_amount, 
			   discount, shipping_fee, tax, tax_enabled, shipping_address, billing_address, 
			   payment_method, promo_code, notes, confirmed_at, cancelled_at, cancelled_reason,
			   created_at, updated_at
		FROM orders
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	
	var orders []*domain.Order
	err := r.conn.DB.SelectContext(ctx, &orders, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}
	
	return orders, nil
}

// GetByStatus retrieves orders by status
func (r *OrderRepository) GetByStatus(ctx context.Context, status domain.OrderStatus) ([]*domain.Order, error) {
	query := `
		SELECT id, customer_id, code, status, source, paid_status, total_amount, 
			   discount, shipping_fee, tax, tax_enabled, shipping_address, billing_address, 
			   payment_method, promo_code, notes, confirmed_at, cancelled_at, cancelled_reason,
			   created_at, updated_at
		FROM orders
		WHERE status = $1
		ORDER BY created_at DESC
	`
	
	var orders []*domain.Order
	err := r.conn.DB.SelectContext(ctx, &orders, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by status: %w", err)
	}
	
	return orders, nil
}

// GetOrdersByDateRange retrieves orders within a date range
func (r *OrderRepository) GetOrdersByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*domain.Order, error) {
	query := `
		SELECT id, customer_id, code, status, source, paid_status, total_amount, 
			   discount, shipping_fee, tax, tax_enabled, shipping_address, billing_address, 
			   payment_method, promo_code, notes, confirmed_at, cancelled_at, cancelled_reason,
			   created_at, updated_at
		FROM orders
		WHERE created_at >= $1 AND created_at <= $2
		ORDER BY created_at DESC
	`
	
	var orders []*domain.Order
	err := r.conn.DB.SelectContext(ctx, &orders, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by date range: %w", err)
	}
	
	return orders, nil
}

// GetOrdersByCustomer is an alias for GetByCustomerID
func (r *OrderRepository) GetOrdersByCustomer(ctx context.Context, customerID uuid.UUID) ([]*domain.Order, error) {
	return r.GetByCustomerID(ctx, customerID)
}

// OrderItemRepository implements the OrderItemRepository interface using PostgreSQL
type OrderItemRepository struct {
	conn *database.Connection
}

// NewOrderItemRepository creates a new PostgreSQL order item repository
func NewOrderItemRepository(conn *database.Connection) domain.OrderItemRepository {
	return &OrderItemRepository{conn: conn}
}

// Create creates a new order item
func (r *OrderItemRepository) Create(ctx context.Context, item *domain.OrderItem) error {
	query := `
		INSERT INTO order_items (id, order_id, product_id, quantity, unit_price, total_price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	
	_, err := r.conn.DB.ExecContext(ctx, query,
		item.ID, item.OrderID, item.ProductID, item.Quantity, item.UnitPrice, 
		item.TotalPrice, item.CreatedAt, item.UpdatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to create order item: %w", err)
	}
	
	return nil
}

// GetByOrderID retrieves all items for an order
func (r *OrderItemRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.OrderItem, error) {
	query := `
		SELECT id, order_id, product_id, quantity, unit_price, total_price, created_at, updated_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY created_at ASC
	`
	
	var items []*domain.OrderItem
	err := r.conn.DB.SelectContext(ctx, &items, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	
	return items, nil
}

// Update updates an existing order item
func (r *OrderItemRepository) Update(ctx context.Context, item *domain.OrderItem) error {
	query := `
		UPDATE order_items SET
			product_id = $2, quantity = $3, unit_price = $4, total_price = $5, updated_at = $6
		WHERE id = $1
	`
	
	result, err := r.conn.DB.ExecContext(ctx, query,
		item.ID, item.ProductID, item.Quantity, item.UnitPrice, item.TotalPrice, item.UpdatedAt,
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
func (r *OrderItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM order_items WHERE id = $1`
	
	result, err := r.conn.DB.ExecContext(ctx, query, id)
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

// DeleteByOrderID deletes all items for an order
func (r *OrderItemRepository) DeleteByOrderID(ctx context.Context, orderID uuid.UUID) error {
	query := `DELETE FROM order_items WHERE order_id = $1`
	
	_, err := r.conn.DB.ExecContext(ctx, query, orderID)
	if err != nil {
		return fmt.Errorf("failed to delete order items: %w", err)
	}
	
	return nil
}

// GetAllOrderItems retrieves all order items (for statistics)
func (r *OrderItemRepository) GetAllOrderItems(ctx context.Context) ([]*domain.OrderItem, error) {
	query := `
		SELECT id, order_id, product_id, quantity, unit_price, total_price, created_at, updated_at
		FROM order_items
		ORDER BY created_at DESC
	`
	
	var items []*domain.OrderItem
	err := r.conn.DB.SelectContext(ctx, &items, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all order items: %w", err)
	}
	
	return items, nil
}
