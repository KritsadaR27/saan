package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/saan-system/services/customer/internal/domain"
)

type customerService struct {
	customerRepo        domain.CustomerRepository
	addressRepo         domain.CustomerAddressRepository
	thaiAddressRepo     domain.ThaiAddressRepository
	deliveryRouteRepo   domain.DeliveryRouteRepository
	cacheRepo           domain.CacheRepository
	eventPublisher      domain.EventPublisher
	loyverseClient      domain.LoyverseClient
	logger              *zap.Logger
}

// NewCustomerService creates a new customer service
func NewCustomerService(
	customerRepo domain.CustomerRepository,
	addressRepo domain.CustomerAddressRepository,
	thaiAddressRepo domain.ThaiAddressRepository,
	deliveryRouteRepo domain.DeliveryRouteRepository,
	cacheRepo domain.CacheRepository,
	eventPublisher domain.EventPublisher,
	loyverseClient domain.LoyverseClient,
	logger *zap.Logger,
) domain.CustomerService {
	return &customerService{
		customerRepo:      customerRepo,
		addressRepo:       addressRepo,
		thaiAddressRepo:   thaiAddressRepo,
		deliveryRouteRepo: deliveryRouteRepo,
		cacheRepo:         cacheRepo,
		eventPublisher:    eventPublisher,
		loyverseClient:    loyverseClient,
		logger:            logger,
	}
}

// CreateCustomer creates a new customer
func (s *customerService) CreateCustomer(ctx context.Context, customer *domain.Customer) (*domain.Customer, error) {
	// Validate customer data
	if err := customer.ValidateCustomer(); err != nil {
		return nil, err
	}

	// Check if customer already exists by email or phone
	existingCustomer, _ := s.customerRepo.GetByEmail(ctx, customer.Email)
	if existingCustomer != nil {
		return nil, domain.ErrCustomerExists
	}

	existingCustomer, _ = s.customerRepo.GetByPhone(ctx, customer.Phone)
	if existingCustomer != nil {
		return nil, domain.ErrCustomerExists
	}

	// Set default values
	customer.ID = uuid.New()
	customer.Tier = domain.TierBronze
	customer.TotalSpent = 0
	customer.OrderCount = 0
	customer.IsActive = true
	now := time.Now()
	customer.CreatedAt = now
	customer.UpdatedAt = now

	// Create customer in database
	if err := s.customerRepo.Create(ctx, customer); err != nil {
		s.logger.Error("Failed to create customer", zap.Error(err))
		return nil, err
	}

	// Publish customer created event
	if err := s.eventPublisher.PublishCustomerCreated(ctx, customer); err != nil {
		s.logger.Error("Failed to publish customer created event", zap.Error(err))
		// Don't return error, just log it
	}

	s.logger.Info("Customer created successfully", zap.String("customer_id", customer.ID.String()))
	return customer, nil
}

// GetCustomer retrieves a customer with addresses by ID
func (s *customerService) GetCustomer(ctx context.Context, id uuid.UUID) (*domain.CustomerWithAddresses, error) {
	// Try to get from cache first
	cacheKey := "customer:" + id.String()
	customer, err := s.cacheRepo.GetCustomer(ctx, cacheKey)
	if err == nil && customer != nil {
		// Get addresses
		addresses, err := s.addressRepo.GetByCustomerID(ctx, id)
		if err != nil {
			s.logger.Error("Failed to get customer addresses", zap.Error(err))
			return nil, err
		}

		return &domain.CustomerWithAddresses{
			Customer:  *customer,
			Addresses: addresses,
		}, nil
	}

	// Get from database
	customer, err = s.customerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the customer
	if err := s.cacheRepo.SetCustomer(ctx, cacheKey, customer, 3600); err != nil {
		s.logger.Error("Failed to cache customer", zap.Error(err))
		// Don't return error, just log it
	}

	// Get addresses
	addresses, err := s.addressRepo.GetByCustomerID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get customer addresses", zap.Error(err))
		return nil, err
	}

	return &domain.CustomerWithAddresses{
		Customer:  *customer,
		Addresses: addresses,
	}, nil
}

// GetCustomerByEmail retrieves a customer by email
func (s *customerService) GetCustomerByEmail(ctx context.Context, email string) (*domain.Customer, error) {
	return s.customerRepo.GetByEmail(ctx, email)
}

// GetCustomerByPhone retrieves a customer by phone
func (s *customerService) GetCustomerByPhone(ctx context.Context, phone string) (*domain.Customer, error) {
	return s.customerRepo.GetByPhone(ctx, phone)
}

// UpdateCustomer updates a customer
func (s *customerService) UpdateCustomer(ctx context.Context, customer *domain.Customer) (*domain.Customer, error) {
	// Validate customer data
	if err := customer.ValidateCustomer(); err != nil {
		return nil, err
	}

	// Check if customer exists
	existingCustomer, err := s.customerRepo.GetByID(ctx, customer.ID)
	if err != nil {
		return nil, err
	}

	// Update timestamp
	customer.UpdatedAt = time.Now()

	// Update in database
	if err := s.customerRepo.Update(ctx, customer); err != nil {
		s.logger.Error("Failed to update customer", zap.Error(err))
		return nil, err
	}

	// Clear cache
	cacheKey := "customer:" + customer.ID.String()
	if err := s.cacheRepo.DeleteCustomer(ctx, cacheKey); err != nil {
		s.logger.Error("Failed to clear customer cache", zap.Error(err))
	}

	// Publish customer updated event
	if err := s.eventPublisher.PublishCustomerUpdated(ctx, customer); err != nil {
		s.logger.Error("Failed to publish customer updated event", zap.Error(err))
	}

	// Check if tier changed and publish tier updated event
	if existingCustomer.Tier != customer.Tier {
		if err := s.eventPublisher.PublishCustomerTierUpdated(ctx, customer.ID, existingCustomer.Tier, customer.Tier); err != nil {
			s.logger.Error("Failed to publish customer tier updated event", zap.Error(err))
		}
	}

	s.logger.Info("Customer updated successfully", zap.String("customer_id", customer.ID.String()))
	return customer, nil
}

// DeleteCustomer soft deletes a customer
func (s *customerService) DeleteCustomer(ctx context.Context, id uuid.UUID) error {
	// Check if customer exists
	_, err := s.customerRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Soft delete customer
	if err := s.customerRepo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete customer", zap.Error(err))
		return err
	}

	// Clear cache
	cacheKey := "customer:" + id.String()
	if err := s.cacheRepo.DeleteCustomer(ctx, cacheKey); err != nil {
		s.logger.Error("Failed to clear customer cache", zap.Error(err))
	}

	// Publish customer deleted event
	if err := s.eventPublisher.PublishCustomerDeleted(ctx, id); err != nil {
		s.logger.Error("Failed to publish customer deleted event", zap.Error(err))
	}

	s.logger.Info("Customer deleted successfully", zap.String("customer_id", id.String()))
	return nil
}

// ListCustomers retrieves customers with filtering and pagination
func (s *customerService) ListCustomers(ctx context.Context, filter domain.CustomerFilter) ([]domain.Customer, int, error) {
	return s.customerRepo.List(ctx, filter)
}

// AddCustomerAddress adds a new address for a customer
func (s *customerService) AddCustomerAddress(ctx context.Context, address *domain.CustomerAddress) (*domain.CustomerAddress, error) {
	// Check if customer exists
	_, err := s.customerRepo.GetByID(ctx, address.CustomerID)
	if err != nil {
		return nil, err
	}

	// Validate address data
	if address.AddressLine1 == "" {
		return nil, domain.ErrInvalidAddressLine1
	}
	if address.PostalCode == "" {
		return nil, domain.ErrInvalidPostalCode
	}

	// Set default values
	address.ID = uuid.New()
	address.IsActive = true
	now := time.Now()
	address.CreatedAt = now
	address.UpdatedAt = now

	// If this is the first address, make it default
	addresses, err := s.addressRepo.GetByCustomerID(ctx, address.CustomerID)
	if err != nil {
		return nil, err
	}
	if len(addresses) == 0 {
		address.IsDefault = true
	}

	// Create address
	if err := s.addressRepo.Create(ctx, address); err != nil {
		s.logger.Error("Failed to create customer address", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Customer address created successfully", zap.String("address_id", address.ID.String()))
	return address, nil
}

// UpdateCustomerAddress updates a customer address
func (s *customerService) UpdateCustomerAddress(ctx context.Context, address *domain.CustomerAddress) (*domain.CustomerAddress, error) {
	// Validate address data
	if address.AddressLine1 == "" {
		return nil, domain.ErrInvalidAddressLine1
	}
	if address.PostalCode == "" {
		return nil, domain.ErrInvalidPostalCode
	}

	// Update timestamp
	address.UpdatedAt = time.Now()

	// Update address
	if err := s.addressRepo.Update(ctx, address); err != nil {
		s.logger.Error("Failed to update customer address", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Customer address updated successfully", zap.String("address_id", address.ID.String()))
	return address, nil
}

// DeleteCustomerAddress soft deletes a customer address
func (s *customerService) DeleteCustomerAddress(ctx context.Context, addressID uuid.UUID) error {
	// Delete address
	if err := s.addressRepo.Delete(ctx, addressID); err != nil {
		s.logger.Error("Failed to delete customer address", zap.Error(err))
		return err
	}

	s.logger.Info("Customer address deleted successfully", zap.String("address_id", addressID.String()))
	return nil
}

// SetDefaultAddress sets an address as default for a customer
func (s *customerService) SetDefaultAddress(ctx context.Context, addressID uuid.UUID, customerID uuid.UUID) error {
	if err := s.addressRepo.SetAsDefault(ctx, addressID, customerID); err != nil {
		s.logger.Error("Failed to set default address", zap.Error(err))
		return err
	}

	s.logger.Info("Default address set successfully", zap.String("address_id", addressID.String()))
	return nil
}

// SearchThaiAddresses searches Thai addresses with autocomplete
func (s *customerService) SearchThaiAddresses(ctx context.Context, query string, limit int) ([]domain.ThaiAddress, error) {
	// Try cache first
	cacheKey := "thai_addresses:" + query
	addresses, err := s.cacheRepo.GetThaiAddresses(ctx, cacheKey)
	if err == nil && addresses != nil {
		return addresses, nil
	}

	// Search in database
	addresses, err = s.thaiAddressRepo.AutoComplete(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	// Cache results
	if err := s.cacheRepo.SetThaiAddresses(ctx, cacheKey, addresses, 3600); err != nil {
		s.logger.Error("Failed to cache Thai addresses", zap.Error(err))
	}

	return addresses, nil
}

// GetThaiAddressByPostalCode retrieves Thai addresses by postal code
func (s *customerService) GetThaiAddressByPostalCode(ctx context.Context, postalCode string) ([]domain.ThaiAddress, error) {
	return s.thaiAddressRepo.GetByPostalCode(ctx, postalCode)
}

// UpdateCustomerTier updates customer tier based on total spent
func (s *customerService) UpdateCustomerTier(ctx context.Context, customerID uuid.UUID, totalSpent float64) error {
	// Get current customer
	customer, err := s.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		return err
	}

	oldTier := customer.Tier

	// Update total spent and tier
	if err := s.customerRepo.UpdateTotalSpent(ctx, customerID, totalSpent); err != nil {
		s.logger.Error("Failed to update customer tier", zap.Error(err))
		return err
	}

	// Get updated customer to check new tier
	updatedCustomer, err := s.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		return err
	}

	// Clear cache
	cacheKey := "customer:" + customerID.String()
	if err := s.cacheRepo.DeleteCustomer(ctx, cacheKey); err != nil {
		s.logger.Error("Failed to clear customer cache", zap.Error(err))
	}

	// Publish tier updated event if tier changed
	if oldTier != updatedCustomer.Tier {
		if err := s.eventPublisher.PublishCustomerTierUpdated(ctx, customerID, oldTier, updatedCustomer.Tier); err != nil {
			s.logger.Error("Failed to publish customer tier updated event", zap.Error(err))
		}
	}

	s.logger.Info("Customer tier updated successfully", 
		zap.String("customer_id", customerID.String()),
		zap.String("old_tier", string(oldTier)),
		zap.String("new_tier", string(updatedCustomer.Tier)))

	return nil
}

// SyncWithLoyverse syncs customer with Loyverse
func (s *customerService) SyncWithLoyverse(ctx context.Context, customerID uuid.UUID) error {
	// Get customer
	customer, err := s.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		return err
	}

	// If customer already has Loyverse ID, skip
	if customer.LoyverseID != nil {
		return nil
	}

	// Try to find customer in Loyverse by email
	loyverseID, err := s.loyverseClient.SearchCustomerByEmail(ctx, customer.Email)
	if err == nil && loyverseID != nil {
		// Update customer with Loyverse ID
		customer.LoyverseID = loyverseID
		customer.UpdatedAt = time.Now()
		
		if err := s.customerRepo.Update(ctx, customer); err != nil {
			return err
		}

		// Publish sync event
		if err := s.eventPublisher.PublishLoyverseCustomerSynced(ctx, customerID, *loyverseID); err != nil {
			s.logger.Error("Failed to publish Loyverse sync event", zap.Error(err))
		}

		s.logger.Info("Customer synced with existing Loyverse customer", 
			zap.String("customer_id", customerID.String()),
			zap.String("loyverse_id", *loyverseID))
		return nil
	}

	// Create new customer in Loyverse
	loyverseID, err = s.CreateLoyverseCustomer(ctx, customer)
	if err != nil {
		return err
	}

	// Update customer with Loyverse ID
	customer.LoyverseID = loyverseID
	customer.UpdatedAt = time.Now()
	
	if err := s.customerRepo.Update(ctx, customer); err != nil {
		return err
	}

	// Clear cache
	cacheKey := "customer:" + customerID.String()
	if err := s.cacheRepo.DeleteCustomer(ctx, cacheKey); err != nil {
		s.logger.Error("Failed to clear customer cache", zap.Error(err))
	}

	// Publish sync event
	if err := s.eventPublisher.PublishLoyverseCustomerSynced(ctx, customerID, *loyverseID); err != nil {
		s.logger.Error("Failed to publish Loyverse sync event", zap.Error(err))
	}

	s.logger.Info("Customer synced with new Loyverse customer", 
		zap.String("customer_id", customerID.String()),
		zap.String("loyverse_id", *loyverseID))

	return nil
}

// CreateLoyverseCustomer creates a customer in Loyverse
func (s *customerService) CreateLoyverseCustomer(ctx context.Context, customer *domain.Customer) (*string, error) {
	return s.loyverseClient.CreateCustomer(ctx, customer)
}
