package application
package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/saan/order-service/internal/domain"
	"github.com/saan/order-service/internal/application/dto"
	"github.com/saan/order-service/pkg/logger"
)

// OrderService provides business logic for order operations
type OrderService struct {
	orderRepo      domain.OrderRepository
	orderItemRepo  domain.OrderItemRepository
	auditRepo      domain.AuditRepository
	eventRepo      domain.EventRepository
	eventPublisher domain.EventPublisher
	logger         logger.Logger
}

// NewOrderService creates a new order service instance
func NewOrderService(
	orderRepo domain.OrderRepository,
	orderItemRepo domain.OrderItemRepository,
	auditRepo domain.AuditRepository,
	eventRepo domain.EventRepository,
	eventPublisher domain.EventPublisher,
	logger logger.Logger,
) *OrderService {
	return &OrderService{
		orderRepo:      orderRepo,
		orderItemRepo:  orderItemRepo,
		auditRepo:      auditRepo,
		eventRepo:      eventRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// CreateOrder creates a new order with items
func (s *OrderService) CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	// Create new order
	order := domain.NewOrder(req.CustomerID, req.ShippingAddress, req.BillingAddress, req.Notes)
	
	// Add items to the order
	for _, itemReq := range req.Items {
		order.AddItem(itemReq.ProductID, itemReq.Quantity, itemReq.UnitPrice)
	}
	
	// Save order to repository
	if err := s.orderRepo.Create(ctx, order); err != nil {
		s.logger.Error("Failed to create order", "error", err, "customer_id", req.CustomerID)
		return nil, err
	}
	
	// Save order items
	for _, item := range order.Items {
		if err := s.orderItemRepo.Create(ctx, &item); err != nil {
			s.logger.Error("Failed to create order item", "error", err, "order_id", order.ID)
			return nil, err
		}
	}
	
	// Create audit log entry
	auditDetails := map[string]interface{}{
		"customer_id":       order.CustomerID,
		"total_amount":      order.TotalAmount,
		"shipping_address":  order.ShippingAddress,
		"billing_address":   order.BillingAddress,
		"items_count":       len(order.Items),
	}
	auditLog := domain.NewAuditLog(order.ID, nil, domain.AuditActionCreate, auditDetails)
	if err := s.auditRepo.Create(ctx, auditLog); err != nil {
		s.logger.Warn("Failed to create audit log", "error", err, "order_id", order.ID)
		// Continue - audit failure shouldn't break order creation
	}
	
	// Create and store event for outbox pattern
	eventPayload := map[string]interface{}{
		"order_id":         order.ID,
		"customer_id":      order.CustomerID,
		"total_amount":     order.TotalAmount,
		"status":          order.Status,
		"shipping_address": order.ShippingAddress,
		"created_at":      order.CreatedAt,
	}
	event := domain.NewOrderEvent(order.ID, domain.EventTypeOrderCreated, eventPayload)
	if err := s.eventRepo.Create(ctx, event); err != nil {
		s.logger.Error("Failed to create order event", "error", err, "order_id", order.ID)
		// Continue - event failure shouldn't break order creation
	}
	
	s.logger.Info("Order created successfully", "order_id", order.ID, "customer_id", req.CustomerID)
	return dto.ToOrderResponse(order), nil
}

// GetOrderByID retrieves an order by its ID
func (s *OrderService) GetOrderByID(ctx context.Context, id uuid.UUID) (*dto.OrderResponse, error) {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// Load order items
	items, err := s.orderItemRepo.GetByOrderID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	order.Items = make([]domain.OrderItem, len(items))
	for i, item := range items {
		order.Items[i] = *item
	}
	
	return dto.ToOrderResponse(order), nil
}

// GetOrdersByCustomerID retrieves all orders for a customer
func (s *OrderService) GetOrdersByCustomerID(ctx context.Context, customerID uuid.UUID) ([]dto.OrderResponse, error) {
	orders, err := s.orderRepo.GetByCustomerID(ctx, customerID)
	if err != nil {
		return nil, err
	}
	
	// Load items for each order
	for _, order := range orders {
		items, err := s.orderItemRepo.GetByOrderID(ctx, order.ID)
		if err != nil {
			return nil, err
		}
		
		order.Items = make([]domain.OrderItem, len(items))
		for i, item := range items {
			order.Items[i] = *item
		}
	}
	
	return dto.ToOrderResponseList(orders), nil
}

// UpdateOrder updates an existing order
func (s *OrderService) UpdateOrder(ctx context.Context, id uuid.UUID, req *dto.UpdateOrderRequest) (*dto.OrderResponse, error) {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// Check if order can be modified
	if order.Status != domain.OrderStatusPending && order.Status != domain.OrderStatusConfirmed {
		return nil, domain.ErrOrderCannotBeModified
	}
	
	// Track changes for audit
	changes := make(map[string]interface{})
	
	// Update fields if provided
	if req.ShippingAddress != nil && *req.ShippingAddress != order.ShippingAddress {
		changes["shipping_address"] = map[string]string{
			"old": order.ShippingAddress,
			"new": *req.ShippingAddress,
		}
		order.ShippingAddress = *req.ShippingAddress
	}
	if req.BillingAddress != nil && *req.BillingAddress != order.BillingAddress {
		changes["billing_address"] = map[string]string{
			"old": order.BillingAddress,
			"new": *req.BillingAddress,
		}
		order.BillingAddress = *req.BillingAddress
	}
	if req.Notes != nil && *req.Notes != order.Notes {
		changes["notes"] = map[string]string{
			"old": order.Notes,
			"new": *req.Notes,
		}
		order.Notes = *req.Notes
	}
	
	// Save updated order
	if err := s.orderRepo.Update(ctx, order); err != nil {
		s.logger.Error("Failed to update order", "error", err, "order_id", id)
		return nil, err
	}
	
	// Create audit log entry if there were changes
	if len(changes) > 0 {
		auditLog := domain.NewAuditLog(order.ID, nil, domain.AuditActionUpdate, changes)
		if err := s.auditRepo.Create(ctx, auditLog); err != nil {
			s.logger.Warn("Failed to create audit log", "error", err, "order_id", order.ID)
		}
		
		// Create and store event for outbox pattern
		eventPayload := map[string]interface{}{
			"order_id":   order.ID,
			"changes":    changes,
			"updated_at": order.UpdatedAt,
		}
		event := domain.NewOrderEvent(order.ID, domain.EventTypeOrderUpdated, eventPayload)
		if err := s.eventRepo.Create(ctx, event); err != nil {
			s.logger.Error("Failed to create order update event", "error", err, "order_id", order.ID)
		}
		
		s.logger.Info("Order updated successfully", "order_id", order.ID, "changes", len(changes))
	}
	
	return dto.ToOrderResponse(order), nil
}

// UpdateOrderStatus updates the status of an order
func (s *OrderService) UpdateOrderStatus(ctx context.Context, id uuid.UUID, req *dto.UpdateOrderStatusRequest) (*dto.OrderResponse, error) {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	oldStatus := order.Status
	
	// Update status with validation
	if err := order.UpdateStatus(req.Status); err != nil {
		return nil, err
	}
	
	// Save updated order
	if err := s.orderRepo.Update(ctx, order); err != nil {
		s.logger.Error("Failed to update order status", "error", err, "order_id", id)
		return nil, err
	}
	
	// Create audit log entry
	auditDetails := map[string]interface{}{
		"old_status": oldStatus,
		"new_status": order.Status,
		"updated_by": "system", // TODO: Get from context
	}
	auditLog := domain.NewAuditLog(order.ID, nil, domain.AuditActionStatusChange, auditDetails)
	if err := s.auditRepo.Create(ctx, auditLog); err != nil {
		s.logger.Warn("Failed to create audit log", "error", err, "order_id", order.ID)
	}
	
	// Determine event type based on new status
	var eventType domain.EventType
	switch order.Status {
	case domain.OrderStatusConfirmed:
		eventType = domain.EventTypeOrderConfirmed
	case domain.OrderStatusShipped:
		eventType = domain.EventTypeOrderShipped
	case domain.OrderStatusDelivered:
		eventType = domain.EventTypeOrderDelivered
	case domain.OrderStatusCancelled:
		eventType = domain.EventTypeOrderCancelled
	default:
		eventType = domain.EventTypeOrderUpdated
	}
	
	// Create and store event for outbox pattern
	eventPayload := map[string]interface{}{
		"order_id":    order.ID,
		"customer_id": order.CustomerID,
		"old_status":  oldStatus,
		"new_status":  order.Status,
		"updated_at":  order.UpdatedAt,
	}
	event := domain.NewOrderEvent(order.ID, eventType, eventPayload)
	if err := s.eventRepo.Create(ctx, event); err != nil {
		s.logger.Error("Failed to create order status change event", "error", err, "order_id", order.ID)
	}
	
	s.logger.Info("Order status updated", "order_id", order.ID, "old_status", oldStatus, "new_status", order.Status)
	return dto.ToOrderResponse(order), nil
}

// DeleteOrder deletes an order by ID
func (s *OrderService) DeleteOrder(ctx context.Context, id uuid.UUID) error {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	
	// Only allow deletion of pending orders
	if order.Status != domain.OrderStatusPending {
		return domain.ErrOrderCannotBeModified
	}
	
	// Create audit log entry before deletion
	auditDetails := map[string]interface{}{
		"deleted_order": map[string]interface{}{
			"customer_id":    order.CustomerID,
			"status":         order.Status,
			"total_amount":   order.TotalAmount,
		},
	}
	auditLog := domain.NewAuditLog(order.ID, nil, domain.AuditActionCancel, auditDetails)
	if err := s.auditRepo.Create(ctx, auditLog); err != nil {
		s.logger.Warn("Failed to create audit log for deletion", "error", err, "order_id", order.ID)
	}
	
	if err := s.orderRepo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete order", "error", err, "order_id", id)
		return err
	}
	
	s.logger.Info("Order deleted successfully", "order_id", id)
	return nil
}

// ListOrders retrieves orders with pagination
func (s *OrderService) ListOrders(ctx context.Context, page, pageSize int) (*dto.OrderListResponse, error) {
	offset := (page - 1) * pageSize
	orders, err := s.orderRepo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, err
	}
	
	// Load items for each order
	for _, order := range orders {
		items, err := s.orderItemRepo.GetByOrderID(ctx, order.ID)
		if err != nil {
			return nil, err
		}
		
		order.Items = make([]domain.OrderItem, len(items))
		for i, item := range items {
			order.Items[i] = *item
		}
	}
	
	return &dto.OrderListResponse{
		Orders:     dto.ToOrderResponseList(orders),
		TotalCount: len(orders), // Note: In production, you'd get this from a separate count query
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// GetOrdersByStatus retrieves orders by status
func (s *OrderService) GetOrdersByStatus(ctx context.Context, status domain.OrderStatus) ([]dto.OrderResponse, error) {
	orders, err := s.orderRepo.GetByStatus(ctx, status)
	if err != nil {
		return nil, err
	}
	
	// Load items for each order
	for _, order := range orders {
		items, err := s.orderItemRepo.GetByOrderID(ctx, order.ID)
		if err != nil {
			return nil, err
		}
		
		order.Items = make([]domain.OrderItem, len(items))
		for i, item := range items {
			order.Items[i] = *item
		}
	}
	
	return dto.ToOrderResponseList(orders), nil
}
