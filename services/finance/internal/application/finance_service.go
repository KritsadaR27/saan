package application

import (
	"database/sql"
	"time"

	"saan/finance/internal/domain"
	"saan/finance/internal/infrastructure/cache"

	"github.com/google/uuid"
)

type financeService struct {
	db    *sql.DB
	redis cache.RedisClient
}

func NewFinanceService(db *sql.DB, redis cache.RedisClient) domain.FinanceService {
	return &financeService{
		db:    db,
		redis: redis,
	}
}

func (f *financeService) ProcessEndOfDay(date time.Time, branchID, vehicleID *uuid.UUID, sales, codCollections float64) (*domain.DailyCashSummary, error) {
	// Calculate allocations
	allocationService := &allocationService{db: f.db, redis: f.redis}
	allocations, err := allocationService.CalculateAllocations(sales, branchID, vehicleID)
	if err != nil {
		return nil, err
	}

	// Create daily summary
	summary := &domain.DailyCashSummary{
		ID:                     uuid.New(),
		BusinessDate:           date,
		BranchID:               branchID,
		VehicleID:              vehicleID,
		TotalSales:             sales,
		CODCollections:         codCollections,
		ProfitAllocation:       allocations[domain.ProfitAccount],
		OwnerPayAllocation:     allocations[domain.OwnerPayAccount],
		TaxAllocation:          allocations[domain.TaxAccount],
		AvailableForExpenses:   allocations[domain.OperatingAccount],
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	// TODO: Add event publishing mechanism if needed

	return summary, nil
}

func (f *financeService) AddExpenseEntry(summaryID uuid.UUID, category, description string, amount float64, enteredBy uuid.UUID) error {
	// TODO: Store expense entry in database
	// TODO: Add event publishing mechanism if needed
	return nil
}

func (f *financeService) CreateTransferBatch(branchID, vehicleID *uuid.UUID, transfers []*domain.CashTransfer, authorizedBy uuid.UUID) (*domain.CashTransferBatch, error) {
	totalAmount := 0.0
	for _, transfer := range transfers {
		totalAmount += transfer.Amount
	}

	batch := &domain.CashTransferBatch{
		ID:             uuid.New(),
		BatchReference: generateBatchReference(),
		BranchID:       branchID,
		VehicleID:      vehicleID,
		TotalAmount:    totalAmount,
		TransferCount:  len(transfers),
		Status:         "pending",
		AuthorizedBy:   authorizedBy,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	return batch, nil
}

func (f *financeService) ExecuteTransferBatch(batchID uuid.UUID) error {
	// Mock implementation - execute all transfers in batch
	// TODO: Add event publishing mechanism if needed
	return nil
}

func (f *financeService) GetCashStatus() (map[string]interface{}, error) {
	// Mock cash status
	return map[string]interface{}{
		"total_cash":        150000.00,
		"profit_account":    45000.00,
		"owner_pay_account": 30000.00,
		"tax_account":       22500.00,
		"operational_cash":  52500.00,
		"last_updated":      time.Now(),
	}, nil
}

func (f *financeService) ReconcileCash(summaryID uuid.UUID, actualCash float64, reconciledBy uuid.UUID) error {
	// TODO: Update summary with reconciliation data
	return nil
}

func generateBatchReference() string {
	return "BATCH_" + time.Now().Format("20060102_150405")
}

type allocationService struct {
	db    *sql.DB
	redis cache.RedisClient
}

func NewAllocationService(db *sql.DB, redis cache.RedisClient) domain.AllocationService {
	return &allocationService{
		db:    db,
		redis: redis,
	}
}

func (a *allocationService) CalculateAllocations(revenue float64, branchID, vehicleID *uuid.UUID) (map[domain.AccountType]float64, error) {
	// Default percentages (could be fetched from database)
	profitPercentage := 30.0
	ownerPayPercentage := 20.0
	taxPercentage := 15.0
	
	allocations := make(map[domain.AccountType]float64)
	allocations[domain.ProfitAccount] = revenue * (profitPercentage / 100)
	allocations[domain.OwnerPayAccount] = revenue * (ownerPayPercentage / 100)
	allocations[domain.TaxAccount] = revenue * (taxPercentage / 100)
	allocations[domain.OperatingAccount] = revenue - allocations[domain.ProfitAccount] - allocations[domain.OwnerPayAccount] - allocations[domain.TaxAccount]

	return allocations, nil
}

func (a *allocationService) GetAllocationRules() (*domain.ProfitAllocationRule, error) {
	// Mock allocation rules
	return &domain.ProfitAllocationRule{
		ProfitPercentage:   30.0,
		OwnerPayPercentage: 20.0,
		TaxPercentage:      15.0,
	}, nil
}

func (a *allocationService) UpdateAllocationRule(rule *domain.ProfitAllocationRule) error {
	// TODO: Update allocation rule in database
	return nil
}

func (a *allocationService) GetCurrentRule(branchID, vehicleID *uuid.UUID) (*domain.ProfitAllocationRule, error) {
	// Mock current rule
	return &domain.ProfitAllocationRule{
		ID:                 uuid.New(),
		BranchID:           branchID,
		VehicleID:          vehicleID,
		ProfitPercentage:   30.0,
		OwnerPayPercentage: 20.0,
		TaxPercentage:      15.0,
		EffectiveFrom:      time.Now().AddDate(0, 0, -30),
		IsActive:           true,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}, nil
}

type cashFlowService struct {
	db    *sql.DB
	redis cache.RedisClient
}

func NewCashFlowService(db *sql.DB, redis cache.RedisClient) domain.CashFlowService {
	return &cashFlowService{
		db:    db,
		redis: redis,
	}
}

func (c *cashFlowService) RecordTransaction(entityType string, entityID uuid.UUID, txType domain.CashFlowType, amount float64, description, reference string, createdBy *uuid.UUID) error {
	// TODO: Store transaction record in database
	// TODO: Add event publishing mechanism if needed
	return nil
}

func (c *cashFlowService) GetEntityCashFlow(entityType string, entityID uuid.UUID, limit int) ([]*domain.CashFlowRecord, error) {
	return []*domain.CashFlowRecord{}, nil
}

func (c *cashFlowService) GetCurrentBalance(entityType string, entityID uuid.UUID) (float64, error) {
	return 50000.0, nil // Mock balance
}
