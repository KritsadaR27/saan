package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/saan-system/services/customer/internal/domain/entity"
	"github.com/saan-system/services/customer/internal/domain/repository"
)

// PointsUsecase handles customer points business logic
type PointsUsecase struct {
	pointsRepo      repository.CustomerPointsRepository
	customerRepo    repository.CustomerRepository
	vipBenefitsRepo repository.VIPTierBenefitsRepository
	eventPublisher  repository.EventPublisher
}

// NewPointsUsecase creates a new points usecase
func NewPointsUsecase(
	pointsRepo repository.CustomerPointsRepository,
	customerRepo repository.CustomerRepository,
	vipBenefitsRepo repository.VIPTierBenefitsRepository,
	eventPublisher repository.EventPublisher,
) *PointsUsecase {
	return &PointsUsecase{
		pointsRepo:      pointsRepo,
		customerRepo:    customerRepo,
		vipBenefitsRepo: vipBenefitsRepo,
		eventPublisher:  eventPublisher,
	}
}

// EarnPointsRequest represents a request to earn points
type EarnPointsRequest struct {
	CustomerID    uuid.UUID  `json:"customer_id" validate:"required"`
	Points        int        `json:"points" validate:"required,min=1"`
	Source        string     `json:"source" validate:"required"`
	Description   string     `json:"description"`
	ReferenceID   *uuid.UUID `json:"reference_id"`
	ReferenceType *string    `json:"reference_type"`
}

// RedeemPointsRequest represents a request to redeem points
type RedeemPointsRequest struct {
	CustomerID    uuid.UUID  `json:"customer_id" validate:"required"`
	Points        int        `json:"points" validate:"required,min=1"`
	Source        string     `json:"source" validate:"required"`
	Description   string     `json:"description"`
	ReferenceID   *uuid.UUID `json:"reference_id"`
	ReferenceType *string    `json:"reference_type"`
}

// EarnPoints adds points to a customer's balance
func (uc *PointsUsecase) EarnPoints(ctx context.Context, req *EarnPointsRequest) (*entity.CustomerPointsTransaction, error) {
	// Get customer to verify exists and get current balance
	customer, err := uc.customerRepo.GetByID(ctx, req.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	// Get VIP benefits for points multiplier
	benefits, err := uc.vipBenefitsRepo.GetByTier(ctx, customer.Tier)
	if err != nil {
		return nil, fmt.Errorf("failed to get VIP benefits: %w", err)
	}

	// Calculate actual points with multiplier
	actualPoints := int(float64(req.Points) * benefits.PointsMultiplier)

	// Create points transaction
	transaction := &entity.CustomerPointsTransaction{
		ID:            uuid.New(),
		CustomerID:    req.CustomerID,
		TransactionID: uuid.New(),
		Type:          entity.PointsEarned,
		Points:        actualPoints,
		Balance:       customer.PointsBalance + actualPoints,
		ReferenceID:   req.ReferenceID,
		ReferenceType: req.ReferenceType,
		Source:        req.Source,
		Description:   req.Description,
		CreatedAt:     time.Now(),
	}

	// Create transaction
	if err := uc.pointsRepo.CreateTransaction(ctx, transaction); err != nil {
		return nil, fmt.Errorf("failed to create points transaction: %w", err)
	}

	// Update customer balance
	customer.PointsBalance += actualPoints
	if err := uc.customerRepo.Update(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to update customer balance: %w", err)
	}

	// Publish event - Use the EarnPoints method instead
	if err := uc.pointsRepo.EarnPoints(ctx, req.CustomerID, actualPoints, req.Source, req.Description, req.ReferenceID, req.ReferenceType); err != nil {
		// Log error but don't fail the operation
		// TODO: Add proper logging
	}

	return transaction, nil
}

// RedeemPoints deducts points from a customer's balance
func (uc *PointsUsecase) RedeemPoints(ctx context.Context, req *RedeemPointsRequest) (*entity.CustomerPointsTransaction, error) {
	// Get customer to verify exists and check balance
	customer, err := uc.customerRepo.GetByID(ctx, req.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	// Check if customer has enough points
	if customer.PointsBalance < req.Points {
		return nil, entity.ErrInsufficientPoints
	}

	// Create points transaction
	transaction := &entity.CustomerPointsTransaction{
		ID:            uuid.New(),
		CustomerID:    req.CustomerID,
		TransactionID: uuid.New(),
		Type:          entity.PointsRedeemed,
		Points:        -req.Points, // Negative for redemption
		Balance:       customer.PointsBalance - req.Points,
		ReferenceID:   req.ReferenceID,
		ReferenceType: req.ReferenceType,
		Source:        req.Source,
		Description:   req.Description,
		CreatedAt:     time.Now(),
	}

	// Create transaction
	if err := uc.pointsRepo.CreateTransaction(ctx, transaction); err != nil {
		return nil, fmt.Errorf("failed to create points transaction: %w", err)
	}

	// Update customer balance
	customer.PointsBalance -= req.Points
	if err := uc.customerRepo.Update(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to update customer balance: %w", err)
	}

	// Use RedeemPoints method instead of event
	if err := uc.pointsRepo.RedeemPoints(ctx, req.CustomerID, req.Points, req.Source, req.Description, req.ReferenceID, req.ReferenceType); err != nil {
		// Log error but don't fail the operation
		// TODO: Add proper logging
	}

	return transaction, nil
}

// GetPointsBalance retrieves a customer's current points balance
func (uc *PointsUsecase) GetPointsBalance(ctx context.Context, customerID uuid.UUID) (int, error) {
	customer, err := uc.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		return 0, fmt.Errorf("failed to get customer: %w", err)
	}

	return customer.PointsBalance, nil
}

// GetPointsHistory retrieves a customer's points transaction history
func (uc *PointsUsecase) GetPointsHistory(ctx context.Context, customerID uuid.UUID, limit, offset int) ([]entity.CustomerPointsTransaction, error) {
	return uc.pointsRepo.GetTransactionsByCustomer(ctx, customerID, limit, offset)
}

// CalculatePointsFromSpending calculates points earned from spending amount
func (uc *PointsUsecase) CalculatePointsFromSpending(ctx context.Context, customerID uuid.UUID, spentAmount float64) (int, error) {
	// Get customer to get their tier
	customer, err := uc.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		return 0, fmt.Errorf("failed to get customer: %w", err)
	}

	// Get VIP benefits for points calculation
	benefits, err := uc.vipBenefitsRepo.GetByTier(ctx, customer.Tier)
	if err != nil {
		return 0, fmt.Errorf("failed to get VIP benefits: %w", err)
	}

	// Base points: 1 point per 100 THB spent
	basePoints := int(spentAmount / 100)
	
	// Apply tier multiplier
	actualPoints := int(float64(basePoints) * benefits.PointsMultiplier)

	return actualPoints, nil
}

// ProcessOrderPoints processes points earning from an order
func (uc *PointsUsecase) ProcessOrderPoints(ctx context.Context, customerID uuid.UUID, orderID uuid.UUID, spentAmount float64) (*entity.CustomerPointsTransaction, error) {
	// Calculate points
	points, err := uc.CalculatePointsFromSpending(ctx, customerID, spentAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate points: %w", err)
	}

	if points <= 0 {
		return nil, nil // No points to award
	}

	// Create earn points request
	req := &EarnPointsRequest{
		CustomerID:    customerID,
		Points:        points,
		Source:        "order",
		Description:   fmt.Sprintf("Points earned from order %.2f THB", spentAmount),
		ReferenceID:   &orderID,
		ReferenceType: func() *string { s := "order"; return &s }(),
	}

	return uc.EarnPoints(ctx, req)
}

// GetPointsStats retrieves points statistics for a customer
func (uc *PointsUsecase) GetPointsStats(ctx context.Context, customerID uuid.UUID) (*entity.CustomerPointsStats, error) {
	// Get customer
	customer, err := uc.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	// Get points transactions for stats
	transactions, err := uc.pointsRepo.GetTransactionsByCustomer(ctx, customerID, 1000, 0) // Get recent transactions
	if err != nil {
		return nil, fmt.Errorf("failed to get points history: %w", err)
	}

	// Calculate stats
	var totalEarned, totalRedeemed int
	for _, txn := range transactions {
		if txn.Points > 0 {
			totalEarned += txn.Points
		} else {
			totalRedeemed += -txn.Points
		}
	}

	return &entity.CustomerPointsStats{
		CurrentBalance: customer.PointsBalance,
		TotalEarned:    totalEarned,
		TotalRedeemed:  totalRedeemed,
		TotalTransactions: len(transactions),
	}, nil
}
