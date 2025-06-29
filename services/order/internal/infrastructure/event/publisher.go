package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/saan/order-service/internal/domain"
)

// EventPublisher defines the interface for publishing events
type EventPublisher interface {
	// PublishEvent publishes an event to the message broker
	PublishEvent(ctx context.Context, event *domain.OrderEvent) error
	
	// Close closes the publisher and releases resources
	Close() error
}

// MockEventPublisher implements EventPublisher for development/testing
type MockEventPublisher struct {
	enabled bool
}

// NewMockEventPublisher creates a new mock event publisher
func NewMockEventPublisher() *MockEventPublisher {
	return &MockEventPublisher{
		enabled: true,
	}
}

// PublishEvent logs the event instead of actually publishing
func (p *MockEventPublisher) PublishEvent(ctx context.Context, event *domain.OrderEvent) error {
	if !p.enabled {
		return nil
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	log.Printf("Mock Event Published: %s", string(eventJSON))
	return nil
}

// Close closes the mock publisher
func (p *MockEventPublisher) Close() error {
	p.enabled = false
	return nil
}
