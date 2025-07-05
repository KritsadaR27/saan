package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/go-playground/validator/v10"
	"order/internal/application"
	"order/internal/application/dto"
	"order/internal/domain"
	"github.com/sirupsen/logrus"
)

// Handler handles HTTP requests for orders using the new service
type Handler struct {
	service   *application.Service
	validator *validator.Validate
	logger    *logrus.Logger
}

// NewHandler creates a new order handler
func NewHandler(service *application.Service, logger *logrus.Logger) *Handler {
	return &Handler{
		service:   service,
		validator: validator.New(),
		logger:    logger,
	}
}

// CreateOrder handles POST /orders
func (h *Handler) CreateOrder(c *gin.Context) {
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.logger.WithError(err).Error("Request validation failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	order, err := h.service.CreateOrder(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create order")
		if err == domain.ErrInvalidOrderData {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	h.logger.WithField("order_id", order.ID).Info("Order created successfully")
	c.JSON(http.StatusCreated, order)
}

// GetOrder handles GET /orders/:id
func (h *Handler) GetOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.WithError(err).WithField("id", idStr).Error("Invalid order ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	order, err := h.service.GetOrder(c.Request.Context(), id)
	if err != nil {
		h.logger.WithError(err).WithField("order_id", id).Error("Failed to get order")
		if err == domain.ErrOrderNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get order"})
		return
	}

	c.JSON(http.StatusOK, order)
}

// ListOrders handles GET /orders
func (h *Handler) ListOrders(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset parameter"})
		return
	}

	orders, err := h.service.ListOrders(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list orders")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"orders": orders,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(orders),
		},
	})
}

// GetOrdersByCustomer handles GET /customers/:customer_id/orders
func (h *Handler) GetOrdersByCustomer(c *gin.Context) {
	customerIDStr := c.Param("customer_id")
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("customer_id", customerIDStr).Error("Invalid customer ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	orders, err := h.service.GetOrdersByCustomer(c.Request.Context(), customerID)
	if err != nil {
		h.logger.WithError(err).WithField("customer_id", customerID).Error("Failed to get customer orders")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get customer orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

// UpdateOrderStatus handles PUT /orders/:id/status
func (h *Handler) UpdateOrderStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.WithError(err).WithField("id", idStr).Error("Invalid order ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req struct {
		Status string `json:"status" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.logger.WithError(err).Error("Request validation failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	// Validate status
	status := domain.OrderStatus(req.Status)
	validStatuses := []domain.OrderStatus{
		domain.OrderStatusPending,
		domain.OrderStatusConfirmed,
		domain.OrderStatusProcessing,
		domain.OrderStatusShipped,
		domain.OrderStatusDelivered,
		domain.OrderStatusCompleted,
		domain.OrderStatusCancelled,
		domain.OrderStatusRefunded,
	}

	isValidStatus := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			isValidStatus = true
			break
		}
	}

	if !isValidStatus {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}

	err = h.service.UpdateOrderStatus(c.Request.Context(), id, status)
	if err != nil {
		h.logger.WithError(err).WithField("order_id", id).Error("Failed to update order status")
		if err == domain.ErrOrderNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		if err == domain.ErrInvalidStatusTransition {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status transition"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"order_id": id,
		"status":   status,
	}).Info("Order status updated successfully")

	c.JSON(http.StatusOK, gin.H{"message": "Order status updated successfully"})
}

// CancelOrder handles POST /orders/:id/cancel
func (h *Handler) CancelOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.WithError(err).WithField("id", idStr).Error("Invalid order ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req struct {
		Reason string `json:"reason" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.logger.WithError(err).Error("Request validation failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	err = h.service.CancelOrder(c.Request.Context(), id, req.Reason)
	if err != nil {
		h.logger.WithError(err).WithField("order_id", id).Error("Failed to cancel order")
		if err == domain.ErrOrderNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		if err == domain.ErrOrderAlreadyCancelled {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Order is already cancelled"})
			return
		}
		if err == domain.ErrOrderCannotBeCancelled {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Order cannot be cancelled"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel order"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"order_id": id,
		"reason":   req.Reason,
	}).Info("Order cancelled successfully")

	c.JSON(http.StatusOK, gin.H{"message": "Order cancelled successfully"})
}
