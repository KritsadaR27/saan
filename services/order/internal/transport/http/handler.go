package http

import (
	"net/http"
	"strconv"

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

// HealthCheck handles GET /health
func (h *OrderHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "order-service",
	})
}
