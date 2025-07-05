package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"payment/internal/application/dto"
	"payment/internal/domain/entity"
	"payment/internal/domain/repository"
)

// PaymentUseCase handles payment business logic
type PaymentUseCase struct {
	paymentRepo         repository.PaymentRepository
	loyverseStoreRepo   repository.LoyverseStoreRepository
	deliveryContextRepo repository.PaymentDeliveryContextRepository
	eventRepo          repository.EventRepository
	logger             *logrus.Logger
}

// NewPaymentUseCase creates a new payment use case
func NewPaymentUseCase(
	paymentRepo repository.PaymentRepository,
	loyverseStoreRepo repository.LoyverseStoreRepository,
	deliveryContextRepo repository.PaymentDeliveryContextRepository,
	eventRepo repository.EventRepository,
	logger *logrus.Logger,
) *PaymentUseCase {
	return &PaymentUseCase{
		paymentRepo:         paymentRepo,
		loyverseStoreRepo:   loyverseStoreRepo,
		deliveryContextRepo: deliveryContextRepo,
		eventRepo:          eventRepo,
		logger:             logger,
	}
}

// CreatePayment creates a new payment transaction
func (uc *PaymentUseCase) CreatePayment(ctx context.Context, req *dto.CreatePaymentRequest) (*dto.PaymentResponse, error) {
	// Create payment entity
	payment := &entity.PaymentTransaction{
		ID:             uuid.New(),
		OrderID:        req.OrderID,
		CustomerID:     req.CustomerID,
		PaymentMethod:  req.PaymentMethod,
		PaymentChannel: req.PaymentChannel,
		PaymentTiming:  req.PaymentTiming,
		Amount:         req.Amount,
		Currency:       req.Currency,
		Status:         entity.PaymentStatusPending,
		AssignedStoreID: req.AssignedStoreID,
		Metadata:       req.Metadata,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Validate payment
	if err := uc.validatePayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("payment validation failed: %w", err)
	}

	// Auto-assign store if needed for Loyverse payments
	if payment.PaymentChannel == entity.PaymentChannelLoyversePOS && payment.AssignedStoreID == nil {
		storeID, err := uc.autoAssignStore(ctx)
		if err != nil {
			uc.logger.WithError(err).Warn("Failed to auto-assign store")
		} else {
			payment.AssignedStoreID = &storeID
		}
	}

	// Save payment
	if err := uc.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Create delivery context for COD payments
	if payment.IsCOD() && req.DeliveryContext != nil {
		deliveryContext := &entity.PaymentDeliveryContext{
			PaymentID:        payment.ID,
			DeliveryID:       req.DeliveryContext.DeliveryID,
			DriverID:         req.DeliveryContext.DriverID,
			DeliveryAddress:  req.DeliveryContext.DeliveryAddress,
			DeliveryStatus:   "pending",
			EstimatedArrival: req.DeliveryContext.EstimatedArrival,
			Instructions:     req.DeliveryContext.Instructions,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		if err := uc.deliveryContextRepo.Create(ctx, deliveryContext); err != nil {
			uc.logger.WithError(err).Error("Failed to create delivery context")
		} else {
			// Publish event
			_ = uc.eventRepo.PublishPaymentEvent(ctx, &repository.PaymentEvent{
				ID:         uuid.New(),
				EventType:  repository.EventTypeDeliveryContextCreated,
				PaymentID:  payment.ID,
				OrderID:    &payment.OrderID,
				CustomerID: &payment.CustomerID,
				Data: map[string]interface{}{
					"delivery_id": deliveryContext.DeliveryID,
					"driver_id":   deliveryContext.DriverID,
				},
				OccurredAt: time.Now(),
				Source:     "payment-service",
				Version:    "1.0",
			})
		}
	}

	// Publish payment created event
	_ = uc.eventRepo.PublishPaymentEvent(ctx, &repository.PaymentEvent{
		ID:         uuid.New(),
		EventType:  repository.EventTypePaymentCreated,
		PaymentID:  payment.ID,
		OrderID:    &payment.OrderID,
		CustomerID: &payment.CustomerID,
		Data: map[string]interface{}{
			"amount":         payment.Amount,
			"currency":       payment.Currency,
			"payment_method": payment.PaymentMethod,
			"payment_timing": payment.PaymentTiming,
		},
		OccurredAt: time.Now(),
		Source:     "payment-service",
		Version:    "1.0",
	})

	return uc.mapToPaymentResponse(ctx, payment), nil
}

// UpdatePaymentStatus updates the status of a payment
func (uc *PaymentUseCase) UpdatePaymentStatus(ctx context.Context, paymentID uuid.UUID, req *dto.UpdatePaymentStatusRequest) (*dto.PaymentResponse, error) {
	// Get existing payment
	payment, err := uc.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	oldStatus := payment.Status
	
	// Update payment
	payment.Status = req.Status
	payment.UpdatedAt = time.Now()
	
	if req.LoyverseReceiptID != nil {
		payment.LoyverseReceiptID = req.LoyverseReceiptID
	}
	if req.LoyversePaymentType != nil {
		payment.LoyversePaymentType = req.LoyversePaymentType
	}
	if req.Metadata != nil {
		payment.Metadata = req.Metadata
	}
	
	// Set paid time if completed
	if payment.Status == entity.PaymentStatusCompleted && payment.PaidAt == nil {
		now := time.Now()
		payment.PaidAt = &now
	}

	// Save updated payment
	if err := uc.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	// Publish status change event
	_ = uc.eventRepo.PublishPaymentStatusChanged(ctx, payment.ID, oldStatus, payment.Status)

	// Publish specific events for certain status changes
	switch payment.Status {
	case entity.PaymentStatusCompleted:
		_ = uc.eventRepo.PublishPaymentEvent(ctx, &repository.PaymentEvent{
			ID:         uuid.New(),
			EventType:  repository.EventTypePaymentCompleted,
			PaymentID:  payment.ID,
			OrderID:    &payment.OrderID,
			CustomerID: &payment.CustomerID,
			Data: map[string]interface{}{
				"amount":              payment.Amount,
				"currency":            payment.Currency,
				"loyverse_receipt_id": payment.LoyverseReceiptID,
			},
			OccurredAt: time.Now(),
			Source:     "payment-service",
			Version:    "1.0",
		})
	case entity.PaymentStatusFailed:
		_ = uc.eventRepo.PublishPaymentEvent(ctx, &repository.PaymentEvent{
			ID:         uuid.New(),
			EventType:  repository.EventTypePaymentFailed,
			PaymentID:  payment.ID,
			OrderID:    &payment.OrderID,
			CustomerID: &payment.CustomerID,
			Data: map[string]interface{}{
				"reason": "Payment processing failed",
			},
			OccurredAt: time.Now(),
			Source:     "payment-service",
			Version:    "1.0",
		})
	}

	return uc.mapToPaymentResponse(ctx, payment), nil
}

// GetPaymentByID retrieves a payment by ID
func (uc *PaymentUseCase) GetPaymentByID(ctx context.Context, paymentID uuid.UUID) (*dto.PaymentResponse, error) {
	payment, err := uc.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return uc.mapToPaymentResponse(ctx, payment), nil
}

// Helper methods
func (uc *PaymentUseCase) validatePayment(ctx context.Context, payment *entity.PaymentTransaction) error {
	// Basic validation
	if payment.Amount <= 0 {
		return fmt.Errorf("payment amount must be greater than 0")
	}

	if payment.Currency == "" {
		return fmt.Errorf("currency is required")
	}

	// Validate store assignment for Loyverse payments
	if payment.PaymentChannel == entity.PaymentChannelLoyversePOS && payment.AssignedStoreID != nil {
		store, err := uc.loyverseStoreRepo.GetByStoreCode(ctx, *payment.AssignedStoreID)
		if err != nil || store == nil {
			return fmt.Errorf("invalid store assignment")
		}

		if !store.IsActive {
			return fmt.Errorf("assigned store is not active")
		}
	}

	return nil
}

func (uc *PaymentUseCase) autoAssignStore(ctx context.Context) (string, error) {
	stores, err := uc.loyverseStoreRepo.GetAvailableStoresForAssignment(ctx)
	if err != nil {
		return "", err
	}

	if len(stores) == 0 {
		return "", fmt.Errorf("no available stores for assignment")
	}

	// Simple assignment logic - use the first available store
	// In production, this would use more sophisticated load balancing
	return stores[0].StoreID, nil
}

func (uc *PaymentUseCase) mapToPaymentResponse(ctx context.Context, payment *entity.PaymentTransaction) *dto.PaymentResponse {
	resp := &dto.PaymentResponse{
		ID:                  payment.ID,
		OrderID:             payment.OrderID,
		CustomerID:          payment.CustomerID,
		PaymentMethod:       payment.PaymentMethod,
		PaymentChannel:      payment.PaymentChannel,
		PaymentTiming:       payment.PaymentTiming,
		Amount:              payment.Amount,
		Currency:            payment.Currency,
		Status:              payment.Status,
		PaidAt:              payment.PaidAt,
		LoyverseReceiptID:   payment.LoyverseReceiptID,
		LoyversePaymentType: payment.LoyversePaymentType,
		AssignedStoreID:     payment.AssignedStoreID,
		Metadata:            payment.Metadata,
		CreatedAt:           payment.CreatedAt,
		UpdatedAt:           payment.UpdatedAt,
	}

	// Load delivery context if it's a COD payment
	if payment.IsCOD() {
		if deliveryContext, err := uc.deliveryContextRepo.GetByPaymentID(ctx, payment.ID); err == nil && deliveryContext != nil {
			resp.DeliveryContext = &dto.DeliveryContextResponse{
				PaymentID:         deliveryContext.PaymentID,
				DeliveryID:        deliveryContext.DeliveryID,
				DriverID:          deliveryContext.DriverID,
				DeliveryAddress:   deliveryContext.DeliveryAddress,
				DeliveryStatus:    deliveryContext.DeliveryStatus,
				EstimatedArrival:  deliveryContext.EstimatedArrival,
				ActualArrival:     deliveryContext.ActualArrival,
				Instructions:      deliveryContext.Instructions,
				CreatedAt:         deliveryContext.CreatedAt,
				UpdatedAt:         deliveryContext.UpdatedAt,
			}
		}
	}

	return resp
}
