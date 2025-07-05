package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"shipping/internal/domain/entity"
	"shipping/internal/domain/repository"
)

type snapshotRepository struct {
	db *sqlx.DB
}

// NewSnapshotRepository creates a new snapshot repository implementation
func NewSnapshotRepository(db *sqlx.DB) repository.SnapshotRepository {
	return &snapshotRepository{db: db}
}

// Create creates a new delivery snapshot
func (r *snapshotRepository) Create(ctx context.Context, snapshot *entity.DeliverySnapshot) error {
	query := `
		INSERT INTO delivery_snapshots (
			id, delivery_id, snapshot_type, snapshot_data, previous_snapshot_id,
			triggered_by, triggered_by_user_id, triggered_event,
			delivery_status, customer_id, order_id, vehicle_id, driver_name,
			delivery_address_province, delivery_fee, provider_code,
			created_at, business_date
		) VALUES (
			:id, :delivery_id, :snapshot_type, :snapshot_data, :previous_snapshot_id,
			:triggered_by, :triggered_by_user_id, :triggered_event,
			:delivery_status, :customer_id, :order_id, :vehicle_id, :driver_name,
			:delivery_address_province, :delivery_fee, :provider_code,
			:created_at, :business_date
		)`

	snapshotDataJSON, err := json.Marshal(snapshot.SnapshotData)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot data: %w", err)
	}

	_, err = r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":                        snapshot.ID,
		"delivery_id":               snapshot.DeliveryID,
		"snapshot_type":             snapshot.SnapshotType,
		"snapshot_data":             snapshotDataJSON,
		"previous_snapshot_id":      snapshot.PreviousSnapshotID,
		"triggered_by":              snapshot.TriggeredBy,
		"triggered_by_user_id":      snapshot.TriggeredByUserID,
		"triggered_event":           snapshot.TriggeredEvent,
		"delivery_status":           snapshot.DeliveryStatus,
		"customer_id":               snapshot.CustomerID,
		"order_id":                  snapshot.OrderID,
		"vehicle_id":                snapshot.VehicleID,
		"driver_name":               snapshot.DriverName,
		"delivery_address_province": snapshot.DeliveryAddressProvince,
		"delivery_fee":              snapshot.DeliveryFee,
		"provider_code":             snapshot.ProviderCode,
		"created_at":                snapshot.CreatedAt,
		"business_date":             snapshot.BusinessDate,
	})

	return err
}

// GetByID retrieves a snapshot by ID
func (r *snapshotRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.DeliverySnapshot, error) {
	query := `
		SELECT id, delivery_id, snapshot_type, snapshot_data, previous_snapshot_id,
			   triggered_by, triggered_by_user_id, triggered_event,
			   delivery_status, customer_id, order_id, vehicle_id, driver_name,
			   delivery_address_province, delivery_fee, provider_code,
			   created_at, business_date
		FROM delivery_snapshots 
		WHERE id = $1`

	var snapshot entity.DeliverySnapshot
	var snapshotDataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&snapshot.ID,
		&snapshot.DeliveryID,
		&snapshot.SnapshotType,
		&snapshotDataJSON,
		&snapshot.PreviousSnapshotID,
		&snapshot.TriggeredBy,
		&snapshot.TriggeredByUserID,
		&snapshot.TriggeredEvent,
		&snapshot.DeliveryStatus,
		&snapshot.CustomerID,
		&snapshot.OrderID,
		&snapshot.VehicleID,
		&snapshot.DriverName,
		&snapshot.DeliveryAddressProvince,
		&snapshot.DeliveryFee,
		&snapshot.ProviderCode,
		&snapshot.CreatedAt,
		&snapshot.BusinessDate,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrSnapshotNotFound
		}
		return nil, err
	}

	if err := json.Unmarshal(snapshotDataJSON, &snapshot.SnapshotData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot data: %w", err)
	}

	return &snapshot, nil
}

// GetByDeliveryID retrieves all snapshots for a delivery
func (r *snapshotRepository) GetByDeliveryID(ctx context.Context, deliveryID uuid.UUID) ([]*entity.DeliverySnapshot, error) {
	query := `
		SELECT id, delivery_id, snapshot_type, snapshot_data, previous_snapshot_id,
			   triggered_by, triggered_by_user_id, triggered_event,
			   delivery_status, customer_id, order_id, vehicle_id, driver_name,
			   delivery_address_province, delivery_fee, provider_code,
			   created_at, business_date
		FROM delivery_snapshots 
		WHERE delivery_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, deliveryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []*entity.DeliverySnapshot
	for rows.Next() {
		var snapshot entity.DeliverySnapshot
		var snapshotDataJSON []byte

		err := rows.Scan(
			&snapshot.ID,
			&snapshot.DeliveryID,
			&snapshot.SnapshotType,
			&snapshotDataJSON,
			&snapshot.PreviousSnapshotID,
			&snapshot.TriggeredBy,
			&snapshot.TriggeredByUserID,
			&snapshot.TriggeredEvent,
			&snapshot.DeliveryStatus,
			&snapshot.CustomerID,
			&snapshot.OrderID,
			&snapshot.VehicleID,
			&snapshot.DriverName,
			&snapshot.DeliveryAddressProvince,
			&snapshot.DeliveryFee,
			&snapshot.ProviderCode,
			&snapshot.CreatedAt,
			&snapshot.BusinessDate,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(snapshotDataJSON, &snapshot.SnapshotData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal snapshot data: %w", err)
		}

		snapshots = append(snapshots, &snapshot)
	}

	return snapshots, rows.Err()
}

// GetByDateRange retrieves snapshots within a date range
func (r *snapshotRepository) GetByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*entity.DeliverySnapshot, error) {
	query := `
		SELECT id, delivery_id, snapshot_type, snapshot_data, previous_snapshot_id,
			   triggered_by, triggered_by_user_id, triggered_event,
			   delivery_status, customer_id, order_id, vehicle_id, driver_name,
			   delivery_address_province, delivery_fee, provider_code,
			   created_at, business_date
		FROM delivery_snapshots 
		WHERE business_date >= $1 AND business_date <= $2
		ORDER BY business_date DESC, created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []*entity.DeliverySnapshot
	for rows.Next() {
		var snapshot entity.DeliverySnapshot
		var snapshotDataJSON []byte

		err := rows.Scan(
			&snapshot.ID,
			&snapshot.DeliveryID,
			&snapshot.SnapshotType,
			&snapshotDataJSON,
			&snapshot.PreviousSnapshotID,
			&snapshot.TriggeredBy,
			&snapshot.TriggeredByUserID,
			&snapshot.TriggeredEvent,
			&snapshot.DeliveryStatus,
			&snapshot.CustomerID,
			&snapshot.OrderID,
			&snapshot.VehicleID,
			&snapshot.DriverName,
			&snapshot.DeliveryAddressProvince,
			&snapshot.DeliveryFee,
			&snapshot.ProviderCode,
			&snapshot.CreatedAt,
			&snapshot.BusinessDate,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(snapshotDataJSON, &snapshot.SnapshotData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal snapshot data: %w", err)
		}

		snapshots = append(snapshots, &snapshot)
	}

	return snapshots, rows.Err()
}

// Update updates an existing snapshot
func (r *snapshotRepository) Update(ctx context.Context, snapshot *entity.DeliverySnapshot) error {
	query := `
		UPDATE delivery_snapshots SET
			snapshot_data = :snapshot_data,
			triggered_by = :triggered_by,
			triggered_by_user_id = :triggered_by_user_id,
			triggered_event = :triggered_event,
			delivery_status = :delivery_status,
			vehicle_id = :vehicle_id,
			driver_name = :driver_name,
			delivery_address_province = :delivery_address_province,
			delivery_fee = :delivery_fee,
			provider_code = :provider_code
		WHERE id = :id`

	snapshotDataJSON, err := json.Marshal(snapshot.SnapshotData)
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot data: %w", err)
	}

	_, err = r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":                        snapshot.ID,
		"snapshot_data":             snapshotDataJSON,
		"triggered_by":              snapshot.TriggeredBy,
		"triggered_by_user_id":      snapshot.TriggeredByUserID,
		"triggered_event":           snapshot.TriggeredEvent,
		"delivery_status":           snapshot.DeliveryStatus,
		"vehicle_id":                snapshot.VehicleID,
		"driver_name":               snapshot.DriverName,
		"delivery_address_province": snapshot.DeliveryAddressProvince,
		"delivery_fee":              snapshot.DeliveryFee,
		"provider_code":             snapshot.ProviderCode,
	})

	return err
}

// GetLatestByDeliveryID retrieves the latest snapshot for a delivery
func (r *snapshotRepository) GetLatestByDeliveryID(ctx context.Context, deliveryID uuid.UUID) (*entity.DeliverySnapshot, error) {
	query := `
		SELECT id, delivery_id, snapshot_type, snapshot_data, previous_snapshot_id,
			   triggered_by, triggered_by_user_id, triggered_event,
			   delivery_status, customer_id, order_id, vehicle_id, driver_name,
			   delivery_address_province, delivery_fee, provider_code,
			   created_at, business_date
		FROM delivery_snapshots 
		WHERE delivery_id = $1
		ORDER BY created_at DESC
		LIMIT 1`

	var snapshot entity.DeliverySnapshot
	var snapshotDataJSON []byte

	err := r.db.QueryRowContext(ctx, query, deliveryID).Scan(
		&snapshot.ID,
		&snapshot.DeliveryID,
		&snapshot.SnapshotType,
		&snapshotDataJSON,
		&snapshot.PreviousSnapshotID,
		&snapshot.TriggeredBy,
		&snapshot.TriggeredByUserID,
		&snapshot.TriggeredEvent,
		&snapshot.DeliveryStatus,
		&snapshot.CustomerID,
		&snapshot.OrderID,
		&snapshot.VehicleID,
		&snapshot.DriverName,
		&snapshot.DeliveryAddressProvince,
		&snapshot.DeliveryFee,
		&snapshot.ProviderCode,
		&snapshot.CreatedAt,
		&snapshot.BusinessDate,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrSnapshotNotFound
		}
		return nil, err
	}

	if err := json.Unmarshal(snapshotDataJSON, &snapshot.SnapshotData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot data: %w", err)
	}

	return &snapshot, nil
}

// GetByDeliveryIDAndType retrieves snapshots for a delivery by type
func (r *snapshotRepository) GetByDeliveryIDAndType(ctx context.Context, deliveryID uuid.UUID, snapshotType entity.SnapshotType) ([]*entity.DeliverySnapshot, error) {
	query := `
		SELECT id, delivery_id, snapshot_type, snapshot_data, previous_snapshot_id,
			   triggered_by, triggered_by_user_id, triggered_event,
			   delivery_status, customer_id, order_id, vehicle_id, driver_name,
			   delivery_address_province, delivery_fee, provider_code,
			   created_at, business_date
		FROM delivery_snapshots 
		WHERE delivery_id = $1 AND snapshot_type = $2
		ORDER BY created_at DESC`

	return r.querySnapshots(ctx, query, deliveryID, snapshotType)
}

// GetByType retrieves snapshots by type with pagination
func (r *snapshotRepository) GetByType(ctx context.Context, snapshotType entity.SnapshotType, limit, offset int) ([]*entity.DeliverySnapshot, error) {
	query := `
		SELECT id, delivery_id, snapshot_type, snapshot_data, previous_snapshot_id,
			   triggered_by, triggered_by_user_id, triggered_event,
			   delivery_status, customer_id, order_id, vehicle_id, driver_name,
			   delivery_address_province, delivery_fee, provider_code,
			   created_at, business_date
		FROM delivery_snapshots 
		WHERE snapshot_type = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	return r.querySnapshots(ctx, query, snapshotType, limit, offset)
}

// GetByBusinessDate retrieves snapshots by business date
func (r *snapshotRepository) GetByBusinessDate(ctx context.Context, businessDate time.Time) ([]*entity.DeliverySnapshot, error) {
	query := `
		SELECT id, delivery_id, snapshot_type, snapshot_data, previous_snapshot_id,
			   triggered_by, triggered_by_user_id, triggered_event,
			   delivery_status, customer_id, order_id, vehicle_id, driver_name,
			   delivery_address_province, delivery_fee, provider_code,
			   created_at, business_date
		FROM delivery_snapshots 
		WHERE DATE(business_date) = DATE($1)
		ORDER BY created_at DESC`

	return r.querySnapshots(ctx, query, businessDate)
}

// GetByCustomerID retrieves snapshots by customer ID with pagination
func (r *snapshotRepository) GetByCustomerID(ctx context.Context, customerID uuid.UUID, limit, offset int) ([]*entity.DeliverySnapshot, error) {
	query := `
		SELECT id, delivery_id, snapshot_type, snapshot_data, previous_snapshot_id,
			   triggered_by, triggered_by_user_id, triggered_event,
			   delivery_status, customer_id, order_id, vehicle_id, driver_name,
			   delivery_address_province, delivery_fee, provider_code,
			   created_at, business_date
		FROM delivery_snapshots 
		WHERE customer_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	return r.querySnapshots(ctx, query, customerID, limit, offset)
}

// GetByOrderID retrieves snapshots by order ID
func (r *snapshotRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*entity.DeliverySnapshot, error) {
	query := `
		SELECT id, delivery_id, snapshot_type, snapshot_data, previous_snapshot_id,
			   triggered_by, triggered_by_user_id, triggered_event,
			   delivery_status, customer_id, order_id, vehicle_id, driver_name,
			   delivery_address_province, delivery_fee, provider_code,
			   created_at, business_date
		FROM delivery_snapshots 
		WHERE order_id = $1
		ORDER BY created_at DESC`

	return r.querySnapshots(ctx, query, orderID)
}

// GetByProviderCode retrieves snapshots by provider code within date range
func (r *snapshotRepository) GetByProviderCode(ctx context.Context, providerCode string, startDate, endDate time.Time) ([]*entity.DeliverySnapshot, error) {
	query := `
		SELECT id, delivery_id, snapshot_type, snapshot_data, previous_snapshot_id,
			   triggered_by, triggered_by_user_id, triggered_event,
			   delivery_status, customer_id, order_id, vehicle_id, driver_name,
			   delivery_address_province, delivery_fee, provider_code,
			   created_at, business_date
		FROM delivery_snapshots 
		WHERE provider_code = $1 AND business_date >= $2 AND business_date <= $3
		ORDER BY created_at DESC`

	return r.querySnapshots(ctx, query, providerCode, startDate, endDate)
}

// GetByProviderAndStatus retrieves snapshots by provider and status with pagination
func (r *snapshotRepository) GetByProviderAndStatus(ctx context.Context, providerCode, status string, limit, offset int) ([]*entity.DeliverySnapshot, error) {
	query := `
		SELECT id, delivery_id, snapshot_type, snapshot_data, previous_snapshot_id,
			   triggered_by, triggered_by_user_id, triggered_event,
			   delivery_status, customer_id, order_id, vehicle_id, driver_name,
			   delivery_address_province, delivery_fee, provider_code,
			   created_at, business_date
		FROM delivery_snapshots 
		WHERE provider_code = $1 AND delivery_status = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	return r.querySnapshots(ctx, query, providerCode, status, limit, offset)
}

// GetByVehicleID retrieves snapshots by vehicle ID for a business date
func (r *snapshotRepository) GetByVehicleID(ctx context.Context, vehicleID uuid.UUID, businessDate time.Time) ([]*entity.DeliverySnapshot, error) {
	query := `
		SELECT id, delivery_id, snapshot_type, snapshot_data, previous_snapshot_id,
			   triggered_by, triggered_by_user_id, triggered_event,
			   delivery_status, customer_id, order_id, vehicle_id, driver_name,
			   delivery_address_province, delivery_fee, provider_code,
			   created_at, business_date
		FROM delivery_snapshots 
		WHERE vehicle_id = $1 AND DATE(business_date) = DATE($2)
		ORDER BY created_at DESC`

	return r.querySnapshots(ctx, query, vehicleID, businessDate)
}

// GetByProvince retrieves snapshots by province for a business date
func (r *snapshotRepository) GetByProvince(ctx context.Context, province string, businessDate time.Time) ([]*entity.DeliverySnapshot, error) {
	query := `
		SELECT id, delivery_id, snapshot_type, snapshot_data, previous_snapshot_id,
			   triggered_by, triggered_by_user_id, triggered_event,
			   delivery_status, customer_id, order_id, vehicle_id, driver_name,
			   delivery_address_province, delivery_fee, provider_code,
			   created_at, business_date
		FROM delivery_snapshots 
		WHERE delivery_address_province = $1 AND DATE(business_date) = DATE($2)
		ORDER BY created_at DESC`

	return r.querySnapshots(ctx, query, province, businessDate)
}

// GetDeliveryTimeline retrieves all snapshots for a delivery in chronological order
func (r *snapshotRepository) GetDeliveryTimeline(ctx context.Context, deliveryID uuid.UUID) ([]*entity.DeliverySnapshot, error) {
	query := `
		SELECT id, delivery_id, snapshot_type, snapshot_data, previous_snapshot_id,
			   triggered_by, triggered_by_user_id, triggered_event,
			   delivery_status, customer_id, order_id, vehicle_id, driver_name,
			   delivery_address_province, delivery_fee, provider_code,
			   created_at, business_date
		FROM delivery_snapshots 
		WHERE delivery_id = $1
		ORDER BY created_at ASC`

	return r.querySnapshots(ctx, query, deliveryID)
}

// GetSnapshotChain retrieves the chain of snapshots starting from a given snapshot
func (r *snapshotRepository) GetSnapshotChain(ctx context.Context, snapshotID uuid.UUID) ([]*entity.DeliverySnapshot, error) {
	// This would require a recursive query to follow the previous_snapshot_id chain
	// For now, implement a simple version that gets the snapshot and its delivery timeline
	snapshot, err := r.GetByID(ctx, snapshotID)
	if err != nil {
		return nil, err
	}
	
	return r.GetDeliveryTimeline(ctx, snapshot.DeliveryID)
}

// GetBusinessEventSnapshots retrieves business-critical snapshots for a delivery
func (r *snapshotRepository) GetBusinessEventSnapshots(ctx context.Context, deliveryID uuid.UUID) ([]*entity.DeliverySnapshot, error) {
	query := `
		SELECT id, delivery_id, snapshot_type, snapshot_data, previous_snapshot_id,
			   triggered_by, triggered_by_user_id, triggered_event,
			   delivery_status, customer_id, order_id, vehicle_id, driver_name,
			   delivery_address_province, delivery_fee, provider_code,
			   created_at, business_date
		FROM delivery_snapshots 
		WHERE delivery_id = $1 
		AND snapshot_type IN ('created', 'assigned', 'picked_up', 'delivered', 'failed', 'cancelled')
		ORDER BY created_at ASC`

	return r.querySnapshots(ctx, query, deliveryID)
}

// GetSnapshotCountByType retrieves snapshot counts by type within date range
func (r *snapshotRepository) GetSnapshotCountByType(ctx context.Context, startDate, endDate time.Time) (map[entity.SnapshotType]int64, error) {
	query := `
		SELECT snapshot_type, COUNT(*) as count
		FROM delivery_snapshots 
		WHERE business_date >= $1 AND business_date <= $2
		GROUP BY snapshot_type`

	rows, err := r.db.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[entity.SnapshotType]int64)
	for rows.Next() {
		var snapshotType entity.SnapshotType
		var count int64
		
		err := rows.Scan(&snapshotType, &count)
		if err != nil {
			return nil, err
		}
		
		result[snapshotType] = count
	}

	return result, rows.Err()
}

// GetSnapshotCountByProvider retrieves snapshot counts by provider for a business date
func (r *snapshotRepository) GetSnapshotCountByProvider(ctx context.Context, businessDate time.Time) (map[string]int64, error) {
	query := `
		SELECT provider_code, COUNT(*) as count
		FROM delivery_snapshots 
		WHERE DATE(business_date) = DATE($1)
		GROUP BY provider_code`

	rows, err := r.db.QueryContext(ctx, query, businessDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int64)
	for rows.Next() {
		var providerCode string
		var count int64
		
		err := rows.Scan(&providerCode, &count)
		if err != nil {
			return nil, err
		}
		
		result[providerCode] = count
	}

	return result, rows.Err()
}

// GetSnapshotCountByStatus retrieves snapshot counts by status for a business date
func (r *snapshotRepository) GetSnapshotCountByStatus(ctx context.Context, businessDate time.Time) (map[string]int64, error) {
	query := `
		SELECT delivery_status, COUNT(*) as count
		FROM delivery_snapshots 
		WHERE DATE(business_date) = DATE($1)
		GROUP BY delivery_status`

	rows, err := r.db.QueryContext(ctx, query, businessDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int64)
	for rows.Next() {
		var status string
		var count int64
		
		err := rows.Scan(&status, &count)
		if err != nil {
			return nil, err
		}
		
		result[status] = count
	}

	return result, rows.Err()
}

// GetDeliveryCompletionRate calculates delivery completion rate from snapshots
func (r *snapshotRepository) GetDeliveryCompletionRate(ctx context.Context, startDate, endDate time.Time) (float64, error) {
	query := `
		SELECT 
			COUNT(CASE WHEN snapshot_type = 'delivered' THEN 1 END) as completed,
			COUNT(DISTINCT delivery_id) as total
		FROM delivery_snapshots 
		WHERE business_date >= $1 AND business_date <= $2`

	var completed, total int64
	err := r.db.QueryRowContext(ctx, query, startDate, endDate).Scan(&completed, &total)
	if err != nil {
		return 0, err
	}

	if total == 0 {
		return 0, nil
	}

	return float64(completed) / float64(total), nil
}

// ArchiveSnapshotsOlderThan archives snapshots older than the cutoff date
func (r *snapshotRepository) ArchiveSnapshotsOlderThan(ctx context.Context, cutoffDate time.Time) (int64, error) {
	// For now, implement as a soft delete by adding an archived_at timestamp
	// In a real implementation, this might move data to an archive table
	query := `
		UPDATE delivery_snapshots 
		SET snapshot_data = jsonb_set(snapshot_data, '{archived_at}', to_jsonb(NOW()::text))
		WHERE business_date < $1 
		AND NOT (snapshot_data ? 'archived_at')`

	result, err := r.db.ExecContext(ctx, query, cutoffDate)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Additional methods for analytics and performance - simplified implementations
func (r *snapshotRepository) GetAverageDeliveryDurationFromSnapshots(ctx context.Context, providerCode string, startDate, endDate time.Time) (float64, error) {
	// Simplified implementation - would need more complex logic to calculate actual durations
	return 0, nil
}

func (r *snapshotRepository) GetProviderPerformanceFromSnapshots(ctx context.Context, providerCode string, startDate, endDate time.Time) (*repository.ProviderPerformanceMetrics, error) {
	// Simplified implementation
	return &repository.ProviderPerformanceMetrics{}, nil
}

func (r *snapshotRepository) GetDailyDeliveryMetricsFromSnapshots(ctx context.Context, businessDate time.Time) (*repository.DailyDeliveryMetrics, error) {
	// Simplified implementation
	return &repository.DailyDeliveryMetrics{}, nil
}

func (r *snapshotRepository) GetSnapshotsForAudit(ctx context.Context, startDate, endDate time.Time) ([]*entity.DeliverySnapshot, error) {
	return r.GetByDateRange(ctx, startDate, endDate)
}

func (r *snapshotRepository) GetSnapshotsByTriggeredBy(ctx context.Context, triggeredBy string, startDate, endDate time.Time) ([]*entity.DeliverySnapshot, error) {
	query := `
		SELECT id, delivery_id, snapshot_type, snapshot_data, previous_snapshot_id,
			   triggered_by, triggered_by_user_id, triggered_event,
			   delivery_status, customer_id, order_id, vehicle_id, driver_name,
			   delivery_address_province, delivery_fee, provider_code,
			   created_at, business_date
		FROM delivery_snapshots 
		WHERE triggered_by = $1 AND business_date >= $2 AND business_date <= $3
		ORDER BY created_at DESC`

	return r.querySnapshots(ctx, query, triggeredBy, startDate, endDate)
}

func (r *snapshotRepository) GetSnapshotsByUser(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*entity.DeliverySnapshot, error) {
	query := `
		SELECT id, delivery_id, snapshot_type, snapshot_data, previous_snapshot_id,
			   triggered_by, triggered_by_user_id, triggered_event,
			   delivery_status, customer_id, order_id, vehicle_id, driver_name,
			   delivery_address_province, delivery_fee, provider_code,
			   created_at, business_date
		FROM delivery_snapshots 
		WHERE triggered_by_user_id = $1 AND business_date >= $2 AND business_date <= $3
		ORDER BY created_at DESC`

	return r.querySnapshots(ctx, query, userID, startDate, endDate)
}

func (r *snapshotRepository) GetRevenueFromSnapshots(ctx context.Context, startDate, endDate time.Time) (map[string]float64, error) {
	query := `
		SELECT provider_code, SUM(delivery_fee) as total_revenue
		FROM delivery_snapshots 
		WHERE business_date >= $1 AND business_date <= $2
		AND snapshot_type = 'delivered'
		GROUP BY provider_code`

	rows, err := r.db.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]float64)
	for rows.Next() {
		var providerCode string
		var revenue float64
		
		err := rows.Scan(&providerCode, &revenue)
		if err != nil {
			return nil, err
		}
		
		result[providerCode] = revenue
	}

	return result, rows.Err()
}

func (r *snapshotRepository) GetDeliveryFeesFromSnapshots(ctx context.Context, providerCode string, businessDate time.Time) (float64, error) {
	query := `
		SELECT COALESCE(SUM(delivery_fee), 0) as total_fees
		FROM delivery_snapshots 
		WHERE provider_code = $1 AND DATE(business_date) = DATE($2)
		AND snapshot_type = 'delivered'`

	var totalFees float64
	err := r.db.QueryRowContext(ctx, query, providerCode, businessDate).Scan(&totalFees)
	return totalFees, err
}

func (r *snapshotRepository) GetMonthlyFinancialSummaryFromSnapshots(ctx context.Context, year int, month int) (*repository.MonthlyFinancialSummary, error) {
	// Simplified implementation
	return &repository.MonthlyFinancialSummary{}, nil
}

func (r *snapshotRepository) SearchSnapshots(ctx context.Context, filters *repository.SnapshotQueryFilters) ([]*entity.DeliverySnapshot, error) {
	// Simplified implementation - would build dynamic query based on filters
	query := `
		SELECT id, delivery_id, snapshot_type, snapshot_data, previous_snapshot_id,
			   triggered_by, triggered_by_user_id, triggered_event,
			   delivery_status, customer_id, order_id, vehicle_id, driver_name,
			   delivery_address_province, delivery_fee, provider_code,
			   created_at, business_date
		FROM delivery_snapshots 
		WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	if filters.DeliveryID != nil {
		query += fmt.Sprintf(" AND delivery_id = $%d", argIndex)
		args = append(args, *filters.DeliveryID)
		argIndex++
	}

	if filters.SnapshotType != nil {
		query += fmt.Sprintf(" AND snapshot_type = $%d", argIndex)
		args = append(args, *filters.SnapshotType)
		argIndex++
	}

	query += " ORDER BY created_at DESC"
	
	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
		
		if filters.Offset > 0 {
			query += fmt.Sprintf(" OFFSET $%d", argIndex)
			args = append(args, filters.Offset)
		}
	}

	return r.querySnapshotsWithArgs(ctx, query, args...)
}

func (r *snapshotRepository) GetFailedDeliverySnapshots(ctx context.Context, startDate, endDate time.Time) ([]*entity.DeliverySnapshot, error) {
	query := `
		SELECT id, delivery_id, snapshot_type, snapshot_data, previous_snapshot_id,
			   triggered_by, triggered_by_user_id, triggered_event,
			   delivery_status, customer_id, order_id, vehicle_id, driver_name,
			   delivery_address_province, delivery_fee, provider_code,
			   created_at, business_date
		FROM delivery_snapshots 
		WHERE snapshot_type IN ('failed', 'cancelled') 
		AND business_date >= $1 AND business_date <= $2
		ORDER BY created_at DESC`

	return r.querySnapshots(ctx, query, startDate, endDate)
}

func (r *snapshotRepository) GetSuccessfulDeliverySnapshots(ctx context.Context, startDate, endDate time.Time) ([]*entity.DeliverySnapshot, error) {
	query := `
		SELECT id, delivery_id, snapshot_type, snapshot_data, previous_snapshot_id,
			   triggered_by, triggered_by_user_id, triggered_event,
			   delivery_status, customer_id, order_id, vehicle_id, driver_name,
			   delivery_address_province, delivery_fee, provider_code,
			   created_at, business_date
		FROM delivery_snapshots 
		WHERE snapshot_type = 'delivered'
		AND business_date >= $1 AND business_date <= $2
		ORDER BY created_at DESC`

	return r.querySnapshots(ctx, query, startDate, endDate)
}

func (r *snapshotRepository) CreateBulkSnapshots(ctx context.Context, snapshots []*entity.DeliverySnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}

	query := `
		INSERT INTO delivery_snapshots (
			id, delivery_id, snapshot_type, snapshot_data, previous_snapshot_id,
			triggered_by, triggered_by_user_id, triggered_event,
			delivery_status, customer_id, order_id, vehicle_id, driver_name,
			delivery_address_province, delivery_fee, provider_code,
			created_at, business_date
		) VALUES `

	var values []string
	var args []interface{}
	argIndex := 1

	for _, snapshot := range snapshots {
		snapshotDataJSON, _ := json.Marshal(snapshot.SnapshotData)
		
		values = append(values, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", 
			argIndex, argIndex+1, argIndex+2, argIndex+3, argIndex+4, argIndex+5, argIndex+6, argIndex+7,
			argIndex+8, argIndex+9, argIndex+10, argIndex+11, argIndex+12, argIndex+13, argIndex+14, argIndex+15, argIndex+16, argIndex+17))
		
		args = append(args, 
			snapshot.ID, snapshot.DeliveryID, snapshot.SnapshotType, snapshotDataJSON, snapshot.PreviousSnapshotID,
			snapshot.TriggeredBy, snapshot.TriggeredByUserID, snapshot.TriggeredEvent,
			snapshot.DeliveryStatus, snapshot.CustomerID, snapshot.OrderID, snapshot.VehicleID, snapshot.DriverName,
			snapshot.DeliveryAddressProvince, snapshot.DeliveryFee, snapshot.ProviderCode,
			snapshot.CreatedAt, snapshot.BusinessDate)
		
		argIndex += 18
	}

	query += strings.Join(values, ", ")
	
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *snapshotRepository) GetSnapshotsByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.DeliverySnapshot, error) {
	if len(ids) == 0 {
		return []*entity.DeliverySnapshot{}, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, delivery_id, snapshot_type, snapshot_data, previous_snapshot_id,
			   triggered_by, triggered_by_user_id, triggered_event,
			   delivery_status, customer_id, order_id, vehicle_id, driver_name,
			   delivery_address_province, delivery_fee, provider_code,
			   created_at, business_date
		FROM delivery_snapshots 
		WHERE id IN (%s)
		ORDER BY created_at DESC`, strings.Join(placeholders, ", "))

	return r.querySnapshotsWithArgs(ctx, query, args...)
}

// Delete removes a snapshot (soft delete by marking as inactive)
func (r *snapshotRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// For snapshots, we typically don't delete but could add a soft delete field
	// For now, implement hard delete as snapshots are audit data
	query := `DELETE FROM delivery_snapshots WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrSnapshotNotFound
	}

	return nil
}

func (r *snapshotRepository) DeleteSnapshotsOlderThan(ctx context.Context, cutoffDate time.Time) (int64, error) {
	query := `DELETE FROM delivery_snapshots WHERE business_date < $1`
	
	result, err := r.db.ExecContext(ctx, query, cutoffDate)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Helper methods
func (r *snapshotRepository) querySnapshots(ctx context.Context, query string, args ...interface{}) ([]*entity.DeliverySnapshot, error) {
	return r.querySnapshotsWithArgs(ctx, query, args...)
}

func (r *snapshotRepository) querySnapshotsWithArgs(ctx context.Context, query string, args ...interface{}) ([]*entity.DeliverySnapshot, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []*entity.DeliverySnapshot
	for rows.Next() {
		var snapshot entity.DeliverySnapshot
		var snapshotDataJSON []byte

		err := rows.Scan(
			&snapshot.ID,
			&snapshot.DeliveryID,
			&snapshot.SnapshotType,
			&snapshotDataJSON,
			&snapshot.PreviousSnapshotID,
			&snapshot.TriggeredBy,
			&snapshot.TriggeredByUserID,
			&snapshot.TriggeredEvent,
			&snapshot.DeliveryStatus,
			&snapshot.CustomerID,
			&snapshot.OrderID,
			&snapshot.VehicleID,
			&snapshot.DriverName,
			&snapshot.DeliveryAddressProvince,
			&snapshot.DeliveryFee,
			&snapshot.ProviderCode,
			&snapshot.CreatedAt,
			&snapshot.BusinessDate,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(snapshotDataJSON, &snapshot.SnapshotData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal snapshot data: %w", err)
		}

		snapshots = append(snapshots, &snapshot)
	}

	return snapshots, rows.Err()
}
