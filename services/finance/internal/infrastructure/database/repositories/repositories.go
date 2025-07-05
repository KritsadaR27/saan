package repositories

import (
	"database/sql"

	"finance/internal/domain"
)

// Repositories struct holds all repository instances
type Repositories struct {
	CashSummary    domain.CashSummaryRepository
	AllocationRule domain.AllocationRuleRepository
	Transfer       domain.TransferRepository
	Expense        domain.ExpenseRepository
	CashFlow       domain.CashFlowRepository
}

// NewRepositories creates and returns all repository instances
func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		CashSummary:    NewCashSummaryRepository(db),
		AllocationRule: NewAllocationRuleRepository(db),
		Transfer:       NewTransferRepository(db),
		Expense:        NewExpenseRepository(db),
		CashFlow:       NewCashFlowRepository(db),
	}
}
