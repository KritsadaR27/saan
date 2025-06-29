package main

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/saan/order-service/internal/application"
	"github.com/saan/order-service/internal/application/template"
	"github.com/saan/order-service/internal/domain"
	"github.com/saan/order-service/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChatOrderIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chat integration test in short mode")
	}

	// Setup mock dependencies
	mockOrderService := &MockOrderService{}
	mockCustomerClient := &MockCustomerClient{}
	mockInventoryClient := &MockInventoryClient{}
	mockNotificationClient := &MockNotificationClient{}
	templateSelector := template.NewTemplateSelector()
	log := logger.NewLogger("debug", "json")

	// Initialize chat order service
	chatOrderService := application.NewChatOrderService(
		mockOrderService,
		mockCustomerClient,
		mockInventoryClient,
		mockNotificationClient,
		templateSelector,
		log,
	)

	t.Run("successful chat order creation", func(t *testing.T) {
		ctx := context.Background()
		chatID := "chat_123"
		customerID := uuid.New()
		productID := uuid.New()

		// Setup mock responses
		mockCustomerClient.SetCustomer(customerID, "John", "Doe", "john@example.com")
		mockInventoryClient.SetProduct(productID, "Test Product", 100.0, true)
		mockInventoryClient.SetStock(productID, 10, true)

		req := &application.ChatOrderRequest{
			ChatID:     chatID,
			CustomerID: customerID.String(),
			Items: []application.ChatOrderItem{
				{
					ProductName: "Test Product",
					ProductID:   &productID,
					Quantity:    2,
					UnitPrice:   &[]float64{100.0}[0],
				},
			},
			PaymentMethod:  &[]string{"cash"}[0],
			DeliveryMethod: &[]string{"delivery"}[0],
			Notes:         "Test order from chat",
		}

		order, err := chatOrderService.CreateOrderFromChat(ctx, req)
		require.NoError(t, err, "Failed to create order from chat")
		assert.NotNil(t, order)
		assert.Equal(t, domain.OrderStatusPending, order.Status)
		assert.Len(t, order.Items, 1)
		assert.Equal(t, 200.0, order.TotalAmount) // 2 items * 100 each

		// Verify notification was sent
		assert.True(t, mockNotificationClient.MessageSent)
		assert.Equal(t, chatID, mockNotificationClient.LastChatID)
		assert.Contains(t, mockNotificationClient.LastMessage, "สรุปออร์เดอร์")
	})

	t.Run("chat order confirmation", func(t *testing.T) {
		ctx := context.Background()
		chatID := "chat_456"
		orderID := uuid.New()

		// Setup mock order
		mockOrderService.SetOrderStatus(orderID, domain.OrderStatusPending)

		order, err := chatOrderService.ConfirmChatOrder(ctx, chatID, orderID)
		require.NoError(t, err, "Failed to confirm chat order")
		assert.NotNil(t, order)
		assert.Equal(t, domain.OrderStatusConfirmed, order.Status)

		// Verify confirmation notification was sent
		assert.True(t, mockNotificationClient.MessageSent)
		assert.Equal(t, chatID, mockNotificationClient.LastChatID)
		assert.Contains(t, mockNotificationClient.LastMessage, "ยืนยันออร์เดอร์เรียบร้อยแล้ว")
	})

	t.Run("chat order cancellation", func(t *testing.T) {
		ctx := context.Background()
		chatID := "chat_789"
		orderID := uuid.New()
		reason := "Customer changed mind"

		// Setup mock order
		mockOrderService.SetOrderStatus(orderID, domain.OrderStatusPending)

		err := chatOrderService.CancelChatOrder(ctx, chatID, orderID, reason)
		require.NoError(t, err, "Failed to cancel chat order")

		// Verify order was cancelled
		assert.Equal(t, domain.OrderStatusCancelled, mockOrderService.GetOrderStatus(orderID))

		// Verify cancellation notification was sent
		assert.True(t, mockNotificationClient.MessageSent)
		assert.Equal(t, chatID, mockNotificationClient.LastChatID)
		assert.Contains(t, mockNotificationClient.LastMessage, "ยกเลิกออร์เดอร์แล้ว")
		assert.Contains(t, mockNotificationClient.LastMessage, reason)
	})
}

// Mock implementations would be defined here
// (Simplified for this example)
