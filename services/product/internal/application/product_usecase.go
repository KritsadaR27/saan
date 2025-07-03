package application

import (
	"context"
	"fmt"
	"time"

	"product-service/internal/domain/entity"
	"product-service/internal/domain/repository"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ProductUsecase handles product business logic
type ProductUsecase struct {
	productRepo repository.ProductRepository
	cache       CacheRepository
	logger      *logrus.Logger
}

// CacheRepository interface for cache operations
type CacheRepository interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string) (interface{}, error)
	Delete(ctx context.Context, key string) error
	GetProduct(ctx context.Context, productID uuid.UUID) (*entity.Product, error)
	SetProduct(ctx context.Context, productID uuid.UUID, product *entity.Product, ttl time.Duration) error
	InvalidateProduct(ctx context.Context, productID uuid.UUID) error
}

// NewProductUsecase creates a new product usecase
func NewProductUsecase(productRepo repository.ProductRepository, cache CacheRepository, logger *logrus.Logger) *ProductUsecase {
	return &ProductUsecase{
		productRepo: productRepo,
		cache:       cache,
		logger:      logger,
	}
}

// CreateProductRequest represents the request to create a product
type CreateProductRequest struct {
	Name         string                      `json:"name" validate:"required"`
	Description  string                      `json:"description"`
	SKU          string                      `json:"sku" validate:"required"`
	Barcode      *string                     `json:"barcode"`
	CategoryID   *uuid.UUID                  `json:"category_id"`
	BasePrice    float64                     `json:"base_price" validate:"required,min=0"`
	Unit         string                      `json:"unit" validate:"required"`
	Weight       *float64                    `json:"weight"`
	Dimensions   *entity.ProductDimensions   `json:"dimensions"`
	Tags         []string                    `json:"tags"`
	IsVIPOnly    bool                        `json:"is_vip_only"`
}

// UpdateProductRequest represents the request to update a product
type UpdateProductRequest struct {
	Name         *string                     `json:"name"`
	Description  *string                     `json:"description"`
	BasePrice    *float64                    `json:"base_price"`
	Weight       *float64                    `json:"weight"`
	Dimensions   *entity.ProductDimensions   `json:"dimensions"`
	Tags         []string                    `json:"tags"`
	IsVIPOnly    *bool                       `json:"is_vip_only"`
	IsActive     *bool                       `json:"is_active"`
}

// CreateProduct creates a new product
func (uc *ProductUsecase) CreateProduct(ctx context.Context, req *CreateProductRequest) (*entity.Product, error) {
	// Validate request
	if req.Name == "" {
		return nil, fmt.Errorf("product name is required")
	}
	if req.SKU == "" {
		return nil, fmt.Errorf("product SKU is required")
	}
	if req.Unit == "" {
		return nil, fmt.Errorf("product unit is required")
	}
	if req.BasePrice < 0 {
		return nil, fmt.Errorf("base price must be non-negative")
	}

	// Check for duplicate SKU
	existing, err := uc.productRepo.GetBySKU(ctx, req.SKU)
	if err != nil {
		return nil, fmt.Errorf("failed to check duplicate SKU: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("product with SKU '%s' already exists", req.SKU)
	}

	// Create new product
	product, err := entity.NewProduct(req.Name, req.SKU, req.Unit, req.BasePrice)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	// Set optional fields
	product.Description = req.Description
	product.Barcode = req.Barcode
	product.CategoryID = req.CategoryID
	product.Weight = req.Weight
	product.Dimensions = req.Dimensions
	product.Tags = req.Tags
	product.IsVIPOnly = req.IsVIPOnly

	// Save to database
	if err := uc.productRepo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to save product: %w", err)
	}

	// Don't cache newly created products automatically
	// Cache only when they become "hot" (frequently accessed)
	
	uc.logger.WithField("product_id", product.ID).Info("Product created successfully")
	return product, nil
}

// GetProduct retrieves a product by ID
func (uc *ProductUsecase) GetProduct(ctx context.Context, id uuid.UUID) (*entity.Product, error) {
	// Get from database (master data - no caching by default)
	product, err := uc.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	
	if product == nil {
		return nil, fmt.Errorf("product not found")
	}

	// Only cache "hot" products (frequently accessed)
	// This would be implemented based on business rules
	// For now, we don't cache by default (following PROJECT_RULES.md)
	
	return product, nil
}

// GetHotProduct retrieves a frequently accessed product with caching
// Following PROJECT_RULES.md: "product:hot:{product_id} → frequently accessed products"
func (uc *ProductUsecase) GetHotProduct(ctx context.Context, id uuid.UUID) (*entity.Product, error) {
	// Try to get from cache first for hot products
	if uc.cache != nil {
		cachedProduct, err := uc.cache.GetProduct(ctx, id)
		if err == nil && cachedProduct != nil {
			uc.logger.WithField("product_id", id).Debug("Hot product retrieved from cache")
			return cachedProduct, nil
		}
	}

	// Get from database
	product, err := uc.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get hot product: %w", err)
	}
	
	if product == nil {
		return nil, fmt.Errorf("product not found")
	}

	// Cache hot products
	if uc.cache != nil {
		if err := uc.cache.SetProduct(ctx, id, product, 1*time.Hour); err != nil {
			uc.logger.WithError(err).WithField("product_id", id).Warn("Failed to cache hot product")
		}
	}

	return product, nil
}

// GetProductBySKU retrieves a product by SKU
func (uc *ProductUsecase) GetProductBySKU(ctx context.Context, sku string) (*entity.Product, error) {
	product, err := uc.productRepo.GetBySKU(ctx, sku)
	if err != nil {
		return nil, fmt.Errorf("failed to get product by SKU: %w", err)
	}
	
	if product == nil {
		return nil, fmt.Errorf("product not found")
	}

	return product, nil
}

// UpdateProduct updates an existing product
func (uc *ProductUsecase) UpdateProduct(ctx context.Context, id uuid.UUID, req *UpdateProductRequest) (*entity.Product, error) {
	// Get existing product
	product, err := uc.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	
	if product == nil {
		return nil, fmt.Errorf("product not found")
	}

	// Check if product can be modified
	if !product.CanBeModified() {
		return nil, fmt.Errorf("product cannot be modified due to master data protection")
	}

	// Update fields if provided
	if req.Name != nil {
		if err := product.UpdateName(*req.Name); err != nil {
			return nil, fmt.Errorf("failed to update name: %w", err)
		}
	}

	if req.Description != nil {
		product.Description = *req.Description
		product.UpdatedAt = time.Now()
	}

	if req.BasePrice != nil {
		if err := product.UpdatePrice(*req.BasePrice); err != nil {
			return nil, fmt.Errorf("failed to update price: %w", err)
		}
	}

	if req.Weight != nil {
		product.Weight = req.Weight
		product.UpdatedAt = time.Now()
	}

	if req.Dimensions != nil {
		product.Dimensions = req.Dimensions
		product.UpdatedAt = time.Now()
	}

	if req.Tags != nil {
		product.Tags = req.Tags
		product.UpdatedAt = time.Now()
	}

	if req.IsVIPOnly != nil {
		product.SetVIPOnly(*req.IsVIPOnly)
	}

	if req.IsActive != nil {
		if *req.IsActive {
			product.Activate()
		} else {
			product.Deactivate()
		}
	}

	// Save changes
	if err := uc.productRepo.Update(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	// Only invalidate cache if product was actually cached (hot products)
	// Most products won't be cached according to PROJECT_RULES.md
	if uc.cache != nil {
		if err := uc.cache.InvalidateProduct(ctx, product.ID); err != nil {
			uc.logger.WithError(err).WithField("product_id", product.ID).Debug("Cache invalidation failed (product may not have been cached)")
		}
	}

	uc.logger.WithField("product_id", product.ID).Info("Product updated successfully")
	return product, nil
}

// DeleteProduct deletes a product
func (uc *ProductUsecase) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	// Get existing product
	product, err := uc.productRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}
	
	if product == nil {
		return fmt.Errorf("product not found")
	}

	// Check if product can be modified
	if !product.CanBeModified() {
		return fmt.Errorf("product cannot be deleted due to master data protection")
	}

	// Delete product
	if err := uc.productRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	// Clean up cache if product was cached
	if uc.cache != nil {
		if err := uc.cache.InvalidateProduct(ctx, id); err != nil {
			uc.logger.WithError(err).WithField("product_id", id).Debug("Cache cleanup failed (product may not have been cached)")
		}
	}

	uc.logger.WithField("product_id", id).Info("Product deleted successfully")
	return nil
}

// ListProducts lists products with filtering
func (uc *ProductUsecase) ListProducts(ctx context.Context, filter repository.ProductFilter) ([]*entity.Product, error) {
	products, err := uc.productRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	return products, nil
}

// SearchProducts searches products
func (uc *ProductUsecase) SearchProducts(ctx context.Context, query string, filter repository.ProductFilter) ([]*entity.Product, error) {
	products, err := uc.productRepo.Search(ctx, query, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	return products, nil
}

// SetManualOverride sets manual override for a product
func (uc *ProductUsecase) SetManualOverride(ctx context.Context, id uuid.UUID, override bool) error {
	// Get existing product
	product, err := uc.productRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}
	
	if product == nil {
		return fmt.Errorf("product not found")
	}

	// Set manual override
	product.SetManualOverride(override)

	// Save changes
	if err := uc.productRepo.Update(ctx, product); err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	uc.logger.WithFields(logrus.Fields{
		"product_id": id,
		"override":   override,
	}).Info("Product manual override updated")

	return nil
}

// GetFeaturedProducts retrieves featured products with caching
// Following PROJECT_RULES.md: "product:featured → featured products list"
func (uc *ProductUsecase) GetFeaturedProducts(ctx context.Context) ([]*entity.Product, error) {
	// Try to get from cache first
	if uc.cache != nil {
		cacheKey := "product:featured"
		cachedData, err := uc.cache.Get(ctx, cacheKey)
		if err == nil && cachedData != nil {
			uc.logger.Debug("Featured products retrieved from cache")
			// Type assertion would be needed here based on cache implementation
			// For now, fall through to database query
		}
	}

	// Get from database
	filter := repository.ProductFilter{
		IsActive: &[]bool{true}[0],
		Tags:     []string{"featured"}, // Use tags to identify featured products
		Limit:    20,                   // Reasonable limit for featured products
	}

	products, err := uc.productRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get featured products: %w", err)
	}

	// Cache featured products with shorter TTL
	if uc.cache != nil {
		cacheKey := "product:featured"
		if err := uc.cache.Set(ctx, cacheKey, products, 30*time.Minute); err != nil {
			uc.logger.WithError(err).Warn("Failed to cache featured products")
		}
	}

	return products, nil
}

// PromoteToHot marks a product as hot (frequently accessed) for caching
func (uc *ProductUsecase) PromoteToHot(ctx context.Context, id uuid.UUID) error {
	product, err := uc.GetProduct(ctx, id)
	if err != nil {
		return err
	}

	// Cache the product as hot
	if uc.cache != nil {
		if err := uc.cache.SetProduct(ctx, id, product, 1*time.Hour); err != nil {
			uc.logger.WithError(err).WithField("product_id", id).Warn("Failed to promote product to hot cache")
			return err
		}
	}

	uc.logger.WithField("product_id", id).Info("Product promoted to hot cache")
	return nil
}
