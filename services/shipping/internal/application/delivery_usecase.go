package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"shipping/internal/domain/entity"
	"shipping/internal/domain/repository"
)

// DeliveryUsecase handles core delivery operations
type DeliveryUsecase struct {
	deliveryRepo     repository.DeliveryRepository
	vehicleRepo      repository.VehicleRepository
	routeRepo        repository.RouteRepository
	providerRepo     repository.ProviderRepository
	snapshotRepo     repository.SnapshotRepository
	coverageAreaRepo repository.CoverageAreaRepository
	eventPublisher   EventPublisher
	cache           Cache
}

// NewDeliveryUsecase creates a new delivery use case
func NewDeliveryUsecase(
	deliveryRepo repository.DeliveryRepository,
	vehicleRepo repository.VehicleRepository,
	routeRepo repository.RouteRepository,
	providerRepo repository.ProviderRepository,
	snapshotRepo repository.SnapshotRepository,
	coverageAreaRepo repository.CoverageAreaRepository,
	eventPublisher EventPublisher,
	cache Cache,
) *DeliveryUsecase {
	return &DeliveryUsecase{
		deliveryRepo:     deliveryRepo,
		vehicleRepo:      vehicleRepo,
		routeRepo:        routeRepo,
		providerRepo:     providerRepo,
		snapshotRepo:     snapshotRepo,
		coverageAreaRepo: coverageAreaRepo,
		eventPublisher:   eventPublisher,
		cache:           cache,
	}
}

// CreateDeliveryRequest represents a request to create a delivery
type CreateDeliveryRequest struct {
	OrderID           uuid.UUID                `json:"order_id"`
	CustomerID        uuid.UUID                `json:"customer_id"`
	CustomerAddressID uuid.UUID                `json:"customer_address_id"`
	DeliveryMethod    entity.DeliveryMethod    `json:"delivery_method"`
	DeliveryFee       float64                  `json:"delivery_fee"`
	CODAmount         float64                  `json:"cod_amount"`
	PlannedDate       time.Time                `json:"planned_delivery_date"`
	Instructions      string                   `json:"delivery_instructions,omitempty"`
	RequiresManual    bool                     `json:"requires_manual_coordination"`
}

// CreateDeliveryResponse represents a response from creating a delivery
type CreateDeliveryResponse struct {
	DeliveryID        uuid.UUID `json:"delivery_id"`
	TrackingNumber    string    `json:"tracking_number,omitempty"`
	Status            entity.DeliveryStatus `json:"status"`
	EstimatedDelivery *time.Time `json:"estimated_delivery,omitempty"`
	ProviderInfo      map[string]interface{} `json:"provider_info,omitempty"`
}

// CreateDelivery creates a new delivery order
func (uc *DeliveryUsecase) CreateDelivery(ctx context.Context, req *CreateDeliveryRequest) (*CreateDeliveryResponse, error) {
	// 1. Validate request
	if err := uc.validateCreateRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	
	// 2. Create delivery entity
	delivery := entity.NewDeliveryOrder(
		req.OrderID,
		req.CustomerID,
		req.CustomerAddressID,
		req.DeliveryMethod,
		req.DeliveryFee,
		req.CODAmount,
	)
	
	if req.Instructions != "" {
		delivery.DeliveryInstructions = &req.Instructions
	}
	
	delivery.PlannedDeliveryDate = req.PlannedDate
	delivery.RequiresManualCoordination = req.RequiresManual
	
	// 3. Persist delivery
	if err := uc.deliveryRepo.Create(ctx, delivery); err != nil {
		return nil, fmt.Errorf("failed to create delivery: %w", err)
	}
	
	// 4. Create snapshot
	if err := uc.createDeliverySnapshot(ctx, delivery, entity.SnapshotTypeCreated, "order_confirmed", "system_auto", nil); err != nil {
		// Log error but don't fail the operation
		// TODO: Add proper logging
	}
	
	// 5. Publish event
	event := DeliveryCreatedEvent{
		DeliveryID:     delivery.ID,
		OrderID:        delivery.OrderID,
		CustomerID:     delivery.CustomerID,
		DeliveryMethod: delivery.DeliveryMethod,
		Status:         delivery.Status,
		CreatedAt:      delivery.CreatedAt,
	}
	
	if err := uc.eventPublisher.Publish(ctx, "delivery.created", event); err != nil {
		// Log error but don't fail the operation
		// TODO: Add proper logging
	}
	
	// 6. Cache delivery info for quick access
	uc.cacheDeliveryInfo(ctx, delivery)
	
	response := &CreateDeliveryResponse{
		DeliveryID: delivery.ID,
		Status:     delivery.Status,
	}
	
	return response, nil
}

// GetDelivery gets a delivery by ID
func (uc *DeliveryUsecase) GetDelivery(ctx context.Context, deliveryID uuid.UUID) (*entity.DeliveryOrder, error) {
	// Try cache first
	if cached := uc.getCachedDelivery(ctx, deliveryID); cached != nil {
		return cached, nil
	}
	
	// Get from repository
	delivery, err := uc.deliveryRepo.GetByID(ctx, deliveryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery: %w", err)
	}
	
	// Cache for future use
	uc.cacheDeliveryInfo(ctx, delivery)
	
	return delivery, nil
}

// UpdateDeliveryStatus updates the status of a delivery
func (uc *DeliveryUsecase) UpdateDeliveryStatus(ctx context.Context, deliveryID uuid.UUID, status entity.DeliveryStatus, userID *uuid.UUID) error {
	// 1. Get current delivery
	delivery, err := uc.deliveryRepo.GetByID(ctx, deliveryID)
	if err != nil {
		return fmt.Errorf("failed to get delivery: %w", err)
	}
	
	oldStatus := delivery.Status
	
	// 2. Update status
	delivery.UpdateStatus(status)
	
	// 3. Persist changes
	if err := uc.deliveryRepo.Update(ctx, delivery); err != nil {
		return fmt.Errorf("failed to update delivery: %w", err)
	}
	
	// 4. Create snapshot
	triggeredBy := "system_auto"
	if userID != nil {
		triggeredBy = "user_action"
	}
	
	if err := uc.createDeliverySnapshot(ctx, delivery, entity.SnapshotTypeStatusUpdated, triggeredBy, "status_update", userID); err != nil {
		// Log error but don't fail the operation
		// TODO: Add proper logging
	}
	
	// 5. Publish event
	event := DeliveryStatusUpdatedEvent{
		DeliveryID:  delivery.ID,
		OrderID:     delivery.OrderID,
		CustomerID:  delivery.CustomerID,
		OldStatus:   oldStatus,
		NewStatus:   status,
		UpdatedAt:   delivery.UpdatedAt,
		UpdatedBy:   userID,
	}
	
	if err := uc.eventPublisher.Publish(ctx, "delivery.status_updated", event); err != nil {
		// Log error but don't fail the operation
		// TODO: Add proper logging
	}
	
	// 6. Clear cache
	uc.clearDeliveryCache(ctx, deliveryID)
	
	// 7. Handle status-specific logic
	switch status {
	case entity.DeliveryStatusDelivered:
		return uc.handleDeliveryCompleted(ctx, delivery)
	case entity.DeliveryStatusFailed:
		return uc.handleDeliveryFailed(ctx, delivery)
	case entity.DeliveryStatusCancelled:
		return uc.handleDeliveryCancelled(ctx, delivery)
	}
	
	return nil
}

// AssignVehicle assigns a vehicle to a delivery
func (uc *DeliveryUsecase) AssignVehicle(ctx context.Context, deliveryID, vehicleID uuid.UUID) error {
	// 1. Validate vehicle exists and is available
	vehicle, err := uc.vehicleRepo.GetByID(ctx, vehicleID)
	if err != nil {
		return fmt.Errorf("failed to get vehicle: %w", err)
	}
	
	if !vehicle.IsAvailable() {
		return fmt.Errorf("vehicle %s is not available", vehicle.LicensePlate)
	}
	
	// 2. Update delivery
	if err := uc.deliveryRepo.AssignVehicle(ctx, deliveryID, vehicleID); err != nil {
		return fmt.Errorf("failed to assign vehicle: %w", err)
	}
	
	// 3. Update vehicle status
	if err := uc.vehicleRepo.UpdateStatus(ctx, vehicleID, entity.VehicleStatusOnRoute); err != nil {
		// Log error but continue
		// TODO: Add proper logging
	}
	
	// 4. Get updated delivery and create snapshot
	delivery, err := uc.deliveryRepo.GetByID(ctx, deliveryID)
	if err == nil {
		uc.createDeliverySnapshot(ctx, delivery, entity.SnapshotTypeAssigned, "vehicle_assigned", "system_auto", nil)
	}
	
	// 5. Clear cache
	uc.clearDeliveryCache(ctx, deliveryID)
	
	return nil
}

// GetDeliveryTimeline gets the timeline of a delivery from snapshots
func (uc *DeliveryUsecase) GetDeliveryTimeline(ctx context.Context, deliveryID uuid.UUID) ([]*entity.DeliverySnapshot, error) {
	snapshots, err := uc.snapshotRepo.GetDeliveryTimeline(ctx, deliveryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery timeline: %w", err)
	}
	
	return snapshots, nil
}

// GetDeliveryMetrics gets delivery metrics for a date range
func (uc *DeliveryUsecase) GetDeliveryMetrics(ctx context.Context, startDate, endDate time.Time) (*repository.DeliveryMetrics, error) {
	// Get counts by status
	statusCounts, err := uc.deliveryRepo.GetDeliveryCountByStatus(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get status counts: %w", err)
	}
	
	// Get counts by method
	methodCounts, err := uc.deliveryRepo.GetDeliveryCountByMethod(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get method counts: %w", err)
	}
	
	// Get average delivery time
	avgTime, err := uc.deliveryRepo.GetAverageDeliveryTime(ctx, "", startDate, endDate)
	if err != nil {
		avgTime = 0 // Default if error
	}
	
	// Get total revenue
	totalRevenue, err := uc.deliveryRepo.GetTotalDeliveryFees(ctx, startDate, endDate)
	if err != nil {
		totalRevenue = 0 // Default if error
	}
	
	// Calculate totals
	var totalDeliveries int64
	for _, count := range statusCounts {
		totalDeliveries += count
	}
	
	// Calculate success rate
	successfulDeliveries := statusCounts[entity.DeliveryStatusDelivered]
	successRate := float64(0)
	if totalDeliveries > 0 {
		successRate = float64(successfulDeliveries) / float64(totalDeliveries) * 100
	}
	
	metrics := &repository.DeliveryMetrics{
		TotalDeliveries:     totalDeliveries,
		DeliveriesByStatus:  statusCounts,
		DeliveriesByMethod:  methodCounts,
		AverageDeliveryTime: avgTime,
		TotalRevenue:        totalRevenue,
		SuccessRate:         successRate,
		OnTimeDeliveryRate:  85.0, // TODO: Calculate actual on-time rate
	}
	
	return metrics, nil
}

// Helper methods

func (uc *DeliveryUsecase) validateCreateRequest(req *CreateDeliveryRequest) error {
	if req.OrderID == uuid.Nil {
		return fmt.Errorf("order ID is required")
	}
	
	if req.CustomerID == uuid.Nil {
		return fmt.Errorf("customer ID is required")
	}
	
	if req.CustomerAddressID == uuid.Nil {
		return fmt.Errorf("customer address ID is required")
	}
	
	if req.DeliveryFee < 0 {
		return fmt.Errorf("delivery fee cannot be negative")
	}
	
	if req.CODAmount < 0 {
		return fmt.Errorf("COD amount cannot be negative")
	}
	
	return nil
}

func (uc *DeliveryUsecase) createDeliverySnapshot(ctx context.Context, delivery *entity.DeliveryOrder, snapshotType entity.SnapshotType, triggeredBy, triggeredEvent string, userID *uuid.UUID) error {
	snapshot, err := entity.NewDeliverySnapshotFromDelivery(
		delivery,
		snapshotType,
		triggeredBy,
		triggeredEvent,
		userID,
		nil, // TODO: Get previous snapshot ID
	)
	if err != nil {
		return err
	}
	
	return uc.snapshotRepo.Create(ctx, snapshot)
}

func (uc *DeliveryUsecase) handleDeliveryCompleted(ctx context.Context, delivery *entity.DeliveryOrder) error {
	// Create completion snapshot
	if err := uc.createDeliverySnapshot(ctx, delivery, entity.SnapshotTypeDelivered, "delivery_completed", "system_auto", nil); err != nil {
		// Log error but continue
	}
	
	// Publish completion event
	event := DeliveryCompletedEvent{
		DeliveryID:        delivery.ID,
		OrderID:           delivery.OrderID,
		CustomerID:        delivery.CustomerID,
		DeliveryMethod:    delivery.DeliveryMethod,
		DeliveryFee:       delivery.DeliveryFee,
		CompletedAt:       *delivery.ActualDeliveryTime,
		VehicleID:         delivery.VehicleID,
	}
	
	return uc.eventPublisher.Publish(ctx, "delivery.completed", event)
}

func (uc *DeliveryUsecase) handleDeliveryFailed(ctx context.Context, delivery *entity.DeliveryOrder) error {
	// Create failure snapshot
	if err := uc.createDeliverySnapshot(ctx, delivery, entity.SnapshotTypeFailed, "delivery_failed", "system_auto", nil); err != nil {
		// Log error but continue
	}
	
	// Release vehicle if assigned
	if delivery.VehicleID != nil {
		uc.vehicleRepo.UpdateStatus(ctx, *delivery.VehicleID, entity.VehicleStatusActive)
	}
	
	// Publish failure event
	event := DeliveryFailedEvent{
		DeliveryID:     delivery.ID,
		OrderID:        delivery.OrderID,
		CustomerID:     delivery.CustomerID,
		DeliveryMethod: delivery.DeliveryMethod,
		FailedAt:       delivery.UpdatedAt,
		Reason:         delivery.Notes,
	}
	
	return uc.eventPublisher.Publish(ctx, "delivery.failed", event)
}

func (uc *DeliveryUsecase) handleDeliveryCancelled(ctx context.Context, delivery *entity.DeliveryOrder) error {
	// Create cancellation snapshot
	if err := uc.createDeliverySnapshot(ctx, delivery, entity.SnapshotTypeCancelled, "delivery_cancelled", "system_auto", nil); err != nil {
		// Log error but continue
	}
	
	// Release vehicle if assigned
	if delivery.VehicleID != nil {
		uc.vehicleRepo.UpdateStatus(ctx, *delivery.VehicleID, entity.VehicleStatusActive)
	}
	
	// Publish cancellation event
	event := DeliveryCancelledEvent{
		DeliveryID:     delivery.ID,
		OrderID:        delivery.OrderID,
		CustomerID:     delivery.CustomerID,
		DeliveryMethod: delivery.DeliveryMethod,
		CancelledAt:    delivery.UpdatedAt,
		Reason:         delivery.Notes,
	}
	
	return uc.eventPublisher.Publish(ctx, "delivery.cancelled", event)
}

func (uc *DeliveryUsecase) cacheDeliveryInfo(ctx context.Context, delivery *entity.DeliveryOrder) {
	key := fmt.Sprintf("delivery:%s", delivery.ID.String())
	// TODO: Implement caching with proper serialization and TTL
	_ = key
}

func (uc *DeliveryUsecase) getCachedDelivery(ctx context.Context, deliveryID uuid.UUID) *entity.DeliveryOrder {
	key := fmt.Sprintf("delivery:%s", deliveryID.String())
	// TODO: Implement cache retrieval with proper deserialization
	_ = key
	return nil
}

func (uc *DeliveryUsecase) clearDeliveryCache(ctx context.Context, deliveryID uuid.UUID) {
	key := fmt.Sprintf("delivery:%s", deliveryID.String())
	// TODO: Implement cache clearing
	_ = key
}

// Event types for delivery operations
type DeliveryCreatedEvent struct {
	DeliveryID     uuid.UUID             `json:"delivery_id"`
	OrderID        uuid.UUID             `json:"order_id"`
	CustomerID     uuid.UUID             `json:"customer_id"`
	DeliveryMethod entity.DeliveryMethod `json:"delivery_method"`
	Status         entity.DeliveryStatus `json:"status"`
	CreatedAt      time.Time             `json:"created_at"`
}

type DeliveryStatusUpdatedEvent struct {
	DeliveryID uuid.UUID             `json:"delivery_id"`
	OrderID    uuid.UUID             `json:"order_id"`
	CustomerID uuid.UUID             `json:"customer_id"`
	OldStatus  entity.DeliveryStatus `json:"old_status"`
	NewStatus  entity.DeliveryStatus `json:"new_status"`
	UpdatedAt  time.Time             `json:"updated_at"`
	UpdatedBy  *uuid.UUID            `json:"updated_by,omitempty"`
}

type DeliveryCompletedEvent struct {
	DeliveryID     uuid.UUID             `json:"delivery_id"`
	OrderID        uuid.UUID             `json:"order_id"`
	CustomerID     uuid.UUID             `json:"customer_id"`
	DeliveryMethod entity.DeliveryMethod `json:"delivery_method"`
	DeliveryFee    float64               `json:"delivery_fee"`
	CompletedAt    time.Time             `json:"completed_at"`
	VehicleID      *uuid.UUID            `json:"vehicle_id,omitempty"`
}

type DeliveryFailedEvent struct {
	DeliveryID     uuid.UUID             `json:"delivery_id"`
	OrderID        uuid.UUID             `json:"order_id"`
	CustomerID     uuid.UUID             `json:"customer_id"`
	DeliveryMethod entity.DeliveryMethod `json:"delivery_method"`
	FailedAt       time.Time             `json:"failed_at"`
	Reason         *string               `json:"reason,omitempty"`
}

type DeliveryCancelledEvent struct {
	DeliveryID     uuid.UUID             `json:"delivery_id"`
	OrderID        uuid.UUID             `json:"order_id"`
	CustomerID     uuid.UUID             `json:"customer_id"`
	DeliveryMethod entity.DeliveryMethod `json:"delivery_method"`
	CancelledAt    time.Time             `json:"cancelled_at"`
	Reason         *string               `json:"reason,omitempty"`
}

// Interface definitions for dependencies
type EventPublisher interface {
	Publish(ctx context.Context, topic string, event interface{}) error
}

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}
