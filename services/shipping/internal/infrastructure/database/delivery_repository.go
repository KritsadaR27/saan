package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"shipping/internal/domain/entity"
	"shipping/internal/domain/repository"
)

// DeliveryRepository implements the delivery repository interface
type DeliveryRepository struct {
	db *sqlx.DB
}

// NewDeliveryRepository creates a new delivery repository
func NewDeliveryRepository(db *sqlx.DB) repository.DeliveryRepository {
	return &DeliveryRepository{db: db}
}

// Create saves a new delivery to the database
func (r *DeliveryRepository) Create(ctx context.Context, delivery *entity.DeliveryOrder) error {
	query := `
		INSERT INTO deliveries (
			id, order_id, customer_id, customer_address_id, delivery_method,
			priority_level, delivery_fee, cod_amount, weight, volume,
			provider_id, provider_order_id, tracking_number, vehicle_id,
			route_id, estimated_delivery_time, status, created_at, updated_at
		) VALUES (
			:id, :order_id, :customer_id, :customer_address_id, :delivery_method,
			:priority_level, :delivery_fee, :cod_amount, :weight, :volume,
			:provider_id, :provider_order_id, :tracking_number, :vehicle_id,
			:route_id, :estimated_delivery_time, :status, :created_at, :updated_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, delivery)
	if err != nil {
		return fmt.Errorf("failed to create delivery: %w", err)
	}

	return nil
}

// GetByID retrieves a delivery by ID
func (r *DeliveryRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.DeliveryOrder, error) {
	query := `
		SELECT id, order_id, customer_id, customer_address_id, delivery_method,
			   priority_level, delivery_fee, cod_amount, weight, volume,
			   provider_id, provider_order_id, tracking_number, vehicle_id,
			   route_id, estimated_delivery_time, status, notes, attempts,
			   completed_at, created_at, updated_at
		FROM deliveries 
		WHERE id = $1`

	var delivery entity.DeliveryOrder
	err := r.db.GetContext(ctx, &delivery, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("delivery not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get delivery: %w", err)
	}

	return &delivery, nil
}

// GetByOrderID retrieves a delivery by order ID
func (r *DeliveryRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*entity.DeliveryOrder, error) {
	query := `
		SELECT id, order_id, customer_id, customer_address_id, delivery_method,
			   priority_level, delivery_fee, cod_amount, weight, volume,
			   provider_id, provider_order_id, tracking_number, vehicle_id,
			   route_id, estimated_delivery_time, status, notes, attempts,
			   completed_at, created_at, updated_at
		FROM deliveries 
		WHERE order_id = $1
		LIMIT 1`

	var delivery entity.DeliveryOrder
	err := r.db.GetContext(ctx, &delivery, query, orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrDeliveryNotFound
		}
		return nil, fmt.Errorf("failed to get delivery by order ID: %w", err)
	}

	return &delivery, nil
}

// GetByTrackingNumber retrieves a delivery by tracking number
func (r *DeliveryRepository) GetByTrackingNumber(ctx context.Context, trackingNumber string) (*entity.DeliveryOrder, error) {
	query := `
		SELECT id, order_id, customer_id, customer_address_id, delivery_method,
			   priority_level, delivery_fee, cod_amount, weight, volume,
			   provider_id, provider_order_id, tracking_number, vehicle_id,
			   route_id, estimated_delivery_time, status, notes, attempts,
			   completed_at, created_at, updated_at
		FROM deliveries 
		WHERE tracking_number = $1`

	var delivery entity.DeliveryOrder
	err := r.db.GetContext(ctx, &delivery, query, trackingNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("delivery not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get delivery by tracking number: %w", err)
	}

	return &delivery, nil
}

// GetByStatus retrieves deliveries by status
func (r *DeliveryRepository) GetByStatus(ctx context.Context, status entity.DeliveryStatus, limit, offset int) ([]*entity.DeliveryOrder, error) {
	query := `
		SELECT id, order_id, customer_id, customer_address_id, delivery_method,
			   priority_level, delivery_fee, cod_amount, weight, volume,
			   provider_id, provider_order_id, tracking_number, vehicle_id,
			   route_id, estimated_delivery_time, status, notes, attempts,
			   completed_at, created_at, updated_at
		FROM deliveries 
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	var deliveries []*entity.DeliveryOrder
	err := r.db.SelectContext(ctx, &deliveries, query, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get deliveries by status: %w", err)
	}

	return deliveries, nil
}

// GetByVehicle retrieves deliveries assigned to a vehicle
func (r *DeliveryRepository) GetByVehicle(ctx context.Context, vehicleID uuid.UUID, limit, offset int) ([]*entity.DeliveryOrder, error) {
	query := `
		SELECT id, order_id, customer_id, customer_address_id, delivery_method,
			   priority_level, delivery_fee, cod_amount, weight, volume,
			   provider_id, provider_order_id, tracking_number, vehicle_id,
			   route_id, estimated_delivery_time, status, notes, attempts,
			   completed_at, created_at, updated_at
		FROM deliveries 
		WHERE vehicle_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	var deliveries []*entity.DeliveryOrder
	err := r.db.SelectContext(ctx, &deliveries, query, vehicleID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get deliveries by vehicle: %w", err)
	}

	return deliveries, nil
}

// GetByRoute retrieves deliveries assigned to a route
func (r *DeliveryRepository) GetByRoute(ctx context.Context, routeID uuid.UUID, limit, offset int) ([]*entity.DeliveryOrder, error) {
	query := `
		SELECT id, order_id, customer_id, customer_address_id, delivery_method,
			   priority_level, delivery_fee, cod_amount, weight, volume,
			   provider_id, provider_order_id, tracking_number, vehicle_id,
			   route_id, estimated_delivery_time, status, notes, attempts,
			   completed_at, created_at, updated_at
		FROM deliveries 
		WHERE route_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	var deliveries []*entity.DeliveryOrder
	err := r.db.SelectContext(ctx, &deliveries, query, routeID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get deliveries by route: %w", err)
	}

	return deliveries, nil
}

// Update updates an existing delivery
func (r *DeliveryRepository) Update(ctx context.Context, delivery *entity.DeliveryOrder) error {
	query := `
		UPDATE deliveries SET
			priority_level = :priority_level,
			delivery_fee = :delivery_fee,
			cod_amount = :cod_amount,
			weight = :weight,
			volume = :volume,
			provider_id = :provider_id,
			provider_order_id = :provider_order_id,
			tracking_number = :tracking_number,
			vehicle_id = :vehicle_id,
			route_id = :route_id,
			estimated_delivery_time = :estimated_delivery_time,
			status = :status,
			notes = :notes,
			attempts = :attempts,
			completed_at = :completed_at,
			updated_at = :updated_at
		WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, query, delivery)
	if err != nil {
		return fmt.Errorf("failed to update delivery: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("delivery not found")
	}

	return nil
}

// UpdateStatus updates the status of a delivery
func (r *DeliveryRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.DeliveryStatus) error {
	query := `
		UPDATE deliveries SET
			status = $1,
			updated_at = $2
		WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update delivery status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("delivery not found")
	}

	return nil
}

// UpdateTrackingInfo updates tracking information for a delivery
func (r *DeliveryRepository) UpdateTrackingInfo(ctx context.Context, id uuid.UUID, trackingNumber string, providerOrderID string) error {
	query := `
		UPDATE deliveries SET
			tracking_number = $1,
			provider_order_id = $2,
			updated_at = $3
		WHERE id = $4`

	result, err := r.db.ExecContext(ctx, query, trackingNumber, providerOrderID, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update tracking info: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("delivery not found")
	}

	return nil
}

// UpdateActualTimes updates actual pickup and delivery times
func (r *DeliveryRepository) UpdateActualTimes(ctx context.Context, id uuid.UUID, pickupTime, deliveryTime *time.Time) error {
	query := `
		UPDATE deliveries SET
			actual_pickup_time = $1,
			actual_delivery_time = $2,
			updated_at = $3
		WHERE id = $4`

	result, err := r.db.ExecContext(ctx, query, pickupTime, deliveryTime, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update actual times: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("delivery not found")
	}

	return nil
}

// AssignVehicle assigns a vehicle to a delivery
func (r *DeliveryRepository) AssignVehicle(ctx context.Context, deliveryID, vehicleID uuid.UUID) error {
	query := `
		UPDATE deliveries SET
			vehicle_id = $1,
			updated_at = $2
		WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, vehicleID, time.Now(), deliveryID)
	if err != nil {
		return fmt.Errorf("failed to assign vehicle: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("delivery not found")
	}

	return nil
}

// AssignRoute assigns a route to a delivery
func (r *DeliveryRepository) AssignRoute(ctx context.Context, deliveryID, routeID uuid.UUID) error {
	query := `
		UPDATE deliveries SET
			route_id = $1,
			updated_at = $2
		WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, routeID, time.Now(), deliveryID)
	if err != nil {
		return fmt.Errorf("failed to assign route: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("delivery not found")
	}

	return nil
}

// UpdateMultipleStatuses updates the status of multiple deliveries
func (r *DeliveryRepository) UpdateMultipleStatuses(ctx context.Context, ids []uuid.UUID, status entity.DeliveryStatus) error {
	if len(ids) == 0 {
		return nil
	}

	query := `
		UPDATE deliveries SET
			status = $1,
			updated_at = $2
		WHERE id = ANY($3)`

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), ids)
	if err != nil {
		return fmt.Errorf("failed to update multiple delivery statuses: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no deliveries found")
	}

	return nil
}

// Delete soft deletes a delivery
func (r *DeliveryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE deliveries SET
			status = $1,
			updated_at = $2
		WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, entity.DeliveryStatusCancelled, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete delivery: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("delivery not found")
	}

	return nil
}

// GetPendingDeliveries retrieves all pending deliveries
func (r *DeliveryRepository) GetPendingDeliveries(ctx context.Context, limit, offset int) ([]*entity.DeliveryOrder, error) {
	return r.GetByStatus(ctx, entity.DeliveryStatusPending, limit, offset)
}

// GetManualCoordinationDeliveries retrieves deliveries requiring manual coordination
func (r *DeliveryRepository) GetManualCoordinationDeliveries(ctx context.Context, limit, offset int) ([]*entity.DeliveryOrder, error) {
	query := `
		SELECT id, order_id, customer_id, customer_address_id, delivery_method,
			   priority_level, delivery_fee, cod_amount, weight, volume,
			   provider_id, provider_order_id, tracking_number, vehicle_id,
			   route_id, estimated_delivery_time, status, notes, attempts,
			   completed_at, created_at, updated_at
		FROM deliveries 
		WHERE delivery_method = $1 
		AND status IN ($2, $3)
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5`

	var deliveries []*entity.DeliveryOrder
	err := r.db.SelectContext(ctx, &deliveries, query, 
		entity.DeliveryMethodInterExpress, 
		entity.DeliveryStatusPending, 
		entity.DeliveryStatusPlanned,
		limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get manual coordination deliveries: %w", err)
	}

	return deliveries, nil
}

// GetByCustomerID retrieves deliveries by customer ID
func (r *DeliveryRepository) GetByCustomerID(ctx context.Context, customerID uuid.UUID, limit, offset int) ([]*entity.DeliveryOrder, error) {
	query := `
		SELECT id, order_id, customer_id, customer_address_id, delivery_method,
			   priority_level, delivery_fee, cod_amount, weight, volume,
			   provider_id, provider_order_id, tracking_number, vehicle_id,
			   route_id, estimated_delivery_time, status, notes, attempts,
			   completed_at, created_at, updated_at
		FROM deliveries 
		WHERE customer_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	var deliveries []*entity.DeliveryOrder
	err := r.db.SelectContext(ctx, &deliveries, query, customerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get deliveries by customer ID: %w", err)
	}

	return deliveries, nil
}

// GetByDateRange retrieves deliveries within a date range
func (r *DeliveryRepository) GetByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*entity.DeliveryOrder, error) {
	query := `
		SELECT id, order_id, customer_id, customer_address_id, delivery_method,
			   priority_level, delivery_fee, cod_amount, weight, volume,
			   provider_id, provider_order_id, tracking_number, vehicle_id,
			   route_id, estimated_delivery_time, status, notes, attempts,
			   completed_at, created_at, updated_at
		FROM deliveries 
		WHERE created_at >= $1 AND created_at <= $2
		ORDER BY created_at DESC`

	var deliveries []*entity.DeliveryOrder
	err := r.db.SelectContext(ctx, &deliveries, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get deliveries by date range: %w", err)
	}

	return deliveries, nil
}

// GetByVehicleID retrieves deliveries by vehicle ID for a specific date
func (r *DeliveryRepository) GetByVehicleID(ctx context.Context, vehicleID uuid.UUID, date time.Time) ([]*entity.DeliveryOrder, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
		SELECT id, order_id, customer_id, customer_address_id, delivery_method,
			   priority_level, delivery_fee, cod_amount, weight, volume,
			   provider_id, provider_order_id, tracking_number, vehicle_id,
			   route_id, estimated_delivery_time, status, notes, attempts,
			   completed_at, created_at, updated_at
		FROM deliveries 
		WHERE vehicle_id = $1 
		AND created_at >= $2 
		AND created_at < $3
		ORDER BY created_at`

	var deliveries []*entity.DeliveryOrder
	err := r.db.SelectContext(ctx, &deliveries, query, vehicleID, startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("failed to get deliveries by vehicle ID: %w", err)
	}

	return deliveries, nil
}

// GetByRouteID retrieves deliveries by route ID  
func (r *DeliveryRepository) GetByRouteID(ctx context.Context, routeID uuid.UUID) ([]*entity.DeliveryOrder, error) {
	query := `
		SELECT id, order_id, customer_id, customer_address_id, delivery_method,
			   priority_level, delivery_fee, cod_amount, weight, volume,
			   provider_id, provider_order_id, tracking_number, vehicle_id,
			   route_id, estimated_delivery_time, status, notes, attempts,
			   completed_at, created_at, updated_at
		FROM deliveries 
		WHERE route_id = $1
		ORDER BY created_at`

	var deliveries []*entity.DeliveryOrder
	err := r.db.SelectContext(ctx, &deliveries, query, routeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deliveries by route ID: %w", err)
	}

	return deliveries, nil
}

// GetByDeliveryMethod retrieves deliveries by delivery method
func (r *DeliveryRepository) GetByDeliveryMethod(ctx context.Context, method entity.DeliveryMethod, limit, offset int) ([]*entity.DeliveryOrder, error) {
	query := `
		SELECT id, order_id, customer_id, customer_address_id, delivery_method,
			   priority_level, delivery_fee, cod_amount, weight, volume,
			   provider_id, provider_order_id, tracking_number, vehicle_id,
			   route_id, estimated_delivery_time, status, notes, attempts,
			   completed_at, created_at, updated_at
		FROM deliveries 
		WHERE delivery_method = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	var deliveries []*entity.DeliveryOrder
	err := r.db.SelectContext(ctx, &deliveries, query, method, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get deliveries by delivery method: %w", err)
	}

	return deliveries, nil
}

// GetPendingInterExpressOrders retrieves pending Inter Express orders for a date
func (r *DeliveryRepository) GetPendingInterExpressOrders(ctx context.Context, date time.Time) ([]*entity.DeliveryOrder, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
		SELECT id, order_id, customer_id, customer_address_id, delivery_method,
			   priority_level, delivery_fee, cod_amount, weight, volume,
			   provider_id, provider_order_id, tracking_number, vehicle_id,
			   route_id, estimated_delivery_time, status, notes, attempts,
			   completed_at, created_at, updated_at
		FROM deliveries 
		WHERE delivery_method = $1 
		AND status = $2
		AND created_at >= $3 
		AND created_at < $4
		ORDER BY created_at`

	var deliveries []*entity.DeliveryOrder
	err := r.db.SelectContext(ctx, &deliveries, query, 
		entity.DeliveryMethodInterExpress, 
		entity.DeliveryStatusPending,
		startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending Inter Express orders: %w", err)
	}

	return deliveries, nil
}

// GetInterExpressByDate retrieves all Inter Express orders for a specific date
func (r *DeliveryRepository) GetInterExpressByDate(ctx context.Context, date time.Time) ([]*entity.DeliveryOrder, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
		SELECT id, order_id, customer_id, customer_address_id, delivery_method,
			   priority_level, delivery_fee, cod_amount, weight, volume,
			   provider_id, provider_order_id, tracking_number, vehicle_id,
			   route_id, estimated_delivery_time, status, notes, attempts,
			   completed_at, created_at, updated_at
		FROM deliveries 
		WHERE delivery_method = $1 
		AND created_at >= $2 
		AND created_at < $3
		ORDER BY created_at`

	var deliveries []*entity.DeliveryOrder
	err := r.db.SelectContext(ctx, &deliveries, query, 
		entity.DeliveryMethodInterExpress, 
		startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("failed to get Inter Express orders by date: %w", err)
	}

	return deliveries, nil
}

// GetActiveDeliveries retrieves all active deliveries
func (r *DeliveryRepository) GetActiveDeliveries(ctx context.Context, limit, offset int) ([]*entity.DeliveryOrder, error) {
	query := `
		SELECT id, order_id, customer_id, customer_address_id, delivery_method,
			   priority_level, delivery_fee, cod_amount, weight, volume,
			   provider_id, provider_order_id, tracking_number, vehicle_id,
			   route_id, estimated_delivery_time, status, notes, attempts,
			   completed_at, created_at, updated_at
		FROM deliveries 
		WHERE status IN ($1, $2, $3, $4)
		ORDER BY created_at DESC
		LIMIT $5 OFFSET $6`

	var deliveries []*entity.DeliveryOrder
	err := r.db.SelectContext(ctx, &deliveries, query, 
		entity.DeliveryStatusPending,
		entity.DeliveryStatusPlanned,
		entity.DeliveryStatusDispatched,
		entity.DeliveryStatusInTransit,
		limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get active deliveries: %w", err)
	}

	return deliveries, nil
}

// GetCompletedDeliveries retrieves completed deliveries
func (r *DeliveryRepository) GetCompletedDeliveries(ctx context.Context, startDate, endDate time.Time) ([]*entity.DeliveryOrder, error) {
	query := `
		SELECT id, order_id, customer_id, customer_address_id, delivery_method,
			   priority_level, delivery_fee, cod_amount, weight, volume,
			   provider_id, provider_order_id, tracking_number, vehicle_id,
			   route_id, estimated_delivery_time, status, notes, attempts,
			   completed_at, created_at, updated_at
		FROM deliveries 
		WHERE status = $1
		AND completed_at >= $2 
		AND completed_at <= $3
		ORDER BY completed_at DESC`

	var deliveries []*entity.DeliveryOrder
	err := r.db.SelectContext(ctx, &deliveries, query, entity.DeliveryStatusDelivered, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get completed deliveries: %w", err)
	}

	return deliveries, nil
}

// GetFailedDeliveries retrieves failed deliveries
func (r *DeliveryRepository) GetFailedDeliveries(ctx context.Context, startDate, endDate time.Time) ([]*entity.DeliveryOrder, error) {
	query := `
		SELECT id, order_id, customer_id, customer_address_id, delivery_method,
			   priority_level, delivery_fee, cod_amount, weight, volume,
			   provider_id, provider_order_id, tracking_number, vehicle_id,
			   route_id, estimated_delivery_time, status, notes, attempts,
			   completed_at, created_at, updated_at
		FROM deliveries 
		WHERE status = $1
		AND created_at >= $2 
		AND created_at <= $3
		ORDER BY created_at DESC`

	var deliveries []*entity.DeliveryOrder
	err := r.db.SelectContext(ctx, &deliveries, query, entity.DeliveryStatusFailed, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed deliveries: %w", err)
	}

	return deliveries, nil
}

// SearchDeliveries searches deliveries with filters
func (r *DeliveryRepository) SearchDeliveries(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*entity.DeliveryOrder, error) {
	query := `
		SELECT id, order_id, customer_id, customer_address_id, delivery_method,
			   priority_level, delivery_fee, cod_amount, weight, volume,
			   provider_id, provider_order_id, tracking_number, vehicle_id,
			   route_id, estimated_delivery_time, status, notes, attempts,
			   completed_at, created_at, updated_at
		FROM deliveries WHERE 1=1`

	args := []interface{}{}
	argCount := 0

	if customerID, ok := filters["customer_id"]; ok && customerID != nil {
		argCount++
		query += fmt.Sprintf(" AND customer_id = $%d", argCount)
		args = append(args, customerID)
	}

	if status, ok := filters["status"]; ok && status != nil {
		argCount++
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
	}

	if deliveryMethod, ok := filters["delivery_method"]; ok && deliveryMethod != nil {
		argCount++
		query += fmt.Sprintf(" AND delivery_method = $%d", argCount)
		args = append(args, deliveryMethod)
	}

	if vehicleID, ok := filters["vehicle_id"]; ok && vehicleID != nil {
		argCount++
		query += fmt.Sprintf(" AND vehicle_id = $%d", argCount)
		args = append(args, vehicleID)
	}

	if startDate, ok := filters["start_date"]; ok && startDate != nil {
		argCount++
		query += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, startDate)
	}

	if endDate, ok := filters["end_date"]; ok && endDate != nil {
		argCount++
		query += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, endDate)
	}

	query += " ORDER BY created_at DESC"

	argCount++
	query += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append(args, limit)

	argCount++
	query += fmt.Sprintf(" OFFSET $%d", argCount)
	args = append(args, offset)

	var deliveries []*entity.DeliveryOrder
	err := r.db.SelectContext(ctx, &deliveries, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search deliveries: %w", err)
	}

	return deliveries, nil
}

// GetDeliveryMetrics calculates delivery metrics for a date range
func (r *DeliveryRepository) GetDeliveryMetrics(ctx context.Context, startDate, endDate time.Time) (*repository.DeliveryMetrics, error) {
	query := `
		SELECT 
			COUNT(*) as total_deliveries,
			COUNT(CASE WHEN status = 'delivered' THEN 1 END) as completed_count,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_count,
			COALESCE(AVG(CASE 
				WHEN status = 'delivered' AND completed_at IS NOT NULL 
				THEN EXTRACT(EPOCH FROM (completed_at - created_at))/3600 
			END), 0) as avg_delivery_time_hours,
			COALESCE(SUM(delivery_fee), 0) as total_revenue
		FROM deliveries 
		WHERE created_at >= $1 AND created_at <= $2`

	var result struct {
		TotalDeliveries      int64   `db:"total_deliveries"`
		CompletedCount       int64   `db:"completed_count"`
		FailedCount          int64   `db:"failed_count"`
		AvgDeliveryTimeHours float64 `db:"avg_delivery_time_hours"`
		TotalRevenue         float64 `db:"total_revenue"`
	}

	err := r.db.GetContext(ctx, &result, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery metrics: %w", err)
	}

	metrics := &repository.DeliveryMetrics{
		TotalDeliveries:     result.TotalDeliveries,
		AverageDeliveryTime: result.AvgDeliveryTimeHours,
		TotalRevenue:        result.TotalRevenue,
	}

	if result.TotalDeliveries > 0 {
		metrics.SuccessRate = float64(result.CompletedCount) / float64(result.TotalDeliveries) * 100
	}

	return metrics, nil
}

// Analytics and reporting methods

// GetDeliveryCountByStatus returns count of deliveries grouped by status
func (r *DeliveryRepository) GetDeliveryCountByStatus(ctx context.Context, startDate, endDate time.Time) (map[entity.DeliveryStatus]int64, error) {
	query := `
		SELECT status, COUNT(*) as count
		FROM deliveries 
		WHERE created_at >= $1 AND created_at <= $2
		GROUP BY status`

	rows, err := r.db.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery count by status: %w", err)
	}
	defer rows.Close()

	result := make(map[entity.DeliveryStatus]int64)
	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan delivery count by status: %w", err)
		}
		result[entity.DeliveryStatus(status)] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate delivery count by status rows: %w", err)
	}

	return result, nil
}

// GetDeliveryCountByMethod returns count of deliveries grouped by method
func (r *DeliveryRepository) GetDeliveryCountByMethod(ctx context.Context, startDate, endDate time.Time) (map[entity.DeliveryMethod]int64, error) {
	query := `
		SELECT delivery_method, COUNT(*) as count
		FROM deliveries 
		WHERE created_at >= $1 AND created_at <= $2
		GROUP BY delivery_method`

	rows, err := r.db.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery count by method: %w", err)
	}
	defer rows.Close()

	result := make(map[entity.DeliveryMethod]int64)
	for rows.Next() {
		var method string
		var count int64
		if err := rows.Scan(&method, &count); err != nil {
			return nil, fmt.Errorf("failed to scan delivery count by method: %w", err)
		}
		result[entity.DeliveryMethod(method)] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate delivery count by method rows: %w", err)
	}

	return result, nil
}

// GetAverageDeliveryTime calculates average delivery time for a method
func (r *DeliveryRepository) GetAverageDeliveryTime(ctx context.Context, method entity.DeliveryMethod, startDate, endDate time.Time) (float64, error) {
	query := `
		SELECT AVG(EXTRACT(EPOCH FROM (completed_at - created_at))/3600) as avg_hours
		FROM deliveries 
		WHERE delivery_method = $1 
		AND completed_at IS NOT NULL 
		AND created_at >= $2 AND created_at <= $3`

	var avgHours sql.NullFloat64
	err := r.db.QueryRowContext(ctx, query, string(method), startDate, endDate).Scan(&avgHours)
	if err != nil {
		return 0, fmt.Errorf("failed to get average delivery time: %w", err)
	}

	if !avgHours.Valid {
		return 0, nil
	}

	return avgHours.Float64, nil
}

// GetTotalDeliveryFees calculates total delivery fees in date range
func (r *DeliveryRepository) GetTotalDeliveryFees(ctx context.Context, startDate, endDate time.Time) (float64, error) {
	query := `
		SELECT COALESCE(SUM(delivery_fee), 0) as total
		FROM deliveries 
		WHERE created_at >= $1 AND created_at <= $2`

	var total float64
	err := r.db.QueryRowContext(ctx, query, startDate, endDate).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get total delivery fees: %w", err)
	}

	return total, nil
}

// GetDeliveriesByIDs retrieves multiple deliveries by their IDs
func (r *DeliveryRepository) GetDeliveriesByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.DeliveryOrder, error) {
	if len(ids) == 0 {
		return []*entity.DeliveryOrder{}, nil
	}

	query := `
		SELECT id, order_id, customer_id, customer_address_id, delivery_method,
			   priority_level, delivery_fee, cod_amount, weight, volume,
			   provider_id, provider_order_id, tracking_number, vehicle_id,
			   route_id, estimated_delivery_time, status, notes, attempts,
			   completed_at, created_at, updated_at
		FROM deliveries 
		WHERE id = ANY($1)
		ORDER BY created_at DESC`

	var deliveries []*entity.DeliveryOrder
	err := r.db.SelectContext(ctx, &deliveries, query, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get deliveries by IDs: %w", err)
	}

	return deliveries, nil
}

// GetOverdueDeliveries retrieves deliveries that are overdue
func (r *DeliveryRepository) GetOverdueDeliveries(ctx context.Context) ([]*entity.DeliveryOrder, error) {
	now := time.Now()
	
	query := `
		SELECT id, order_id, customer_id, customer_address_id, delivery_method,
			   priority_level, delivery_fee, cod_amount, weight, volume,
			   provider_id, provider_order_id, tracking_number, vehicle_id,
			   route_id, estimated_delivery_time, status, notes, attempts,
			   completed_at, created_at, updated_at
		FROM deliveries 
		WHERE estimated_delivery_time < $1 
		AND status NOT IN ($2, $3, $4, $5)
		ORDER BY estimated_delivery_time ASC`

	var deliveries []*entity.DeliveryOrder
	err := r.db.SelectContext(ctx, &deliveries, query, now,
		entity.DeliveryStatusDelivered,
		entity.DeliveryStatusCancelled,
		entity.DeliveryStatusFailed)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue deliveries: %w", err)
	}

	return deliveries, nil
}
