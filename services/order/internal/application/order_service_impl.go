package application
package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/saan/order-service/internal/domain"
	"github.com/saan/order-service/internal/dto"
	"github.com/saan/order-service/pkg/logger"
)

// OrderService defines the interface for order business logic
type OrderService interface {
	CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.OrderResponse, error)
	GetOrder(ctx context.Context, id uuid.UUID) (*dto.OrderResponse, error)
	UpdateOrder(ctx context.Context, id uuid.UUID, req *dto.UpdateOrderRequest) (*dto.OrderResponse, error)
	DeleteOrder(ctx context.Context, id uuid.UUID) error
	ListOrders(ctx context.Context, filter *dto.OrderFilterRequest) ([]*dto.OrderResponse, int, error)
}

// OrderServiceImpl implements OrderService
type OrderServiceImpl struct {
	repo   domain.OrderRepository
	logger *logger.Logger
}

// NewOrderService creates a new OrderService
func NewOrderService(repo domain.OrderRepository, logger *logger.Logger) OrderService {
	return &OrderServiceImpl{
		repo:   repo,
		logger: logger,
	}
}

// CreateOrder creates a new order
func (s *OrderServiceImpl) CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	// Create domain order
	order := domain.NewOrder(req.CustomerID)
	
	// Add items
	for _, itemReq := range req.Items {
		item, err := domain.NewOrderItem(itemReq.ProductID, itemReq.Quantity, itemReq.Price)
		if err != nil {
			return nil, err
		}
		
		if err := order.AddItem(*item); err != nil {
			return nil, err
		}
	}
	
	// Validate order
	if err := order.Validate(); err != nil {
		return nil, err
	}
	
	// Save to repository
	if err := s.repo.Create(ctx, order); err != nil {
		s.logger.Error("Failed to create order: ", err)
		return nil, err
	}
	
	s.logger.Info("Order created successfully: ", order.ID.String())
	
	return s.domainToDTO(order), nil
}

// GetOrder retrieves an order by ID
func (s *OrderServiceImpl) GetOrder(ctx context.Context, id uuid.UUID) (*dto.OrderResponse, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get order: ", err)
		return nil, err
	}
	
	return s.domainToDTO(order), nil
}

// UpdateOrder updates an existing order
func (s *OrderServiceImpl) UpdateOrder(ctx context.Context, id uuid.UUID, req *dto.UpdateOrderRequest) (*dto.OrderResponse, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// Update status if provided
	if req.Status != nil {
		status := domain.OrderStatus(*req.Status)
		if err := order.UpdateStatus(status); err != nil {
			return nil, err
		}
	}
	
	// Update metadata if provided
	if req.Metadata != nil {
		order.Metadata = req.Metadata
	}
	
	// Save changes
	if err := s.repo.Update(ctx, order); err != nil {
		s.logger.Error("Failed to update order: ", err)
		return nil, err
	}
	
	s.logger.Info("Order updated successfully: ", order.ID.String())
	
	return s.domainToDTO(order), nil
}

// DeleteOrder deletes an order
func (s *OrderServiceImpl) DeleteOrder(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete order: ", err)
		return err
	}
	
	s.logger.Info("Order deleted successfully: ", id.String())
	return nil
}

// ListOrders lists orders with filtering
func (s *OrderServiceImpl) ListOrders(ctx context.Context, filter *dto.OrderFilterRequest) ([]*dto.OrderResponse, int, error) {
	domainFilter := domain.OrderFilter{
		CustomerID: filter.CustomerID,
		Status:     (*domain.OrderStatus)(filter.Status),
		Offset:     (filter.Page - 1) * filter.Limit,
		Limit:      filter.Limit,
	}
	
	orders, total, err := s.repo.List(ctx, domainFilter)
	if err != nil {
		s.logger.Error("Failed to list orders: ", err)
		return nil, 0, err
	}
	
	// Convert to DTOs
	dtoOrders := make([]*dto.OrderResponse, len(orders))
	for i, order := range orders {
		dtoOrders[i] = s.domainToDTO(order)
	}
	
	return dtoOrders, total, nil
}

// domainToDTO converts domain Order to DTO
func (s *OrderServiceImpl) domainToDTO(order *domain.Order) *dto.OrderResponse {
	items := make([]dto.OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		items[i] = dto.OrderItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
			Total:     item.Total,
		}
	}
	
	return &dto.OrderResponse{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		Status:     string(order.Status),
		Total:      order.Total,
		Items:      items,
		Metadata:   order.Metadata,
		CreatedAt:  order.CreatedAt,
		UpdatedAt:  order.UpdatedAt,
	}
}
