package repository

import (
	"context"

	"product-service/internal/domain/entity"

	"github.com/google/uuid"
)

// CategoryFilter represents filters for category queries
type CategoryFilter struct {
	ParentID   *uuid.UUID `json:"parent_id,omitempty"`
	IsActive   *bool      `json:"is_active,omitempty"`
	DataSource *string    `json:"data_source,omitempty"`
	Search     *string    `json:"search,omitempty"`
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
	OrderBy    string     `json:"order_by"`
	OrderDir   string     `json:"order_dir"`
}

// CategoryRepository defines the interface for category data access
type CategoryRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, category *entity.Category) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Category, error)
	GetByLoyverseID(ctx context.Context, loyverseID string) (*entity.Category, error)
	Update(ctx context.Context, category *entity.Category) error
	Delete(ctx context.Context, id uuid.UUID) error

	// List operations
	List(ctx context.Context, filter CategoryFilter) ([]*entity.Category, error)
	GetRoot(ctx context.Context) ([]*entity.Category, error)
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]*entity.Category, error)

	// Sync operations
	UpsertFromLoyverse(ctx context.Context, category *entity.Category) error
}
