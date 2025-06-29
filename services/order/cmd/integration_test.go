package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/saan/order-service/internal/application"
	"github.com/saan/order-service/internal/application/dto"
	"github.com/saan/order-service/internal/domain"
	"github.com/saan/order-service/internal/infrastructure/event"
	"github.com/saan/order-service/pkg/logger"
)

// MockEventRepository implements domain.OrderEventRepository for testing
type MockEventRepository struct {
	events []domain.OrderEvent
}

func NewMockEventRepository() *MockEventRepository {
	return &MockEventRepository{
		events: make([]domain.OrderEvent, 0),
	}
}

func (m *MockEventRepository) Create(ctx context.Context, event *domain.OrderEvent) error {
	m.events = append(m.events, *event)
	return nil
}

func (m *MockEventRepository) GetPendingEvents(ctx context.Context, limit int) ([]*domain.OrderEvent, error) {
	var pending []*domain.OrderEvent
	for i := range m.events {
		if m.events[i].Status == domain.EventStatusPending {
			pending = append(pending, &m.events[i])
			if len(pending) >= limit {
				break
			}
		}
	}
	return pending, nil
}

func (m *MockEventRepository) GetFailedEvents(ctx context.Context, maxRetries int, limit int) ([]*domain.OrderEvent, error) {
	var failed []*domain.OrderEvent
	for i := range m.events {
		if m.events[i].Status == domain.EventStatusFailed && m.events[i].RetryCount < maxRetries {
			failed = append(failed, &m.events[i])
			if len(failed) >= limit {
				break
			}
		}
	}
	return failed, nil
}

func (m *MockEventRepository) UpdateStatus(ctx context.Context, eventID uuid.UUID, status domain.EventStatus) error {
	for i := range m.events {
		if m.events[i].ID == eventID {
			m.events[i].Status = status
			return nil
		}
	}
	return fmt.Errorf("event not found")
}

func (m *MockEventRepository) MarkAsSent(ctx context.Context, eventID uuid.UUID) error {
	for i := range m.events {
		if m.events[i].ID == eventID {
			m.events[i].MarkAsSent()
			return nil
		}
	}
	return fmt.Errorf("event not found")
}

func (m *MockEventRepository) MarkAsFailed(ctx context.Context, eventID uuid.UUID) error {
	for i := range m.events {
		if m.events[i].ID == eventID {
			m.events[i].MarkAsFailed()
			return nil
		}
	}
	return fmt.Errorf("event not found")
}

func (m *MockEventRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.OrderEvent, error) {
	var orderEvents []*domain.OrderEvent
	for i := range m.events {
		if m.events[i].OrderID == orderID {
			orderEvents = append(orderEvents, &m.events[i])
		}
	}
	return orderEvents, nil
}

func (m *MockEventRepository) Delete(ctx context.Context, eventID uuid.UUID) error {
	for i := range m.events {
		if m.events[i].ID == eventID {
			m.events = append(m.events[:i], m.events[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("event not found")
}

// MockAuditRepository implements domain.AuditRepository for testing
type MockAuditRepository struct {
	logs []domain.AuditLog
}

func NewMockAuditRepository() *MockAuditRepository {
	return &MockAuditRepository{
		logs: make([]domain.AuditLog, 0),
	}
}

func (m *MockAuditRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	m.logs = append(m.logs, *log)
	return nil
}

func (m *MockAuditRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.AuditLog, error) {
	var orderLogs []*domain.AuditLog
	for i := range m.logs {
		if m.logs[i].OrderID == orderID {
			orderLogs = append(orderLogs, &m.logs[i])
		}
	}
	return orderLogs, nil
}

func (m *MockAuditRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.AuditLog, error) {
	var userLogs []*domain.AuditLog
	for i := range m.logs {
		if m.logs[i].UserID != nil && *m.logs[i].UserID == userID {
			userLogs = append(userLogs, &m.logs[i])
		}
	}
	return userLogs, nil
}

func (m *MockAuditRepository) GetByAction(ctx context.Context, action domain.AuditAction) ([]*domain.AuditLog, error) {
	var actionLogs []*domain.AuditLog
	for i := range m.logs {
		if m.logs[i].Action == action {
			actionLogs = append(actionLogs, &m.logs[i])
		}
	}
	return actionLogs, nil
}

func (m *MockAuditRepository) List(ctx context.Context, limit, offset int) ([]*domain.AuditLog, error) {
	var logs []*domain.AuditLog
	start := offset
	end := offset + limit
	if start > len(m.logs) {
		return logs, nil
	}
	if end > len(m.logs) {
		end = len(m.logs)
	}
	for i := start; i < end; i++ {
		logs = append(logs, &m.logs[i])
	}
	return logs, nil
}

// MockEventPublisher implements domain.EventPublisher for testing
type MockEventPublisher struct {
	publishedEvents []domain.OrderEvent
}

func NewMockEventPublisher() *MockEventPublisher {
	return &MockEventPublisher{
		publishedEvents: make([]domain.OrderEvent, 0),
	}
}

func (m *MockEventPublisher) PublishEvent(ctx context.Context, event *domain.OrderEvent) error {
	m.publishedEvents = append(m.publishedEvents, *event)
	return nil
}

// MockOrderRepository implements domain.OrderRepository for testing
type MockOrderRepository struct {
	orders map[uuid.UUID]*domain.Order
}

func NewMockOrderRepository() *MockOrderRepository {
	return &MockOrderRepository{
		orders: make(map[uuid.UUID]*domain.Order),
	}
}

func (m *MockOrderRepository) Create(ctx context.Context, order *domain.Order) error {
	m.orders[order.ID] = order
	return nil
}

func (m *MockOrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	order, exists := m.orders[id]
	if !exists {
		return nil, domain.ErrOrderNotFound
	}
	return order, nil
}

func (m *MockOrderRepository) Update(ctx context.Context, order *domain.Order) error {
	if _, exists := m.orders[order.ID]; !exists {
		return domain.ErrOrderNotFound
	}
	m.orders[order.ID] = order
	return nil
}

func (m *MockOrderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if _, exists := m.orders[id]; !exists {
		return domain.ErrOrderNotFound
	}
	delete(m.orders, id)
	return nil
}

func (m *MockOrderRepository) GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*domain.Order, error) {
	var customerOrders []*domain.Order
	for _, order := range m.orders {
		if order.CustomerID == customerID {
			customerOrders = append(customerOrders, order)
		}
	}
	return customerOrders, nil
}

func (m *MockOrderRepository) List(ctx context.Context, limit, offset int) ([]*domain.Order, error) {
	var orders []*domain.Order
	count := 0
	for _, order := range m.orders {
		if count < offset {
			count++
			continue
		}
		if len(orders) >= limit {
			break
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (m *MockOrderRepository) GetByStatus(ctx context.Context, status domain.OrderStatus) ([]*domain.Order, error) {
	var statusOrders []*domain.Order
	for _, order := range m.orders {
		if order.Status == status {
			statusOrders = append(statusOrders, order)
		}
	}
	return statusOrders, nil
}

// MockOrderItemRepository implements domain.OrderItemRepository for testing
type MockOrderItemRepository struct {
	items map[uuid.UUID][]*domain.OrderItem
}

func NewMockOrderItemRepository() *MockOrderItemRepository {
	return &MockOrderItemRepository{
		items: make(map[uuid.UUID][]*domain.OrderItem),
	}
}

func (m *MockOrderItemRepository) Create(ctx context.Context, item *domain.OrderItem) error {
	m.items[item.OrderID] = append(m.items[item.OrderID], item)
	return nil
}

func (m *MockOrderItemRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.OrderItem, error) {
	items, exists := m.items[orderID]
	if !exists {
		return []*domain.OrderItem{}, nil
	}
	return items, nil
}

func (m *MockOrderItemRepository) Update(ctx context.Context, item *domain.OrderItem) error {
	items, exists := m.items[item.OrderID]
	if !exists {
		return fmt.Errorf("order not found")
	}
	
	for i, existingItem := range items {
		if existingItem.ID == item.ID {
			items[i] = item
			return nil
		}
	}
	return fmt.Errorf("item not found")
}

func (m *MockOrderItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	for orderID, items := range m.items {
		for i, item := range items {
			if item.ID == id {
				m.items[orderID] = append(items[:i], items[i+1:]...)
				return nil
			}
		}
	}
	return fmt.Errorf("item not found")
}

// Integration test for order creation with audit and events
func TestOrderCreationIntegration(t *testing.T) {
	// Setup
	orderRepo := NewMockOrderRepository()
	orderItemRepo := NewMockOrderItemRepository()
	auditRepo := NewMockAuditRepository()
	eventRepo := NewMockEventRepository()
	eventPublisher := NewMockEventPublisher()
	logger := logger.NewLogger("info", "text")
	
	orderService := application.NewOrderService(
		orderRepo, orderItemRepo, auditRepo, eventRepo, eventPublisher, logger,
	)
	
	// Test data
	customerID := uuid.New()
	productID := uuid.New()
	
	req := &dto.CreateOrderRequest{
		CustomerID:      customerID,
		ShippingAddress: "123 Test Street",
		BillingAddress:  "123 Test Street",
		Notes:          "Test order",
		Items: []dto.CreateOrderItemRequest{
			{
				ProductID: productID,
				Quantity:  2,
				UnitPrice: 100.50,
			},
		},
	}
	
	// Execute
	ctx := context.Background()
	response, err := orderService.CreateOrder(ctx, req)
	
	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if response == nil {
		t.Fatal("Expected response, got nil")
	}
	
	// Verify order was created
	if len(orderRepo.orders) != 1 {
		t.Errorf("Expected 1 order, got %d", len(orderRepo.orders))
	}
	
	// Verify order items were created
	items, err := orderItemRepo.GetByOrderID(ctx, response.ID)
	if err != nil {
		t.Fatalf("Error getting order items: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("Expected 1 order item, got %d", len(items))
	}
	
	// Verify audit log was created
	auditLogs, err := auditRepo.GetByOrderID(ctx, response.ID)
	if err != nil {
		t.Fatalf("Error getting audit logs: %v", err)
	}
	if len(auditLogs) != 1 {
		t.Errorf("Expected 1 audit log, got %d", len(auditLogs))
	}
	if auditLogs[0].Action != domain.AuditActionCreate {
		t.Errorf("Expected audit action CREATE, got %s", auditLogs[0].Action)
	}
	
	// Verify event was created
	events, err := eventRepo.GetByOrderID(ctx, response.ID)
	if err != nil {
		t.Fatalf("Error getting events: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
	if events[0].EventType != domain.EventTypeOrderCreated {
		t.Errorf("Expected event type OrderCreated, got %s", events[0].EventType)
	}
	
	t.Logf("✅ Order created successfully with ID: %s", response.ID)
	t.Logf("✅ Audit log created with action: %s", auditLogs[0].Action)
	t.Logf("✅ Event created with type: %s", events[0].EventType)
}

// Integration test for order status update with audit and events
func TestOrderStatusUpdateIntegration(t *testing.T) {
	// Setup
	orderRepo := NewMockOrderRepository()
	orderItemRepo := NewMockOrderItemRepository()
	auditRepo := NewMockAuditRepository()
	eventRepo := NewMockEventRepository()
	eventPublisher := NewMockEventPublisher()
	logger := logger.NewLogger("info", "text")
	
	orderService := application.NewOrderService(
		orderRepo, orderItemRepo, auditRepo, eventRepo, eventPublisher, logger,
	)
	
	// Create an order first
	customerID := uuid.New()
	order := domain.NewOrder(customerID, "123 Test St", "123 Test St", "Test")
	err := orderRepo.Create(context.Background(), order)
	if err != nil {
		t.Fatalf("Error creating test order: %v", err)
	}
	
	// Test status update
	ctx := context.Background()
	req := &dto.UpdateOrderStatusRequest{
		Status: domain.OrderStatusConfirmed,
	}
	
	// Execute
	response, err := orderService.UpdateOrderStatus(ctx, order.ID, req)
	
	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if response.Status != domain.OrderStatusConfirmed {
		t.Errorf("Expected status confirmed, got %s", response.Status)
	}
	
	// Verify audit log was created for status change
	auditLogs, err := auditRepo.GetByOrderID(ctx, order.ID)
	if err != nil {
		t.Fatalf("Error getting audit logs: %v", err)
	}
	if len(auditLogs) != 1 {
		t.Errorf("Expected 1 audit log, got %d", len(auditLogs))
	}
	if auditLogs[0].Action != domain.AuditActionStatusChange {
		t.Errorf("Expected audit action CHANGE_STATUS, got %s", auditLogs[0].Action)
	}
	
	// Verify event was created for status change
	events, err := eventRepo.GetByOrderID(ctx, order.ID)
	if err != nil {
		t.Fatalf("Error getting events: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
	if events[0].EventType != domain.EventTypeOrderConfirmed {
		t.Errorf("Expected event type OrderConfirmed, got %s", events[0].EventType)
	}
	
	t.Logf("✅ Order status updated successfully to: %s", response.Status)
	t.Logf("✅ Audit log created for status change")
	t.Logf("✅ Event created with type: %s", events[0].EventType)
}

// Integration test for outbox worker processing
func TestOutboxWorkerIntegration(t *testing.T) {
	// Setup
	eventRepo := NewMockEventRepository()
	eventPublisher := NewMockEventPublisher()
	logger := logger.NewLogger("info", "text")
	
	// Create test events
	ctx := context.Background()
	orderID := uuid.New()
	
	event1 := domain.NewOrderEvent(orderID, domain.EventTypeOrderCreated, map[string]interface{}{
		"order_id": orderID,
		"status":   "pending",
	})
	event2 := domain.NewOrderEvent(orderID, domain.EventTypeOrderConfirmed, map[string]interface{}{
		"order_id": orderID,
		"status":   "confirmed",
	})
	
	err := eventRepo.Create(ctx, event1)
	if err != nil {
		t.Fatalf("Error creating test event 1: %v", err)
	}
	err = eventRepo.Create(ctx, event2)
	if err != nil {
		t.Fatalf("Error creating test event 2: %v", err)
	}
	
	// Setup outbox worker
	config := event.OutboxWorkerConfig{
		PollingInterval: 100 * time.Millisecond,
		BatchSize:       10,
		MaxRetries:      3,
		RetryDelay:      1 * time.Second,
	}
	
	worker := event.NewOutboxWorker(eventRepo, eventPublisher, config, logger)
	
	// Start worker for a short time
	workerCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	
	worker.Start(workerCtx)
	time.Sleep(300 * time.Millisecond) // Let it process events
	worker.Stop()
	
	// Verify events were published
	if len(eventPublisher.publishedEvents) != 2 {
		t.Errorf("Expected 2 published events, got %d", len(eventPublisher.publishedEvents))
	}
	
	// Verify events were marked as sent
	events, err := eventRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		t.Fatalf("Error getting events: %v", err)
	}
	
	sentCount := 0
	for _, event := range events {
		if event.Status == domain.EventStatusSent {
			sentCount++
		}
	}
	
	if sentCount != 2 {
		t.Errorf("Expected 2 events marked as sent, got %d", sentCount)
	}
	
	t.Logf("✅ Outbox worker processed %d events successfully", sentCount)
	t.Logf("✅ Events published: %d", len(eventPublisher.publishedEvents))
}

// Test the health check endpoint
func TestHealthCheckEndpoint(t *testing.T) {
	// This would require setting up the HTTP server
	// For now, we'll just verify the HTTP status
	resp, err := http.Get("http://localhost:8081/health")
	if err != nil {
		t.Skip("Skipping health check test - service not running")
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	
	t.Logf("✅ Health check endpoint responded with status: %d", resp.StatusCode)
}
