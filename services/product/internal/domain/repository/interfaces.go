package repository

import (
	"context"
	"time"

	"product/internal/domain/entity"

	"github.com/google/uuid"
)

// ProductRepository defines product data access operations
type ProductRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, product *entity.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Product, error)
	GetBySKU(ctx context.Context, sku string) (*entity.Product, error)
	GetByLoyverseID(ctx context.Context, loyverseID string) (*entity.Product, error)
	Update(ctx context.Context, product *entity.Product) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Batch operations
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.Product, error)
	CreateBatch(ctx context.Context, products []*entity.Product) error
	UpdateBatch(ctx context.Context, products []*entity.Product) error

	// Search and filtering
	List(ctx context.Context, filter ProductFilter) ([]*entity.Product, error)
	Search(ctx context.Context, query string, filter ProductFilter) ([]*entity.Product, error)
	GetByCategory(ctx context.Context, categoryID uuid.UUID, filter ProductFilter) ([]*entity.Product, error)

	// Master Data Protection
	GetByDataSource(ctx context.Context, dataSourceType string, dataSourceID string) (*entity.Product, error)
	GetManualOverrides(ctx context.Context) ([]*entity.Product, error)
	SetManualOverride(ctx context.Context, productID uuid.UUID, override bool) error

	// Sync operations
	GetProductsToSync(ctx context.Context, lastSyncTime time.Time) ([]*entity.Product, error)
	UpdateSyncStatus(ctx context.Context, productID uuid.UUID, syncTime time.Time) error

	// Statistics
	GetCount(ctx context.Context, filter ProductFilter) (int64, error)
	GetActiveCount(ctx context.Context) (int64, error)
	GetCategoryStats(ctx context.Context) (map[uuid.UUID]int64, error)
}

// CategoryRepository defines category data access operations
type CategoryRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, category *entity.Category) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Category, error)
	GetByLoyverseID(ctx context.Context, loyverseID string) (*entity.Category, error)
	Update(ctx context.Context, category *entity.Category) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Hierarchy operations
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]*entity.Category, error)
	GetParent(ctx context.Context, childID uuid.UUID) (*entity.Category, error)
	GetTree(ctx context.Context) ([]*entity.Category, error)
	GetRoot(ctx context.Context) ([]*entity.Category, error)
	GetPath(ctx context.Context, categoryID uuid.UUID) ([]*entity.Category, error)

	// Batch operations
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.Category, error)
	CreateBatch(ctx context.Context, categories []*entity.Category) error
	UpdateBatch(ctx context.Context, categories []*entity.Category) error

	// Search and filtering
	List(ctx context.Context, filter CategoryFilter) ([]*entity.Category, error)
	Search(ctx context.Context, query string) ([]*entity.Category, error)

	// Master Data Protection
	GetByDataSource(ctx context.Context, dataSourceType string, dataSourceID string) (*entity.Category, error)
	GetManualOverrides(ctx context.Context) ([]*entity.Category, error)
	SetManualOverride(ctx context.Context, categoryID uuid.UUID, override bool) error

	// Statistics
	GetCount(ctx context.Context, filter CategoryFilter) (int64, error)
	GetProductCount(ctx context.Context, categoryID uuid.UUID) (int64, error)
}

// PriceRepository defines price data access operations
type PriceRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, price *entity.Price) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Price, error)
	Update(ctx context.Context, price *entity.Price) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Product pricing operations
	GetByProductID(ctx context.Context, productID uuid.UUID) ([]*entity.Price, error)
	GetActiveByProductID(ctx context.Context, productID uuid.UUID) ([]*entity.Price, error)
	GetByType(ctx context.Context, productID uuid.UUID, priceType string) ([]*entity.Price, error)

	// Batch operations
	CreateBatch(ctx context.Context, prices []*entity.Price) error
	UpdateBatch(ctx context.Context, prices []*entity.Price) error
	DeleteByProductID(ctx context.Context, productID uuid.UUID) error

	// Price calculation
	GetBestPrice(ctx context.Context, req PriceRequest) (*entity.Price, error)
	GetApplicablePrices(ctx context.Context, req PriceRequest) ([]*entity.Price, error)
	GetEffectivePrice(ctx context.Context, productID uuid.UUID, priceType string, customerInfo *CustomerInfo) (*entity.Price, error)

	// VIP pricing
	GetVIPPrices(ctx context.Context, productID uuid.UUID) ([]*entity.Price, error)
	SetVIPPrice(ctx context.Context, productID uuid.UUID, vipTierID uuid.UUID, price float64) error

	// Bulk pricing
	GetBulkPrices(ctx context.Context, productID uuid.UUID) ([]*entity.Price, error)
	SetBulkPrice(ctx context.Context, productID uuid.UUID, minQty, maxQty *int, price float64) error

	// Promotional pricing
	GetPromotionalPrices(ctx context.Context, productID uuid.UUID) ([]*entity.Price, error)
	GetActivePromotions(ctx context.Context, productID uuid.UUID) ([]*entity.Price, error)
	GetExpiredPromotions(ctx context.Context) ([]*entity.Price, error)
	SetPromotionalPrice(ctx context.Context, req PromotionalPriceRequest) error

	// Statistics
	GetPriceHistory(ctx context.Context, productID uuid.UUID, days int) ([]*entity.Price, error)
	GetAveragePrice(ctx context.Context, productID uuid.UUID, days int) (float64, error)
}

// CustomerInfo represents customer information for pricing
type CustomerInfo struct {
	CustomerID uuid.UUID
	LocationID uuid.UUID
	Quantity   int
	VIPTierID  *uuid.UUID
	GroupIDs   []uuid.UUID
}

// InventoryRepository defines inventory data access operations
type InventoryRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, inventory *entity.Inventory) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Inventory, error)
	GetByProductAndLocation(ctx context.Context, productID, locationID uuid.UUID) (*entity.Inventory, error)
	Update(ctx context.Context, inventory *entity.Inventory) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Product inventory operations
	GetByProductID(ctx context.Context, productID uuid.UUID) ([]*entity.Inventory, error)
	GetByLocationID(ctx context.Context, locationID uuid.UUID) ([]*entity.Inventory, error)

	// Batch operations
	GetByProductIDs(ctx context.Context, productIDs []uuid.UUID, locationID uuid.UUID) ([]*entity.Inventory, error)
	UpdateBatch(ctx context.Context, inventories []*entity.Inventory) error

	// Stock management
	UpdateStockLevel(ctx context.Context, productID, locationID uuid.UUID, delta float64) error
	ReserveStock(ctx context.Context, productID, locationID uuid.UUID, quantity float64) error
	ReleaseStock(ctx context.Context, productID, locationID uuid.UUID, quantity float64) error
	GetAvailableStock(ctx context.Context, productID, locationID uuid.UUID) (float64, error)

	// Availability
	GetAvailability(ctx context.Context, productID uuid.UUID) (*entity.ProductAvailability, error)
	GetAvailabilityByLocation(ctx context.Context, productID, locationID uuid.UUID) (*entity.LocationAvailability, error)
	SetAvailability(ctx context.Context, productID, locationID uuid.UUID, available bool, reason *string) error

	// Low stock alerts
	GetLowStockItems(ctx context.Context, locationID uuid.UUID) ([]*entity.Inventory, error)
	GetOutOfStockItems(ctx context.Context, locationID uuid.UUID) ([]*entity.Inventory, error)

	// Statistics
	GetTotalValue(ctx context.Context, locationID uuid.UUID) (float64, error)
	GetTurnoverRate(ctx context.Context, productID uuid.UUID, days int) (float64, error)
}

// CacheRepository defines caching operations
type CacheRepository interface {
	// Basic cache operations
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string) (interface{}, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// Batch operations
	SetBatch(ctx context.Context, items map[string]interface{}, ttl time.Duration) error
	GetBatch(ctx context.Context, keys []string) (map[string]interface{}, error)
	DeleteBatch(ctx context.Context, keys []string) error

	// Pattern operations
	DeletePattern(ctx context.Context, pattern string) error
	GetKeys(ctx context.Context, pattern string) ([]string, error)

	// Product-specific cache operations
	SetProduct(ctx context.Context, productID uuid.UUID, product *entity.Product, ttl time.Duration) error
	GetProduct(ctx context.Context, productID uuid.UUID) (*entity.Product, error)
	SetProductList(ctx context.Context, key string, products []*entity.Product, ttl time.Duration) error
	GetProductList(ctx context.Context, key string) ([]*entity.Product, error)

	// Price cache operations
	SetPrice(ctx context.Context, productID uuid.UUID, priceCalc *entity.PriceCalculation, ttl time.Duration) error
	GetPrice(ctx context.Context, productID uuid.UUID) (*entity.PriceCalculation, error)

	// Inventory cache operations
	SetAvailability(ctx context.Context, productID uuid.UUID, availability *entity.ProductAvailability, ttl time.Duration) error
	GetAvailability(ctx context.Context, productID uuid.UUID) (*entity.ProductAvailability, error)

	// Statistics cache
	SetStats(ctx context.Context, key string, stats interface{}, ttl time.Duration) error
	GetStats(ctx context.Context, key string) (interface{}, error)
}

// Filter types for repository operations
type ProductFilter struct {
	CategoryID *uuid.UUID
	IsActive   *bool
	IsVIPOnly  *bool
	MinPrice   *float64
	MaxPrice   *float64
	SKU        *string
	Name       *string
	Tags       []string
	LocationID *uuid.UUID
	DataSource *string
	HasStock   *bool
	Search     *string
	Limit      int
	Offset     int
	OrderBy    string
	OrderDir   string
}

type CategoryFilter struct {
	ParentID   *uuid.UUID
	IsActive   *bool
	DataSource *string
	Limit      int
	Offset     int
	OrderBy    string
	OrderDir   string
}

type PriceRequest struct {
	ProductID        uuid.UUID
	CustomerID       *uuid.UUID
	LocationID       *uuid.UUID
	Quantity         *int
	VIPTierID        *uuid.UUID
	CustomerGroupIDs []uuid.UUID
	RequestTime      time.Time
}

type PromotionalPriceRequest struct {
	ProductID        uuid.UUID
	Price            float64
	ValidFrom        time.Time
	ValidTo          time.Time
	PromotionName    string
	DiscountPercent  *float64
	LocationIDs      []uuid.UUID
	CustomerGroupIDs []uuid.UUID
	MinQuantity      *int
	MaxQuantity      *int
	Priority         int
}
