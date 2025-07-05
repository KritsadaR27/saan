package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"order/internal/domain"
	"order/internal/infrastructure/database"
)

// EventRepository implements the OrderEventRepository interface using PostgreSQL
type EventRepository struct {
	conn *database.Connection
}

// NewEventRepository creates a new PostgreSQL event repository
func NewEventRepository(conn *database.Connection) domain.OrderEventRepository {
	return &EventRepository{conn: conn}
}

// Create creates a new event in the outbox
func (r *EventRepository) Create(ctx context.Context, event *domain.OrderEventOutbox) error {
	query := `
		INSERT INTO order_events (id, order_id, event_type, payload, status, created_at, sent_at, retry_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	// Convert payload to JSON
	payloadJSON, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	_, err = r.conn.DB.ExecContext(ctx, query,
		event.ID, event.OrderID, event.EventType, payloadJSON, event.Status,
		event.CreatedAt, event.SentAt, event.RetryCount,
	)

	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

// GetPendingEvents retrieves all pending events for processing
func (r *EventRepository) GetPendingEvents(ctx context.Context, limit int) ([]*domain.OrderEventOutbox, error) {
	query := `
		SELECT id, order_id, event_type, payload, status, created_at, sent_at, retry_count
		FROM order_events
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT $2
	`

	rows, err := r.conn.DB.QueryContext(ctx, query, domain.EventStatusPending, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending events: %w", err)
	}
	defer rows.Close()

	var events []*domain.OrderEventOutbox
	for rows.Next() {
		event, err := r.scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate pending events: %w", err)
	}

	return events, nil
}

// GetFailedEvents retrieves failed events that can be retried
func (r *EventRepository) GetFailedEvents(ctx context.Context, maxRetries int, limit int) ([]*domain.OrderEventOutbox, error) {
	query := `
		SELECT id, order_id, event_type, payload, status, created_at, sent_at, retry_count
		FROM order_events
		WHERE status = $1 AND retry_count < $2
		ORDER BY created_at ASC
		LIMIT $3
	`

	rows, err := r.conn.DB.QueryContext(ctx, query, domain.EventStatusFailed, maxRetries, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query failed events: %w", err)
	}
	defer rows.Close()

	var events []*domain.OrderEventOutbox
	for rows.Next() {
		event, err := r.scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate failed events: %w", err)
	}

	return events, nil
}

// UpdateStatus updates the status of an event
func (r *EventRepository) UpdateStatus(ctx context.Context, eventID uuid.UUID, status domain.EventStatus) error {
	query := `UPDATE order_events SET status = $1 WHERE id = $2`

	result, err := r.conn.DB.ExecContext(ctx, query, status, eventID)
	if err != nil {
		return fmt.Errorf("failed to update event status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrEventNotFound
	}

	return nil
}

// MarkAsSent marks an event as successfully sent
func (r *EventRepository) MarkAsSent(ctx context.Context, eventID uuid.UUID) error {
	query := `UPDATE order_events SET status = $1, sent_at = $2 WHERE id = $3`

	now := time.Now()
	result, err := r.conn.DB.ExecContext(ctx, query, domain.EventStatusSent, now, eventID)
	if err != nil {
		return fmt.Errorf("failed to mark event as sent: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrEventNotFound
	}

	return nil
}

// MarkAsFailed marks an event as failed and increments retry count
func (r *EventRepository) MarkAsFailed(ctx context.Context, eventID uuid.UUID) error {
	query := `UPDATE order_events SET status = $1, retry_count = retry_count + 1 WHERE id = $2`

	result, err := r.conn.DB.ExecContext(ctx, query, domain.EventStatusFailed, eventID)
	if err != nil {
		return fmt.Errorf("failed to mark event as failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrEventNotFound
	}

	return nil
}

// GetByOrderID retrieves all events for an order
func (r *EventRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.OrderEventOutbox, error) {
	query := `
		SELECT id, order_id, event_type, payload, status, created_at, sent_at, retry_count
		FROM order_events
		WHERE order_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.conn.DB.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query events by order ID: %w", err)
	}
	defer rows.Close()

	var events []*domain.OrderEventOutbox
	for rows.Next() {
		event, err := r.scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate events: %w", err)
	}

	return events, nil
}

// Delete removes old processed events (for cleanup)
func (r *EventRepository) Delete(ctx context.Context, eventID uuid.UUID) error {
	query := `DELETE FROM order_events WHERE id = $1`

	result, err := r.conn.DB.ExecContext(ctx, query, eventID)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrEventNotFound
	}

	return nil
}

// Helper methods

func (r *EventRepository) scanEvent(rows *sql.Rows) (*domain.OrderEventOutbox, error) {
	var event domain.OrderEventOutbox
	var payloadJSON []byte

	err := rows.Scan(
		&event.ID,
		&event.OrderID,
		&event.EventType,
		&payloadJSON,
		&event.Status,
		&event.CreatedAt,
		&event.SentAt,
		&event.RetryCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan event: %w", err)
	}

	// Unmarshal payload JSON
	if len(payloadJSON) > 0 {
		if err := json.Unmarshal(payloadJSON, &event.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}
	}

	return &event, nil
}
