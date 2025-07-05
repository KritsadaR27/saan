package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"shipping/internal/domain/entity"
	"shipping/internal/domain/repository"
)

// TrackingUseCase handles delivery tracking operations
type TrackingUseCase struct {
	deliveryRepo   repository.DeliveryRepository
	snapshotRepo   repository.SnapshotRepository
	eventPublisher EventPublisher
	cache          Cache
}

// NewTrackingUseCase creates a new tracking use case
func NewTrackingUseCase(
	deliveryRepo repository.DeliveryRepository,
	snapshotRepo repository.SnapshotRepository,
	eventPublisher EventPublisher,
	cache Cache,
) *TrackingUseCase {
	return &TrackingUseCase{
		deliveryRepo:   deliveryRepo,
		snapshotRepo:   snapshotRepo,
		eventPublisher: eventPublisher,
		cache:          cache,
	}
}

// Location represents geographic coordinates
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address,omitempty"`
}

// DeliveryLocation represents a delivery with location information
type DeliveryLocation struct {
	Delivery         *entity.DeliveryOrder `json:"delivery"`
	CurrentLocation  *Location             `json:"current_location,omitempty"`
	LastUpdated      time.Time             `json:"last_updated"`
	EstimatedArrival *time.Time            `json:"estimated_arrival,omitempty"`
}

// TrackingResponse represents the tracking information for a delivery
type TrackingResponse struct {
	DeliveryID       uuid.UUID             `json:"delivery_id"`
	TrackingNumber   string                `json:"tracking_number"`
	Status           entity.DeliveryStatus `json:"status"`
	CurrentLocation  *Location             `json:"current_location,omitempty"`
	LastUpdated      time.Time             `json:"last_updated"`
	EstimatedArrival *time.Time            `json:"estimated_arrival,omitempty"`
	Updates          []TrackingUpdate      `json:"updates"`
}

// TrackingUpdate represents a single tracking update
type TrackingUpdate struct {
	Timestamp   time.Time `json:"timestamp"`
	Status      string    `json:"status"`
	Location    *Location `json:"location,omitempty"`
	Description string    `json:"description"`
}

// LocationUpdateRequest represents a request to update delivery location
type LocationUpdateRequest struct {
	DeliveryID       uuid.UUID `json:"delivery_id" validate:"required"`
	CurrentLocation  Location  `json:"current_location" validate:"required"`
	EstimatedArrival *time.Time `json:"estimated_arrival,omitempty"`
	UpdatedBy        string    `json:"updated_by" validate:"required"`
}

// TrackDelivery retrieves tracking information for a delivery
func (uc *TrackingUseCase) TrackDelivery(ctx context.Context, trackingNumber string) (*TrackingResponse, error) {
	// Get delivery by tracking number
	delivery, err := uc.deliveryRepo.GetByTrackingNumber(ctx, trackingNumber)
	if err != nil {
		return nil, fmt.Errorf("delivery not found: %w", err)
	}

	// Get snapshots for tracking history
	snapshots, err := uc.snapshotRepo.GetByDeliveryID(ctx, delivery.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tracking snapshots: %w", err)
	}

	// Build tracking response
	response := &TrackingResponse{
		DeliveryID:     delivery.ID,
		TrackingNumber: trackingNumber,
		Status:         delivery.Status,
		LastUpdated:    delivery.UpdatedAt,
	}

	// Add tracking updates from snapshots
	for _, snapshot := range snapshots {
		if snapshot.SnapshotData != nil {
			if statusUpdate, ok := snapshot.SnapshotData["status"].(string); ok {
				update := TrackingUpdate{
					Timestamp:   snapshot.CreatedAt,
					Status:      statusUpdate,
					Description: fmt.Sprintf("Status updated to %s", statusUpdate),
				}

				// Add location if available
				if lat, latOk := snapshot.SnapshotData["latitude"].(float64); latOk {
					if lng, lngOk := snapshot.SnapshotData["longitude"].(float64); lngOk {
						update.Location = &Location{
							Latitude:  lat,
							Longitude: lng,
						}
						if addr, addrOk := snapshot.SnapshotData["address"].(string); addrOk {
							update.Location.Address = addr
						}
					}
				}

				response.Updates = append(response.Updates, update)
			}
		}
	}

	return response, nil
}

// TrackDeliveryByID retrieves tracking information for a delivery by ID
func (uc *TrackingUseCase) TrackDeliveryByID(ctx context.Context, deliveryID uuid.UUID) (*TrackingResponse, error) {
	// Get delivery
	delivery, err := uc.deliveryRepo.GetByID(ctx, deliveryID)
	if err != nil {
		return nil, fmt.Errorf("delivery not found: %w", err)
	}

	trackingNumber := ""
	if delivery.TrackingNumber != nil {
		trackingNumber = *delivery.TrackingNumber
	}

	return uc.TrackDelivery(ctx, trackingNumber)
}

// UpdateDeliveryLocation updates the current location of a delivery
func (uc *TrackingUseCase) UpdateDeliveryLocation(ctx context.Context, req LocationUpdateRequest) error {
	// Get delivery
	delivery, err := uc.deliveryRepo.GetByID(ctx, req.DeliveryID)
	if err != nil {
		return fmt.Errorf("delivery not found: %w", err)
	}

	// Create snapshot for location update
	snapshot := &entity.DeliverySnapshot{
		ID:                uuid.New(),
		DeliveryID:        delivery.ID,
		SnapshotType:      entity.SnapshotTypeInTransit,
		SnapshotData: map[string]interface{}{
			"latitude":           req.CurrentLocation.Latitude,
			"longitude":          req.CurrentLocation.Longitude,
			"address":            req.CurrentLocation.Address,
			"estimated_arrival":  req.EstimatedArrival,
			"updated_by":         req.UpdatedBy,
		},
		TriggeredBy:       req.UpdatedBy,
		TriggeredEvent:    "location_updated",
		CreatedAt:         time.Now(),
	}

	// Save snapshot
	if err := uc.snapshotRepo.Create(ctx, snapshot); err != nil {
		return fmt.Errorf("failed to save location snapshot: %w", err)
	}

	// Publish event
	uc.eventPublisher.Publish(ctx, "delivery.location_updated", map[string]interface{}{
		"delivery_id":        delivery.ID.String(),
		"tracking_number":    delivery.TrackingNumber,
		"current_location":   req.CurrentLocation,
		"estimated_arrival":  req.EstimatedArrival,
		"updated_by":         req.UpdatedBy,
		"updated_at":         time.Now(),
	})

	return nil
}

// GetActiveDeliveries retrieves all deliveries that are currently active (in transit)
func (uc *TrackingUseCase) GetActiveDeliveries(ctx context.Context) ([]*DeliveryLocation, error) {
	deliveries, err := uc.deliveryRepo.GetActiveDeliveries(ctx, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get active deliveries: %w", err)
	}

	var responses []*DeliveryLocation
	for _, delivery := range deliveries {
		// Get latest location snapshot for each delivery
		snapshots, err := uc.snapshotRepo.GetByDeliveryID(ctx, delivery.ID)
		if err != nil {
			continue // Skip if can't get snapshots
		}

		response := &DeliveryLocation{
			Delivery:    delivery,
			LastUpdated: delivery.UpdatedAt,
		}

		// Find latest location from snapshots
		for _, snapshot := range snapshots {
			if snapshot.TriggeredEvent == "location_updated" && snapshot.SnapshotData != nil {
				if lat, latOk := snapshot.SnapshotData["latitude"].(float64); latOk {
					if lng, lngOk := snapshot.SnapshotData["longitude"].(float64); lngOk {
						response.CurrentLocation = &Location{
							Latitude:  lat,
							Longitude: lng,
						}
						if addr, addrOk := snapshot.SnapshotData["address"].(string); addrOk {
							response.CurrentLocation.Address = addr
						}
						response.LastUpdated = snapshot.CreatedAt
						
						// Set estimated arrival if available
						if arrivalStr, ok := snapshot.SnapshotData["estimated_arrival"].(string); ok {
							if arrival, parseErr := time.Parse(time.RFC3339, arrivalStr); parseErr == nil {
								response.EstimatedArrival = &arrival
							}
						}
						break // Use the latest snapshot
					}
				}
			}
		}

		responses = append(responses, response)
	}

	return responses, nil
}

// GetDelayedDeliveries retrieves deliveries that are delayed
func (uc *TrackingUseCase) GetDelayedDeliveries(ctx context.Context) ([]*TrackingResponse, error) {
	// Get deliveries where estimated delivery time has passed but status is not delivered
	deliveries, err := uc.deliveryRepo.GetOverdueDeliveries(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get delayed deliveries: %w", err)
	}

	var responses []*TrackingResponse
	for _, delivery := range deliveries {
		trackingNum := ""
		if delivery.TrackingNumber != nil {
			trackingNum = *delivery.TrackingNumber
		}

		response := &TrackingResponse{
			DeliveryID:     delivery.ID,
			TrackingNumber: trackingNum,
			Status:         delivery.Status,
			LastUpdated:    delivery.UpdatedAt,
		}

		responses = append(responses, response)
	}

	return responses, nil
}

// GetDeliveryMetrics retrieves delivery performance metrics for a date range
func (uc *TrackingUseCase) GetDeliveryMetrics(ctx context.Context, startDate, endDate time.Time) (*DeliveryMetrics, error) {
	// Get delivery counts by status
	statusCounts, err := uc.deliveryRepo.GetDeliveryCountByStatus(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery status counts: %w", err)
	}

	// Get delivery counts by method
	methodCounts, err := uc.deliveryRepo.GetDeliveryCountByMethod(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery method counts: %w", err)
	}

	// Get average delivery time (using self delivery method as default)
	avgDeliveryTime, err := uc.deliveryRepo.GetAverageDeliveryTime(ctx, entity.DeliveryMethodSelfDelivery, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get average delivery time: %w", err)
	}

	// Get total revenue
	totalRevenue, err := uc.deliveryRepo.GetTotalDeliveryFees(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get total delivery fees: %w", err)
	}

	// Calculate success rate
	var totalDeliveries int64
	var successfulDeliveries int64
	for status, count := range statusCounts {
		totalDeliveries += count
		if status == entity.DeliveryStatusDelivered {
			successfulDeliveries += count
		}
	}

	successRate := float64(0)
	if totalDeliveries > 0 {
		successRate = float64(successfulDeliveries) / float64(totalDeliveries) * 100
	}

	// Create metrics response
	metrics := &DeliveryMetrics{
		TotalDeliveries:     totalDeliveries,
		DeliveriesByStatus:  statusCounts,
		DeliveriesByMethod:  methodCounts,
		AverageDeliveryTime: avgDeliveryTime,
		TotalRevenue:        totalRevenue,
		SuccessRate:         successRate,
		OnTimeDeliveryRate:  0, // Calculate this based on estimated vs actual delivery times
	}

	return metrics, nil
}

// DeliveryMetrics represents delivery performance metrics
type DeliveryMetrics struct {
	TotalDeliveries     int64                                    `json:"total_deliveries"`
	DeliveriesByStatus  map[entity.DeliveryStatus]int64          `json:"deliveries_by_status"`
	DeliveriesByMethod  map[entity.DeliveryMethod]int64          `json:"deliveries_by_method"`
	AverageDeliveryTime float64                                  `json:"average_delivery_time_hours"`
	TotalRevenue        float64                                  `json:"total_revenue"`
	SuccessRate         float64                                  `json:"success_rate_percentage"`
	OnTimeDeliveryRate  float64                                  `json:"on_time_delivery_rate"`
}

// GetDeliveryHistory retrieves the complete history of a delivery
func (uc *TrackingUseCase) GetDeliveryHistory(ctx context.Context, deliveryID uuid.UUID) ([]*entity.DeliverySnapshot, error) {
	snapshots, err := uc.snapshotRepo.GetByDeliveryID(ctx, deliveryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery history: %w", err)
	}

	return snapshots, nil
}

// CreateTrackingSnapshot creates a new tracking snapshot for a delivery
func (uc *TrackingUseCase) CreateTrackingSnapshot(ctx context.Context, deliveryID uuid.UUID, eventType string, eventData map[string]interface{}, createdBy string) error {
	snapshot := &entity.DeliverySnapshot{
		ID:             uuid.New(),
		DeliveryID:     deliveryID,
		SnapshotType:   entity.SnapshotTypeInTransit,
		SnapshotData:   eventData,
		TriggeredBy:    createdBy,
		TriggeredEvent: eventType,
		CreatedAt:      time.Now(),
	}

	if err := uc.snapshotRepo.Create(ctx, snapshot); err != nil {
		return fmt.Errorf("failed to create tracking snapshot: %w", err)
	}

	// Publish event
	uc.eventPublisher.Publish(ctx, "delivery.tracking_update", map[string]interface{}{
		"delivery_id":  deliveryID.String(),
		"event_type":   eventType,
		"event_data":   eventData,
		"created_by":   createdBy,
		"created_at":   snapshot.CreatedAt,
	})

	return nil
}
