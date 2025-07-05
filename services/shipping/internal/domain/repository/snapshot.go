package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"shipping/internal/domain/entity"
)

// SnapshotRepository defines the contract for delivery snapshot data persistence
type SnapshotRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, snapshot *entity.DeliverySnapshot) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.DeliverySnapshot, error)
	Update(ctx context.Context, snapshot *entity.DeliverySnapshot) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Query operations by delivery
	GetByDeliveryID(ctx context.Context, deliveryID uuid.UUID) ([]*entity.DeliverySnapshot, error)
	GetLatestByDeliveryID(ctx context.Context, deliveryID uuid.UUID) (*entity.DeliverySnapshot, error)
	GetByDeliveryIDAndType(ctx context.Context, deliveryID uuid.UUID, snapshotType entity.SnapshotType) ([]*entity.DeliverySnapshot, error)
	
	// Query operations by filters
	GetByType(ctx context.Context, snapshotType entity.SnapshotType, limit, offset int) ([]*entity.DeliverySnapshot, error)
	GetByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*entity.DeliverySnapshot, error)
	GetByBusinessDate(ctx context.Context, businessDate time.Time) ([]*entity.DeliverySnapshot, error)
	GetByCustomerID(ctx context.Context, customerID uuid.UUID, limit, offset int) ([]*entity.DeliverySnapshot, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*entity.DeliverySnapshot, error)
	
	// Provider-specific queries
	GetByProviderCode(ctx context.Context, providerCode string, startDate, endDate time.Time) ([]*entity.DeliverySnapshot, error)
	GetByProviderAndStatus(ctx context.Context, providerCode, status string, limit, offset int) ([]*entity.DeliverySnapshot, error)
	
	// Vehicle and route queries
	GetByVehicleID(ctx context.Context, vehicleID uuid.UUID, businessDate time.Time) ([]*entity.DeliverySnapshot, error)
	GetByProvince(ctx context.Context, province string, businessDate time.Time) ([]*entity.DeliverySnapshot, error)
	
	// Timeline and audit queries
	GetDeliveryTimeline(ctx context.Context, deliveryID uuid.UUID) ([]*entity.DeliverySnapshot, error)
	GetSnapshotChain(ctx context.Context, snapshotID uuid.UUID) ([]*entity.DeliverySnapshot, error)
	GetBusinessEventSnapshots(ctx context.Context, deliveryID uuid.UUID) ([]*entity.DeliverySnapshot, error)
	
	// Analytics and reporting
	GetSnapshotCountByType(ctx context.Context, startDate, endDate time.Time) (map[entity.SnapshotType]int64, error)
	GetSnapshotCountByProvider(ctx context.Context, businessDate time.Time) (map[string]int64, error)
	GetSnapshotCountByStatus(ctx context.Context, businessDate time.Time) (map[string]int64, error)
	GetDeliveryCompletionRate(ctx context.Context, startDate, endDate time.Time) (float64, error)
	
	// Performance metrics from snapshots
	GetAverageDeliveryDurationFromSnapshots(ctx context.Context, providerCode string, startDate, endDate time.Time) (float64, error)
	GetProviderPerformanceFromSnapshots(ctx context.Context, providerCode string, startDate, endDate time.Time) (*ProviderPerformanceMetrics, error)
	GetDailyDeliveryMetricsFromSnapshots(ctx context.Context, businessDate time.Time) (*DailyDeliveryMetrics, error)
	
	// Audit and compliance
	GetSnapshotsForAudit(ctx context.Context, startDate, endDate time.Time) ([]*entity.DeliverySnapshot, error)
	GetSnapshotsByTriggeredBy(ctx context.Context, triggeredBy string, startDate, endDate time.Time) ([]*entity.DeliverySnapshot, error)
	GetSnapshotsByUser(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*entity.DeliverySnapshot, error)
	
	// Financial reporting from snapshots
	GetRevenueFromSnapshots(ctx context.Context, startDate, endDate time.Time) (map[string]float64, error)
	GetDeliveryFeesFromSnapshots(ctx context.Context, providerCode string, businessDate time.Time) (float64, error)
	GetMonthlyFinancialSummaryFromSnapshots(ctx context.Context, year int, month int) (*MonthlyFinancialSummary, error)
	
	// Search and filtering
	SearchSnapshots(ctx context.Context, filters *SnapshotQueryFilters) ([]*entity.DeliverySnapshot, error)
	GetFailedDeliverySnapshots(ctx context.Context, startDate, endDate time.Time) ([]*entity.DeliverySnapshot, error)
	GetSuccessfulDeliverySnapshots(ctx context.Context, startDate, endDate time.Time) ([]*entity.DeliverySnapshot, error)
	
	// Bulk operations
	CreateBulkSnapshots(ctx context.Context, snapshots []*entity.DeliverySnapshot) error
	GetSnapshotsByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.DeliverySnapshot, error)
	
	// Data retention and cleanup
	DeleteSnapshotsOlderThan(ctx context.Context, cutoffDate time.Time) (int64, error)
	ArchiveSnapshotsOlderThan(ctx context.Context, cutoffDate time.Time) (int64, error)
}

// SnapshotQueryFilters represents filters for snapshot queries
type SnapshotQueryFilters struct {
	DeliveryID        *uuid.UUID             `json:"delivery_id,omitempty"`
	SnapshotType      *entity.SnapshotType   `json:"snapshot_type,omitempty"`
	CustomerID        *uuid.UUID             `json:"customer_id,omitempty"`
	OrderID           *uuid.UUID             `json:"order_id,omitempty"`
	VehicleID         *uuid.UUID             `json:"vehicle_id,omitempty"`
	ProviderCode      *string                `json:"provider_code,omitempty"`
	DeliveryStatus    *string                `json:"delivery_status,omitempty"`
	Province          *string                `json:"province,omitempty"`
	TriggeredBy       *string                `json:"triggered_by,omitempty"`
	TriggeredByUserID *uuid.UUID             `json:"triggered_by_user_id,omitempty"`
	StartDate         *time.Time             `json:"start_date,omitempty"`
	EndDate           *time.Time             `json:"end_date,omitempty"`
	BusinessDate      *time.Time             `json:"business_date,omitempty"`
	MinDeliveryFee    *float64               `json:"min_delivery_fee,omitempty"`
	MaxDeliveryFee    *float64               `json:"max_delivery_fee,omitempty"`
	Limit             int                    `json:"limit"`
	Offset            int                    `json:"offset"`
}

// ProviderPerformanceMetrics represents provider performance metrics derived from snapshots
type ProviderPerformanceMetrics struct {
	ProviderCode         string    `json:"provider_code"`
	TotalDeliveries      int64     `json:"total_deliveries"`
	SuccessfulDeliveries int64     `json:"successful_deliveries"`
	FailedDeliveries     int64     `json:"failed_deliveries"`
	CancelledDeliveries  int64     `json:"cancelled_deliveries"`
	AverageDeliveryTime  float64   `json:"average_delivery_time_hours"`
	SuccessRate          float64   `json:"success_rate_percentage"`
	TotalRevenue         float64   `json:"total_revenue"`
	AverageDeliveryFee   float64   `json:"average_delivery_fee"`
	PeriodStart          time.Time `json:"period_start"`
	PeriodEnd            time.Time `json:"period_end"`
}

// DailyDeliveryMetrics represents daily delivery metrics derived from snapshots
type DailyDeliveryMetrics struct {
	BusinessDate         time.Time                   `json:"business_date"`
	TotalDeliveries      int64                       `json:"total_deliveries"`
	CompletedDeliveries  int64                       `json:"completed_deliveries"`
	FailedDeliveries     int64                       `json:"failed_deliveries"`
	CancelledDeliveries  int64                       `json:"cancelled_deliveries"`
	PendingDeliveries    int64                       `json:"pending_deliveries"`
	TotalRevenue         float64                     `json:"total_revenue"`
	AverageDeliveryFee   float64                     `json:"average_delivery_fee"`
	ProviderBreakdown    map[string]int64            `json:"provider_breakdown"`
	StatusBreakdown      map[string]int64            `json:"status_breakdown"`
	ProvinceBreakdown    map[string]int64            `json:"province_breakdown"`
	HourlyDistribution   map[string]int64            `json:"hourly_distribution"`
}

// MonthlyFinancialSummary represents monthly financial summary from snapshots
type MonthlyFinancialSummary struct {
	Year                 int                    `json:"year"`
	Month                int                    `json:"month"`
	TotalRevenue         float64                `json:"total_revenue"`
	TotalDeliveries      int64                  `json:"total_deliveries"`
	AverageOrderValue    float64                `json:"average_order_value"`
	RevenueByProvider    map[string]float64     `json:"revenue_by_provider"`
	DeliveriesByProvider map[string]int64       `json:"deliveries_by_provider"`
	DailyBreakdown       map[string]float64     `json:"daily_breakdown"`
	GrowthRate           float64                `json:"growth_rate_percentage"`
	TopProvinces         map[string]float64     `json:"top_provinces_revenue"`
}

// SnapshotAuditInfo represents audit information for snapshots
type SnapshotAuditInfo struct {
	SnapshotID       uuid.UUID              `json:"snapshot_id"`
	DeliveryID       uuid.UUID              `json:"delivery_id"`
	SnapshotType     entity.SnapshotType    `json:"snapshot_type"`
	TriggeredBy      string                 `json:"triggered_by"`
	TriggeredEvent   string                 `json:"triggered_event"`
	UserID           *uuid.UUID             `json:"user_id,omitempty"`
	Changes          map[string]interface{} `json:"changes"`
	CreatedAt        time.Time              `json:"created_at"`
	BusinessDate     time.Time              `json:"business_date"`
}
