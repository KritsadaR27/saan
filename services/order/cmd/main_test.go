package main

import (
	"testing"
	"order/internal/domain"
	"github.com/google/uuid"
)

func TestOrderCreation(t *testing.T) {
	customerID := uuid.New()
	order := domain.NewOrder(customerID, "123 Main St", "123 Main St", "Test order")
	
	if order.ID == uuid.Nil {
		t.Error("Order ID should not be nil")
	}
	
	if order.CustomerID != customerID {
		t.Error("Customer ID should match")
	}
	
	if order.Status != domain.OrderStatusPending {
		t.Error("Initial status should be pending")
	}
	
	if order.TotalAmount != 0 {
		t.Error("Initial total amount should be 0")
	}
}

func TestOrderItemAddition(t *testing.T) {
	customerID := uuid.New()
	productID := uuid.New()
	order := domain.NewOrder(customerID, "123 Main St", "123 Main St", "Test order")
	
	order.AddItem(productID, 2, 29.99)
	
	if len(order.Items) != 1 {
		t.Error("Should have 1 item")
	}
	
	if order.TotalAmount != 59.98 {
		t.Errorf("Total amount should be 59.98, got %f", order.TotalAmount)
	}
}

func TestOrderStatusTransition(t *testing.T) {
	customerID := uuid.New()
	order := domain.NewOrder(customerID, "123 Main St", "123 Main St", "Test order")
	
	// Valid transition
	err := order.UpdateStatus(domain.OrderStatusConfirmed)
	if err != nil {
		t.Errorf("Should allow transition from pending to confirmed: %v", err)
	}
	
	// Invalid transition
	err = order.UpdateStatus(domain.OrderStatusDelivered)
	if err == nil {
		t.Error("Should not allow transition from confirmed to delivered")
	}
}
