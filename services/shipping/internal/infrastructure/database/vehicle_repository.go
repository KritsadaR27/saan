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

// VehicleRepository implements the vehicle repository interface
type VehicleRepository struct {
	db *sqlx.DB
}

// NewVehicleRepository creates a new vehicle repository
func NewVehicleRepository(db *sqlx.DB) repository.VehicleRepository {
	return &VehicleRepository{db: db}
}

// Create saves a new vehicle to the database
func (r *VehicleRepository) Create(ctx context.Context, vehicle *entity.DeliveryVehicle) error {
	query := `
		INSERT INTO vehicles (
			id, license_plate, vehicle_code, vehicle_type, brand, model, year,
			max_weight, max_volume, fuel_type, driver_id, status,
			current_latitude, current_longitude, notes, is_active,
			created_at, updated_at
		) VALUES (
			:id, :license_plate, :vehicle_code, :vehicle_type, :brand, :model, :year,
			:max_weight, :max_volume, :fuel_type, :driver_id, :status,
			:current_latitude, :current_longitude, :notes, :is_active,
			:created_at, :updated_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, vehicle)
	if err != nil {
		return fmt.Errorf("failed to create vehicle: %w", err)
	}

	return nil
}

// GetByID retrieves a vehicle by ID
func (r *VehicleRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.DeliveryVehicle, error) {
	query := `
		SELECT id, license_plate, vehicle_code, vehicle_type, brand, model, year,
			   max_weight, max_volume, fuel_type, driver_id, status,
			   current_latitude, current_longitude, notes, is_active,
			   created_at, updated_at
		FROM vehicles 
		WHERE id = $1 AND is_active = true`

	var vehicle entity.DeliveryVehicle
	err := r.db.GetContext(ctx, &vehicle, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("vehicle not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get vehicle: %w", err)
	}

	return &vehicle, nil
}

// GetByCode retrieves a vehicle by license plate
func (r *VehicleRepository) GetByCode(ctx context.Context, code string) (*entity.DeliveryVehicle, error) {
	query := `
		SELECT id, license_plate, vehicle_code, vehicle_type, brand, model, year,
			   max_weight, max_volume, fuel_type, driver_id, status,
			   current_latitude, current_longitude, notes, is_active,
			   created_at, updated_at
		FROM vehicles 
		WHERE license_plate = $1 AND is_active = true`

	var vehicle entity.DeliveryVehicle
	err := r.db.GetContext(ctx, &vehicle, query, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("vehicle not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get vehicle by code: %w", err)
	}

	return &vehicle, nil
}

// GetAvailable retrieves all available vehicles
func (r *VehicleRepository) GetAvailable(ctx context.Context) ([]*entity.DeliveryVehicle, error) {
	query := `
		SELECT id, license_plate, vehicle_code, vehicle_type, brand, model, year,
			   max_weight, max_volume, fuel_type, driver_id, status,
			   current_latitude, current_longitude, notes, is_active,
			   created_at, updated_at
		FROM vehicles 
		WHERE status = $1 AND is_active = true
		ORDER BY license_plate`

	var vehicles []*entity.DeliveryVehicle
	err := r.db.SelectContext(ctx, &vehicles, query, entity.VehicleStatusActive)
	if err != nil {
		return nil, fmt.Errorf("failed to get available vehicles: %w", err)
	}

	return vehicles, nil
}

// GetByType retrieves vehicles by type
func (r *VehicleRepository) GetByType(ctx context.Context, vehicleType entity.VehicleType) ([]*entity.DeliveryVehicle, error) {
	query := `
		SELECT id, license_plate, vehicle_code, vehicle_type, brand, model, year,
			   max_weight, max_volume, fuel_type, driver_id, status,
			   current_latitude, current_longitude, notes, is_active,
			   created_at, updated_at
		FROM vehicles 
		WHERE vehicle_type = $1 AND is_active = true
		ORDER BY license_plate`

	var vehicles []*entity.DeliveryVehicle
	err := r.db.SelectContext(ctx, &vehicles, query, vehicleType)
	if err != nil {
		return nil, fmt.Errorf("failed to get vehicles by type: %w", err)
	}

	return vehicles, nil
}

// GetByStatus retrieves vehicles by status
func (r *VehicleRepository) GetByStatus(ctx context.Context, status entity.VehicleStatus) ([]*entity.DeliveryVehicle, error) {
	query := `
		SELECT id, license_plate, vehicle_code, vehicle_type, brand, model, year,
			   max_weight, max_volume, fuel_type, driver_id, status,
			   current_latitude, current_longitude, notes, is_active,
			   created_at, updated_at
		FROM vehicles 
		WHERE status = $1 AND is_active = true
		ORDER BY license_plate`

	var vehicles []*entity.DeliveryVehicle
	err := r.db.SelectContext(ctx, &vehicles, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get vehicles by status: %w", err)
	}

	return vehicles, nil
}

// Update updates an existing vehicle
func (r *VehicleRepository) Update(ctx context.Context, vehicle *entity.DeliveryVehicle) error {
	query := `
		UPDATE vehicles SET
			brand = :brand,
			model = :model,
			year = :year,
			max_weight = :max_weight,
			max_volume = :max_volume,
			fuel_type = :fuel_type,
			driver_id = :driver_id,
			status = :status,
			current_latitude = :current_latitude,
			current_longitude = :current_longitude,
			notes = :notes,
			updated_at = :updated_at
		WHERE id = :id AND is_active = true`

	result, err := r.db.NamedExecContext(ctx, query, vehicle)
	if err != nil {
		return fmt.Errorf("failed to update vehicle: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("vehicle not found or already deleted")
	}

	return nil
}

// UpdateStatus updates the status of a vehicle
func (r *VehicleRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.VehicleStatus) error {
	query := `
		UPDATE vehicles SET
			status = $1,
			updated_at = $2
		WHERE id = $3 AND is_active = true`

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update vehicle status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("vehicle not found or already deleted")
	}

	return nil
}

// AssignDriver assigns a driver to a vehicle
func (r *VehicleRepository) AssignDriver(ctx context.Context, vehicleID, driverID uuid.UUID) error {
	query := `
		UPDATE vehicles SET
			driver_id = $1,
			updated_at = $2
		WHERE id = $3 AND is_active = true`

	result, err := r.db.ExecContext(ctx, query, driverID, time.Now(), vehicleID)
	if err != nil {
		return fmt.Errorf("failed to assign driver: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("vehicle not found or already deleted")
	}

	return nil
}

// UnassignDriver removes a driver from a vehicle
func (r *VehicleRepository) UnassignDriver(ctx context.Context, vehicleID uuid.UUID) error {
	query := `
		UPDATE vehicles SET
			driver_id = NULL,
			updated_at = $1
		WHERE id = $2 AND is_active = true`

	result, err := r.db.ExecContext(ctx, query, time.Now(), vehicleID)
	if err != nil {
		return fmt.Errorf("failed to unassign driver: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("vehicle not found or already deleted")
	}

	return nil
}

// SetMaintenanceMode sets maintenance mode for a vehicle
func (r *VehicleRepository) SetMaintenanceMode(ctx context.Context, vehicleID uuid.UUID, enabled bool) error {
	status := entity.VehicleStatusActive
	if enabled {
		status = entity.VehicleStatusMaintenance
	}

	return r.UpdateStatus(ctx, vehicleID, status)
}

// UpdateLocation updates the current location of a vehicle
func (r *VehicleRepository) UpdateLocation(ctx context.Context, vehicleID uuid.UUID, latitude, longitude float64) error {
	query := `
		UPDATE vehicles SET
			current_latitude = $1,
			current_longitude = $2,
			updated_at = $3
		WHERE id = $4 AND is_active = true`

	result, err := r.db.ExecContext(ctx, query, latitude, longitude, time.Now(), vehicleID)
	if err != nil {
		return fmt.Errorf("failed to update vehicle location: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("vehicle not found or already deleted")
	}

	return nil
}

// Delete soft deletes a vehicle
func (r *VehicleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE vehicles SET
			is_active = false,
			updated_at = $1
		WHERE id = $2 AND is_active = true`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete vehicle: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("vehicle not found or already deleted")
	}

	return nil
}

// GetVehicleMetrics retrieves performance metrics for a vehicle
func (r *VehicleRepository) GetVehicleMetrics(ctx context.Context, vehicleID uuid.UUID, startDate, endDate time.Time) (*repository.VehicleMetrics, error) {
	query := `
		SELECT 
			COUNT(CASE WHEN d.status = 'completed' THEN 1 END) as total_deliveries,
			COALESCE(SUM(CASE WHEN d.status = 'completed' THEN 1 ELSE 0 END), 0) as completed_deliveries,
			COALESCE(AVG(EXTRACT(EPOCH FROM (d.completed_at - d.created_at))/60), 0) as avg_delivery_time_minutes,
			COALESCE(SUM(d.delivery_fee), 0) as total_revenue
		FROM deliveries d 
		WHERE d.vehicle_id = $1 
		AND d.created_at >= $2 
		AND d.created_at <= $3`

	var metrics repository.VehicleMetrics
	err := r.db.GetContext(ctx, &metrics, query, vehicleID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get vehicle metrics: %w", err)
	}

	return &metrics, nil
}

// GetFleetUtilization retrieves utilization metrics for all vehicles on a specific date
func (r *VehicleRepository) GetFleetUtilization(ctx context.Context, date time.Time) (map[uuid.UUID]float64, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
		SELECT 
			v.id,
			COALESCE(COUNT(d.id), 0) as delivery_count,
			CASE 
				WHEN COUNT(d.id) = 0 THEN 0.0
				ELSE LEAST(COUNT(d.id) / 8.0, 1.0)  -- Assuming 8 deliveries per day is 100% utilization
			END as utilization
		FROM vehicles v
		LEFT JOIN deliveries d ON v.id = d.vehicle_id 
			AND d.created_at >= $1 
			AND d.created_at < $2
		WHERE v.is_active = true
		GROUP BY v.id`

	rows, err := r.db.QueryContext(ctx, query, startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("failed to get fleet utilization: %w", err)
	}
	defer rows.Close()

	utilization := make(map[uuid.UUID]float64)
	for rows.Next() {
		var vehicleID uuid.UUID
		var deliveryCount int
		var util float64

		if err := rows.Scan(&vehicleID, &deliveryCount, &util); err != nil {
			return nil, fmt.Errorf("failed to scan utilization row: %w", err)
		}

		utilization[vehicleID] = util
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating utilization rows: %w", err)
	}

	return utilization, nil
}

// GetAll retrieves all vehicles with pagination
func (r *VehicleRepository) GetAll(ctx context.Context, limit, offset int) ([]*entity.DeliveryVehicle, error) {
	query := `
		SELECT id, license_plate, vehicle_code, vehicle_type, brand, model, year,
			   max_weight, max_volume, fuel_type, driver_id, status,
			   current_latitude, current_longitude, notes, is_active,
			   created_at, updated_at
		FROM vehicles 
		WHERE is_active = true
		ORDER BY license_plate
		LIMIT $1 OFFSET $2`

	var vehicles []*entity.DeliveryVehicle
	err := r.db.SelectContext(ctx, &vehicles, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get all vehicles: %w", err)
	}

	return vehicles, nil
}

// GetByRoute retrieves vehicles assigned to a specific route
func (r *VehicleRepository) GetByRoute(ctx context.Context, route string) ([]*entity.DeliveryVehicle, error) {
	query := `
		SELECT id, license_plate, vehicle_code, vehicle_type, brand, model, year,
			   max_weight, max_volume, fuel_type, driver_id, status,
			   current_latitude, current_longitude, notes, is_active,
			   created_at, updated_at
		FROM vehicles 
		WHERE vehicle_code = $1 AND is_active = true
		ORDER BY license_plate`

	var vehicles []*entity.DeliveryVehicle
	err := r.db.SelectContext(ctx, &vehicles, query, route)
	if err != nil {
		return nil, fmt.Errorf("failed to get vehicles by route: %w", err)
	}

	return vehicles, nil
}

// AssignRoute assigns a route to a vehicle
func (r *VehicleRepository) AssignRoute(ctx context.Context, id uuid.UUID, route string) error {
	query := `
		UPDATE vehicles SET
			vehicle_code = $1,
			updated_at = $2
		WHERE id = $3 AND is_active = true`

	result, err := r.db.ExecContext(ctx, query, route, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to assign route: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("vehicle not found or already deleted")
	}

	return nil
}

// UnassignRoute removes route assignment from a vehicle
func (r *VehicleRepository) UnassignRoute(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE vehicles SET
			vehicle_code = NULL,
			updated_at = $1
		WHERE id = $2 AND is_active = true`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to unassign route: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("vehicle not found or already deleted")
	}

	return nil
}

// UpdateDailyCapacity updates the daily capacity of a vehicle
func (r *VehicleRepository) UpdateDailyCapacity(ctx context.Context, id uuid.UUID, capacity int) error {
	query := `
		UPDATE vehicles SET
			max_weight = $1,
			updated_at = $2
		WHERE id = $3 AND is_active = true`

	result, err := r.db.ExecContext(ctx, query, float64(capacity), time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update daily capacity: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("vehicle not found or already deleted")
	}

	return nil
}

// GetVehicleUtilization gets utilization for a specific vehicle on a date
func (r *VehicleRepository) GetVehicleUtilization(ctx context.Context, id uuid.UUID, date time.Time) (int, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
		SELECT COUNT(d.id) as delivery_count
		FROM deliveries d 
		WHERE d.vehicle_id = $1 
		AND d.created_at >= $2 
		AND d.created_at < $3`

	var count int
	err := r.db.GetContext(ctx, &count, query, id, startOfDay, endOfDay)
	if err != nil {
		return 0, fmt.Errorf("failed to get vehicle utilization: %w", err)
	}

	return count, nil
}

// UpdateOperatingCosts updates operating costs for a vehicle
func (r *VehicleRepository) UpdateOperatingCosts(ctx context.Context, id uuid.UUID, dailyCost, perKmCost float64) error {
	// Note: This would require additional fields in the vehicles table
	// For now, we'll return nil as a placeholder
	return nil
}

// GetVehiclesByCapacityRange retrieves vehicles within a capacity range
func (r *VehicleRepository) GetVehiclesByCapacityRange(ctx context.Context, minCapacity, maxCapacity int) ([]*entity.DeliveryVehicle, error) {
	query := `
		SELECT id, license_plate, vehicle_code, vehicle_type, brand, model, year,
			   max_weight, max_volume, fuel_type, driver_id, status,
			   current_latitude, current_longitude, notes, is_active,
			   created_at, updated_at
		FROM vehicles 
		WHERE max_weight >= $1 AND max_weight <= $2 AND is_active = true
		ORDER BY max_weight`

	var vehicles []*entity.DeliveryVehicle
	err := r.db.SelectContext(ctx, &vehicles, query, float64(minCapacity), float64(maxCapacity))
	if err != nil {
		return nil, fmt.Errorf("failed to get vehicles by capacity range: %w", err)
	}

	return vehicles, nil
}

// SearchVehicles searches vehicles with filters
func (r *VehicleRepository) SearchVehicles(ctx context.Context, filters *repository.VehicleQueryFilters) ([]*entity.DeliveryVehicle, error) {
	query := `
		SELECT id, license_plate, vehicle_code, vehicle_type, brand, model, year,
			   max_weight, max_volume, fuel_type, driver_id, status,
			   current_latitude, current_longitude, notes, is_active,
			   created_at, updated_at
		FROM vehicles 
		WHERE is_active = true`

	args := []interface{}{}
	argCount := 0

	if filters.VehicleType != nil {
		argCount++
		query += fmt.Sprintf(" AND vehicle_type = $%d", argCount)
		args = append(args, *filters.VehicleType)
	}

	if filters.Status != nil {
		argCount++
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, *filters.Status)
	}

	if filters.DriverID != nil {
		argCount++
		query += fmt.Sprintf(" AND driver_id = $%d", argCount)
		args = append(args, *filters.DriverID)
	}

	if filters.MinWeight != nil {
		argCount++
		query += fmt.Sprintf(" AND max_weight >= $%d", argCount)
		args = append(args, *filters.MinWeight)
	}

	if filters.MaxWeight != nil {
		argCount++
		query += fmt.Sprintf(" AND max_weight <= $%d", argCount)
		args = append(args, *filters.MaxWeight)
	}

	query += " ORDER BY license_plate"

	if filters.Limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filters.Limit)

		if filters.Offset > 0 {
			argCount++
			query += fmt.Sprintf(" OFFSET $%d", argCount)
			args = append(args, filters.Offset)
		}
	}

	var vehicles []*entity.DeliveryVehicle
	err := r.db.SelectContext(ctx, &vehicles, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search vehicles: %w", err)
	}

	return vehicles, nil
}

// GetActiveVehicles retrieves all active vehicles
func (r *VehicleRepository) GetActiveVehicles(ctx context.Context) ([]*entity.DeliveryVehicle, error) {
	query := `
		SELECT id, license_plate, vehicle_code, vehicle_type, brand, model, year,
			   max_weight, max_volume, fuel_type, driver_id, status,
			   current_latitude, current_longitude, notes, is_active,
			   created_at, updated_at
		FROM vehicles 
		WHERE is_active = true AND status = $1
		ORDER BY license_plate`

	var vehicles []*entity.DeliveryVehicle
	err := r.db.SelectContext(ctx, &vehicles, query, entity.VehicleStatusActive)
	if err != nil {
		return nil, fmt.Errorf("failed to get active vehicles: %w", err)
	}

	return vehicles, nil
}

// GetVehiclesNeedingMaintenance retrieves vehicles that need maintenance
func (r *VehicleRepository) GetVehiclesNeedingMaintenance(ctx context.Context) ([]*entity.DeliveryVehicle, error) {
	query := `
		SELECT id, license_plate, vehicle_code, vehicle_type, brand, model, year,
			   max_weight, max_volume, fuel_type, driver_id, status,
			   current_latitude, current_longitude, notes, is_active,
			   created_at, updated_at
		FROM vehicles 
		WHERE is_active = true AND status = $1
		ORDER BY license_plate`

	var vehicles []*entity.DeliveryVehicle
	err := r.db.SelectContext(ctx, &vehicles, query, entity.VehicleStatusMaintenance)
	if err != nil {
		return nil, fmt.Errorf("failed to get vehicles needing maintenance: %w", err)
	}

	return vehicles, nil
}
