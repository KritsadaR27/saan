package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/saan/order-service/internal/domain"
)

// CreateOrderRequest represents the request to create a new order
type CreateOrderRequest struct {
	CustomerID      uuid.UUID               `json:"customer_id" validate:"required"`
	Source          *domain.OrderSource     `json:"source,omitempty"`
	ShippingAddress string                  `json:"shipping_address" validate:"required"`
	BillingAddress  string                  `json:"billing_address" validate:"required"`
	PaymentMethod   *domain.PaymentMethod   `json:"payment_method,omitempty"`
	PromoCode       *string                 `json:"promo_code,omitempty"`
	Notes           string                  `json:"notes"`
	TaxEnabled      *bool                   `json:"tax_enabled,omitempty"`
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
	ShippingAddress *string                `json:"shipping_address,omitempty"`
	BillingAddress  *string                `json:"billing_address,omitempty"`
	PaymentMethod   *domain.PaymentMethod  `json:"payment_method,omitempty"`
	PromoCode       *string                `json:"promo_code,omitempty"`
	Notes           *string                `json:"notes,omitempty"`
	Discount        *float64               `json:"discount,omitempty"`
	ShippingFee     *float64               `json:"shipping_fee,omitempty"`
	TaxEnabled      *bool                  `json:"tax_enabled,omitempty"`
}

// UpdateOrderStatusRequest represents the request to update order status
type UpdateOrderStatusRequest struct {
	Status domain.OrderStatus `json:"status" validate:"required"`
}

// UpdatePaidStatusRequest represents the request to update payment status
type UpdatePaidStatusRequest struct {
	PaidStatus domain.PaidStatus `json:"paid_status" validate:"required"`
}

// ConfirmOrderRequest represents the request to confirm an order
type ConfirmOrderRequest struct {
	// Could include additional confirmation data in the future
}

// CancelOrderRequest represents the request to cancel an order
type CancelOrderRequest struct {
	Reason string `json:"reason" validate:"required"`
}

// ApplyDiscountRequest represents the request to apply discount
type ApplyDiscountRequest struct {
	Amount float64 `json:"amount" validate:"required,min=0"`
}

// SetShippingFeeRequest represents the request to set shipping fee
type SetShippingFeeRequest struct {
	Fee float64 `json:"fee" validate:"required,min=0"`
}

// CalculateTaxRequest represents the request to calculate tax
type CalculateTaxRequest struct {
	TaxRate float64 `json:"tax_rate" validate:"required,min=0,max=1"`
}

// OrderResponse represents an order in the response
type OrderResponse struct {
	ID               uuid.UUID              `json:"id"`
	CustomerID       uuid.UUID              `json:"customer_id"`
	Code             *string                `json:"code,omitempty"`
	Status           domain.OrderStatus     `json:"status"`
	Source           domain.OrderSource     `json:"source"`
	PaidStatus       domain.PaidStatus      `json:"paid_status"`
	TotalAmount      float64                `json:"total_amount"`
	Discount         float64                `json:"discount"`
	ShippingFee      float64                `json:"shipping_fee"`
	Tax              float64                `json:"tax"`
	TaxEnabled       bool                   `json:"tax_enabled"`
	ShippingAddress  string                 `json:"shipping_address"`
	BillingAddress   string                 `json:"billing_address"`
	PaymentMethod    *domain.PaymentMethod  `json:"payment_method,omitempty"`
	PromoCode        *string                `json:"promo_code,omitempty"`
	Notes            string                 `json:"notes"`
	ConfirmedAt      *time.Time             `json:"confirmed_at,omitempty"`
	CancelledAt      *time.Time             `json:"cancelled_at,omitempty"`
	CancelledReason  *string                `json:"cancelled_reason,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	Items            []OrderItemResponse    `json:"items,omitempty"`
}

// OrderItemResponse represents an order item in the response
type OrderItemResponse struct {
	ID             uuid.UUID `json:"id"`
	ProductID      uuid.UUID `json:"product_id"`
	Quantity       int       `json:"quantity"`
	UnitPrice      float64   `json:"unit_price"`
	TotalPrice     float64   `json:"total_price"`
	IsOverride     bool      `json:"is_override"`
	OverrideReason *string   `json:"override_reason,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// OrderListResponse represents a paginated list of orders
type OrderListResponse struct {
	Orders     []OrderResponse `json:"orders"`
	TotalCount int            `json:"total_count"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
}

// AuditLogResponse represents an audit log entry in the response
type AuditLogResponse struct {
	ID        uuid.UUID              `json:"id"`
	OrderID   uuid.UUID              `json:"order_id"`
	UserID    *string                `json:"user_id,omitempty"`
	Action    domain.AuditAction     `json:"action"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// EventResponse represents an event in the response
type EventResponse struct {
	ID         uuid.UUID              `json:"id"`
	OrderID    uuid.UUID              `json:"order_id"`
	EventType  domain.EventType       `json:"event_type"`
	Payload    map[string]interface{} `json:"payload"`
	Status     domain.EventStatus     `json:"status"`
	CreatedAt  time.Time              `json:"created_at"`
	SentAt     *time.Time             `json:"sent_at,omitempty"`
	RetryCount int                    `json:"retry_count"`
}

// ConfirmOrderWithStockOverrideRequest represents the request to confirm an order with stock override
type ConfirmOrderWithStockOverrideRequest struct {
	UserID         uuid.UUID                    `json:"user_id" validate:"required"` // User performing the override
	UserRole       string                      `json:"user_role" validate:"required"` // User role for permission check
	OverrideItems  []StockOverrideItem         `json:"override_items" validate:"required,min=1"`
}

// StockOverrideItem represents an item that needs stock override
type StockOverrideItem struct {
	ProductID      uuid.UUID `json:"product_id" validate:"required"`
	Quantity       int       `json:"quantity" validate:"required,min=1"`
	OverrideReason string    `json:"override_reason" validate:"required"`
}

// ToOrderResponse converts a domain order to response DTO
func ToOrderResponse(order *domain.Order) *OrderResponse {
	response := &OrderResponse{
		ID:               order.ID,
		CustomerID:       order.CustomerID,
		Code:             order.Code,
		Status:           order.Status,
		Source:           order.Source,
		PaidStatus:       order.PaidStatus,
		TotalAmount:      order.TotalAmount,
		Discount:         order.Discount,
		ShippingFee:      order.ShippingFee,
		Tax:              order.Tax,
		TaxEnabled:       order.TaxEnabled,
		ShippingAddress:  order.ShippingAddress,
		BillingAddress:   order.BillingAddress,
		PaymentMethod:    order.PaymentMethod,
		PromoCode:        order.PromoCode,
		Notes:            order.Notes,
		ConfirmedAt:      order.ConfirmedAt,
		CancelledAt:      order.CancelledAt,
		CancelledReason:  order.CancelledReason,
		CreatedAt:        order.CreatedAt,
		UpdatedAt:        order.UpdatedAt,
		Items:            make([]OrderItemResponse, len(order.Items)),
	}
	
	for i, item := range order.Items {
		response.Items[i] = OrderItemResponse{
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

// ToAuditLogResponse converts a domain audit log to response DTO
func ToAuditLogResponse(auditLog *domain.OrderAuditLog) *AuditLogResponse {
	return &AuditLogResponse{
		ID:        auditLog.ID,
		OrderID:   auditLog.OrderID,
		UserID:    auditLog.UserID,
		Action:    auditLog.Action,
		Details:   auditLog.Details,
		Timestamp: auditLog.Timestamp,
	}
}

// ToEventResponse converts a domain event to response DTO
func ToEventResponse(event *domain.OrderEventOutbox) *EventResponse {
	return &EventResponse{
		ID:         event.ID,
		OrderID:    event.OrderID,
		EventType:  event.EventType,
		Payload:    event.Payload,
		Status:     event.Status,
		CreatedAt:  event.CreatedAt,
		SentAt:     event.SentAt,
		RetryCount: event.RetryCount,
	}
}
