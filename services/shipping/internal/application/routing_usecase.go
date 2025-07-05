package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"shipping/internal/domain/entity"
	"shipping/internal/domain/repository"
)

// RoutingUseCase handles route planning and optimization operations
type RoutingUseCase struct {
	routeRepo    repository.RouteRepository
	deliveryRepo repository.DeliveryRepository
	vehicleRepo  repository.VehicleRepository
	eventPub     EventPublisher
	cache        Cache
}

// NewRoutingUseCase creates a new routing use case
func NewRoutingUseCase(
	routeRepo repository.RouteRepository,
	deliveryRepo repository.DeliveryRepository,
	vehicleRepo repository.VehicleRepository,
	eventPub EventPublisher,
	cache Cache,
) *RoutingUseCase {
	return &RoutingUseCase{
		routeRepo:    routeRepo,
		deliveryRepo: deliveryRepo,
		vehicleRepo:  vehicleRepo,
		eventPub:     eventPub,
		cache:        cache,
	}
}

// CreateRouteRequest represents a request to create a new route
type CreateRouteRequest struct {
	RouteName string    `json:"route_name" validate:"required"`
	RouteDate time.Time `json:"route_date" validate:"required"`
	CreatedBy string    `json:"created_by" validate:"required"`
}

// UpdateRouteRequest represents a request to update a route
type UpdateRouteRequest struct {
	RouteID          uuid.UUID   `json:"route_id" validate:"required"`
	RouteName        *string     `json:"route_name,omitempty"`
	PlannedStartTime *time.Time  `json:"planned_start_time,omitempty"`
	PlannedEndTime   *time.Time  `json:"planned_end_time,omitempty"`
	UpdatedBy        string      `json:"updated_by" validate:"required"`
}

// AssignVehicleRequest represents a request to assign a vehicle to a route
type AssignVehicleRequest struct {
	RouteID    uuid.UUID  `json:"route_id" validate:"required"`
	VehicleID  uuid.UUID  `json:"vehicle_id" validate:"required"`
	DriverID   *uuid.UUID `json:"driver_id,omitempty"`
	AssignedBy string     `json:"assigned_by" validate:"required"`
}

// RouteOptimizationRequest represents a request to optimize a route
type RouteOptimizationRequest struct {
	RouteID     uuid.UUID `json:"route_id" validate:"required"`
	Algorithm   string    `json:"algorithm,omitempty"`
	OptimizedBy string    `json:"optimized_by" validate:"required"`
}

// CreateRoute creates a new delivery route
func (uc *RoutingUseCase) CreateRoute(ctx context.Context, req CreateRouteRequest) (*entity.DeliveryRoute, error) {
	// Create new route entity
	route, err := entity.NewDeliveryRoute(req.RouteName, req.RouteDate)
	if err != nil {
		return nil, fmt.Errorf("failed to create route entity: %w", err)
	}

	// Save to repository
	if err := uc.routeRepo.Create(ctx, route); err != nil {
		return nil, fmt.Errorf("failed to save route: %w", err)
	}

	// Publish event
	uc.eventPub.Publish(ctx, "route.created", map[string]interface{}{
		"route_id":   route.ID.String(),
		"route_name": route.RouteName,
		"route_date": route.RouteDate,
		"created_by": req.CreatedBy,
		"created_at": route.CreatedAt,
	})

	return route, nil
}

// GetRoute retrieves a route by ID
func (uc *RoutingUseCase) GetRoute(ctx context.Context, routeID uuid.UUID) (*entity.DeliveryRoute, error) {
	route, err := uc.routeRepo.GetByID(ctx, routeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get route: %w", err)
	}
	
	return route, nil
}

// GetRoutesByVehicle retrieves routes assigned to a specific vehicle
func (uc *RoutingUseCase) GetRoutesByVehicle(ctx context.Context, vehicleID uuid.UUID, date time.Time) ([]*entity.DeliveryRoute, error) {
	routes, err := uc.routeRepo.GetByVehicle(ctx, vehicleID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get routes by vehicle: %w", err)
	}
	
	return routes, nil
}

// GetRoutesByDriver retrieves routes assigned to a specific driver
func (uc *RoutingUseCase) GetRoutesByDriver(ctx context.Context, driverID uuid.UUID, date time.Time) ([]*entity.DeliveryRoute, error) {
	routes, err := uc.routeRepo.GetByDriver(ctx, driverID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get routes by driver: %w", err)
	}
	
	return routes, nil
}

// GetActiveRoutes retrieves all active routes
func (uc *RoutingUseCase) GetActiveRoutes(ctx context.Context) ([]*entity.DeliveryRoute, error) {
	routes, err := uc.routeRepo.GetActiveRoutes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active routes: %w", err)
	}
	
	return routes, nil
}

// GetRoutesByStatus retrieves routes by status
func (uc *RoutingUseCase) GetRoutesByStatus(ctx context.Context, status entity.RouteStatus) ([]*entity.DeliveryRoute, error) {
	routes, err := uc.routeRepo.GetByStatus(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get routes by status: %w", err)
	}
	
	return routes, nil
}

// UpdateRoute updates an existing route
func (uc *RoutingUseCase) UpdateRoute(ctx context.Context, req UpdateRouteRequest) (*entity.DeliveryRoute, error) {
	// Get existing route
	route, err := uc.routeRepo.GetByID(ctx, req.RouteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get route: %w", err)
	}

	// Update fields if provided
	if req.RouteName != nil {
		route.RouteName = *req.RouteName
	}

	if req.PlannedStartTime != nil && req.PlannedEndTime != nil {
		if err := route.SetPlanning(*req.PlannedStartTime, *req.PlannedEndTime, route.TotalPlannedDistance, route.TotalPlannedOrders); err != nil {
			return nil, fmt.Errorf("failed to update route planning: %w", err)
		}
	}

	route.UpdatedAt = time.Now()

	// Save updated route
	if err := uc.routeRepo.Update(ctx, route); err != nil {
		return nil, fmt.Errorf("failed to update route: %w", err)
	}

	// Publish event
	uc.eventPub.Publish(ctx, "route.updated", map[string]interface{}{
		"route_id":   route.ID.String(),
		"route_name": route.RouteName,
		"updated_by": req.UpdatedBy,
		"updated_at": route.UpdatedAt,
	})

	return route, nil
}

// AssignVehicleToRoute assigns a vehicle and optionally a driver to a route
func (uc *RoutingUseCase) AssignVehicleToRoute(ctx context.Context, req AssignVehicleRequest) error {
	// Get route
	route, err := uc.routeRepo.GetByID(ctx, req.RouteID)
	if err != nil {
		return fmt.Errorf("failed to get route: %w", err)
	}

	// Assign vehicle to route
	if err := route.AssignVehicle(req.VehicleID, req.DriverID); err != nil {
		return fmt.Errorf("failed to assign vehicle to route: %w", err)
	}

	// Update route in repository
	if err := uc.routeRepo.Update(ctx, route); err != nil {
		return fmt.Errorf("failed to update route: %w", err)
	}

	// Update vehicle status
	if err := uc.vehicleRepo.UpdateStatus(ctx, req.VehicleID, entity.VehicleStatusOnRoute); err != nil {
		return fmt.Errorf("failed to update vehicle status: %w", err)
	}

	// Publish event
	uc.eventPub.Publish(ctx, "route.vehicle_assigned", map[string]interface{}{
		"route_id":    route.ID.String(),
		"vehicle_id":  req.VehicleID.String(),
		"driver_id":   req.DriverID,
		"assigned_by": req.AssignedBy,
		"assigned_at": time.Now(),
	})

	return nil
}

// StartRoute starts route execution
func (uc *RoutingUseCase) StartRoute(ctx context.Context, routeID uuid.UUID, startedBy string) error {
	// Get route
	route, err := uc.routeRepo.GetByID(ctx, routeID)
	if err != nil {
		return fmt.Errorf("failed to get route: %w", err)
	}

	// Validate route can be started
	if route.AssignedVehicleID == nil {
		return fmt.Errorf("route must have an assigned vehicle before starting")
	}

	// Start the route
	if err := route.StartRoute(); err != nil {
		return fmt.Errorf("failed to start route: %w", err)
	}

	// Update route in repository
	if err := uc.routeRepo.Update(ctx, route); err != nil {
		return fmt.Errorf("failed to update route: %w", err)
	}

	// Update vehicle status
	if err := uc.vehicleRepo.UpdateStatus(ctx, *route.AssignedVehicleID, entity.VehicleStatusOnRoute); err != nil {
		return fmt.Errorf("failed to update vehicle status: %w", err)
	}

	// Publish event
	uc.eventPub.Publish(ctx, "route.started", map[string]interface{}{
		"route_id":    route.ID.String(),
		"vehicle_id":  route.AssignedVehicleID.String(),
		"driver_id":   route.AssignedDriverID,
		"started_by":  startedBy,
		"started_at":  route.ActualStartTime,
	})

	return nil
}

// CompleteRoute completes route execution
func (uc *RoutingUseCase) CompleteRoute(ctx context.Context, routeID uuid.UUID, actualDistance float64, actualOrdersDelivered int, completedBy string) error {
	// Get route
	route, err := uc.routeRepo.GetByID(ctx, routeID)
	if err != nil {
		return fmt.Errorf("failed to get route: %w", err)
	}

	// Complete the route
	if err := route.CompleteRoute(actualDistance, actualOrdersDelivered); err != nil {
		return fmt.Errorf("failed to complete route: %w", err)
	}

	// Update route in repository
	if err := uc.routeRepo.Update(ctx, route); err != nil {
		return fmt.Errorf("failed to update route: %w", err)
	}

	// Update vehicle status if assigned
	if route.AssignedVehicleID != nil {
		if err := uc.vehicleRepo.UpdateStatus(ctx, *route.AssignedVehicleID, entity.VehicleStatusActive); err != nil {
			return fmt.Errorf("failed to update vehicle status: %w", err)
		}
	}

	// Publish event
	uc.eventPub.Publish(ctx, "route.completed", map[string]interface{}{
		"route_id":                route.ID.String(),
		"completed_by":            completedBy,
		"completed_at":            route.ActualEndTime,
		"actual_distance":         route.ActualDistance,
		"actual_orders_delivered": route.ActualOrdersDelivered,
		"efficiency":              route.GetEfficiency(),
		"duration":                route.GetDuration().String(),
	})

	return nil
}

// CancelRoute cancels a route
func (uc *RoutingUseCase) CancelRoute(ctx context.Context, routeID uuid.UUID, reason, cancelledBy string) error {
	// Get route
	route, err := uc.routeRepo.GetByID(ctx, routeID)
	if err != nil {
		return fmt.Errorf("failed to get route: %w", err)
	}

	// Cancel the route
	if err := route.CancelRoute(); err != nil {
		return fmt.Errorf("failed to cancel route: %w", err)
	}

	// Update route in repository
	if err := uc.routeRepo.Update(ctx, route); err != nil {
		return fmt.Errorf("failed to update route: %w", err)
	}

	// Update vehicle status if assigned
	if route.AssignedVehicleID != nil {
		if err := uc.vehicleRepo.UpdateStatus(ctx, *route.AssignedVehicleID, entity.VehicleStatusActive); err != nil {
			return fmt.Errorf("failed to update vehicle status: %w", err)
		}
	}

	// Get deliveries assigned to this route and unassign them
	deliveries, err := uc.deliveryRepo.GetByRouteID(ctx, routeID)
	if err != nil {
		return fmt.Errorf("failed to get route deliveries: %w", err)
	}

	// Update delivery statuses back to pending
	var deliveryIDs []uuid.UUID
	for _, delivery := range deliveries {
		deliveryIDs = append(deliveryIDs, delivery.ID)
	}

	if len(deliveryIDs) > 0 {
		if err := uc.deliveryRepo.UpdateMultipleStatuses(ctx, deliveryIDs, entity.DeliveryStatusPending); err != nil {
			return fmt.Errorf("failed to update delivery statuses: %w", err)
		}
	}

	// Publish event
	uc.eventPub.Publish(ctx, "route.cancelled", map[string]interface{}{
		"route_id":     route.ID.String(),
		"reason":       reason,
		"cancelled_by": cancelledBy,
		"cancelled_at": route.UpdatedAt,
	})

	return nil
}

// OptimizeRoute optimizes the route using specified algorithm
func (uc *RoutingUseCase) OptimizeRoute(ctx context.Context, req RouteOptimizationRequest) (*entity.DeliveryRoute, error) {
	// Get route
	route, err := uc.routeRepo.GetByID(ctx, req.RouteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get route: %w", err)
	}

	// Get deliveries for this route
	deliveries, err := uc.deliveryRepo.GetByRouteID(ctx, req.RouteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get route deliveries: %w", err)
	}

	if len(deliveries) == 0 {
		return route, nil // No deliveries to optimize
	}

	// Create optimization data
	optimizationData := map[string]interface{}{
		"algorithm":      req.Algorithm,
		"optimized_by":   req.OptimizedBy,
		"optimized_at":   time.Now(),
		"total_stops":    len(deliveries),
		"delivery_count": len(deliveries),
	}

	// Set optimization data
	route.SetOptimizationData(optimizationData)

	// Update route in repository
	if err := uc.routeRepo.Update(ctx, route); err != nil {
		return nil, fmt.Errorf("failed to update route: %w", err)
	}

	// Publish event
	uc.eventPub.Publish(ctx, "route.optimized", map[string]interface{}{
		"route_id":     route.ID.String(),
		"algorithm":    req.Algorithm,
		"optimized_by": req.OptimizedBy,
		"optimized_at": time.Now(),
		"total_stops":  len(deliveries),
	})

	return route, nil
}

// GetRouteMetrics retrieves metrics for a specific route
func (uc *RoutingUseCase) GetRouteMetrics(ctx context.Context, routeID uuid.UUID) (*repository.RouteMetrics, error) {
	metrics, err := uc.routeRepo.GetRouteMetrics(ctx, routeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get route metrics: %w", err)
	}

	return metrics, nil
}

// GetPlannedRoutes retrieves all planned routes for a specific date
func (uc *RoutingUseCase) GetPlannedRoutes(ctx context.Context, date time.Time) ([]*entity.DeliveryRoute, error) {
	routes, err := uc.routeRepo.GetPlannedRoutes(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get planned routes: %w", err)
	}

	return routes, nil
}

// GetInProgressRoutes retrieves all routes currently in progress
func (uc *RoutingUseCase) GetInProgressRoutes(ctx context.Context) ([]*entity.DeliveryRoute, error) {
	routes, err := uc.routeRepo.GetInProgressRoutes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get in-progress routes: %w", err)
	}

	return routes, nil
}

// DeleteRoute deletes a route (only if not started)
func (uc *RoutingUseCase) DeleteRoute(ctx context.Context, routeID uuid.UUID, deletedBy string) error {
	// Get route
	route, err := uc.routeRepo.GetByID(ctx, routeID)
	if err != nil {
		return fmt.Errorf("failed to get route: %w", err)
	}

	// Only allow deletion of planned routes
	if route.Status != entity.RouteStatusPlanned {
		return fmt.Errorf("only planned routes can be deleted")
	}

	// Check if route has assigned deliveries
	deliveries, err := uc.deliveryRepo.GetByRouteID(ctx, routeID)
	if err != nil {
		return fmt.Errorf("failed to check route deliveries: %w", err)
	}

	if len(deliveries) > 0 {
		return fmt.Errorf("cannot delete route with assigned deliveries")
	}

	// Delete route
	if err := uc.routeRepo.Delete(ctx, routeID); err != nil {
		return fmt.Errorf("failed to delete route: %w", err)
	}

	// Publish event
	uc.eventPub.Publish(ctx, "route.deleted", map[string]interface{}{
		"route_id":   routeID.String(),
		"deleted_by": deletedBy,
		"deleted_at": time.Now(),
	})

	return nil
}
