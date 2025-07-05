package events

import (
	"context"

	"github.com/google/uuid"
	"customer/internal/domain/entity"
)

// NoOpPublisher is a no-op implementation of Publisher for development/testing
type NoOpPublisher struct{}

// NewNoOpPublisher creates a new no-op publisher
func NewNoOpPublisher() Publisher {
	return &NoOpPublisher{}
}

// PublishCustomerCreated is a no-op implementation
func (p *NoOpPublisher) PublishCustomerCreated(ctx context.Context, customer *entity.Customer) error {
	return nil
}

// PublishCustomerUpdated is a no-op implementation
func (p *NoOpPublisher) PublishCustomerUpdated(ctx context.Context, customer *entity.Customer) error {
	return nil
}

// PublishCustomerDeleted is a no-op implementation
func (p *NoOpPublisher) PublishCustomerDeleted(ctx context.Context, customerID uuid.UUID) error {
	return nil
}

// PublishCustomerTierUpdated is a no-op implementation
func (p *NoOpPublisher) PublishCustomerTierUpdated(ctx context.Context, customerID uuid.UUID, oldTier, newTier entity.CustomerTier) error {
	return nil
}

// PublishLoyverseCustomerSynced is a no-op implementation  
func (p *NoOpPublisher) PublishLoyverseCustomerSynced(ctx context.Context, customerID uuid.UUID, loyverseID string) error {
	return nil
}

// PublishCustomerPointsUpdated is a no-op implementation
func (p *NoOpPublisher) PublishCustomerPointsUpdated(ctx context.Context, customerID uuid.UUID, pointsChange, totalPoints int, transactionType, description string) error {
	return nil
}

// PublishCustomerAddressAdded is a no-op implementation
func (p *NoOpPublisher) PublishCustomerAddressAdded(ctx context.Context, customerID, addressID uuid.UUID, address *entity.CustomerAddress) error {
	return nil
}

// PublishCustomerAddressUpdated is a no-op implementation
func (p *NoOpPublisher) PublishCustomerAddressUpdated(ctx context.Context, customerID, addressID uuid.UUID, address *entity.CustomerAddress) error {
	return nil
}

// PublishCustomerAddressDeleted is a no-op implementation
func (p *NoOpPublisher) PublishCustomerAddressDeleted(ctx context.Context, customerID, addressID uuid.UUID) error {
	return nil
}

// PublishLoyverseSynced is a no-op implementation
func (p *NoOpPublisher) PublishLoyverseSynced(ctx context.Context, customerID uuid.UUID, loyverseID string, syncStatus string) error {
	return nil
}

// PublishCustomerTierUpdatedWithReason is a no-op implementation
func (p *NoOpPublisher) PublishCustomerTierUpdatedWithReason(ctx context.Context, customerID uuid.UUID, oldTier, newTier entity.CustomerTier, reason string) error {
	return nil
}

// PublishLoyverseSyncedWithStatus is a no-op implementation
func (p *NoOpPublisher) PublishLoyverseSyncedWithStatus(ctx context.Context, customerID uuid.UUID, loyverseID string, syncStatus string) error {
	return nil
}

// Close is a no-op implementation
func (p *NoOpPublisher) Close() error {
	return nil
}
