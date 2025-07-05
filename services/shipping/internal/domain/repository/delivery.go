package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"shipping/internal/domain/entity"
)

// DeliveryRepository defines the contract for delivery data persistence
type DeliveryRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, delivery *entity.DeliveryOrder) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.DeliveryOrder, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) (*entity.DeliveryOrder, error)
	GetByTrackingNumber(ctx context.Context, trackingNumber string) (*entity.DeliveryOrder, error)
	Update(ctx context.Context, delivery *entity.DeliveryOrder) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Query operations
	GetByCustomerID(ctx context.Context, customerID uuid.UUID, limit, offset int) ([]*entity.DeliveryOrder, error)
	GetByStatus(ctx context.Context, status entity.DeliveryStatus, limit, offset int) ([]*entity.DeliveryOrder, error)
	GetByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*entity.DeliveryOrder, error)
	GetByVehicleID(ctx context.Context, vehicleID uuid.UUID, date time.Time) ([]*entity.DeliveryOrder, error)
	GetByRouteID(ctx context.Context, routeID uuid.UUID) ([]*entity.DeliveryOrder, error)
	
	// Delivery method specific queries
	GetByDeliveryMethod(ctx context.Context, method entity.DeliveryMethod, limit, offset int) ([]*entity.DeliveryOrder, error)
	GetManualCoordinationDeliveries(ctx context.Context, limit, offset int) ([]*entity.DeliveryOrder, error)
	GetPendingDeliveries(ctx context.Context, limit, offset int) ([]*entity.DeliveryOrder, error)
	
	// Inter Express specific queries
	GetPendingInterExpressOrders(ctx context.Context, date time.Time) ([]*entity.DeliveryOrder, error)
	GetInterExpressByDate(ctx context.Context, date time.Time) ([]*entity.DeliveryOrder, error)
	
	// Analytics and reporting
	GetDeliveryCountByStatus(ctx context.Context, startDate, endDate time.Time) (map[entity.DeliveryStatus]int64, error)
	GetDeliveryCountByMethod(ctx context.Context, startDate, endDate time.Time) (map[entity.DeliveryMethod]int64, error)
	GetAverageDeliveryTime(ctx context.Context, method entity.DeliveryMethod, startDate, endDate time.Time) (float64, error)
	GetTotalDeliveryFees(ctx context.Context, startDate, endDate time.Time) (float64, error)
	
	// Status updates
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.DeliveryStatus) error
	UpdateTrackingInfo(ctx context.Context, id uuid.UUID, trackingNumber string, providerOrderID string) error
	UpdateActualTimes(ctx context.Context, id uuid.UUID, pickupTime, deliveryTime *time.Time) error
	
	// Vehicle and route assignment
	AssignVehicle(ctx context.Context, id uuid.UUID, vehicleID uuid.UUID) error
	AssignRoute(ctx context.Context, id uuid.UUID, routeID uuid.UUID) error
	
	// Bulk operations
	UpdateMultipleStatuses(ctx context.Context, ids []uuid.UUID, status entity.DeliveryStatus) error
	GetDeliveriesByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.DeliveryOrder, error)
	
	// Search and filtering
	SearchDeliveries(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*entity.DeliveryOrder, error)
	GetActiveDeliveries(ctx context.Context, limit, offset int) ([]*entity.DeliveryOrder, error)
	GetOverdueDeliveries(ctx context.Context) ([]*entity.DeliveryOrder, error)
}

// DeliveryQueryFilters represents filters for delivery queries
type DeliveryQueryFilters struct {
	CustomerID      *uuid.UUID             `json:"customer_id,omitempty"`
	Status          *entity.DeliveryStatus `json:"status,omitempty"`
	DeliveryMethod  *entity.DeliveryMethod `json:"delivery_method,omitempty"`
	VehicleID       *uuid.UUID             `json:"vehicle_id,omitempty"`
	RouteID         *uuid.UUID             `json:"route_id,omitempty"`
	StartDate       *time.Time             `json:"start_date,omitempty"`
	EndDate         *time.Time             `json:"end_date,omitempty"`
	TrackingNumber  *string                `json:"tracking_number,omitempty"`
	RequiresManual  *bool                  `json:"requires_manual,omitempty"`
	IsActive        *bool                  `json:"is_active,omitempty"`
}

// DeliveryMetrics represents delivery performance metrics
type DeliveryMetrics struct {
	TotalDeliveries       int64                                    `json:"total_deliveries"`
	DeliveriesByStatus    map[entity.DeliveryStatus]int64          `json:"deliveries_by_status"`
	DeliveriesByMethod    map[entity.DeliveryMethod]int64          `json:"deliveries_by_method"`
	AverageDeliveryTime   float64                                  `json:"average_delivery_time_hours"`
	TotalRevenue          float64                                  `json:"total_revenue"`
	SuccessRate           float64                                  `json:"success_rate_percentage"`
	OnTimeDeliveryRate    float64                                  `json:"on_time_delivery_rate"`
}
