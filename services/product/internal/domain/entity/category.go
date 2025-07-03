package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Category represents product categories
type Category struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	LoyverseID  *string    `json:"loyverse_id" gorm:"uniqueIndex"`
	Name        string     `json:"name" gorm:"not null"`
	Description string     `json:"description"`
	ParentID    *uuid.UUID `json:"parent_id" gorm:"type:uuid"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	SortOrder   int        `json:"sort_order" gorm:"default:0"`

	// Master Data Protection
	DataSourceType   string     `json:"data_source_type" gorm:"not null"`
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

// NewCategory creates a new category with validation
func NewCategory(name string) (*Category, error) {
	if name == "" {
		return nil, errors.New("category name is required")
	}

	return &Category{
		ID:             uuid.New(),
		Name:           name,
		IsActive:       true,
		SortOrder:      0,
		DataSourceType: "manual",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Version:        1,
	}, nil
}

// UpdateName updates the category name
func (c *Category) UpdateName(name string) error {
	if name == "" {
		return errors.New("category name is required")
	}
	c.Name = name
	c.UpdatedAt = time.Now()
	return nil
}

// SetDescription sets the category description
func (c *Category) SetDescription(description string) {
	c.Description = description
	c.UpdatedAt = time.Now()
}

// SetParent sets the parent category
func (c *Category) SetParent(parentID uuid.UUID) {
	c.ParentID = &parentID
	c.UpdatedAt = time.Now()
}

// RemoveParent removes the parent category
func (c *Category) RemoveParent() {
	c.ParentID = nil
	c.UpdatedAt = time.Now()
}

// Activate activates the category
func (c *Category) Activate() {
	c.IsActive = true
	c.UpdatedAt = time.Now()
}

// Deactivate deactivates the category
func (c *Category) Deactivate() {
	c.IsActive = false
	c.UpdatedAt = time.Now()
}

// SetSortOrder sets the sort order
func (c *Category) SetSortOrder(order int) {
	c.SortOrder = order
	c.UpdatedAt = time.Now()
}

// SetManualOverride sets manual override flag
func (c *Category) SetManualOverride(override bool) {
	c.IsManualOverride = override
	c.UpdatedAt = time.Now()
}

// MarkSynced marks the category as synced from external source
func (c *Category) MarkSynced() {
	now := time.Now()
	c.LastSyncedAt = &now
	c.UpdatedAt = now
}

// IsFromLoyverse checks if category is from Loyverse
func (c *Category) IsFromLoyverse() bool {
	return c.DataSourceType == "loyverse"
}

// CanBeModified checks if category can be modified (not protected by master data)
func (c *Category) CanBeModified() bool {
	return c.IsManualOverride || c.DataSourceType == "manual"
}

// IsRootCategory checks if this is a root category (no parent)
func (c *Category) IsRootCategory() bool {
	return c.ParentID == nil
}

// Validate validates the category
func (c *Category) Validate() error {
	if c.Name == "" {
		return errors.New("category name is required")
	}
	if c.DataSourceType == "" {
		return errors.New("data source type is required")
	}

	return nil
}
