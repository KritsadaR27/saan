package loyverse

import (
	"context"
	"fmt"
	"time"

	"product/internal/application"
	"product/internal/infrastructure/events"

	"github.com/sirupsen/logrus"
)

// SyncService handles synchronization with Loyverse
type SyncService struct {
	client      *Client
	syncUsecase *application.SyncUsecase
	eventPub    events.Publisher
	logger      *logrus.Logger
}

// NewSyncService creates a new Loyverse sync service
func NewSyncService(
	client *Client,
	syncUsecase *application.SyncUsecase,
	eventPub events.Publisher,
	logger *logrus.Logger,
) *SyncService {
	return &SyncService{
		client:      client,
		syncUsecase: syncUsecase,
		eventPub:    eventPub,
		logger:      logger,
	}
}

// SyncAllProducts syncs all products from Loyverse
func (s *SyncService) SyncAllProducts(ctx context.Context) (*SyncResult, error) {
	startTime := time.Now()
	result := &SyncResult{
		StartTime: startTime,
		Errors:    make([]string, 0),
	}

	s.logger.Info("Starting full product sync from Loyverse")

	// First sync categories
	categoriesProcessed, err := s.syncCategories(ctx)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Category sync failed: %v", err))
		s.logger.WithError(err).Error("Failed to sync categories")
	}
	result.CategoriesProcessed = categoriesProcessed

	// Then sync products
	productsProcessed, err := s.syncProducts(ctx)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Product sync failed: %v", err))
		s.logger.WithError(err).Error("Failed to sync products")
	}
	result.ProductsProcessed = productsProcessed

	// Calculate duration
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	// Publish sync completion event
	if err := s.publishSyncEvent(ctx, result); err != nil {
		s.logger.WithError(err).Error("Failed to publish sync completion event")
		result.Errors = append(result.Errors, fmt.Sprintf("Event publishing failed: %v", err))
	}

	s.logger.WithFields(logrus.Fields{
		"products_processed":   result.ProductsProcessed,
		"categories_processed": result.CategoriesProcessed,
		"duration":             result.Duration.String(),
		"errors_count":         len(result.Errors),
	}).Info("Loyverse sync completed")

	return result, nil
}

// SyncProduct syncs a single product by Loyverse ID
func (s *SyncService) SyncProduct(ctx context.Context, loyverseProductID string) error {
	s.logger.WithField("loyverse_product_id", loyverseProductID).Info("Syncing single product from Loyverse")

	// Fetch product from Loyverse
	loyverseProduct, err := s.client.GetProduct(ctx, loyverseProductID)
	if err != nil {
		return fmt.Errorf("failed to fetch product from Loyverse: %w", err)
	}		// Convert and sync each variant as a separate product
		var syncRequests []application.SyncProductRequest
		for _, variant := range loyverseProduct.Variants {
			syncRequest := s.convertVariantToSyncRequest(loyverseProduct, variant)
			syncRequests = append(syncRequests, *syncRequest)
		}
		
		if len(syncRequests) > 0 {
			_, err := s.syncUsecase.SyncProductsFromLoyverse(ctx, syncRequests)
			if err != nil {
				s.logger.WithError(err).WithField("loyverse_product_id", loyverseProductID).Error("Failed to sync product variants")
				return fmt.Errorf("failed to sync product variants: %w", err)
			}
		}

	s.logger.WithField("loyverse_product_id", loyverseProductID).Info("Successfully synced product from Loyverse")
	return nil
}

// syncCategories syncs all categories from Loyverse
func (s *SyncService) syncCategories(ctx context.Context) (int, error) {
	s.logger.Info("Syncing categories from Loyverse")

	// Fetch categories from Loyverse
	categoriesResp, err := s.client.GetCategories(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch categories: %w", err)
	}

	processed := 0
	var syncRequests []application.SyncCategoryRequest
	for _, loyverseCategory := range categoriesResp.Categories {
		syncRequest := s.convertCategoryToSyncRequest(loyverseCategory)
		syncRequests = append(syncRequests, *syncRequest)
	}

	if len(syncRequests) > 0 {
		_, err = s.syncUsecase.SyncCategoriesFromLoyverse(ctx, syncRequests)
		if err != nil {
			s.logger.WithError(err).Error("Failed to sync categories batch")
			return 0, fmt.Errorf("failed to sync categories: %w", err)
		}
		processed = len(syncRequests)
	}

	s.logger.WithField("categories_processed", processed).Info("Categories sync completed")
	return processed, nil
}

// syncProducts syncs all products from Loyverse with pagination
func (s *SyncService) syncProducts(ctx context.Context) (int, error) {
	s.logger.Info("Syncing products from Loyverse")

	var cursor string
	processed := 0
	const batchSize = 100 // Loyverse API limit

	for {
		// Fetch batch of products
		productsResp, err := s.client.GetProducts(ctx, cursor, batchSize)
		if err != nil {
			return processed, fmt.Errorf("failed to fetch products: %w", err)
		}

		// Process each product
		var batchRequests []application.SyncProductRequest
		for _, loyverseProduct := range productsResp.Products {
			// Each variant becomes a separate product in our system
			for _, variant := range loyverseProduct.Variants {
				syncRequest := s.convertVariantToSyncRequest(&loyverseProduct, variant)
				batchRequests = append(batchRequests, *syncRequest)
			}
		}

		if len(batchRequests) > 0 {
			_, err = s.syncUsecase.SyncProductsFromLoyverse(ctx, batchRequests)
			if err != nil {
				s.logger.WithError(err).Error("Failed to sync products batch")
				// Continue with next batch instead of failing completely
			} else {
				processed += len(batchRequests)
			}
		}

		// Check if there are more pages
		if productsResp.Cursor == "" {
			break
		}
		cursor = productsResp.Cursor

		s.logger.WithFields(logrus.Fields{
			"processed_so_far": processed,
			"next_cursor":      cursor,
		}).Debug("Processed batch of products")
	}

	s.logger.WithField("products_processed", processed).Info("Products sync completed")
	return processed, nil
}

// convertVariantToSyncRequest converts a Loyverse variant to a sync request
func (s *SyncService) convertVariantToSyncRequest(product *LoyverseProduct, variant LoyverseVariant) *application.SyncProductRequest {
	// Build product name with variant options
	name := product.ItemName
	if variant.Option1Value != "" {
		name += " - " + variant.Option1Value
	}
	if variant.Option2Value != "" {
		name += " - " + variant.Option2Value
	}
	if variant.Option3Value != "" {
		name += " - " + variant.Option3Value
	}

	// Convert status
	status := "active"
	if len(variant.Stores) == 0 {
		status = "inactive"
	} else {
		// Check if available in any store
		available := false
		for _, store := range variant.Stores {
			if store.Available {
				available = true
				break
			}
		}
		if !available {
			status = "inactive"
		}
	}

	// Handle barcode
	var barcode *string
	if variant.Barcode != "" {
		barcode = &variant.Barcode
	}

	// Handle category
	var categoryID *string
	if product.CategoryID != "" {
		categoryID = &product.CategoryID
	}

	// Handle cost price
	var costPrice *float64
	if variant.Cost > 0 {
		costPrice = &variant.Cost
	}

	return &application.SyncProductRequest{
		LoyverseID:  variant.ID, // Use variant ID as the unique identifier
		Name:        name,
		Description: product.Description,
		SKU:         variant.SKU,
		Barcode:     barcode,
		CategoryID:  categoryID,
		BasePrice:   variant.DefaultPrice,
		CostPrice:   costPrice,
		Unit:        "pcs", // Default unit, can be configured
		Status:      status,
		Metadata: map[string]interface{}{
			"loyverse_product_id": product.ID,
			"image_url":           product.ImageURL,
			"updated_at":          variant.UpdatedAt,
		},
	}
}

// convertCategoryToSyncRequest converts a Loyverse category to a sync request
func (s *SyncService) convertCategoryToSyncRequest(category LoyverseCategory) *application.SyncCategoryRequest {
	return &application.SyncCategoryRequest{
		LoyverseID:  category.ID,
		Name:        category.Name,
		Description: fmt.Sprintf("Category synced from Loyverse (Color: %s)", category.Color),
		Status:      "active",
	}
}

// publishSyncEvent publishes a sync completion event
func (s *SyncService) publishSyncEvent(ctx context.Context, result *SyncResult) error {
	status := "completed"
	errorMessage := ""
	if len(result.Errors) > 0 {
		status = "failed"
		errorMessage = fmt.Sprintf("%d errors occurred during sync", len(result.Errors))
	}

	event := &events.SyncEvent{
		BaseEvent: events.BaseEvent{
			ID:        fmt.Sprintf("sync_%d", time.Now().Unix()),
			Type:      events.LoyverseSyncCompletedEvent,
			Source:    "product-service",
			Timestamp: time.Now(),
			Version:   "1.0",
		},
		SyncType:     "loyverse_full_sync",
		SourceSystem: "loyverse",
		RecordsCount: result.ProductsProcessed + result.CategoriesProcessed,
		SuccessCount: result.ProductsProcessed + result.CategoriesProcessed - len(result.Errors),
		FailureCount: len(result.Errors),
		Status:       status,
		ErrorMessage: errorMessage,
	}

	if len(result.Errors) > 0 {
		event.Type = events.SyncFailedEvent
	}

	return s.eventPub.PublishSyncEvent(ctx, event)
}
