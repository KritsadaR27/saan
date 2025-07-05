package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"shipping/internal/domain/entity"
)

// ProviderRepository defines the contract for delivery provider data persistence
type ProviderRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, provider *entity.DeliveryProvider) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.DeliveryProvider, error)
	GetByCode(ctx context.Context, providerCode string) (*entity.DeliveryProvider, error)
	Update(ctx context.Context, provider *entity.DeliveryProvider) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Query operations
	GetAll(ctx context.Context, limit, offset int) ([]*entity.DeliveryProvider, error)
	GetActive(ctx context.Context) ([]*entity.DeliveryProvider, error)
	GetByType(ctx context.Context, providerType entity.ProviderType) ([]*entity.DeliveryProvider, error)
	GetAPIProviders(ctx context.Context) ([]*entity.DeliveryProvider, error)
	GetManualProviders(ctx context.Context) ([]*entity.DeliveryProvider, error)
	
	// Provider capabilities
	GetProvidersForArea(ctx context.Context, province, district string) ([]*entity.DeliveryProvider, error)
	GetProvidersForWeight(ctx context.Context, weight float64) ([]*entity.DeliveryProvider, error)
	GetProvidersForServiceLevel(ctx context.Context, serviceLevel string) ([]*entity.DeliveryProvider, error)
	GetRateComparisonProviders(ctx context.Context) ([]*entity.DeliveryProvider, error)
	
	// Configuration management
	UpdateAPIConfiguration(ctx context.Context, id uuid.UUID, config map[string]interface{}) error
	UpdatePricing(ctx context.Context, id uuid.UUID, pricing map[string]interface{}) error
	UpdateServiceLevels(ctx context.Context, id uuid.UUID, serviceLevels map[string]interface{}) error
	UpdateCoverageAreas(ctx context.Context, id uuid.UUID, areas map[string]interface{}) error
	
	// Performance tracking
	UpdatePerformanceMetrics(ctx context.Context, id uuid.UUID, metrics map[string]interface{}) error
	GetProviderMetrics(ctx context.Context, id uuid.UUID, startDate, endDate time.Time) (*ProviderMetrics, error)
	GetTopPerformingProviders(ctx context.Context, metric string, limit int) ([]*entity.DeliveryProvider, error)
	
	// Status management
	ActivateProvider(ctx context.Context, id uuid.UUID) error
	DeactivateProvider(ctx context.Context, id uuid.UUID) error
	SetPriority(ctx context.Context, id uuid.UUID, priority int) error
	
	// Special provider configurations
	GetInterExpressProvider(ctx context.Context) (*entity.DeliveryProvider, error)
	GetNimExpressProvider(ctx context.Context) (*entity.DeliveryProvider, error)
	GetRotRaoProvider(ctx context.Context) (*entity.DeliveryProvider, error)
	
	// Search and filtering
	SearchProviders(ctx context.Context, filters *ProviderQueryFilters) ([]*entity.DeliveryProvider, error)
	GetProvidersByPriceRange(ctx context.Context, minPrice, maxPrice float64) ([]*entity.DeliveryProvider, error)
	GetProvidersByRating(ctx context.Context, minRating float64) ([]*entity.DeliveryProvider, error)
	
	// Bulk operations
	UpdateMultipleProviderStatuses(ctx context.Context, providerIDs []uuid.UUID, isActive bool) error
	GetProvidersByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.DeliveryProvider, error)
}

// ProviderQueryFilters represents filters for provider queries
type ProviderQueryFilters struct {
	ProviderType      *entity.ProviderType `json:"provider_type,omitempty"`
	IsActive          *bool                `json:"is_active,omitempty"`
	HasAPI            *bool                `json:"has_api,omitempty"`
	ManualCoordination *bool               `json:"manual_coordination,omitempty"`
	SupportsRateComparison *bool           `json:"supports_rate_comparison,omitempty"`
	Province          *string              `json:"province,omitempty"`
	MinRating         *float64             `json:"min_rating,omitempty"`
	MaxBaseRate       *float64             `json:"max_base_rate,omitempty"`
	SameDayAvailable  *bool                `json:"same_day_available,omitempty"`
	CODAvailable      *bool                `json:"cod_available,omitempty"`
	Limit             int                  `json:"limit"`
	Offset            int                  `json:"offset"`
}

// ProviderMetrics represents provider performance metrics
type ProviderMetrics struct {
	ProviderID           uuid.UUID `json:"provider_id"`
	ProviderCode         string    `json:"provider_code"`
	ProviderName         string    `json:"provider_name"`
	TotalDeliveries      int64     `json:"total_deliveries"`
	SuccessfulDeliveries int64     `json:"successful_deliveries"`
	FailedDeliveries     int64     `json:"failed_deliveries"`
	CancelledDeliveries  int64     `json:"cancelled_deliveries"`
	AverageDeliveryTime  float64   `json:"average_delivery_time_hours"`
	OnTimeDeliveryRate   float64   `json:"on_time_delivery_rate_percentage"`
	SuccessRate          float64   `json:"success_rate_percentage"`
	CustomerRating       float64   `json:"customer_rating"`
	TotalRevenue         float64   `json:"total_revenue"`
	AverageOrderValue    float64   `json:"average_order_value"`
	PeriodStart          time.Time `json:"period_start"`
	PeriodEnd            time.Time `json:"period_end"`
}

// ProviderCapability represents what a provider can handle
type ProviderCapability struct {
	ProviderID        uuid.UUID              `json:"provider_id"`
	SupportedProvinces []string              `json:"supported_provinces"`
	MaxWeight         float64                `json:"max_weight_kg"`
	MaxDimensions     map[string]float64     `json:"max_dimensions"`
	ServiceLevels     []string               `json:"service_levels"`
	PricingTiers      map[string]interface{} `json:"pricing_tiers"`
	Features          []string               `json:"features"`
}

// ProviderRateQuote represents a rate quote from a provider
type ProviderRateQuote struct {
	ProviderID        uuid.UUID   `json:"provider_id"`
	ProviderCode      string      `json:"provider_code"`
	ProviderName      string      `json:"provider_name"`
	ServiceLevel      string      `json:"service_level"`
	BaseRate          float64     `json:"base_rate"`
	DistanceCharge    float64     `json:"distance_charge"`
	WeightCharge      float64     `json:"weight_charge"`
	Surcharges        float64     `json:"surcharges"`
	TotalCost         float64     `json:"total_cost"`
	Currency          string      `json:"currency"`
	EstimatedDelivery time.Time   `json:"estimated_delivery"`
	ValidUntil        time.Time   `json:"valid_until"`
	QuoteReference    string      `json:"quote_reference"`
	Terms             []string    `json:"terms"`
	Restrictions      []string    `json:"restrictions"`
}
