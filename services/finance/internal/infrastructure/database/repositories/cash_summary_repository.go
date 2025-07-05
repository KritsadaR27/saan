package repositories

import (
	"database/sql"
	"time"

	"finance/internal/domain"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type cashSummaryRepository struct {
	db *sql.DB
}

func NewCashSummaryRepository(db *sql.DB) domain.CashSummaryRepository {
	return &cashSummaryRepository{
		db: db,
	}
}

func (r *cashSummaryRepository) Create(summary *domain.DailyCashSummary) error {
	query := `
		INSERT INTO daily_cash_summaries (
			id, business_date, branch_id, vehicle_id, opening_cash, total_sales, 
			cod_collections, profit_allocation, owner_pay_allocation, tax_allocation,
			available_for_expenses, manual_expenses, supplier_transfers, other_transfers,
			closing_cash, reconciled, reconciled_by_user_id, reconciled_at, notes,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21
		)`

	_, err := r.db.Exec(query,
		summary.ID,
		summary.BusinessDate,
		summary.BranchID,
		summary.VehicleID,
		summary.OpeningCash,
		summary.TotalSales,
		summary.CODCollections,
		summary.ProfitAllocation,
		summary.OwnerPayAllocation,
		summary.TaxAllocation,
		summary.AvailableForExpenses,
		summary.ManualExpenses,
		summary.SupplierTransfers,
		summary.OtherTransfers,
		summary.ClosingCash,
		summary.Reconciled,
		summary.ReconciledByUserID,
		summary.ReconciledAt,
		summary.Notes,
		summary.CreatedAt,
		summary.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				return domain.ErrCashSummaryExists
			}
		}
		return err
	}

	return nil
}

func (r *cashSummaryRepository) GetByID(id uuid.UUID) (*domain.DailyCashSummary, error) {
	query := `
		SELECT 
			id, business_date, branch_id, vehicle_id, opening_cash, total_sales,
			cod_collections, profit_allocation, owner_pay_allocation, tax_allocation,
			available_for_expenses, manual_expenses, supplier_transfers, other_transfers,
			closing_cash, reconciled, reconciled_by_user_id, reconciled_at, notes,
			created_at, updated_at
		FROM daily_cash_summaries 
		WHERE id = $1`

	summary := &domain.DailyCashSummary{}
	err := r.db.QueryRow(query, id).Scan(
		&summary.ID,
		&summary.BusinessDate,
		&summary.BranchID,
		&summary.VehicleID,
		&summary.OpeningCash,
		&summary.TotalSales,
		&summary.CODCollections,
		&summary.ProfitAllocation,
		&summary.OwnerPayAllocation,
		&summary.TaxAllocation,
		&summary.AvailableForExpenses,
		&summary.ManualExpenses,
		&summary.SupplierTransfers,
		&summary.OtherTransfers,
		&summary.ClosingCash,
		&summary.Reconciled,
		&summary.ReconciledByUserID,
		&summary.ReconciledAt,
		&summary.Notes,
		&summary.CreatedAt,
		&summary.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrCashSummaryNotFound
		}
		return nil, err
	}

	return summary, nil
}

func (r *cashSummaryRepository) GetByDateAndEntity(date time.Time, branchID, vehicleID *uuid.UUID) (*domain.DailyCashSummary, error) {
	query := `
		SELECT 
			id, business_date, branch_id, vehicle_id, opening_cash, total_sales,
			cod_collections, profit_allocation, owner_pay_allocation, tax_allocation,
			available_for_expenses, manual_expenses, supplier_transfers, other_transfers,
			closing_cash, reconciled, reconciled_by_user_id, reconciled_at, notes,
			created_at, updated_at
		FROM daily_cash_summaries 
		WHERE business_date = $1 
		AND ($2::uuid IS NULL OR branch_id = $2) 
		AND ($3::uuid IS NULL OR vehicle_id = $3)`

	summary := &domain.DailyCashSummary{}
	err := r.db.QueryRow(query, date, branchID, vehicleID).Scan(
		&summary.ID,
		&summary.BusinessDate,
		&summary.BranchID,
		&summary.VehicleID,
		&summary.OpeningCash,
		&summary.TotalSales,
		&summary.CODCollections,
		&summary.ProfitAllocation,
		&summary.OwnerPayAllocation,
		&summary.TaxAllocation,
		&summary.AvailableForExpenses,
		&summary.ManualExpenses,
		&summary.SupplierTransfers,
		&summary.OtherTransfers,
		&summary.ClosingCash,
		&summary.Reconciled,
		&summary.ReconciledByUserID,
		&summary.ReconciledAt,
		&summary.Notes,
		&summary.CreatedAt,
		&summary.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrCashSummaryNotFound
		}
		return nil, err
	}

	return summary, nil
}

func (r *cashSummaryRepository) UpdateReconciliation(id uuid.UUID, reconciledBy uuid.UUID) error {
	query := `
		UPDATE daily_cash_summaries 
		SET reconciled = true, 
			reconciled_by_user_id = $2, 
			reconciled_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND reconciled = false`

	result, err := r.db.Exec(query, id, reconciledBy)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrCashSummaryNotFound
	}

	return nil
}

func (r *cashSummaryRepository) Update(summary *domain.DailyCashSummary) error {
	query := `
		UPDATE daily_cash_summaries 
		SET opening_cash = $2, 
			total_sales = $3,
			cod_collections = $4,
			profit_allocation = $5,
			owner_pay_allocation = $6,
			tax_allocation = $7,
			available_for_expenses = $8,
			manual_expenses = $9,
			supplier_transfers = $10,
			other_transfers = $11,
			closing_cash = $12,
			notes = $13,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND reconciled = false`

	result, err := r.db.Exec(query,
		summary.ID,
		summary.OpeningCash,
		summary.TotalSales,
		summary.CODCollections,
		summary.ProfitAllocation,
		summary.OwnerPayAllocation,
		summary.TaxAllocation,
		summary.AvailableForExpenses,
		summary.ManualExpenses,
		summary.SupplierTransfers,
		summary.OtherTransfers,
		summary.ClosingCash,
		summary.Notes,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrCannotModifyReconciled
	}

	return nil
}

func (r *cashSummaryRepository) GetByDateRange(startDate, endDate time.Time, branchID, vehicleID *uuid.UUID) ([]*domain.DailyCashSummary, error) {
	query := `
		SELECT 
			id, business_date, branch_id, vehicle_id, opening_cash, total_sales,
			cod_collections, profit_allocation, owner_pay_allocation, tax_allocation,
			available_for_expenses, manual_expenses, supplier_transfers, other_transfers,
			closing_cash, reconciled, reconciled_by_user_id, reconciled_at, notes,
			created_at, updated_at
		FROM daily_cash_summaries 
		WHERE business_date >= $1 AND business_date <= $2
		AND ($3::uuid IS NULL OR branch_id = $3) 
		AND ($4::uuid IS NULL OR vehicle_id = $4)
		ORDER BY business_date DESC`

	rows, err := r.db.Query(query, startDate, endDate, branchID, vehicleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []*domain.DailyCashSummary
	for rows.Next() {
		summary := &domain.DailyCashSummary{}
		err := rows.Scan(
			&summary.ID,
			&summary.BusinessDate,
			&summary.BranchID,
			&summary.VehicleID,
			&summary.OpeningCash,
			&summary.TotalSales,
			&summary.CODCollections,
			&summary.ProfitAllocation,
			&summary.OwnerPayAllocation,
			&summary.TaxAllocation,
			&summary.AvailableForExpenses,
			&summary.ManualExpenses,
			&summary.SupplierTransfers,
			&summary.OtherTransfers,
			&summary.ClosingCash,
			&summary.Reconciled,
			&summary.ReconciledByUserID,
			&summary.ReconciledAt,
			&summary.Notes,
			&summary.CreatedAt,
			&summary.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}
