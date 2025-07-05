package entity

import (
	"time"
	"github.com/google/uuid"
)

// PaymentMethod represents the method of payment
type PaymentMethod string

const (
	PaymentMethodCash         PaymentMethod = "cash"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
	PaymentMethodCODCash      PaymentMethod = "cod_cash"
	PaymentMethodCODTransfer  PaymentMethod = "cod_transfer"
	PaymentMethodDigitalWallet PaymentMethod = "digital_wallet"
)

// PaymentChannel represents where the payment was initiated
type PaymentChannel string

const (
	PaymentChannelLoyversePOS PaymentChannel = "loyverse_pos"
	PaymentChannelSAANApp     PaymentChannel = "saan_app"
	PaymentChannelSAANChat    PaymentChannel = "saan_chat"
	PaymentChannelDelivery    PaymentChannel = "delivery"
	PaymentChannelWebPortal   PaymentChannel = "web_portal"
)

// PaymentTiming represents when the payment occurs
type PaymentTiming string

const (
	PaymentTimingPrepaid PaymentTiming = "prepaid"
	PaymentTimingCOD     PaymentTiming = "cod"
)

// PaymentStatus represents the current status of payment
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusCompleted  PaymentStatus = "completed"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusRefunded   PaymentStatus = "refunded"
	PaymentStatusCancelled  PaymentStatus = "cancelled"
)

// PaymentTransaction represents a payment transaction
type PaymentTransaction struct {
	ID         uuid.UUID `json:"id" db:"id"`
	OrderID    uuid.UUID `json:"order_id" db:"order_id"`
	CustomerID uuid.UUID `json:"customer_id" db:"customer_id"`

	// Payment details
	PaymentMethod  PaymentMethod  `json:"payment_method" db:"payment_method"`
	PaymentChannel PaymentChannel `json:"payment_channel" db:"payment_channel"`
	PaymentTiming  PaymentTiming  `json:"payment_timing" db:"payment_timing"`
	Amount         float64        `json:"amount" db:"amount"`
	Currency       string         `json:"currency" db:"currency"`

	// Payment status
	Status PaymentStatus `json:"status" db:"status"`
	PaidAt *time.Time    `json:"paid_at,omitempty" db:"paid_at"`

	// Loyverse integration
	LoyverseReceiptID   *string `json:"loyverse_receipt_id,omitempty" db:"loyverse_receipt_id"`
	LoyversePaymentType *string `json:"loyverse_payment_type,omitempty" db:"loyverse_payment_type"`
	AssignedStoreID     *string `json:"assigned_store_id,omitempty" db:"assigned_store_id"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty" db:"metadata"`

	// Audit fields
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty" db:"updated_by"`
}

// Business logic methods
func (pt *PaymentTransaction) IsCompleted() bool {
	return pt.Status == PaymentStatusCompleted
}

func (pt *PaymentTransaction) IsPending() bool {
	return pt.Status == PaymentStatusPending
}

func (pt *PaymentTransaction) IsCOD() bool {
	return pt.PaymentTiming == PaymentTimingCOD
}

func (pt *PaymentTransaction) IsOnlinePayment() bool {
	return pt.PaymentChannel == PaymentChannelSAANApp || 
		   pt.PaymentChannel == PaymentChannelSAANChat ||
		   pt.PaymentChannel == PaymentChannelWebPortal
}

func (pt *PaymentTransaction) GetLoyversePaymentType() string {
	switch pt.PaymentMethod {
	case PaymentMethodCash, PaymentMethodCODCash:
		return "cash"
	case PaymentMethodBankTransfer, PaymentMethodCODTransfer:
		return "bank_transfer"
	case PaymentMethodDigitalWallet:
		return "digital_wallet"
	default:
		return "cash"
	}
}

func (pt *PaymentTransaction) CanUpdateStatus(newStatus PaymentStatus) bool {
	validTransitions := map[PaymentStatus][]PaymentStatus{
		PaymentStatusPending: {
			PaymentStatusProcessing, 
			PaymentStatusCompleted, 
			PaymentStatusFailed, 
			PaymentStatusCancelled,
		},
		PaymentStatusProcessing: {
			PaymentStatusCompleted, 
			PaymentStatusFailed,
		},
		PaymentStatusCompleted: {
			PaymentStatusRefunded,
		},
		PaymentStatusFailed: {
			PaymentStatusPending,
		},
	}

	allowedStatuses, exists := validTransitions[pt.Status]
	if !exists {
		return false
	}

	for _, allowed := range allowedStatuses {
		if allowed == newStatus {
			return true
		}
	}
	return false
}

func (pt *PaymentTransaction) Validate() error {
	if pt.OrderID == uuid.Nil {
		return ErrInvalidOrderID
	}
	if pt.CustomerID == uuid.Nil {
		return ErrInvalidCustomerID
	}
	if pt.Amount <= 0 {
		return ErrInvalidAmount
	}
	if pt.Currency == "" {
		return ErrInvalidCurrency
	}
	return nil
}
