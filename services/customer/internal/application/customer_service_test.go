package application

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/saan-system/services/customer/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Mock implementations for testing
type MockCustomerRepository struct {
	mock.Mock
}

func (m *MockCustomerRepository) Create(ctx context.Context, customer *domain.Customer) error {
	args := m.Called(ctx, customer)
	return args.Error(0)
}

func (m *MockCustomerRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Customer), args.Error(1)
}

func (m *MockCustomerRepository) GetByEmail(ctx context.Context, email string) (*domain.Customer, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Customer), args.Error(1)
}

func (m *MockCustomerRepository) GetByPhone(ctx context.Context, phone string) (*domain.Customer, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Customer), args.Error(1)
}

func (m *MockCustomerRepository) GetByLoyverseID(ctx context.Context, loyverseID string) (*domain.Customer, error) {
	args := m.Called(ctx, loyverseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Customer), args.Error(1)
}

func (m *MockCustomerRepository) Update(ctx context.Context, customer *domain.Customer) error {
	args := m.Called(ctx, customer)
	return args.Error(0)
}

func (m *MockCustomerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCustomerRepository) List(ctx context.Context, filter domain.CustomerFilter) ([]domain.Customer, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]domain.Customer), args.Int(1), args.Error(2)
}

func (m *MockCustomerRepository) UpdateTotalSpent(ctx context.Context, customerID uuid.UUID, amount float64) error {
	args := m.Called(ctx, customerID, amount)
	return args.Error(0)
}

func (m *MockCustomerRepository) GetTopCustomers(ctx context.Context, limit int) ([]domain.Customer, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]domain.Customer), args.Error(1)
}

// MockCustomerAddressRepository
type MockCustomerAddressRepository struct {
	mock.Mock
}

func (m *MockCustomerAddressRepository) Create(ctx context.Context, address *domain.CustomerAddress) error {
	args := m.Called(ctx, address)
	return args.Error(0)
}

func (m *MockCustomerAddressRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.CustomerAddress, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CustomerAddress), args.Error(1)
}

func (m *MockCustomerAddressRepository) GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]domain.CustomerAddress, error) {
	args := m.Called(ctx, customerID)
	return args.Get(0).([]domain.CustomerAddress), args.Error(1)
}

func (m *MockCustomerAddressRepository) GetDefaultAddress(ctx context.Context, customerID uuid.UUID) (*domain.CustomerAddress, error) {
	args := m.Called(ctx, customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CustomerAddress), args.Error(1)
}

func (m *MockCustomerAddressRepository) Update(ctx context.Context, address *domain.CustomerAddress) error {
	args := m.Called(ctx, address)
	return args.Error(0)
}

func (m *MockCustomerAddressRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCustomerAddressRepository) SetAsDefault(ctx context.Context, addressID uuid.UUID, customerID uuid.UUID) error {
	args := m.Called(ctx, addressID, customerID)
	return args.Error(0)
}

// MockThaiAddressRepository
type MockThaiAddressRepository struct {
	mock.Mock
}

func (m *MockThaiAddressRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ThaiAddress, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ThaiAddress), args.Error(1)
}

func (m *MockThaiAddressRepository) GetByPostalCode(ctx context.Context, postalCode string) ([]domain.ThaiAddress, error) {
	args := m.Called(ctx, postalCode)
	return args.Get(0).([]domain.ThaiAddress), args.Error(1)
}

func (m *MockThaiAddressRepository) SearchByProvince(ctx context.Context, province string) ([]domain.ThaiAddress, error) {
	args := m.Called(ctx, province)
	return args.Get(0).([]domain.ThaiAddress), args.Error(1)
}

func (m *MockThaiAddressRepository) SearchByDistrict(ctx context.Context, district string) ([]domain.ThaiAddress, error) {
	args := m.Called(ctx, district)
	return args.Get(0).([]domain.ThaiAddress), args.Error(1)
}

func (m *MockThaiAddressRepository) SearchBySubdistrict(ctx context.Context, subdistrict string) ([]domain.ThaiAddress, error) {
	args := m.Called(ctx, subdistrict)
	return args.Get(0).([]domain.ThaiAddress), args.Error(1)
}

func (m *MockThaiAddressRepository) AutoComplete(ctx context.Context, query string, limit int) ([]domain.ThaiAddress, error) {
	args := m.Called(ctx, query, limit)
	return args.Get(0).([]domain.ThaiAddress), args.Error(1)
}

// MockDeliveryRouteRepository
type MockDeliveryRouteRepository struct {
	mock.Mock
}

func (m *MockDeliveryRouteRepository) Create(ctx context.Context, route *domain.DeliveryRoute) error {
	args := m.Called(ctx, route)
	return args.Error(0)
}

func (m *MockDeliveryRouteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.DeliveryRoute, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DeliveryRoute), args.Error(1)
}

func (m *MockDeliveryRouteRepository) GetAll(ctx context.Context) ([]domain.DeliveryRoute, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.DeliveryRoute), args.Error(1)
}

func (m *MockDeliveryRouteRepository) Update(ctx context.Context, route *domain.DeliveryRoute) error {
	args := m.Called(ctx, route)
	return args.Error(0)
}

func (m *MockDeliveryRouteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockCacheRepository
type MockCacheRepository struct {
	mock.Mock
}

func (m *MockCacheRepository) GetCustomer(ctx context.Context, key string) (*domain.Customer, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Customer), args.Error(1)
}

func (m *MockCacheRepository) SetCustomer(ctx context.Context, key string, customer *domain.Customer, ttl int) error {
	args := m.Called(ctx, key, customer, ttl)
	return args.Error(0)
}

func (m *MockCacheRepository) DeleteCustomer(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheRepository) GetThaiAddresses(ctx context.Context, key string) ([]domain.ThaiAddress, error) {
	args := m.Called(ctx, key)
	return args.Get(0).([]domain.ThaiAddress), args.Error(1)
}

func (m *MockCacheRepository) SetThaiAddresses(ctx context.Context, key string, addresses []domain.ThaiAddress, ttl int) error {
	args := m.Called(ctx, key, addresses, ttl)
	return args.Error(0)
}

// MockEventPublisher
type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) PublishCustomerCreated(ctx context.Context, customer *domain.Customer) error {
	args := m.Called(ctx, customer)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishCustomerUpdated(ctx context.Context, customer *domain.Customer) error {
	args := m.Called(ctx, customer)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishCustomerDeleted(ctx context.Context, customerID uuid.UUID) error {
	args := m.Called(ctx, customerID)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishCustomerTierUpdated(ctx context.Context, customerID uuid.UUID, oldTier, newTier domain.CustomerTier) error {
	args := m.Called(ctx, customerID, oldTier, newTier)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishLoyverseCustomerSynced(ctx context.Context, customerID uuid.UUID, loyverseID string) error {
	args := m.Called(ctx, customerID, loyverseID)
	return args.Error(0)
}

// MockLoyverseClient
type MockLoyverseClient struct {
	mock.Mock
}

func (m *MockLoyverseClient) CreateCustomer(ctx context.Context, customer *domain.Customer) (*string, error) {
	args := m.Called(ctx, customer)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*string), args.Error(1)
}

func (m *MockLoyverseClient) GetCustomer(ctx context.Context, loyverseID string) (*domain.Customer, error) {
	args := m.Called(ctx, loyverseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Customer), args.Error(1)
}

func (m *MockLoyverseClient) UpdateCustomer(ctx context.Context, loyverseID string, customer *domain.Customer) error {
	args := m.Called(ctx, loyverseID, customer)
	return args.Error(0)
}

func (m *MockLoyverseClient) SearchCustomerByEmail(ctx context.Context, email string) (*string, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*string), args.Error(1)
}

func (m *MockLoyverseClient) SearchCustomerByPhone(ctx context.Context, phone string) (*string, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*string), args.Error(1)
}

func TestCreateCustomer(t *testing.T) {
	// Create mocks
	customerRepo := new(MockCustomerRepository)
	customerAddressRepo := new(MockCustomerAddressRepository)
	thaiAddressRepo := new(MockThaiAddressRepository)
	deliveryRouteRepo := new(MockDeliveryRouteRepository)
	cacheRepo := new(MockCacheRepository)
	eventPublisher := new(MockEventPublisher)
	loyverseClient := new(MockLoyverseClient)

	logger := zap.NewNop()

	// Create service
	service := NewCustomerService(
		customerRepo,
		customerAddressRepo,
		thaiAddressRepo,
		deliveryRouteRepo,
		cacheRepo,
		eventPublisher,
		loyverseClient,
		logger,
	)

	// Create test customer
	customerID := uuid.New()
	customer := &domain.Customer{
		ID:        customerID,
		Email:     "test@example.com",
		Phone:     "0812345678",
		FirstName: "John",
		LastName:  "Doe",
		Tier:      domain.TierBronze,
		IsActive:  true,
	}

	// Setup expectations
	customerRepo.On("GetByEmail", mock.Anything, customer.Email).Return((*domain.Customer)(nil), domain.ErrCustomerNotFound)
	customerRepo.On("GetByPhone", mock.Anything, customer.Phone).Return((*domain.Customer)(nil), domain.ErrCustomerNotFound)
	customerRepo.On("Create", mock.Anything, customer).Return(nil)
	eventPublisher.On("PublishCustomerCreated", mock.Anything, customer).Return(nil)

	// Test
	ctx := context.Background()
	result, err := service.CreateCustomer(ctx, customer)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, customer, result)
	customerRepo.AssertExpectations(t)
	eventPublisher.AssertExpectations(t)
}
