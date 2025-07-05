package repository

import (
	"context"
	"time"

	"customer/internal/domain/entity"
	"github.com/google/uuid"
)

// CustomerFilter defines filters for customer queries
type CustomerFilter struct {
	Email           *string
	Phone           *string
	Tier            *entity.CustomerTier
	IsActive        *bool
	DeliveryRouteID *uuid.UUID
	Limit           int
	Offset          int
	SortBy          string
	SortOrder       string
}

// CustomerRepository defines the interface for customer data operations
type CustomerRepository interface {
	// Customer CRUD operations
	Create(ctx context.Context, customer *entity.Customer) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Customer, error)
	GetByEmail(ctx context.Context, email string) (*entity.Customer, error)
	GetByPhone(ctx context.Context, phone string) (*entity.Customer, error)
	GetByLoyverseID(ctx context.Context, loyverseID string) (*entity.Customer, error)
	GetByLineUserID(ctx context.Context, lineUserID string) (*entity.Customer, error)
	Update(ctx context.Context, customer *entity.Customer) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter CustomerFilter) ([]entity.Customer, int, error)

	// Customer statistics
	UpdateTotalSpent(ctx context.Context, customerID uuid.UUID, amount float64) error
	GetTopCustomers(ctx context.Context, limit int) ([]entity.Customer, error)
}

// CustomerAddressRepository defines the interface for customer address operations
type CustomerAddressRepository interface {
	Create(ctx context.Context, address *entity.CustomerAddress) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.CustomerAddress, error)
	GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]entity.CustomerAddress, error)
	GetDefaultAddress(ctx context.Context, customerID uuid.UUID) (*entity.CustomerAddress, error)
	Update(ctx context.Context, address *entity.CustomerAddress) error
	Delete(ctx context.Context, id uuid.UUID) error
	SetAsDefault(ctx context.Context, addressID uuid.UUID, customerID uuid.UUID) error
}

// ThaiAddressRepository defines the interface for Thai address operations
type ThaiAddressRepository interface {
	// Address suggestions
	GetAddressSuggestions(ctx context.Context, query string, limit int) ([]entity.AddressSuggestion, error)
	GetBySubdistrict(ctx context.Context, subdistrict string) ([]entity.ThaiAddress, error)
	GetProvinceDeliveryInfo(ctx context.Context, province string) (*entity.DeliveryRoute, error)

	// Existing methods used by customer service
	AutoComplete(ctx context.Context, query string, limit int) ([]entity.ThaiAddress, error)
	GetByPostalCode(ctx context.Context, postalCode string) ([]entity.ThaiAddress, error)
	SearchByProvince(ctx context.Context, province string) ([]entity.ThaiAddress, error)
	SearchByDistrict(ctx context.Context, district string) ([]entity.ThaiAddress, error)
	SearchBySubdistrict(ctx context.Context, subdistrict string) ([]entity.ThaiAddress, error)
}

// DeliveryRouteRepository defines the interface for delivery route operations
type DeliveryRouteRepository interface {
	Create(ctx context.Context, route *entity.DeliveryRoute) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.DeliveryRoute, error)
	GetAll(ctx context.Context) ([]entity.DeliveryRoute, error)
	Update(ctx context.Context, route *entity.DeliveryRoute) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// VIPTierBenefitsRepository defines the interface for VIP tier benefits operations
type VIPTierBenefitsRepository interface {
	GetByTier(ctx context.Context, tier entity.CustomerTier) (*entity.VIPTierBenefits, error)
	GetAll(ctx context.Context) ([]entity.VIPTierBenefits, error)
	Update(ctx context.Context, benefits *entity.VIPTierBenefits) error
}

// CustomerPointsRepository defines the interface for customer points operations
type CustomerPointsRepository interface {
	// Points transactions
	CreateTransaction(ctx context.Context, transaction *entity.CustomerPointsTransaction) error
	GetTransactionsByCustomer(ctx context.Context, customerID uuid.UUID, limit int, offset int) ([]entity.CustomerPointsTransaction, error)
	GetPointsBalance(ctx context.Context, customerID uuid.UUID) (int, error)

	// Points operations
	EarnPoints(ctx context.Context, customerID uuid.UUID, points int, source, description string, referenceID *uuid.UUID, referenceType *string) error
	RedeemPoints(ctx context.Context, customerID uuid.UUID, points int, source, description string, referenceID *uuid.UUID, referenceType *string) error
	ExpirePoints(ctx context.Context, customerID uuid.UUID, points int, description string) error
}

// CustomerAnalyticsRepository defines the interface for customer analytics operations
type CustomerAnalyticsRepository interface {
	GetCustomerInsights(ctx context.Context, customerID uuid.UUID) (*entity.CustomerAnalytics, error)
	UpdatePurchaseAnalytics(ctx context.Context, customerID uuid.UUID, orderValue float64, orderDate time.Time) error
	GetSegmentationData(ctx context.Context, customerID uuid.UUID) (map[string]interface{}, error)
	GetRecommendations(ctx context.Context, customerID uuid.UUID) ([]entity.UpsellSuggestion, error)
}

// LINEService defines the interface for LINE integration operations
type LINEService interface {
	// Digital card operations
	GenerateDigitalCard(ctx context.Context, customer *entity.Customer) (map[string]interface{}, error)
	SendDigitalCard(ctx context.Context, lineUserID string, customer *entity.Customer) error

	// Rich menu operations
	SetupRichMenu(ctx context.Context) error
	GetRichMenuAnalytics(ctx context.Context) (map[string]interface{}, error)

	// Customer registration
	RegisterCustomerFromLINE(ctx context.Context, lineUserID, displayName, phone string) (*entity.Customer, error)
	LinkLINEAccount(ctx context.Context, customerID uuid.UUID, lineUserID, displayName string) error
}

// LoyverseReceiptService defines the interface for Loyverse receipt operations
type LoyverseReceiptService interface {
	CreateReceipt(ctx context.Context, orderData OrderReceiptData) (*LoyverseReceipt, error)
	HandleReceiptWebhook(ctx context.Context, webhookData LoyverseReceiptWebhook) error
	SyncCustomerFromReceipt(ctx context.Context, receiptData LoyverseReceiptWebhook) error
}

// Supporting types for Loyverse integration
type OrderReceiptData struct {
	OrderNumber string      `json:"order_number"`
	CustomerID  uuid.UUID   `json:"customer_id"`
	TotalAmount float64     `json:"total_amount"`
	Items       []OrderItem `json:"items"`
	Payments    []PaymentMethod `json:"payments"`
}

type OrderItem struct {
	ProductID uuid.UUID `json:"product_id"`
	Name      string    `json:"name"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
}

type PaymentMethod struct {
	Type   string  `json:"type"`
	Amount float64 `json:"amount"`
}

type LoyverseReceipt struct {
	ID            string    `json:"id"`
	ReceiptNumber string    `json:"receipt_number"`
	CustomerID    string    `json:"customer_id"`
	TotalMoney    float64   `json:"total_money"`
	PointsEarned  float64   `json:"points_earned"`
	CreatedAt     time.Time `json:"created_at"`
}

type LoyverseReceiptWebhook struct {
	ReceiptNumber       string    `json:"receipt_number"`
	CustomerID          string    `json:"customer_id"`
	TotalMoney          float64   `json:"total_money"`
	PointsEarned        float64   `json:"points_earned"`
	PointsBalance       float64   `json:"points_balance"`
	CustomerTotalVisits int       `json:"customer_total_visits"`
	CustomerTotalSpent  float64   `json:"customer_total_spent"`
	ReceiptDate         time.Time `json:"created_at"`
}

// CustomerService defines the business logic interface for customer operations
type CustomerService interface {
	// Customer operations
	CreateCustomer(ctx context.Context, customer *entity.Customer) (*entity.Customer, error)
	GetCustomer(ctx context.Context, id uuid.UUID) (*entity.CustomerWithAddresses, error)
	GetCustomerByEmail(ctx context.Context, email string) (*entity.Customer, error)
	GetCustomerByPhone(ctx context.Context, phone string) (*entity.Customer, error)
	UpdateCustomer(ctx context.Context, customer *entity.Customer) (*entity.Customer, error)
	DeleteCustomer(ctx context.Context, id uuid.UUID) error
	ListCustomers(ctx context.Context, filter CustomerFilter) ([]entity.Customer, int, error)

	// Address operations
	AddCustomerAddress(ctx context.Context, address *entity.CustomerAddress) (*entity.CustomerAddress, error)
	UpdateCustomerAddress(ctx context.Context, address *entity.CustomerAddress) (*entity.CustomerAddress, error)
	DeleteCustomerAddress(ctx context.Context, addressID uuid.UUID) error
	SetDefaultAddress(ctx context.Context, addressID uuid.UUID, customerID uuid.UUID) error

	// Thai address operations
	SearchThaiAddresses(ctx context.Context, query string, limit int) ([]entity.ThaiAddress, error)
	GetThaiAddressByPostalCode(ctx context.Context, postalCode string) ([]entity.ThaiAddress, error)

	// Tier management
	UpdateCustomerTier(ctx context.Context, customerID uuid.UUID, totalSpent float64) error

	// Loyverse integration
	SyncWithLoyverse(ctx context.Context, customerID uuid.UUID) error
	CreateLoyverseCustomer(ctx context.Context, customer *entity.Customer) (*string, error)
}

// EventPublisher defines the interface for publishing domain events
type EventPublisher interface {
	PublishCustomerCreated(ctx context.Context, customer *entity.Customer) error
	PublishCustomerUpdated(ctx context.Context, customer *entity.Customer) error
	PublishCustomerDeleted(ctx context.Context, customerID uuid.UUID) error
	PublishCustomerTierUpdated(ctx context.Context, customerID uuid.UUID, oldTier, newTier entity.CustomerTier) error
	PublishLoyverseCustomerSynced(ctx context.Context, customerID uuid.UUID, loyverseID string) error
}

// CacheRepository defines the interface for caching operations
type CacheRepository interface {
	GetCustomer(ctx context.Context, key string) (*entity.Customer, error)
	SetCustomer(ctx context.Context, key string, customer *entity.Customer, ttl int) error
	DeleteCustomer(ctx context.Context, key string) error
	GetThaiAddresses(ctx context.Context, key string) ([]entity.ThaiAddress, error)
	SetThaiAddresses(ctx context.Context, key string, addresses []entity.ThaiAddress, ttl int) error
}

// LoyverseClient defines the interface for Loyverse API integration
type LoyverseClient interface {
	CreateCustomer(ctx context.Context, customer *entity.Customer) (*string, error)
	GetCustomer(ctx context.Context, loyverseID string) (*entity.Customer, error)
	UpdateCustomer(ctx context.Context, loyverseID string, customer *entity.Customer) error
	SearchCustomerByEmail(ctx context.Context, email string) (*string, error)
	SearchCustomerByPhone(ctx context.Context, phone string) (*string, error)
}

// ThaiAddressService defines the interface for Thai address operations
type ThaiAddressService interface {
	GetAddressSuggestions(ctx context.Context, query string, limit int) ([]entity.AddressSuggestion, error)
	GetBySubdistrict(ctx context.Context, subdistrict string) ([]entity.ThaiAddress, error)
	GetProvinceDeliveryInfo(ctx context.Context, province string) (*entity.DeliveryRoute, error)
}
