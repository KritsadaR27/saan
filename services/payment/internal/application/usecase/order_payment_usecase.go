package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"payment/internal/application/dto"
	"payment/internal/domain/repository"
)

// Type 3: Order-based data retrieval use case
type OrderPaymentUseCase struct {
	paymentRepo         repository.PaymentRepository
	deliveryContextRepo repository.PaymentDeliveryContextRepository
}

// NewOrderPaymentUseCase creates a new order payment use case
func NewOrderPaymentUseCase(
	paymentRepo repository.PaymentRepository,
	deliveryContextRepo repository.PaymentDeliveryContextRepository,
) *OrderPaymentUseCase {
	return &OrderPaymentUseCase{
		paymentRepo:         paymentRepo,
		deliveryContextRepo: deliveryContextRepo,
	}
}

// GetOrderPayments retrieves all payments for a specific order
func (uc *OrderPaymentUseCase) GetOrderPayments(ctx context.Context, req *dto.GetOrderPaymentsRequest) (*dto.PaymentListResponse, error) {
	// Get payments for order
	payments, err := uc.paymentRepo.GetByOrderID(ctx, req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order payments: %w", err)
	}

	// Convert to response with delivery context for COD payments
	paymentResponses := make([]*dto.PaymentResponse, len(payments))
	for i, payment := range payments {
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

		// Load delivery context for COD payments
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

		paymentResponses[i] = resp
	}

	return &dto.PaymentListResponse{
		Payments: paymentResponses,
		Total:    len(payments),
		Page:     1,
		PerPage:  len(payments),
		HasMore:  false, // Order payments are always returned completely
	}, nil
}

// GetOrderPaymentSummary retrieves payment summary for a specific order
func (uc *OrderPaymentUseCase) GetOrderPaymentSummary(ctx context.Context, req *dto.GetOrderPaymentSummaryRequest) (*dto.OrderPaymentSummaryResponse, error) {
	// Get payment summary for order
	summary, err := uc.paymentRepo.GetOrderPaymentSummary(ctx, req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order payment summary: %w", err)
	}

	return &dto.OrderPaymentSummaryResponse{
		OrderID:           summary.OrderID,
		TotalAmount:       summary.TotalAmount,
		PaidAmount:        summary.PaidAmount,
		PendingAmount:     summary.PendingAmount,
		RefundedAmount:    summary.RefundedAmount,
		Currency:          summary.Currency,
		PaymentStatus:     summary.PaymentStatus,
		TransactionCount:  summary.TransactionCount,
		LastPaymentAt:     summary.LastPaymentAt,
		PaymentMethods:    summary.PaymentMethods,
	}, nil
}

// ProcessOrderPayment processes a payment for an order with business logic
func (uc *OrderPaymentUseCase) ProcessOrderPayment(ctx context.Context, orderID uuid.UUID, paymentAmount float64) (*OrderPaymentProcessResult, error) {
	// Get current order payment summary
	summary, err := uc.paymentRepo.GetOrderPaymentSummary(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order payment summary: %w", err)
	}

	// Calculate remaining amount needed
	remainingAmount := summary.TotalAmount - summary.PaidAmount

	// Determine payment processing result
	result := &OrderPaymentProcessResult{
		OrderID:           orderID,
		ProcessedAmount:   paymentAmount,
		RemainingAmount:   remainingAmount - paymentAmount,
		OverpaidAmount:    0,
		PaymentStatus:     "partially_paid",
	}

	if paymentAmount >= remainingAmount {
		// Payment completes the order
		result.PaymentStatus = "fully_paid"
		if paymentAmount > remainingAmount {
			result.OverpaidAmount = paymentAmount - remainingAmount
		}
		result.RemainingAmount = 0
	}

	if paymentAmount < remainingAmount {
		result.PaymentStatus = "partially_paid"
	}

	return result, nil
}

// ValidateOrderPayment validates if a payment can be processed for an order
func (uc *OrderPaymentUseCase) ValidateOrderPayment(ctx context.Context, orderID uuid.UUID, paymentAmount float64) error {
	// Get current order payment summary
	summary, err := uc.paymentRepo.GetOrderPaymentSummary(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order payment summary: %w", err)
	}

	// Check if order is already fully paid
	if summary.PaymentStatus == "fully_paid" {
		return fmt.Errorf("order %s is already fully paid", orderID)
	}

	// Check if payment amount is valid
	if paymentAmount <= 0 {
		return fmt.Errorf("payment amount must be greater than 0")
	}

	// Calculate remaining amount
	remainingAmount := summary.TotalAmount - summary.PaidAmount
	
	// Allow overpayment but warn if it's excessive (more than 10% over)
	if paymentAmount > remainingAmount*1.1 {
		return fmt.Errorf("payment amount %.2f is significantly higher than remaining amount %.2f", paymentAmount, remainingAmount)
	}

	return nil
}

// GetOrderPaymentTimeline retrieves the payment timeline for an order
func (uc *OrderPaymentUseCase) GetOrderPaymentTimeline(ctx context.Context, orderID uuid.UUID) (*OrderPaymentTimeline, error) {
	// Get all payments for order sorted by creation time
	payments, err := uc.paymentRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order payments: %w", err)
	}

	timeline := &OrderPaymentTimeline{
		OrderID: orderID,
		Events:  make([]PaymentTimelineEvent, 0),
	}

	// Create timeline events from payments
	for _, payment := range payments {
		// Payment created event
		timeline.Events = append(timeline.Events, PaymentTimelineEvent{
			ID:          payment.ID,
			EventType:   "payment_created",
			Amount:      payment.Amount,
			Status:      string(payment.Status),
			Method:      string(payment.PaymentMethod),
			Channel:     string(payment.PaymentChannel),
			Timestamp:   payment.CreatedAt,
			Description: fmt.Sprintf("Payment of %.2f created via %s", payment.Amount, payment.PaymentMethod),
		})

		// Payment completion event (if completed)
		if payment.IsCompleted() && payment.PaidAt != nil {
			timeline.Events = append(timeline.Events, PaymentTimelineEvent{
				ID:          payment.ID,
				EventType:   "payment_completed",
				Amount:      payment.Amount,
				Status:      string(payment.Status),
				Method:      string(payment.PaymentMethod),
				Channel:     string(payment.PaymentChannel),
				Timestamp:   *payment.PaidAt,
				Description: fmt.Sprintf("Payment of %.2f completed", payment.Amount),
			})
		}
	}

	return timeline, nil
}

// Supporting types for order payment operations

// OrderPaymentProcessResult represents the result of processing a payment for an order
type OrderPaymentProcessResult struct {
	OrderID         uuid.UUID `json:"order_id"`
	ProcessedAmount float64   `json:"processed_amount"`
	RemainingAmount float64   `json:"remaining_amount"`
	OverpaidAmount  float64   `json:"overpaid_amount"`
	PaymentStatus   string    `json:"payment_status"`
}

// OrderPaymentTimeline represents the payment timeline for an order
type OrderPaymentTimeline struct {
	OrderID uuid.UUID              `json:"order_id"`
	Events  []PaymentTimelineEvent `json:"events"`
}

// PaymentTimelineEvent represents a single event in the payment timeline
type PaymentTimelineEvent struct {
	ID          uuid.UUID `json:"id"`
	EventType   string    `json:"event_type"`
	Amount      float64   `json:"amount"`
	Status      string    `json:"status"`
	Method      string    `json:"method"`
	Channel     string    `json:"channel"`
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
}
