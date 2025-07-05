package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"product/internal/domain/entity"
)

// ProductEnhancedRepository defines the interface for enhanced product operations
type ProductEnhancedRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, product *entity.ProductEnhanced) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.ProductEnhanced, error)
	GetByExternalID(ctx context.Context, externalID string) (*entity.ProductEnhanced, error)
	GetBySKU(ctx context.Context, sku string) (*entity.ProductEnhanced, error)
	Update(ctx context.Context, product *entity.ProductEnhanced) error
	UpdateLoyverseFields(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	UpdateAdminFields(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID, reason string) error

	// Query operations
	List(ctx context.Context, filters ProductFilters) ([]*entity.ProductEnhanced, error)
	Search(ctx context.Context, query string, filters ProductFilters) ([]*entity.ProductEnhanced, error)
	GetFeaturedProducts(ctx context.Context, limit int) ([]*entity.ProductEnhanced, error)
	GetByCategory(ctx context.Context, categoryID uuid.UUID, filters ProductFilters) ([]*entity.ProductEnhanced, error)
	GetVIPProducts(ctx context.Context, vipLevel string) ([]*entity.ProductEnhanced, error)

	// Availability operations
	CheckAvailability(ctx context.Context, productID uuid.UUID, customerVIPLevel *string) (bool, string, error)
	UpdateAvailability(ctx context.Context, productID uuid.UUID, isActive bool, reason string, changedBy *uuid.UUID) error
	ScheduleAvailability(ctx context.Context, productID uuid.UUID, inactiveUntil time.Time, autoReactivate bool, reason string, changedBy *uuid.UUID) error
	GetAvailabilitySchedule(ctx context.Context, dateFrom, dateTo time.Time) ([]*entity.ProductEnhanced, error)
	ProcessScheduledReactivations(ctx context.Context) ([]uuid.UUID, error)

	// Master Data Protection operations
	UpsertFromLoyverse(ctx context.Context, externalID string, loyverseData map[string]interface{}) (*entity.ProductEnhanced, error)
	GetFieldPolicy(ctx context.Context) (*entity.ProductFieldPolicy, error)
	UpdateFieldPolicy(ctx context.Context, policy *entity.ProductFieldPolicy) error
	GetSyncConflicts(ctx context.Context) ([]*SyncConflict, error)
	ResolveSyncConflict(ctx context.Context, conflictID uuid.UUID, resolution ConflictResolution) error

	// Bulk operations
	CreateBatch(ctx context.Context, products []*entity.ProductEnhanced) error
	UpdateBatch(ctx context.Context, updates []ProductUpdate) error
	SyncFromLoyverseBatch(ctx context.Context, loyverseProducts []LoyverseProductData) error

	// Analytics operations
	GetProductMetrics(ctx context.Context, productID uuid.UUID, dateFrom, dateTo time.Time) (*ProductMetrics, error)
	GetTopSellingProducts(ctx context.Context, limit int, dateFrom, dateTo time.Time) ([]*ProductSalesData, error)
	GetLowStockProducts(ctx context.Context, threshold int) ([]*entity.ProductEnhanced, error)
}

// PricingRepository defines the interface for pricing operations
type PricingRepository interface {
	// Pricing tier operations
	CreatePricingTier(ctx context.Context, tier *entity.ProductPricingTier) error
	GetPricingTiers(ctx context.Context, productID uuid.UUID) ([]*entity.ProductPricingTier, error)
	UpdatePricingTier(ctx context.Context, tier *entity.ProductPricingTier) error
	DeletePricingTier(ctx context.Context, tierID uuid.UUID) error

	// Customer group pricing
	CreateCustomerGroupPricing(ctx context.Context, pricing *entity.CustomerGroupPricing) error
	GetCustomerGroupPricing(ctx context.Context, productID uuid.UUID, customerGroup string) (*entity.CustomerGroupPricing, error)
	UpdateCustomerGroupPricing(ctx context.Context, pricing *entity.CustomerGroupPricing) error
	DeleteCustomerGroupPricing(ctx context.Context, pricingID uuid.UUID) error

	// VIP pricing benefits
	CreateVIPBenefits(ctx context.Context, benefits *entity.VIPPricingBenefits) error
	GetVIPBenefits(ctx context.Context, vipLevel string) (*entity.VIPPricingBenefits, error)
	UpdateVIPBenefits(ctx context.Context, benefits *entity.VIPPricingBenefits) error
	DeleteVIPBenefits(ctx context.Context, benefitsID uuid.UUID) error
	ListVIPBenefits(ctx context.Context) ([]*entity.VIPPricingBenefits, error)

	// Price calculation
	CalculatePrice(ctx context.Context, request PricingRequest) (*PricingResult, error)
	GetOptimalPricingTier(ctx context.Context, productID uuid.UUID, quantity int) (*entity.ProductPricingTier, error)
	ValidatePricingTiers(ctx context.Context, productID uuid.UUID) error
}

// AvailabilityLogRepository defines the interface for availability audit log operations
type AvailabilityLogRepository interface {
	LogAvailabilityChange(ctx context.Context, log *entity.ProductAvailabilityLog) error
	GetAvailabilityHistory(ctx context.Context, productID uuid.UUID, limit int) ([]*entity.ProductAvailabilityLog, error)
	GetAvailabilityChanges(ctx context.Context, dateFrom, dateTo time.Time) ([]*entity.ProductAvailabilityLog, error)
	GetChangesByUser(ctx context.Context, userID uuid.UUID, dateFrom, dateTo time.Time) ([]*entity.ProductAvailabilityLog, error)
}

// Supporting types

// ProductFilters represents filters for product queries
type ProductFilters struct {
	Status           *string
	SourceSystem     *string
	CategoryID       *uuid.UUID
	IsActive         *bool
	IsAdminActive    *bool
	VIPOnly          *bool
	IsFeatured       *bool
	CustomerVIPLevel *string
	AvailableOnly    bool
	MinPrice         *decimal.Decimal
	MaxPrice         *decimal.Decimal
	Tags             []string
	SearchQuery      *string
	Limit            int
	Offset           int
	SortBy           string
	SortOrder        string // "ASC" or "DESC"
}

// PricingRequest represents a pricing calculation request
type PricingRequest struct {
	ProductID     uuid.UUID
	Quantity      int
	CustomerGroup *string
	VIPLevel      *string
	LocationID    *uuid.UUID
	CalculateAt   *time.Time
}

// PricingResult represents the result of price calculation
type PricingResult struct {
	ProductID      uuid.UUID        `json:"product_id"`
	BasePrice      decimal.Decimal  `json:"base_price"`
	TierPrice      decimal.Decimal  `json:"tier_price"`
	GroupDiscount  decimal.Decimal  `json:"group_discount"`
	VIPDiscount    decimal.Decimal  `json:"vip_discount"`
	FinalPrice     decimal.Decimal  `json:"final_price"`
	TotalPrice     decimal.Decimal  `json:"total_price"`
	Savings        decimal.Decimal  `json:"savings"`
	AppliedTier    *entity.ProductPricingTier `json:"applied_tier,omitempty"`
	AppliedBenefits *entity.VIPPricingBenefits `json:"applied_benefits,omitempty"`
	Breakdown      PriceBreakdown   `json:"breakdown"`
}

// PriceBreakdown provides detailed price calculation breakdown
type PriceBreakdown struct {
	OriginalPrice     decimal.Decimal `json:"original_price"`
	TierDiscountApplied bool           `json:"tier_discount_applied"`
	TierDiscountAmount  decimal.Decimal `json:"tier_discount_amount"`
	GroupDiscountApplied bool          `json:"group_discount_applied"`
	GroupDiscountAmount  decimal.Decimal `json:"group_discount_amount"`
	VIPDiscountApplied   bool           `json:"vip_discount_applied"`
	VIPDiscountAmount    decimal.Decimal `json:"vip_discount_amount"`
	VIPQuantityMultiplier decimal.Decimal `json:"vip_quantity_multiplier"`
}

// ProductUpdate represents a product update operation
type ProductUpdate struct {
	ProductID uuid.UUID
	Fields    map[string]interface{}
	UpdateType string // "admin", "loyverse", "system"
}

// LoyverseProductData represents product data from Loyverse sync
type LoyverseProductData struct {
	ExternalID   string
	Name         string
	Description  *string
	SKU          *string
	Barcode      *string
	CategoryID   *uuid.UUID
	CostPrice    *decimal.Decimal
	SellingPrice *decimal.Decimal
	Status       string
	SyncedAt     time.Time
}

// SyncConflict represents a conflict during Loyverse sync
type SyncConflict struct {
	ID              uuid.UUID              `json:"id"`
	ProductID       uuid.UUID              `json:"product_id"`
	ExternalID      string                 `json:"external_id"`
	ConflictType    string                 `json:"conflict_type"` // "field_mismatch", "deletion_conflict"
	ConflictField   string                 `json:"conflict_field"`
	LocalValue      interface{}            `json:"local_value"`
	RemoteValue     interface{}            `json:"remote_value"`
	ConflictedAt    time.Time              `json:"conflicted_at"`
	ResolvedAt      *time.Time             `json:"resolved_at,omitempty"`
	Resolution      *ConflictResolution    `json:"resolution,omitempty"`
	AdditionalInfo  map[string]interface{} `json:"additional_info,omitempty"`
}

// ConflictResolution represents how a sync conflict should be resolved
type ConflictResolution struct {
	Type        string      `json:"type"` // "use_local", "use_remote", "manual"
	ManualValue interface{} `json:"manual_value,omitempty"`
	ResolvedBy  uuid.UUID   `json:"resolved_by"`
	Reason      string      `json:"reason"`
}

// ProductMetrics represents product performance metrics
type ProductMetrics struct {
	ProductID       uuid.UUID       `json:"product_id"`
	ViewCount       int             `json:"view_count"`
	SearchCount     int             `json:"search_count"`
	OrderCount      int             `json:"order_count"`
	TotalSold       int             `json:"total_sold"`
	Revenue         decimal.Decimal `json:"revenue"`
	AvgOrderValue   decimal.Decimal `json:"avg_order_value"`
	ConversionRate  decimal.Decimal `json:"conversion_rate"`
	DateFrom        time.Time       `json:"date_from"`
	DateTo          time.Time       `json:"date_to"`
}

// ProductSalesData represents product sales information
type ProductSalesData struct {
	Product     *entity.ProductEnhanced `json:"product"`
	TotalSold   int                     `json:"total_sold"`
	Revenue     decimal.Decimal         `json:"revenue"`
	OrderCount  int                     `json:"order_count"`
	AvgPrice    decimal.Decimal         `json:"avg_price"`
	LastSold    *time.Time              `json:"last_sold,omitempty"`
}

// Constants for repository operations
const (
	UpdateTypeLoyverse = "loyverse"
	UpdateTypeAdmin    = "admin"
	UpdateTypeSystem   = "system"

	ConflictTypeFieldMismatch    = "field_mismatch"
	ConflictTypeDeletionConflict = "deletion_conflict"
	ConflictTypeNewProduct       = "new_product"

	ResolutionTypeUseLocal  = "use_local"
	ResolutionTypeUseRemote = "use_remote"
	ResolutionTypeManual    = "manual"

	SortByName      = "name"
	SortByPrice     = "price"
	SortByCreated   = "created_at"
	SortByUpdated   = "updated_at"
	SortByPopularity = "popularity"

	SortOrderASC  = "ASC"
	SortOrderDESC = "DESC"

	DefaultLimit = 50
	MaxLimit     = 1000
)
