package events

import (
	"context"

	"order/internal/domain"
	"github.com/sirupsen/logrus"
)

// DomainEventAdapter adapts our new event publisher to the domain EventPublisher interface
type DomainEventAdapter struct {
	publisher Publisher
	logger    *logrus.Logger
}

// NewDomainEventAdapter creates a new domain event adapter
func NewDomainEventAdapter(publisher Publisher, logger *logrus.Logger) domain.EventPublisher {
	return &DomainEventAdapter{
		publisher: publisher,
		logger:    logger,
	}
}

// PublishEvent adapts domain.OrderEvent to our new event system
func (a *DomainEventAdapter) PublishEvent(ctx context.Context, event *domain.OrderEvent) error {
	// Convert domain event to our event format and publish
	eventData := map[string]interface{}{
		"id":          event.ID.String(),
		"order_id":    event.OrderID.String(),
		"event_type":  string(event.EventType),
		"payload":     event.Payload,
		"status":      string(event.Status),
		"created_at":  event.CreatedAt,
		"retry_count": event.RetryCount,
	}

	// Map domain event types to our event types and publish accordingly
	switch event.EventType {
	case domain.EventOrderCreated:
		return a.publisher.PublishOrderCreated(ctx, event.OrderID.String(), "", eventData)
	case domain.EventOrderUpdated:
		return a.publisher.PublishOrderUpdated(ctx, event.OrderID.String(), "", eventData)
	case domain.EventOrderCancelled:
		return a.publisher.PublishOrderCancelled(ctx, event.OrderID.String(), "", "")
	case domain.EventOrderDelivered:
		return a.publisher.PublishOrderCompleted(ctx, event.OrderID.String(), "")
	default:
		// For unknown event types, use order updated as fallback
		a.logger.WithField("event_type", event.EventType).Warn("Unknown event type, using order updated as fallback")
		return a.publisher.PublishOrderUpdated(ctx, event.OrderID.String(), "", eventData)
	}
}
