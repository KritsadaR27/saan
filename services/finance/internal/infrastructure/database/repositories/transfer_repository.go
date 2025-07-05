package repositories

import (
	"database/sql"

	"finance/internal/domain"

	"github.com/google/uuid"
)

type transferRepository struct {
	db *sql.DB
}

func NewTransferRepository(db *sql.DB) domain.TransferRepository {
	return &transferRepository{
		db: db,
	}
}

func (r *transferRepository) CreateBatch(batch *domain.CashTransferBatch) error {
	query := `
		INSERT INTO cash_transfer_batches (
			id, batch_reference, branch_id, vehicle_id, total_amount, transfer_count,
			status, scheduled_at, processed_at, completed_at, authorized_by, notes,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)`

	_, err := r.db.Exec(query,
		batch.ID,
		batch.BatchReference,
		batch.BranchID,
		batch.VehicleID,
		batch.TotalAmount,
		batch.TransferCount,
		batch.Status,
		batch.ScheduledAt,
		batch.ProcessedAt,
		batch.CompletedAt,
		batch.AuthorizedBy,
		batch.Notes,
		batch.CreatedAt,
		batch.UpdatedAt,
	)

	return err
}

func (r *transferRepository) CreateTransfer(transfer *domain.CashTransfer) error {
	query := `
		INSERT INTO cash_transfers (
			id, batch_id, transfer_type, recipient_name, recipient_account, amount,
			currency, reference, description, status, bank_name, account_number,
			transaction_ref, scheduled_at, executed_at, confirmed_at, failure_reason,
			created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		)`

	_, err := r.db.Exec(query,
		transfer.ID,
		transfer.BatchID,
		transfer.TransferType,
		transfer.RecipientName,
		transfer.RecipientAccount,
		transfer.Amount,
		transfer.Currency,
		transfer.Reference,
		transfer.Description,
		transfer.Status,
		transfer.BankName,
		transfer.AccountNumber,
		transfer.TransactionRef,
		transfer.ScheduledAt,
		transfer.ExecutedAt,
		transfer.ConfirmedAt,
		transfer.FailureReason,
		transfer.CreatedBy,
		transfer.CreatedAt,
		transfer.UpdatedAt,
	)

	return err
}

func (r *transferRepository) GetBatchByID(id uuid.UUID) (*domain.CashTransferBatch, error) {
	query := `
		SELECT 
			id, batch_reference, branch_id, vehicle_id, total_amount, transfer_count,
			status, scheduled_at, processed_at, completed_at, authorized_by, notes,
			created_at, updated_at
		FROM cash_transfer_batches 
		WHERE id = $1`

	batch := &domain.CashTransferBatch{}
	err := r.db.QueryRow(query, id).Scan(
		&batch.ID,
		&batch.BatchReference,
		&batch.BranchID,
		&batch.VehicleID,
		&batch.TotalAmount,
		&batch.TransferCount,
		&batch.Status,
		&batch.ScheduledAt,
		&batch.ProcessedAt,
		&batch.CompletedAt,
		&batch.AuthorizedBy,
		&batch.Notes,
		&batch.CreatedAt,
		&batch.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrTransferBatchNotFound
		}
		return nil, err
	}

	return batch, nil
}

func (r *transferRepository) GetTransfersByBatch(batchID uuid.UUID) ([]*domain.CashTransfer, error) {
	query := `
		SELECT 
			id, batch_id, transfer_type, recipient_name, recipient_account, amount,
			currency, reference, description, status, bank_name, account_number,
			transaction_ref, scheduled_at, executed_at, confirmed_at, failure_reason,
			created_by, created_at, updated_at
		FROM cash_transfers 
		WHERE batch_id = $1
		ORDER BY created_at ASC`

	rows, err := r.db.Query(query, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transfers []*domain.CashTransfer
	for rows.Next() {
		transfer := &domain.CashTransfer{}
		err := rows.Scan(
			&transfer.ID,
			&transfer.BatchID,
			&transfer.TransferType,
			&transfer.RecipientName,
			&transfer.RecipientAccount,
			&transfer.Amount,
			&transfer.Currency,
			&transfer.Reference,
			&transfer.Description,
			&transfer.Status,
			&transfer.BankName,
			&transfer.AccountNumber,
			&transfer.TransactionRef,
			&transfer.ScheduledAt,
			&transfer.ExecutedAt,
			&transfer.ConfirmedAt,
			&transfer.FailureReason,
			&transfer.CreatedBy,
			&transfer.CreatedAt,
			&transfer.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		transfers = append(transfers, transfer)
	}

	return transfers, nil
}

func (r *transferRepository) UpdateTransferStatus(id uuid.UUID, status string) error {
	query := `
		UPDATE cash_transfers 
		SET status = $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	result, err := r.db.Exec(query, id, status)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrTransferNotFound
	}

	return nil
}

func (r *transferRepository) UpdateBatchStatus(id uuid.UUID, status string) error {
	query := `
		UPDATE cash_transfer_batches 
		SET status = $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	result, err := r.db.Exec(query, id, status)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrTransferBatchNotFound
	}

	return nil
}

func (r *transferRepository) GetTransferByID(id uuid.UUID) (*domain.CashTransfer, error) {
	query := `
		SELECT 
			id, batch_id, transfer_type, recipient_name, recipient_account, amount,
			currency, reference, description, status, bank_name, account_number,
			transaction_ref, scheduled_at, executed_at, confirmed_at, failure_reason,
			created_by, created_at, updated_at
		FROM cash_transfers 
		WHERE id = $1`

	transfer := &domain.CashTransfer{}
	err := r.db.QueryRow(query, id).Scan(
		&transfer.ID,
		&transfer.BatchID,
		&transfer.TransferType,
		&transfer.RecipientName,
		&transfer.RecipientAccount,
		&transfer.Amount,
		&transfer.Currency,
		&transfer.Reference,
		&transfer.Description,
		&transfer.Status,
		&transfer.BankName,
		&transfer.AccountNumber,
		&transfer.TransactionRef,
		&transfer.ScheduledAt,
		&transfer.ExecutedAt,
		&transfer.ConfirmedAt,
		&transfer.FailureReason,
		&transfer.CreatedBy,
		&transfer.CreatedAt,
		&transfer.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrTransferNotFound
		}
		return nil, err
	}

	return transfer, nil
}

func (r *transferRepository) GetPendingBatches() ([]*domain.CashTransferBatch, error) {
	query := `
		SELECT 
			id, batch_reference, branch_id, vehicle_id, total_amount, transfer_count,
			status, scheduled_at, processed_at, completed_at, authorized_by, notes,
			created_at, updated_at
		FROM cash_transfer_batches 
		WHERE status = 'pending'
		ORDER BY created_at ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var batches []*domain.CashTransferBatch
	for rows.Next() {
		batch := &domain.CashTransferBatch{}
		err := rows.Scan(
			&batch.ID,
			&batch.BatchReference,
			&batch.BranchID,
			&batch.VehicleID,
			&batch.TotalAmount,
			&batch.TransferCount,
			&batch.Status,
			&batch.ScheduledAt,
			&batch.ProcessedAt,
			&batch.CompletedAt,
			&batch.AuthorizedBy,
			&batch.Notes,
			&batch.CreatedAt,
			&batch.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		batches = append(batches, batch)
	}

	return batches, nil
}
