package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// DeliverySnapshot represents a snapshot of delivery state for audit trail
type DeliverySnapshot struct {
	ID                  uuid.UUID              `json:"id"`
	DeliveryID          uuid.UUID              `json:"delivery_id"`
	
	// Snapshot Metadata
	SnapshotType        SnapshotType           `json:"snapshot_type"`
	SnapshotData        map[string]interface{} `json:"snapshot_data"`
	PreviousSnapshotID  *uuid.UUID             `json:"previous_snapshot_id"`
	
	// Audit Information
	TriggeredBy         string                 `json:"triggered_by"`
	TriggeredByUserID   *uuid.UUID             `json:"triggered_by_user_id"`
	TriggeredEvent      string                 `json:"triggered_event"`
	
	// Quick Access Fields (denormalized for performance)
	DeliveryStatus              string                 `json:"delivery_status"`
	CustomerID                  uuid.UUID              `json:"customer_id"`
	OrderID                     uuid.UUID              `json:"order_id"`
	VehicleID                   *uuid.UUID             `json:"vehicle_id"`
	DriverName                  string                 `json:"driver_name,omitempty"`
	DeliveryAddressProvince     string                 `json:"delivery_address_province"`
	DeliveryFee                 float64                `json:"delivery_fee"`
	ProviderCode                string                 `json:"provider_code"`
	
	// Timestamps
	CreatedAt                   time.Time              `json:"created_at"`
	BusinessDate                time.Time              `json:"business_date"`
}

// SnapshotType represents the type of snapshot
type SnapshotType string

const (
	SnapshotTypeCreated         SnapshotType = "created"
	SnapshotTypeAssigned        SnapshotType = "assigned"
	SnapshotTypePickedUp        SnapshotType = "picked_up"
	SnapshotTypeInTransit       SnapshotType = "in_transit"
	SnapshotTypeDelivered       SnapshotType = "delivered"
	SnapshotTypeFailed          SnapshotType = "failed"
	SnapshotTypeCancelled       SnapshotType = "cancelled"
	SnapshotTypeStatusUpdated   SnapshotType = "status_updated"
	SnapshotTypeProviderUpdated SnapshotType = "provider_updated"
)

// Domain errors
var (
	ErrSnapshotInvalidType      = errors.New("invalid snapshot type")
	ErrSnapshotInvalidData      = errors.New("snapshot data cannot be empty")
	ErrSnapshotInvalidTrigger   = errors.New("triggered by cannot be empty")
	ErrSnapshotInvalidDelivery  = errors.New("delivery ID cannot be empty")
	ErrSnapshotInvalidCustomer  = errors.New("customer ID cannot be empty")
	ErrSnapshotInvalidOrder     = errors.New("order ID cannot be empty")
)

// NewDeliverySnapshot creates a new delivery snapshot with validation
func NewDeliverySnapshot(
	deliveryID uuid.UUID,
	snapshotType SnapshotType,
	snapshotData map[string]interface{},
	triggeredBy, triggeredEvent string,
	triggeredByUserID *uuid.UUID,
) (*DeliverySnapshot, error) {
	if err := validateDeliveryID(deliveryID); err != nil {
		return nil, err
	}
	
	if err := validateSnapshotType(snapshotType); err != nil {
		return nil, err
	}
	
	if err := validateSnapshotData(snapshotData); err != nil {
		return nil, err
	}
	
	if err := validateTriggeredBy(triggeredBy); err != nil {
		return nil, err
	}
	
	now := time.Now()
	businessDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	
	return &DeliverySnapshot{
		ID:                  uuid.New(),
		DeliveryID:          deliveryID,
		SnapshotType:        snapshotType,
		SnapshotData:        snapshotData,
		TriggeredBy:         triggeredBy,
		TriggeredByUserID:   triggeredByUserID,
		TriggeredEvent:      triggeredEvent,
		CreatedAt:           now,
		BusinessDate:        businessDate,
	}, nil
}

// NewDeliverySnapshotFromDelivery creates a snapshot from a delivery entity
func NewDeliverySnapshotFromDelivery(
	delivery *DeliveryOrder,
	snapshotType SnapshotType,
	triggeredBy, triggeredEvent string,
	triggeredByUserID *uuid.UUID,
	previousSnapshotID *uuid.UUID,
) (*DeliverySnapshot, error) {
	// Create snapshot data from delivery
	snapshotData := map[string]interface{}{
		"delivery_id":              delivery.ID,
		"order_id":                 delivery.OrderID,
		"customer_id":              delivery.CustomerID,
		"customer_address_id":      delivery.CustomerAddressID,
		"delivery_method":          delivery.DeliveryMethod,
		"provider_id":              delivery.ProviderID,
		"vehicle_id":               delivery.VehicleID,
		"route_id":                 delivery.RouteID,
		"tracking_number":          delivery.TrackingNumber,
		"provider_order_id":        delivery.ProviderOrderID,
		"scheduled_pickup_time":    delivery.ScheduledPickupTime,
		"planned_delivery_date":    delivery.PlannedDeliveryDate,
		"estimated_delivery_time":  delivery.EstimatedDeliveryTime,
		"actual_pickup_time":       delivery.ActualPickupTime,
		"actual_delivery_time":     delivery.ActualDeliveryTime,
		"delivery_fee":             delivery.DeliveryFee,
		"cod_amount":               delivery.CODAmount,
		"status":                   delivery.Status,
		"notes":                    delivery.Notes,
		"delivery_instructions":    delivery.DeliveryInstructions,
		"requires_manual_coordination": delivery.RequiresManualCoordination,
		"is_active":                delivery.IsActive,
		"created_at":               delivery.CreatedAt,
		"updated_at":               delivery.UpdatedAt,
		"snapshot_created_at":      time.Now(),
	}
	
	snapshot := &DeliverySnapshot{
		ID:                      uuid.New(),
		DeliveryID:              delivery.ID,
		SnapshotType:            snapshotType,
		SnapshotData:            snapshotData,
		PreviousSnapshotID:      previousSnapshotID,
		TriggeredBy:             triggeredBy,
		TriggeredByUserID:       triggeredByUserID,
		TriggeredEvent:          triggeredEvent,
		DeliveryStatus:          string(delivery.Status),
		CustomerID:              delivery.CustomerID,
		OrderID:                 delivery.OrderID,
		VehicleID:               delivery.VehicleID,
		DeliveryFee:             delivery.DeliveryFee,
		ProviderCode:            string(delivery.DeliveryMethod), // Use delivery method as provider code
		CreatedAt:               time.Now(),
		BusinessDate:            time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location()),
	}
	
	return snapshot, nil
}

// SetQuickAccessFields sets the denormalized fields for performance
func (s *DeliverySnapshot) SetQuickAccessFields(
	deliveryStatus string,
	customerID, orderID uuid.UUID,
	vehicleID *uuid.UUID,
	driverName, province string,
	deliveryFee float64,
	providerCode string,
) {
	s.DeliveryStatus = deliveryStatus
	s.CustomerID = customerID
	s.OrderID = orderID
	s.VehicleID = vehicleID
	s.DriverName = driverName
	s.DeliveryAddressProvince = province
	s.DeliveryFee = deliveryFee
	s.ProviderCode = providerCode
}

// GetSnapshotDataValue safely gets a value from snapshot data
func (s *DeliverySnapshot) GetSnapshotDataValue(key string) interface{} {
	if s.SnapshotData == nil {
		return nil
	}
	return s.SnapshotData[key]
}

// GetSnapshotDataString safely gets a string value from snapshot data
func (s *DeliverySnapshot) GetSnapshotDataString(key string) string {
	value := s.GetSnapshotDataValue(key)
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}

// GetSnapshotDataFloat safely gets a float64 value from snapshot data
func (s *DeliverySnapshot) GetSnapshotDataFloat(key string) float64 {
	value := s.GetSnapshotDataValue(key)
	if f, ok := value.(float64); ok {
		return f
	}
	return 0.0
}

// GetSnapshotDataBool safely gets a bool value from snapshot data
func (s *DeliverySnapshot) GetSnapshotDataBool(key string) bool {
	value := s.GetSnapshotDataValue(key)
	if b, ok := value.(bool); ok {
		return b
	}
	return false
}

// GetSnapshotDataTime safely gets a time.Time value from snapshot data
func (s *DeliverySnapshot) GetSnapshotDataTime(key string) *time.Time {
	value := s.GetSnapshotDataValue(key)
	if t, ok := value.(time.Time); ok {
		return &t
	}
	if str, ok := value.(string); ok {
		if t, err := time.Parse(time.RFC3339, str); err == nil {
			return &t
		}
	}
	return nil
}

// IsBusinessEvent returns true if this snapshot represents a business event
func (s *DeliverySnapshot) IsBusinessEvent() bool {
	businessEvents := map[SnapshotType]bool{
		SnapshotTypeCreated:    true,
		SnapshotTypeAssigned:   true,
		SnapshotTypePickedUp:   true,
		SnapshotTypeInTransit:  true,
		SnapshotTypeDelivered:  true,
		SnapshotTypeFailed:     true,
		SnapshotTypeCancelled:  true,
	}
	
	return businessEvents[s.SnapshotType]
}

// IsSuccessfulDelivery returns true if this snapshot represents a successful delivery
func (s *DeliverySnapshot) IsSuccessfulDelivery() bool {
	return s.SnapshotType == SnapshotTypeDelivered
}

// IsFailedDelivery returns true if this snapshot represents a failed delivery
func (s *DeliverySnapshot) IsFailedDelivery() bool {
	return s.SnapshotType == SnapshotTypeFailed || s.SnapshotType == SnapshotTypeCancelled
}

// GetDeliveryDuration calculates delivery duration from snapshot data
func (s *DeliverySnapshot) GetDeliveryDuration() time.Duration {
	pickupTime := s.GetSnapshotDataTime("actual_pickup_time")
	deliveryTime := s.GetSnapshotDataTime("actual_delivery_time")
	
	if pickupTime != nil && deliveryTime != nil {
		return deliveryTime.Sub(*pickupTime)
	}
	
	return 0
}

// CompareWithPrevious compares this snapshot with previous snapshot data
func (s *DeliverySnapshot) CompareWithPrevious(previousSnapshot *DeliverySnapshot) map[string]interface{} {
	if previousSnapshot == nil {
		return map[string]interface{}{
			"has_changes": false,
			"changes":     []string{},
		}
	}
	
	changes := []string{}
	
	// Compare key fields
	if s.DeliveryStatus != previousSnapshot.DeliveryStatus {
		changes = append(changes, "status")
	}
	
	if s.VehicleID != previousSnapshot.VehicleID {
		changes = append(changes, "vehicle_assignment")
	}
	
	if s.DeliveryFee != previousSnapshot.DeliveryFee {
		changes = append(changes, "delivery_fee")
	}
	
	if s.ProviderCode != previousSnapshot.ProviderCode {
		changes = append(changes, "provider")
	}
	
	return map[string]interface{}{
		"has_changes": len(changes) > 0,
		"changes":     changes,
		"previous_snapshot_id": previousSnapshot.ID,
		"time_diff_minutes": s.CreatedAt.Sub(previousSnapshot.CreatedAt).Minutes(),
	}
}

// Validation functions
func validateDeliveryID(id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrSnapshotInvalidDelivery
	}
	return nil
}

func validateSnapshotType(snapshotType SnapshotType) error {
	validTypes := map[SnapshotType]bool{
		SnapshotTypeCreated:         true,
		SnapshotTypeAssigned:        true,
		SnapshotTypePickedUp:        true,
		SnapshotTypeInTransit:       true,
		SnapshotTypeDelivered:       true,
		SnapshotTypeFailed:          true,
		SnapshotTypeCancelled:       true,
		SnapshotTypeStatusUpdated:   true,
		SnapshotTypeProviderUpdated: true,
	}
	
	if !validTypes[snapshotType] {
		return ErrSnapshotInvalidType
	}
	return nil
}

func validateSnapshotData(data map[string]interface{}) error {
	if data == nil || len(data) == 0 {
		return ErrSnapshotInvalidData
	}
	return nil
}

func validateTriggeredBy(triggeredBy string) error {
	if triggeredBy == "" {
		return ErrSnapshotInvalidTrigger
	}
	return nil
}
