package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"shipping/internal/domain/entity"
)

// RouteRepository defines the contract for route data persistence
type RouteRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, route *entity.DeliveryRoute) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.DeliveryRoute, error)
	GetByName(ctx context.Context, routeName string, routeDate time.Time) (*entity.DeliveryRoute, error)
	Update(ctx context.Context, route *entity.DeliveryRoute) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Query operations
	GetAll(ctx context.Context, limit, offset int) ([]*entity.DeliveryRoute, error)
	GetByDate(ctx context.Context, date time.Time) ([]*entity.DeliveryRoute, error)
	GetByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*entity.DeliveryRoute, error)
	GetByVehicle(ctx context.Context, vehicleID uuid.UUID, date time.Time) ([]*entity.DeliveryRoute, error)
	GetByDriver(ctx context.Context, driverID uuid.UUID, date time.Time) ([]*entity.DeliveryRoute, error)
	GetByStatus(ctx context.Context, status entity.RouteStatus) ([]*entity.DeliveryRoute, error)
	
	// Route planning and optimization
	GetActiveRoutes(ctx context.Context) ([]*entity.DeliveryRoute, error)
	GetPlannedRoutes(ctx context.Context, date time.Time) ([]*entity.DeliveryRoute, error)
	GetInProgressRoutes(ctx context.Context) ([]*entity.DeliveryRoute, error)
	GetCompletedRoutes(ctx context.Context, date time.Time) ([]*entity.DeliveryRoute, error)
	
	// Route assignment
	AssignVehicle(ctx context.Context, routeID, vehicleID uuid.UUID, driverID *uuid.UUID) error
	UnassignVehicle(ctx context.Context, routeID uuid.UUID) error
	
	// Route execution
	StartRoute(ctx context.Context, routeID uuid.UUID) error
	CompleteRoute(ctx context.Context, routeID uuid.UUID, actualDistance float64, actualOrdersDelivered int) error
	CancelRoute(ctx context.Context, routeID uuid.UUID) error
	
	// Route optimization
	UpdateOptimizationData(ctx context.Context, routeID uuid.UUID, data map[string]interface{}) error
	GetOptimizedRoutes(ctx context.Context, date time.Time, maxDistance float64) ([]*entity.DeliveryRoute, error)
	
	// Planning operations
	SetPlanning(ctx context.Context, routeID uuid.UUID, startTime, endTime time.Time, distance float64, orderCount int) error
	UpdatePlanning(ctx context.Context, routeID uuid.UUID, updates map[string]interface{}) error
	
	// Analytics and reporting
	GetRouteMetrics(ctx context.Context, routeID uuid.UUID) (*RouteMetrics, error)
	GetRoutesEfficiency(ctx context.Context, startDate, endDate time.Time) (map[uuid.UUID]float64, error)
	GetAverageRouteTime(ctx context.Context, startDate, endDate time.Time) (float64, error)
	GetRouteUtilization(ctx context.Context, date time.Time) (float64, error)
	
	// Search and filtering
	SearchRoutes(ctx context.Context, filters *RouteQueryFilters) ([]*entity.DeliveryRoute, error)
	GetRoutesByPattern(ctx context.Context, namePattern string) ([]*entity.DeliveryRoute, error)
	GetUnassignedRoutes(ctx context.Context, date time.Time) ([]*entity.DeliveryRoute, error)
	
	// Bulk operations
	CreateMultipleRoutes(ctx context.Context, routes []*entity.DeliveryRoute) error
	UpdateMultipleRouteStatuses(ctx context.Context, routeIDs []uuid.UUID, status entity.RouteStatus) error
	DeleteRoutesByDate(ctx context.Context, date time.Time) error
}

// RouteQueryFilters represents filters for route queries
type RouteQueryFilters struct {
	RouteDate     *time.Time            `json:"route_date,omitempty"`
	StartDate     *time.Time            `json:"start_date,omitempty"`
	EndDate       *time.Time            `json:"end_date,omitempty"`
	Status        *entity.RouteStatus   `json:"status,omitempty"`
	VehicleID     *uuid.UUID            `json:"vehicle_id,omitempty"`
	DriverID      *uuid.UUID            `json:"driver_id,omitempty"`
	RouteName     *string               `json:"route_name,omitempty"`
	MinDistance   *float64              `json:"min_distance,omitempty"`
	MaxDistance   *float64              `json:"max_distance,omitempty"`
	MinOrders     *int                  `json:"min_orders,omitempty"`
	MaxOrders     *int                  `json:"max_orders,omitempty"`
	Limit         int                   `json:"limit"`
	Offset        int                   `json:"offset"`
}

// RouteMetrics represents route performance metrics
type RouteMetrics struct {
	RouteID              uuid.UUID     `json:"route_id"`
	RouteName            string        `json:"route_name"`
	TotalRoutes          int64         `json:"total_routes"`
	CompletedRoutes      int64         `json:"completed_routes"`
	CancelledRoutes      int64         `json:"cancelled_routes"`
	AverageDistance      float64       `json:"average_distance_km"`
	AverageDuration      float64       `json:"average_duration_hours"`
	AverageOrdersPerRoute int64        `json:"average_orders_per_route"`
	EfficiencyRate       float64       `json:"efficiency_rate_percentage"`
	OnTimeRate           float64       `json:"on_time_rate_percentage"`
	TotalDeliveries      int64         `json:"total_deliveries"`
	SuccessfulDeliveries int64         `json:"successful_deliveries"`
	PeriodStart          time.Time     `json:"period_start"`
	PeriodEnd            time.Time     `json:"period_end"`
}

// RouteOptimization represents route optimization data
type RouteOptimization struct {
	RouteID           uuid.UUID                `json:"route_id"`
	OptimizedSequence []RouteStop              `json:"optimized_sequence"`
	TotalDistance     float64                  `json:"total_distance_km"`
	EstimatedDuration time.Duration            `json:"estimated_duration"`
	OptimizationAlgorithm string               `json:"optimization_algorithm"`
	CreatedAt         time.Time                `json:"created_at"`
	Parameters        map[string]interface{}   `json:"parameters"`
}

// RouteStop represents a stop in an optimized route
type RouteStop struct {
	DeliveryID        uuid.UUID   `json:"delivery_id"`
	Sequence          int         `json:"sequence"`
	Address           string      `json:"address"`
	Coordinates       Coordinates `json:"coordinates"`
	EstimatedArrival  time.Time   `json:"estimated_arrival"`
	EstimatedDuration int         `json:"estimated_duration_minutes"`
	StopType          string      `json:"stop_type"` // pickup, delivery
}

// Coordinates represents geographic coordinates
type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
