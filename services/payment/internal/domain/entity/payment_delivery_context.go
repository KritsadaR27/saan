package entity

import (
	"time"
	"github.com/google/uuid"
)

// PaymentDeliveryContext represents delivery context for COD payments
type PaymentDeliveryContext struct {
	PaymentID        uuid.UUID  `json:"payment_id" db:"payment_id"`
	DeliveryID       uuid.UUID  `json:"delivery_id" db:"delivery_id"`
	DriverID         *uuid.UUID `json:"driver_id,omitempty" db:"driver_id"`
	
	// Delivery details
	DeliveryAddress   string     `json:"delivery_address" db:"delivery_address"`
	DeliveryStatus    string     `json:"delivery_status" db:"delivery_status"`
	EstimatedArrival  *time.Time `json:"estimated_arrival,omitempty" db:"estimated_arrival"`
	ActualArrival     *time.Time `json:"actual_arrival,omitempty" db:"actual_arrival"`
	Instructions      string     `json:"instructions,omitempty" db:"instructions"`
	
	// COD collection details
	CODAmount         *float64   `json:"cod_amount,omitempty" db:"cod_amount"`
	CODCollectedAt    *time.Time `json:"cod_collected_at,omitempty" db:"cod_collected_at"`
	CODCollectionMethod *string  `json:"cod_collection_method,omitempty" db:"cod_collection_method"`
	
	// GPS tracking
	PickupLat     *float64 `json:"pickup_lat,omitempty" db:"pickup_lat"`
	PickupLng     *float64 `json:"pickup_lng,omitempty" db:"pickup_lng"`
	DeliveryLat   *float64 `json:"delivery_lat,omitempty" db:"delivery_lat"`
	DeliveryLng   *float64 `json:"delivery_lng,omitempty" db:"delivery_lng"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty" db:"metadata"`

	// Audit fields
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Business logic methods

// IsCompleted checks if delivery is completed
func (pdc *PaymentDeliveryContext) IsCompleted() bool {
	return pdc.DeliveryStatus == "completed" && pdc.ActualArrival != nil
}

// IsPending checks if delivery is pending
func (pdc *PaymentDeliveryContext) IsPending() bool {
	return pdc.DeliveryStatus == "pending"
}

// IsInProgress checks if delivery is in progress
func (pdc *PaymentDeliveryContext) IsInProgress() bool {
	return pdc.DeliveryStatus == "in_progress" || pdc.DeliveryStatus == "picked_up"
}

// IsCODCollected checks if COD has been collected
func (pdc *PaymentDeliveryContext) IsCODCollected() bool {
	return pdc.CODCollectedAt != nil
}

// HasGPSTracking checks if GPS coordinates are available
func (pdc *PaymentDeliveryContext) HasGPSTracking() bool {
	return pdc.PickupLat != nil && pdc.PickupLng != nil && 
		   pdc.DeliveryLat != nil && pdc.DeliveryLng != nil
}

// UpdateDeliveryStatus updates the delivery status
func (pdc *PaymentDeliveryContext) UpdateDeliveryStatus(status string) error {
	if status == "" {
		return ErrInvalidDeliveryStatus
	}
	
	pdc.DeliveryStatus = status
	pdc.UpdatedAt = time.Now()
	
	// Set actual arrival time when completed
	if status == "completed" && pdc.ActualArrival == nil {
		now := time.Now()
		pdc.ActualArrival = &now
	}
	
	return nil
}

// UpdateGPSLocation updates GPS coordinates
func (pdc *PaymentDeliveryContext) UpdateGPSLocation(pickupLat, pickupLng, deliveryLat, deliveryLng *float64) error {
	if pickupLat != nil && pickupLng != nil {
		if !isValidLatitude(*pickupLat) || !isValidLongitude(*pickupLng) {
			return ErrInvalidGPSCoordinates
		}
		pdc.PickupLat = pickupLat
		pdc.PickupLng = pickupLng
	}
	
	if deliveryLat != nil && deliveryLng != nil {
		if !isValidLatitude(*deliveryLat) || !isValidLongitude(*deliveryLng) {
			return ErrInvalidGPSCoordinates
		}
		pdc.DeliveryLat = deliveryLat
		pdc.DeliveryLng = deliveryLng
	}
	
	pdc.UpdatedAt = time.Now()
	return nil
}

// CollectCOD marks COD as collected
func (pdc *PaymentDeliveryContext) CollectCOD(amount float64, method string) error {
	if amount <= 0 {
		return ErrInvalidCODAmount
	}
	
	if method == "" {
		method = "cash"
	}
	
	now := time.Now()
	pdc.CODAmount = &amount
	pdc.CODCollectedAt = &now
	pdc.CODCollectionMethod = &method
	pdc.UpdatedAt = now
	
	return nil
}

// Validate validates the delivery context
func (pdc *PaymentDeliveryContext) Validate() error {
	if pdc.PaymentID == uuid.Nil {
		return ErrInvalidPaymentTransactionID
	}
	
	if pdc.DeliveryID == uuid.Nil {
		return ErrInvalidDeliveryID
	}
	
	if pdc.DeliveryAddress == "" {
		return ErrInvalidDeliveryAddress
	}
	
	if pdc.DeliveryStatus == "" {
		return ErrInvalidDeliveryStatus
	}
	
	// Validate GPS coordinates if provided
	if pdc.PickupLat != nil && !isValidLatitude(*pdc.PickupLat) {
		return ErrInvalidGPSCoordinates
	}
	
	if pdc.PickupLng != nil && !isValidLongitude(*pdc.PickupLng) {
		return ErrInvalidGPSCoordinates
	}
	
	if pdc.DeliveryLat != nil && !isValidLatitude(*pdc.DeliveryLat) {
		return ErrInvalidGPSCoordinates
	}
	
	if pdc.DeliveryLng != nil && !isValidLongitude(*pdc.DeliveryLng) {
		return ErrInvalidGPSCoordinates
	}
	
	return nil
}

// Helper functions for GPS validation
func isValidLatitude(lat float64) bool {
	return lat >= -90.0 && lat <= 90.0
}

func isValidLongitude(lng float64) bool {
	return lng >= -180.0 && lng <= 180.0
}
