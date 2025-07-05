package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"payment/internal/domain/entity"
	"payment/internal/domain/repository"
)

// PostgresPaymentDeliveryContextRepository implements PaymentDeliveryContextRepository using PostgreSQL
type PostgresPaymentDeliveryContextRepository struct {
	db *sqlx.DB
}

// NewPostgresPaymentDeliveryContextRepository creates a new PostgreSQL payment delivery context repository
func NewPostgresPaymentDeliveryContextRepository(db *sqlx.DB) repository.PaymentDeliveryContextRepository {
	return &PostgresPaymentDeliveryContextRepository{db: db}
}

// Create creates a new payment delivery context
func (r *PostgresPaymentDeliveryContextRepository) Create(ctx context.Context, context *entity.PaymentDeliveryContext) error {
	query := `
		INSERT INTO payment_delivery_contexts (
			payment_id, delivery_id, driver_id, delivery_address, delivery_status,
			estimated_arrival, actual_arrival, instructions, cod_amount, cod_collected_at,
			cod_collection_method, pickup_lat, pickup_lng, delivery_lat, delivery_lng,
			metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`
	
	_, err := r.db.ExecContext(ctx, query,
		context.PaymentID,
		context.DeliveryID,
		context.DriverID,
		context.DeliveryAddress,
		context.DeliveryStatus,
		context.EstimatedArrival,
		context.ActualArrival,
		context.Instructions,
		context.CODAmount,
		context.CODCollectedAt,
		context.CODCollectionMethod,
		context.PickupLat,
		context.PickupLng,
		context.DeliveryLat,
		context.DeliveryLng,
		context.Metadata,
		context.CreatedAt,
		context.UpdatedAt,
	)
	
	return err
}

// GetByPaymentID retrieves a payment delivery context by payment ID
func (r *PostgresPaymentDeliveryContextRepository) GetByPaymentID(ctx context.Context, paymentID uuid.UUID) (*entity.PaymentDeliveryContext, error) {
	query := `
		SELECT payment_id, delivery_id, driver_id, delivery_address, delivery_status,
			   estimated_arrival, actual_arrival, instructions, cod_amount, cod_collected_at,
			   cod_collection_method, pickup_lat, pickup_lng, delivery_lat, delivery_lng,
			   metadata, created_at, updated_at
		FROM payment_delivery_contexts 
		WHERE payment_id = $1
	`
	
	var context entity.PaymentDeliveryContext
	err := r.db.GetContext(ctx, &context, query, paymentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrPaymentNotFound
		}
		return nil, err
	}
	
	return &context, nil
}

// Update updates a payment delivery context
func (r *PostgresPaymentDeliveryContextRepository) Update(ctx context.Context, context *entity.PaymentDeliveryContext) error {
	query := `
		UPDATE payment_delivery_contexts 
		SET delivery_id = $2, driver_id = $3, delivery_address = $4, delivery_status = $5,
			estimated_arrival = $6, actual_arrival = $7, instructions = $8, cod_amount = $9,
			cod_collected_at = $10, cod_collection_method = $11, pickup_lat = $12, pickup_lng = $13,
			delivery_lat = $14, delivery_lng = $15, metadata = $16, updated_at = $17
		WHERE payment_id = $1
	`
	
	result, err := r.db.ExecContext(ctx, query,
		context.PaymentID,
		context.DeliveryID,
		context.DriverID,
		context.DeliveryAddress,
		context.DeliveryStatus,
		context.EstimatedArrival,
		context.ActualArrival,
		context.Instructions,
		context.CODAmount,
		context.CODCollectedAt,
		context.CODCollectionMethod,
		context.PickupLat,
		context.PickupLng,
		context.DeliveryLat,
		context.DeliveryLng,
		context.Metadata,
		context.UpdatedAt,
	)
	
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return entity.ErrPaymentNotFound
	}
	
	return nil
}

// Delete deletes a payment delivery context
func (r *PostgresPaymentDeliveryContextRepository) Delete(ctx context.Context, paymentID uuid.UUID) error {
	query := `DELETE FROM payment_delivery_contexts WHERE payment_id = $1`
	
	result, err := r.db.ExecContext(ctx, query, paymentID)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return entity.ErrPaymentNotFound
	}
	
	return nil
}

// GetByDeliveryID retrieves payment delivery contexts by delivery ID
func (r *PostgresPaymentDeliveryContextRepository) GetByDeliveryID(ctx context.Context, deliveryID uuid.UUID) ([]*entity.PaymentDeliveryContext, error) {
	query := `
		SELECT payment_id, delivery_id, driver_id, delivery_address, delivery_status,
			   estimated_arrival, actual_arrival, instructions, cod_amount, cod_collected_at,
			   cod_collection_method, pickup_lat, pickup_lng, delivery_lat, delivery_lng,
			   metadata, created_at, updated_at
		FROM payment_delivery_contexts 
		WHERE delivery_id = $1
		ORDER BY created_at DESC
	`
	
	var contexts []*entity.PaymentDeliveryContext
	err := r.db.SelectContext(ctx, &contexts, query, deliveryID)
	if err != nil {
		return nil, err
	}
	
	return contexts, nil
}

// GetByDriverID retrieves payment delivery contexts by driver ID
func (r *PostgresPaymentDeliveryContextRepository) GetByDriverID(ctx context.Context, driverID uuid.UUID) ([]*entity.PaymentDeliveryContext, error) {
	query := `
		SELECT payment_id, delivery_id, driver_id, delivery_address, delivery_status,
			   estimated_arrival, actual_arrival, instructions, cod_amount, cod_collected_at,
			   cod_collection_method, pickup_lat, pickup_lng, delivery_lat, delivery_lng,
			   metadata, created_at, updated_at
		FROM payment_delivery_contexts 
		WHERE driver_id = $1
		ORDER BY created_at DESC
	`
	
	var contexts []*entity.PaymentDeliveryContext
	err := r.db.SelectContext(ctx, &contexts, query, driverID)
	if err != nil {
		return nil, err
	}
	
	return contexts, nil
}

// GetContextsByDateRange retrieves payment delivery contexts by date range
func (r *PostgresPaymentDeliveryContextRepository) GetContextsByDateRange(ctx context.Context, dateFrom, dateTo time.Time) ([]*entity.PaymentDeliveryContext, error) {
	query := `
		SELECT payment_id, delivery_id, driver_id, delivery_address, delivery_status,
			   estimated_arrival, actual_arrival, instructions, cod_amount, cod_collected_at,
			   cod_collection_method, pickup_lat, pickup_lng, delivery_lat, delivery_lng,
			   metadata, created_at, updated_at
		FROM payment_delivery_contexts 
		WHERE created_at >= $1 AND created_at <= $2
		ORDER BY created_at DESC
	`
	
	var contexts []*entity.PaymentDeliveryContext
	err := r.db.SelectContext(ctx, &contexts, query, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}
	
	return contexts, nil
}

// GetCODContexts retrieves COD payment delivery contexts with filters
func (r *PostgresPaymentDeliveryContextRepository) GetCODContexts(ctx context.Context, filters repository.ContextFilters) ([]*entity.PaymentDeliveryContext, error) {
	// Build the base query
	query := `
		SELECT payment_id, delivery_id, driver_id, delivery_address, delivery_status,
			   estimated_arrival, actual_arrival, instructions, cod_amount, cod_collected_at,
			   cod_collection_method, pickup_lat, pickup_lng, delivery_lat, delivery_lng,
			   metadata, created_at, updated_at
		FROM payment_delivery_contexts 
		WHERE cod_amount > 0
	`
	
	args := []interface{}{}
	argIndex := 1
	
	// Apply filters
	if filters.DriverID != nil {
		query += fmt.Sprintf(" AND driver_id = $%d", argIndex)
		args = append(args, *filters.DriverID)
		argIndex++
	}
	
	if filters.DeliveryStatus != nil {
		query += fmt.Sprintf(" AND delivery_status = $%d", argIndex)
		args = append(args, *filters.DeliveryStatus)
		argIndex++
	}
	
	if filters.DateFrom != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *filters.DateFrom)
		argIndex++
	}
	
	if filters.DateTo != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *filters.DateTo)
		argIndex++
	}
	
	// Add ordering and pagination
	query += " ORDER BY created_at DESC"
	
	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
	}
	
	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filters.Offset)
		argIndex++
	}
	
	var contexts []*entity.PaymentDeliveryContext
	err := r.db.SelectContext(ctx, &contexts, query, args...)
	if err != nil {
		return nil, err
	}
	
	return contexts, nil
}
