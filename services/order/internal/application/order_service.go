package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/saan/order-service/internal/domain"
	"github.com/saan/order-service/internal/application/dto"
)

// OrderService provides business logic for order operations
type OrderService struct {
	orderRepo     domain.OrderRepository
	orderItemRepo domain.OrderItemRepository
}

// NewOrderService creates a new order service instance
func NewOrderService(orderRepo domain.OrderRepository, orderItemRepo domain.OrderItemRepository) *OrderService {
	return &OrderService{
		orderRepo:     orderRepo,
		orderItemRepo: orderItemRepo,
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
		return nil, err
	}
	
	// Save order items
	for _, item := range order.Items {
		if err := s.orderItemRepo.Create(ctx, &item); err != nil {
			return nil, err
		}
	}
	
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
	
	// Update fields if provided
	if req.ShippingAddress != nil {
		order.ShippingAddress = *req.ShippingAddress
	}
	if req.BillingAddress != nil {
		order.BillingAddress = *req.BillingAddress
	}
	if req.Notes != nil {
		order.Notes = *req.Notes
	}
	
	// Save updated order
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, err
	}
	
	return dto.ToOrderResponse(order), nil
}

// UpdateOrderStatus updates the status of an order
func (s *OrderService) UpdateOrderStatus(ctx context.Context, id uuid.UUID, req *dto.UpdateOrderStatusRequest) (*dto.OrderResponse, error) {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// Update status with validation
	if err := order.UpdateStatus(req.Status); err != nil {
		return nil, err
	}
	
	// Save updated order
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, err
	}
	
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
	
	return s.orderRepo.Delete(ctx, id)
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
