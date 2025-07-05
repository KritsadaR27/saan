package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"shipping/internal/domain/entity"
	"shipping/internal/domain/repository"
)

// VehicleUseCase handles vehicle management operations
type VehicleUseCase struct {
	vehicleRepo    repository.VehicleRepository
	eventPublisher EventPublisher
}

// NewVehicleUseCase creates a new vehicle use case
func NewVehicleUseCase(
	vehicleRepo repository.VehicleRepository,
	eventPublisher EventPublisher,
) *VehicleUseCase {
	return &VehicleUseCase{
		vehicleRepo:    vehicleRepo,
		eventPublisher: eventPublisher,
	}
}

// CreateVehicleRequest represents the request to create a vehicle
type CreateVehicleRequest struct {
	LicensePlate string              `json:"license_plate"`
	VehicleType  entity.VehicleType  `json:"vehicle_type"`
	Brand        string              `json:"brand"`
	Model        string              `json:"model"`
	Year         int                 `json:"year"`
	MaxWeight    float64             `json:"max_weight"`
	MaxVolume    float64             `json:"max_volume"`
	FuelType     string              `json:"fuel_type"`
	DriverID     *uuid.UUID          `json:"driver_id,omitempty"`
	Notes        *string             `json:"notes,omitempty"`
}

// UpdateVehicleRequest represents the request to update a vehicle
type UpdateVehicleRequest struct {
	ID        uuid.UUID           `json:"id"`
	Brand     *string             `json:"brand,omitempty"`
	Model     *string             `json:"model,omitempty"`
	Year      *int                `json:"year,omitempty"`
	MaxWeight *float64            `json:"max_weight,omitempty"`
	MaxVolume *float64            `json:"max_volume,omitempty"`
	FuelType  *string             `json:"fuel_type,omitempty"`
	Notes     *string             `json:"notes,omitempty"`
}

// CreateVehicle creates a new vehicle
func (uc *VehicleUseCase) CreateVehicle(ctx context.Context, req CreateVehicleRequest) (*entity.DeliveryVehicle, error) {
	// Create vehicle entity
	vehicle := &entity.DeliveryVehicle{
		ID:           uuid.New(),
		LicensePlate: req.LicensePlate,
		VehicleType:  req.VehicleType,
		Brand:        req.Brand,
		Model:        req.Model,
		Year:         req.Year,
		MaxWeight:    req.MaxWeight,
		MaxVolume:    req.MaxVolume,
		FuelType:     req.FuelType,
		DriverID:     req.DriverID,
		Status:       entity.VehicleStatusActive,
		Notes:        req.Notes,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	// Save vehicle
	if err := uc.vehicleRepo.Create(ctx, vehicle); err != nil {
		return nil, fmt.Errorf("failed to save vehicle: %w", err)
	}
	
	// Publish event
	event := map[string]interface{}{
		"event_type":    "vehicle_created",
		"vehicle_id":    vehicle.ID.String(),
		"license_plate": vehicle.LicensePlate,
		"vehicle_type":  string(vehicle.VehicleType),
		"created_at":    vehicle.CreatedAt,
	}
	
	if err := uc.eventPublisher.Publish(ctx, "vehicle.created", event); err != nil {
		// Don't fail the operation for event publishing errors, just log
		fmt.Printf("Failed to publish vehicle created event: %v\n", err)
	}
	
	return vehicle, nil
}

// GetVehicle retrieves a vehicle by ID
func (uc *VehicleUseCase) GetVehicle(ctx context.Context, vehicleID uuid.UUID) (*entity.DeliveryVehicle, error) {
	vehicle, err := uc.vehicleRepo.GetByID(ctx, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get vehicle: %w", err)
	}
	
	return vehicle, nil
}

// GetVehicleByCode retrieves a vehicle by code (license plate)
func (uc *VehicleUseCase) GetVehicleByCode(ctx context.Context, vehicleCode string) (*entity.DeliveryVehicle, error) {
	vehicle, err := uc.vehicleRepo.GetByCode(ctx, vehicleCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get vehicle by code: %w", err)
	}
	
	return vehicle, nil
}

// GetAvailableVehicles retrieves all available vehicles
func (uc *VehicleUseCase) GetAvailableVehicles(ctx context.Context) ([]*entity.DeliveryVehicle, error) {
	vehicles, err := uc.vehicleRepo.GetAvailable(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get available vehicles: %w", err)
	}
	
	return vehicles, nil
}

// GetVehiclesByType retrieves vehicles by type
func (uc *VehicleUseCase) GetVehiclesByType(ctx context.Context, vehicleType entity.VehicleType) ([]*entity.DeliveryVehicle, error) {
	vehicles, err := uc.vehicleRepo.GetByType(ctx, vehicleType)
	if err != nil {
		return nil, fmt.Errorf("failed to get vehicles by type: %w", err)
	}
	
	return vehicles, nil
}

// GetVehiclesByStatus retrieves vehicles by status
func (uc *VehicleUseCase) GetVehiclesByStatus(ctx context.Context, status entity.VehicleStatus) ([]*entity.DeliveryVehicle, error) {
	vehicles, err := uc.vehicleRepo.GetByStatus(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get vehicles by status: %w", err)
	}
	
	return vehicles, nil
}

// UpdateVehicle updates an existing vehicle
func (uc *VehicleUseCase) UpdateVehicle(ctx context.Context, req UpdateVehicleRequest) (*entity.DeliveryVehicle, error) {
	// Get existing vehicle
	vehicle, err := uc.vehicleRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get vehicle for update: %w", err)
	}
	
	// Update fields if provided
	if req.Brand != nil {
		vehicle.Brand = *req.Brand
	}
	if req.Model != nil {
		vehicle.Model = *req.Model
	}
	if req.Year != nil {
		vehicle.Year = *req.Year
	}
	if req.MaxWeight != nil {
		vehicle.MaxWeight = *req.MaxWeight
	}
	if req.MaxVolume != nil {
		vehicle.MaxVolume = *req.MaxVolume
	}
	if req.FuelType != nil {
		vehicle.FuelType = *req.FuelType
	}
	if req.Notes != nil {
		vehicle.Notes = req.Notes
	}
	
	vehicle.UpdatedAt = time.Now()
	
	// Save updated vehicle
	if err := uc.vehicleRepo.Update(ctx, vehicle); err != nil {
		return nil, fmt.Errorf("failed to save vehicle: %w", err)
	}
	
	// Publish event
	event := map[string]interface{}{
		"event_type": "vehicle_updated",
		"vehicle_id": vehicle.ID.String(),
		"updated_at": vehicle.UpdatedAt,
	}
	
	if err := uc.eventPublisher.Publish(ctx, "vehicle.updated", event); err != nil {
		fmt.Printf("Failed to publish vehicle updated event: %v\n", err)
	}
	
	return vehicle, nil
}

// SetVehicleStatus sets the status of a vehicle
func (uc *VehicleUseCase) SetVehicleStatus(ctx context.Context, vehicleID uuid.UUID, status entity.VehicleStatus) error {
	// Get vehicle
	vehicle, err := uc.vehicleRepo.GetByID(ctx, vehicleID)
	if err != nil {
		return fmt.Errorf("failed to get vehicle for status update: %w", err)
	}
	
	oldStatus := vehicle.Status
	vehicle.Status = status
	vehicle.UpdatedAt = time.Now()
	
	// Update status using repository method
	if err := uc.vehicleRepo.UpdateStatus(ctx, vehicleID, status); err != nil {
		return fmt.Errorf("failed to update vehicle status: %w", err)
	}
	
	// Publish event
	event := map[string]interface{}{
		"event_type":  "vehicle_status_changed",
		"vehicle_id":  vehicle.ID.String(),
		"old_status":  string(oldStatus),
		"new_status":  string(status),
		"updated_at":  time.Now(),
	}
	
	if err := uc.eventPublisher.Publish(ctx, "vehicle.status_changed", event); err != nil {
		fmt.Printf("Failed to publish vehicle status changed event: %v\n", err)
	}
	
	return nil
}

// AssignDriver assigns a driver to a vehicle
func (uc *VehicleUseCase) AssignDriver(ctx context.Context, vehicleID, driverID uuid.UUID) error {
	// Check if vehicle exists and is available
	vehicle, err := uc.vehicleRepo.GetByID(ctx, vehicleID)
	if err != nil {
		return fmt.Errorf("failed to get vehicle for driver assignment: %w", err)
	}
	
	if vehicle.Status != entity.VehicleStatusActive {
		return errors.New("vehicle is not available for driver assignment")
	}
	
	// Assign driver using repository method
	if err := uc.vehicleRepo.AssignDriver(ctx, vehicleID, driverID); err != nil {
		return fmt.Errorf("failed to assign driver: %w", err)
	}
	
	// Publish event
	event := map[string]interface{}{
		"event_type":   "driver_assigned",
		"vehicle_id":   vehicleID.String(),
		"driver_id":    driverID.String(),
		"assigned_at":  time.Now(),
	}
	
	if err := uc.eventPublisher.Publish(ctx, "vehicle.driver_assigned", event); err != nil {
		fmt.Printf("Failed to publish driver assigned event: %v\n", err)
	}
	
	return nil
}

// UnassignDriver removes a driver from a vehicle
func (uc *VehicleUseCase) UnassignDriver(ctx context.Context, vehicleID uuid.UUID) error {
	// Get vehicle to check current driver
	vehicle, err := uc.vehicleRepo.GetByID(ctx, vehicleID)
	if err != nil {
		return fmt.Errorf("failed to get vehicle for driver unassignment: %w", err)
	}
	
	if vehicle.DriverID == nil {
		return errors.New("vehicle has no assigned driver")
	}
	
	driverID := *vehicle.DriverID
	
	// Unassign driver using repository method
	if err := uc.vehicleRepo.UnassignDriver(ctx, vehicleID); err != nil {
		return fmt.Errorf("failed to unassign driver: %w", err)
	}
	
	// Publish event
	event := map[string]interface{}{
		"event_type":     "driver_unassigned",
		"vehicle_id":     vehicleID.String(),
		"driver_id":      driverID.String(),
		"unassigned_at":  time.Now(),
	}
	
	if err := uc.eventPublisher.Publish(ctx, "vehicle.driver_unassigned", event); err != nil {
		fmt.Printf("Failed to publish driver unassigned event: %v\n", err)
	}
	
	return nil
}

// SetMaintenanceMode sets a vehicle to maintenance mode
func (uc *VehicleUseCase) SetMaintenanceMode(ctx context.Context, vehicleID uuid.UUID, enable bool) error {
	if err := uc.vehicleRepo.SetMaintenanceMode(ctx, vehicleID, enable); err != nil {
		return fmt.Errorf("failed to set maintenance mode: %w", err)
	}
	
	// Update status based on maintenance mode
	status := entity.VehicleStatusActive
	if enable {
		status = entity.VehicleStatusMaintenance
	}
	
	if err := uc.vehicleRepo.UpdateStatus(ctx, vehicleID, status); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	
	// Publish event
	event := map[string]interface{}{
		"event_type":        "maintenance_mode_changed",
		"vehicle_id":        vehicleID.String(),
		"maintenance_mode":  enable,
		"status":           string(status),
		"updated_at":       time.Now(),
	}
	
	if err := uc.eventPublisher.Publish(ctx, "vehicle.maintenance_mode_changed", event); err != nil {
		fmt.Printf("Failed to publish maintenance mode changed event: %v\n", err)
	}
	
	return nil
}

// UpdateLocation updates the current location of a vehicle
func (uc *VehicleUseCase) UpdateLocation(ctx context.Context, vehicleID uuid.UUID, latitude, longitude float64) error {
	if err := uc.vehicleRepo.UpdateLocation(ctx, vehicleID, latitude, longitude); err != nil {
		return fmt.Errorf("failed to update vehicle location: %w", err)
	}
	
	// Publish event
	event := map[string]interface{}{
		"event_type": "vehicle_location_updated",
		"vehicle_id": vehicleID.String(),
		"latitude":   latitude,
		"longitude":  longitude,
		"updated_at": time.Now(),
	}
	
	if err := uc.eventPublisher.Publish(ctx, "vehicle.location_updated", event); err != nil {
		fmt.Printf("Failed to publish vehicle location updated event: %v\n", err)
	}
	
	return nil
}

// DeleteVehicle soft deletes a vehicle
func (uc *VehicleUseCase) DeleteVehicle(ctx context.Context, vehicleID uuid.UUID) error {
	// Get vehicle to check if it can be deleted
	vehicle, err := uc.vehicleRepo.GetByID(ctx, vehicleID)
	if err != nil {
		return fmt.Errorf("failed to get vehicle for deletion: %w", err)
	}
	
	if vehicle.Status == entity.VehicleStatusOnRoute {
		return errors.New("cannot delete vehicle that is currently on route")
	}
	
	// Soft delete vehicle
	if err := uc.vehicleRepo.Delete(ctx, vehicleID); err != nil {
		return fmt.Errorf("failed to delete vehicle: %w", err)
	}
	
	// Publish event
	event := map[string]interface{}{
		"event_type":    "vehicle_deleted",
		"vehicle_id":    vehicle.ID.String(),
		"license_plate": vehicle.LicensePlate,
		"deleted_at":    time.Now(),
	}
	
	if err := uc.eventPublisher.Publish(ctx, "vehicle.deleted", event); err != nil {
		fmt.Printf("Failed to publish vehicle deleted event: %v\n", err)
	}
	
	return nil
}

// GetVehicleMetrics retrieves performance metrics for a vehicle
func (uc *VehicleUseCase) GetVehicleMetrics(ctx context.Context, vehicleID uuid.UUID, startDate, endDate time.Time) (*repository.VehicleMetrics, error) {
	metrics, err := uc.vehicleRepo.GetVehicleMetrics(ctx, vehicleID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get vehicle metrics: %w", err)
	}
	
	return metrics, nil
}

// GetFleetUtilization retrieves utilization metrics for all vehicles
func (uc *VehicleUseCase) GetFleetUtilization(ctx context.Context, date time.Time) (map[uuid.UUID]float64, error) {
	utilization, err := uc.vehicleRepo.GetFleetUtilization(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get fleet utilization: %w", err)
	}
	
	return utilization, nil
}
