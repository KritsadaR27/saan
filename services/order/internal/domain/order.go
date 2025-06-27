package domain

import (
	"time"

	"github.com/google/uuid"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusRefunded   OrderStatus = "refunded"
)

// OrderItem represents an item in an order
type OrderItem struct {
	ID          uuid.UUID `json:"id" db:"id"`
	OrderID     uuid.UUID `json:"order_id" db:"order_id"`
	ProductID   uuid.UUID `json:"product_id" db:"product_id"`
	Quantity    int       `json:"quantity" db:"quantity"`
	UnitPrice   float64   `json:"unit_price" db:"unit_price"`
	TotalPrice  float64   `json:"total_price" db:"total_price"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Order represents an order in the system
type Order struct {
	ID              uuid.UUID   `json:"id" db:"id"`
	CustomerID      uuid.UUID   `json:"customer_id" db:"customer_id"`
	Status          OrderStatus `json:"status" db:"status"`
	TotalAmount     float64     `json:"total_amount" db:"total_amount"`
	ShippingAddress string      `json:"shipping_address" db:"shipping_address"`
	BillingAddress  string      `json:"billing_address" db:"billing_address"`
	Notes           string      `json:"notes" db:"notes"`
	CreatedAt       time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at" db:"updated_at"`
	Items           []OrderItem `json:"items,omitempty"`
}

// NewOrder creates a new order with generated ID and current timestamp
func NewOrder(customerID uuid.UUID, shippingAddress, billingAddress, notes string) *Order {
	now := time.Now()
	return &Order{
		ID:              uuid.New(),
		CustomerID:      customerID,
		Status:          OrderStatusPending,
		TotalAmount:     0,
		ShippingAddress: shippingAddress,
		BillingAddress:  billingAddress,
		Notes:           notes,
		CreatedAt:       now,
		UpdatedAt:       now,
		Items:           []OrderItem{},
	}
}

// AddItem adds an item to the order and recalculates total
func (o *Order) AddItem(productID uuid.UUID, quantity int, unitPrice float64) {
	item := OrderItem{
		ID:         uuid.New(),
		OrderID:    o.ID,
		ProductID:  productID,
		Quantity:   quantity,
		UnitPrice:  unitPrice,
		TotalPrice: float64(quantity) * unitPrice,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	
	o.Items = append(o.Items, item)
	o.CalculateTotal()
	o.UpdatedAt = time.Now()
}

// CalculateTotal recalculates the total amount of the order
func (o *Order) CalculateTotal() {
	total := 0.0
	for _, item := range o.Items {
		total += item.TotalPrice
	}
	o.TotalAmount = total
	o.UpdatedAt = time.Now()
}

// UpdateStatus updates the order status
func (o *Order) UpdateStatus(status OrderStatus) error {
	if !o.IsValidStatusTransition(o.Status, status) {
		return ErrInvalidStatusTransition
	}
	o.Status = status
	o.UpdatedAt = time.Now()
	return nil
}

// IsValidStatusTransition checks if a status transition is valid
func (o *Order) IsValidStatusTransition(from, to OrderStatus) bool {
	validTransitions := map[OrderStatus][]OrderStatus{
		OrderStatusPending:    {OrderStatusConfirmed, OrderStatusCancelled},
		OrderStatusConfirmed:  {OrderStatusProcessing, OrderStatusCancelled},
		OrderStatusProcessing: {OrderStatusShipped, OrderStatusCancelled},
		OrderStatusShipped:    {OrderStatusDelivered},
		OrderStatusDelivered:  {OrderStatusRefunded},
		OrderStatusCancelled:  {},
		OrderStatusRefunded:   {},
	}
	
	allowed, exists := validTransitions[from]
	if !exists {
		return false
	}
	
	for _, allowedStatus := range allowed {
		if allowedStatus == to {
			return true
		}
	}
	
	return false
}