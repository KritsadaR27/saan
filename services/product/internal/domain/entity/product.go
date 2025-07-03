package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Product represents a product in the system
type Product struct {
	ID          uuid.UUID          `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	LoyverseID  *string            `json:"loyverse_id" gorm:"uniqueIndex"`
	Name        string             `json:"name" gorm:"not null"`
	Description string             `json:"description"`
	SKU         string             `json:"sku" gorm:"uniqueIndex;not null"`
	Barcode     *string            `json:"barcode" gorm:"uniqueIndex"`
	CategoryID  *uuid.UUID         `json:"category_id" gorm:"type:uuid"`
	BasePrice   float64            `json:"base_price" gorm:"not null"`
	Unit        string             `json:"unit" gorm:"not null"`
	Weight      *float64           `json:"weight"`
	Dimensions  *ProductDimensions `json:"dimensions" gorm:"embedded"`
	IsActive    bool               `json:"is_active" gorm:"default:true"`
	IsVIPOnly   bool               `json:"is_vip_only" gorm:"default:false"`
	Tags        []string           `json:"tags" gorm:"type:text[]"`

	// Master Data Protection
	DataSourceType   string     `json:"data_source_type" gorm:"not null"` // "loyverse", "manual"
	DataSourceID     *string    `json:"data_source_id"`
	LastSyncedAt     *time.Time `json:"last_synced_at"`
	IsManualOverride bool       `json:"is_manual_override" gorm:"default:false"`

	// Audit fields
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedBy *uuid.UUID `json:"created_by" gorm:"type:uuid"`
	UpdatedBy *uuid.UUID `json:"updated_by" gorm:"type:uuid"`
	Version   int        `json:"version" gorm:"default:1"`
}

// ProductDimensions represents product physical dimensions
type ProductDimensions struct {
	Length float64 `json:"length"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Unit   string  `json:"unit"` // "cm", "m", "in", "ft"
}

// NewProduct creates a new product with validation
func NewProduct(name, sku, unit string, basePrice float64) (*Product, error) {
	if name == "" {
		return nil, errors.New("product name is required")
	}
	if sku == "" {
		return nil, errors.New("product SKU is required")
	}
	if unit == "" {
		return nil, errors.New("product unit is required")
	}
	if basePrice < 0 {
		return nil, errors.New("base price must be non-negative")
	}

	return &Product{
		ID:             uuid.New(),
		Name:           name,
		SKU:            sku,
		Unit:           unit,
		BasePrice:      basePrice,
		IsActive:       true,
		IsVIPOnly:      false,
		DataSourceType: "manual",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Version:        1,
	}, nil
}

// UpdateName updates the product name
func (p *Product) UpdateName(name string) error {
	if name == "" {
		return errors.New("product name is required")
	}
	p.Name = name
	p.UpdatedAt = time.Now()
	return nil
}

// UpdatePrice updates the base price
func (p *Product) UpdatePrice(price float64) error {
	if price < 0 {
		return errors.New("base price must be non-negative")
	}
	p.BasePrice = price
	p.UpdatedAt = time.Now()
	return nil
}

// SetVIPOnly sets the VIP only flag
func (p *Product) SetVIPOnly(vipOnly bool) {
	p.IsVIPOnly = vipOnly
	p.UpdatedAt = time.Now()
}

// Activate activates the product
func (p *Product) Activate() {
	p.IsActive = true
	p.UpdatedAt = time.Now()
}

// Deactivate deactivates the product
func (p *Product) Deactivate() {
	p.IsActive = false
	p.UpdatedAt = time.Now()
}

// SetManualOverride sets manual override flag
func (p *Product) SetManualOverride(override bool) {
	p.IsManualOverride = override
	p.UpdatedAt = time.Now()
}

// MarkSynced marks the product as synced from external source
func (p *Product) MarkSynced() {
	now := time.Now()
	p.LastSyncedAt = &now
	p.UpdatedAt = now
}

// IsFromLoyverse checks if product is from Loyverse
func (p *Product) IsFromLoyverse() bool {
	return p.DataSourceType == "loyverse"
}

// CanBeModified checks if product can be modified (not protected by master data)
func (p *Product) CanBeModified() bool {
	return p.IsManualOverride || p.DataSourceType == "manual"
}

// AddTag adds a tag to the product
func (p *Product) AddTag(tag string) {
	if tag == "" {
		return
	}

	// Check if tag already exists
	for _, existingTag := range p.Tags {
		if existingTag == tag {
			return
		}
	}

	p.Tags = append(p.Tags, tag)
	p.UpdatedAt = time.Now()
}

// RemoveTag removes a tag from the product
func (p *Product) RemoveTag(tag string) {
	var newTags []string
	for _, existingTag := range p.Tags {
		if existingTag != tag {
			newTags = append(newTags, existingTag)
		}
	}
	p.Tags = newTags
	p.UpdatedAt = time.Now()
}

// Validate validates the product
func (p *Product) Validate() error {
	if p.Name == "" {
		return errors.New("product name is required")
	}
	if p.SKU == "" {
		return errors.New("product SKU is required")
	}
	if p.Unit == "" {
		return errors.New("product unit is required")
	}
	if p.BasePrice < 0 {
		return errors.New("base price must be non-negative")
	}
	if p.DataSourceType == "" {
		return errors.New("data source type is required")
	}

	return nil
}
