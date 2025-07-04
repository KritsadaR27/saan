package repository
package repository

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/saan/order-service/internal/domain"
)

// PostgresEventRepository implements the EventRepository interface
type PostgresEventRepository struct {
	db *sqlx.DB
}

// NewPostgresEventRepository creates a new PostgresEventRepository
func NewPostgresEventRepository(db *sqlx.DB) *PostgresEventRepository {
	return &PostgresEventRepository{db: db}
}

// Create stores a new event in the outbox table
func (r *PostgresEventRepository) Create(ctx context.Context, event *domain.OrderEvent) error {
	query := `
		INSERT INTO events_outbox (id, order_id, event_type, payload, status, created_at, retry_count)
		VALUES (:id, :order_id, :event_type, :payload, :status, :created_at, :retry_count)
	`
	
	_, err := r.db.NamedExecContext(ctx, query, event)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}
	
	return nil
}

// GetPendingEvents retrieves pending events from the outbox table
func (r *PostgresEventRepository) GetPendingEvents(ctx context.Context, limit int) ([]*domain.OrderEvent, error) {
	query := `
		SELECT id, order_id, event_type, payload, status, created_at, sent_at, retry_count
		FROM events_outbox
		WHERE status IN ('pending', 'failed')
		ORDER BY created_at ASC
		LIMIT $1
	`
	
	var events []*domain.OrderEvent
	err := r.db.SelectContext(ctx, &events, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending events: %w", err)
	}
	
	return events, nil
}

// Update updates an existing event in the outbox table
func (r *PostgresEventRepository) Update(ctx context.Context, event *domain.OrderEvent) error {
	query := `
		UPDATE events_outbox
		SET status = :status, sent_at = :sent_at, retry_count = :retry_count
		WHERE id = :id
	`
	
	_, err := r.db.NamedExecContext(ctx, query, event)
	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}
	
	return nil
}

// Delete removes an event from the outbox table
func (r *PostgresEventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM events_outbox WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}
	
	return nil
}

// JSONBPayload is a custom type for handling JSONB in PostgreSQL
type JSONBPayload map[string]interface{}

// Value implements the driver.Valuer interface for database/sql
func (j JSONBPayload) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for database/sql
func (j *JSONBPayload) Scan(value interface{}) error {
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
		return fmt.Errorf("cannot scan %T into JSONBPayload", value)
	}
}
