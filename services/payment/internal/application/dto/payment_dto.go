package dto

import (
	"time"

	"github.com/google/uuid"
	"payment/internal/domain/entity"
)

// CreatePaymentRequest represents a request to create a new payment
type CreatePaymentRequest struct {
	OrderID        uuid.UUID                  `json:"order_id" validate:"required"`
	CustomerID     uuid.UUID                  `json:"customer_id" validate:"required"`
	PaymentMethod  entity.PaymentMethod       `json:"payment_method" validate:"required"`
	PaymentChannel entity.PaymentChannel      `json:"payment_channel" validate:"required"`
	PaymentTiming  entity.PaymentTiming       `json:"payment_timing" validate:"required"`
	Amount         float64                    `json:"amount" validate:"required,gt=0"`
	Currency       string                     `json:"currency" validate:"required,len=3"`
	AssignedStoreID *string                   `json:"assigned_store_id,omitempty"`
	Metadata       map[string]interface{}     `json:"metadata,omitempty"`
	
	// Delivery context (for COD payments)
	DeliveryContext *CreateDeliveryContextRequest `json:"delivery_context,omitempty"`
}

// CreateDeliveryContextRequest represents delivery context for COD payments
type CreateDeliveryContextRequest struct {
	DeliveryID       uuid.UUID `json:"delivery_id" validate:"required"`
	DriverID         *uuid.UUID `json:"driver_id,omitempty"`
	DeliveryAddress  string    `json:"delivery_address" validate:"required"`
	EstimatedArrival *time.Time `json:"estimated_arrival,omitempty"`
	Instructions     string    `json:"instructions,omitempty"`
}

// UpdatePaymentStatusRequest represents a request to update payment status
type UpdatePaymentStatusRequest struct {
	Status           entity.PaymentStatus `json:"status" validate:"required"`
	LoyverseReceiptID *string             `json:"loyverse_receipt_id,omitempty"`
	LoyversePaymentType *string          `json:"loyverse_payment_type,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// PaymentFiltersRequest represents filters for payment queries
type PaymentFiltersRequest struct {
	Status         *entity.PaymentStatus  `json:"status,omitempty"`
	PaymentMethod  *entity.PaymentMethod  `json:"payment_method,omitempty"`
	PaymentChannel *entity.PaymentChannel `json:"payment_channel,omitempty"`
	PaymentTiming  *entity.PaymentTiming  `json:"payment_timing,omitempty"`
	DateFrom       *time.Time             `json:"date_from,omitempty"`
	DateTo         *time.Time             `json:"date_to,omitempty"`
	MinAmount      *float64               `json:"min_amount,omitempty"`
	MaxAmount      *float64               `json:"max_amount,omitempty"`
	Limit          int                    `json:"limit,omitempty"`
	Offset         int                    `json:"offset,omitempty"`
	SortBy         string                 `json:"sort_by,omitempty"`
	SortOrder      string                 `json:"sort_order,omitempty"`
}

// Type 1 DTOs: Store-based data retrieval
type GetStorePaymentsRequest struct {
	StoreID string                 `json:"store_id" validate:"required"`
	Filters PaymentFiltersRequest  `json:"filters,omitempty"`
}

type GetStoreAnalyticsRequest struct {
	StoreID  string    `json:"store_id" validate:"required"`
	DateFrom time.Time `json:"date_from" validate:"required"`
	DateTo   time.Time `json:"date_to" validate:"required"`
}

// Type 2 DTOs: Customer-based data retrieval
type GetCustomerPaymentsRequest struct {
	CustomerID uuid.UUID             `json:"customer_id" validate:"required"`
	Filters    PaymentFiltersRequest `json:"filters,omitempty"`
}

type GetCustomerPaymentHistoryRequest struct {
	CustomerID uuid.UUID `json:"customer_id" validate:"required"`
	Limit      int       `json:"limit,omitempty"`
}

// Type 3 DTOs: Order-based data retrieval
type GetOrderPaymentsRequest struct {
	OrderID uuid.UUID `json:"order_id" validate:"required"`
}

type GetOrderPaymentSummaryRequest struct {
	OrderID uuid.UUID `json:"order_id" validate:"required"`
}

// Response DTOs
type PaymentResponse struct {
	ID                  uuid.UUID                  `json:"id"`
	OrderID             uuid.UUID                  `json:"order_id"`
	CustomerID          uuid.UUID                  `json:"customer_id"`
	PaymentMethod       entity.PaymentMethod       `json:"payment_method"`
	PaymentChannel      entity.PaymentChannel      `json:"payment_channel"`
	PaymentTiming       entity.PaymentTiming       `json:"payment_timing"`
	Amount              float64                    `json:"amount"`
	Currency            string                     `json:"currency"`
	Status              entity.PaymentStatus       `json:"status"`
	PaidAt              *time.Time                 `json:"paid_at,omitempty"`
	LoyverseReceiptID   *string                    `json:"loyverse_receipt_id,omitempty"`
	LoyversePaymentType *string                    `json:"loyverse_payment_type,omitempty"`
	AssignedStoreID     *string                    `json:"assigned_store_id,omitempty"`
	Metadata            map[string]interface{}     `json:"metadata,omitempty"`
	CreatedAt           time.Time                  `json:"created_at"`
	UpdatedAt           time.Time                  `json:"updated_at"`
	
	// Extended information
	DeliveryContext     *DeliveryContextResponse   `json:"delivery_context,omitempty"`
}

type DeliveryContextResponse struct {
	PaymentID         uuid.UUID  `json:"payment_id"`
	DeliveryID        uuid.UUID  `json:"delivery_id"`
	DriverID          *uuid.UUID `json:"driver_id,omitempty"`
	DeliveryAddress   string     `json:"delivery_address"`
	DeliveryStatus    string     `json:"delivery_status"`
	EstimatedArrival  *time.Time `json:"estimated_arrival,omitempty"`
	ActualArrival     *time.Time `json:"actual_arrival,omitempty"`
	Instructions      string     `json:"instructions,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type PaymentListResponse struct {
	Payments    []*PaymentResponse `json:"payments"`
	Total       int                `json:"total"`
	Page        int                `json:"page"`
	PerPage     int                `json:"per_page"`
	HasMore     bool               `json:"has_more"`
}

type StoreAnalyticsResponse struct {
	StoreID           string                      `json:"store_id"`
	TotalTransactions int                         `json:"total_transactions"`
	TotalAmount       float64                     `json:"total_amount"`
	AvgAmount         float64                     `json:"avg_amount"`
	Currency          string                      `json:"currency"`
	DateFrom          time.Time                   `json:"date_from"`
	DateTo            time.Time                   `json:"date_to"`
	PaymentMethodStats []PaymentMethodStatResponse `json:"payment_method_stats"`
	DailyStats        []DailyPaymentStatResponse   `json:"daily_stats"`
}

type PaymentMethodStatResponse struct {
	Method           entity.PaymentMethod `json:"method"`
	Count            int                  `json:"count"`
	TotalAmount      float64              `json:"total_amount"`
	PercentageCount  float64              `json:"percentage_count"`
	PercentageAmount float64              `json:"percentage_amount"`
}

type DailyPaymentStatResponse struct {
	Date   time.Time `json:"date"`
	Count  int       `json:"count"`
	Amount float64   `json:"amount"`
}

type OrderPaymentSummaryResponse struct {
	OrderID           uuid.UUID  `json:"order_id"`
	TotalAmount       float64    `json:"total_amount"`
	PaidAmount        float64    `json:"paid_amount"`
	PendingAmount     float64    `json:"pending_amount"`
	RefundedAmount    float64    `json:"refunded_amount"`
	Currency          string     `json:"currency"`
	PaymentStatus     string     `json:"payment_status"`
	TransactionCount  int        `json:"transaction_count"`
	LastPaymentAt     *time.Time `json:"last_payment_at"`
	PaymentMethods    []string   `json:"payment_methods"`
}

// Error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Code    string                 `json:"code"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
