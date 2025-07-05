package repositories

import (
	"database/sql"

	"finance/internal/domain"

	"github.com/google/uuid"
)

type cashFlowRepository struct {
	db *sql.DB
}

func NewCashFlowRepository(db *sql.DB) domain.CashFlowRepository {
	return &cashFlowRepository{
		db: db,
	}
}

func (r *cashFlowRepository) Create(record *domain.CashFlowRecord) error {
	query := `
		INSERT INTO cash_flow_records (
			id, entity_type, entity_id, transaction_type, amount, description,
			reference, running_balance, created_by, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)`

	_, err := r.db.Exec(query,
		record.ID,
		record.EntityType,
		record.EntityID,
		record.TransactionType,
		record.Amount,
		record.Description,
		record.Reference,
		record.RunningBalance,
		record.CreatedBy,
		record.CreatedAt,
	)

	return err
}

func (r *cashFlowRepository) GetByEntity(entityType string, entityID uuid.UUID, limit int) ([]*domain.CashFlowRecord, error) {
	query := `
		SELECT 
			id, entity_type, entity_id, transaction_type, amount, description,
			reference, running_balance, created_by, created_at
		FROM cash_flow_records 
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC
		LIMIT $3`

	rows, err := r.db.Query(query, entityType, entityID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*domain.CashFlowRecord
	for rows.Next() {
		record := &domain.CashFlowRecord{}
		err := rows.Scan(
			&record.ID,
			&record.EntityType,
			&record.EntityID,
			&record.TransactionType,
			&record.Amount,
			&record.Description,
			&record.Reference,
			&record.RunningBalance,
			&record.CreatedBy,
			&record.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

func (r *cashFlowRepository) GetCurrentBalance(entityType string, entityID uuid.UUID) (float64, error) {
	query := `
		SELECT running_balance 
		FROM cash_flow_records 
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC
		LIMIT 1`

	var balance float64
	err := r.db.QueryRow(query, entityType, entityID).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0.0, nil // No records means zero balance
		}
		return 0, err
	}

	return balance, nil
}

func (r *cashFlowRepository) GetByID(id uuid.UUID) (*domain.CashFlowRecord, error) {
	query := `
		SELECT 
			id, entity_type, entity_id, transaction_type, amount, description,
			reference, running_balance, created_by, created_at
		FROM cash_flow_records 
		WHERE id = $1`

	record := &domain.CashFlowRecord{}
	err := r.db.QueryRow(query, id).Scan(
		&record.ID,
		&record.EntityType,
		&record.EntityID,
		&record.TransactionType,
		&record.Amount,
		&record.Description,
		&record.Reference,
		&record.RunningBalance,
		&record.CreatedBy,
		&record.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrCashFlowNotFound
		}
		return nil, err
	}

	return record, nil
}

func (r *cashFlowRepository) GetBalanceHistory(entityType string, entityID uuid.UUID, limit int) ([]*domain.CashFlowRecord, error) {
	query := `
		SELECT 
			id, entity_type, entity_id, transaction_type, amount, description,
			reference, running_balance, created_by, created_at
		FROM cash_flow_records 
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC
		LIMIT $3`

	rows, err := r.db.Query(query, entityType, entityID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*domain.CashFlowRecord
	for rows.Next() {
		record := &domain.CashFlowRecord{}
		err := rows.Scan(
			&record.ID,
			&record.EntityType,
			&record.EntityID,
			&record.TransactionType,
			&record.Amount,
			&record.Description,
			&record.Reference,
			&record.RunningBalance,
			&record.CreatedBy,
			&record.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

// CalculateRunningBalance calculates the running balance by considering the current balance
// and the new transaction amount and type
func (r *cashFlowRepository) CalculateRunningBalance(entityType string, entityID uuid.UUID, transactionType domain.CashFlowType, amount float64) (float64, error) {
	currentBalance, err := r.GetCurrentBalance(entityType, entityID)
	if err != nil {
		return 0, err
	}

	switch transactionType {
	case domain.CashInflow:
		return currentBalance + amount, nil
	case domain.CashOutflow:
		return currentBalance - amount, nil
	default:
		return currentBalance, domain.ErrInvalidAmount
	}
}

func (r *cashFlowRepository) GetTotalInflows(entityType string, entityID uuid.UUID) (float64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0) 
		FROM cash_flow_records 
		WHERE entity_type = $1 AND entity_id = $2 AND transaction_type = 'inflow'`

	var total float64
	err := r.db.QueryRow(query, entityType, entityID).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (r *cashFlowRepository) GetTotalOutflows(entityType string, entityID uuid.UUID) (float64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0) 
		FROM cash_flow_records 
		WHERE entity_type = $1 AND entity_id = $2 AND transaction_type = 'outflow'`

	var total float64
	err := r.db.QueryRow(query, entityType, entityID).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}
