package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"payment/internal/application/dto"
	"payment/internal/domain/repository"
)

// Type 2: Customer-based data retrieval use case
type CustomerPaymentUseCase struct {
	paymentRepo         repository.PaymentRepository
	deliveryContextRepo repository.PaymentDeliveryContextRepository
}

// NewCustomerPaymentUseCase creates a new customer payment use case
func NewCustomerPaymentUseCase(
	paymentRepo repository.PaymentRepository,
	deliveryContextRepo repository.PaymentDeliveryContextRepository,
) *CustomerPaymentUseCase {
	return &CustomerPaymentUseCase{
		paymentRepo:         paymentRepo,
		deliveryContextRepo: deliveryContextRepo,
	}
}

// GetCustomerPayments retrieves payments for a specific customer
func (uc *CustomerPaymentUseCase) GetCustomerPayments(ctx context.Context, req *dto.GetCustomerPaymentsRequest) (*dto.PaymentListResponse, error) {
	// Convert filters
	filters := repository.PaymentFilters{
		Status:         req.Filters.Status,
		PaymentMethod:  req.Filters.PaymentMethod,
		PaymentChannel: req.Filters.PaymentChannel,
		PaymentTiming:  req.Filters.PaymentTiming,
		DateFrom:       req.Filters.DateFrom,
		DateTo:         req.Filters.DateTo,
		MinAmount:      req.Filters.MinAmount,
		MaxAmount:      req.Filters.MaxAmount,
		Limit:          req.Filters.Limit,
		Offset:         req.Filters.Offset,
		SortBy:         req.Filters.SortBy,
		SortOrder:      req.Filters.SortOrder,
	}

	if filters.Limit == 0 {
		filters.Limit = 50 // Default limit
	}

	// Get payments for customer
	payments, err := uc.paymentRepo.GetByCustomerID(ctx, req.CustomerID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer payments: %w", err)
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
		Page:     (filters.Offset / filters.Limit) + 1,
		PerPage:  filters.Limit,
		HasMore:  len(payments) == filters.Limit,
	}, nil
}

// GetCustomerPaymentHistory retrieves payment history for a customer
func (uc *CustomerPaymentUseCase) GetCustomerPaymentHistory(ctx context.Context, req *dto.GetCustomerPaymentHistoryRequest) (*dto.PaymentListResponse, error) {
	limit := req.Limit
	if limit == 0 {
		limit = 20 // Default limit for history
	}

	// Get payment history
	payments, err := uc.paymentRepo.GetCustomerPaymentHistory(ctx, req.CustomerID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer payment history: %w", err)
	}

	// Convert to response
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
		PerPage:  limit,
		HasMore:  len(payments) == limit,
	}, nil
}

// GetCustomerPaymentStats retrieves payment statistics for a customer
func (uc *CustomerPaymentUseCase) GetCustomerPaymentStats(ctx context.Context, customerID uuid.UUID) (*CustomerPaymentStats, error) {
	// Get all customer payments
	filters := repository.PaymentFilters{
		Limit: 1000, // Get all payments for stats
	}
	
	payments, err := uc.paymentRepo.GetByCustomerID(ctx, customerID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer payments for stats: %w", err)
	}

	// Calculate statistics
	stats := &CustomerPaymentStats{
		CustomerID:      customerID,
		TotalPayments:   len(payments),
		TotalAmount:     0,
		CompletedCount:  0,
		PendingCount:    0,
		FailedCount:     0,
		CODCount:        0,
		OnlineCount:     0,
		MethodStats:     make(map[string]int),
		ChannelStats:    make(map[string]int),
	}

	for _, payment := range payments {
		stats.TotalAmount += payment.Amount
		
		switch payment.Status {
		case "completed":
			stats.CompletedCount++
		case "pending":
			stats.PendingCount++
		case "failed":
			stats.FailedCount++
		}

		if payment.IsCOD() {
			stats.CODCount++
		} else {
			stats.OnlineCount++
		}

		stats.MethodStats[string(payment.PaymentMethod)]++
		stats.ChannelStats[string(payment.PaymentChannel)]++
	}

	if stats.TotalPayments > 0 {
		stats.AvgAmount = stats.TotalAmount / float64(stats.TotalPayments)
	}

	return stats, nil
}

// CustomerPaymentStats represents payment statistics for a customer
type CustomerPaymentStats struct {
	CustomerID     uuid.UUID          `json:"customer_id"`
	TotalPayments  int                `json:"total_payments"`
	TotalAmount    float64            `json:"total_amount"`
	AvgAmount      float64            `json:"avg_amount"`
	CompletedCount int                `json:"completed_count"`
	PendingCount   int                `json:"pending_count"`
	FailedCount    int                `json:"failed_count"`
	CODCount       int                `json:"cod_count"`
	OnlineCount    int                `json:"online_count"`
	MethodStats    map[string]int     `json:"method_stats"`
	ChannelStats   map[string]int     `json:"channel_stats"`
}
