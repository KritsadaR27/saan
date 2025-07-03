package repository

import (
	"context"

	"product-service/internal/domain/entity"

	"github.com/google/uuid"
)

// ProductFilter represents filters for product queries
type ProductFilter struct {
	CategoryID *uuid.UUID `json:"category_id,omitempty"`
	IsActive   *bool      `json:"is_active,omitempty"`
	IsVIPOnly  *bool      `json:"is_vip_only,omitempty"`
	SKU        *string    `json:"sku,omitempty"`
	Name       *string    `json:"name,omitempty"`
	Tags       []string   `json:"tags,omitempty"`
	MinPrice   *float64   `json:"min_price,omitempty"`
	MaxPrice   *float64   `json:"max_price,omitempty"`
	DataSource *string    `json:"data_source,omitempty"`
	Search     *string    `json:"search,omitempty"`
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
	OrderBy    string     `json:"order_by"`
	OrderDir   string     `json:"order_dir"`
}

// ProductRepository defines the interface for product data access
type ProductRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, product *entity.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Product, error)
	GetBySKU(ctx context.Context, sku string) (*entity.Product, error)
	GetByLoyverseID(ctx context.Context, loyverseID string) (*entity.Product, error)
	Update(ctx context.Context, product *entity.Product) error
	Delete(ctx context.Context, id uuid.UUID) error

	// List and search operations
	List(ctx context.Context, filter ProductFilter) ([]*entity.Product, error)
	Count(ctx context.Context, filter ProductFilter) (int64, error)
	Search(ctx context.Context, query string, filter ProductFilter) ([]*entity.Product, error)

	// Batch operations
	CreateBatch(ctx context.Context, products []*entity.Product) error
	UpdateBatch(ctx context.Context, products []*entity.Product) error

	// Sync operations
	UpsertFromLoyverse(ctx context.Context, product *entity.Product) error
	GetProductsForSync(ctx context.Context, dataSource string, limit int) ([]*entity.Product, error)

	// Relationship operations
	GetByCategory(ctx context.Context, categoryID uuid.UUID, filter ProductFilter) ([]*entity.Product, error)
	GetWithPrices(ctx context.Context, id uuid.UUID) (*entity.Product, error)
	GetWithInventory(ctx context.Context, id uuid.UUID) (*entity.Product, error)
}
