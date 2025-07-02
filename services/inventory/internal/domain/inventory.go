package domain

import (
	"time"
)

// Product represents a product in the inventory system
type Product struct {
	ID             string             `json:"id" db:"id"`
	Name           string             `json:"name" db:"name"`
	SKU            string             `json:"sku" db:"sku"`
	Barcode        string             `json:"barcode" db:"barcode"`
	CategoryID     string             `json:"category_id" db:"category_id"`
	CategoryName   string             `json:"category_name" db:"category_name"`
	SupplierID     string             `json:"supplier_id" db:"supplier_id"`
	SupplierName   string             `json:"supplier_name" db:"supplier_name"`
	CostPrice      float64            `json:"cost_price" db:"cost_price"`
	SellPrice      float64            `json:"sell_price" db:"sell_price"`
	Unit           string             `json:"unit" db:"unit"`
	Description    string             `json:"description" db:"description"`
	IsActive       bool               `json:"is_active" db:"is_active"`
	StockLevels    []StockLevel       `json:"stock_levels,omitempty"`
	MovementRecord []StockMovement    `json:"movement_record,omitempty"`
	LastUpdated    time.Time          `json:"last_updated" db:"last_updated"`
}

// StockLevel represents stock quantity at a specific store
type StockLevel struct {
	ProductID      string    `json:"product_id" db:"product_id"`
	StoreID        string    `json:"store_id" db:"store_id"`
	StoreName      string    `json:"store_name" db:"store_name"`
	QuantityOnHand float64   `json:"quantity_on_hand" db:"quantity_on_hand"`
	ReorderLevel   float64   `json:"reorder_level" db:"reorder_level"`
	MaxStock       float64   `json:"max_stock" db:"max_stock"`
	IsLowStock     bool      `json:"is_low_stock" db:"is_low_stock"`
	LastUpdated    time.Time `json:"last_updated" db:"last_updated"`
}

// StockMovement represents inventory movement records
type StockMovement struct {
	ID            string    `json:"id" db:"id"`
	ProductID     string    `json:"product_id" db:"product_id"`
	StoreID       string    `json:"store_id" db:"store_id"`
	MovementType  string    `json:"movement_type" db:"movement_type"` // SALE, PURCHASE, ADJUSTMENT, TRANSFER
	Quantity      float64   `json:"quantity" db:"quantity"`
	QuantityBefore float64  `json:"quantity_before" db:"quantity_before"`
	QuantityAfter  float64  `json:"quantity_after" db:"quantity_after"`
	Reference     string    `json:"reference" db:"reference"` // Receipt ID, Adjustment ID, etc.
	Notes         string    `json:"notes" db:"notes"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// InventoryAlert represents low stock or other inventory alerts
type InventoryAlert struct {
	ID          string    `json:"id" db:"id"`
	ProductID   string    `json:"product_id" db:"product_id"`
	ProductName string    `json:"product_name" db:"product_name"`
	StoreID     string    `json:"store_id" db:"store_id"`
	StoreName   string    `json:"store_name" db:"store_name"`
	AlertType   string    `json:"alert_type" db:"alert_type"` // LOW_STOCK, OUT_OF_STOCK, OVERSTOCKED
	CurrentQty  float64   `json:"current_qty" db:"current_qty"`
	ThresholdQty float64  `json:"threshold_qty" db:"threshold_qty"`
	Severity    string    `json:"severity" db:"severity"` // HIGH, MEDIUM, LOW
	IsResolved  bool      `json:"is_resolved" db:"is_resolved"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
}

// Store represents a store location
type Store struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Address     string    `json:"address" db:"address"`
	Phone       string    `json:"phone" db:"phone"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Category represents a product category
type Category struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Color     string    `json:"color" db:"color"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
