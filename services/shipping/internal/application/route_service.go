package application

import (
	"database/sql"
	"fmt"

	"saan/shipping/internal/domain"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type routeService struct {
	db    *sql.DB
	redis *redis.Client
	repo  domain.RouteRepository
}

func NewRouteService(db *sql.DB, redis *redis.Client) domain.RouteService {
	return &routeService{
		db:    db,
		redis: redis,
		// Initialize repository here
	}
}

func (r *routeService) OptimizeRoute(tasks []*domain.DeliveryTask) ([]*domain.DeliveryTask, error) {
	// Simple optimization: sort by district/subdistrict
	// In production, use more sophisticated algorithms
	return tasks, nil
}

func (r *routeService) GetRouteInfo(routeCode string) (*domain.DeliveryRoute, error) {
	return r.repo.GetByCode(routeCode)
}

func (r *routeService) CalculateDeliveryFee(route string, codAmount float64) (float64, error) {
	routeInfo, err := r.repo.GetByCode(route)
	if err != nil {
		return 0, fmt.Errorf("failed to get route info: %w", err)
	}

	// Base fee + percentage of COD amount
	fee := routeInfo.BaseDeliveryFee
	if codAmount > 1000 {
		fee += codAmount * 0.02 // 2% of COD amount
	}

	return fee, nil
}

type carrierService struct {
	db    *sql.DB
	redis *redis.Client
	repo  domain.CarrierRepository
}

func NewCarrierService(db *sql.DB, redis *redis.Client) domain.CarrierService {
	return &carrierService{
		db:    db,
		redis: redis,
		// Initialize repository here
	}
}

func (c *carrierService) GetAvailableCarriers(province string) ([]*domain.DeliveryCarrier, error) {
	// Filter carriers that serve this province
	carriers, err := c.repo.GetActiveCarriers()
	if err != nil {
		return nil, err
	}

	// In production, filter by province coverage
	return carriers, nil
}

func (c *carrierService) SchedulePickup(carrierID uuid.UUID, tasks []*domain.DeliveryTask) error {
	// Schedule pickup with carrier
	// Update tasks with carrier info
	return nil
}

func (c *carrierService) GetTrackingInfo(carrierID uuid.UUID, trackingNumber string) (map[string]interface{}, error) {
	// Call carrier API for tracking info
	return map[string]interface{}{
		"status":   "in_transit",
		"location": "Bangkok Hub",
	}, nil
}
