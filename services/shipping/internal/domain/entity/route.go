package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// DeliveryRoute represents an optimized delivery route for vehicles
type DeliveryRoute struct {
	ID                    uuid.UUID  `json:"id"`
	RouteName             string     `json:"route_name"`
	RouteDate             time.Time  `json:"route_date"`
	AssignedVehicleID     *uuid.UUID `json:"assigned_vehicle_id"`
	AssignedDriverID      *uuid.UUID `json:"assigned_driver_id"`
	
	// Planning
	PlannedStartTime      *time.Time `json:"planned_start_time"`
	PlannedEndTime        *time.Time `json:"planned_end_time"`
	TotalPlannedDistance  float64    `json:"total_planned_distance_km"`
	TotalPlannedOrders    int        `json:"total_planned_orders"`
	
	// Status
	Status                RouteStatus `json:"status"`
	ActualStartTime       *time.Time  `json:"actual_start_time"`
	ActualEndTime         *time.Time  `json:"actual_end_time"`
	ActualDistance        float64     `json:"actual_distance_km"`
	ActualOrdersDelivered int         `json:"actual_orders_delivered"`
	
	// Optimization
	RouteOptimizationData map[string]interface{} `json:"route_optimization_data"`
	
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// RouteStatus represents the status of a delivery route
type RouteStatus string

const (
	RouteStatusPlanned     RouteStatus = "planned"
	RouteStatusInProgress  RouteStatus = "in_progress"
	RouteStatusCompleted   RouteStatus = "completed"
	RouteStatusCancelled   RouteStatus = "cancelled"
)

// Domain errors
var (
	ErrRouteInvalidName        = errors.New("route name cannot be empty")
	ErrRouteInvalidDate        = errors.New("route date cannot be in the past")
	ErrRouteInvalidTimeSlot    = errors.New("planned end time must be after start time")
	ErrRouteInvalidDistance    = errors.New("planned distance must be positive")
	ErrRouteInvalidOrderCount  = errors.New("planned orders count must be positive")
	ErrRouteAlreadyStarted     = errors.New("route has already been started")
	ErrRouteNotStarted         = errors.New("route has not been started yet")
	ErrRouteAlreadyCompleted   = errors.New("route is already completed")
	ErrRouteInvalidStatus      = errors.New("invalid route status")
)

// NewDeliveryRoute creates a new delivery route with validation
func NewDeliveryRoute(routeName string, routeDate time.Time) (*DeliveryRoute, error) {
	if err := validateRouteName(routeName); err != nil {
		return nil, err
	}
	
	if err := validateRouteDate(routeDate); err != nil {
		return nil, err
	}
	
	now := time.Now()
	return &DeliveryRoute{
		ID:                    uuid.New(),
		RouteName:             routeName,
		RouteDate:             routeDate,
		Status:                RouteStatusPlanned,
		RouteOptimizationData: make(map[string]interface{}),
		CreatedAt:             now,
		UpdatedAt:             now,
	}, nil
}

// AssignVehicle assigns a vehicle to the route
func (r *DeliveryRoute) AssignVehicle(vehicleID uuid.UUID, driverID *uuid.UUID) error {
	if r.Status != RouteStatusPlanned {
		return ErrRouteAlreadyStarted
	}
	
	r.AssignedVehicleID = &vehicleID
	r.AssignedDriverID = driverID
	r.UpdatedAt = time.Now()
	
	return nil
}

// SetPlanning sets the route planning details
func (r *DeliveryRoute) SetPlanning(startTime, endTime time.Time, distance float64, orderCount int) error {
	if err := validateTimeSlot(startTime, endTime); err != nil {
		return err
	}
	
	if distance <= 0 {
		return ErrRouteInvalidDistance
	}
	
	if orderCount <= 0 {
		return ErrRouteInvalidOrderCount
	}
	
	r.PlannedStartTime = &startTime
	r.PlannedEndTime = &endTime
	r.TotalPlannedDistance = distance
	r.TotalPlannedOrders = orderCount
	r.UpdatedAt = time.Now()
	
	return nil
}

// StartRoute starts the route execution
func (r *DeliveryRoute) StartRoute() error {
	if r.Status != RouteStatusPlanned {
		return ErrRouteAlreadyStarted
	}
	
	if r.AssignedVehicleID == nil {
		return errors.New("vehicle must be assigned before starting route")
	}
	
	now := time.Now()
	r.Status = RouteStatusInProgress
	r.ActualStartTime = &now
	r.UpdatedAt = now
	
	return nil
}

// CompleteRoute completes the route execution
func (r *DeliveryRoute) CompleteRoute(actualDistance float64, actualOrdersDelivered int) error {
	if r.Status != RouteStatusInProgress {
		return ErrRouteNotStarted
	}
	
	if actualDistance < 0 {
		return ErrRouteInvalidDistance
	}
	
	if actualOrdersDelivered < 0 {
		return ErrRouteInvalidOrderCount
	}
	
	now := time.Now()
	r.Status = RouteStatusCompleted
	r.ActualEndTime = &now
	r.ActualDistance = actualDistance
	r.ActualOrdersDelivered = actualOrdersDelivered
	r.UpdatedAt = now
	
	return nil
}

// CancelRoute cancels the route
func (r *DeliveryRoute) CancelRoute() error {
	if r.Status == RouteStatusCompleted {
		return ErrRouteAlreadyCompleted
	}
	
	r.Status = RouteStatusCancelled
	r.UpdatedAt = time.Now()
	
	return nil
}

// SetOptimizationData sets the route optimization data
func (r *DeliveryRoute) SetOptimizationData(data map[string]interface{}) {
	r.RouteOptimizationData = data
	r.UpdatedAt = time.Now()
}

// IsActive returns true if the route is active (planned or in progress)
func (r *DeliveryRoute) IsActive() bool {
	return r.Status == RouteStatusPlanned || r.Status == RouteStatusInProgress
}

// GetEfficiency calculates route efficiency (actual vs planned)
func (r *DeliveryRoute) GetEfficiency() float64 {
	if r.Status != RouteStatusCompleted || r.TotalPlannedOrders == 0 {
		return 0.0
	}
	
	return float64(r.ActualOrdersDelivered) / float64(r.TotalPlannedOrders) * 100.0
}

// GetDuration calculates the route duration
func (r *DeliveryRoute) GetDuration() time.Duration {
	if r.ActualStartTime == nil {
		return 0
	}
	
	endTime := r.ActualEndTime
	if endTime == nil {
		endTime = &time.Time{}
		*endTime = time.Now()
	}
	
	return endTime.Sub(*r.ActualStartTime)
}

// Validation functions
func validateRouteName(name string) error {
	if name == "" {
		return ErrRouteInvalidName
	}
	return nil
}

func validateRouteDate(date time.Time) error {
	if date.Before(time.Now().Truncate(24 * time.Hour)) {
		return ErrRouteInvalidDate
	}
	return nil
}

func validateTimeSlot(start, end time.Time) error {
	if end.Before(start) || end.Equal(start) {
		return ErrRouteInvalidTimeSlot
	}
	return nil
}
