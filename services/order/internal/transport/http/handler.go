package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/go-playground/validator/v10"
	"github.com/saan/order-service/internal/application"
	"github.com/saan/order-service/internal/application/dto"
	"github.com/saan/order-service/internal/domain"
	"github.com/saan/order-service/pkg/logger"
)

// OrderHandler handles HTTP requests for orders
type OrderHandler struct {
	orderService *application.OrderService
	validator    *validator.Validate
	logger       logger.Logger
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(orderService *application.OrderService, logger logger.Logger) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		validator:    validator.New(),
		logger:       logger,
	}
}

// CreateOrder handles POST /orders
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithField("error", err).Error("Failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.logger.WithField("error", err).Error("Request validation failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	order, err := h.orderService.CreateOrder(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithField("error", err).Error("Failed to create order")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	h.logger.WithField("order_id", order.ID).Info("Order created successfully")
	c.JSON(http.StatusCreated, order)
}

// GetOrder handles GET /orders/:id
func (h *OrderHandler) GetOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	order, err := h.orderService.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		if err == domain.ErrOrderNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		h.logger.WithField("error", err).Error("Failed to get order")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get order"})
		return
	}

	c.JSON(http.StatusOK, order)
}

// GetOrdersByCustomer handles GET /customers/:customerId/orders
func (h *OrderHandler) GetOrdersByCustomer(c *gin.Context) {
	customerIDStr := c.Param("customerId")
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	orders, err := h.orderService.GetOrdersByCustomerID(c.Request.Context(), customerID)
	if err != nil {
		h.logger.WithField("error", err).Error("Failed to get orders by customer")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

// UpdateOrder handles PUT /orders/:id
func (h *OrderHandler) UpdateOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req dto.UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithField("error", err).Error("Failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	order, err := h.orderService.UpdateOrder(c.Request.Context(), id, &req)
	if err != nil {
		if err == domain.ErrOrderNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		if err == domain.ErrOrderCannotBeModified {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Order cannot be modified in current status"})
			return
		}
		h.logger.WithField("error", err).Error("Failed to update order")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order"})
		return
	}

	h.logger.WithField("order_id", id).Info("Order updated successfully")
	c.JSON(http.StatusOK, order)
}

// UpdateOrderStatus handles PATCH /orders/:id/status
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req dto.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithField("error", err).Error("Failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.logger.WithField("error", err).Error("Request validation failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	order, err := h.orderService.UpdateOrderStatus(c.Request.Context(), id, &req)
	if err != nil {
		if err == domain.ErrOrderNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		if err == domain.ErrInvalidStatusTransition {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status transition"})
			return
		}
		h.logger.WithField("error", err).Error("Failed to update order status")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	h.logger.WithField("order_id", id).WithField("status", req.Status).Info("Order status updated successfully")
	c.JSON(http.StatusOK, order)
}

// DeleteOrder handles DELETE /orders/:id
func (h *OrderHandler) DeleteOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	err = h.orderService.DeleteOrder(c.Request.Context(), id)
	if err != nil {
		if err == domain.ErrOrderNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		if err == domain.ErrOrderCannotBeModified {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Order cannot be deleted in current status"})
			return
		}
		h.logger.WithField("error", err).Error("Failed to delete order")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete order"})
		return
	}

	h.logger.WithField("order_id", id).Info("Order deleted successfully")
	c.JSON(http.StatusNoContent, nil)
}

// ListOrders handles GET /orders
func (h *OrderHandler) ListOrders(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	result, err := h.orderService.ListOrders(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.WithField("error", err).Error("Failed to list orders")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list orders"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetOrdersByStatus handles GET /orders/status/:status
func (h *OrderHandler) GetOrdersByStatus(c *gin.Context) {
	statusStr := c.Param("status")
	status := domain.OrderStatus(statusStr)

	// Validate status
	validStatuses := []domain.OrderStatus{
		domain.OrderStatusPending,
		domain.OrderStatusConfirmed,
		domain.OrderStatusProcessing,
		domain.OrderStatusShipped,
		domain.OrderStatusDelivered,
		domain.OrderStatusCancelled,
		domain.OrderStatusRefunded,
	}

	isValid := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			isValid = true
			break
		}
	}

	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order status"})
		return
	}

	orders, err := h.orderService.GetOrdersByStatus(c.Request.Context(), status)
	if err != nil {
		h.logger.WithField("error", err).Error("Failed to get orders by status")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

// ConfirmOrderWithStockOverride handles POST /orders/{id}/confirm-with-override
func (h *OrderHandler) ConfirmOrderWithStockOverride(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		h.logger.Error("Invalid order ID format", "order_id", orderIDStr, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID format"})
		return
	}

	var req dto.ConfirmOrderWithStockOverrideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for stock override", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.logger.Error("Stock override request validation failed", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	order, err := h.orderService.ConfirmOrderWithStockOverride(c.Request.Context(), orderID, &req)
	if err != nil {
		h.logger.Error("Failed to confirm order with stock override", "order_id", orderID, "error", err)

		// Handle specific error cases
		switch err {
		case domain.ErrUnauthorizedStockOverride:
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to perform stock override"})
		case domain.ErrInvalidOrderStatus:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Order status is not valid for stock override"})
		case domain.ErrOrderNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		case domain.ErrOrderItemNotFound:
			c.JSON(http.StatusBadRequest, gin.H{"error": "One or more order items not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to confirm order with stock override"})
		}
		return
	}

	h.logger.Info("Order confirmed with stock override", "order_id", orderID, "user_id", req.UserID)
	c.JSON(http.StatusOK, gin.H{"data": order})
}

// Admin-specific order management endpoints

// CreateOrderForCustomer handles POST /admin/orders - Admin creates order for customer
func (h *OrderHandler) CreateOrderForCustomer(c *gin.Context) {
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for admin order creation", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.logger.Error("Admin order creation validation failed", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	// Get admin user info from context (set by auth middleware)
	adminUser, exists := c.Get("user")
	if !exists {
		h.logger.Error("Admin user not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	order, err := h.orderService.CreateOrder(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create order for customer", "admin_user", adminUser, "customer_id", req.CustomerID, "error", err)
		
		switch err {
		case domain.ErrInvalidCustomerID:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		}
		return
	}

	h.logger.Info("Order created by admin for customer", "admin_user", adminUser, "order_id", order.ID, "customer_id", req.CustomerID)
	c.JSON(http.StatusCreated, gin.H{"data": order})
}

// LinkOrderToChat handles POST /admin/orders/{id}/link-chat
func (h *OrderHandler) LinkOrderToChat(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		h.logger.Error("Invalid order ID format", "order_id", orderIDStr, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID format"})
		return
	}

	var req struct {
		ChatID   string `json:"chat_id" binding:"required"`
		Platform string `json:"platform" binding:"required"` // "line", "facebook", etc.
		UserID   string `json:"user_id,omitempty"`           // chat user ID
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for chat linking", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get admin user info from context
	adminUser, _ := c.Get("user")

	// Get the order first to verify it exists
	_, err = h.orderService.GetOrderByID(c.Request.Context(), orderID)
	if err != nil {
		h.logger.Error("Failed to get order for chat linking", "order_id", orderID, "error", err)
		
		if err == domain.ErrOrderNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get order"})
		}
		return
	}

	// Create update request to add chat metadata
	updateReq := &dto.UpdateOrderRequest{
		Notes: &[]string{fmt.Sprintf("Linked to %s chat: %s", req.Platform, req.ChatID)}[0],
	}

	updatedOrder, err := h.orderService.UpdateOrder(c.Request.Context(), orderID, updateReq)
	if err != nil {
		h.logger.Error("Failed to link order to chat", "order_id", orderID, "chat_id", req.ChatID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link order to chat"})
		return
	}

	h.logger.Info("Order linked to chat successfully", 
		"admin_user", adminUser, 
		"order_id", orderID, 
		"chat_id", req.ChatID, 
		"platform", req.Platform)

	c.JSON(http.StatusOK, gin.H{
		"data": updatedOrder,
		"chat_link": gin.H{
			"chat_id":  req.ChatID,
			"platform": req.Platform,
			"user_id":  req.UserID,
		},
	})
}

// BulkUpdateOrderStatus handles POST /admin/orders/bulk-status
func (h *OrderHandler) BulkUpdateOrderStatus(c *gin.Context) {
	var req struct {
		OrderIDs []string            `json:"order_ids" binding:"required,min=1"`
		Status   domain.OrderStatus  `json:"status" binding:"required"`
		Reason   string             `json:"reason,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for bulk status update", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate order IDs
	orderIDs := make([]uuid.UUID, len(req.OrderIDs))
	for i, idStr := range req.OrderIDs {
		orderID, err := uuid.Parse(idStr)
		if err != nil {
			h.logger.Error("Invalid order ID in bulk update", "order_id", idStr, "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid order ID: %s", idStr)})
			return
		}
		orderIDs[i] = orderID
	}

	// Get admin user info
	adminUser, _ := c.Get("user")

	// Process each order
	results := make([]gin.H, len(orderIDs))
	successCount := 0
	errorCount := 0

	for i, orderID := range orderIDs {
		statusReq := &dto.UpdateOrderStatusRequest{
			Status: req.Status,
		}

		order, err := h.orderService.UpdateOrderStatus(c.Request.Context(), orderID, statusReq)
		if err != nil {
			h.logger.Error("Failed to update order status in bulk", "order_id", orderID, "status", req.Status, "error", err)
			results[i] = gin.H{
				"order_id": orderID.String(),
				"success":  false,
				"error":    err.Error(),
			}
			errorCount++
		} else {
			results[i] = gin.H{
				"order_id": orderID.String(),
				"success":  true,
				"status":   order.Status,
			}
			successCount++
		}
	}

	h.logger.Info("Bulk order status update completed", 
		"admin_user", adminUser,
		"total_orders", len(orderIDs),
		"success_count", successCount,
		"error_count", errorCount,
		"new_status", req.Status)

	c.JSON(http.StatusOK, gin.H{
		"total_orders":  len(orderIDs),
		"success_count": successCount,
		"error_count":   errorCount,
		"results":       results,
	})
}

// ExportOrders handles GET /admin/orders/export
func (h *OrderHandler) ExportOrders(c *gin.Context) {
	// Get query parameters for filtering
	format := c.DefaultQuery("format", "csv") // csv or excel
	status := c.Query("status")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	customerID := c.Query("customer_id")

	// Validate format
	if format != "csv" && format != "excel" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid format. Use 'csv' or 'excel'"})
		return
	}

	// Get admin user info
	adminUser, _ := c.Get("user")

	// For now, export all orders (in production, you'd implement filtering)
	orders, err := h.orderService.ListOrders(c.Request.Context(), 1, 1000) // Max 1000 orders
	if err != nil {
		h.logger.Error("Failed to get orders for export", "admin_user", adminUser, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get orders for export"})
		return
	}

	// Apply filters if provided
	filteredOrders := h.filterOrdersForExport(orders.Orders, status, startDate, endDate, customerID)

	switch format {
	case "csv":
		h.exportOrdersAsCSV(c, filteredOrders, adminUser)
	case "excel":
		h.exportOrdersAsExcel(c, filteredOrders, adminUser)
	}
}

// Helper functions for export functionality
func (h *OrderHandler) filterOrdersForExport(orders []dto.OrderResponse, status, startDate, endDate, customerID string) []dto.OrderResponse {
	var filtered []dto.OrderResponse
	
	for _, order := range orders {
		// Apply status filter
		if status != "" && string(order.Status) != status {
			continue
		}
		
		// Apply customer ID filter
		if customerID != "" && order.CustomerID.String() != customerID {
			continue
		}
		
		// Date filtering would be implemented here with proper date parsing
		// For now, include all orders that pass other filters
		
		filtered = append(filtered, order)
	}
	
	return filtered
}

func (h *OrderHandler) exportOrdersAsCSV(c *gin.Context, orders []dto.OrderResponse, adminUser interface{}) {
	// Set headers for CSV download
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=orders_export.csv")

	// Create CSV content
	var csvContent strings.Builder
	csvContent.WriteString("Order ID,Customer ID,Status,Total Amount,Created At,Updated At\n")
	
	for _, order := range orders {
		csvContent.WriteString(fmt.Sprintf("%s,%s,%s,%.2f,%s,%s\n",
			order.ID.String(),
			order.CustomerID.String(),
			string(order.Status),
			order.TotalAmount,
			order.CreatedAt.Format("2006-01-02 15:04:05"),
			order.UpdatedAt.Format("2006-01-02 15:04:05"),
		))
	}

	h.logger.Info("Orders exported as CSV", "admin_user", adminUser, "order_count", len(orders))
	c.String(http.StatusOK, csvContent.String())
}

func (h *OrderHandler) exportOrdersAsExcel(c *gin.Context, orders []dto.OrderResponse, adminUser interface{}) {
	// Import excelize at the top level when building
	h.logger.Info("Excel export requested", "admin_user", adminUser, "order_count", len(orders))
	
	// Create new Excel file using excelize
	// For now, return CSV as Excel is more complex to implement
	// This can be extended later with full Excel formatting
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=orders_export.csv")
	
	// Delegate to CSV export for now
	h.exportOrdersAsCSV(c, orders, adminUser)
}

// HealthCheck handles GET /health
func (h *OrderHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "order-service",
	})
}
