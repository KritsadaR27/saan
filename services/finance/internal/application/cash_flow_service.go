package application

import (
	"time"

	"finance/internal/domain"
	"finance/internal/infrastructure/cache"
	"finance/internal/infrastructure/database/repositories"

	"github.com/google/uuid"
)

type cashFlowService struct {
	repos *repositories.Repositories
	redis cache.RedisClient
}

func NewCashFlowService(repos *repositories.Repositories, redis cache.RedisClient) domain.CashFlowService {
	return &cashFlowService{
		repos: repos,
		redis: redis,
	}
}

func (c *cashFlowService) RecordTransaction(entityType string, entityID uuid.UUID, txType domain.CashFlowType, amount float64, description, reference string, createdBy *uuid.UUID) error {
	// Validate inputs
	if amount <= 0 {
		return domain.ErrInvalidAmount
	}
	if description == "" {
		return domain.ErrMissingDescription
	}

	// Calculate running balance
	runningBalance, err := c.repos.CashFlow.CalculateRunningBalance(entityType, entityID, txType, amount)
	if err != nil {
		return err
	}

	// Create cash flow record
	record := &domain.CashFlowRecord{
		ID:             uuid.New(),
		EntityType:     entityType,
		EntityID:       entityID,
		TransactionType: txType,
		Amount:         amount,
		Description:    description,
		Reference:      reference,
		RunningBalance: runningBalance,
		CreatedBy:      createdBy,
		CreatedAt:      time.Now(),
	}

	return c.repos.CashFlow.Create(record)
}

func (c *cashFlowService) GetEntityCashFlow(entityType string, entityID uuid.UUID, limit int) ([]*domain.CashFlowRecord, error) {
	if limit <= 0 {
		limit = 50 // Default limit
	}
	if limit > 1000 {
		limit = 1000 // Maximum limit
	}

	return c.repos.CashFlow.GetByEntity(entityType, entityID, limit)
}

func (c *cashFlowService) GetCurrentBalance(entityType string, entityID uuid.UUID) (float64, error) {
	return c.repos.CashFlow.GetCurrentBalance(entityType, entityID)
}

func (c *cashFlowService) GetBalanceHistory(entityType string, entityID uuid.UUID, limit int) ([]*domain.CashFlowRecord, error) {
	if limit <= 0 {
		limit = 100 // Default limit for balance history
	}
	
	return c.repos.CashFlow.GetBalanceHistory(entityType, entityID, limit)
}

func (c *cashFlowService) RecordInflow(entityType string, entityID uuid.UUID, amount float64, description, reference string, createdBy *uuid.UUID) error {
	return c.RecordTransaction(entityType, entityID, domain.CashInflow, amount, description, reference, createdBy)
}

func (c *cashFlowService) RecordOutflow(entityType string, entityID uuid.UUID, amount float64, description, reference string, createdBy *uuid.UUID) error {
	return c.RecordTransaction(entityType, entityID, domain.CashOutflow, amount, description, reference, createdBy)
}

func (c *cashFlowService) GetTotalInflows(entityType string, entityID uuid.UUID) (float64, error) {
	return c.repos.CashFlow.GetTotalInflows(entityType, entityID)
}

func (c *cashFlowService) GetTotalOutflows(entityType string, entityID uuid.UUID) (float64, error) {
	return c.repos.CashFlow.GetTotalOutflows(entityType, entityID)
}

func (c *cashFlowService) GetNetCashFlow(entityType string, entityID uuid.UUID) (float64, error) {
	inflows, err := c.GetTotalInflows(entityType, entityID)
	if err != nil {
		return 0, err
	}

	outflows, err := c.GetTotalOutflows(entityType, entityID)
	if err != nil {
		return 0, err
	}

	return inflows - outflows, nil
}

func (c *cashFlowService) ValidateEntity(entityType string, entityID uuid.UUID) error {
	switch entityType {
	case "branch", "vehicle", "central":
		return nil
	default:
		return domain.ErrInvalidEntity
	}
}
