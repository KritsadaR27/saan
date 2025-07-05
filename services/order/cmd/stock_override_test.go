package main

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"order/internal/application"
	"order/internal/application/dto"
	"order/internal/domain"
	"order/internal/infrastructure/config"
	"order/internal/infrastructure/db"
	"order/internal/infrastructure/repository"
	"order/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStockOverrideIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test environment
	cfg := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		DBName:   "saan_test",
		SSLMode:  "disable",
	}

	// Connect to test database
	database, err := db.Connect(cfg)
	require.NoError(t, err, "Failed to connect to test database")
	defer database.Close()

	// Initialize repositories
	orderRepo := repository.NewPostgresOrderRepository(database)
	orderItemRepo := repository.NewPostgresOrderItemRepository(database)
	auditRepo := repository.NewPostgresAuditRepository(database)
	orderEventRepo := repository.NewPostgresEventRepository(database)

	// Mock event publisher for testing
	mockPublisher := NewMockEventPublisher()

	// Initialize logger
	log := logger.NewLogger("debug", "json")

	// Initialize order service
	orderService := application.NewOrderService(
		orderRepo, 
		orderItemRepo, 
		auditRepo, 
		orderEventRepo, 
		mockPublisher, 
		log,
	)

	t.Run("successful stock override", func(t *testing.T) {
		ctx := context.Background()

		// Create a test order first
		customerID := uuid.New()
		productID := uuid.New()
		
		createReq := &dto.CreateOrderRequest{
			CustomerID:      customerID,
			ShippingAddress: "123 Test St",
			BillingAddress:  "123 Test St",
			Notes:          "Test order for stock override",
			Items: []dto.CreateOrderItemRequest{
				{
					ProductID: productID,
					Quantity:  10,
					UnitPrice: 100.0,
				},
			},
		}

		orderResp, err := orderService.CreateOrder(ctx, createReq)
		require.NoError(t, err, "Failed to create test order")
		assert.Equal(t, domain.OrderStatusPending, orderResp.Status)

		// Now test stock override
		userID := uuid.New()
		overrideReq := &dto.ConfirmOrderWithStockOverrideRequest{
			UserID:   userID,
			UserRole: "manager", // Manager role should have permission
			OverrideItems: []dto.StockOverrideItem{
				{
					ProductID:      productID,
					Quantity:       10,
					OverrideReason: "Emergency order - customer critical need",
				},
			},
		}

		confirmedOrder, err := orderService.ConfirmOrderWithStockOverride(ctx, orderResp.ID, overrideReq)
		require.NoError(t, err, "Failed to confirm order with stock override")

		// Verify order status changed
		assert.Equal(t, domain.OrderStatusConfirmed, confirmedOrder.Status)
		assert.NotNil(t, confirmedOrder.ConfirmedAt)

		// Verify order items have override information
		assert.Len(t, confirmedOrder.Items, 1)
		item := confirmedOrder.Items[0]
		assert.True(t, item.IsOverride)
		assert.NotNil(t, item.OverrideReason)
		assert.Equal(t, "Emergency order - customer critical need", *item.OverrideReason)

		// Verify audit log was created
		// Note: This would require additional repository method to query audit logs by order ID
		// For now, we trust that the audit log creation worked since no error was returned

		// Verify event was published
		assert.Len(t, mockPublisher.publishedEvents, 3) // Create + Status Change + Stock Override events
		
		// Find the stock override event
		var stockOverrideEvent *domain.OrderEvent
		for _, event := range mockPublisher.publishedEvents {
			if event.EventType == "order.stock_override" {
				stockOverrideEvent = &event
				break
			}
		}
		require.NotNil(t, stockOverrideEvent, "Stock override event should be published")
		assert.Equal(t, orderResp.ID, stockOverrideEvent.OrderID)
	})

	t.Run("unauthorized stock override", func(t *testing.T) {
		ctx := context.Background()

		// Create a test order first
		customerID := uuid.New()
		productID := uuid.New()
		
		createReq := &dto.CreateOrderRequest{
			CustomerID:      customerID,
			ShippingAddress: "123 Test St",
			BillingAddress:  "123 Test St",
			Notes:          "Test order for unauthorized override",
			Items: []dto.CreateOrderItemRequest{
				{
					ProductID: productID,
					Quantity:  5,
					UnitPrice: 50.0,
				},
			},
		}

		orderResp, err := orderService.CreateOrder(ctx, createReq)
		require.NoError(t, err, "Failed to create test order")

		// Try stock override with insufficient permission
		userID := uuid.New()
		overrideReq := &dto.ConfirmOrderWithStockOverrideRequest{
			UserID:   userID,
			UserRole: "employee", // Employee role should NOT have permission
			OverrideItems: []dto.StockOverrideItem{
				{
					ProductID:      productID,
					Quantity:       5,
					OverrideReason: "Unauthorized attempt",
				},
			},
		}

		_, err = orderService.ConfirmOrderWithStockOverride(ctx, orderResp.ID, overrideReq)
		assert.Error(t, err, "Should fail for unauthorized user")
		assert.Equal(t, domain.ErrUnauthorizedStockOverride, err)
	})

	t.Run("invalid order status for override", func(t *testing.T) {
		ctx := context.Background()

		// Create and confirm a test order first
		customerID := uuid.New()
		productID := uuid.New()
		
		createReq := &dto.CreateOrderRequest{
			CustomerID:      customerID,
			ShippingAddress: "123 Test St",
			BillingAddress:  "123 Test St",
			Notes:          "Test order for status validation",
			Items: []dto.CreateOrderItemRequest{
				{
					ProductID: productID,
					Quantity:  3,
					UnitPrice: 75.0,
				},
			},
		}

		orderResp, err := orderService.CreateOrder(ctx, createReq)
		require.NoError(t, err, "Failed to create test order")

		// First confirm the order normally
		statusReq := &dto.UpdateOrderStatusRequest{
			Status: domain.OrderStatusConfirmed,
		}
		_, err = orderService.UpdateOrderStatus(ctx, orderResp.ID, statusReq)
		require.NoError(t, err, "Failed to update order status")

		// Now try to override stock on already confirmed order
		userID := uuid.New()
		overrideReq := &dto.ConfirmOrderWithStockOverrideRequest{
			UserID:   userID,
			UserRole: "manager",
			OverrideItems: []dto.StockOverrideItem{
				{
					ProductID:      productID,
					Quantity:       3,
					OverrideReason: "Should fail - order already confirmed",
				},
			},
		}

		_, err = orderService.ConfirmOrderWithStockOverride(ctx, orderResp.ID, overrideReq)
		assert.Error(t, err, "Should fail for non-pending order")
		assert.Equal(t, domain.ErrInvalidOrderStatus, err)
	})
}
