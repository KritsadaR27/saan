package domain

import (
	"time"
	"github.com/google/uuid"
)

// CashFlowType represents the type of cash flow
type CashFlowType string

const (
	CashInflow  CashFlowType = "inflow"
	CashOutflow CashFlowType = "outflow"
)

// AccountType represents different account types in Profit First system
type AccountType string

const (
	ProfitAccount    AccountType = "profit"
	OwnerPayAccount  AccountType = "owner_pay"
	TaxAccount       AccountType = "tax"
	OperatingAccount AccountType = "operating"
	RevenueAccount   AccountType = "revenue"
)

// DailyCashSummary represents end-of-day cash summary
type DailyCashSummary struct {
	ID                     uuid.UUID  `json:"id" db:"id"`
	BusinessDate           time.Time  `json:"business_date" db:"business_date"`
	BranchID               *uuid.UUID `json:"branch_id,omitempty" db:"branch_id"`
	VehicleID              *uuid.UUID `json:"vehicle_id,omitempty" db:"vehicle_id"`
	OpeningCash            float64    `json:"opening_cash" db:"opening_cash"`
	TotalSales             float64    `json:"total_sales" db:"total_sales"`
	CODCollections         float64    `json:"cod_collections" db:"cod_collections"`
	
	// Profit First allocations
	ProfitAllocation       float64    `json:"profit_allocation" db:"profit_allocation"`
	OwnerPayAllocation     float64    `json:"owner_pay_allocation" db:"owner_pay_allocation"`
	TaxAllocation          float64    `json:"tax_allocation" db:"tax_allocation"`
	AvailableForExpenses   float64    `json:"available_for_expenses" db:"available_for_expenses"`
	
	// Manual entries
	ManualExpenses         float64    `json:"manual_expenses" db:"manual_expenses"`
	SupplierTransfers      float64    `json:"supplier_transfers" db:"supplier_transfers"`
	OtherTransfers         float64    `json:"other_transfers" db:"other_transfers"`
	
	ClosingCash            float64    `json:"closing_cash" db:"closing_cash"`
	Reconciled             bool       `json:"reconciled" db:"reconciled"`
	ReconciledByUserID     *uuid.UUID `json:"reconciled_by_user_id,omitempty" db:"reconciled_by_user_id"`
	ReconciledAt           *time.Time `json:"reconciled_at,omitempty" db:"reconciled_at"`
	Notes                  string     `json:"notes" db:"notes"`
	CreatedAt              time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at" db:"updated_at"`
}

// ProfitAllocationRule represents configurable profit allocation rules
type ProfitAllocationRule struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	BranchID             *uuid.UUID `json:"branch_id,omitempty" db:"branch_id"`
	VehicleID            *uuid.UUID `json:"vehicle_id,omitempty" db:"vehicle_id"`
	ProfitPercentage     float64    `json:"profit_percentage" db:"profit_percentage"`
	OwnerPayPercentage   float64    `json:"owner_pay_percentage" db:"owner_pay_percentage"`
	TaxPercentage        float64    `json:"tax_percentage" db:"tax_percentage"`
	EffectiveFrom        time.Time  `json:"effective_from" db:"effective_from"`
	EffectiveTo          *time.Time `json:"effective_to,omitempty" db:"effective_to"`
	IsActive             bool       `json:"is_active" db:"is_active"`
	UpdatedByUserID      uuid.UUID  `json:"updated_by_user_id" db:"updated_by_user_id"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
}

// CashTransferBatch represents a batch of cash transfers
type CashTransferBatch struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	BatchReference  string     `json:"batch_reference" db:"batch_reference"`
	BranchID        *uuid.UUID `json:"branch_id,omitempty" db:"branch_id"`
	VehicleID       *uuid.UUID `json:"vehicle_id,omitempty" db:"vehicle_id"`
	TotalAmount     float64    `json:"total_amount" db:"total_amount"`
	TransferCount   int        `json:"transfer_count" db:"transfer_count"`
	Status          string     `json:"status" db:"status"` // pending, processing, completed, failed
	ScheduledAt     *time.Time `json:"scheduled_at,omitempty" db:"scheduled_at"`
	ProcessedAt     *time.Time `json:"processed_at,omitempty" db:"processed_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	AuthorizedBy    uuid.UUID  `json:"authorized_by" db:"authorized_by"`
	Notes           string     `json:"notes" db:"notes"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// CashTransfer represents individual cash transfer
type CashTransfer struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	BatchID          *uuid.UUID `json:"batch_id,omitempty" db:"batch_id"`
	TransferType     string     `json:"transfer_type" db:"transfer_type"` // supplier_payment, expense, central_transfer
	RecipientName    string     `json:"recipient_name" db:"recipient_name"`
	RecipientAccount string     `json:"recipient_account" db:"recipient_account"`
	Amount           float64    `json:"amount" db:"amount"`
	Currency         string     `json:"currency" db:"currency"`
	Reference        string     `json:"reference" db:"reference"`
	Description      string     `json:"description" db:"description"`
	Status           string     `json:"status" db:"status"` // pending, processing, completed, failed
	
	// Bank transfer details
	BankName         *string    `json:"bank_name,omitempty" db:"bank_name"`
	AccountNumber    *string    `json:"account_number,omitempty" db:"account_number"`
	TransactionRef   *string    `json:"transaction_ref,omitempty" db:"transaction_ref"`
	
	ScheduledAt      *time.Time `json:"scheduled_at,omitempty" db:"scheduled_at"`
	ExecutedAt       *time.Time `json:"executed_at,omitempty" db:"executed_at"`
	ConfirmedAt      *time.Time `json:"confirmed_at,omitempty" db:"confirmed_at"`
	FailureReason    *string    `json:"failure_reason,omitempty" db:"failure_reason"`
	CreatedBy        uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// ExpenseEntry represents manual expense entries
type ExpenseEntry struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	SummaryID   uuid.UUID  `json:"summary_id" db:"summary_id"`
	Category    string     `json:"category" db:"category"` // fuel, meals, utilities, maintenance, etc.
	Description string     `json:"description" db:"description"`
	Amount      float64    `json:"amount" db:"amount"`
	Receipt     *string    `json:"receipt,omitempty" db:"receipt"` // File path or URL
	EnteredBy   uuid.UUID  `json:"entered_by" db:"entered_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// CashFlowRecord represents real-time cash flow tracking
type CashFlowRecord struct {
	ID              uuid.UUID    `json:"id" db:"id"`
	EntityType      string       `json:"entity_type" db:"entity_type"` // branch, vehicle, central
	EntityID        uuid.UUID    `json:"entity_id" db:"entity_id"`
	TransactionType CashFlowType `json:"transaction_type" db:"transaction_type"`
	Amount          float64      `json:"amount" db:"amount"`
	Description     string       `json:"description" db:"description"`
	Reference       string       `json:"reference" db:"reference"`
	RunningBalance  float64      `json:"running_balance" db:"running_balance"`
	CreatedBy       *uuid.UUID   `json:"created_by,omitempty" db:"created_by"`
	CreatedAt       time.Time    `json:"created_at" db:"created_at"`
}

// Repository interfaces
type CashSummaryRepository interface {
	Create(summary *DailyCashSummary) error
	GetByID(id uuid.UUID) (*DailyCashSummary, error)
	GetByDateAndEntity(date time.Time, branchID, vehicleID *uuid.UUID) (*DailyCashSummary, error)
	UpdateReconciliation(id uuid.UUID, reconciledBy uuid.UUID) error
}

type AllocationRuleRepository interface {
	GetActiveRule(branchID, vehicleID *uuid.UUID) (*ProfitAllocationRule, error)
	Create(rule *ProfitAllocationRule) error
	Update(rule *ProfitAllocationRule) error
}

type TransferRepository interface {
	CreateBatch(batch *CashTransferBatch) error
	CreateTransfer(transfer *CashTransfer) error
	GetBatchByID(id uuid.UUID) (*CashTransferBatch, error)
	GetTransfersByBatch(batchID uuid.UUID) ([]*CashTransfer, error)
	UpdateTransferStatus(id uuid.UUID, status string) error
}

type ExpenseRepository interface {
	Create(expense *ExpenseEntry) error
	GetBySummaryID(summaryID uuid.UUID) ([]*ExpenseEntry, error)
}

type CashFlowRepository interface {
	Create(record *CashFlowRecord) error
	GetByEntity(entityType string, entityID uuid.UUID, limit int) ([]*CashFlowRecord, error)
	GetCurrentBalance(entityType string, entityID uuid.UUID) (float64, error)
}

// Service interfaces
type FinanceService interface {
	ProcessEndOfDay(date time.Time, branchID, vehicleID *uuid.UUID, sales, codCollections float64) (*DailyCashSummary, error)
	AddExpenseEntry(summaryID uuid.UUID, category, description string, amount float64, enteredBy uuid.UUID) error
	CreateTransferBatch(branchID, vehicleID *uuid.UUID, transfers []*CashTransfer, authorizedBy uuid.UUID) (*CashTransferBatch, error)
	ExecuteTransferBatch(batchID uuid.UUID) error
	GetCashStatus() (map[string]interface{}, error)
	ReconcileCash(summaryID uuid.UUID, actualCash float64, reconciledBy uuid.UUID) error
}

type AllocationService interface {
	CalculateAllocations(revenue float64, branchID, vehicleID *uuid.UUID) (map[AccountType]float64, error)
	UpdateAllocationRule(rule *ProfitAllocationRule) error
	GetCurrentRule(branchID, vehicleID *uuid.UUID) (*ProfitAllocationRule, error)
}

type CashFlowService interface {
	RecordTransaction(entityType string, entityID uuid.UUID, txType CashFlowType, amount float64, description, reference string, createdBy *uuid.UUID) error
	GetEntityCashFlow(entityType string, entityID uuid.UUID, limit int) ([]*CashFlowRecord, error)
	GetCurrentBalance(entityType string, entityID uuid.UUID) (float64, error)
}

type TransferService interface {
	CreateSupplierPayment(amount float64, supplierName, account string, createdBy uuid.UUID) (*CashTransfer, error)
	CreateExpenseTransfer(amount float64, description string, createdBy uuid.UUID) (*CashTransfer, error)
	ProcessBankTransfer(transferID uuid.UUID) error
}
