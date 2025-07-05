package events

import (
	"crypto/rand"
	"fmt"
	"time"
)

// Event types for order domain
const (
	OrderCreated      = "order.created"
	OrderUpdated      = "order.updated"
	OrderCancelled    = "order.cancelled"
	OrderPaid         = "order.paid"
	OrderShipped      = "order.shipped"
	OrderDelivered    = "order.delivered"
	OrderCompleted    = "order.completed"
	OrderRefunded     = "order.refunded"
	
	// Order item events
	OrderItemAdded    = "order.item_added"
	OrderItemUpdated  = "order.item_updated"
	OrderItemRemoved  = "order.item_removed"
	
	// Payment events
	PaymentProcessed  = "order.payment_processed"
	PaymentFailed     = "order.payment_failed"
	
	// Inventory events
	InventoryReserved = "order.inventory_reserved"
	InventoryReleased = "order.inventory_released"
	
	// Notification events
	NotificationSent  = "order.notification_sent"
)

// Topic definitions following SAAN standards
const (
	OrderEventsTopic        = "order-events"
	PaymentEventsTopic      = "payment-events"
	InventoryEventsTopic    = "inventory-events"
	NotificationEventsTopic = "notification-events"
)

// Base event structure
type BaseEvent struct {
	EventID     string    `json:"event_id"`
	EventType   string    `json:"event_type"`
	AggregateID string    `json:"aggregate_id"`
	Timestamp   time.Time `json:"timestamp"`
	Version     int       `json:"version"`
}

// OrderEvent represents order-related events
type OrderEvent struct {
	BaseEvent
	OrderID      string                 `json:"order_id"`
	CustomerID   string                 `json:"customer_id"`
	Status       string                 `json:"status,omitempty"`
	TotalAmount  float64                `json:"total_amount,omitempty"`
	Currency     string                 `json:"currency,omitempty"`
	OrderData    map[string]interface{} `json:"order_data,omitempty"`
	Changes      map[string]interface{} `json:"changes,omitempty"`
}

// OrderItemEvent represents order item events
type OrderItemEvent struct {
	BaseEvent
	OrderID     string  `json:"order_id"`
	ItemID      string  `json:"item_id"`
	ProductID   string  `json:"product_id"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	TotalPrice  float64 `json:"total_price"`
}

// PaymentEvent represents payment-related events
type PaymentEvent struct {
	BaseEvent
	OrderID       string  `json:"order_id"`
	PaymentID     string  `json:"payment_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	PaymentMethod string  `json:"payment_method"`
	Status        string  `json:"status"`
	ProviderData  map[string]interface{} `json:"provider_data,omitempty"`
}

// InventoryEvent represents inventory-related events
type InventoryEvent struct {
	BaseEvent
	OrderID     string `json:"order_id"`
	ProductID   string `json:"product_id"`
	Quantity    int    `json:"quantity"`
	Action      string `json:"action"` // reserve, release, deduct
	ReservationID string `json:"reservation_id,omitempty"`
}

// NotificationEvent represents notification events
type NotificationEvent struct {
	BaseEvent
	OrderID        string                 `json:"order_id"`
	CustomerID     string                 `json:"customer_id"`
	NotificationType string               `json:"notification_type"`
	Channel        string                 `json:"channel"` // email, sms, push
	Template       string                 `json:"template"`
	Data           map[string]interface{} `json:"data"`
}

// generateID generates a unique ID for events
func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("%x-%x-%x-%x-%x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:])
}

// NewOrderEvent creates a new order event
func NewOrderEvent(eventType string, orderID string, customerID string) *OrderEvent {
	return &OrderEvent{
		BaseEvent: BaseEvent{
			EventID:     generateID(),
			EventType:   eventType,
			AggregateID: orderID,
			Timestamp:   time.Now(),
			Version:     1,
		},
		OrderID:    orderID,
		CustomerID: customerID,
	}
}

// NewOrderItemEvent creates a new order item event
func NewOrderItemEvent(eventType string, orderID string, itemID string, productID string, quantity int, price float64) *OrderItemEvent {
	return &OrderItemEvent{
		BaseEvent: BaseEvent{
			EventID:     generateID(),
			EventType:   eventType,
			AggregateID: orderID,
			Timestamp:   time.Now(),
			Version:     1,
		},
		OrderID:    orderID,
		ItemID:     itemID,
		ProductID:  productID,
		Quantity:   quantity,
		Price:      price,
		TotalPrice: float64(quantity) * price,
	}
}

// NewPaymentEvent creates a new payment event
func NewPaymentEvent(eventType string, orderID string, paymentID string, amount float64, currency string) *PaymentEvent {
	return &PaymentEvent{
		BaseEvent: BaseEvent{
			EventID:     generateID(),
			EventType:   eventType,
			AggregateID: orderID,
			Timestamp:   time.Now(),
			Version:     1,
		},
		OrderID:   orderID,
		PaymentID: paymentID,
		Amount:    amount,
		Currency:  currency,
	}
}

// NewInventoryEvent creates a new inventory event
func NewInventoryEvent(eventType string, orderID string, productID string, quantity int, action string) *InventoryEvent {
	return &InventoryEvent{
		BaseEvent: BaseEvent{
			EventID:     generateID(),
			EventType:   eventType,
			AggregateID: orderID,
			Timestamp:   time.Now(),
			Version:     1,
		},
		OrderID:   orderID,
		ProductID: productID,
		Quantity:  quantity,
		Action:    action,
	}
}

// NewNotificationEvent creates a new notification event
func NewNotificationEvent(eventType string, orderID string, customerID string, notificationType string, channel string) *NotificationEvent {
	return &NotificationEvent{
		BaseEvent: BaseEvent{
			EventID:     generateID(),
			EventType:   eventType,
			AggregateID: orderID,
			Timestamp:   time.Now(),
			Version:     1,
		},
		OrderID:          orderID,
		CustomerID:       customerID,
		NotificationType: notificationType,
		Channel:          channel,
	}
}
