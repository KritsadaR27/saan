package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"payment/internal/domain/entity"
	"payment/internal/domain/repository"
)

// PostgresLoyverseStoreRepository implements LoyverseStoreRepository using PostgreSQL
type PostgresLoyverseStoreRepository struct {
	db *sqlx.DB
}

// NewPostgresLoyverseStoreRepository creates a new PostgreSQL Loyverse store repository
func NewPostgresLoyverseStoreRepository(db *sqlx.DB) repository.LoyverseStoreRepository {
	return &PostgresLoyverseStoreRepository{db: db}
}

// Create creates a new Loyverse store
func (r *PostgresLoyverseStoreRepository) Create(ctx context.Context, store *entity.LoyverseStore) error {
	query := `
		INSERT INTO loyverse_stores (
			id, store_id, store_name, store_type, is_active, is_default,
			accepts_cash, accepts_transfer, accepts_cod, 
			delivery_driver_phone, delivery_route, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	
	_, err := r.db.ExecContext(ctx, query,
		store.ID,
		store.StoreID,
		store.StoreName,
		store.StoreType,
		store.IsActive,
		store.IsDefault,
		store.AcceptsCash,
		store.AcceptsTransfer,
		store.AcceptsCOD,
		store.DeliveryDriverPhone,
		store.DeliveryRoute,
		store.CreatedAt,
		store.UpdatedAt,
	)
	
	return err
}

// GetByID retrieves a Loyverse store by ID
func (r *PostgresLoyverseStoreRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.LoyverseStore, error) {
	query := `
		SELECT id, store_id, store_name, store_type, is_active, is_default,
			   accepts_cash, accepts_transfer, accepts_cod, 
			   delivery_driver_phone, delivery_route, created_at, updated_at
		FROM loyverse_stores 
		WHERE id = $1
	`
	
	var store entity.LoyverseStore
	err := r.db.GetContext(ctx, &store, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrStoreNotFound
		}
		return nil, err
	}
	
	return &store, nil
}

// GetByStoreCode retrieves a Loyverse store by store code
func (r *PostgresLoyverseStoreRepository) GetByStoreCode(ctx context.Context, storeCode string) (*entity.LoyverseStore, error) {
	query := `
		SELECT id, store_id, store_name, store_type, is_active, is_default,
			   accepts_cash, accepts_transfer, accepts_cod, 
			   delivery_driver_phone, delivery_route, created_at, updated_at
		FROM loyverse_stores 
		WHERE store_id = $1
	`
	
	var store entity.LoyverseStore
	err := r.db.GetContext(ctx, &store, query, storeCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrStoreNotFound
		}
		return nil, err
	}
	
	return &store, nil
}

// Update updates a Loyverse store
func (r *PostgresLoyverseStoreRepository) Update(ctx context.Context, store *entity.LoyverseStore) error {
	query := `
		UPDATE loyverse_stores 
		SET store_name = $2, store_type = $3, is_active = $4, is_default = $5,
			accepts_cash = $6, accepts_transfer = $7, accepts_cod = $8, 
			delivery_driver_phone = $9, delivery_route = $10, updated_at = $11
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(ctx, query,
		store.ID,
		store.StoreName,
		store.StoreType,
		store.IsActive,
		store.IsDefault,
		store.AcceptsCash,
		store.AcceptsTransfer,
		store.AcceptsCOD,
		store.DeliveryDriverPhone,
		store.DeliveryRoute,
		store.UpdatedAt,
	)
	
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return entity.ErrStoreNotFound
	}
	
	return nil
}

// Delete deletes a Loyverse store
func (r *PostgresLoyverseStoreRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM loyverse_stores WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return entity.ErrStoreNotFound
	}
	
	return nil
}

// GetAllStores retrieves all Loyverse stores
func (r *PostgresLoyverseStoreRepository) GetAllStores(ctx context.Context) ([]*entity.LoyverseStore, error) {
	query := `
		SELECT id, store_id, store_name, store_type, is_active, is_default,
			   accepts_cash, accepts_transfer, accepts_cod, 
			   delivery_driver_phone, delivery_route, created_at, updated_at
		FROM loyverse_stores 
		ORDER BY store_name
	`
	
	var stores []*entity.LoyverseStore
	err := r.db.SelectContext(ctx, &stores, query)
	if err != nil {
		return nil, err
	}
	
	return stores, nil
}

// GetActiveStores retrieves all active Loyverse stores
func (r *PostgresLoyverseStoreRepository) GetActiveStores(ctx context.Context) ([]*entity.LoyverseStore, error) {
	query := `
		SELECT id, store_id, store_name, store_type, is_active, is_default,
			   accepts_cash, accepts_transfer, accepts_cod, 
			   delivery_driver_phone, delivery_route, created_at, updated_at
		FROM loyverse_stores 
		WHERE is_active = true
		ORDER BY store_name
	`
	
	var stores []*entity.LoyverseStore
	err := r.db.SelectContext(ctx, &stores, query)
	if err != nil {
		return nil, err
	}
	
	return stores, nil
}

// GetStoresByRegion retrieves stores by region (using delivery_route as region indicator)
func (r *PostgresLoyverseStoreRepository) GetStoresByRegion(ctx context.Context, region string) ([]*entity.LoyverseStore, error) {
	query := `
		SELECT id, store_id, store_name, store_type, is_active, is_default,
			   accepts_cash, accepts_transfer, accepts_cod, 
			   delivery_driver_phone, delivery_route, created_at, updated_at
		FROM loyverse_stores 
		WHERE delivery_route = $1 AND is_active = true
		ORDER BY store_name
	`
	
	var stores []*entity.LoyverseStore
	err := r.db.SelectContext(ctx, &stores, query, region)
	if err != nil {
		return nil, err
	}
	
	return stores, nil
}

// GetStoresByManager retrieves stores by manager ID (placeholder - no manager field in current entity)
func (r *PostgresLoyverseStoreRepository) GetStoresByManager(ctx context.Context, managerID uuid.UUID) ([]*entity.LoyverseStore, error) {
	// Since we don't have manager info in the current entity, return empty for now
	// This would need to be implemented with proper manager association
	return []*entity.LoyverseStore{}, nil
}

// GetAvailableStoresForAssignment retrieves stores available for assignment
func (r *PostgresLoyverseStoreRepository) GetAvailableStoresForAssignment(ctx context.Context) ([]*entity.LoyverseStore, error) {
	query := `
		SELECT id, store_id, store_name, store_type, is_active, is_default,
			   accepts_cash, accepts_transfer, accepts_cod, 
			   delivery_driver_phone, delivery_route, created_at, updated_at
		FROM loyverse_stores 
		WHERE is_active = true AND store_type IN ('main', 'delivery')
		ORDER BY store_name
	`
	
	var stores []*entity.LoyverseStore
	err := r.db.SelectContext(ctx, &stores, query)
	if err != nil {
		return nil, err
	}
	
	return stores, nil
}

// GetStoreWorkload returns workload information for a store (placeholder implementation)
func (r *PostgresLoyverseStoreRepository) GetStoreWorkload(ctx context.Context, storeCode string, dateFrom, dateTo time.Time) (*repository.StoreWorkload, error) {
	// This would typically calculate workload based on orders, payments, etc.
	// For now, returning a basic implementation
	return &repository.StoreWorkload{
		StoreCode:           storeCode,
		PendingOrders:       0,
		ProcessingOrders:    0,
		TotalOrdersToday:    0,
		AvgProcessingTime:   0.0,
		CurrentCapacity:     0.0,
		LastUpdated:         time.Now(),
	}, nil
}

// UpdateStoreMetrics updates store metrics (placeholder implementation)
func (r *PostgresLoyverseStoreRepository) UpdateStoreMetrics(ctx context.Context, storeCode string, metrics *repository.StoreMetrics) error {
	// This would update store performance metrics
	// For now, just returning nil
	return nil
}
