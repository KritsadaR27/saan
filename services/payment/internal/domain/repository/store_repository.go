package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"payment/internal/domain/entity"
)

// LoyverseStoreRepository defines the interface for Loyverse store operations
type LoyverseStoreRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, store *entity.LoyverseStore) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.LoyverseStore, error)
	GetByStoreCode(ctx context.Context, storeCode string) (*entity.LoyverseStore, error)
	Update(ctx context.Context, store *entity.LoyverseStore) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Query operations
	GetAllStores(ctx context.Context) ([]*entity.LoyverseStore, error)
	GetActiveStores(ctx context.Context) ([]*entity.LoyverseStore, error)
	GetStoresByRegion(ctx context.Context, region string) ([]*entity.LoyverseStore, error)
	GetStoresByManager(ctx context.Context, managerID uuid.UUID) ([]*entity.LoyverseStore, error)

	// Store assignment operations
	GetAvailableStoresForAssignment(ctx context.Context) ([]*entity.LoyverseStore, error)
	GetStoreWorkload(ctx context.Context, storeCode string, dateFrom, dateTo time.Time) (*StoreWorkload, error)
	UpdateStoreMetrics(ctx context.Context, storeCode string, metrics *StoreMetrics) error
}

// PaymentDeliveryContextRepository defines the interface for payment delivery context operations
type PaymentDeliveryContextRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, context *entity.PaymentDeliveryContext) error
	GetByPaymentID(ctx context.Context, paymentID uuid.UUID) (*entity.PaymentDeliveryContext, error)
	Update(ctx context.Context, context *entity.PaymentDeliveryContext) error
	Delete(ctx context.Context, paymentID uuid.UUID) error

	// Query operations
	GetByDeliveryID(ctx context.Context, deliveryID uuid.UUID) ([]*entity.PaymentDeliveryContext, error)
	GetByDriverID(ctx context.Context, driverID uuid.UUID) ([]*entity.PaymentDeliveryContext, error)
	GetContextsByDateRange(ctx context.Context, dateFrom, dateTo time.Time) ([]*entity.PaymentDeliveryContext, error)
	GetCODContexts(ctx context.Context, filters ContextFilters) ([]*entity.PaymentDeliveryContext, error)
}

// EventRepository defines the interface for event operations
type EventRepository interface {
	PublishPaymentEvent(ctx context.Context, event *PaymentEvent) error
	PublishPaymentStatusChanged(ctx context.Context, paymentID uuid.UUID, oldStatus, newStatus entity.PaymentStatus) error
	PublishLoyversePaymentCreated(ctx context.Context, paymentID uuid.UUID, receiptID string) error
	PublishCODPaymentCollected(ctx context.Context, paymentID uuid.UUID, deliveryID uuid.UUID) error
}

// Supporting types

// StoreWorkload represents the workload metrics for a store
type StoreWorkload struct {
	StoreCode           string    `json:"store_code"`
	PendingOrders       int       `json:"pending_orders"`
	ProcessingOrders    int       `json:"processing_orders"`
	TotalOrdersToday    int       `json:"total_orders_today"`
	AvgProcessingTime   float64   `json:"avg_processing_time_minutes"`
	CurrentCapacity     float64   `json:"current_capacity_percentage"`
	LastUpdated         time.Time `json:"last_updated"`
}

// StoreMetrics represents performance metrics for a store
type StoreMetrics struct {
	OrdersProcessed     int       `json:"orders_processed"`
	AvgProcessingTime   float64   `json:"avg_processing_time"`
	CustomerSatisfaction float64   `json:"customer_satisfaction"`
	PaymentSuccessRate  float64   `json:"payment_success_rate"`
	LastUpdated         time.Time `json:"last_updated"`
}

// ContextFilters represents filters for delivery context queries
type ContextFilters struct {
	PaymentStatus   *entity.PaymentStatus
	DeliveryStatus  *string
	DriverID        *uuid.UUID
	CustomerID      *uuid.UUID
	StoreCode       *string
	DateFrom        *time.Time
	DateTo          *time.Time
	CODOnly         bool
	Limit           int
	Offset          int
}

// PaymentEvent represents a payment-related event
type PaymentEvent struct {
	ID          uuid.UUID              `json:"id"`
	EventType   string                 `json:"event_type"`
	PaymentID   uuid.UUID              `json:"payment_id"`
	OrderID     *uuid.UUID             `json:"order_id,omitempty"`
	CustomerID  *uuid.UUID             `json:"customer_id,omitempty"`
	Data        map[string]interface{} `json:"data"`
	OccurredAt  time.Time              `json:"occurred_at"`
	Source      string                 `json:"source"`
	Version     string                 `json:"version"`
}

// Event types constants
const (
	EventTypePaymentCreated           = "payment.created"
	EventTypePaymentStatusChanged     = "payment.status_changed"
	EventTypePaymentCompleted         = "payment.completed"
	EventTypePaymentFailed            = "payment.failed"
	EventTypePaymentRefunded          = "payment.refunded"
	EventTypeLoyversePaymentCreated   = "loyverse.payment_created"
	EventTypeCODPaymentCollected      = "cod.payment_collected"
	EventTypeStoreAssigned            = "store.assigned"
	EventTypeDeliveryContextCreated   = "delivery_context.created"
)
