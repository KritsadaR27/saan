package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"order/internal/domain"
	"order/internal/infrastructure/database"
)

// AuditRepository implements the OrderAuditRepository interface using PostgreSQL
type AuditRepository struct {
	conn *database.Connection
}

// NewAuditRepository creates a new PostgreSQL audit repository
func NewAuditRepository(conn *database.Connection) domain.OrderAuditRepository {
	return &AuditRepository{conn: conn}
}

// Create creates a new audit log entry
func (r *AuditRepository) Create(ctx context.Context, auditLog *domain.OrderAuditLog) error {
	query := `
		INSERT INTO order_audit_logs (id, order_id, user_id, action, details, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	// Convert details map to JSON
	detailsJSON, err := marshalDetails(auditLog.Details)
	if err != nil {
		return fmt.Errorf("failed to marshal details: %w", err)
	}

	_, err = r.conn.DB.ExecContext(ctx, query,
		auditLog.ID, auditLog.OrderID, auditLog.UserID, auditLog.Action, detailsJSON, auditLog.Timestamp,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// GetByOrderID retrieves all audit logs for an order
func (r *AuditRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.OrderAuditLog, error) {
	query := `
		SELECT id, order_id, user_id, action, details, timestamp
		FROM order_audit_logs
		WHERE order_id = $1
		ORDER BY timestamp DESC
	`

	rows, err := r.conn.DB.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var auditLogs []*domain.OrderAuditLog
	for rows.Next() {
		auditLog, err := r.scanAuditLog(rows)
		if err != nil {
			return nil, err
		}
		auditLogs = append(auditLogs, auditLog)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate audit logs: %w", err)
	}

	return auditLogs, nil
}

// GetByUserID retrieves audit logs by user ID
func (r *AuditRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.OrderAuditLog, error) {
	query := `
		SELECT id, order_id, user_id, action, details, timestamp
		FROM order_audit_logs
		WHERE user_id = $1
		ORDER BY timestamp DESC
	`

	rows, err := r.conn.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs by user ID: %w", err)
	}
	defer rows.Close()

	var auditLogs []*domain.OrderAuditLog
	for rows.Next() {
		auditLog, err := r.scanAuditLog(rows)
		if err != nil {
			return nil, err
		}
		auditLogs = append(auditLogs, auditLog)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate audit logs: %w", err)
	}

	return auditLogs, nil
}

// GetByAction retrieves audit logs by action type
func (r *AuditRepository) GetByAction(ctx context.Context, action domain.AuditAction) ([]*domain.OrderAuditLog, error) {
	query := `
		SELECT id, order_id, user_id, action, details, timestamp
		FROM order_audit_logs
		WHERE action = $1
		ORDER BY timestamp DESC
	`

	rows, err := r.conn.DB.QueryContext(ctx, query, action)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs by action: %w", err)
	}
	defer rows.Close()

	var auditLogs []*domain.OrderAuditLog
	for rows.Next() {
		auditLog, err := r.scanAuditLog(rows)
		if err != nil {
			return nil, err
		}
		auditLogs = append(auditLogs, auditLog)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate audit logs: %w", err)
	}

	return auditLogs, nil
}

// List retrieves audit logs with pagination
func (r *AuditRepository) List(ctx context.Context, limit, offset int) ([]*domain.OrderAuditLog, error) {
	query := `
		SELECT id, order_id, user_id, action, details, timestamp
		FROM order_audit_logs
		ORDER BY timestamp DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.conn.DB.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var auditLogs []*domain.OrderAuditLog
	for rows.Next() {
		auditLog, err := r.scanAuditLog(rows)
		if err != nil {
			return nil, err
		}
		auditLogs = append(auditLogs, auditLog)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate audit logs: %w", err)
	}

	return auditLogs, nil
}

// Helper methods

func (r *AuditRepository) scanAuditLog(rows *sql.Rows) (*domain.OrderAuditLog, error) {
	var auditLog domain.OrderAuditLog
	var detailsJSON []byte

	err := rows.Scan(
		&auditLog.ID,
		&auditLog.OrderID,
		&auditLog.UserID,
		&auditLog.Action,
		&detailsJSON,
		&auditLog.Timestamp,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan audit log: %w", err)
	}

	// Unmarshal details JSON
	if len(detailsJSON) > 0 {
		details, err := unmarshalDetails(detailsJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal details: %w", err)
		}
		auditLog.Details = details
	}

	return &auditLog, nil
}

func marshalDetails(details map[string]interface{}) ([]byte, error) {
	if details == nil {
		return nil, nil
	}
	return []byte(fmt.Sprintf("%v", details)), nil
}

func unmarshalDetails(data []byte) (map[string]interface{}, error) {
	if len(data) == 0 {
		return nil, nil
	}
	// For simplicity, we'll store as string. In production, you'd use proper JSON marshaling
	return map[string]interface{}{"raw": string(data)}, nil
}
