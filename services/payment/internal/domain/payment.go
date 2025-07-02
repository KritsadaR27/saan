package domain

import (
	"time"
	"github.com/google/uuid"
)

// PaymentMethod represents the payment method
type PaymentMethod string

const (
	Cash         PaymentMethod = "cash"
	BankTransfer PaymentMethod = "bank_transfer"
	CreditCard   PaymentMethod = "credit_card"
	QRCode       PaymentMethod = "qr_code"
	Omise        PaymentMethod = "omise"
	C2P          PaymentMethod = "c2p"
	TrueMoney    PaymentMethod = "true_money"
	COD          PaymentMethod = "cod"
)

// PaymentStatus represents payment status
type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "pending"
	PaymentProcessing PaymentStatus = "processing"
	PaymentCompleted PaymentStatus = "completed"
	PaymentFailed    PaymentStatus = "failed"
	PaymentCancelled PaymentStatus = "cancelled"
	PaymentRefunded  PaymentStatus = "refunded"
)

// Payment represents a payment transaction
type Payment struct {
	ID              uuid.UUID     `json:"id" db:"id"`
	OrderID         uuid.UUID     `json:"order_id" db:"order_id"`
	CustomerID      uuid.UUID     `json:"customer_id" db:"customer_id"`
	PaymentMethod   PaymentMethod `json:"payment_method" db:"payment_method"`
	PaymentProvider string        `json:"payment_provider" db:"payment_provider"` // omise, c2p, etc.
	Amount          float64       `json:"amount" db:"amount"`
	Currency        string        `json:"currency" db:"currency"`
	Status          PaymentStatus `json:"status" db:"status"`
	
	// External payment details
	ExternalTransactionID *string   `json:"external_transaction_id,omitempty" db:"external_transaction_id"`
	ExternalPaymentID     *string   `json:"external_payment_id,omitempty" db:"external_payment_id"`
	PaymentGatewayFee     float64   `json:"payment_gateway_fee" db:"payment_gateway_fee"`
	
	// Timing
	PaymentDueDate     *time.Time `json:"payment_due_date,omitempty" db:"payment_due_date"`
	ProcessedAt        *time.Time `json:"processed_at,omitempty" db:"processed_at"`
	CompletedAt        *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	
	// Metadata
	PaymentMetadata    string    `json:"payment_metadata" db:"payment_metadata"` // JSON
	FailureReason      *string   `json:"failure_reason,omitempty" db:"failure_reason"`
	Notes              string    `json:"notes" db:"notes"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// PaymentSnapshot represents a snapshot of payment state
type PaymentSnapshot struct {
	ID            uuid.UUID     `json:"id" db:"id"`
	PaymentID     uuid.UUID     `json:"payment_id" db:"payment_id"`
	SnapshotType  string        `json:"snapshot_type" db:"snapshot_type"` // created, processing, completed, failed
	Status        PaymentStatus `json:"status" db:"status"`
	Amount        float64       `json:"amount" db:"amount"`
	PaymentMethod PaymentMethod `json:"payment_method" db:"payment_method"`
	SnapshotData  string        `json:"snapshot_data" db:"snapshot_data"` // JSON of payment state
	CreatedBy     *uuid.UUID    `json:"created_by,omitempty" db:"created_by"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
}

// Refund represents a refund transaction
type Refund struct {
	ID              uuid.UUID    `json:"id" db:"id"`
	PaymentID       uuid.UUID    `json:"payment_id" db:"payment_id"`
	OrderID         uuid.UUID    `json:"order_id" db:"order_id"`
	RefundAmount    float64      `json:"refund_amount" db:"refund_amount"`
	RefundReason    string       `json:"refund_reason" db:"refund_reason"`
	RefundMethod    PaymentMethod `json:"refund_method" db:"refund_method"`
	Status          PaymentStatus `json:"status" db:"status"`
	ExternalRefundID *string     `json:"external_refund_id,omitempty" db:"external_refund_id"`
	ProcessedAt     *time.Time   `json:"processed_at,omitempty" db:"processed_at"`
	CompletedAt     *time.Time   `json:"completed_at,omitempty" db:"completed_at"`
	Notes           string       `json:"notes" db:"notes"`
	CreatedBy       uuid.UUID    `json:"created_by" db:"created_by"`
	CreatedAt       time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at" db:"updated_at"`
}

// PaymentRequest represents a payment creation request
type PaymentRequest struct {
	OrderID         uuid.UUID     `json:"order_id" binding:"required"`
	CustomerID      uuid.UUID     `json:"customer_id" binding:"required"`
	Amount          float64       `json:"amount" binding:"required"`
	PaymentMethod   PaymentMethod `json:"payment_method" binding:"required"`
	PaymentProvider string        `json:"payment_provider,omitempty"`
	Currency        string        `json:"currency"`
	PaymentDueDate  *time.Time    `json:"payment_due_date,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// PaymentGatewayResponse represents response from payment gateway
type PaymentGatewayResponse struct {
	Success           bool                   `json:"success"`
	TransactionID     string                 `json:"transaction_id"`
	PaymentID         string                 `json:"payment_id"`
	Status            PaymentStatus          `json:"status"`
	Amount            float64                `json:"amount"`
	Fee               float64                `json:"fee"`
	PaymentURL        string                 `json:"payment_url,omitempty"`
	QRCodeData        string                 `json:"qr_code_data,omitempty"`
	ExpiresAt         *time.Time             `json:"expires_at,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	ErrorMessage      string                 `json:"error_message,omitempty"`
}

// Repository interfaces
type PaymentRepository interface {
	Create(payment *Payment) error
	GetByID(id uuid.UUID) (*Payment, error)
	GetByOrderID(orderID uuid.UUID) ([]*Payment, error)
	UpdateStatus(id uuid.UUID, status PaymentStatus) error
	GetPendingPayments() ([]*Payment, error)
}

type PaymentSnapshotRepository interface {
	Create(snapshot *PaymentSnapshot) error
	GetByPaymentID(paymentID uuid.UUID) ([]*PaymentSnapshot, error)
}

type RefundRepository interface {
	Create(refund *Refund) error
	GetByID(id uuid.UUID) (*Refund, error)
	GetByPaymentID(paymentID uuid.UUID) ([]*Refund, error)
	UpdateStatus(id uuid.UUID, status PaymentStatus) error
}

// Service interfaces
type PaymentService interface {
	CreatePayment(req *PaymentRequest) (*Payment, error)
	ProcessPayment(paymentID uuid.UUID) (*PaymentGatewayResponse, error)
	CompletePayment(paymentID uuid.UUID, externalTransactionID string) error
	FailPayment(paymentID uuid.UUID, reason string) error
	GetPaymentByID(id uuid.UUID) (*Payment, error)
	GetPaymentsByOrderID(orderID uuid.UUID) ([]*Payment, error)
	GetPaymentHistory(paymentID uuid.UUID) ([]*PaymentSnapshot, error)
}

type RefundService interface {
	CreateRefund(paymentID uuid.UUID, amount float64, reason string, createdBy uuid.UUID) (*Refund, error)
	ProcessRefund(refundID uuid.UUID) error
	GetRefundByID(id uuid.UUID) (*Refund, error)
	GetRefundsByPaymentID(paymentID uuid.UUID) ([]*Refund, error)
}

type PaymentGatewayService interface {
	CreatePayment(payment *Payment) (*PaymentGatewayResponse, error)
	CheckPaymentStatus(externalTransactionID string) (*PaymentGatewayResponse, error)
	ProcessRefund(payment *Payment, refundAmount float64) (*PaymentGatewayResponse, error)
}
