package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/saan-system/services/customer/internal/domain"
)

type deliveryRouteRepository struct {
	db *sql.DB
}

// NewDeliveryRouteRepository creates a new delivery route repository
func NewDeliveryRouteRepository(db *sql.DB) domain.DeliveryRouteRepository {
	return &deliveryRouteRepository{db: db}
}

// Create creates a new delivery route
func (r *deliveryRouteRepository) Create(ctx context.Context, route *domain.DeliveryRoute) error {
	query := `
		INSERT INTO delivery_routes (id, name, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(ctx, query,
		route.ID, route.Name, route.Description, route.IsActive,
		route.CreatedAt, route.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create delivery route: %w", err)
	}

	return nil
}

// GetByID retrieves a delivery route by ID
func (r *deliveryRouteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.DeliveryRoute, error) {
	query := `
		SELECT id, name, description, is_active, created_at, updated_at
		FROM delivery_routes 
		WHERE id = $1`

	route := &domain.DeliveryRoute{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&route.ID, &route.Name, &route.Description, &route.IsActive,
		&route.CreatedAt, &route.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrDeliveryRouteNotFound
		}
		return nil, fmt.Errorf("failed to get delivery route: %w", err)
	}

	return route, nil
}

// GetAll retrieves all delivery routes
func (r *deliveryRouteRepository) GetAll(ctx context.Context) ([]domain.DeliveryRoute, error) {
	query := `
		SELECT id, name, description, is_active, created_at, updated_at
		FROM delivery_routes 
		ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery routes: %w", err)
	}
	defer rows.Close()

	var routes []domain.DeliveryRoute
	for rows.Next() {
		route := domain.DeliveryRoute{}
		err := rows.Scan(
			&route.ID, &route.Name, &route.Description, &route.IsActive,
			&route.CreatedAt, &route.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan delivery route: %w", err)
		}
		routes = append(routes, route)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return routes, nil
}

// Update updates a delivery route
func (r *deliveryRouteRepository) Update(ctx context.Context, route *domain.DeliveryRoute) error {
	query := `
		UPDATE delivery_routes 
		SET name = $2, description = $3, is_active = $4, updated_at = $5
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		route.ID, route.Name, route.Description, route.IsActive, route.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update delivery route: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrDeliveryRouteNotFound
	}

	return nil
}

// Delete deletes a delivery route
func (r *deliveryRouteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM delivery_routes WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete delivery route: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrDeliveryRouteNotFound
	}

	return nil
}
