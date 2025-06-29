package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/saan/order-service/internal/application/dto"
	"github.com/saan/order-service/internal/domain"
	"github.com/saan/order-service/pkg/logger"
)

// Mock implementations for testing
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Create(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockOrderRepository) Update(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*domain.Order, error) {
	args := m.Called(ctx, customerID)
	return args.Get(0).([]*domain.Order), args.Error(1)
}

func (m *MockOrderRepository) GetByStatus(ctx context.Context, status domain.OrderStatus) ([]*domain.Order, error) {
	args := m.Called(ctx, status)
	return args.Get(0).([]*domain.Order), args.Error(1)
}

func (m *MockOrderRepository) GetOrdersByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*domain.Order, error) {
	args := m.Called(ctx, startDate, endDate)
	return args.Get(0).([]*domain.Order), args.Error(1)
}

func (m *MockOrderRepository) GetOrdersByCustomer(ctx context.Context, customerID uuid.UUID) ([]*domain.Order, error) {
	args := m.Called(ctx, customerID)
	return args.Get(0).([]*domain.Order), args.Error(1)
}

func (m *MockOrderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrderRepository) List(ctx context.Context, limit, offset int) ([]*domain.Order, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*domain.Order), args.Error(1)
}

type MockOrderItemRepository struct {
	mock.Mock
}

func (m *MockOrderItemRepository) Create(ctx context.Context, orderItem *domain.OrderItem) error {
	args := m.Called(ctx, orderItem)
	return args.Error(0)
}

func (m *MockOrderItemRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.OrderItem, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).([]*domain.OrderItem), args.Error(1)
}

func (m *MockOrderItemRepository) Update(ctx context.Context, orderItem *domain.OrderItem) error {
	args := m.Called(ctx, orderItem)
	return args.Error(0)
}

func (m *MockOrderItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrderItemRepository) GetByProductID(ctx context.Context, productID uuid.UUID) ([]*domain.OrderItem, error) {
	args := m.Called(ctx, productID)
	return args.Get(0).([]*domain.OrderItem), args.Error(1)
}

func (m *MockOrderItemRepository) GetAllOrderItems(ctx context.Context) ([]*domain.OrderItem, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.OrderItem), args.Error(1)
}

type MockOrderAuditRepository struct {
	mock.Mock
}

func (m *MockOrderAuditRepository) Create(ctx context.Context, auditLog *domain.OrderAuditLog) error {
	args := m.Called(ctx, auditLog)
	return args.Error(0)
}

func (m *MockOrderAuditRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.OrderAuditLog, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).([]*domain.OrderAuditLog), args.Error(1)
}

func (m *MockOrderAuditRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.OrderAuditLog, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*domain.OrderAuditLog), args.Error(1)
}

func (m *MockOrderAuditRepository) GetByAction(ctx context.Context, action domain.AuditAction) ([]*domain.OrderAuditLog, error) {
	args := m.Called(ctx, action)
	return args.Get(0).([]*domain.OrderAuditLog), args.Error(1)
}

func (m *MockOrderAuditRepository) List(ctx context.Context, limit, offset int) ([]*domain.OrderAuditLog, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*domain.OrderAuditLog), args.Error(1)
}

type MockOrderEventRepository struct {
	mock.Mock
}

func (m *MockOrderEventRepository) Create(ctx context.Context, event *domain.OrderEventOutbox) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockOrderEventRepository) GetPendingEvents(ctx context.Context, limit int) ([]*domain.OrderEventOutbox, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*domain.OrderEventOutbox), args.Error(1)
}

func (m *MockOrderEventRepository) GetFailedEvents(ctx context.Context, maxRetries int, limit int) ([]*domain.OrderEventOutbox, error) {
	args := m.Called(ctx, maxRetries, limit)
	return args.Get(0).([]*domain.OrderEventOutbox), args.Error(1)
}

func (m *MockOrderEventRepository) UpdateStatus(ctx context.Context, eventID uuid.UUID, status domain.EventStatus) error {
	args := m.Called(ctx, eventID, status)
	return args.Error(0)
}

func (m *MockOrderEventRepository) MarkAsSent(ctx context.Context, eventID uuid.UUID) error {
	args := m.Called(ctx, eventID)
	return args.Error(0)
}

func (m *MockOrderEventRepository) MarkAsFailed(ctx context.Context, eventID uuid.UUID) error {
	args := m.Called(ctx, eventID)
	return args.Error(0)
}

func (m *MockOrderEventRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.OrderEventOutbox, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).([]*domain.OrderEventOutbox), args.Error(1)
}

func (m *MockOrderEventRepository) Delete(ctx context.Context, eventID uuid.UUID) error {
	args := m.Called(ctx, eventID)
	return args.Error(0)
}

type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) PublishEvent(ctx context.Context, event *domain.OrderEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockLogger struct{}

func (l *MockLogger) Debug(args ...interface{})                             {}
func (l *MockLogger) Info(args ...interface{})                              {}
func (l *MockLogger) Warn(args ...interface{})                              {}
func (l *MockLogger) Error(args ...interface{})                             {}
func (l *MockLogger) Fatal(args ...interface{})                             {}
func (l *MockLogger) Debugf(format string, args ...interface{})             {}
func (l *MockLogger) Infof(format string, args ...interface{})              {}
func (l *MockLogger) Warnf(format string, args ...interface{})              {}
func (l *MockLogger) Errorf(format string, args ...interface{})             {}
func (l *MockLogger) Fatalf(format string, args ...interface{})             {}
func (l *MockLogger) WithField(key string, value interface{}) logger.Logger { return l }
func (l *MockLogger) WithFields(fields map[string]interface{}) logger.Logger { return l }

// Test Suite
type OrderServiceTestSuite struct {
	suite.Suite
	orderService         *OrderService
	mockOrderRepo        *MockOrderRepository
	mockOrderItemRepo    *MockOrderItemRepository
	mockAuditRepo        *MockOrderAuditRepository
	mockEventRepo        *MockOrderEventRepository
	mockEventPublisher   *MockEventPublisher
	logger               logger.Logger
}

func (suite *OrderServiceTestSuite) SetupTest() {
	suite.mockOrderRepo = new(MockOrderRepository)
	suite.mockOrderItemRepo = new(MockOrderItemRepository)
	suite.mockAuditRepo = new(MockOrderAuditRepository)
	suite.mockEventRepo = new(MockOrderEventRepository)
	suite.mockEventPublisher = new(MockEventPublisher)
	suite.logger = &MockLogger{}

	suite.orderService = NewOrderService(
		suite.mockOrderRepo,
		suite.mockOrderItemRepo,
		suite.mockAuditRepo,
		suite.mockEventRepo,
		suite.mockEventPublisher,
		suite.logger,
	)
}

func (suite *OrderServiceTestSuite) TearDownTest() {
	suite.mockOrderRepo.AssertExpectations(suite.T())
	suite.mockOrderItemRepo.AssertExpectations(suite.T())
	suite.mockAuditRepo.AssertExpectations(suite.T())
	suite.mockEventRepo.AssertExpectations(suite.T())
	suite.mockEventPublisher.AssertExpectations(suite.T())
}

// Test Order Creation Flow
func (suite *OrderServiceTestSuite) TestCreateOrder_Success() {
	// Arrange
	ctx := context.Background()
	customerID := uuid.New()
	createOrderRequest := &dto.CreateOrderRequest{
		CustomerID:      customerID,
		ShippingAddress: "123 Main St",
		BillingAddress:  "123 Main St",
		Notes:           "Test order",
		Items: []dto.CreateOrderItemRequest{
			{
				ProductID: uuid.New(),
				Quantity:  2,
				UnitPrice: 10.99,
			},
		},
	}
	// Setup mocks
	suite.mockOrderRepo.On("Create", ctx, mock.AnythingOfType("*domain.Order")).Return(nil)
	suite.mockOrderItemRepo.On("Create", ctx, mock.AnythingOfType("*domain.OrderItem")).Return(nil)
	suite.mockAuditRepo.On("Create", ctx, mock.AnythingOfType("*domain.OrderAuditLog")).Return(nil)
	suite.mockEventRepo.On("Create", ctx, mock.AnythingOfType("*domain.OrderEventOutbox")).Return(nil)
	suite.mockEventPublisher.On("PublishEvent", ctx, mock.AnythingOfType("*domain.OrderEvent")).Return(nil)

	// Act
	result, err := suite.orderService.CreateOrder(ctx, createOrderRequest)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), customerID, result.CustomerID)
	assert.Equal(suite.T(), domain.OrderStatusPending, result.Status)
	assert.Equal(suite.T(), float64(21.98), result.TotalAmount)
}

func (suite *OrderServiceTestSuite) TestCreateOrder_EmptyItems() {
	// Arrange
	ctx := context.Background()
	createOrderRequest := &dto.CreateOrderRequest{
		CustomerID:      uuid.New(),
		ShippingAddress: "123 Main St",
		BillingAddress:  "123 Main St",
		Notes:           "Test order",
		Items:           []dto.CreateOrderItemRequest{}, // Empty items
	}

	// Act
	result, err := suite.orderService.CreateOrder(ctx, createOrderRequest)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "order must have at least one item")
}

// Test Status Transitions
func (suite *OrderServiceTestSuite) TestUpdateOrderStatus_ValidTransition() {
	// Arrange
	ctx := context.Background()
	orderID := uuid.New()
	existingOrder := &domain.Order{
		ID:          orderID,
		CustomerID:  uuid.New(),
		Status:      domain.OrderStatusPending,
		TotalAmount: 100.0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Setup mocks
	suite.mockOrderRepo.On("GetByID", ctx, orderID).Return(existingOrder, nil)
	suite.mockOrderRepo.On("Update", ctx, mock.AnythingOfType("*domain.Order")).Return(nil)
	suite.mockAuditRepo.On("Create", ctx, mock.AnythingOfType("*domain.OrderAuditLog")).Return(nil)
	suite.mockEventRepo.On("Create", ctx, mock.AnythingOfType("*domain.OrderEventOutbox")).Return(nil)
	suite.mockEventPublisher.On("PublishEvent", ctx, mock.AnythingOfType("*domain.OrderEvent")).Return(nil)

	// Act
	_, err := suite.orderService.UpdateOrderStatus(ctx, orderID, &dto.UpdateOrderStatusRequest{
		Status: domain.OrderStatusConfirmed,
	})

	// Assert
	assert.NoError(suite.T(), err)
}

func (suite *OrderServiceTestSuite) TestUpdateOrderStatus_InvalidTransition() {
	// Arrange
	ctx := context.Background()
	orderID := uuid.New()
	existingOrder := &domain.Order{
		ID:          orderID,
		CustomerID:  uuid.New(),
		Status:      domain.OrderStatusDelivered, // Already delivered
		TotalAmount: 100.0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Setup mocks
	suite.mockOrderRepo.On("GetByID", ctx, orderID).Return(existingOrder, nil)

	// Act
	_, err := suite.orderService.UpdateOrderStatus(ctx, orderID, &dto.UpdateOrderStatusRequest{
		Status: domain.OrderStatusPending,
	})

	// Assert
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid status transition")
}

// Test Stock Override
func (suite *OrderServiceTestSuite) TestCreateOrderWithStockOverride_Success() {
	// Arrange
	ctx := context.Background()
	customerID := uuid.New()
	createOrderRequest := &dto.CreateOrderRequest{
		CustomerID:      customerID,
		ShippingAddress: "123 Main St",
		BillingAddress:  "123 Main St",
		Notes:           "Test order with stock override",
		Items: []dto.CreateOrderItemRequest{
			{
				ProductID: uuid.New(),
				Quantity:  10,
				UnitPrice: 15.99,
			},
		},
	}

	// Setup mocks
	suite.mockOrderRepo.On("Create", ctx, mock.AnythingOfType("*domain.Order")).Return(nil)
	suite.mockOrderItemRepo.On("Create", ctx, mock.AnythingOfType("*domain.OrderItem")).Return(nil)
	suite.mockAuditRepo.On("Create", ctx, mock.AnythingOfType("*domain.OrderAuditLog")).Return(nil)
	suite.mockEventRepo.On("Create", ctx, mock.AnythingOfType("*domain.OrderEventOutbox")).Return(nil)
	suite.mockEventPublisher.On("PublishEvent", ctx, mock.AnythingOfType("*domain.OrderEvent")).Return(nil)

	// Act
	result, err := suite.orderService.CreateOrder(ctx, createOrderRequest)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), float64(159.9), result.TotalAmount)
}

// Test Event Publishing
func (suite *OrderServiceTestSuite) TestCreateOrder_EventPublishing() {
	// Arrange
	ctx := context.Background()
	customerID := uuid.New()
	createOrderRequest := &dto.CreateOrderRequest{
		CustomerID:      customerID,
		ShippingAddress: "123 Main St",
		BillingAddress:  "123 Main St",
		Notes:           "Test order",
		Items: []dto.CreateOrderItemRequest{
			{
				ProductID: uuid.New(),
				Quantity:  1,
				UnitPrice: 19.99,
			},
		},
	}

	// Setup mocks
	suite.mockOrderRepo.On("Create", ctx, mock.AnythingOfType("*domain.Order")).Return(nil)
	suite.mockOrderItemRepo.On("Create", ctx, mock.AnythingOfType("*domain.OrderItem")).Return(nil)
	suite.mockAuditRepo.On("Create", ctx, mock.AnythingOfType("*domain.OrderAuditLog")).Return(nil)
	suite.mockEventRepo.On("Create", ctx, mock.AnythingOfType("*domain.OrderEventOutbox")).Return(nil)
	
	// Verify event publishing is called with correct event type
	suite.mockEventPublisher.On("PublishEvent", ctx, mock.MatchedBy(func(event *domain.OrderEvent) bool {
		return event.EventType == "order.created"
	})).Return(nil)

	// Act
	result, err := suite.orderService.CreateOrder(ctx, createOrderRequest)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
}

func (suite *OrderServiceTestSuite) TestGetOrderByID_Success() {
	// Arrange
	ctx := context.Background()
	orderID := uuid.New()
	expectedOrder := &domain.Order{
		ID:          orderID,
		CustomerID:  uuid.New(),
		Status:      domain.OrderStatusPending,
		TotalAmount: 50.0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Setup mocks
	suite.mockOrderRepo.On("GetByID", ctx, orderID).Return(expectedOrder, nil)

	// Act
	result, err := suite.orderService.GetOrderByID(ctx, orderID)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), orderID, result.ID)
	assert.Equal(suite.T(), expectedOrder.Status, result.Status)
}

func (suite *OrderServiceTestSuite) TestGetOrderByID_NotFound() {
	// Arrange
	ctx := context.Background()
	orderID := uuid.New()

	// Setup mocks
	suite.mockOrderRepo.On("GetByID", ctx, orderID).Return((*domain.Order)(nil), domain.ErrOrderNotFound)

	// Act
	result, err := suite.orderService.GetOrderByID(ctx, orderID)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), domain.ErrOrderNotFound, err)
}

// Run the test suite
func TestOrderServiceSuite(t *testing.T) {
	suite.Run(t, new(OrderServiceTestSuite))
}
