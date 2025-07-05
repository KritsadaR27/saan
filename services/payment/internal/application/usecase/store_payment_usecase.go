package usecase

import (
	"context"
	"fmt"

	"payment/internal/application/dto"
	"payment/internal/domain/repository"
)

// Type 1: Store-based data retrieval use case
type StorePaymentUseCase struct {
	paymentRepo       repository.PaymentRepository
	loyverseStoreRepo repository.LoyverseStoreRepository
}

// NewStorePaymentUseCase creates a new store payment use case
func NewStorePaymentUseCase(
	paymentRepo repository.PaymentRepository,
	loyverseStoreRepo repository.LoyverseStoreRepository,
) *StorePaymentUseCase {
	return &StorePaymentUseCase{
		paymentRepo:       paymentRepo,
		loyverseStoreRepo: loyverseStoreRepo,
	}
}

// GetStorePayments retrieves payments for a specific store
func (uc *StorePaymentUseCase) GetStorePayments(ctx context.Context, req *dto.GetStorePaymentsRequest) (*dto.PaymentListResponse, error) {
	// Validate store exists
	store, err := uc.loyverseStoreRepo.GetByStoreCode(ctx, req.StoreID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate store: %w", err)
	}
	if store == nil {
		return nil, fmt.Errorf("store not found: %s", req.StoreID)
	}

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

	// Get payments for store
	payments, err := uc.paymentRepo.GetByStoreID(ctx, req.StoreID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get store payments: %w", err)
	}

	// Convert to response
	paymentResponses := make([]*dto.PaymentResponse, len(payments))
	for i, payment := range payments {
		paymentResponses[i] = &dto.PaymentResponse{
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
	}

	return &dto.PaymentListResponse{
		Payments: paymentResponses,
		Total:    len(payments),
		Page:     (filters.Offset / filters.Limit) + 1,
		PerPage:  filters.Limit,
		HasMore:  len(payments) == filters.Limit,
	}, nil
}

// GetStoreAnalytics retrieves analytics for a specific store
func (uc *StorePaymentUseCase) GetStoreAnalytics(ctx context.Context, req *dto.GetStoreAnalyticsRequest) (*dto.StoreAnalyticsResponse, error) {
	// Validate store exists
	store, err := uc.loyverseStoreRepo.GetByStoreCode(ctx, req.StoreID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate store: %w", err)
	}
	if store == nil {
		return nil, fmt.Errorf("store not found: %s", req.StoreID)
	}

	// Get analytics data
	analytics, err := uc.paymentRepo.GetStoreAnalytics(ctx, req.StoreID, req.DateFrom, req.DateTo)
	if err != nil {
		return nil, fmt.Errorf("failed to get store analytics: %w", err)
	}

	// Convert to response
	methodStats := make([]dto.PaymentMethodStatResponse, len(analytics.PaymentMethodStats))
	for i, stat := range analytics.PaymentMethodStats {
		methodStats[i] = dto.PaymentMethodStatResponse{
			Method:           stat.Method,
			Count:            stat.Count,
			TotalAmount:      stat.TotalAmount,
			PercentageCount:  stat.PercentageCount,
			PercentageAmount: stat.PercentageAmount,
		}
	}

	dailyStats := make([]dto.DailyPaymentStatResponse, len(analytics.DailyStats))
	for i, stat := range analytics.DailyStats {
		dailyStats[i] = dto.DailyPaymentStatResponse{
			Date:   stat.Date,
			Count:  stat.Count,
			Amount: stat.Amount,
		}
	}

	return &dto.StoreAnalyticsResponse{
		StoreID:            analytics.StoreID,
		TotalTransactions:  analytics.TotalTransactions,
		TotalAmount:        analytics.TotalAmount,
		AvgAmount:          analytics.AvgAmount,
		Currency:           analytics.Currency,
		DateFrom:           analytics.DateFrom,
		DateTo:             analytics.DateTo,
		PaymentMethodStats: methodStats,
		DailyStats:         dailyStats,
	}, nil
}
