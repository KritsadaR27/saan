package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"order/internal/domain"
	"order/internal/application/dto"
	"order/internal/infrastructure/cache"
	"order/internal/infrastructure/events"
	"github.com/sirupsen/logrus"
)

// Service provides business logic for order operations following clean architecture
type Service struct {
	orderRepo      domain.OrderRepository
	orderItemRepo  domain.OrderItemRepository
	auditRepo      domain.OrderAuditRepository
	eventRepo      domain.OrderEventRepository
	eventPublisher events.Publisher
	cache          *cache.RedisClient
	logger         *logrus.Logger
}

// NewService creates a new order service instance
func NewService(
	orderRepo domain.OrderRepository,
	orderItemRepo domain.OrderItemRepository,
	auditRepo domain.OrderAuditRepository,
	eventRepo domain.OrderEventRepository,
	eventPublisher events.Publisher,
	cache *cache.RedisClient,
	logger *logrus.Logger,
) *Service {
	return &Service{
		orderRepo:      orderRepo,
		orderItemRepo:  orderItemRepo,
		auditRepo:      auditRepo,
		eventRepo:      eventRepo,
		eventPublisher: eventPublisher,
		cache:          cache,
		logger:         logger,
	}
}

// CreateOrder creates a new order with items
func (s *Service) CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	// Create new order
	order := domain.NewOrder(req.CustomerID, req.ShippingAddress, req.BillingAddress, req.Notes)
	
	// Add items to the order
	for _, itemReq := range req.Items {
		order.AddItem(itemReq.ProductID, itemReq.Quantity, itemReq.UnitPrice)
	}

	// Validate the order
	if err := order.Validate(); err != nil {
		s.logger.WithError(err).Error("Order validation failed")
		return nil, err
	}

	// Save order to database
	if err := s.orderRepo.Create(ctx, order); err != nil {
		s.logger.WithError(err).Error("Failed to create order")
		return nil, err
	}

	// Save order items
	for i := range order.Items {
		if err := s.orderItemRepo.Create(ctx, &order.Items[i]); err != nil {
			s.logger.WithError(err).Error("Failed to create order item")
			return nil, err
		}
	}

	// Cache the order (simplified - just add to customer orders)
	if err := s.cache.AddCustomerOrder(ctx, order.CustomerID.String(), order.ID.String(), 24*time.Hour); err != nil {
		s.logger.WithError(err).Warn("Failed to cache customer order")
		// Don't fail the operation for cache errors
	}

	// Create and publish order created event
	orderData := map[string]interface{}{
		"total_amount":      order.TotalAmount,
		"status":           string(order.Status),
		"shipping_address": order.ShippingAddress,
		"billing_address":  order.BillingAddress,
		"items":            convertItemsToEventData(order.Items),
	}

	if err := s.eventPublisher.PublishOrderCreated(ctx, order.ID.String(), order.CustomerID.String(), orderData); err != nil {
		s.logger.WithError(err).Error("Failed to publish order created event")
		// Don't fail the operation for event publishing errors
	}

	// Create audit record
	audit := domain.NewAuditLog(order.ID, nil, domain.AuditActionCreate, nil)
	if err := s.auditRepo.Create(ctx, audit); err != nil {
		s.logger.WithError(err).Warn("Failed to create audit record")
		// Don't fail the operation for audit errors
	}

	return s.orderToResponse(order), nil
}

// GetOrder retrieves an order by ID
func (s *Service) GetOrder(ctx context.Context, id uuid.UUID) (*dto.OrderResponse, error) {
	// Get from database
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("order_id", id).Error("Failed to get order")
		return nil, err
	}

	// Get order items
	items, err := s.orderItemRepo.GetByOrderID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("order_id", id).Error("Failed to get order items")
		return nil, err
	}
	
	// Convert items to slice
	order.Items = make([]domain.OrderItem, len(items))
	for i, item := range items {
		order.Items[i] = *item
	}

	return s.orderToResponse(order), nil
}

// UpdateOrderStatus updates the status of an order
func (s *Service) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("order_id", id).Error("Failed to get order for status update")
		return err
	}

	oldStatus := order.Status
	order.Status = status
	order.UpdatedAt = time.Now()

	// Update specific status-related timestamps
	switch status {
	case domain.OrderStatusConfirmed:
		now := time.Now()
		order.ConfirmedAt = &now
	case domain.OrderStatusCancelled:
		now := time.Now()
		order.CancelledAt = &now
	}

	// Save to database
	if err := s.orderRepo.Update(ctx, order); err != nil {
		s.logger.WithError(err).WithField("order_id", id).Error("Failed to update order status")
		return err
	}

	// Invalidate cache (simplified)
	if err := s.cache.DeleteOrder(ctx, id.String()); err != nil {
		s.logger.WithError(err).Warn("Failed to invalidate order cache")
	}

	// Publish status change event using the correct method signature
	changes := map[string]interface{}{
		"old_status": string(oldStatus),
		"new_status": string(status),
	}
	if err := s.eventPublisher.PublishOrderUpdated(ctx, order.ID.String(), order.CustomerID.String(), changes); err != nil {
		s.logger.WithError(err).Error("Failed to publish order status changed event")
	}

	// Create audit record
	auditChanges := map[string]interface{}{
		"old_status": string(oldStatus),
		"new_status": string(status),
	}
	audit := domain.NewAuditLog(order.ID, nil, domain.AuditActionStatusChange, auditChanges)
	if err := s.auditRepo.Create(ctx, audit); err != nil {
		s.logger.WithError(err).Warn("Failed to create audit record")
	}

	return nil
}

// ListOrders retrieves orders with pagination
func (s *Service) ListOrders(ctx context.Context, limit, offset int) ([]*dto.OrderResponse, error) {
	orders, err := s.orderRepo.List(ctx, limit, offset)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list orders")
		return nil, err
	}

	responses := make([]*dto.OrderResponse, len(orders))
	for i, order := range orders {
		// Get order items for each order
		items, err := s.orderItemRepo.GetByOrderID(ctx, order.ID)
		if err != nil {
			s.logger.WithError(err).WithField("order_id", order.ID).Error("Failed to get order items")
			return nil, err
		}
		
		// Convert items to slice
		order.Items = make([]domain.OrderItem, len(items))
		for j, item := range items {
			order.Items[j] = *item
		}
		
		responses[i] = s.orderToResponse(order)
	}

	return responses, nil
}

// GetOrdersByCustomer retrieves all orders for a customer
func (s *Service) GetOrdersByCustomer(ctx context.Context, customerID uuid.UUID) ([]*dto.OrderResponse, error) {
	// Get from database (skip cache complexity for now)
	orders, err := s.orderRepo.GetByCustomerID(ctx, customerID)
	if err != nil {
		s.logger.WithError(err).WithField("customer_id", customerID).Error("Failed to get customer orders")
		return nil, err
	}

	// Get items for each order
	for _, order := range orders {
		items, err := s.orderItemRepo.GetByOrderID(ctx, order.ID)
		if err != nil {
			s.logger.WithError(err).WithField("order_id", order.ID).Error("Failed to get order items")
			return nil, err
		}
		
		// Convert items to slice
		order.Items = make([]domain.OrderItem, len(items))
		for j, item := range items {
			order.Items[j] = *item
		}
	}

	responses := make([]*dto.OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = s.orderToResponse(order)
	}

	return responses, nil
}

// CancelOrder cancels an order
func (s *Service) CancelOrder(ctx context.Context, id uuid.UUID, reason string) error {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("order_id", id).Error("Failed to get order for cancellation")
		return err
	}

	if order.Status == domain.OrderStatusCancelled {
		return domain.ErrOrderAlreadyCancelled
	}

	if order.Status == domain.OrderStatusDelivered {
		return domain.ErrOrderCannotBeCancelled
	}

	oldStatus := order.Status
	order.Status = domain.OrderStatusCancelled
	order.CancelledReason = &reason
	now := time.Now()
	order.CancelledAt = &now
	order.UpdatedAt = now

	// Save to database
	if err := s.orderRepo.Update(ctx, order); err != nil {
		s.logger.WithError(err).WithField("order_id", id).Error("Failed to cancel order")
		return err
	}

	// Invalidate cache (simplified)
	if err := s.cache.DeleteOrder(ctx, id.String()); err != nil {
		s.logger.WithError(err).Warn("Failed to invalidate order cache")
	}

	// Publish cancellation event using correct method signature
	if err := s.eventPublisher.PublishOrderCancelled(ctx, order.ID.String(), order.CustomerID.String(), reason); err != nil {
		s.logger.WithError(err).Error("Failed to publish order cancelled event")
	}

	// Create audit record
	changes := map[string]interface{}{
		"old_status": string(oldStatus),
		"new_status": string(domain.OrderStatusCancelled),
		"reason":     reason,
	}
	audit := domain.NewAuditLog(order.ID, nil, domain.AuditActionCancel, changes)
	if err := s.auditRepo.Create(ctx, audit); err != nil {
		s.logger.WithError(err).Warn("Failed to create audit record")
	}

	return nil
}

// Helper functions

func (s *Service) orderToResponse(order *domain.Order) *dto.OrderResponse {
	items := make([]dto.OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		items[i] = dto.OrderItemResponse{
			ID:             item.ID,
			ProductID:      item.ProductID,
			Quantity:       item.Quantity,
			UnitPrice:      item.UnitPrice,
			TotalPrice:     item.TotalPrice,
			IsOverride:     item.IsOverride,
			OverrideReason: item.OverrideReason,
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
		}
	}

	return &dto.OrderResponse{
		ID:              order.ID,
		CustomerID:      order.CustomerID,
		Code:            order.Code,
		Status:          order.Status,
		Source:          order.Source,
		PaidStatus:      order.PaidStatus,
		TotalAmount:     order.TotalAmount,
		Discount:        order.Discount,
		ShippingFee:     order.ShippingFee,
		Tax:             order.Tax,
		TaxEnabled:      order.TaxEnabled,
		ShippingAddress: order.ShippingAddress,
		BillingAddress:  order.BillingAddress,
		PaymentMethod:   order.PaymentMethod,
		PromoCode:       order.PromoCode,
		Notes:           order.Notes,
		Items:           items,
		ConfirmedAt:     order.ConfirmedAt,
		CancelledAt:     order.CancelledAt,
		CancelledReason: order.CancelledReason,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}
}

func convertItemsToEventData(items []domain.OrderItem) []map[string]interface{} {
	eventItems := make([]map[string]interface{}, len(items))
	for i, item := range items {
		eventItems[i] = map[string]interface{}{
			"product_id":  item.ProductID.String(),
			"quantity":    item.Quantity,
			"unit_price":  item.UnitPrice,
			"total_price": item.TotalPrice,
		}
	}
	return eventItems
}
