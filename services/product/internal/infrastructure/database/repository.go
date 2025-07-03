package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"product-service/internal/domain/entity"
	"product-service/internal/domain/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// productRepository implements the ProductRepository interface
type productRepository struct {
	db *gorm.DB
}

// NewProductRepository creates a new product repository
func NewProductRepository(db *gorm.DB) repository.ProductRepository {
	return &productRepository{db: db}
}

// Create creates a new product
func (r *productRepository) Create(ctx context.Context, product *entity.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

// GetByID retrieves a product by ID
func (r *productRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Product, error) {
	var product entity.Product
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&product).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

// GetBySKU retrieves a product by SKU
func (r *productRepository) GetBySKU(ctx context.Context, sku string) (*entity.Product, error) {
	var product entity.Product
	err := r.db.WithContext(ctx).Where("sku = ?", sku).First(&product).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

// GetByLoyverseID retrieves a product by Loyverse ID
func (r *productRepository) GetByLoyverseID(ctx context.Context, loyverseID string) (*entity.Product, error) {
	var product entity.Product
	err := r.db.WithContext(ctx).Where("loyverse_id = ?", loyverseID).First(&product).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

// Update updates a product
func (r *productRepository) Update(ctx context.Context, product *entity.Product) error {
	return r.db.WithContext(ctx).Save(product).Error
}

// Delete deletes a product
func (r *productRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Product{}, id).Error
}

// List lists products with filtering
func (r *productRepository) List(ctx context.Context, filter repository.ProductFilter) ([]*entity.Product, error) {
	var products []*entity.Product

	query := r.db.WithContext(ctx).Model(&entity.Product{})

	// Apply filters
	query = r.applyFilters(query, filter)

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	// Apply ordering
	if filter.OrderBy != "" {
		direction := "ASC"
		if filter.OrderDir == "DESC" {
			direction = "DESC"
		}
		query = query.Order(fmt.Sprintf("%s %s", filter.OrderBy, direction))
	} else {
		query = query.Order("created_at DESC")
	}

	err := query.Find(&products).Error
	return products, err
}

// Count counts products with filtering
func (r *productRepository) Count(ctx context.Context, filter repository.ProductFilter) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&entity.Product{})
	query = r.applyFilters(query, filter)

	err := query.Count(&count).Error
	return count, err
}

// Search searches products
func (r *productRepository) Search(ctx context.Context, searchQuery string, filter repository.ProductFilter) ([]*entity.Product, error) {
	var products []*entity.Product

	query := r.db.WithContext(ctx).Model(&entity.Product{})

	// Apply search
	if searchQuery != "" {
		search := "%" + searchQuery + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ? OR sku ILIKE ?", search, search, search)
	}

	// Apply other filters
	query = r.applyFilters(query, filter)

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	// Apply ordering
	if filter.OrderBy != "" {
		direction := "ASC"
		if filter.OrderDir == "DESC" {
			direction = "DESC"
		}
		query = query.Order(fmt.Sprintf("%s %s", filter.OrderBy, direction))
	} else {
		query = query.Order("created_at DESC")
	}

	err := query.Find(&products).Error
	return products, err
}

// CreateBatch creates multiple products
func (r *productRepository) CreateBatch(ctx context.Context, products []*entity.Product) error {
	return r.db.WithContext(ctx).CreateInBatches(products, 100).Error
}

// UpdateBatch updates multiple products
func (r *productRepository) UpdateBatch(ctx context.Context, products []*entity.Product) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, product := range products {
			if err := tx.Save(product).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// UpsertFromLoyverse upserts a product from Loyverse
func (r *productRepository) UpsertFromLoyverse(ctx context.Context, product *entity.Product) error {
	if product.LoyverseID == nil {
		return fmt.Errorf("loyverse ID is required for upsert operation")
	}

	var existing entity.Product
	err := r.db.WithContext(ctx).Where("loyverse_id = ?", *product.LoyverseID).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new product
		product.DataSourceType = "loyverse"
		return r.db.WithContext(ctx).Create(product).Error
	} else if err != nil {
		return err
	}

	// Update existing product if not manually overridden
	if !existing.IsManualOverride {
		product.ID = existing.ID
		product.CreatedAt = existing.CreatedAt
		product.CreatedBy = existing.CreatedBy
		product.Version = existing.Version + 1
		product.DataSourceType = "loyverse"
		return r.db.WithContext(ctx).Save(product).Error
	}

	return nil // Skip update due to manual override
}

// GetProductsForSync gets products that need syncing
func (r *productRepository) GetProductsForSync(ctx context.Context, dataSource string, limit int) ([]*entity.Product, error) {
	var products []*entity.Product

	query := r.db.WithContext(ctx).Model(&entity.Product{}).
		Where("data_source_type = ?", dataSource).
		Order("last_synced_at ASC NULLS FIRST")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&products).Error
	return products, err
}

// GetByCategory gets products by category
func (r *productRepository) GetByCategory(ctx context.Context, categoryID uuid.UUID, filter repository.ProductFilter) ([]*entity.Product, error) {
	filter.CategoryID = &categoryID
	return r.List(ctx, filter)
}

// GetWithPrices gets a product with its prices
func (r *productRepository) GetWithPrices(ctx context.Context, id uuid.UUID) (*entity.Product, error) {
	var product entity.Product
	err := r.db.WithContext(ctx).Preload("Prices").Where("id = ?", id).First(&product).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

// GetWithInventory gets a product with its inventory
func (r *productRepository) GetWithInventory(ctx context.Context, id uuid.UUID) (*entity.Product, error) {
	var product entity.Product
	err := r.db.WithContext(ctx).Preload("Inventory").Where("id = ?", id).First(&product).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

// GetActiveCount returns the count of active products
func (r *productRepository) GetActiveCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Product{}).
		Where("is_active = ?", true).
		Count(&count).Error
	return count, err
}

// GetCategoryStats returns product count by category
func (r *productRepository) GetCategoryStats(ctx context.Context) (map[uuid.UUID]int64, error) {
	type categoryStats struct {
		CategoryID uuid.UUID `json:"category_id"`
		Count      int64     `json:"count"`
	}

	var stats []categoryStats
	err := r.db.WithContext(ctx).Model(&entity.Product{}).
		Select("category_id, COUNT(*) as count").
		Where("category_id IS NOT NULL AND is_active = ?", true).
		Group("category_id").
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]int64)
	for _, stat := range stats {
		result[stat.CategoryID] = stat.Count
	}

	return result, nil
}

// GetByIDs returns products by their IDs
func (r *productRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.Product, error) {
	var products []*entity.Product
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&products).Error
	return products, err
}

// GetByDataSource returns product by data source type and ID
func (r *productRepository) GetByDataSource(ctx context.Context, dataSourceType, dataSourceID string) (*entity.Product, error) {
	var product entity.Product
	err := r.db.WithContext(ctx).Where("data_source_type = ? AND data_source_id = ?", dataSourceType, dataSourceID).First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

// GetManualOverrides returns products with manual overrides
func (r *productRepository) GetManualOverrides(ctx context.Context) ([]*entity.Product, error) {
	var products []*entity.Product
	err := r.db.WithContext(ctx).Where("is_manual_override = ?", true).Find(&products).Error
	return products, err
}

// SetManualOverride sets manual override flag for a product
func (r *productRepository) SetManualOverride(ctx context.Context, productID uuid.UUID, override bool) error {
	return r.db.WithContext(ctx).Model(&entity.Product{}).
		Where("id = ?", productID).
		Update("is_manual_override", override).Error
}

// GetProductsToSync returns products that need to be synced
func (r *productRepository) GetProductsToSync(ctx context.Context, lastSyncTime time.Time) ([]*entity.Product, error) {
	var products []*entity.Product
	err := r.db.WithContext(ctx).
		Where("updated_at > ? OR last_synced_at IS NULL OR last_synced_at < ?", lastSyncTime, lastSyncTime).
		Find(&products).Error
	return products, err
}

// UpdateSyncStatus updates the sync status of a product
func (r *productRepository) UpdateSyncStatus(ctx context.Context, productID uuid.UUID, syncTime time.Time) error {
	return r.db.WithContext(ctx).Model(&entity.Product{}).
		Where("id = ?", productID).
		Update("last_synced_at", syncTime).Error
}

// GetCount returns the count of products matching the filter
func (r *productRepository) GetCount(ctx context.Context, filter repository.ProductFilter) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&entity.Product{})
	query = r.applyFilters(query, filter)
	err := query.Count(&count).Error
	return count, err
}

// applyFilters applies filters to the query
func (r *productRepository) applyFilters(query *gorm.DB, filter repository.ProductFilter) *gorm.DB {
	if filter.CategoryID != nil {
		query = query.Where("category_id = ?", *filter.CategoryID)
	}

	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	if filter.IsVIPOnly != nil {
		query = query.Where("is_vip_only = ?", *filter.IsVIPOnly)
	}

	if filter.SKU != nil {
		query = query.Where("sku = ?", *filter.SKU)
	}

	if filter.Name != nil {
		query = query.Where("name ILIKE ?", "%"+*filter.Name+"%")
	}

	if len(filter.Tags) > 0 {
		query = query.Where("tags @> ?", filter.Tags)
	}

	if filter.MinPrice != nil {
		query = query.Where("base_price >= ?", *filter.MinPrice)
	}

	if filter.MaxPrice != nil {
		query = query.Where("base_price <= ?", *filter.MaxPrice)
	}

	if filter.DataSource != nil {
		query = query.Where("data_source_type = ?", *filter.DataSource)
	}

	if filter.Search != nil {
		search := "%" + *filter.Search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ? OR sku ILIKE ?", search, search, search)
	}

	return query
}
