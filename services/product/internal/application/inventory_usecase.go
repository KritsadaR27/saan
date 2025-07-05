package application

import (
	"context"
	"fmt"
	"time"

	"product/internal/domain/entity"
	"product/internal/domain/repository"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// InventoryUsecase handles inventory business logic
type InventoryUsecase struct {
	inventoryRepo repository.InventoryRepository
	productRepo   repository.ProductRepository
	logger        *logrus.Logger
}

// NewInventoryUsecase creates a new inventory usecase
func NewInventoryUsecase(inventoryRepo repository.InventoryRepository, productRepo repository.ProductRepository, logger *logrus.Logger) *InventoryUsecase {
	return &InventoryUsecase{
		inventoryRepo: inventoryRepo,
		productRepo:   productRepo,
		logger:        logger,
	}
}

// CreateInventoryRequest represents the request to create inventory
type CreateInventoryRequest struct {
	ProductID         uuid.UUID `json:"product_id" validate:"required"`
	LocationID        uuid.UUID `json:"location_id" validate:"required"`
	StockLevel        float64   `json:"stock_level" validate:"required,min=0"`
	ReorderPoint      float64   `json:"reorder_point" validate:"min=0"`
	MaxStockLevel     float64   `json:"max_stock_level" validate:"min=0"`
	Unit              string    `json:"unit"`
	CostPerUnit       float64   `json:"cost_per_unit" validate:"min=0"`
	IsTrackingStock   bool      `json:"is_tracking_stock"`
	IsAvailable       bool      `json:"is_available"`
	UnavailableReason *string   `json:"unavailable_reason"`
}

// UpdateInventoryRequest represents the request to update inventory
type UpdateInventoryRequest struct {
	StockLevel        *float64 `json:"stock_level"`
	ReorderPoint      *float64 `json:"reorder_point"`
	MaxStockLevel     *float64 `json:"max_stock_level"`
	Unit              *string  `json:"unit"`
	CostPerUnit       *float64 `json:"cost_per_unit"`
	IsTrackingStock   *bool    `json:"is_tracking_stock"`
	IsAvailable       *bool    `json:"is_available"`
	UnavailableReason *string  `json:"unavailable_reason"`
}

// StockAdjustmentRequest represents the request to adjust stock
type StockAdjustmentRequest struct {
	ProductID      uuid.UUID `json:"product_id" validate:"required"`
	LocationID     uuid.UUID `json:"location_id" validate:"required"`
	Quantity       float64   `json:"quantity" validate:"required"`
	AdjustmentType string    `json:"adjustment_type" validate:"required"` // "increase", "decrease", "set"
	Reason         string    `json:"reason"`
	Notes          string    `json:"notes"`
}

// StockReservationRequest represents the request to reserve stock
type StockReservationRequest struct {
	ProductID  uuid.UUID  `json:"product_id" validate:"required"`
	LocationID uuid.UUID  `json:"location_id" validate:"required"`
	Quantity   float64    `json:"quantity" validate:"required,min=0"`
	ReservedBy string     `json:"reserved_by"`
	OrderID    *uuid.UUID `json:"order_id"`
	ExpiresAt  *time.Time `json:"expires_at"`
}

// CreateInventory creates a new inventory record
func (uc *InventoryUsecase) CreateInventory(ctx context.Context, req *CreateInventoryRequest) (*entity.Inventory, error) {
	// Validate request
	if req.ProductID == uuid.Nil {
		return nil, fmt.Errorf("product ID is required")
	}
	if req.LocationID == uuid.Nil {
		return nil, fmt.Errorf("location ID is required")
	}
	if req.StockLevel < 0 {
		return nil, fmt.Errorf("stock level must be non-negative")
	}

	// Check if product exists
	product, err := uc.productRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, fmt.Errorf("failed to check product: %w", err)
	}
	if product == nil {
		return nil, fmt.Errorf("product not found")
	}

	// Check if inventory already exists for this product and location
	existingInventory, err := uc.inventoryRepo.GetByProductAndLocation(ctx, req.ProductID, req.LocationID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing inventory: %w", err)
	}
	if existingInventory != nil {
		return nil, fmt.Errorf("inventory already exists for this product and location")
	}

	// Create new inventory
	inventory := entity.NewInventory(req.ProductID, req.LocationID)

	// Set initial stock level
	if err := inventory.UpdateStockLevel(req.StockLevel); err != nil {
		return nil, fmt.Errorf("failed to set stock level: %w", err)
	}

	// Set optional fields
	inventory.SetThresholds(nil, &req.ReorderPoint, &req.MaxStockLevel)
	inventory.SetAvailability(req.IsAvailable, req.UnavailableReason)

	// Save to database
	if err := uc.inventoryRepo.Create(ctx, inventory); err != nil {
		return nil, fmt.Errorf("failed to save inventory: %w", err)
	}

	uc.logger.WithFields(logrus.Fields{
		"inventory_id": inventory.ID,
		"product_id":   req.ProductID,
		"location_id":  req.LocationID,
	}).Info("Inventory created successfully")

	return inventory, nil
}

// GetInventory retrieves an inventory by ID
func (uc *InventoryUsecase) GetInventory(ctx context.Context, id uuid.UUID) (*entity.Inventory, error) {
	inventory, err := uc.inventoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	if inventory == nil {
		return nil, fmt.Errorf("inventory not found")
	}

	return inventory, nil
}

// GetInventoryByProductAndLocation retrieves inventory for a specific product and location
func (uc *InventoryUsecase) GetInventoryByProductAndLocation(ctx context.Context, productID, locationID uuid.UUID) (*entity.Inventory, error) {
	inventory, err := uc.inventoryRepo.GetByProductAndLocation(ctx, productID, locationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	if inventory == nil {
		return nil, fmt.Errorf("inventory not found")
	}

	return inventory, nil
}

// GetProductInventories retrieves all inventory records for a product
func (uc *InventoryUsecase) GetProductInventories(ctx context.Context, productID uuid.UUID) ([]*entity.Inventory, error) {
	inventories, err := uc.inventoryRepo.GetByProductID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product inventories: %w", err)
	}

	return inventories, nil
}

// GetLocationInventories retrieves all inventory records for a location
func (uc *InventoryUsecase) GetLocationInventories(ctx context.Context, locationID uuid.UUID) ([]*entity.Inventory, error) {
	inventories, err := uc.inventoryRepo.GetByLocationID(ctx, locationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get location inventories: %w", err)
	}

	return inventories, nil
}

// UpdateInventory updates an existing inventory
func (uc *InventoryUsecase) UpdateInventory(ctx context.Context, id uuid.UUID, req *UpdateInventoryRequest) (*entity.Inventory, error) {
	// Get existing inventory
	inventory, err := uc.inventoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	if inventory == nil {
		return nil, fmt.Errorf("inventory not found")
	}

	// Update fields if provided
	if req.StockLevel != nil {
		if *req.StockLevel < 0 {
			return nil, fmt.Errorf("stock level must be non-negative")
		}
		if err := inventory.UpdateStockLevel(*req.StockLevel); err != nil {
			return nil, fmt.Errorf("failed to update stock level: %w", err)
		}
	}

	if req.ReorderPoint != nil || req.MaxStockLevel != nil {
		inventory.SetThresholds(nil, req.ReorderPoint, req.MaxStockLevel)
	}

	if req.IsAvailable != nil {
		inventory.SetAvailability(*req.IsAvailable, req.UnavailableReason)
	}

	// Save changes
	if err := uc.inventoryRepo.Update(ctx, inventory); err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	uc.logger.WithField("inventory_id", inventory.ID).Info("Inventory updated successfully")
	return inventory, nil
}

// DeleteInventory deletes an inventory record
func (uc *InventoryUsecase) DeleteInventory(ctx context.Context, id uuid.UUID) error {
	// Get existing inventory
	inventory, err := uc.inventoryRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get inventory: %w", err)
	}

	if inventory == nil {
		return fmt.Errorf("inventory not found")
	}

	// Check if there are any reserved stocks
	if inventory.ReservedLevel > 0 {
		return fmt.Errorf("cannot delete inventory with reserved stock")
	}

	// Delete inventory
	if err := uc.inventoryRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete inventory: %w", err)
	}

	uc.logger.WithField("inventory_id", id).Info("Inventory deleted successfully")
	return nil
}

// AdjustStock adjusts the stock level for a product at a location
func (uc *InventoryUsecase) AdjustStock(ctx context.Context, req *StockAdjustmentRequest) (*entity.Inventory, error) {
	// Validate request
	if req.ProductID == uuid.Nil {
		return nil, fmt.Errorf("product ID is required")
	}
	if req.LocationID == uuid.Nil {
		return nil, fmt.Errorf("location ID is required")
	}
	if req.AdjustmentType == "" {
		return nil, fmt.Errorf("adjustment type is required")
	}

	// Get existing inventory
	inventory, err := uc.inventoryRepo.GetByProductAndLocation(ctx, req.ProductID, req.LocationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}
	if inventory == nil {
		return nil, fmt.Errorf("inventory not found")
	}

	// Calculate new stock level based on adjustment type
	var newStockLevel float64
	switch req.AdjustmentType {
	case "increase":
		newStockLevel = inventory.StockLevel + req.Quantity
	case "decrease":
		newStockLevel = inventory.StockLevel - req.Quantity
		if newStockLevel < 0 {
			return nil, fmt.Errorf("insufficient stock for adjustment")
		}
	case "set":
		newStockLevel = req.Quantity
		if newStockLevel < 0 {
			return nil, fmt.Errorf("stock level must be non-negative")
		}
	default:
		return nil, fmt.Errorf("invalid adjustment type: %s", req.AdjustmentType)
	}

	// Update stock level
	if err := inventory.UpdateStockLevel(newStockLevel); err != nil {
		return nil, fmt.Errorf("failed to update stock level: %w", err)
	}

	// Note: In a real implementation, you would want to record stock adjustments
	// in a separate StockAdjustment entity or table for audit purposes

	// Save changes
	if err := uc.inventoryRepo.Update(ctx, inventory); err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	uc.logger.WithFields(logrus.Fields{
		"inventory_id":    inventory.ID,
		"adjustment_type": req.AdjustmentType,
		"quantity":        req.Quantity,
		"new_stock_level": newStockLevel,
	}).Info("Stock adjustment completed")

	return inventory, nil
}

// ReserveStock reserves stock for an order
func (uc *InventoryUsecase) ReserveStock(ctx context.Context, req *StockReservationRequest) error {
	// Validate request
	if req.ProductID == uuid.Nil {
		return fmt.Errorf("product ID is required")
	}
	if req.LocationID == uuid.Nil {
		return fmt.Errorf("location ID is required")
	}
	if req.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}

	// Get available stock
	availableStock, err := uc.inventoryRepo.GetAvailableStock(ctx, req.ProductID, req.LocationID)
	if err != nil {
		return fmt.Errorf("failed to check available stock: %w", err)
	}

	if availableStock < req.Quantity {
		return fmt.Errorf("insufficient stock available: %f required, %f available", req.Quantity, availableStock)
	}

	// Reserve stock
	if err := uc.inventoryRepo.ReserveStock(ctx, req.ProductID, req.LocationID, req.Quantity); err != nil {
		return fmt.Errorf("failed to reserve stock: %w", err)
	}

	uc.logger.WithFields(logrus.Fields{
		"product_id":  req.ProductID,
		"location_id": req.LocationID,
		"quantity":    req.Quantity,
		"reserved_by": req.ReservedBy,
	}).Info("Stock reserved successfully")

	return nil
}

// ReleaseStock releases reserved stock
func (uc *InventoryUsecase) ReleaseStock(ctx context.Context, productID, locationID uuid.UUID, quantity float64) error {
	// Validate parameters
	if productID == uuid.Nil {
		return fmt.Errorf("product ID is required")
	}
	if locationID == uuid.Nil {
		return fmt.Errorf("location ID is required")
	}
	if quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}

	// Release stock
	if err := uc.inventoryRepo.ReleaseStock(ctx, productID, locationID, quantity); err != nil {
		return fmt.Errorf("failed to release stock: %w", err)
	}

	uc.logger.WithFields(logrus.Fields{
		"product_id":  productID,
		"location_id": locationID,
		"quantity":    quantity,
	}).Info("Stock released successfully")

	return nil
}

// GetAvailableStock gets available stock for a product at a location
func (uc *InventoryUsecase) GetAvailableStock(ctx context.Context, productID, locationID uuid.UUID) (float64, error) {
	availableStock, err := uc.inventoryRepo.GetAvailableStock(ctx, productID, locationID)
	if err != nil {
		return 0, fmt.Errorf("failed to get available stock: %w", err)
	}

	return availableStock, nil
}

// GetProductAvailability gets availability status for a product across all locations
func (uc *InventoryUsecase) GetProductAvailability(ctx context.Context, productID uuid.UUID) (*entity.ProductAvailability, error) {
	availability, err := uc.inventoryRepo.GetAvailability(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product availability: %w", err)
	}

	return availability, nil
}

// GetLocationAvailability gets availability status for a product at a specific location
func (uc *InventoryUsecase) GetLocationAvailability(ctx context.Context, productID, locationID uuid.UUID) (*entity.LocationAvailability, error) {
	availability, err := uc.inventoryRepo.GetAvailabilityByLocation(ctx, productID, locationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get location availability: %w", err)
	}

	return availability, nil
}

// SetAvailability sets availability status for a product at a location
func (uc *InventoryUsecase) SetAvailability(ctx context.Context, productID, locationID uuid.UUID, available bool, reason *string) error {
	if err := uc.inventoryRepo.SetAvailability(ctx, productID, locationID, available, reason); err != nil {
		return fmt.Errorf("failed to set availability: %w", err)
	}

	uc.logger.WithFields(logrus.Fields{
		"product_id":  productID,
		"location_id": locationID,
		"available":   available,
		"reason":      reason,
	}).Info("Availability updated successfully")

	return nil
}

// GetLowStockItems gets items with low stock at a location
func (uc *InventoryUsecase) GetLowStockItems(ctx context.Context, locationID uuid.UUID) ([]*entity.Inventory, error) {
	items, err := uc.inventoryRepo.GetLowStockItems(ctx, locationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get low stock items: %w", err)
	}

	return items, nil
}

// GetOutOfStockItems gets items that are out of stock at a location
func (uc *InventoryUsecase) GetOutOfStockItems(ctx context.Context, locationID uuid.UUID) ([]*entity.Inventory, error) {
	items, err := uc.inventoryRepo.GetOutOfStockItems(ctx, locationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get out of stock items: %w", err)
	}

	return items, nil
}

// GetInventoryValue calculates the total value of inventory at a location
func (uc *InventoryUsecase) GetInventoryValue(ctx context.Context, locationID uuid.UUID) (float64, error) {
	totalValue, err := uc.inventoryRepo.GetTotalValue(ctx, locationID)
	if err != nil {
		return 0, fmt.Errorf("failed to get inventory value: %w", err)
	}

	return totalValue, nil
}

// GetTurnoverRate calculates the turnover rate for a product
func (uc *InventoryUsecase) GetTurnoverRate(ctx context.Context, productID uuid.UUID, days int) (float64, error) {
	if days <= 0 {
		return 0, fmt.Errorf("days must be positive")
	}

	turnoverRate, err := uc.inventoryRepo.GetTurnoverRate(ctx, productID, days)
	if err != nil {
		return 0, fmt.Errorf("failed to get turnover rate: %w", err)
	}

	return turnoverRate, nil
}

// CheckReorderPoints checks if any products need reordering at a location
func (uc *InventoryUsecase) CheckReorderPoints(ctx context.Context, locationID uuid.UUID) ([]*entity.Inventory, error) {
	// Get all inventories for the location
	inventories, err := uc.inventoryRepo.GetByLocationID(ctx, locationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventories: %w", err)
	}

	var reorderItems []*entity.Inventory
	for _, inventory := range inventories {
		if inventory.NeedsReorder() {
			reorderItems = append(reorderItems, inventory)
		}
	}

	return reorderItems, nil
}
