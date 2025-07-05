package application

import (
	"time"

	"finance/internal/domain"
	"finance/internal/infrastructure/cache"
	"finance/internal/infrastructure/database/repositories"

	"github.com/google/uuid"
)

type financeService struct {
	repos *repositories.Repositories
	redis cache.RedisClient
}

func NewFinanceService(repos *repositories.Repositories, redis cache.RedisClient) domain.FinanceService {
	return &financeService{
		repos: repos,
		redis: redis,
	}
}

func (f *financeService) ProcessEndOfDay(date time.Time, branchID, vehicleID *uuid.UUID, sales, codCollections float64) (*domain.DailyCashSummary, error) {
	// Check if summary already exists
	existing, err := f.repos.CashSummary.GetByDateAndEntity(date, branchID, vehicleID)
	if err != nil && err != domain.ErrCashSummaryNotFound {
		return nil, err
	}
	if existing != nil {
		return existing, nil // Return existing summary
	}

	// Calculate allocations using allocation service
	allocationService := NewAllocationService(f.repos, f.redis)
	allocations, err := allocationService.CalculateAllocations(sales, branchID, vehicleID)
	if err != nil {
		return nil, err
	}

	// Create daily summary
	summary := &domain.DailyCashSummary{
		ID:                   uuid.New(),
		BusinessDate:         date,
		BranchID:             branchID,
		VehicleID:            vehicleID,
		TotalSales:           sales,
		CODCollections:       codCollections,
		ProfitAllocation:     allocations[domain.ProfitAccount],
		OwnerPayAllocation:   allocations[domain.OwnerPayAccount],
		TaxAllocation:        allocations[domain.TaxAccount],
		AvailableForExpenses: allocations[domain.OperatingAccount],
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	err = f.repos.CashSummary.Create(summary)
	if err != nil {
		return nil, err
	}

	return summary, nil
}

func (f *financeService) AddExpenseEntry(summaryID uuid.UUID, category, description string, amount float64, enteredBy uuid.UUID) error {
	// Validate the summary exists
	_, err := f.repos.CashSummary.GetByID(summaryID)
	if err != nil {
		return err
	}

	// Create expense entry
	expense := &domain.ExpenseEntry{
		ID:          uuid.New(),
		SummaryID:   summaryID,
		Category:    category,
		Description: description,
		Amount:      amount,
		EnteredBy:   enteredBy,
		CreatedAt:   time.Now(),
	}

	err = f.repos.Expense.Create(expense)
	if err != nil {
		return err
	}

	// Update summary with new expense total
	summary, err := f.repos.CashSummary.GetByID(summaryID)
	if err != nil {
		return err
	}

	// Get total expenses and update summary
	totalExpenses, err := f.repos.Expense.GetTotalBySummaryID(summaryID)
	if err != nil {
		return err
	}

	summary.ManualExpenses = totalExpenses
	return f.repos.CashSummary.Update(summary)
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

	err := f.repos.Transfer.CreateBatch(batch)
	if err != nil {
		return nil, err
	}

	// Create individual transfers
	for _, transfer := range transfers {
		transfer.BatchID = &batch.ID
		transfer.CreatedAt = time.Now()
		transfer.UpdatedAt = time.Now()
		if transfer.ID == uuid.Nil {
			transfer.ID = uuid.New()
		}

		err = f.repos.Transfer.CreateTransfer(transfer)
		if err != nil {
			return nil, err
		}
	}

	return batch, nil
}

func (f *financeService) ExecuteTransferBatch(batchID uuid.UUID) error {
	// Get batch
	batch, err := f.repos.Transfer.GetBatchByID(batchID)
	if err != nil {
		return err
	}

	if batch.Status != "pending" {
		return domain.ErrTransferInProgress
	}

	// Update batch status to processing
	err = f.repos.Transfer.UpdateBatchStatus(batchID, "processing")
	if err != nil {
		return err
	}

	// Get all transfers in batch
	transfers, err := f.repos.Transfer.GetTransfersByBatch(batchID)
	if err != nil {
		return err
	}

	// Execute each transfer
	failedTransfers := 0
	for _, transfer := range transfers {
		err := f.executeIndividualTransfer(transfer)
		if err != nil {
			failedTransfers++
			f.repos.Transfer.UpdateTransferStatus(transfer.ID, "failed")
		} else {
			f.repos.Transfer.UpdateTransferStatus(transfer.ID, "completed")
		}
	}

	// Update batch status based on results
	var finalStatus string
	if failedTransfers == 0 {
		finalStatus = "completed"
	} else if failedTransfers == len(transfers) {
		finalStatus = "failed"
	} else {
		finalStatus = "partial"
	}

	return f.repos.Transfer.UpdateBatchStatus(batchID, finalStatus)
}

func (f *financeService) executeIndividualTransfer(transfer *domain.CashTransfer) error {
	// In a real implementation, this would integrate with banking APIs
	// For now, we'll simulate the transfer
	time.Sleep(100 * time.Millisecond) // Simulate processing time
	return nil
}

func (f *financeService) GetCashStatus() (map[string]interface{}, error) {
	// Get current balances for all entities
	// This would typically aggregate across all branches and vehicles
	
	status := map[string]interface{}{
		"total_cash":        150000.00,
		"profit_account":    45000.00,
		"owner_pay_account": 30000.00,
		"tax_account":       22500.00,
		"operational_cash":  52500.00,
		"last_updated":      time.Now(),
	}

	return status, nil
}

func (f *financeService) ReconcileCash(summaryID uuid.UUID, actualCash float64, reconciledBy uuid.UUID) error {
	// Get the summary
	summary, err := f.repos.CashSummary.GetByID(summaryID)
	if err != nil {
		return err
	}

	if summary.Reconciled {
		return domain.ErrCannotModifyReconciled
	}

	// Update closing cash with actual amount
	summary.ClosingCash = actualCash

	// Update the summary
	err = f.repos.CashSummary.Update(summary)
	if err != nil {
		return err
	}

	// Mark as reconciled
	return f.repos.CashSummary.UpdateReconciliation(summaryID, reconciledBy)
}

func generateBatchReference() string {
	return "BATCH_" + time.Now().Format("20060102_150405")
}
