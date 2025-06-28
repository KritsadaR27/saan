package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/saan/order-service/internal/domain"
)

// PostgresAuditRepository implements the OrderAuditRepository interface
type PostgresAuditRepository struct {
	db *sqlx.DB
}

// NewPostgresAuditRepository creates a new PostgreSQL audit repository
func NewPostgresAuditRepository(db *sqlx.DB) domain.OrderAuditRepository {
	return &PostgresAuditRepository{db: db}
}

// Create creates a new audit log entry
func (r *PostgresAuditRepository) Create(ctx context.Context, auditLog *domain.OrderAuditLog) error {
	query := `
		INSERT INTO order_audit_log (id, order_id, user_id, action, details, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	
	// Convert details map to JSON
	var detailsJSON []byte
	var err error
	if auditLog.Details != nil {
		detailsJSON, err = json.Marshal(auditLog.Details)
		if err != nil {
			return fmt.Errorf("failed to marshal audit details: %w", err)
		}
	}
	
	_, err = r.db.ExecContext(ctx, query,
		auditLog.ID, auditLog.OrderID, auditLog.UserID, auditLog.Action,
		detailsJSON, auditLog.Timestamp,
	)
	
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}
	
	return nil
}

// GetByOrderID retrieves all audit logs for an order
func (r *PostgresAuditRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.OrderAuditLog, error) {
	query := `
		SELECT id, order_id, user_id, action, details, timestamp
		FROM order_audit_log
		WHERE order_id = $1
		ORDER BY timestamp DESC
	`
	
	var auditLogs []*domain.OrderAuditLog
	err := r.db.SelectContext(ctx, &auditLogs, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs by order ID: %w", err)
	}
	
	return auditLogs, nil
}

// GetByUserID retrieves audit logs by user ID
func (r *PostgresAuditRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.OrderAuditLog, error) {
	query := `
		SELECT id, order_id, user_id, action, details, timestamp
		FROM order_audit_log
		WHERE user_id = $1
		ORDER BY timestamp DESC
	`
	
	var auditLogs []*domain.OrderAuditLog
	err := r.db.SelectContext(ctx, &auditLogs, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs by user ID: %w", err)
	}
	
	return auditLogs, nil
}

// GetByAction retrieves audit logs by action type
func (r *PostgresAuditRepository) GetByAction(ctx context.Context, action domain.AuditAction) ([]*domain.OrderAuditLog, error) {
	query := `
		SELECT id, order_id, user_id, action, details, timestamp
		FROM order_audit_log
		WHERE action = $1
		ORDER BY timestamp DESC
	`
	
	var auditLogs []*domain.OrderAuditLog
	err := r.db.SelectContext(ctx, &auditLogs, query, action)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs by action: %w", err)
	}
	
	return auditLogs, nil
}

// List retrieves audit logs with pagination
func (r *PostgresAuditRepository) List(ctx context.Context, limit, offset int) ([]*domain.OrderAuditLog, error) {
	query := `
		SELECT id, order_id, user_id, action, details, timestamp
		FROM order_audit_log
		ORDER BY timestamp DESC
		LIMIT $1 OFFSET $2
	`
	
	var auditLogs []*domain.OrderAuditLog
	err := r.db.SelectContext(ctx, &auditLogs, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}
	
	return auditLogs, nil
}

// PostgresEventRepository implements the OrderEventRepository interface
type PostgresEventRepository struct {
	db *sqlx.DB
}

// NewPostgresEventRepository creates a new PostgreSQL event repository
func NewPostgresEventRepository(db *sqlx.DB) domain.OrderEventRepository {
	return &PostgresEventRepository{db: db}
}

// Create creates a new event in the outbox
func (r *PostgresEventRepository) Create(ctx context.Context, event *domain.OrderEventOutbox) error {
	query := `
		INSERT INTO order_events_outbox (id, order_id, event_type, payload, status, created_at, sent_at, retry_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	
	// Convert payload map to JSON
	var payloadJSON []byte
	var err error
	if event.Payload != nil {
		payloadJSON, err = json.Marshal(event.Payload)
		if err != nil {
			return fmt.Errorf("failed to marshal event payload: %w", err)
		}
	}
	
	_, err = r.db.ExecContext(ctx, query,
		event.ID, event.OrderID, event.EventType, payloadJSON,
		event.Status, event.CreatedAt, event.SentAt, event.RetryCount,
	)
	
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}
	
	return nil
}

// GetPendingEvents retrieves all pending events for processing
func (r *PostgresEventRepository) GetPendingEvents(ctx context.Context, limit int) ([]*domain.OrderEventOutbox, error) {
	query := `
		SELECT id, order_id, event_type, payload, status, created_at, sent_at, retry_count
		FROM order_events_outbox
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1
	`
	
	var events []*domain.OrderEventOutbox
	err := r.db.SelectContext(ctx, &events, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending events: %w", err)
	}
	
	return events, nil
}

// GetFailedEvents retrieves failed events that can be retried
func (r *PostgresEventRepository) GetFailedEvents(ctx context.Context, maxRetries int, limit int) ([]*domain.OrderEventOutbox, error) {
	query := `
		SELECT id, order_id, event_type, payload, status, created_at, sent_at, retry_count
		FROM order_events_outbox
		WHERE status = 'failed' AND retry_count < $1
		ORDER BY created_at ASC
		LIMIT $2
	`
	
	var events []*domain.OrderEventOutbox
	err := r.db.SelectContext(ctx, &events, query, maxRetries, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed events: %w", err)
	}
	
	return events, nil
}

// UpdateStatus updates the status of an event
func (r *PostgresEventRepository) UpdateStatus(ctx context.Context, eventID uuid.UUID, status domain.EventStatus) error {
	query := `
		UPDATE order_events_outbox
		SET status = $2
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(ctx, query, eventID, status)
	if err != nil {
		return fmt.Errorf("failed to update event status: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("event not found")
	}
	
	return nil
}

// MarkAsSent marks an event as successfully sent
func (r *PostgresEventRepository) MarkAsSent(ctx context.Context, eventID uuid.UUID) error {
	query := `
		UPDATE order_events_outbox
		SET status = 'sent', sent_at = NOW()
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(ctx, query, eventID)
	if err != nil {
		return fmt.Errorf("failed to mark event as sent: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("event not found")
	}
	
	return nil
}

// MarkAsFailed marks an event as failed and increments retry count
func (r *PostgresEventRepository) MarkAsFailed(ctx context.Context, eventID uuid.UUID) error {
	query := `
		UPDATE order_events_outbox
		SET status = 'failed', retry_count = retry_count + 1
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(ctx, query, eventID)
	if err != nil {
		return fmt.Errorf("failed to mark event as failed: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("event not found")
	}
	
	return nil
}

// GetByOrderID retrieves all events for an order
func (r *PostgresEventRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.OrderEventOutbox, error) {
	query := `
		SELECT id, order_id, event_type, payload, status, created_at, sent_at, retry_count
		FROM order_events_outbox
		WHERE order_id = $1
		ORDER BY created_at DESC
	`
	
	var events []*domain.OrderEventOutbox
	err := r.db.SelectContext(ctx, &events, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get events by order ID: %w", err)
	}
	
	return events, nil
}

// Delete removes old processed events (for cleanup)
func (r *PostgresEventRepository) Delete(ctx context.Context, eventID uuid.UUID) error {
	query := `DELETE FROM order_events_outbox WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, eventID)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("event not found")
	}
	
	return nil
}
