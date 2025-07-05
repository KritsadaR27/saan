package events

import (
	"context"

	"github.com/google/uuid"
	"customer/internal/domain/entity"
	"customer/internal/domain/repository"
)

// Publisher extends the domain repository interface with additional event types
type Publisher interface {
	repository.EventPublisher // Embed the domain interface
	
	// Additional customer points events
	PublishCustomerPointsUpdated(ctx context.Context, customerID uuid.UUID, pointsChange, totalPoints int, transactionType, description string) error
	
	// Additional customer address events
	PublishCustomerAddressAdded(ctx context.Context, customerID, addressID uuid.UUID, address *entity.CustomerAddress) error
	PublishCustomerAddressUpdated(ctx context.Context, customerID, addressID uuid.UUID, address *entity.CustomerAddress) error
	PublishCustomerAddressDeleted(ctx context.Context, customerID, addressID uuid.UUID) error
	
	// Enhanced tier update with reason
	PublishCustomerTierUpdatedWithReason(ctx context.Context, customerID uuid.UUID, oldTier, newTier entity.CustomerTier, reason string) error
	
	// Enhanced sync with status
	PublishLoyverseSyncedWithStatus(ctx context.Context, customerID uuid.UUID, loyverseID string, syncStatus string) error
	
	// Lifecycle
	Close() error
}
