package repositories

import (
	"database/sql"

	"finance/internal/domain"

	"github.com/google/uuid"
)

type expenseRepository struct {
	db *sql.DB
}

func NewExpenseRepository(db *sql.DB) domain.ExpenseRepository {
	return &expenseRepository{
		db: db,
	}
}

func (r *expenseRepository) Create(expense *domain.ExpenseEntry) error {
	query := `
		INSERT INTO expense_entries (
			id, summary_id, category, description, amount, receipt, entered_by, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)`

	_, err := r.db.Exec(query,
		expense.ID,
		expense.SummaryID,
		expense.Category,
		expense.Description,
		expense.Amount,
		expense.Receipt,
		expense.EnteredBy,
		expense.CreatedAt,
	)

	return err
}

func (r *expenseRepository) GetBySummaryID(summaryID uuid.UUID) ([]*domain.ExpenseEntry, error) {
	query := `
		SELECT 
			id, summary_id, category, description, amount, receipt, entered_by, created_at
		FROM expense_entries 
		WHERE summary_id = $1
		ORDER BY created_at ASC`

	rows, err := r.db.Query(query, summaryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []*domain.ExpenseEntry
	for rows.Next() {
		expense := &domain.ExpenseEntry{}
		err := rows.Scan(
			&expense.ID,
			&expense.SummaryID,
			&expense.Category,
			&expense.Description,
			&expense.Amount,
			&expense.Receipt,
			&expense.EnteredBy,
			&expense.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}

	return expenses, nil
}

func (r *expenseRepository) GetByID(id uuid.UUID) (*domain.ExpenseEntry, error) {
	query := `
		SELECT 
			id, summary_id, category, description, amount, receipt, entered_by, created_at
		FROM expense_entries 
		WHERE id = $1`

	expense := &domain.ExpenseEntry{}
	err := r.db.QueryRow(query, id).Scan(
		&expense.ID,
		&expense.SummaryID,
		&expense.Category,
		&expense.Description,
		&expense.Amount,
		&expense.Receipt,
		&expense.EnteredBy,
		&expense.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrExpenseNotFound
		}
		return nil, err
	}

	return expense, nil
}

func (r *expenseRepository) Update(expense *domain.ExpenseEntry) error {
	query := `
		UPDATE expense_entries 
		SET category = $2,
			description = $3,
			amount = $4,
			receipt = $5
		WHERE id = $1`

	result, err := r.db.Exec(query,
		expense.ID,
		expense.Category,
		expense.Description,
		expense.Amount,
		expense.Receipt,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrExpenseNotFound
	}

	return nil
}

func (r *expenseRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM expense_entries WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrExpenseNotFound
	}

	return nil
}

func (r *expenseRepository) GetTotalBySummaryID(summaryID uuid.UUID) (float64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0) 
		FROM expense_entries 
		WHERE summary_id = $1`

	var total float64
	err := r.db.QueryRow(query, summaryID).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (r *expenseRepository) GetByCategoryAndSummaryID(summaryID uuid.UUID, category string) ([]*domain.ExpenseEntry, error) {
	query := `
		SELECT 
			id, summary_id, category, description, amount, receipt, entered_by, created_at
		FROM expense_entries 
		WHERE summary_id = $1 AND category = $2
		ORDER BY created_at ASC`

	rows, err := r.db.Query(query, summaryID, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []*domain.ExpenseEntry
	for rows.Next() {
		expense := &domain.ExpenseEntry{}
		err := rows.Scan(
			&expense.ID,
			&expense.SummaryID,
			&expense.Category,
			&expense.Description,
			&expense.Amount,
			&expense.Receipt,
			&expense.EnteredBy,
			&expense.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}

	return expenses, nil
}
