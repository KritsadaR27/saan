package application

import (
	"time"

	"finance/internal/domain"
	"finance/internal/infrastructure/cache"
	"finance/internal/infrastructure/database/repositories"

	"github.com/google/uuid"
)

type allocationService struct {
	repos *repositories.Repositories
	redis cache.RedisClient
}

func NewAllocationService(repos *repositories.Repositories, redis cache.RedisClient) domain.AllocationService {
	return &allocationService{
		repos: repos,
		redis: redis,
	}
}

func (a *allocationService) CalculateAllocations(revenue float64, branchID, vehicleID *uuid.UUID) (map[domain.AccountType]float64, error) {
	// Get the active allocation rule for this entity
	rule, err := a.repos.AllocationRule.GetActiveRule(branchID, vehicleID)
	if err != nil {
		return nil, err
	}

	// Calculate allocations based on percentages
	allocations := make(map[domain.AccountType]float64)

	// Calculate each allocation
	allocations[domain.ProfitAccount] = revenue * rule.ProfitPercentage / 100
	allocations[domain.OwnerPayAccount] = revenue * rule.OwnerPayPercentage / 100
	allocations[domain.TaxAccount] = revenue * rule.TaxPercentage / 100

	// Operating account gets the remainder (should be 30% by default)
	operatingPercentage := 100.0 - rule.ProfitPercentage - rule.OwnerPayPercentage - rule.TaxPercentage
	allocations[domain.OperatingAccount] = revenue * operatingPercentage / 100

	// Revenue account tracks the total revenue
	allocations[domain.RevenueAccount] = revenue

	return allocations, nil
}

func (a *allocationService) UpdateAllocationRule(rule *domain.ProfitAllocationRule) error {
	// Validate percentages
	totalPercentage := rule.ProfitPercentage + rule.OwnerPayPercentage + rule.TaxPercentage
	if totalPercentage > 100.0 {
		return domain.ErrPercentageSum
	}

	// Set timestamps
	rule.UpdatedAt = time.Now()
	
	// If this is a new rule, set creation timestamp and generate ID
	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
		rule.CreatedAt = time.Now()
		return a.repos.AllocationRule.Create(rule)
	}

	// Otherwise, update existing rule
	return a.repos.AllocationRule.Update(rule)
}

func (a *allocationService) GetCurrentRule(branchID, vehicleID *uuid.UUID) (*domain.ProfitAllocationRule, error) {
	return a.repos.AllocationRule.GetActiveRule(branchID, vehicleID)
}

func (a *allocationService) CreateNewRule(branchID, vehicleID *uuid.UUID, profitPct, ownerPayPct, taxPct float64, updatedBy uuid.UUID) (*domain.ProfitAllocationRule, error) {
	// Validate percentages
	if profitPct < 0 || ownerPayPct < 0 || taxPct < 0 {
		return nil, domain.ErrInvalidPercentage
	}

	totalPercentage := profitPct + ownerPayPct + taxPct
	if totalPercentage > 100.0 {
		return nil, domain.ErrPercentageSum
	}

	rule := &domain.ProfitAllocationRule{
		ID:                 uuid.New(),
		BranchID:           branchID,
		VehicleID:          vehicleID,
		ProfitPercentage:   profitPct,
		OwnerPayPercentage: ownerPayPct,
		TaxPercentage:      taxPct,
		EffectiveFrom:      time.Now(),
		IsActive:           true,
		UpdatedByUserID:    updatedBy,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	err := a.repos.AllocationRule.Create(rule)
	if err != nil {
		return nil, err
	}

	return rule, nil
}

func (a *allocationService) DeactivateRule(ruleID uuid.UUID, updatedBy uuid.UUID) error {
	return a.repos.AllocationRule.DeactivateRule(ruleID, updatedBy)
}

func (a *allocationService) GetRuleHistory(branchID, vehicleID *uuid.UUID) ([]*domain.ProfitAllocationRule, error) {
	return a.repos.AllocationRule.GetByEntity(branchID, vehicleID)
}

func (a *allocationService) ValidateAllocation(revenue float64, branchID, vehicleID *uuid.UUID) error {
	if revenue < 0 {
		return domain.ErrInvalidAmount
	}

	_, err := a.repos.AllocationRule.GetActiveRule(branchID, vehicleID)
	if err != nil {
		return err
	}

	return nil
}
