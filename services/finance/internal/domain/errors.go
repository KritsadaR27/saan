package domain

import "errors"

// Finance service specific errors
var (
	ErrCashSummaryNotFound     = errors.New("daily cash summary not found")
	ErrCashSummaryExists       = errors.New("daily cash summary already exists for this date and entity")
	ErrAllocationRuleNotFound  = errors.New("allocation rule not found")
	ErrInvalidAllocationRule   = errors.New("invalid allocation rule")
	ErrTransferBatchNotFound   = errors.New("transfer batch not found")
	ErrTransferNotFound        = errors.New("transfer not found")
	ErrExpenseNotFound         = errors.New("expense entry not found")
	ErrCashFlowNotFound        = errors.New("cash flow record not found")
	ErrInvalidAmount           = errors.New("invalid amount")
	ErrInvalidDate             = errors.New("invalid date")
	ErrInvalidEntity           = errors.New("invalid entity specification")
	ErrReconciliationFailed    = errors.New("cash reconciliation failed")
	ErrInsufficientBalance     = errors.New("insufficient balance")
	ErrInvalidStatus           = errors.New("invalid status")
	ErrUnauthorized            = errors.New("unauthorized operation")
	ErrDatabaseConnection      = errors.New("database connection error")
	ErrTransactionFailed       = errors.New("database transaction failed")
)

// Validation errors
var (
	ErrMissingEntityID      = errors.New("entity ID is required")
	ErrMissingUserID        = errors.New("user ID is required")
	ErrMissingDescription   = errors.New("description is required")
	ErrMissingRecipient     = errors.New("recipient information is required")
	ErrInvalidPercentage    = errors.New("percentage must be between 0 and 100")
	ErrPercentageSum        = errors.New("allocation percentages must not exceed 100%")
	ErrInvalidCurrency      = errors.New("invalid currency code")
	ErrInvalidAccountNumber = errors.New("invalid account number format")
)

// Business rule errors
var (
	ErrCannotModifyReconciled = errors.New("cannot modify reconciled cash summary")
	ErrCannotDeleteActiveRule = errors.New("cannot delete active allocation rule")
	ErrDuplicateActiveRule    = errors.New("only one active rule allowed per entity")
	ErrTransferInProgress     = errors.New("transfer batch is already in progress")
	ErrInvalidDateRange       = errors.New("effective date range is invalid")
)
