package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/saan/order-service/internal/domain"
)

// CreateOrderRequest represents the request to create a new order
type CreateOrderRequest struct {
	CustomerID      uuid.UUID           `json:"customer_id" validate:"required"`
	ShippingAddress string              `json:"shipping_address" validate:"required"`
	BillingAddress  string              `json:"billing_address" validate:"required"`
	Notes           string              `json:"notes"`
	Items           []CreateOrderItemRequest `json:"items" validate:"required,min=1"`
}

// CreateOrderItemRequest represents an item in the create order request
type CreateOrderItemRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Quantity  int       `json:"quantity" validate:"required,min=1"`
	UnitPrice float64   `json:"unit_price" validate:"required,min=0"`
}

// UpdateOrderRequest represents the request to update an order
type UpdateOrderRequest struct {
	ShippingAddress *string `json:"shipping_address,omitempty"`
	BillingAddress  *string `json:"billing_address,omitempty"`
	Notes           *string `json:"notes,omitempty"`
}

// UpdateOrderStatusRequest represents the request to update order status
type UpdateOrderStatusRequest struct {
	Status domain.OrderStatus `json:"status" validate:"required"`
}

// OrderResponse represents an order in the response
type OrderResponse struct {
	ID              uuid.UUID           `json:"id"`
	CustomerID      uuid.UUID           `json:"customer_id"`
	Status          domain.OrderStatus  `json:"status"`
	TotalAmount     float64             `json:"total_amount"`
	ShippingAddress string              `json:"shipping_address"`
	BillingAddress  string              `json:"billing_address"`
	Notes           string              `json:"notes"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	Items           []OrderItemResponse `json:"items,omitempty"`
}

// OrderItemResponse represents an order item in the response
type OrderItemResponse struct {
	ID         uuid.UUID `json:"id"`
	ProductID  uuid.UUID `json:"product_id"`
	Quantity   int       `json:"quantity"`
	UnitPrice  float64   `json:"unit_price"`
	TotalPrice float64   `json:"total_price"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// OrderListResponse represents a paginated list of orders
type OrderListResponse struct {
	Orders     []OrderResponse `json:"orders"`
	TotalCount int            `json:"total_count"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
}

// ToOrderResponse converts a domain order to response DTO
func ToOrderResponse(order *domain.Order) *OrderResponse {
	response := &OrderResponse{
		ID:              order.ID,
		CustomerID:      order.CustomerID,
		Status:          order.Status,
		TotalAmount:     order.TotalAmount,
		ShippingAddress: order.ShippingAddress,
		BillingAddress:  order.BillingAddress,
		Notes:           order.Notes,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
		Items:           make([]OrderItemResponse, len(order.Items)),
	}
	
	for i, item := range order.Items {
		response.Items[i] = OrderItemResponse{
			ID:         item.ID,
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			UnitPrice:  item.UnitPrice,
			TotalPrice: item.TotalPrice,
			CreatedAt:  item.CreatedAt,
			UpdatedAt:  item.UpdatedAt,
		}
	}
	
	return response
}

// ToOrderResponseList converts a slice of domain orders to response DTOs
func ToOrderResponseList(orders []*domain.Order) []OrderResponse {
	responses := make([]OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = *ToOrderResponse(order)
	}
	return responses
}
