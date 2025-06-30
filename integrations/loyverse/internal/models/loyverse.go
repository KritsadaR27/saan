// integrations/loyverse/internal/models/loyverse.go
package models

import (
	"time"
)

// Employee represents Loyverse employee
type Employee struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Phone     string     `json:"phone_number"`
	StoreID   string     `json:"store_id"`
	Roles     []string   `json:"pos_roles"`
	IsOwner   bool       `json:"is_owner"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

// Category represents Loyverse category
type Category struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Color     string     `json:"color"`
	ParentID  *string    `json:"parent_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

// Supplier represents Loyverse supplier
type Supplier struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	ContactName string     `json:"contact_name"`
	Phone       string     `json:"phone_number"`
	Email       string     `json:"email"`
	Address     string     `json:"address"`
	Note        string     `json:"note"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at"`
}

// Discount represents Loyverse discount
type Discount struct {
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	Type              string     `json:"type"` // FIXED_AMOUNT or PERCENTAGE
	Value             float64    `json:"value"`
	IsRestricted      bool       `json:"is_restricted"`
	RestrictedAccess  []string   `json:"restricted_access"`
	ApplicableToItems []string   `json:"applicable_to_items"`
	MinimumAmount     *float64   `json:"minimum_subtotal"`
	ValidSince        *time.Time `json:"valid_since"`
	ValidUntil        *time.Time `json:"valid_until"`
	StoreIDs          []string   `json:"stores"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at"`
}

// PaymentType represents Loyverse payment type
type PaymentType struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Type           string     `json:"type"` // CASH, CARD, etc.
	ShowInPOS      bool       `json:"show_in_pos"`
	ShowInReceipts bool       `json:"show_in_receipts"`
	OrderIndex     int        `json:"order_index"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at"`
}

// Store represents Loyverse store
type Store struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Address       string     `json:"address"`
	Phone         string     `json:"phone_number"`
	Email         string     `json:"email"`
	Description   string     `json:"description"`
	ReceiptFooter string     `json:"receipt_footer_text"`
	TaxNumber     string     `json:"tax_number"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at"`
}

// Item represents Loyverse item (product)
type Item struct {
	ID                string      `json:"id"`
	Name              string      `json:"item_name"`
	Description       string      `json:"description"`
	CategoryID        *string     `json:"category_id"`
	PrimarySupplierID *string     `json:"primary_supplier_id"`
	ImageURL          string      `json:"image_url"`
	SKU               string      `json:"sku"`
	Barcode           string      `json:"barcode"`
	TrackStock        bool        `json:"track_stock"`
	SoldByWeight      bool        `json:"sold_by_weight"`
	IsComposite       bool        `json:"is_composite"`
	UseProduction     bool        `json:"use_production"`
	Components        []Component `json:"components"`
	Variants          []Variant   `json:"variants"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
	DeletedAt         *time.Time  `json:"deleted_at"`
}

// Component for composite items
type Component struct {
	ItemID    string  `json:"item_id"`
	VariantID string  `json:"variant_id"`
	Quantity  float64 `json:"quantity"`
}

// Variant represents item variant
type Variant struct {
	ID           string       `json:"variant_id"`
	ItemID       string       `json:"item_id"`
	SKU          string       `json:"sku"`
	Barcode      string       `json:"barcode"`
	Option1Name  string       `json:"option1_name"`
	Option1Value string       `json:"option1_value"`
	Option2Name  string       `json:"option2_name"`
	Option2Value string       `json:"option2_value"`
	Option3Name  string       `json:"option3_name"`
	Option3Value string       `json:"option3_value"`
	Cost         float64      `json:"cost"`
	PurchaseCost float64      `json:"purchase_cost"`
	DefaultPrice float64      `json:"default_price"`
	Stores       []StorePrice `json:"stores"`
}

// StorePrice represents variant price in specific store
type StorePrice struct {
	StoreID string  `json:"store_id"`
	Price   float64 `json:"price"`
}

// InventoryLevel represents stock level
type InventoryLevel struct {
	VariantID string    `json:"variant_id"`
	StoreID   string    `json:"store_id"`
	InStock   float64   `json:"in_stock"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Receipt represents Loyverse receipt
type Receipt struct {
	ID             string     `json:"receipt_id"`
	Number         string     `json:"receipt_number"`
	Note           string     `json:"note"`
	ReceiptType    string     `json:"receipt_type"` // SALE or REFUND
	RefundFor      *string    `json:"refund_for"`
	Order          *string    `json:"order"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	ReceiptDate    time.Time  `json:"receipt_date"`
	CancelledAt    *time.Time `json:"cancelled_at"`
	Source         string     `json:"source"`
	TotalMoney     float64    `json:"total_money"`
	TotalTax       float64    `json:"total_tax"`
	PointsEarned   int        `json:"points_earned"`
	PointsDeducted int        `json:"points_deducted"`
	PointsBalance  int        `json:"points_balance"`
	CustomerID     *string    `json:"customer_id"`
	CustomerName   string     `json:"customer_name"`
	CustomerPhone  string     `json:"customer_phone_number"`
	TotalDiscount  float64    `json:"total_discount"`
	EmployeeID     string     `json:"employee_id"`
	StoreID        string     `json:"store_id"`
	PosDeviceID    string     `json:"pos_device_id"`
	PosDeviceName  string     `json:"pos_device_name"`
	LineItems      []LineItem `json:"line_items"`
	Payments       []Payment  `json:"payments"`
}

// LineItem represents receipt line item
type LineItem struct {
	ID              string         `json:"id"`
	ItemID          string         `json:"item_id"`
	VariantID       string         `json:"variant_id"`
	ItemName        string         `json:"item_name"`
	VariantName     string         `json:"variant_name"`
	SKU             string         `json:"sku"`
	Quantity        float64        `json:"quantity"`
	Price           float64        `json:"price"`
	GrossTotalMoney float64        `json:"gross_total_money"`
	TotalMoney      float64        `json:"total_money"`
	Cost            float64        `json:"cost"`
	CostTotal       float64        `json:"cost_total"`
	LineNote        string         `json:"line_note"`
	TotalDiscount   float64        `json:"total_discount"`
	LineTaxes       []LineTax      `json:"line_taxes"`
	LineDiscounts   []LineDiscount `json:"line_discounts"`
	LineModifiers   []LineModifier `json:"line_modifiers"`
}

// LineTax represents tax applied to line item
type LineTax struct {
	ID     string  `json:"id"`
	TaxID  string  `json:"tax_id"`
	Name   string  `json:"name"`
	Rate   float64 `json:"rate"`
	Amount float64 `json:"tax_amount"`
}

// LineDiscount represents discount applied to line item
type LineDiscount struct {
	DiscountID string  `json:"discount_id"`
	Name       string  `json:"name"`
	Amount     float64 `json:"discount_amount"`
}

// LineModifier represents modifier applied to line item
type LineModifier struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Quantity float64 `json:"quantity"`
	Price    float64 `json:"price"`
}

// Payment represents receipt payment
type Payment struct {
	PaymentTypeID string  `json:"payment_type_id"`
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	Amount        float64 `json:"money_amount"`
}

// Customer represents Loyverse customer
type Customer struct {
	ID                string     `json:"customer_id"`
	Name              string     `json:"name"`
	Email             string     `json:"email"`
	Phone             string     `json:"phone_number"`
	Address           string     `json:"address"`
	City              string     `json:"city"`
	PostalCode        string     `json:"postal_code"`
	CountryCode       string     `json:"country_code"`
	Note              string     `json:"note"`
	FirstVisit        time.Time  `json:"first_visit"`
	LastVisit         time.Time  `json:"last_visit"`
	TotalSpent        float64    `json:"total_spent"`
	TotalOrders       int        `json:"total_orders"`
	AverageOrderValue float64    `json:"average_order"`
	PointsBalance     int        `json:"points_balance"`
	CustomerCode      string     `json:"customer_code"`
	TaxNumber         string     `json:"tax_number"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at"`
}

// Response wrappers for paginated API responses
type EmployeesResponse struct {
	Employees []Employee `json:"employees"`
	Cursor    string     `json:"cursor"`
}

type CategoriesResponse struct {
	Categories []Category `json:"categories"`
	Cursor     string     `json:"cursor"`
}

type SuppliersResponse struct {
	Suppliers []Supplier `json:"suppliers"`
	Cursor    string     `json:"cursor"`
}

type DiscountsResponse struct {
	Discounts []Discount `json:"discounts"`
	Cursor    string     `json:"cursor"`
}

type PaymentTypesResponse struct {
	PaymentTypes []PaymentType `json:"payment_types"`
	Cursor       string        `json:"cursor"`
}

type StoresResponse struct {
	Stores []Store `json:"stores"`
	Cursor string  `json:"cursor"`
}

type ItemsResponse struct {
	Items  []Item `json:"items"`
	Cursor string `json:"cursor"`
}

type InventoryLevelsResponse struct {
	InventoryLevels []InventoryLevel `json:"inventory_levels"`
	Cursor          string           `json:"cursor"`
}

type ReceiptsResponse struct {
	Receipts []Receipt `json:"receipts"`
	Cursor   string    `json:"cursor"`
}

type CustomersResponse struct {
	Customers []Customer `json:"customers"`
	Cursor    string     `json:"cursor"`
}
