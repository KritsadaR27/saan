package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"shipping/internal/domain/entity"
)

// CoverageAreaRepository defines the contract for coverage area data persistence
type CoverageAreaRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, area *entity.CoverageArea) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.CoverageArea, error)
	Update(ctx context.Context, area *entity.CoverageArea) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Query operations
	GetAll(ctx context.Context, limit, offset int) ([]*entity.CoverageArea, error)
	GetActive(ctx context.Context) ([]*entity.CoverageArea, error)
	GetByProvince(ctx context.Context, province string) ([]*entity.CoverageArea, error)
	GetSelfDeliveryAreas(ctx context.Context) ([]*entity.CoverageArea, error)
	GetThirdPartyAreas(ctx context.Context) ([]*entity.CoverageArea, error)
	
	// Location matching
	FindByLocation(ctx context.Context, province, district, subdistrict, postalCode string) ([]*entity.CoverageArea, error)
	GetBestMatchForLocation(ctx context.Context, province, district, subdistrict, postalCode string) (*entity.CoverageArea, error)
	GetByPostalCode(ctx context.Context, postalCode string) ([]*entity.CoverageArea, error)
	
	// Route and zone operations
	GetByRoute(ctx context.Context, route string) ([]*entity.CoverageArea, error)
	GetByZone(ctx context.Context, zone string) ([]*entity.CoverageArea, error)
	GetAvailableRoutes(ctx context.Context) ([]string, error)
	GetAvailableZones(ctx context.Context) ([]string, error)
	
	// Configuration management
	GetByPriorityOrder(ctx context.Context) ([]*entity.CoverageArea, error)
	UpdatePriority(ctx context.Context, id uuid.UUID, priority int) error
	ActivateArea(ctx context.Context, id uuid.UUID) error
	DeactivateArea(ctx context.Context, id uuid.UUID) error
	
	// Pricing operations
	UpdatePricing(ctx context.Context, id uuid.UUID, baseDeliveryFee, perKmRate, freeDeliveryThreshold decimal.Decimal) error
	GetAreasWithFreeDelivery(ctx context.Context, orderValue decimal.Decimal) ([]*entity.CoverageArea, error)
	GetAreasByPriceRange(ctx context.Context, minFee, maxFee decimal.Decimal) ([]*entity.CoverageArea, error)
	
	// Service level operations
	UpdateServiceLevels(ctx context.Context, id uuid.UUID, standardHours, expressHours int, sameDayAvailable bool) error
	GetAreasBySameDaySupport(ctx context.Context, supported bool) ([]*entity.CoverageArea, error)
	GetAreasByServiceLevel(ctx context.Context, serviceLevel string) ([]*entity.CoverageArea, error)
	
	// Capacity management
	UpdateCapacity(ctx context.Context, id uuid.UUID, maxDailyCapacity int) error
	GetAreasByCapacityRange(ctx context.Context, minCapacity, maxCapacity int) ([]*entity.CoverageArea, error)
	GetAreasWithAvailableCapacity(ctx context.Context, requiredCapacity int) ([]*entity.CoverageArea, error)
	
	// Auto assignment
	EnableAutoAssign(ctx context.Context, id uuid.UUID) error
	DisableAutoAssign(ctx context.Context, id uuid.UUID) error
	GetAutoAssignAreas(ctx context.Context) ([]*entity.CoverageArea, error)
	
	// Search and filtering
	SearchAreas(ctx context.Context, filters *CoverageAreaQueryFilters) ([]*entity.CoverageArea, error)
	GetAreasByPattern(ctx context.Context, locationPattern string) ([]*entity.CoverageArea, error)
	
	// Bulk operations
	UpdateMultipleAreaStatuses(ctx context.Context, areaIDs []uuid.UUID, isActive bool) error
	GetAreasByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.CoverageArea, error)
	CreateBulkAreas(ctx context.Context, areas []*entity.CoverageArea) error
	
	// Analytics and reporting
	GetCoverageStats(ctx context.Context) (*CoverageStats, error)
	GetProvinceCoverage(ctx context.Context) (map[string]*ProvinceCoverage, error)
	GetUnservicedAreas(ctx context.Context) ([]string, error)
}

// CoverageAreaQueryFilters represents filters for coverage area queries
type CoverageAreaQueryFilters struct {
	Province           *string          `json:"province,omitempty"`
	District           *string          `json:"district,omitempty"`
	Subdistrict        *string          `json:"subdistrict,omitempty"`
	PostalCode         *string          `json:"postal_code,omitempty"`
	IsSelfDeliveryArea *bool            `json:"is_self_delivery_area,omitempty"`
	DeliveryRoute      *string          `json:"delivery_route,omitempty"`
	DeliveryZone       *string          `json:"delivery_zone,omitempty"`
	IsActive           *bool            `json:"is_active,omitempty"`
	AutoAssign         *bool            `json:"auto_assign,omitempty"`
	SameDayAvailable   *bool            `json:"same_day_available,omitempty"`
	MinPriority        *int             `json:"min_priority,omitempty"`
	MaxPriority        *int             `json:"max_priority,omitempty"`
	MinBaseFee         *decimal.Decimal `json:"min_base_fee,omitempty"`
	MaxBaseFee         *decimal.Decimal `json:"max_base_fee,omitempty"`
	MinCapacity        *int             `json:"min_capacity,omitempty"`
	MaxCapacity        *int             `json:"max_capacity,omitempty"`
	Limit              int              `json:"limit"`
	Offset             int              `json:"offset"`
}

// CoverageStats represents coverage statistics
type CoverageStats struct {
	TotalAreas              int64   `json:"total_areas"`
	ActiveAreas             int64   `json:"active_areas"`
	SelfDeliveryAreas       int64   `json:"self_delivery_areas"`
	ThirdPartyAreas         int64   `json:"third_party_areas"`
	SameDayAreas            int64   `json:"same_day_areas"`
	AutoAssignAreas         int64   `json:"auto_assign_areas"`
	TotalProvincesCovered   int64   `json:"total_provinces_covered"`
	AverageBaseFee          float64 `json:"average_base_fee"`
	AveragePerKmRate        float64 `json:"average_per_km_rate"`
	AverageFreeThreshold    float64 `json:"average_free_threshold"`
	AverageStandardHours    float64 `json:"average_standard_hours"`
	AverageExpressHours     float64 `json:"average_express_hours"`
	TotalCapacity           int64   `json:"total_capacity"`
	AverageCapacity         float64 `json:"average_capacity"`
}

// ProvinceCoverage represents coverage information for a province
type ProvinceCoverage struct {
	Province            string                   `json:"province"`
	TotalAreas          int64                    `json:"total_areas"`
	SelfDeliveryAreas   int64                    `json:"self_delivery_areas"`
	ThirdPartyAreas     int64                    `json:"third_party_areas"`
	ActiveAreas         int64                    `json:"active_areas"`
	SameDaySupport      bool                     `json:"same_day_support"`
	Routes              []string                 `json:"routes"`
	Zones               []string                 `json:"zones"`
	AverageBaseFee      float64                  `json:"average_base_fee"`
	MinBaseFee          float64                  `json:"min_base_fee"`
	MaxBaseFee          float64                  `json:"max_base_fee"`
	TotalCapacity       int64                    `json:"total_capacity"`
	DistrictCoverage    map[string]int64         `json:"district_coverage"`
	ServiceLevels       map[string]bool          `json:"service_levels"`
}

// DeliveryOption represents a delivery option for a location
type DeliveryOption struct {
	CoverageAreaID      uuid.UUID       `json:"coverage_area_id"`
	LocationString      string          `json:"location_string"`
	IsSelfDelivery      bool            `json:"is_self_delivery"`
	DeliveryRoute       string          `json:"delivery_route,omitempty"`
	DeliveryZone        string          `json:"delivery_zone,omitempty"`
	BaseDeliveryFee     decimal.Decimal `json:"base_delivery_fee"`
	PerKmRate           decimal.Decimal `json:"per_km_rate"`
	FreeDeliveryThreshold decimal.Decimal `json:"free_delivery_threshold"`
	StandardDeliveryHours int           `json:"standard_delivery_hours"`
	ExpressDeliveryHours  int           `json:"express_delivery_hours"`
	SameDayAvailable    bool            `json:"same_day_available"`
	MaxDailyCapacity    int             `json:"max_daily_capacity"`
	PriorityOrder       int             `json:"priority_order"`
	IsActive            bool            `json:"is_active"`
	AutoAssign          bool            `json:"auto_assign"`
}
