package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"shipping/internal/domain/entity"
)

// VehicleRepository defines the contract for vehicle data persistence
type VehicleRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, vehicle *entity.DeliveryVehicle) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.DeliveryVehicle, error)
	GetByCode(ctx context.Context, vehicleCode string) (*entity.DeliveryVehicle, error)
	Update(ctx context.Context, vehicle *entity.DeliveryVehicle) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Query operations
	GetAll(ctx context.Context, limit, offset int) ([]*entity.DeliveryVehicle, error)
	GetByType(ctx context.Context, vehicleType entity.VehicleType) ([]*entity.DeliveryVehicle, error)
	GetByStatus(ctx context.Context, status entity.VehicleStatus) ([]*entity.DeliveryVehicle, error)
	GetAvailable(ctx context.Context) ([]*entity.DeliveryVehicle, error)
	GetByRoute(ctx context.Context, route string) ([]*entity.DeliveryVehicle, error)
	
	// Status management
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.VehicleStatus) error
	UpdateLocation(ctx context.Context, id uuid.UUID, latitude, longitude float64) error
	
	// Route assignment
	AssignRoute(ctx context.Context, id uuid.UUID, route string) error
	UnassignRoute(ctx context.Context, id uuid.UUID) error
	
	// Driver assignment
	AssignDriver(ctx context.Context, vehicleID, driverID uuid.UUID) error
	UnassignDriver(ctx context.Context, vehicleID uuid.UUID) error
	
	// Capacity management
	UpdateDailyCapacity(ctx context.Context, id uuid.UUID, capacity int) error
	GetVehicleUtilization(ctx context.Context, id uuid.UUID, date time.Time) (int, error)
	
	// Maintenance and operational
	SetMaintenanceMode(ctx context.Context, id uuid.UUID, enable bool) error
	UpdateOperatingCosts(ctx context.Context, id uuid.UUID, dailyCost, perKmCost float64) error
	
	// Analytics and reporting
	GetVehicleMetrics(ctx context.Context, id uuid.UUID, startDate, endDate time.Time) (*VehicleMetrics, error)
	GetFleetUtilization(ctx context.Context, date time.Time) (map[uuid.UUID]float64, error)
	GetVehiclesByCapacityRange(ctx context.Context, minCapacity, maxCapacity int) ([]*entity.DeliveryVehicle, error)
	
	// Search and filtering
	SearchVehicles(ctx context.Context, filters *VehicleQueryFilters) ([]*entity.DeliveryVehicle, error)
	GetActiveVehicles(ctx context.Context) ([]*entity.DeliveryVehicle, error)
	GetVehiclesNeedingMaintenance(ctx context.Context) ([]*entity.DeliveryVehicle, error)
}

// VehicleQueryFilters represents filters for vehicle queries
type VehicleQueryFilters struct {
	VehicleType   *entity.VehicleType   `json:"vehicle_type,omitempty"`
	Status        *entity.VehicleStatus `json:"status,omitempty"`
	Route         *string               `json:"route,omitempty"`
	DriverID      *uuid.UUID            `json:"driver_id,omitempty"`
	MinCapacity   *int                  `json:"min_capacity,omitempty"`
	MaxCapacity   *int                  `json:"max_capacity,omitempty"`
	MinWeight     *float64              `json:"min_weight,omitempty"`
	MaxWeight     *float64              `json:"max_weight,omitempty"`
	IsActive      *bool                 `json:"is_active,omitempty"`
	HomeBase      *string               `json:"home_base,omitempty"`
	Limit         int                   `json:"limit"`
	Offset        int                   `json:"offset"`
}

// VehicleMetrics represents vehicle performance metrics
type VehicleMetrics struct {
	VehicleID           uuid.UUID `json:"vehicle_id"`
	TotalDeliveries     int64     `json:"total_deliveries"`
	CompletedDeliveries int64     `json:"completed_deliveries"`
	FailedDeliveries    int64     `json:"failed_deliveries"`
	TotalDistance       float64   `json:"total_distance_km"`
	TotalDuration       float64   `json:"total_duration_hours"`
	UtilizationRate     float64   `json:"utilization_rate_percentage"`
	AverageDeliveryTime float64   `json:"average_delivery_time_minutes"`
	FuelCost            float64   `json:"fuel_cost"`
	MaintenanceCost     float64   `json:"maintenance_cost"`
	TotalRevenue        float64   `json:"total_revenue"`
}
