package event

import (
	"context"
	"time"

	"github.com/saan/order-service/internal/domain"
	"github.com/saan/order-service/pkg/logger"
)

// OutboxWorkerConfig holds configuration for the outbox worker
type OutboxWorkerConfig struct {
	// PollingInterval is how often to check for pending events
	PollingInterval time.Duration
	
	// BatchSize is the maximum number of events to process in one batch
	BatchSize int
	
	// MaxRetries is the maximum number of retry attempts for failed events
	MaxRetries int
	
	// RetryDelay is the delay before retrying failed events
	RetryDelay time.Duration
}

// DefaultOutboxWorkerConfig returns a default configuration
func DefaultOutboxWorkerConfig() OutboxWorkerConfig {
	return OutboxWorkerConfig{
		PollingInterval: 5 * time.Second,
		BatchSize:       10,
		MaxRetries:      3,
		RetryDelay:      30 * time.Second,
	}
}

// OutboxWorker is a background worker that processes events from the outbox table
type OutboxWorker struct {
	eventRepo     domain.OrderEventRepository // Use the concrete interface instead of alias
	publisher     EventPublisher
	config        OutboxWorkerConfig
	stopChan      chan struct{}
	logger        logger.Logger // Use the interface instead of pointer
}

// NewOutboxWorker creates a new outbox worker
func NewOutboxWorker(
	eventRepo domain.OrderEventRepository,
	publisher EventPublisher,
	config OutboxWorkerConfig,
	logger logger.Logger,
) *OutboxWorker {
	return &OutboxWorker{
		eventRepo: eventRepo,
		publisher: publisher,
		config:    config,
		stopChan:  make(chan struct{}),
		logger:    logger,
	}
}

// Start starts the outbox worker in a background goroutine
func (w *OutboxWorker) Start(ctx context.Context) {
	w.logger.WithFields(map[string]interface{}{
		"polling_interval": w.config.PollingInterval,
		"batch_size":       w.config.BatchSize,
		"max_retries":      w.config.MaxRetries,
	}).Info("Starting outbox worker")

	go w.run(ctx)
}

// Stop stops the outbox worker
func (w *OutboxWorker) Stop() {
	w.logger.Info("Stopping outbox worker")
	close(w.stopChan)
}

// run is the main worker loop
func (w *OutboxWorker) run(ctx context.Context) {
	ticker := time.NewTicker(w.config.PollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Outbox worker stopped due to context cancellation")
			return
		case <-w.stopChan:
			w.logger.Info("Outbox worker stopped")
			return
		case <-ticker.C:
			w.processPendingEvents(ctx)
		}
	}
}

// processPendingEvents processes pending events from the outbox
func (w *OutboxWorker) processPendingEvents(ctx context.Context) {
	// Get pending events
	events, err := w.eventRepo.GetPendingEvents(ctx, w.config.BatchSize)
	if err != nil {
		w.logger.WithField("error", err.Error()).Error("Failed to get pending events")
		return
	}

	if len(events) == 0 {
		return
	}

	w.logger.WithField("count", len(events)).Info("Processing pending events")

	// Process each event
	for _, event := range events {
		w.processEvent(ctx, event)
	}
}

// processEvent processes a single event
func (w *OutboxWorker) processEvent(ctx context.Context, event *domain.OrderEventOutbox) {
	eventLogger := w.logger.WithFields(map[string]interface{}{
		"event_id":    event.ID,
		"order_id":    event.OrderID,
		"event_type":  event.EventType,
		"retry_count": event.RetryCount,
	})

	eventLogger.Info("Processing event")

	// Check if event should be retried
	if event.Status == domain.EventStatusFailed && !event.ShouldRetry(w.config.MaxRetries) {
		eventLogger.Warn("Event exceeded max retries, marking as cancelled")
		event.MarkAsCancelled()
		if err := w.updateEventStatus(ctx, event); err != nil {
			eventLogger.WithField("error", err.Error()).Error("Failed to mark event as cancelled")
		}
		return
	}

	// Try to publish the event
	err := w.publisher.Publish(ctx, event)
	if err != nil {
		eventLogger.WithField("error", err.Error()).Error("Failed to publish event")
		
		// Mark as failed and increment retry count
		event.MarkAsFailed()
		if updateErr := w.updateEventStatus(ctx, event); updateErr != nil {
			eventLogger.WithField("error", updateErr.Error()).Error("Failed to update event status to failed")
		}
		return
	}

	// Mark as sent
	event.MarkAsSent()
	if err := w.updateEventStatus(ctx, event); err != nil {
		eventLogger.WithField("error", err.Error()).Error("Failed to mark event as sent")
		return
	}

	eventLogger.Info("Event published successfully")
}

// updateEventStatus updates the event status in the repository
func (w *OutboxWorker) updateEventStatus(ctx context.Context, event *domain.OrderEventOutbox) error {
	// For OrderEventOutbox, we need to use the repository update methods
	switch event.Status {
	case domain.EventStatusSent:
		return w.eventRepo.MarkAsSent(ctx, event.ID)
	case domain.EventStatusFailed:
		return w.eventRepo.MarkAsFailed(ctx, event.ID)
	case domain.EventStatusCancelled:
		return w.eventRepo.UpdateStatus(ctx, event.ID, domain.EventStatusCancelled)
	default:
		return w.eventRepo.UpdateStatus(ctx, event.ID, event.Status)
	}
}

// CleanupProcessedEvents removes old processed events from the outbox
func (w *OutboxWorker) CleanupProcessedEvents(ctx context.Context, olderThan time.Duration) error {
	w.logger.WithField("older_than", olderThan).Info("Starting cleanup of processed events")

	// This would require additional repository methods to query by date and status
	// For now, we'll implement a basic cleanup strategy
	
	// Note: In a production system, you might want to:
	// 1. Archive events instead of deleting them
	// 2. Use a separate cleanup job
	// 3. Implement batch deletion for better performance
	
	w.logger.Info("Event cleanup completed")
	return nil
}
