package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// DeliveryMethod represents the method of delivery
type DeliveryMethod string

const (
	DeliveryMethodSelfDelivery   DeliveryMethod = "self_delivery"
	DeliveryMethodGrab           DeliveryMethod = "grab"
	DeliveryMethodLineMan        DeliveryMethod = "lineman"
	DeliveryMethodLalamove       DeliveryMethod = "lalamove"
	DeliveryMethodInterExpress   DeliveryMethod = "inter_express"
	DeliveryMethodNimExpress     DeliveryMethod = "nim_express"
	DeliveryMethodRotRao         DeliveryMethod = "rot_rao"
)

// DeliveryStatus represents delivery status
type DeliveryStatus string

const (
	DeliveryStatusPending      DeliveryStatus = "pending"
	DeliveryStatusPlanned      DeliveryStatus = "planned"
	DeliveryStatusDispatched   DeliveryStatus = "dispatched"
	DeliveryStatusInTransit    DeliveryStatus = "in_transit"
	DeliveryStatusDelivered    DeliveryStatus = "delivered"
	DeliveryStatusFailed       DeliveryStatus = "failed"
	DeliveryStatusCancelled    DeliveryStatus = "cancelled"
)

// DeliveryOrder represents a delivery order
type DeliveryOrder struct {
	ID                      uuid.UUID      `json:"id" db:"id"`
	OrderID                 uuid.UUID      `json:"order_id" db:"order_id"`
	CustomerID              uuid.UUID      `json:"customer_id" db:"customer_id"`
	CustomerAddressID       uuid.UUID      `json:"customer_address_id" db:"customer_address_id"`
	DeliveryMethod          DeliveryMethod `json:"delivery_method" db:"delivery_method"`
	ProviderID              *uuid.UUID     `json:"provider_id" db:"provider_id"`
	VehicleID               *uuid.UUID     `json:"vehicle_id" db:"vehicle_id"`
	RouteID                 *uuid.UUID     `json:"route_id" db:"route_id"`
	TrackingNumber          *string        `json:"tracking_number" db:"tracking_number"`
	ProviderOrderID         *string        `json:"provider_order_id" db:"provider_order_id"`
	ScheduledPickupTime     *time.Time     `json:"scheduled_pickup_time" db:"scheduled_pickup_time"`
	PlannedDeliveryDate     time.Time      `json:"planned_delivery_date" db:"planned_delivery_date"`
	EstimatedDeliveryTime   *time.Time     `json:"estimated_delivery_time" db:"estimated_delivery_time"`
	ActualPickupTime        *time.Time     `json:"actual_pickup_time" db:"actual_pickup_time"`
	ActualDeliveryTime      *time.Time     `json:"actual_delivery_time" db:"actual_delivery_time"`
	DeliveryFee             float64        `json:"delivery_fee" db:"delivery_fee"`
	CODAmount               float64        `json:"cod_amount" db:"cod_amount"`
	Status                  DeliveryStatus `json:"status" db:"status"`
	Notes                   *string        `json:"notes" db:"notes"`
	DeliveryInstructions    *string        `json:"delivery_instructions" db:"delivery_instructions"`
	RequiresManualCoordination bool        `json:"requires_manual_coordination" db:"requires_manual_coordination"`
	IsActive                bool           `json:"is_active" db:"is_active"`
	CreatedAt               time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at" db:"updated_at"`
}

// NewDeliveryOrder creates a new delivery order
func NewDeliveryOrder(orderID, customerID, customerAddressID uuid.UUID, method DeliveryMethod, fee, codAmount float64) *DeliveryOrder {
	now := time.Now()
	return &DeliveryOrder{
		ID:                         uuid.New(),
		OrderID:                    orderID,
		CustomerID:                 customerID,
		CustomerAddressID:          customerAddressID,
		DeliveryMethod:             method,
		DeliveryFee:                fee,
		CODAmount:                  codAmount,
		Status:                     DeliveryStatusPending,
		PlannedDeliveryDate:        now.AddDate(0, 0, 1), // Default next day
		RequiresManualCoordination: method == DeliveryMethodInterExpress || method == DeliveryMethodNimExpress || method == DeliveryMethodRotRao,
		IsActive:                   true,
		CreatedAt:                  now,
		UpdatedAt:                  now,
	}
}

// UpdateStatus updates the delivery status
func (d *DeliveryOrder) UpdateStatus(status DeliveryStatus) {
	d.Status = status
	d.UpdatedAt = time.Now()
	
	// Set actual times based on status
	now := time.Now()
	switch status {
	case DeliveryStatusDispatched:
		if d.ActualPickupTime == nil {
			d.ActualPickupTime = &now
		}
	case DeliveryStatusDelivered:
		if d.ActualDeliveryTime == nil {
			d.ActualDeliveryTime = &now
		}
	}
}

// SetTrackingInfo sets tracking information from provider
func (d *DeliveryOrder) SetTrackingInfo(trackingNumber, providerOrderID string) {
	d.TrackingNumber = &trackingNumber
	d.ProviderOrderID = &providerOrderID
	d.UpdatedAt = time.Now()
}

// AssignVehicle assigns a vehicle for self-delivery
func (d *DeliveryOrder) AssignVehicle(vehicleID uuid.UUID, routeID uuid.UUID) {
	d.VehicleID = &vehicleID
	d.RouteID = &routeID
	d.UpdatedAt = time.Now()
}

// AssignProvider assigns a third-party provider
func (d *DeliveryOrder) AssignProvider(providerID uuid.UUID) {
	d.ProviderID = &providerID
	d.UpdatedAt = time.Now()
}

// Validate validates the delivery order
func (d *DeliveryOrder) Validate() error {
	if d.OrderID == uuid.Nil {
		return ErrInvalidOrderID
	}
	if d.CustomerID == uuid.Nil {
		return ErrInvalidCustomerID
	}
	if d.CustomerAddressID == uuid.Nil {
		return ErrInvalidAddressID
	}
	if d.DeliveryMethod == "" {
		return ErrInvalidDeliveryMethod
	}
	if d.DeliveryFee < 0 {
		return ErrInvalidDeliveryFee
	}
	if d.CODAmount < 0 {
		return ErrInvalidCODAmount
	}
	return nil
}

// IsManualCoordination returns true if delivery requires manual coordination
func (d *DeliveryOrder) IsManualCoordination() bool {
	return d.RequiresManualCoordination
}

// CanCancel returns true if delivery can be cancelled
func (d *DeliveryOrder) CanCancel() bool {
	return d.Status == DeliveryStatusPending || d.Status == DeliveryStatusPlanned
}

// CanUpdate returns true if delivery can be updated
func (d *DeliveryOrder) CanUpdate() bool {
	return d.Status != DeliveryStatusDelivered && d.Status != DeliveryStatusCancelled
}

// GetEstimatedDeliveryDuration returns estimated delivery duration in hours
func (d *DeliveryOrder) GetEstimatedDeliveryDuration() time.Duration {
	switch d.DeliveryMethod {
	case DeliveryMethodSelfDelivery:
		return 24 * time.Hour // Same day or next day
	case DeliveryMethodGrab, DeliveryMethodLineMan, DeliveryMethodLalamove:
		return 2 * time.Hour // 2 hours for on-demand
	case DeliveryMethodInterExpress:
		return 48 * time.Hour // 2 days
	case DeliveryMethodNimExpress:
		return 24 * time.Hour // 1 day
	case DeliveryMethodRotRao:
		return 72 * time.Hour // 3 days (traditional)
	default:
		return 24 * time.Hour
	}
}

// Domain errors for delivery
var (
	ErrDeliveryOrderNotFound    = errors.New("delivery order not found")
	ErrInvalidOrderID           = errors.New("invalid order ID")
	ErrInvalidCustomerID        = errors.New("invalid customer ID")
	ErrInvalidAddressID         = errors.New("invalid address ID")
	ErrInvalidDeliveryMethod    = errors.New("invalid delivery method")
	ErrInvalidDeliveryFee       = errors.New("invalid delivery fee")
	ErrInvalidCODAmount         = errors.New("invalid COD amount")
	ErrDeliveryCannotBeCancelled = errors.New("delivery cannot be cancelled")
	ErrDeliveryCannotBeUpdated  = errors.New("delivery cannot be updated")
	ErrInvalidStatusTransition  = errors.New("invalid status transition")
)
