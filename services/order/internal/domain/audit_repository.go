package domain

import (
	"context"

	"github.com/google/uuid"
)

// AuditRepository defines the interface for audit log operations
type AuditRepository interface {
	// Create creates a new audit log entry
	Create(ctx context.Context, log *AuditLog) error

	// GetByOrderID retrieves all audit logs for an order
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*AuditLog, error)
}
