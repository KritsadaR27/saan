package application

import (
	"context"
	"fmt"
	"time"

	"product-service/internal/domain/entity"
	"product-service/internal/domain/repository"
	"product-service/internal/infrastructure/events"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// SyncUsecase handles synchronization logic
type SyncUsecase struct {
	productRepo  repository.ProductRepository
	categoryRepo repository.CategoryRepository
	eventPub     events.Publisher
	logger       *logrus.Logger
}

// NewSyncUsecase creates a new sync usecase
func NewSyncUsecase(
	productRepo repository.ProductRepository,
	categoryRepo repository.CategoryRepository,
	eventPub events.Publisher,
	logger *logrus.Logger,
) *SyncUsecase {
	return &SyncUsecase{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		eventPub:     eventPub,
		logger:       logger,
	}
}

// SyncProductRequest represents a product sync request from Loyverse
type SyncProductRequest struct {
	LoyverseID  string                 `json:"loyverse_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	SKU         string                 `json:"sku"`
	Barcode     *string                `json:"barcode"`
	CategoryID  *string                `json:"category_id"`
	BasePrice   float64                `json:"base_price"`
	CostPrice   *float64               `json:"cost_price"`
	Unit        string                 `json:"unit"`
	Status      string                 `json:"status"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SyncCategoryRequest represents a category sync request from Loyverse
type SyncCategoryRequest struct {
	LoyverseID  string  `json:"loyverse_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	ParentID    *string `json:"parent_id,omitempty"`
	Status      string  `json:"status"`
}

// SyncResponse represents the result of a sync operation
type SyncResponse struct {
	SyncID       string    `json:"sync_id"`
	SyncType     string    `json:"sync_type"`
	Status       string    `json:"status"`
	RecordsTotal int       `json:"records_total"`
	RecordsSync  int       `json:"records_synced"`
	RecordsFail  int       `json:"records_failed"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// SyncProductsFromLoyverse syncs products from Loyverse API
func (uc *SyncUsecase) SyncProductsFromLoyverse(ctx context.Context, products []SyncProductRequest) (*SyncResponse, error) {
	syncID := uuid.New().String()
	startTime := time.Now()
	
	uc.logger.WithFields(logrus.Fields{
		"sync_id":       syncID,
		"product_count": len(products),
	}).Info("Starting product sync from Loyverse")

	response := &SyncResponse{
		SyncID:       syncID,
		SyncType:     "loyverse_products",
		Status:       "in_progress",
		RecordsTotal: len(products),
		StartTime:    startTime,
	}

	var syncedCount, failedCount int
	var lastError error

	for _, productReq := range products {
		err := uc.syncSingleProduct(ctx, productReq, syncID)
		if err != nil {
			failedCount++
			lastError = err
			uc.logger.WithError(err).WithFields(logrus.Fields{
				"loyverse_id": productReq.LoyverseID,
				"sku":         productReq.SKU,
			}).Error("Failed to sync product")
		} else {
			syncedCount++
		}
	}

	// Update response
	response.EndTime = time.Now()
	response.RecordsSync = syncedCount
	response.RecordsFail = failedCount
	
	if failedCount > 0 {
		response.Status = "partial_success"
		if lastError != nil {
			response.ErrorMessage = lastError.Error()
		}
	} else {
		response.Status = "completed"
	}

	// Publish sync event
	syncEvent := events.NewSyncEvent(
		events.LoyverseSyncCompletedEvent,
		"loyverse_products",
		"loyverse",
		response.RecordsTotal,
		response.RecordsSync,
		response.RecordsFail,
		response.Status,
		response.ErrorMessage,
	)

	if err := uc.eventPub.PublishSyncEvent(ctx, syncEvent); err != nil {
		uc.logger.WithError(err).Error("Failed to publish sync event")
	}

	uc.logger.WithFields(logrus.Fields{
		"sync_id":        syncID,
		"records_synced": syncedCount,
		"records_failed": failedCount,
		"duration":       response.EndTime.Sub(response.StartTime),
	}).Info("Product sync completed")

	return response, nil
}

// syncSingleProduct syncs a single product using Master Data Protection pattern
func (uc *SyncUsecase) syncSingleProduct(ctx context.Context, req SyncProductRequest, syncID string) error {
	// Check if product exists by Loyverse ID
	existing, err := uc.productRepo.GetByLoyverseID(ctx, req.LoyverseID)
	if err != nil {
		return fmt.Errorf("failed to check existing product: %w", err)
	}

	if existing != nil {
		// Update existing product - only Loyverse-controlled fields
		return uc.updateProductFromLoyverse(ctx, existing, req, syncID)
	} else {
		// Create new product
		return uc.createProductFromLoyverse(ctx, req, syncID)
	}
}

// updateProductFromLoyverse updates existing product with Loyverse data (Master Data Protection)
func (uc *SyncUsecase) updateProductFromLoyverse(ctx context.Context, existing *entity.Product, req SyncProductRequest, syncID string) error {
	// Store old values for event
	oldName := existing.Name
	oldPrice := existing.BasePrice
	
	// Update only Loyverse-controlled fields (following MASTER_DATA_PROTECTION_PATTERN.md)
	existing.Name = req.Name
	existing.Description = req.Description
	existing.SKU = req.SKU
	existing.BasePrice = req.BasePrice
	existing.Unit = req.Unit
	
	if req.Barcode != nil {
		existing.Barcode = req.Barcode
	}
	
	// Parse category if provided
	if req.CategoryID != nil {
		if categoryUUID, err := uuid.Parse(*req.CategoryID); err == nil {
			existing.CategoryID = &categoryUUID
		}
	}
	
	// Update sync metadata
	existing.DataSourceType = "loyverse"
	existing.DataSourceID = &req.LoyverseID
	existing.LoyverseID = &req.LoyverseID
	existing.LastSyncedAt = &[]time.Time{time.Now()}[0]
	existing.UpdatedAt = time.Now()

	// Save to database
	if err := uc.productRepo.Update(ctx, existing); err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	// Publish product updated event
	changes := map[string]interface{}{
		"sync_id":     syncID,
		"old_name":    oldName,
		"new_name":    req.Name,
		"old_price":   oldPrice,
		"new_price":   req.BasePrice,
		"source":      "loyverse_sync",
	}

	productEvent := events.NewProductEvent(
		events.ProductUpdatedEvent,
		existing.ID,
		existing.SKU,
		existing.Name,
		"loyverse_sync",
		changes,
	)

	if err := uc.eventPub.PublishProductEvent(ctx, productEvent); err != nil {
		uc.logger.WithError(err).Error("Failed to publish product updated event")
	}

	return nil
}

// createProductFromLoyverse creates a new product from Loyverse data
func (uc *SyncUsecase) createProductFromLoyverse(ctx context.Context, req SyncProductRequest, syncID string) error {
	// Create new product entity
	product, err := entity.NewProduct(req.Name, req.SKU, req.Unit, req.BasePrice)
	if err != nil {
		return fmt.Errorf("failed to create product entity: %w", err)
	}

	// Set Loyverse-specific fields
	product.Description = req.Description
	product.DataSourceType = "loyverse"
	product.DataSourceID = &req.LoyverseID
	product.LoyverseID = &req.LoyverseID
	product.LastSyncedAt = &[]time.Time{time.Now()}[0]
	
	if req.Barcode != nil {
		product.Barcode = req.Barcode
	}
	
	// Parse category if provided
	if req.CategoryID != nil {
		if categoryUUID, err := uuid.Parse(*req.CategoryID); err == nil {
			product.CategoryID = &categoryUUID
		}
	}

	// Set cost price if provided
	if req.CostPrice != nil {
		// Note: CostPrice would need to be added to entity.Product if not already there
		// For now, we'll store it in metadata or extend the entity
	}

	// Save to database
	if err := uc.productRepo.Create(ctx, product); err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	// Publish product created event
	changes := map[string]interface{}{
		"sync_id": syncID,
		"source":  "loyverse_sync",
	}

	productEvent := events.NewProductEvent(
		events.ProductCreatedEvent,
		product.ID,
		product.SKU,
		product.Name,
		"loyverse_sync",
		changes,
	)

	if err := uc.eventPub.PublishProductEvent(ctx, productEvent); err != nil {
		uc.logger.WithError(err).Error("Failed to publish product created event")
	}

	return nil
}

// SyncCategoriesFromLoyverse syncs categories from Loyverse API
func (uc *SyncUsecase) SyncCategoriesFromLoyverse(ctx context.Context, categories []SyncCategoryRequest) (*SyncResponse, error) {
	syncID := uuid.New().String()
	startTime := time.Now()
	
	uc.logger.WithFields(logrus.Fields{
		"sync_id":        syncID,
		"category_count": len(categories),
	}).Info("Starting category sync from Loyverse")

	response := &SyncResponse{
		SyncID:       syncID,
		SyncType:     "loyverse_categories",
		Status:       "in_progress",
		RecordsTotal: len(categories),
		StartTime:    startTime,
	}

	var syncedCount, failedCount int
	var lastError error

	for _, categoryReq := range categories {
		err := uc.syncSingleCategory(ctx, categoryReq, syncID)
		if err != nil {
			failedCount++
			lastError = err
			uc.logger.WithError(err).WithField("loyverse_id", categoryReq.LoyverseID).Error("Failed to sync category")
		} else {
			syncedCount++
		}
	}

	// Update response
	response.EndTime = time.Now()
	response.RecordsSync = syncedCount
	response.RecordsFail = failedCount
	
	if failedCount > 0 {
		response.Status = "partial_success"
		if lastError != nil {
			response.ErrorMessage = lastError.Error()
		}
	} else {
		response.Status = "completed"
	}

	// Publish sync event
	syncEvent := events.NewSyncEvent(
		events.LoyverseSyncCompletedEvent,
		"loyverse_categories",
		"loyverse",
		response.RecordsTotal,
		response.RecordsSync,
		response.RecordsFail,
		response.Status,
		response.ErrorMessage,
	)

	if err := uc.eventPub.PublishSyncEvent(ctx, syncEvent); err != nil {
		uc.logger.WithError(err).Error("Failed to publish sync event")
	}

	return response, nil
}

// syncSingleCategory syncs a single category
func (uc *SyncUsecase) syncSingleCategory(ctx context.Context, req SyncCategoryRequest, syncID string) error {
	// Check if category exists by Loyverse ID
	existing, err := uc.categoryRepo.GetByLoyverseID(ctx, req.LoyverseID)
	if err != nil {
		return fmt.Errorf("failed to check existing category: %w", err)
	}

	if existing != nil {
		// Update existing category
		existing.Name = req.Name
		existing.Description = req.Description
		existing.DataSourceType = "loyverse"
		existing.DataSourceID = &req.LoyverseID
		existing.LoyverseID = &req.LoyverseID
		existing.LastSyncedAt = &[]time.Time{time.Now()}[0]
		existing.UpdatedAt = time.Now()

		// Parse parent ID if provided
		if req.ParentID != nil {
			if parentUUID, err := uuid.Parse(*req.ParentID); err == nil {
				existing.ParentID = &parentUUID
			}
		}

		return uc.categoryRepo.Update(ctx, existing)
	} else {
		// Create new category
		category, err := entity.NewCategory(req.Name)
		if err != nil {
			return fmt.Errorf("failed to create category entity: %w", err)
		}

		category.Description = req.Description
		category.DataSourceType = "loyverse"
		category.DataSourceID = &req.LoyverseID
		category.LoyverseID = &req.LoyverseID
		category.LastSyncedAt = &[]time.Time{time.Now()}[0]

		// Parse parent ID if provided
		if req.ParentID != nil {
			if parentUUID, err := uuid.Parse(*req.ParentID); err == nil {
				category.ParentID = &parentUUID
			}
		}

		return uc.categoryRepo.Create(ctx, category)
	}
}

// GetSyncStatus returns the status of sync operations
func (uc *SyncUsecase) GetSyncStatus(ctx context.Context, syncID string) (*SyncResponse, error) {
	// This would typically query a sync_history table
	// For now, return a placeholder
	return &SyncResponse{
		SyncID:   syncID,
		Status:   "completed",
		SyncType: "unknown",
	}, nil
}

// GetLastSyncTime returns the last sync time for a specific sync type
func (uc *SyncUsecase) GetLastSyncTime(ctx context.Context, syncType string) (time.Time, error) {
	// This would typically query the last successful sync from sync_history table
	// For now, return a placeholder
	return time.Now().Add(-24 * time.Hour), nil
}
