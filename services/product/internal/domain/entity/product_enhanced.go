package entity

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ProductEnhanced represents a product with Master Data Protection pattern
type ProductEnhanced struct {
	ID uuid.UUID `json:"id" db:"id"`

	// ✅ Loyverse-controlled fields (sync updates these)
	ExternalID             *string           `json:"external_id,omitempty" db:"external_id"`
	SourceSystem           string            `json:"source_system" db:"source_system"`
	Name                   string            `json:"name" db:"name"`
	Description            *string           `json:"description,omitempty" db:"description"`
	SKU                    *string           `json:"sku,omitempty" db:"sku"`
	Barcode                *string           `json:"barcode,omitempty" db:"barcode"`
	CategoryID             *uuid.UUID        `json:"category_id,omitempty" db:"category_id"`
	SupplierID             *uuid.UUID        `json:"supplier_id,omitempty" db:"supplier_id"`
	CostPrice              *decimal.Decimal  `json:"cost_price,omitempty" db:"cost_price"`
	SellingPrice           *decimal.Decimal  `json:"selling_price,omitempty" db:"selling_price"`
	Status                 string            `json:"status" db:"status"`
	LastSyncFromLoyverse   *time.Time        `json:"last_sync_from_loyverse,omitempty" db:"last_sync_from_loyverse"`

	// 🔒 Admin-controlled fields (sync never touches these)
	DisplayName            *string           `json:"display_name,omitempty" db:"display_name"`
	InternalCategory       *string           `json:"internal_category,omitempty" db:"internal_category"`
	InternalNotes          *string           `json:"internal_notes,omitempty" db:"internal_notes"`
	IsFeatured             bool              `json:"is_featured" db:"is_featured"`
	ProfitMarginTarget     *decimal.Decimal  `json:"profit_margin_target,omitempty" db:"profit_margin_target"`
	SalesTags              map[string]interface{} `json:"sales_tags,omitempty" db:"sales_tags"`

	// Product Specifications
	WeightGrams            *decimal.Decimal  `json:"weight_grams,omitempty" db:"weight_grams"`
	UnitsPerPack           int               `json:"units_per_pack" db:"units_per_pack"`
	UnitType               string            `json:"unit_type" db:"unit_type"`

	// Advanced Availability Control
	IsAdminActive          bool              `json:"is_admin_active" db:"is_admin_active"`
	InactiveReason         *string           `json:"inactive_reason,omitempty" db:"inactive_reason"`
	InactiveUntil          *time.Time        `json:"inactive_until,omitempty" db:"inactive_until"`
	AutoReactivate         bool              `json:"auto_reactivate" db:"auto_reactivate"`
	InactiveSchedule       map[string]interface{} `json:"inactive_schedule,omitempty" db:"inactive_schedule"`

	// VIP Access Control
	VIPOnly                bool              `json:"vip_only" db:"vip_only"`
	MinVIPLevel            *string           `json:"min_vip_level,omitempty" db:"min_vip_level"`
	VIPEarlyAccess         bool              `json:"vip_early_access" db:"vip_early_access"`
	EarlyAccessUntil       *time.Time        `json:"early_access_until,omitempty" db:"early_access_until"`

	// System fields
	CreatedAt              time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time         `json:"updated_at" db:"updated_at"`
}

// ProductPricingTier represents quantity-based pricing tiers
type ProductPricingTier struct {
	ID          uuid.UUID        `json:"id" db:"id"`
	ProductID   uuid.UUID        `json:"product_id" db:"product_id"`
	MinQuantity int              `json:"min_quantity" db:"min_quantity"`
	MaxQuantity *int             `json:"max_quantity,omitempty" db:"max_quantity"`
	Price       decimal.Decimal  `json:"price" db:"price"`
	TierName    *string          `json:"tier_name,omitempty" db:"tier_name"`
	IsActive    bool             `json:"is_active" db:"is_active"`
	ValidFrom   *time.Time       `json:"valid_from,omitempty" db:"valid_from"`
	ValidUntil  *time.Time       `json:"valid_until,omitempty" db:"valid_until"`
	CreatedAt   time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at" db:"updated_at"`
}

// CustomerGroupPricing represents customer group-specific pricing
type CustomerGroupPricing struct {
	ID                 uuid.UUID        `json:"id" db:"id"`
	ProductID          uuid.UUID        `json:"product_id" db:"product_id"`
	CustomerGroup      string           `json:"customer_group" db:"customer_group"`
	BasePrice          *decimal.Decimal `json:"base_price,omitempty" db:"base_price"`
	DiscountPercentage *decimal.Decimal `json:"discount_percentage,omitempty" db:"discount_percentage"`
	IsActive           bool             `json:"is_active" db:"is_active"`
	CreatedAt          time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at" db:"updated_at"`
}

// VIPPricingBenefits represents VIP-level pricing benefits
type VIPPricingBenefits struct {
	ID                        uuid.UUID        `json:"id" db:"id"`
	VIPLevel                  string           `json:"vip_level" db:"vip_level"`
	GlobalDiscountPercentage  *decimal.Decimal `json:"global_discount_percentage,omitempty" db:"global_discount_percentage"`
	QuantityMultiplier        decimal.Decimal  `json:"quantity_multiplier" db:"quantity_multiplier"`
	IsActive                  bool             `json:"is_active" db:"is_active"`
	CreatedAt                 time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt                 time.Time        `json:"updated_at" db:"updated_at"`
}

// ProductAvailabilityLog represents availability change audit log
type ProductAvailabilityLog struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	ProductID      uuid.UUID  `json:"product_id" db:"product_id"`
	ChangedBy      *uuid.UUID `json:"changed_by,omitempty" db:"changed_by"`
	ChangeType     string     `json:"change_type" db:"change_type"`
	OldStatus      *bool      `json:"old_status,omitempty" db:"old_status"`
	NewStatus      *bool      `json:"new_status,omitempty" db:"new_status"`
	Reason         *string    `json:"reason,omitempty" db:"reason"`
	ScheduledUntil *time.Time `json:"scheduled_until,omitempty" db:"scheduled_until"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

// ProductFieldPolicy defines which fields can be updated by sync vs admin
type ProductFieldPolicy struct {
	// ✅ Loyverse-controlled (sync updates these)
	SourceFields []string `json:"source_fields"`

	// 🔒 Admin-controlled (sync never touches)
	AdminFields []string `json:"admin_fields"`

	// 🔒 Related tables (sync never touches)
	RelatedTables []string `json:"related_tables"`
}

// GetDefaultFieldPolicy returns the default field policy for Master Data Protection
func GetDefaultFieldPolicy() ProductFieldPolicy {
	return ProductFieldPolicy{
		SourceFields: []string{
			"external_id", "source_system", "name", "description", "sku", "barcode",
			"category_id", "supplier_id", "cost_price", "selling_price", "status",
			"last_sync_from_loyverse",
		},
		AdminFields: []string{
			"display_name", "internal_category", "internal_notes", "is_featured",
			"profit_margin_target", "sales_tags", "weight_grams", "units_per_pack",
			"unit_type", "is_admin_active", "inactive_reason", "inactive_until",
			"auto_reactivate", "inactive_schedule", "vip_only", "min_vip_level",
			"vip_early_access", "early_access_until",
		},
		RelatedTables: []string{
			"product_pricing_tiers", "customer_group_pricing", "vip_pricing_benefits",
			"product_availability_log",
		},
	}
}

// Business logic methods for ProductEnhanced

// IsAvailable checks if the product is available for a customer
func (p *ProductEnhanced) IsAvailable(customerVIPLevel *string) (bool, string) {
	// 1. Check Loyverse status
	if p.Status != "active" {
		return false, "Product inactive in Loyverse"
	}

	// 2. Check admin override
	if !p.IsAdminActive {
		reason := "Administratively disabled"
		if p.InactiveReason != nil {
			reason = *p.InactiveReason
		}
		return false, reason
	}

	// 3. Check time-based inactive
	if p.InactiveUntil != nil && time.Now().Before(*p.InactiveUntil) {
		return false, "Temporarily unavailable"
	}

	// 4. Check VIP access
	if p.VIPOnly && (customerVIPLevel == nil || *customerVIPLevel == "") {
		return false, "VIP members only"
	}

	// 5. Check VIP level requirement
	if p.MinVIPLevel != nil && *p.MinVIPLevel != "" {
		if customerVIPLevel == nil || *customerVIPLevel == "" {
			return false, "Requires VIP membership"
		}
		// Note: VIP level comparison logic would go here
	}

	// 6. Check early access
	if p.VIPEarlyAccess && p.EarlyAccessUntil != nil {
		if time.Now().Before(*p.EarlyAccessUntil) {
			if customerVIPLevel == nil || *customerVIPLevel == "" {
				return false, "VIP early access period"
			}
		}
	}

	return true, ""
}

// GetDisplayName returns the display name, falling back to name if not set
func (p *ProductEnhanced) GetDisplayName() string {
	if p.DisplayName != nil && *p.DisplayName != "" {
		return *p.DisplayName
	}
	return p.Name
}

// IsFromLoyverse checks if the product originates from Loyverse
func (p *ProductEnhanced) IsFromLoyverse() bool {
	return p.SourceSystem == "loyverse" && p.ExternalID != nil
}

// ShouldAutoReactivate checks if the product should be automatically reactivated
func (p *ProductEnhanced) ShouldAutoReactivate() bool {
	if !p.AutoReactivate || p.InactiveUntil == nil {
		return false
	}
	return time.Now().After(*p.InactiveUntil)
}

// Validation methods

// Validate validates the product data
func (p *ProductEnhanced) Validate() error {
	if p.Name == "" {
		return errors.New("product name is required")
	}

	if p.SourceSystem == "" {
		return errors.New("source system is required")
	}

	if p.SourceSystem == "loyverse" && p.ExternalID == nil {
		return errors.New("external ID is required for Loyverse products")
	}

	if p.UnitsPerPack <= 0 {
		return errors.New("units per pack must be greater than 0")
	}

	if p.UnitType == "" {
		return errors.New("unit type is required")
	}

	return nil
}

// NewProductFromLoyverse creates a new product from Loyverse data
func NewProductFromLoyverse(externalID, name string, sellingPrice decimal.Decimal) *ProductEnhanced {
	now := time.Now()
	return &ProductEnhanced{
		ID:                   uuid.New(),
		ExternalID:           &externalID,
		SourceSystem:         "loyverse",
		Name:                 name,
		SellingPrice:         &sellingPrice,
		Status:               "active",
		LastSyncFromLoyverse: &now,
		UnitsPerPack:         1,
		UnitType:             "piece",
		IsAdminActive:        true,
		VIPOnly:              false,
		VIPEarlyAccess:       false,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

// NewManualProduct creates a new manually created product
func NewManualProduct(name string, unitType string) *ProductEnhanced {
	now := time.Now()
	return &ProductEnhanced{
		ID:            uuid.New(),
		SourceSystem:  "manual",
		Name:          name,
		Status:        "active",
		UnitsPerPack:  1,
		UnitType:      unitType,
		IsAdminActive: true,
		VIPOnly:       false,
		VIPEarlyAccess: false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// Helper methods for JSON handling

// ScanSalesTags scans JSONB data into SalesTags
func (p *ProductEnhanced) ScanSalesTags(value interface{}) error {
	if value == nil {
		p.SalesTags = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, &p.SalesTags)
	case string:
		return json.Unmarshal([]byte(v), &p.SalesTags)
	default:
		return errors.New("cannot scan SalesTags")
	}
}

// ScanInactiveSchedule scans JSONB data into InactiveSchedule
func (p *ProductEnhanced) ScanInactiveSchedule(value interface{}) error {
	if value == nil {
		p.InactiveSchedule = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, &p.InactiveSchedule)
	case string:
		return json.Unmarshal([]byte(v), &p.InactiveSchedule)
	default:
		return errors.New("cannot scan InactiveSchedule")
	}
}

// Constants for validation
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusDeleted  = "deleted"

	SourceSystemLoyverse = "loyverse"
	SourceSystemManual   = "manual"

	UnitTypePiece  = "piece"
	UnitTypeKg     = "kg"
	UnitTypeGram   = "gram"
	UnitTypeLiter  = "liter"
	UnitTypeML     = "ml"
	UnitTypeBox    = "box"
	UnitTypePack   = "pack"

	VIPLevelBronze   = "bronze"
	VIPLevelSilver   = "silver"
	VIPLevelGold     = "gold"
	VIPLevelPlatinum = "platinum"

	CustomerGroupRetail    = "retail"
	CustomerGroupWholesale = "wholesale"
	CustomerGroupVIP       = "vip"
	CustomerGroupEmployee  = "employee"

	ChangeTypeActivated     = "activated"
	ChangeTypeDeactivated   = "deactivated"
	ChangeTypeScheduled     = "scheduled"
	ChangeTypeAutoReactivated = "auto_reactivated"
)

// Error definitions
var (
	ErrProductNotFound        = errors.New("product not found")
	ErrProductAlreadyExists   = errors.New("product already exists")
	ErrInvalidProductData     = errors.New("invalid product data")
	ErrInvalidPricingData     = errors.New("invalid pricing data")
	ErrVIPAccessRequired      = errors.New("VIP access required")
	ErrProductNotAvailable    = errors.New("product not available")
	ErrInvalidVIPLevel        = errors.New("invalid VIP level")
)
