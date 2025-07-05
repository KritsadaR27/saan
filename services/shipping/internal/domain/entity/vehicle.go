package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// VehicleType represents the type of delivery vehicle
type VehicleType string

const (
	VehicleTypeMotorcycle VehicleType = "motorcycle"
	VehicleTypeCar        VehicleType = "car"
	VehicleTypeTruck      VehicleType = "truck"
	VehicleTypeVan        VehicleType = "van"
)

// VehicleStatus represents vehicle status
type VehicleStatus string

const (
	VehicleStatusActive      VehicleStatus = "active"
	VehicleStatusInactive    VehicleStatus = "inactive"
	VehicleStatusMaintenance VehicleStatus = "maintenance"
	VehicleStatusOnRoute     VehicleStatus = "on_route"
)

// DeliveryVehicle represents a delivery vehicle for self-delivery
type DeliveryVehicle struct {
	ID              uuid.UUID     `json:"id" db:"id"`
	LicensePlate    string        `json:"license_plate" db:"license_plate"`
	VehicleType     VehicleType   `json:"vehicle_type" db:"vehicle_type"`
	Brand           string        `json:"brand" db:"brand"`
	Model           string        `json:"model" db:"model"`
	Year            int           `json:"year" db:"year"`
	MaxWeight       float64       `json:"max_weight" db:"max_weight"`       // kg
	MaxVolume       float64       `json:"max_volume" db:"max_volume"`       // cubic meters
	FuelType        string        `json:"fuel_type" db:"fuel_type"`         // gasoline, diesel, electric
	DriverID        *uuid.UUID    `json:"driver_id" db:"driver_id"`
	Status          VehicleStatus `json:"status" db:"status"`
	CurrentLocation *string       `json:"current_location" db:"current_location"` // JSON coordinates
	LastMaintenance *time.Time    `json:"last_maintenance" db:"last_maintenance"`
	NextMaintenance *time.Time    `json:"next_maintenance" db:"next_maintenance"`
	Notes           *string       `json:"notes" db:"notes"`
	IsActive        bool          `json:"is_active" db:"is_active"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at" db:"updated_at"`
}

// NewDeliveryVehicle creates a new delivery vehicle
func NewDeliveryVehicle(licensePlate, brand, model string, vehicleType VehicleType, year int, maxWeight, maxVolume float64) *DeliveryVehicle {
	now := time.Now()
	return &DeliveryVehicle{
		ID:           uuid.New(),
		LicensePlate: licensePlate,
		VehicleType:  vehicleType,
		Brand:        brand,
		Model:        model,
		Year:         year,
		MaxWeight:    maxWeight,
		MaxVolume:    maxVolume,
		Status:       VehicleStatusActive,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// AssignDriver assigns a driver to the vehicle
func (v *DeliveryVehicle) AssignDriver(driverID uuid.UUID) {
	v.DriverID = &driverID
	v.UpdatedAt = time.Now()
}

// UnassignDriver removes driver from the vehicle
func (v *DeliveryVehicle) UnassignDriver() {
	v.DriverID = nil
	v.UpdatedAt = time.Now()
}

// UpdateStatus updates the vehicle status
func (v *DeliveryVehicle) UpdateStatus(status VehicleStatus) {
	v.Status = status
	v.UpdatedAt = time.Now()
}

// UpdateLocation updates the vehicle's current location
func (v *DeliveryVehicle) UpdateLocation(latitude, longitude float64) {
	// In a real implementation, this would update the vehicle's location in the database
	// For now, we just update the timestamp
	v.UpdatedAt = time.Now()
}

// ScheduleMaintenance schedules maintenance for the vehicle
func (v *DeliveryVehicle) ScheduleMaintenance(maintenanceDate time.Time) {
	v.NextMaintenance = &maintenanceDate
	v.UpdatedAt = time.Now()
}

// CompleteMaintenance marks maintenance as completed
func (v *DeliveryVehicle) CompleteMaintenance() {
	now := time.Now()
	v.LastMaintenance = &now
	v.NextMaintenance = nil
	v.Status = VehicleStatusActive
	v.UpdatedAt = now
}

// IsAvailable returns true if vehicle is available for delivery
func (v *DeliveryVehicle) IsAvailable() bool {
	return v.IsActive && (v.Status == VehicleStatusActive || v.Status == VehicleStatusInactive)
}

// CanCarry returns true if vehicle can carry the specified weight and volume
func (v *DeliveryVehicle) CanCarry(weight, volume float64) bool {
	return weight <= v.MaxWeight && volume <= v.MaxVolume
}

// Validate validates the vehicle data
func (v *DeliveryVehicle) Validate() error {
	if v.LicensePlate == "" {
		return ErrInvalidLicensePlate
	}
	if v.VehicleType == "" {
		return ErrInvalidVehicleType
	}
	if v.Brand == "" {
		return ErrInvalidBrand
	}
	if v.Model == "" {
		return ErrInvalidModel
	}
	if v.Year <= 0 {
		return ErrInvalidYear
	}
	if v.MaxWeight <= 0 {
		return ErrInvalidMaxWeight
	}
	if v.MaxVolume <= 0 {
		return ErrInvalidMaxVolume
	}
	return nil
}

// Vehicle domain errors
var (
	ErrVehicleNotFound       = errors.New("vehicle not found")
	ErrInvalidLicensePlate   = errors.New("invalid license plate")
	ErrInvalidVehicleType    = errors.New("invalid vehicle type")
	ErrInvalidBrand          = errors.New("invalid brand")
	ErrInvalidModel          = errors.New("invalid model")
	ErrInvalidYear           = errors.New("invalid year")
	ErrInvalidMaxWeight      = errors.New("invalid max weight")
	ErrInvalidMaxVolume      = errors.New("invalid max volume")
	ErrVehicleNotAvailable   = errors.New("vehicle not available")
	ErrInsufficientCapacity  = errors.New("insufficient vehicle capacity")
)
