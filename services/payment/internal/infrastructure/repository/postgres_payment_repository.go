package repository

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"payment/internal/domain/entity"
	"payment/internal/domain/repository"
)

// PostgresPaymentRepository implements PaymentRepository using PostgreSQL
type PostgresPaymentRepository struct {
	db *sqlx.DB
}

// NewPostgresPaymentRepository creates a new PostgreSQL payment repository
func NewPostgresPaymentRepository(db *sqlx.DB) repository.PaymentRepository {
	return &PostgresPaymentRepository{db: db}
}

// JSONB type for handling PostgreSQL JSONB
type JSONB map[string]interface{}

// Value implements driver.Valuer interface for JSONB
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements sql.Scanner interface for JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	default:
		return fmt.Errorf("cannot scan %T into JSONB", value)
	}
}

// Create creates a new payment transaction
func (r *PostgresPaymentRepository) Create(ctx context.Context, payment *entity.PaymentTransaction) error {
	query := `
		INSERT INTO payment_transactions (
			id, order_id, customer_id, payment_method, payment_channel, payment_timing,
			amount, currency, status, paid_at, loyverse_receipt_id, loyverse_payment_type,
			assigned_store_id, metadata, created_at, updated_at, created_by, updated_by
		) VALUES (
			:id, :order_id, :customer_id, :payment_method, :payment_channel, :payment_timing,
			:amount, :currency, :status, :paid_at, :loyverse_receipt_id, :loyverse_payment_type,
			:assigned_store_id, :metadata, :created_at, :updated_at, :created_by, :updated_by
		)`

	_, err := r.db.NamedExecContext(ctx, query, payment)
	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}

	return nil
}

// GetByID retrieves a payment by ID
func (r *PostgresPaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.PaymentTransaction, error) {
	query := `
		SELECT id, order_id, customer_id, payment_method, payment_channel, payment_timing,
			   amount, currency, status, paid_at, loyverse_receipt_id, loyverse_payment_type,
			   assigned_store_id, metadata, created_at, updated_at, created_by, updated_by
		FROM payment_transactions 
		WHERE id = $1`

	var payment entity.PaymentTransaction
	err := r.db.GetContext(ctx, &payment, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return &payment, nil
}

// Update updates a payment transaction
func (r *PostgresPaymentRepository) Update(ctx context.Context, payment *entity.PaymentTransaction) error {
	query := `
		UPDATE payment_transactions SET
			payment_method = :payment_method,
			payment_channel = :payment_channel,
			payment_timing = :payment_timing,
			amount = :amount,
			currency = :currency,
			status = :status,
			paid_at = :paid_at,
			loyverse_receipt_id = :loyverse_receipt_id,
			loyverse_payment_type = :loyverse_payment_type,
			assigned_store_id = :assigned_store_id,
			metadata = :metadata,
			updated_at = :updated_at,
			updated_by = :updated_by
		WHERE id = :id`

	_, err := r.db.NamedExecContext(ctx, query, payment)
	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	return nil
}

// Delete deletes a payment transaction
func (r *PostgresPaymentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM payment_transactions WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete payment: %w", err)
	}
	return nil
}

// GetByStoreID retrieves payments for a specific store
func (r *PostgresPaymentRepository) GetByStoreID(ctx context.Context, storeID string, filters repository.PaymentFilters) ([]*entity.PaymentTransaction, error) {
	query := `
		SELECT id, order_id, customer_id, payment_method, payment_channel, payment_timing,
			   amount, currency, status, paid_at, loyverse_receipt_id, loyverse_payment_type,
			   assigned_store_id, metadata, created_at, updated_at, created_by, updated_by
		FROM payment_transactions 
		WHERE assigned_store_id = $1`

	args := []interface{}{storeID}
	argIndex := 2

	// Add filters
	query, args, argIndex = r.applyFilters(query, filters, args, argIndex)

	// Add ordering and pagination
	sortBy := "created_at"
	if filters.SortBy != "" {
		sortBy = filters.SortBy
	}
	sortOrder := "DESC"
	if filters.SortOrder != "" {
		sortOrder = filters.SortOrder
	}

	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filters.Offset)
	}

	var payments []*entity.PaymentTransaction
	err := r.db.SelectContext(ctx, &payments, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments by store ID: %w", err)
	}

	return payments, nil
}

// GetStoreAnalytics retrieves analytics for a specific store
func (r *PostgresPaymentRepository) GetStoreAnalytics(ctx context.Context, storeID string, dateFrom, dateTo time.Time) (*repository.StorePaymentAnalytics, error) {
	// Main analytics query
	analyticsQuery := `
		SELECT 
			COUNT(*) as total_transactions,
			COALESCE(SUM(amount), 0) as total_amount,
			COALESCE(AVG(amount), 0) as avg_amount,
			currency
		FROM payment_transactions 
		WHERE assigned_store_id = $1 
		  AND created_at >= $2 
		  AND created_at <= $3
		  AND status = 'completed'
		GROUP BY currency`

	var analyticsRow struct {
		TotalTransactions int     `db:"total_transactions"`
		TotalAmount       float64 `db:"total_amount"`
		AvgAmount         float64 `db:"avg_amount"`
		Currency          string  `db:"currency"`
	}

	err := r.db.GetContext(ctx, &analyticsRow, analyticsQuery, storeID, dateFrom, dateTo)
	if err != nil {
		return nil, fmt.Errorf("failed to get store analytics: %w", err)
	}

	analytics := &repository.StorePaymentAnalytics{
		StoreID:           storeID,
		TotalTransactions: analyticsRow.TotalTransactions,
		TotalAmount:       analyticsRow.TotalAmount,
		AvgAmount:         analyticsRow.AvgAmount,
		Currency:          analyticsRow.Currency,
		DateFrom:          dateFrom,
		DateTo:            dateTo,
	}

	// Get payment method stats
	methodStatsQuery := `
		SELECT 
			payment_method,
			COUNT(*) as count,
			SUM(amount) as total_amount
		FROM payment_transactions 
		WHERE assigned_store_id = $1 
		  AND created_at >= $2 
		  AND created_at <= $3
		  AND status = 'completed'
		GROUP BY payment_method`

	var methodRows []struct {
		PaymentMethod string  `db:"payment_method"`
		Count         int     `db:"count"`
		TotalAmount   float64 `db:"total_amount"`
	}

	err = r.db.SelectContext(ctx, &methodRows, methodStatsQuery, storeID, dateFrom, dateTo)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment method stats: %w", err)
	}

	methodStats := make([]repository.PaymentMethodStat, len(methodRows))
	for i, row := range methodRows {
		methodStats[i] = repository.PaymentMethodStat{
			Method:           entity.PaymentMethod(row.PaymentMethod),
			Count:            row.Count,
			TotalAmount:      row.TotalAmount,
			PercentageCount:  float64(row.Count) / float64(analytics.TotalTransactions) * 100,
			PercentageAmount: row.TotalAmount / analytics.TotalAmount * 100,
		}
	}
	analytics.PaymentMethodStats = methodStats

	// Get daily stats
	dailyStatsQuery := `
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as count,
			SUM(amount) as amount
		FROM payment_transactions 
		WHERE assigned_store_id = $1 
		  AND created_at >= $2 
		  AND created_at <= $3
		  AND status = 'completed'
		GROUP BY DATE(created_at)
		ORDER BY date`

	var dailyRows []struct {
		Date   time.Time `db:"date"`
		Count  int       `db:"count"`
		Amount float64   `db:"amount"`
	}

	err = r.db.SelectContext(ctx, &dailyRows, dailyStatsQuery, storeID, dateFrom, dateTo)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily stats: %w", err)
	}

	dailyStats := make([]repository.DailyPaymentStat, len(dailyRows))
	for i, row := range dailyRows {
		dailyStats[i] = repository.DailyPaymentStat{
			Date:   row.Date,
			Count:  row.Count,
			Amount: row.Amount,
		}
	}
	analytics.DailyStats = dailyStats

	return analytics, nil
}

// GetByCustomerID retrieves payments for a specific customer
func (r *PostgresPaymentRepository) GetByCustomerID(ctx context.Context, customerID uuid.UUID, filters repository.PaymentFilters) ([]*entity.PaymentTransaction, error) {
	query := `
		SELECT id, order_id, customer_id, payment_method, payment_channel, payment_timing,
			   amount, currency, status, paid_at, loyverse_receipt_id, loyverse_payment_type,
			   assigned_store_id, metadata, created_at, updated_at, created_by, updated_by
		FROM payment_transactions 
		WHERE customer_id = $1`

	args := []interface{}{customerID}
	argIndex := 2

	// Add filters
	query, args, argIndex = r.applyFilters(query, filters, args, argIndex)

	// Add ordering and pagination
	sortBy := "created_at"
	if filters.SortBy != "" {
		sortBy = filters.SortBy
	}
	sortOrder := "DESC"
	if filters.SortOrder != "" {
		sortOrder = filters.SortOrder
	}

	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filters.Offset)
	}

	var payments []*entity.PaymentTransaction
	err := r.db.SelectContext(ctx, &payments, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments by customer ID: %w", err)
	}

	return payments, nil
}

// GetCustomerPaymentHistory retrieves payment history for a customer
func (r *PostgresPaymentRepository) GetCustomerPaymentHistory(ctx context.Context, customerID uuid.UUID, limit int) ([]*entity.PaymentTransaction, error) {
	query := `
		SELECT id, order_id, customer_id, payment_method, payment_channel, payment_timing,
			   amount, currency, status, paid_at, loyverse_receipt_id, loyverse_payment_type,
			   assigned_store_id, metadata, created_at, updated_at, created_by, updated_by
		FROM payment_transactions 
		WHERE customer_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	var payments []*entity.PaymentTransaction
	err := r.db.SelectContext(ctx, &payments, query, customerID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer payment history: %w", err)
	}

	return payments, nil
}

// GetByOrderID retrieves payments for a specific order
func (r *PostgresPaymentRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*entity.PaymentTransaction, error) {
	query := `
		SELECT id, order_id, customer_id, payment_method, payment_channel, payment_timing,
			   amount, currency, status, paid_at, loyverse_receipt_id, loyverse_payment_type,
			   assigned_store_id, metadata, created_at, updated_at, created_by, updated_by
		FROM payment_transactions 
		WHERE order_id = $1
		ORDER BY created_at ASC`

	var payments []*entity.PaymentTransaction
	err := r.db.SelectContext(ctx, &payments, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments by order ID: %w", err)
	}

	return payments, nil
}

// GetOrderPaymentSummary retrieves payment summary for an order
func (r *PostgresPaymentRepository) GetOrderPaymentSummary(ctx context.Context, orderID uuid.UUID) (*repository.OrderPaymentSummary, error) {
	query := `
		SELECT 
			order_id,
			COUNT(*) as transaction_count,
			SUM(amount) as total_amount,
			SUM(CASE WHEN status = 'completed' THEN amount ELSE 0 END) as paid_amount,
			SUM(CASE WHEN status = 'pending' THEN amount ELSE 0 END) as pending_amount,
			SUM(CASE WHEN status = 'refunded' THEN amount ELSE 0 END) as refunded_amount,
			currency,
			MAX(CASE WHEN status = 'completed' THEN paid_at END) as last_payment_at,
			CASE 
				WHEN SUM(CASE WHEN status = 'completed' THEN amount ELSE 0 END) >= SUM(amount) THEN 'fully_paid'
				WHEN SUM(CASE WHEN status = 'completed' THEN amount ELSE 0 END) > 0 THEN 'partially_paid'
				ELSE 'unpaid'
			END as payment_status,
			ARRAY_AGG(DISTINCT payment_method) as payment_methods
		FROM payment_transactions
		WHERE order_id = $1
		GROUP BY order_id, currency`

	var summary repository.OrderPaymentSummary
	err := r.db.GetContext(ctx, &summary, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order payment summary: %w", err)
	}

	return &summary, nil
}

// GetByLoyverseReceiptID retrieves payment by Loyverse receipt ID
func (r *PostgresPaymentRepository) GetByLoyverseReceiptID(ctx context.Context, receiptID string) (*entity.PaymentTransaction, error) {
	query := `
		SELECT id, order_id, customer_id, payment_method, payment_channel, payment_timing,
			   amount, currency, status, paid_at, loyverse_receipt_id, loyverse_payment_type,
			   assigned_store_id, metadata, created_at, updated_at, created_by, updated_by
		FROM payment_transactions 
		WHERE loyverse_receipt_id = $1`

	var payment entity.PaymentTransaction
	err := r.db.GetContext(ctx, &payment, query, receiptID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment by Loyverse receipt ID: %w", err)
	}

	return &payment, nil
}

// GetPendingPayments retrieves pending payments
func (r *PostgresPaymentRepository) GetPendingPayments(ctx context.Context, limit int) ([]*entity.PaymentTransaction, error) {
	query := `
		SELECT id, order_id, customer_id, payment_method, payment_channel, payment_timing,
			   amount, currency, status, paid_at, loyverse_receipt_id, loyverse_payment_type,
			   assigned_store_id, metadata, created_at, updated_at, created_by, updated_by
		FROM payment_transactions 
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1`

	var payments []*entity.PaymentTransaction
	err := r.db.SelectContext(ctx, &payments, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending payments: %w", err)
	}

	return payments, nil
}

// GetPaymentsByDateRange retrieves payments within a date range
func (r *PostgresPaymentRepository) GetPaymentsByDateRange(ctx context.Context, dateFrom, dateTo time.Time, filters repository.PaymentFilters) ([]*entity.PaymentTransaction, error) {
	query := `
		SELECT id, order_id, customer_id, payment_method, payment_channel, payment_timing,
			   amount, currency, status, paid_at, loyverse_receipt_id, loyverse_payment_type,
			   assigned_store_id, metadata, created_at, updated_at, created_by, updated_by
		FROM payment_transactions 
		WHERE created_at >= $1 AND created_at <= $2`

	args := []interface{}{dateFrom, dateTo}
	argIndex := 3

	// Add filters
	query, args, argIndex = r.applyFilters(query, filters, args, argIndex)

	// Add ordering and pagination
	sortBy := "created_at"
	if filters.SortBy != "" {
		sortBy = filters.SortBy
	}
	sortOrder := "DESC"
	if filters.SortOrder != "" {
		sortOrder = filters.SortOrder
	}

	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filters.Offset)
	}

	var payments []*entity.PaymentTransaction
	err := r.db.SelectContext(ctx, &payments, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments by date range: %w", err)
	}

	return payments, nil
}

// GetPaymentsByChannel retrieves payments by channel
func (r *PostgresPaymentRepository) GetPaymentsByChannel(ctx context.Context, channel entity.PaymentChannel, filters repository.PaymentFilters) ([]*entity.PaymentTransaction, error) {
	query := `
		SELECT id, order_id, customer_id, payment_method, payment_channel, payment_timing,
			   amount, currency, status, paid_at, loyverse_receipt_id, loyverse_payment_type,
			   assigned_store_id, metadata, created_at, updated_at, created_by, updated_by
		FROM payment_transactions 
		WHERE payment_channel = $1`

	args := []interface{}{string(channel)}
	argIndex := 2

	// Add filters
	query, args, argIndex = r.applyFilters(query, filters, args, argIndex)

	// Add ordering and pagination
	sortBy := "created_at"
	if filters.SortBy != "" {
		sortBy = filters.SortBy
	}
	sortOrder := "DESC"
	if filters.SortOrder != "" {
		sortOrder = filters.SortOrder
	}

	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filters.Offset)
	}

	var payments []*entity.PaymentTransaction
	err := r.db.SelectContext(ctx, &payments, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments by channel: %w", err)
	}

	return payments, nil
}

// GetPaymentsByMethod retrieves payments by method
func (r *PostgresPaymentRepository) GetPaymentsByMethod(ctx context.Context, method entity.PaymentMethod, filters repository.PaymentFilters) ([]*entity.PaymentTransaction, error) {
	query := `
		SELECT id, order_id, customer_id, payment_method, payment_channel, payment_timing,
			   amount, currency, status, paid_at, loyverse_receipt_id, loyverse_payment_type,
			   assigned_store_id, metadata, created_at, updated_at, created_by, updated_by
		FROM payment_transactions 
		WHERE payment_method = $1`

	args := []interface{}{string(method)}
	argIndex := 2

	// Add filters
	query, args, argIndex = r.applyFilters(query, filters, args, argIndex)

	// Add ordering and pagination
	sortBy := "created_at"
	if filters.SortBy != "" {
		sortBy = filters.SortBy
	}
	sortOrder := "DESC"
	if filters.SortOrder != "" {
		sortOrder = filters.SortOrder
	}

	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filters.Offset)
	}

	var payments []*entity.PaymentTransaction
	err := r.db.SelectContext(ctx, &payments, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments by method: %w", err)
	}

	return payments, nil
}

// CreateBatch creates multiple payment transactions in a batch
func (r *PostgresPaymentRepository) CreateBatch(ctx context.Context, payments []*entity.PaymentTransaction) error {
	if len(payments) == 0 {
		return nil
	}

	query := `
		INSERT INTO payment_transactions (
			id, order_id, customer_id, payment_method, payment_channel, payment_timing,
			amount, currency, status, paid_at, loyverse_receipt_id, loyverse_payment_type,
			assigned_store_id, metadata, created_at, updated_at, created_by, updated_by
		) VALUES (
			:id, :order_id, :customer_id, :payment_method, :payment_channel, :payment_timing,
			:amount, :currency, :status, :paid_at, :loyverse_receipt_id, :loyverse_payment_type,
			:assigned_store_id, :metadata, :created_at, :updated_at, :created_by, :updated_by
		)`

	_, err := r.db.NamedExecContext(ctx, query, payments)
	if err != nil {
		return fmt.Errorf("failed to create payments batch: %w", err)
	}

	return nil
}

// UpdateStatus updates payment status
func (r *PostgresPaymentRepository) UpdateStatus(ctx context.Context, paymentID uuid.UUID, status entity.PaymentStatus) error {
	query := `
		UPDATE payment_transactions 
		SET status = $1, updated_at = NOW()
		WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, string(status), paymentID)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	return nil
}

// UpdateStatusBatch updates multiple payment statuses
func (r *PostgresPaymentRepository) UpdateStatusBatch(ctx context.Context, paymentIDs []uuid.UUID, status entity.PaymentStatus) error {
	if len(paymentIDs) == 0 {
		return nil
	}

	// Create placeholders for the IN clause
	placeholders := make([]string, len(paymentIDs))
	args := make([]interface{}, len(paymentIDs)+1)
	args[0] = string(status)
	
	for i, id := range paymentIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}

	query := fmt.Sprintf(`
		UPDATE payment_transactions 
		SET status = $1, updated_at = NOW()
		WHERE id IN (%s)`, strings.Join(placeholders, ","))

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update payment statuses batch: %w", err)
	}

	return nil
}

// Helper method to apply filters to queries
func (r *PostgresPaymentRepository) applyFilters(query string, filters repository.PaymentFilters, args []interface{}, argIndex int) (string, []interface{}, int) {
	conditions := []string{}

	if filters.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, string(*filters.Status))
		argIndex++
	}

	if filters.PaymentMethod != nil {
		conditions = append(conditions, fmt.Sprintf("payment_method = $%d", argIndex))
		args = append(args, string(*filters.PaymentMethod))
		argIndex++
	}

	if filters.PaymentChannel != nil {
		conditions = append(conditions, fmt.Sprintf("payment_channel = $%d", argIndex))
		args = append(args, string(*filters.PaymentChannel))
		argIndex++
	}

	if filters.PaymentTiming != nil {
		conditions = append(conditions, fmt.Sprintf("payment_timing = $%d", argIndex))
		args = append(args, string(*filters.PaymentTiming))
		argIndex++
	}

	if filters.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *filters.DateFrom)
		argIndex++
	}

	if filters.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *filters.DateTo)
		argIndex++
	}

	if filters.MinAmount != nil {
		conditions = append(conditions, fmt.Sprintf("amount >= $%d", argIndex))
		args = append(args, *filters.MinAmount)
		argIndex++
	}

	if filters.MaxAmount != nil {
		conditions = append(conditions, fmt.Sprintf("amount <= $%d", argIndex))
		args = append(args, *filters.MaxAmount)
		argIndex++
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	return query, args, argIndex
}

// Additional method implementations would continue here...
// GetByCustomerID, GetByOrderID, GetOrderPaymentSummary, etc.
// Following the same pattern as above
