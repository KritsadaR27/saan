package domain

import (
	"context"

	"github.com/google/uuid"
)

// CustomerRepository defines the interface for customer data operations
type CustomerRepository interface {
	// Customer CRUD operations
	Create(ctx context.Context, customer *Customer) error
	GetByID(ctx context.Context, id uuid.UUID) (*Customer, error)
	GetByEmail(ctx context.Context, email string) (*Customer, error)
	GetByPhone(ctx context.Context, phone string) (*Customer, error)
	GetByLoyverseID(ctx context.Context, loyverseID string) (*Customer, error)
	Update(ctx context.Context, customer *Customer) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter CustomerFilter) ([]Customer, int, error)
	
	// Customer statistics
	UpdateTotalSpent(ctx context.Context, customerID uuid.UUID, amount float64) error
	GetTopCustomers(ctx context.Context, limit int) ([]Customer, error)
}

// CustomerAddressRepository defines the interface for customer address operations
type CustomerAddressRepository interface {
	Create(ctx context.Context, address *CustomerAddress) error
	GetByID(ctx context.Context, id uuid.UUID) (*CustomerAddress, error)
	GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]CustomerAddress, error)
	GetDefaultAddress(ctx context.Context, customerID uuid.UUID) (*CustomerAddress, error)
	Update(ctx context.Context, address *CustomerAddress) error
	Delete(ctx context.Context, id uuid.UUID) error
	SetAsDefault(ctx context.Context, addressID uuid.UUID, customerID uuid.UUID) error
}

// ThaiAddressRepository defines the interface for Thai address operations
type ThaiAddressRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*ThaiAddress, error)
	GetByPostalCode(ctx context.Context, postalCode string) ([]ThaiAddress, error)
	SearchByProvince(ctx context.Context, province string) ([]ThaiAddress, error)
	SearchByDistrict(ctx context.Context, district string) ([]ThaiAddress, error)
	SearchBySubdistrict(ctx context.Context, subdistrict string) ([]ThaiAddress, error)
	AutoComplete(ctx context.Context, query string, limit int) ([]ThaiAddress, error)
}

// DeliveryRouteRepository defines the interface for delivery route operations
type DeliveryRouteRepository interface {
	Create(ctx context.Context, route *DeliveryRoute) error
	GetByID(ctx context.Context, id uuid.UUID) (*DeliveryRoute, error)
	GetAll(ctx context.Context) ([]DeliveryRoute, error)
	Update(ctx context.Context, route *DeliveryRoute) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// CustomerFilter defines filters for customer queries
type CustomerFilter struct {
	Email           *string
	Phone           *string
	Tier            *CustomerTier
	IsActive        *bool
	DeliveryRouteID *uuid.UUID
	Limit           int
	Offset          int
	SortBy          string
	SortOrder       string
}

// CustomerService defines the business logic interface for customer operations
type CustomerService interface {
	// Customer operations
	CreateCustomer(ctx context.Context, customer *Customer) (*Customer, error)
	GetCustomer(ctx context.Context, id uuid.UUID) (*CustomerWithAddresses, error)
	GetCustomerByEmail(ctx context.Context, email string) (*Customer, error)
	GetCustomerByPhone(ctx context.Context, phone string) (*Customer, error)
	UpdateCustomer(ctx context.Context, customer *Customer) (*Customer, error)
	DeleteCustomer(ctx context.Context, id uuid.UUID) error
	ListCustomers(ctx context.Context, filter CustomerFilter) ([]Customer, int, error)
	
	// Address operations
	AddCustomerAddress(ctx context.Context, address *CustomerAddress) (*CustomerAddress, error)
	UpdateCustomerAddress(ctx context.Context, address *CustomerAddress) (*CustomerAddress, error)
	DeleteCustomerAddress(ctx context.Context, addressID uuid.UUID) error
	SetDefaultAddress(ctx context.Context, addressID uuid.UUID, customerID uuid.UUID) error
	
	// Thai address operations
	SearchThaiAddresses(ctx context.Context, query string, limit int) ([]ThaiAddress, error)
	GetThaiAddressByPostalCode(ctx context.Context, postalCode string) ([]ThaiAddress, error)
	
	// Tier management
	UpdateCustomerTier(ctx context.Context, customerID uuid.UUID, totalSpent float64) error
	
	// Loyverse integration
	SyncWithLoyverse(ctx context.Context, customerID uuid.UUID) error
	CreateLoyverseCustomer(ctx context.Context, customer *Customer) (*string, error)
}

// EventPublisher defines the interface for publishing domain events
type EventPublisher interface {
	PublishCustomerCreated(ctx context.Context, customer *Customer) error
	PublishCustomerUpdated(ctx context.Context, customer *Customer) error
	PublishCustomerDeleted(ctx context.Context, customerID uuid.UUID) error
	PublishCustomerTierUpdated(ctx context.Context, customerID uuid.UUID, oldTier, newTier CustomerTier) error
	PublishLoyverseCustomerSynced(ctx context.Context, customerID uuid.UUID, loyverseID string) error
}

// CacheRepository defines the interface for caching operations
type CacheRepository interface {
	GetCustomer(ctx context.Context, key string) (*Customer, error)
	SetCustomer(ctx context.Context, key string, customer *Customer, ttl int) error
	DeleteCustomer(ctx context.Context, key string) error
	GetThaiAddresses(ctx context.Context, key string) ([]ThaiAddress, error)
	SetThaiAddresses(ctx context.Context, key string, addresses []ThaiAddress, ttl int) error
}

// LoyverseClient defines the interface for Loyverse API integration
type LoyverseClient interface {
	CreateCustomer(ctx context.Context, customer *Customer) (*string, error)
	GetCustomer(ctx context.Context, loyverseID string) (*Customer, error)
	UpdateCustomer(ctx context.Context, loyverseID string, customer *Customer) error
	SearchCustomerByEmail(ctx context.Context, email string) (*string, error)
	SearchCustomerByPhone(ctx context.Context, phone string) (*string, error)
}
