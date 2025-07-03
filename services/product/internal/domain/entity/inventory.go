package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Inventory represents product inventory levels
type Inventory struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProductID  uuid.UUID `json:"product_id" gorm:"type:uuid;not null"`
	LocationID uuid.UUID `json:"location_id" gorm:"type:uuid;not null"`

	// Stock levels
	StockLevel     float64 `json:"stock_level" gorm:"not null;default:0"`
	ReservedLevel  float64 `json:"reserved_level" gorm:"not null;default:0"`
	AvailableLevel float64 `json:"available_level" gorm:"not null;default:0"`

	// Thresholds
	LowStockThreshold *float64 `json:"low_stock_threshold"`
	ReorderPoint      *float64 `json:"reorder_point"`
	MaxStockLevel     *float64 `json:"max_stock_level"`

	// Availability
	IsAvailable       bool    `json:"is_available" gorm:"default:true"`
	AvailabilityNotes *string `json:"availability_notes"`

	// Audit fields
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	Version   int       `json:"version" gorm:"default:1"`
}

// ProductAvailability represents product availability status
type ProductAvailability struct {
	ProductID        uuid.UUID              `json:"product_id"`
	LocationID       uuid.UUID              `json:"location_id"`
	IsAvailable      bool                   `json:"is_available"`
	StockLevel       float64                `json:"stock_level"`
	ReservedLevel    float64                `json:"reserved_level"`
	AvailableLevel   float64                `json:"available_level"`
	EstimatedRestock *time.Time             `json:"estimated_restock"`
	Reason           *string                `json:"reason"`
	Locations        []LocationAvailability `json:"locations"`
}

// LocationAvailability represents availability at specific location
type LocationAvailability struct {
	LocationID       uuid.UUID  `json:"location_id"`
	LocationName     string     `json:"location_name"`
	IsAvailable      bool       `json:"is_available"`
	StockLevel       float64    `json:"stock_level"`
	AvailableLevel   float64    `json:"available_level"`
	EstimatedRestock *time.Time `json:"estimated_restock"`
}

// NewInventory creates a new inventory record
func NewInventory(productID, locationID uuid.UUID) *Inventory {
	return &Inventory{
		ID:             uuid.New(),
		ProductID:      productID,
		LocationID:     locationID,
		StockLevel:     0,
		ReservedLevel:  0,
		AvailableLevel: 0,
		IsAvailable:    true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Version:        1,
	}
}

// UpdateStockLevel updates the stock level and recalculates available level
func (i *Inventory) UpdateStockLevel(level float64) error {
	if level < 0 {
		return errors.New("stock level cannot be negative")
	}
	i.StockLevel = level
	i.AvailableLevel = i.StockLevel - i.ReservedLevel
	i.UpdatedAt = time.Now()
	return nil
}

// ReserveStock reserves stock for an order
func (i *Inventory) ReserveStock(quantity float64) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	if i.AvailableLevel < quantity {
		return errors.New("insufficient stock available")
	}

	i.ReservedLevel += quantity
	i.AvailableLevel = i.StockLevel - i.ReservedLevel
	i.UpdatedAt = time.Now()
	return nil
}

// ReleaseStock releases reserved stock
func (i *Inventory) ReleaseStock(quantity float64) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	if i.ReservedLevel < quantity {
		return errors.New("not enough reserved stock to release")
	}

	i.ReservedLevel -= quantity
	i.AvailableLevel = i.StockLevel - i.ReservedLevel
	i.UpdatedAt = time.Now()
	return nil
}

// ConsumeStock consumes stock (removes from both stock and reserved levels)
func (i *Inventory) ConsumeStock(quantity float64) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	if i.ReservedLevel < quantity {
		return errors.New("not enough reserved stock to consume")
	}

	i.StockLevel -= quantity
	i.ReservedLevel -= quantity
	i.AvailableLevel = i.StockLevel - i.ReservedLevel
	i.UpdatedAt = time.Now()
	return nil
}

// SetThresholds sets inventory thresholds
func (i *Inventory) SetThresholds(lowStock, reorderPoint, maxStock *float64) {
	i.LowStockThreshold = lowStock
	i.ReorderPoint = reorderPoint
	i.MaxStockLevel = maxStock
	i.UpdatedAt = time.Now()
}

// SetAvailability sets the availability status
func (i *Inventory) SetAvailability(available bool, notes *string) {
	i.IsAvailable = available
	i.AvailabilityNotes = notes
	i.UpdatedAt = time.Now()
}

// IsLowStock checks if stock is below low stock threshold
func (i *Inventory) IsLowStock() bool {
	return i.LowStockThreshold != nil && i.StockLevel <= *i.LowStockThreshold
}

// NeedsReorder checks if stock needs reordering
func (i *Inventory) NeedsReorder() bool {
	return i.ReorderPoint != nil && i.StockLevel <= *i.ReorderPoint
}

// IsOverStock checks if stock is above max threshold
func (i *Inventory) IsOverStock() bool {
	return i.MaxStockLevel != nil && i.StockLevel > *i.MaxStockLevel
}

// GetAvailableStock returns the available stock level
func (i *Inventory) GetAvailableStock() float64 {
	return i.AvailableLevel
}

// Validate validates the inventory record
func (i *Inventory) Validate() error {
	if i.ProductID == uuid.Nil {
		return errors.New("product ID is required")
	}
	if i.LocationID == uuid.Nil {
		return errors.New("location ID is required")
	}
	if i.StockLevel < 0 {
		return errors.New("stock level cannot be negative")
	}
	if i.ReservedLevel < 0 {
		return errors.New("reserved level cannot be negative")
	}
	if i.ReservedLevel > i.StockLevel {
		return errors.New("reserved level cannot exceed stock level")
	}

	return nil
}
