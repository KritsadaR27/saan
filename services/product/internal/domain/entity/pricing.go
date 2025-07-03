package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Price represents product pricing with advanced features
type Price struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProductID uuid.UUID `json:"product_id" gorm:"type:uuid;not null"`
	Product   *Product  `json:"product" gorm:"foreignKey:ProductID"`
	PriceType string    `json:"price_type" gorm:"not null"` // "base", "vip", "bulk", "promotional"
	Price     float64   `json:"price" gorm:"not null"`
	Currency  string    `json:"currency" gorm:"not null;default:'THB'"`

	// Bulk pricing
	MinQuantity *int `json:"min_quantity"`
	MaxQuantity *int `json:"max_quantity"`

	// VIP pricing
	VIPTierID *uuid.UUID `json:"vip_tier_id" gorm:"type:uuid"`

	// Promotional pricing
	ValidFrom       *time.Time `json:"valid_from"`
	ValidTo         *time.Time `json:"valid_to"`
	PromotionName   *string    `json:"promotion_name"`
	DiscountPercent *float64   `json:"discount_percent"`

	// Conditions
	LocationIDs      []uuid.UUID `json:"location_ids" gorm:"type:uuid[]"`
	CustomerGroupIDs []uuid.UUID `json:"customer_group_ids" gorm:"type:uuid[]"`

	IsActive bool `json:"is_active" gorm:"default:true"`
	Priority int  `json:"priority" gorm:"default:0"`

	// Audit fields
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedBy *uuid.UUID `json:"created_by" gorm:"type:uuid"`
	UpdatedBy *uuid.UUID `json:"updated_by" gorm:"type:uuid"`
	Version   int        `json:"version" gorm:"default:1"`
}

// PriceCalculation represents calculated pricing for a product
type PriceCalculation struct {
	ProductID       uuid.UUID  `json:"product_id"`
	BasePrice       float64    `json:"base_price"`
	AppliedPrice    float64    `json:"applied_price"`
	PriceType       string     `json:"price_type"`
	Currency        string     `json:"currency"`
	DiscountPercent *float64   `json:"discount_percent"`
	Savings         *float64   `json:"savings"`
	ValidFrom       *time.Time `json:"valid_from"`
	ValidTo         *time.Time `json:"valid_to"`
	Conditions      []string   `json:"conditions"`
}

// PriceRequest represents pricing request parameters
type PriceRequest struct {
	ProductID        uuid.UUID   `json:"product_id"`
	CustomerID       *uuid.UUID  `json:"customer_id"`
	LocationID       *uuid.UUID  `json:"location_id"`
	Quantity         int         `json:"quantity"`
	CustomerGroupIDs []uuid.UUID `json:"customer_group_ids"`
	VIPTierID        *uuid.UUID  `json:"vip_tier_id"`
	RequestTime      time.Time   `json:"request_time"`
}

// BulkPriceRequest represents bulk pricing setup request
type BulkPriceRequest struct {
	ProductID   uuid.UUID   `json:"product_id"`
	MinQuantity int         `json:"min_quantity"`
	MaxQuantity *int        `json:"max_quantity"`
	Price       float64     `json:"price"`
	Currency    string      `json:"currency"`
	LocationIDs []uuid.UUID `json:"location_ids"`
}

// VIPPriceRequest represents VIP pricing setup request
type VIPPriceRequest struct {
	ProductID   uuid.UUID   `json:"product_id"`
	VIPTierID   uuid.UUID   `json:"vip_tier_id"`
	Price       float64     `json:"price"`
	Currency    string      `json:"currency"`
	LocationIDs []uuid.UUID `json:"location_ids"`
}

// PromotionalPriceRequest represents promotional pricing setup request
type PromotionalPriceRequest struct {
	ProductID        uuid.UUID   `json:"product_id"`
	PromotionName    string      `json:"promotion_name"`
	Price            *float64    `json:"price"`
	DiscountPercent  *float64    `json:"discount_percent"`
	ValidFrom        time.Time   `json:"valid_from"`
	ValidTo          time.Time   `json:"valid_to"`
	LocationIDs      []uuid.UUID `json:"location_ids"`
	CustomerGroupIDs []uuid.UUID `json:"customer_group_ids"`
}

// NewPrice creates a new price record
func NewPrice(productID uuid.UUID, priceType string, price float64) (*Price, error) {
	if productID == uuid.Nil {
		return nil, errors.New("product ID is required")
	}
	if priceType == "" {
		return nil, errors.New("price type is required")
	}
	if price < 0 {
		return nil, errors.New("price must be non-negative")
	}

	return &Price{
		ID:        uuid.New(),
		ProductID: productID,
		PriceType: priceType,
		Price:     price,
		Currency:  "THB",
		IsActive:  true,
		Priority:  0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}, nil
}

// UpdatePrice updates the price value
func (p *Price) UpdatePrice(price float64) error {
	if price < 0 {
		return errors.New("price must be non-negative")
	}
	p.Price = price
	p.UpdatedAt = time.Now()
	return nil
}

// SetCurrency sets the currency
func (p *Price) SetCurrency(currency string) {
	p.Currency = currency
	p.UpdatedAt = time.Now()
}

// SetQuantityRange sets quantity range for bulk pricing
func (p *Price) SetQuantityRange(minQty, maxQty *int) {
	p.MinQuantity = minQty
	p.MaxQuantity = maxQty
	p.UpdatedAt = time.Now()
}

// SetVIPTier sets VIP tier for VIP pricing
func (p *Price) SetVIPTier(vipTierID uuid.UUID) {
	p.VIPTierID = &vipTierID
	p.UpdatedAt = time.Now()
}

// SetValidityPeriod sets validity period for promotional pricing
func (p *Price) SetValidityPeriod(validFrom, validTo *time.Time) {
	p.ValidFrom = validFrom
	p.ValidTo = validTo
	p.UpdatedAt = time.Now()
}

// SetPromotionDetails sets promotion details
func (p *Price) SetPromotionDetails(promotionName string, discountPercent *float64) {
	p.PromotionName = &promotionName
	p.DiscountPercent = discountPercent
	p.UpdatedAt = time.Now()
}

// SetLocationIDs sets location IDs where price applies
func (p *Price) SetLocationIDs(locationIDs []uuid.UUID) {
	p.LocationIDs = locationIDs
	p.UpdatedAt = time.Now()
}

// SetCustomerGroupIDs sets customer group IDs for targeted pricing
func (p *Price) SetCustomerGroupIDs(groupIDs []uuid.UUID) {
	p.CustomerGroupIDs = groupIDs
	p.UpdatedAt = time.Now()
}

// SetPriority sets the priority for price calculation
func (p *Price) SetPriority(priority int) {
	p.Priority = priority
	p.UpdatedAt = time.Now()
}

// Activate activates the price
func (p *Price) Activate() {
	p.IsActive = true
	p.UpdatedAt = time.Now()
}

// Deactivate deactivates the price
func (p *Price) Deactivate() {
	p.IsActive = false
	p.UpdatedAt = time.Now()
}

// IsValid checks if the price is valid for the given time
func (p *Price) IsValid(t time.Time) bool {
	if !p.IsActive {
		return false
	}

	if p.ValidFrom != nil && t.Before(*p.ValidFrom) {
		return false
	}

	if p.ValidTo != nil && t.After(*p.ValidTo) {
		return false
	}

	return true
}

// IsPromotional checks if this is a promotional price
func (p *Price) IsPromotional() bool {
	return p.PriceType == "promotional"
}

// IsVIP checks if this is a VIP price
func (p *Price) IsVIP() bool {
	return p.PriceType == "vip"
}

// IsBulk checks if this is a bulk price
func (p *Price) IsBulk() bool {
	return p.PriceType == "bulk"
}

// Validate validates the price record
func (p *Price) Validate() error {
	if p.ProductID == uuid.Nil {
		return errors.New("product ID is required")
	}
	if p.PriceType == "" {
		return errors.New("price type is required")
	}
	if p.Price < 0 {
		return errors.New("price must be non-negative")
	}
	if p.Currency == "" {
		return errors.New("currency is required")
	}

	// Validate promotional pricing
	if p.IsPromotional() {
		if p.ValidFrom == nil || p.ValidTo == nil {
			return errors.New("promotional pricing requires valid from and to dates")
		}
		if p.ValidFrom.After(*p.ValidTo) {
			return errors.New("valid from date must be before valid to date")
		}
	}

	// Validate bulk pricing
	if p.IsBulk() {
		if p.MinQuantity == nil {
			return errors.New("bulk pricing requires minimum quantity")
		}
		if *p.MinQuantity <= 0 {
			return errors.New("minimum quantity must be positive")
		}
		if p.MaxQuantity != nil && *p.MaxQuantity < *p.MinQuantity {
			return errors.New("maximum quantity must be greater than minimum quantity")
		}
	}

	// Validate VIP pricing
	if p.IsVIP() {
		if p.VIPTierID == nil {
			return errors.New("VIP pricing requires VIP tier ID")
		}
	}

	return nil
}
