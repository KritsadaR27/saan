package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"payment/internal/domain/entity"
)

// PaymentRepository defines the interface for payment data operations
type PaymentRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, payment *entity.PaymentTransaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.PaymentTransaction, error)
	Update(ctx context.Context, payment *entity.PaymentTransaction) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Query operations for the three data retrieval types
	
	// Type 1: Store-based queries (Loyverse integration)
	GetByStoreID(ctx context.Context, storeID string, filters PaymentFilters) ([]*entity.PaymentTransaction, error)
	GetStoreAnalytics(ctx context.Context, storeID string, dateFrom, dateTo time.Time) (*StorePaymentAnalytics, error)
	
	// Type 2: Customer-based queries
	GetByCustomerID(ctx context.Context, customerID uuid.UUID, filters PaymentFilters) ([]*entity.PaymentTransaction, error)
	GetCustomerPaymentHistory(ctx context.Context, customerID uuid.UUID, limit int) ([]*entity.PaymentTransaction, error)
	
	// Type 3: Order-based queries
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*entity.PaymentTransaction, error)
	GetOrderPaymentSummary(ctx context.Context, orderID uuid.UUID) (*OrderPaymentSummary, error)

	// Advanced queries
	GetByLoyverseReceiptID(ctx context.Context, receiptID string) (*entity.PaymentTransaction, error)
	GetPendingPayments(ctx context.Context, limit int) ([]*entity.PaymentTransaction, error)
	GetPaymentsByDateRange(ctx context.Context, dateFrom, dateTo time.Time, filters PaymentFilters) ([]*entity.PaymentTransaction, error)
	GetPaymentsByChannel(ctx context.Context, channel entity.PaymentChannel, filters PaymentFilters) ([]*entity.PaymentTransaction, error)
	GetPaymentsByMethod(ctx context.Context, method entity.PaymentMethod, filters PaymentFilters) ([]*entity.PaymentTransaction, error)

	// Bulk operations
	CreateBatch(ctx context.Context, payments []*entity.PaymentTransaction) error
	UpdateStatus(ctx context.Context, paymentID uuid.UUID, status entity.PaymentStatus) error
	UpdateStatusBatch(ctx context.Context, paymentIDs []uuid.UUID, status entity.PaymentStatus) error
}

// PaymentFilters represents filters for payment queries
type PaymentFilters struct {
	Status         *entity.PaymentStatus
	PaymentMethod  *entity.PaymentMethod
	PaymentChannel *entity.PaymentChannel
	PaymentTiming  *entity.PaymentTiming
	DateFrom       *time.Time
	DateTo         *time.Time
	MinAmount      *float64
	MaxAmount      *float64
	Limit          int
	Offset         int
	SortBy         string
	SortOrder      string // "ASC" or "DESC"
}

// StorePaymentAnalytics represents analytics data for a store
type StorePaymentAnalytics struct {
	StoreID           string    `json:"store_id"`
	TotalTransactions int       `json:"total_transactions"`
	TotalAmount       float64   `json:"total_amount"`
	AvgAmount         float64   `json:"avg_amount"`
	Currency          string    `json:"currency"`
	DateFrom          time.Time `json:"date_from"`
	DateTo            time.Time `json:"date_to"`
	
	// Payment method breakdown
	PaymentMethodStats []PaymentMethodStat `json:"payment_method_stats"`
	
	// Daily breakdown
	DailyStats []DailyPaymentStat `json:"daily_stats"`
}

// PaymentMethodStat represents payment statistics by method
type PaymentMethodStat struct {
	Method          entity.PaymentMethod `json:"method"`
	Count           int                  `json:"count"`
	TotalAmount     float64              `json:"total_amount"`
	PercentageCount float64              `json:"percentage_count"`
	PercentageAmount float64             `json:"percentage_amount"`
}

// DailyPaymentStat represents daily payment statistics
type DailyPaymentStat struct {
	Date   time.Time `json:"date"`
	Count  int       `json:"count"`
	Amount float64   `json:"amount"`
}

// OrderPaymentSummary represents payment summary for an order
type OrderPaymentSummary struct {
	OrderID           uuid.UUID `json:"order_id"`
	TotalAmount       float64   `json:"total_amount"`
	PaidAmount        float64   `json:"paid_amount"`
	PendingAmount     float64   `json:"pending_amount"`
	RefundedAmount    float64   `json:"refunded_amount"`
	Currency          string    `json:"currency"`
	PaymentStatus     string    `json:"payment_status"` // "fully_paid", "partially_paid", "unpaid"
	TransactionCount  int       `json:"transaction_count"`
	LastPaymentAt     *time.Time `json:"last_payment_at"`
	PaymentMethods    []string   `json:"payment_methods"`
}
