package entity

import (
	"time"
	"github.com/google/uuid"
)

// StoreType represents the type of Loyverse store
type StoreType string

const (
	StoreTypeMain      StoreType = "main"
	StoreTypeDelivery  StoreType = "delivery"
	StoreTypeWarehouse StoreType = "warehouse"
)

// LoyverseStore represents a Loyverse POS store configuration
type LoyverseStore struct {
	ID        uuid.UUID `json:"id" db:"id"`
	StoreID   string    `json:"store_id" db:"store_id"`
	StoreName string    `json:"store_name" db:"store_name"`
	StoreType StoreType `json:"store_type" db:"store_type"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	IsDefault bool      `json:"is_default" db:"is_default"`

	// Store capabilities
	AcceptsCash     bool `json:"accepts_cash" db:"accepts_cash"`
	AcceptsTransfer bool `json:"accepts_transfer" db:"accepts_transfer"`
	AcceptsCOD      bool `json:"accepts_cod" db:"accepts_cod"`

	// Delivery configuration
	DeliveryDriverPhone *string `json:"delivery_driver_phone,omitempty" db:"delivery_driver_phone"`
	DeliveryRoute       *string `json:"delivery_route,omitempty" db:"delivery_route"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Business logic methods for LoyverseStore
func (ls *LoyverseStore) CanProcessPaymentMethod(method PaymentMethod) bool {
	switch method {
	case PaymentMethodCash:
		return ls.AcceptsCash
	case PaymentMethodBankTransfer:
		return ls.AcceptsTransfer
	case PaymentMethodCODCash, PaymentMethodCODTransfer:
		return ls.AcceptsCOD
	default:
		return false
	}
}

func (ls *LoyverseStore) IsDeliveryStore() bool {
	return ls.StoreType == StoreTypeDelivery
}

func (ls *LoyverseStore) HasDeliveryDriver() bool {
	return ls.DeliveryDriverPhone != nil && *ls.DeliveryDriverPhone != ""
}

func (ls *LoyverseStore) GetDriverInfo() (string, string) {
	if ls.DeliveryDriverPhone == nil {
		return "", ""
	}
	
	phone := *ls.DeliveryDriverPhone
	route := ""
	if ls.DeliveryRoute != nil {
		route = *ls.DeliveryRoute
	}
	
	return phone, route
}

// Validation methods
func (ls *LoyverseStore) Validate() error {
	if ls.StoreID == "" {
		return ErrInvalidStoreID
	}
	if ls.StoreName == "" {
		return ErrInvalidStoreName
	}
	return nil
}
