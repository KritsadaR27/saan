package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"shipping/internal/domain/entity"
	"shipping/internal/domain/repository"
)

type routeRepository struct {
	db *sqlx.DB
}

// NewRouteRepository creates a new route repository implementation
func NewRouteRepository(db *sqlx.DB) repository.RouteRepository {
	return &routeRepository{db: db}
}

// Create creates a new delivery route
func (r *routeRepository) Create(ctx context.Context, route *entity.DeliveryRoute) error {
	query := `
		INSERT INTO delivery_routes (
			id, route_name, route_date, status, vehicle_id, driver_id,
			start_time, end_time, estimated_distance, estimated_orders,
			actual_distance, actual_orders_delivered, optimization_data,
			planning_data, created_at, updated_at
		) VALUES (
			:id, :route_name, :route_date, :status, :vehicle_id, :driver_id,
			:start_time, :end_time, :estimated_distance, :estimated_orders,
			:actual_distance, :actual_orders_delivered, :optimization_data,
			:planning_data, :created_at, :updated_at
		)`

	optimizationJSON, _ := json.Marshal(route.RouteOptimizationData)
	planningJSON, _ := json.Marshal(map[string]interface{}{
		"planned_start_time": route.PlannedStartTime,
		"planned_end_time":   route.PlannedEndTime,
		"total_planned_distance": route.TotalPlannedDistance,
		"total_planned_orders": route.TotalPlannedOrders,
	})

	_, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":                      route.ID,
		"route_name":              route.RouteName,
		"route_date":              route.RouteDate,
		"status":                  route.Status,
		"vehicle_id":              route.AssignedVehicleID,
		"driver_id":               route.AssignedDriverID,
		"start_time":              route.PlannedStartTime,
		"end_time":                route.PlannedEndTime,
		"estimated_distance":      route.TotalPlannedDistance,
		"estimated_orders":        route.TotalPlannedOrders,
		"actual_distance":         route.ActualDistance,
		"actual_orders_delivered": route.ActualOrdersDelivered,
		"optimization_data":       optimizationJSON,
		"planning_data":           planningJSON,
		"created_at":              route.CreatedAt,
		"updated_at":              route.UpdatedAt,
	})

	return err
}

// GetByID retrieves a route by ID
func (r *routeRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.DeliveryRoute, error) {
	query := `
		SELECT id, route_name, route_date, status, vehicle_id, driver_id,
			   start_time, end_time, estimated_distance, estimated_orders,
			   actual_distance, actual_orders_delivered, optimization_data,
			   planning_data, created_at, updated_at
		FROM delivery_routes 
		WHERE id = $1`

	var route entity.DeliveryRoute
	var optimizationJSON, planningJSON []byte

	err := r.db.QueryRowxContext(ctx, query, id).Scan(
		&route.ID, &route.RouteName, &route.RouteDate, &route.Status,
		&route.AssignedVehicleID, &route.AssignedDriverID, &route.PlannedStartTime, &route.PlannedEndTime,
		&route.TotalPlannedDistance, &route.TotalPlannedOrders, &route.ActualDistance,
		&route.ActualOrdersDelivered, &optimizationJSON, &planningJSON,
		&route.CreatedAt, &route.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("route not found")
		}
		return nil, err
	}

	if optimizationJSON != nil {
		json.Unmarshal(optimizationJSON, &route.RouteOptimizationData)
	}

	return &route, nil
}

// GetByName retrieves a route by name and date
func (r *routeRepository) GetByName(ctx context.Context, routeName string, routeDate time.Time) (*entity.DeliveryRoute, error) {
	query := `
		SELECT id, route_name, route_date, status, vehicle_id, driver_id,
			   start_time, end_time, estimated_distance, estimated_orders,
			   actual_distance, actual_orders_delivered, optimization_data,
			   planning_data, created_at, updated_at
		FROM delivery_routes 
		WHERE route_name = $1 AND route_date = $2`

	var route entity.DeliveryRoute
	var optimizationJSON, planningJSON []byte

	err := r.db.QueryRowxContext(ctx, query, routeName, routeDate).Scan(
		&route.ID, &route.RouteName, &route.RouteDate, &route.Status,
		&route.AssignedVehicleID, &route.AssignedDriverID, &route.PlannedStartTime, &route.PlannedEndTime,
		&route.TotalPlannedDistance, &route.TotalPlannedOrders, &route.ActualDistance,
		&route.ActualOrdersDelivered, &optimizationJSON, &planningJSON,
		&route.CreatedAt, &route.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("route not found")
		}
		return nil, err
	}

	if optimizationJSON != nil {
		json.Unmarshal(optimizationJSON, &route.RouteOptimizationData)
	}

	return &route, nil
}

// Update updates an existing route
func (r *routeRepository) Update(ctx context.Context, route *entity.DeliveryRoute) error {
	query := `
		UPDATE delivery_routes SET
			route_name = :route_name,
			route_date = :route_date,
			status = :status,
			vehicle_id = :vehicle_id,
			driver_id = :driver_id,
			start_time = :start_time,
			end_time = :end_time,
			estimated_distance = :estimated_distance,
			estimated_orders = :estimated_orders,
			actual_distance = :actual_distance,
			actual_orders_delivered = :actual_orders_delivered,
			optimization_data = :optimization_data,
			planning_data = :planning_data,
			updated_at = :updated_at
		WHERE id = :id`

	optimizationJSON, _ := json.Marshal(route.RouteOptimizationData)
	planningJSON, _ := json.Marshal(map[string]interface{}{
		"planned_start_time": route.PlannedStartTime,
		"planned_end_time":   route.PlannedEndTime,
		"total_planned_distance": route.TotalPlannedDistance,
		"total_planned_orders": route.TotalPlannedOrders,
	})
	route.UpdatedAt = time.Now()

	_, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":                      route.ID,
		"route_name":              route.RouteName,
		"route_date":              route.RouteDate,
		"status":                  route.Status,
		"vehicle_id":              route.AssignedVehicleID,
		"driver_id":               route.AssignedDriverID,
		"start_time":              route.PlannedStartTime,
		"end_time":                route.PlannedEndTime,
		"estimated_distance":      route.TotalPlannedDistance,
		"estimated_orders":        route.TotalPlannedOrders,
		"actual_distance":         route.ActualDistance,
		"actual_orders_delivered": route.ActualOrdersDelivered,
		"optimization_data":       optimizationJSON,
		"planning_data":           planningJSON,
		"updated_at":              route.UpdatedAt,
	})

	return err
}

// Delete deletes a route by ID
func (r *routeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM delivery_routes WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GetAll retrieves all routes with pagination
func (r *routeRepository) GetAll(ctx context.Context, limit, offset int) ([]*entity.DeliveryRoute, error) {
	query := `
		SELECT id, route_name, route_date, status, vehicle_id, driver_id,
			   start_time, end_time, estimated_distance, estimated_orders,
			   actual_distance, actual_orders_delivered, optimization_data,
			   planning_data, created_at, updated_at
		FROM delivery_routes 
		ORDER BY route_date DESC, route_name
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryxContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRoutes(rows)
}

// GetByDate retrieves routes by specific date
func (r *routeRepository) GetByDate(ctx context.Context, date time.Time) ([]*entity.DeliveryRoute, error) {
	query := `
		SELECT id, route_name, route_date, status, vehicle_id, driver_id,
			   start_time, end_time, estimated_distance, estimated_orders,
			   actual_distance, actual_orders_delivered, optimization_data,
			   planning_data, created_at, updated_at
		FROM delivery_routes 
		WHERE route_date = $1
		ORDER BY route_name`

	rows, err := r.db.QueryxContext(ctx, query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRoutes(rows)
}

// GetByDateRange retrieves routes within a date range
func (r *routeRepository) GetByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*entity.DeliveryRoute, error) {
	query := `
		SELECT id, route_name, route_date, status, vehicle_id, driver_id,
			   start_time, end_time, estimated_distance, estimated_orders,
			   actual_distance, actual_orders_delivered, optimization_data,
			   planning_data, created_at, updated_at
		FROM delivery_routes 
		WHERE route_date BETWEEN $1 AND $2
		ORDER BY route_date DESC, route_name`

	rows, err := r.db.QueryxContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRoutes(rows)
}

// GetByVehicle retrieves routes assigned to a vehicle on a specific date
func (r *routeRepository) GetByVehicle(ctx context.Context, vehicleID uuid.UUID, date time.Time) ([]*entity.DeliveryRoute, error) {
	query := `
		SELECT id, route_name, route_date, status, vehicle_id, driver_id,
			   start_time, end_time, estimated_distance, estimated_orders,
			   actual_distance, actual_orders_delivered, optimization_data,
			   planning_data, created_at, updated_at
		FROM delivery_routes 
		WHERE vehicle_id = $1 AND route_date = $2
		ORDER BY start_time`

	rows, err := r.db.QueryxContext(ctx, query, vehicleID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRoutes(rows)
}

// GetByDriver retrieves routes assigned to a driver on a specific date
func (r *routeRepository) GetByDriver(ctx context.Context, driverID uuid.UUID, date time.Time) ([]*entity.DeliveryRoute, error) {
	query := `
		SELECT id, route_name, route_date, status, vehicle_id, driver_id,
			   start_time, end_time, estimated_distance, estimated_orders,
			   actual_distance, actual_orders_delivered, optimization_data,
			   planning_data, created_at, updated_at
		FROM delivery_routes 
		WHERE driver_id = $1 AND route_date = $2
		ORDER BY start_time`

	rows, err := r.db.QueryxContext(ctx, query, driverID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRoutes(rows)
}

// GetByStatus retrieves routes by status
func (r *routeRepository) GetByStatus(ctx context.Context, status entity.RouteStatus) ([]*entity.DeliveryRoute, error) {
	query := `
		SELECT id, route_name, route_date, status, vehicle_id, driver_id,
			   start_time, end_time, estimated_distance, estimated_orders,
			   actual_distance, actual_orders_delivered, optimization_data,
			   planning_data, created_at, updated_at
		FROM delivery_routes 
		WHERE status = $1
		ORDER BY route_date DESC, route_name`

	rows, err := r.db.QueryxContext(ctx, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRoutes(rows)
}

// GetActiveRoutes retrieves all active routes (in_progress status)
func (r *routeRepository) GetActiveRoutes(ctx context.Context) ([]*entity.DeliveryRoute, error) {
	return r.GetByStatus(ctx, entity.RouteStatusInProgress)
}

// GetPlannedRoutes retrieves planned routes for a specific date
func (r *routeRepository) GetPlannedRoutes(ctx context.Context, date time.Time) ([]*entity.DeliveryRoute, error) {
	query := `
		SELECT id, route_name, route_date, status, vehicle_id, driver_id,
			   start_time, end_time, estimated_distance, estimated_orders,
			   actual_distance, actual_orders_delivered, optimization_data,
			   planning_data, created_at, updated_at
		FROM delivery_routes 
		WHERE route_date = $1 AND status = $2
		ORDER BY start_time`

	rows, err := r.db.QueryxContext(ctx, query, date, entity.RouteStatusPlanned)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRoutes(rows)
}

// GetInProgressRoutes retrieves all in-progress routes
func (r *routeRepository) GetInProgressRoutes(ctx context.Context) ([]*entity.DeliveryRoute, error) {
	return r.GetByStatus(ctx, entity.RouteStatusInProgress)
}

// GetCompletedRoutes retrieves completed routes for a specific date
func (r *routeRepository) GetCompletedRoutes(ctx context.Context, date time.Time) ([]*entity.DeliveryRoute, error) {
	query := `
		SELECT id, route_name, route_date, status, vehicle_id, driver_id,
			   start_time, end_time, estimated_distance, estimated_orders,
			   actual_distance, actual_orders_delivered, optimization_data,
			   planning_data, created_at, updated_at
		FROM delivery_routes 
		WHERE route_date = $1 AND status = $2
		ORDER BY end_time DESC`

	rows, err := r.db.QueryxContext(ctx, query, date, entity.RouteStatusCompleted)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRoutes(rows)
}

// AssignVehicle assigns a vehicle and optionally a driver to a route
func (r *routeRepository) AssignVehicle(ctx context.Context, routeID, vehicleID uuid.UUID, driverID *uuid.UUID) error {
	query := `
		UPDATE delivery_routes 
		SET vehicle_id = $2, driver_id = $3, updated_at = $4
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, routeID, vehicleID, driverID, time.Now())
	return err
}

// UnassignVehicle removes vehicle and driver assignment from a route
func (r *routeRepository) UnassignVehicle(ctx context.Context, routeID uuid.UUID) error {
	query := `
		UPDATE delivery_routes 
		SET vehicle_id = NULL, driver_id = NULL, updated_at = $2
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, routeID, time.Now())
	return err
}

// StartRoute starts a route (changes status to in_progress)
func (r *routeRepository) StartRoute(ctx context.Context, routeID uuid.UUID) error {
	query := `
		UPDATE delivery_routes 
		SET status = $2, start_time = $3, updated_at = $4
		WHERE id = $1 AND status = $5`

	result, err := r.db.ExecContext(ctx, query, routeID, entity.RouteStatusInProgress, time.Now(), time.Now(), entity.RouteStatusPlanned)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("route not found or not in planned status")
	}

	return nil
}

// CompleteRoute completes a route with actual metrics
func (r *routeRepository) CompleteRoute(ctx context.Context, routeID uuid.UUID, actualDistance float64, actualOrdersDelivered int) error {
	query := `
		UPDATE delivery_routes 
		SET status = $2, end_time = $3, actual_distance = $4, 
			actual_orders_delivered = $5, updated_at = $6
		WHERE id = $1 AND status = $7`

	result, err := r.db.ExecContext(ctx, query, routeID, entity.RouteStatusCompleted, time.Now(), 
		actualDistance, actualOrdersDelivered, time.Now(), entity.RouteStatusInProgress)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("route not found or not in progress")
	}

	return nil
}

// CancelRoute cancels a route
func (r *routeRepository) CancelRoute(ctx context.Context, routeID uuid.UUID) error {
	query := `
		UPDATE delivery_routes 
		SET status = $2, updated_at = $3
		WHERE id = $1 AND status IN ($4, $5)`

	_, err := r.db.ExecContext(ctx, query, routeID, entity.RouteStatusCancelled, time.Now(), 
		entity.RouteStatusPlanned, entity.RouteStatusInProgress)
	return err
}

// UpdateOptimizationData updates route optimization data
func (r *routeRepository) UpdateOptimizationData(ctx context.Context, routeID uuid.UUID, data map[string]interface{}) error {
	optimizationJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	query := `
		UPDATE delivery_routes 
		SET optimization_data = $2, updated_at = $3
		WHERE id = $1`

	_, err = r.db.ExecContext(ctx, query, routeID, optimizationJSON, time.Now())
	return err
}

// GetOptimizedRoutes retrieves optimized routes for a date with distance constraints
func (r *routeRepository) GetOptimizedRoutes(ctx context.Context, date time.Time, maxDistance float64) ([]*entity.DeliveryRoute, error) {
	query := `
		SELECT id, route_name, route_date, status, vehicle_id, driver_id,
			   start_time, end_time, estimated_distance, estimated_orders,
			   actual_distance, actual_orders_delivered, optimization_data,
			   planning_data, created_at, updated_at
		FROM delivery_routes 
		WHERE route_date = $1 AND estimated_distance <= $2 
			  AND optimization_data IS NOT NULL
		ORDER BY estimated_distance`

	rows, err := r.db.QueryxContext(ctx, query, date, maxDistance)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRoutes(rows)
}

// SetPlanning sets planning data for a route
func (r *routeRepository) SetPlanning(ctx context.Context, routeID uuid.UUID, startTime, endTime time.Time, distance float64, orderCount int) error {
	planningData := map[string]interface{}{
		"planned_start_time": startTime,
		"planned_end_time":   endTime,
		"planned_distance":   distance,
		"planned_orders":     orderCount,
		"planned_at":         time.Now(),
	}

	planningJSON, err := json.Marshal(planningData)
	if err != nil {
		return err
	}

	query := `
		UPDATE delivery_routes 
		SET start_time = $2, end_time = $3, estimated_distance = $4, 
			estimated_orders = $5, planning_data = $6, updated_at = $7
		WHERE id = $1`

	_, err = r.db.ExecContext(ctx, query, routeID, startTime, endTime, distance, orderCount, planningJSON, time.Now())
	return err
}

// UpdatePlanning updates planning data for a route
func (r *routeRepository) UpdatePlanning(ctx context.Context, routeID uuid.UUID, updates map[string]interface{}) error {
	planningJSON, err := json.Marshal(updates)
	if err != nil {
		return err
	}

	query := `
		UPDATE delivery_routes 
		SET planning_data = $2, updated_at = $3
		WHERE id = $1`

	_, err = r.db.ExecContext(ctx, query, routeID, planningJSON, time.Now())
	return err
}

// GetRouteMetrics retrieves route performance metrics
func (r *routeRepository) GetRouteMetrics(ctx context.Context, routeID uuid.UUID) (*repository.RouteMetrics, error) {
	query := `
		SELECT 
			r.id,
			r.route_name,
			COUNT(*) as total_routes,
			COUNT(CASE WHEN r.status = 'completed' THEN 1 END) as completed_routes,
			COUNT(CASE WHEN r.status = 'cancelled' THEN 1 END) as cancelled_routes,
			COALESCE(AVG(r.actual_distance), 0) as average_distance,
			COALESCE(AVG(EXTRACT(EPOCH FROM (r.end_time - r.start_time))/3600), 0) as average_duration,
			COALESCE(AVG(r.actual_orders_delivered), 0) as average_orders_per_route,
			COALESCE(AVG(CASE WHEN r.estimated_distance > 0 THEN (r.actual_distance / r.estimated_distance) * 100 END), 0) as efficiency_rate,
			MIN(r.route_date) as period_start,
			MAX(r.route_date) as period_end
		FROM delivery_routes r
		WHERE r.id = $1
		GROUP BY r.id, r.route_name`

	var metrics repository.RouteMetrics
	err := r.db.QueryRowxContext(ctx, query, routeID).Scan(
		&metrics.RouteID, &metrics.RouteName, &metrics.TotalRoutes,
		&metrics.CompletedRoutes, &metrics.CancelledRoutes, &metrics.AverageDistance,
		&metrics.AverageDuration, &metrics.AverageOrdersPerRoute, &metrics.EfficiencyRate,
		&metrics.PeriodStart, &metrics.PeriodEnd,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("route metrics not found")
		}
		return nil, err
	}

	// Calculate additional metrics
	if metrics.TotalRoutes > 0 {
		metrics.OnTimeRate = (float64(metrics.CompletedRoutes) / float64(metrics.TotalRoutes)) * 100
	}

	return &metrics, nil
}

// GetRoutesEfficiency retrieves efficiency metrics for routes in a date range
func (r *routeRepository) GetRoutesEfficiency(ctx context.Context, startDate, endDate time.Time) (map[uuid.UUID]float64, error) {
	query := `
		SELECT id, 
			   CASE WHEN estimated_distance > 0 THEN 
				   (actual_distance / estimated_distance) * 100 
			   ELSE 0 END as efficiency
		FROM delivery_routes 
		WHERE route_date BETWEEN $1 AND $2 
			  AND status = 'completed'
			  AND actual_distance IS NOT NULL 
			  AND estimated_distance > 0`

	rows, err := r.db.QueryxContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	efficiency := make(map[uuid.UUID]float64)
	for rows.Next() {
		var id uuid.UUID
		var eff float64
		if err := rows.Scan(&id, &eff); err != nil {
			return nil, err
		}
		efficiency[id] = eff
	}

	return efficiency, nil
}

// GetAverageRouteTime retrieves average route completion time
func (r *routeRepository) GetAverageRouteTime(ctx context.Context, startDate, endDate time.Time) (float64, error) {
	query := `
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (end_time - start_time))/3600), 0) as avg_hours
		FROM delivery_routes 
		WHERE route_date BETWEEN $1 AND $2 
			  AND status = 'completed'
			  AND start_time IS NOT NULL 
			  AND end_time IS NOT NULL`

	var avgHours float64
	err := r.db.QueryRowxContext(ctx, query, startDate, endDate).Scan(&avgHours)
	return avgHours, err
}

// GetRouteUtilization retrieves route utilization for a specific date
func (r *routeRepository) GetRouteUtilization(ctx context.Context, date time.Time) (float64, error) {
	query := `
		SELECT 
			CASE WHEN COUNT(*) > 0 THEN 
				(COUNT(CASE WHEN vehicle_id IS NOT NULL THEN 1 END)::FLOAT / COUNT(*)::FLOAT) * 100 
			ELSE 0 END as utilization
		FROM delivery_routes 
		WHERE route_date = $1`

	var utilization float64
	err := r.db.QueryRowxContext(ctx, query, date).Scan(&utilization)
	return utilization, err
}

// SearchRoutes searches routes based on filters
func (r *routeRepository) SearchRoutes(ctx context.Context, filters *repository.RouteQueryFilters) ([]*entity.DeliveryRoute, error) {
	query := `
		SELECT id, route_name, route_date, status, vehicle_id, driver_id,
			   start_time, end_time, estimated_distance, estimated_orders,
			   actual_distance, actual_orders_delivered, optimization_data,
			   planning_data, created_at, updated_at
		FROM delivery_routes WHERE 1=1`
	
	args := []interface{}{}
	argIndex := 1

	if filters.RouteDate != nil {
		query += fmt.Sprintf(" AND route_date = $%d", argIndex)
		args = append(args, *filters.RouteDate)
		argIndex++
	}

	if filters.StartDate != nil && filters.EndDate != nil {
		query += fmt.Sprintf(" AND route_date BETWEEN $%d AND $%d", argIndex, argIndex+1)
		args = append(args, *filters.StartDate, *filters.EndDate)
		argIndex += 2
	}

	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *filters.Status)
		argIndex++
	}

	if filters.VehicleID != nil {
		query += fmt.Sprintf(" AND vehicle_id = $%d", argIndex)
		args = append(args, *filters.VehicleID)
		argIndex++
	}

	if filters.DriverID != nil {
		query += fmt.Sprintf(" AND driver_id = $%d", argIndex)
		args = append(args, *filters.DriverID)
		argIndex++
	}

	if filters.RouteName != nil {
		query += fmt.Sprintf(" AND route_name ILIKE $%d", argIndex)
		args = append(args, "%"+*filters.RouteName+"%")
		argIndex++
	}

	if filters.MinDistance != nil {
		query += fmt.Sprintf(" AND estimated_distance >= $%d", argIndex)
		args = append(args, *filters.MinDistance)
		argIndex++
	}

	if filters.MaxDistance != nil {
		query += fmt.Sprintf(" AND estimated_distance <= $%d", argIndex)
		args = append(args, *filters.MaxDistance)
		argIndex++
	}

	if filters.MinOrders != nil {
		query += fmt.Sprintf(" AND estimated_orders >= $%d", argIndex)
		args = append(args, *filters.MinOrders)
		argIndex++
	}

	if filters.MaxOrders != nil {
		query += fmt.Sprintf(" AND estimated_orders <= $%d", argIndex)
		args = append(args, *filters.MaxOrders)
		argIndex++
	}

	query += " ORDER BY route_date DESC, route_name"

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filters.Offset)
	}

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRoutes(rows)
}

// GetRoutesByPattern retrieves routes matching a name pattern
func (r *routeRepository) GetRoutesByPattern(ctx context.Context, namePattern string) ([]*entity.DeliveryRoute, error) {
	query := `
		SELECT id, route_name, route_date, status, vehicle_id, driver_id,
			   start_time, end_time, estimated_distance, estimated_orders,
			   actual_distance, actual_orders_delivered, optimization_data,
			   planning_data, created_at, updated_at
		FROM delivery_routes 
		WHERE route_name ILIKE $1
		ORDER BY route_date DESC, route_name`

	rows, err := r.db.QueryxContext(ctx, query, "%"+namePattern+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRoutes(rows)
}

// GetUnassignedRoutes retrieves routes without vehicle assignment for a date
func (r *routeRepository) GetUnassignedRoutes(ctx context.Context, date time.Time) ([]*entity.DeliveryRoute, error) {
	query := `
		SELECT id, route_name, route_date, status, vehicle_id, driver_id,
			   start_time, end_time, estimated_distance, estimated_orders,
			   actual_distance, actual_orders_delivered, optimization_data,
			   planning_data, created_at, updated_at
		FROM delivery_routes 
		WHERE route_date = $1 AND vehicle_id IS NULL
		ORDER BY route_name`

	rows, err := r.db.QueryxContext(ctx, query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanRoutes(rows)
}

// CreateMultipleRoutes creates multiple routes in a single transaction
func (r *routeRepository) CreateMultipleRoutes(ctx context.Context, routes []*entity.DeliveryRoute) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO delivery_routes (
			id, route_name, route_date, status, vehicle_id, driver_id,
			start_time, end_time, estimated_distance, estimated_orders,
			actual_distance, actual_orders_delivered, optimization_data,
			planning_data, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)`

	for _, route := range routes {
		optimizationJSON, _ := json.Marshal(route.RouteOptimizationData)
		planningJSON, _ := json.Marshal(map[string]interface{}{
			"planned_start_time": route.PlannedStartTime,
			"planned_end_time":   route.PlannedEndTime,
			"total_planned_distance": route.TotalPlannedDistance,
			"total_planned_orders": route.TotalPlannedOrders,
		})

		_, err = tx.ExecContext(ctx, query,
			route.ID, route.RouteName, route.RouteDate, route.Status,
			route.AssignedVehicleID, route.AssignedDriverID, route.PlannedStartTime, route.PlannedEndTime,
			route.TotalPlannedDistance, route.TotalPlannedOrders, route.ActualDistance,
			route.ActualOrdersDelivered, optimizationJSON, planningJSON,
			route.CreatedAt, route.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// UpdateMultipleRouteStatuses updates status for multiple routes
func (r *routeRepository) UpdateMultipleRouteStatuses(ctx context.Context, routeIDs []uuid.UUID, status entity.RouteStatus) error {
	query := `
		UPDATE delivery_routes 
		SET status = $1, updated_at = $2
		WHERE id = ANY($3)`

	_, err := r.db.ExecContext(ctx, query, status, time.Now(), pq.Array(routeIDs))
	return err
}

// DeleteRoutesByDate deletes all routes for a specific date
func (r *routeRepository) DeleteRoutesByDate(ctx context.Context, date time.Time) error {
	query := `DELETE FROM delivery_routes WHERE route_date = $1`
	_, err := r.db.ExecContext(ctx, query, date)
	return err
}

// scanRoutes is a helper method to scan multiple routes from query results
func (r *routeRepository) scanRoutes(rows *sqlx.Rows) ([]*entity.DeliveryRoute, error) {
	var routes []*entity.DeliveryRoute

	for rows.Next() {
		var route entity.DeliveryRoute
		var optimizationJSON, planningJSON []byte

		err := rows.Scan(
			&route.ID, &route.RouteName, &route.RouteDate, &route.Status,
			&route.AssignedVehicleID, &route.AssignedDriverID, &route.PlannedStartTime, &route.PlannedEndTime,
			&route.TotalPlannedDistance, &route.TotalPlannedOrders, &route.ActualDistance,
			&route.ActualOrdersDelivered, &optimizationJSON, &planningJSON,
			&route.CreatedAt, &route.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if optimizationJSON != nil {
			json.Unmarshal(optimizationJSON, &route.RouteOptimizationData)
		}

		routes = append(routes, &route)
	}

	return routes, nil
}
