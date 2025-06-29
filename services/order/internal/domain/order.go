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

// OrderSource represents the source channel of an order
type OrderSource string

const (
	OrderSourceOnline      OrderSource = "online"
	OrderSourcePOS         OrderSource = "POS"
	OrderSourceMarketplace OrderSource = "marketplace"
	OrderSourceLINE        OrderSource = "LINE"
	OrderSourceFacebook    OrderSource = "Facebook"
)

// PaidStatus represents the payment status of an order
type PaidStatus string

const (
	PaidStatusUnpaid     PaidStatus = "unpaid"
	PaidStatusPaid       PaidStatus = "paid"
	PaidStatusPartialPaid PaidStatus = "partial_paid"
	PaidStatusRefunded   PaidStatus = "refunded"
)

// PaymentMethod represents the payment method used
type PaymentMethod string

const (
	PaymentMethodCash         PaymentMethod = "cash"
	PaymentMethodCreditCard   PaymentMethod = "credit_card"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
	PaymentMethodQRCode       PaymentMethod = "qr_code"
	PaymentMethodWallet       PaymentMethod = "wallet"
	PaymentMethodInstallment  PaymentMethod = "installment"
)

// OrderItem represents an item in an order
type OrderItem struct {
	ID             uuid.UUID `json:"id" db:"id"`
	OrderID        uuid.UUID `json:"order_id" db:"order_id"`
	ProductID      uuid.UUID `json:"product_id" db:"product_id"`
	Quantity       int       `json:"quantity" db:"quantity"`
	UnitPrice      float64   `json:"unit_price" db:"unit_price"`
	TotalPrice     float64   `json:"total_price" db:"total_price"`
	IsOverride     bool      `json:"is_override" db:"is_override"`
	OverrideReason *string   `json:"override_reason,omitempty" db:"override_reason"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Order represents an order in the system
type Order struct {
	ID               uuid.UUID      `json:"id" db:"id"`
	CustomerID       uuid.UUID      `json:"customer_id" db:"customer_id"`
	Code             *string        `json:"code,omitempty" db:"code"`
	Status           OrderStatus    `json:"status" db:"status"`
	Source           OrderSource    `json:"source" db:"source"`
	PaidStatus       PaidStatus     `json:"paid_status" db:"paid_status"`
	TotalAmount      float64        `json:"total_amount" db:"total_amount"`
	Discount         float64        `json:"discount" db:"discount"`
	ShippingFee      float64        `json:"shipping_fee" db:"shipping_fee"`
	Tax              float64        `json:"tax" db:"tax"`
	TaxEnabled       bool           `json:"tax_enabled" db:"tax_enabled"`
	ShippingAddress  string         `json:"shipping_address" db:"shipping_address"`
	BillingAddress   string         `json:"billing_address" db:"billing_address"`
	PaymentMethod    *PaymentMethod `json:"payment_method,omitempty" db:"payment_method"`
	PromoCode        *string        `json:"promo_code,omitempty" db:"promo_code"`
	Notes            string         `json:"notes" db:"notes"`
	ConfirmedAt      *time.Time     `json:"confirmed_at,omitempty" db:"confirmed_at"`
	CancelledAt      *time.Time     `json:"cancelled_at,omitempty" db:"cancelled_at"`
	CancelledReason  *string        `json:"cancelled_reason,omitempty" db:"cancelled_reason"`
	CreatedAt        time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at" db:"updated_at"`
	Items            []OrderItem    `json:"items,omitempty"`
}

// NewOrder creates a new order with generated ID and current timestamp
func NewOrder(customerID uuid.UUID, shippingAddress, billingAddress, notes string) *Order {
	now := time.Now()
	return &Order{
		ID:              uuid.New(),
		CustomerID:      customerID,
		Status:          OrderStatusPending,
		Source:          OrderSourceOnline, // Default to online
		PaidStatus:      PaidStatusUnpaid,
		TotalAmount:     0,
		Discount:        0,
		ShippingFee:     0,
		Tax:             0,
		TaxEnabled:      true,
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

// AddItemWithOverride adds an item to the order with stock override capability
func (o *Order) AddItemWithOverride(productID uuid.UUID, quantity int, unitPrice float64, isOverride bool, overrideReason *string) {
	item := OrderItem{
		ID:             uuid.New(),
		OrderID:        o.ID,
		ProductID:      productID,
		Quantity:       quantity,
		UnitPrice:      unitPrice,
		TotalPrice:     float64(quantity) * unitPrice,
		IsOverride:     isOverride,
		OverrideReason: overrideReason,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	o.Items = append(o.Items, item)
	o.CalculateTotal()
	o.UpdatedAt = time.Now()
}

// CalculateTotal recalculates the total amount of the order including shipping and tax
func (o *Order) CalculateTotal() {
	itemsTotal := 0.0
	for _, item := range o.Items {
		itemsTotal += item.TotalPrice
	}
	
	// Calculate final total: (items - discount) + shipping + tax
	subtotal := itemsTotal - o.Discount
	o.TotalAmount = subtotal + o.ShippingFee + o.Tax
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

// GenerateOrderCode generates a unique order code
func (o *Order) GenerateOrderCode() string {
	// Generate order code based on timestamp in format ORD20250001
	codeStr := "ORD" + o.CreatedAt.Format("20060102") + o.ID.String()[:4]
	o.Code = &codeStr
	o.UpdatedAt = time.Now()
	return codeStr
}

// SetPaymentMethod sets the payment method for the order
func (o *Order) SetPaymentMethod(method PaymentMethod) {
	o.PaymentMethod = &method
	o.UpdatedAt = time.Now()
}

// SetPromoCode applies a promotional code to the order
func (o *Order) SetPromoCode(code string) {
	o.PromoCode = &code
	o.UpdatedAt = time.Now()
}

// ApplyDiscount applies a discount to the order
func (o *Order) ApplyDiscount(amount float64) error {
	if amount < 0 {
		return ErrInvalidAmount
	}
	o.Discount = amount
	o.CalculateTotal()
	return nil
}

// SetShippingFee sets the shipping fee for the order
func (o *Order) SetShippingFee(fee float64) error {
	if fee < 0 {
		return ErrInvalidAmount
	}
	o.ShippingFee = fee
	o.CalculateTotal()
	return nil
}

// CalculateTax calculates tax based on the subtotal
func (o *Order) CalculateTax(taxRate float64) {
	if o.TaxEnabled && taxRate > 0 {
		subtotal := o.GetSubtotal()
		o.Tax = subtotal * taxRate
	} else {
		o.Tax = 0
	}
	o.CalculateTotal()
}

// GetSubtotal returns the subtotal (items total - discount)
func (o *Order) GetSubtotal() float64 {
	itemsTotal := 0.0
	for _, item := range o.Items {
		itemsTotal += item.TotalPrice
	}
	return itemsTotal - o.Discount
}

// ConfirmOrder confirms the order and sets confirmed timestamp
func (o *Order) ConfirmOrder() error {
	if o.Status != OrderStatusPending {
		return ErrInvalidStatusTransition
	}
	
	now := time.Now()
	o.Status = OrderStatusConfirmed
	o.ConfirmedAt = &now
	o.UpdatedAt = now
	
	return nil
}

// CancelOrder cancels the order with reason
func (o *Order) CancelOrder(reason string) error {
	// Can cancel from any status except delivered, refunded, or already cancelled
	if o.Status == OrderStatusDelivered || o.Status == OrderStatusRefunded || o.Status == OrderStatusCancelled {
		return ErrInvalidStatusTransition
	}
	
	now := time.Now()
	o.Status = OrderStatusCancelled
	o.PaidStatus = PaidStatusUnpaid // Set to unpaid when cancelled
	o.CancelledAt = &now
	o.CancelledReason = &reason
	o.UpdatedAt = now
	
	return nil
}

// UpdatePaidStatus updates the payment status of the order
func (o *Order) UpdatePaidStatus(status PaidStatus) error {
	// Validate payment status transitions
	if !o.IsValidPaidStatusTransition(o.PaidStatus, status) {
		return ErrInvalidStatusTransition
	}
	
	o.PaidStatus = status
	o.UpdatedAt = time.Now()
	return nil
}

// IsValidPaidStatusTransition checks if a payment status transition is valid
func (o *Order) IsValidPaidStatusTransition(from, to PaidStatus) bool {
	validTransitions := map[PaidStatus][]PaidStatus{
		PaidStatusUnpaid:     {PaidStatusPartialPaid, PaidStatusPaid},
		PaidStatusPartialPaid: {PaidStatusPaid, PaidStatusRefunded},
		PaidStatusPaid:       {PaidStatusRefunded},
		PaidStatusRefunded:   {},
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