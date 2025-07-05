package events

import (
	"context"
)

// Publisher defines the interface for event publishing
type Publisher interface {
	// Order events
	PublishOrderCreated(ctx context.Context, orderID, customerID string, orderData map[string]interface{}) error
	PublishOrderUpdated(ctx context.Context, orderID, customerID string, changes map[string]interface{}) error
	PublishOrderCancelled(ctx context.Context, orderID, customerID string, reason string) error
	PublishOrderCompleted(ctx context.Context, orderID, customerID string) error
	
	// Payment events
	PublishPaymentProcessed(ctx context.Context, orderID, paymentID string, amount float64, currency string) error
	PublishPaymentFailed(ctx context.Context, orderID, paymentID string, amount float64, currency string, reason string) error
	
	// Inventory events
	PublishInventoryReserved(ctx context.Context, orderID, productID string, quantity int, reservationID string) error
	PublishInventoryReleased(ctx context.Context, orderID, productID string, quantity int, reservationID string) error
	
	// Notification events
	PublishNotificationSent(ctx context.Context, orderID, customerID string, notificationType, channel string, data map[string]interface{}) error
	
	// Lifecycle
	Close() error
}
