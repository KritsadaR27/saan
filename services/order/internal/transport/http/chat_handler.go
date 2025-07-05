package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"order/internal/application"
	"order/pkg/logger"
)

// ChatOrderHandler handles HTTP requests for chat-based order operations
type ChatOrderHandler struct {
	chatOrderService *application.ChatOrderService
	logger           logger.Logger
}

// NewChatOrderHandler creates a new chat order handler
func NewChatOrderHandler(chatOrderService *application.ChatOrderService, logger logger.Logger) *ChatOrderHandler {
	return &ChatOrderHandler{
		chatOrderService: chatOrderService,
		logger:           logger,
	}
}

// CreateOrderFromChat handles POST /chat/orders
func (h *ChatOrderHandler) CreateOrderFromChat(c *gin.Context) {
	var req application.ChatOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for chat order", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Basic validation
	if req.ChatID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chat_id is required"})
		return
	}
	if req.CustomerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "customer_id is required"})
		return
	}
	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "items are required"})
		return
	}

	order, err := h.chatOrderService.CreateOrderFromChat(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create order from chat", "chat_id", req.ChatID, "customer_id", req.CustomerID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order", "details": err.Error()})
		return
	}

	h.logger.Info("Order created from chat successfully", "chat_id", req.ChatID, "order_id", order.ID)
	c.JSON(http.StatusCreated, gin.H{"data": order})
}

// ConfirmChatOrder handles POST /chat/orders/{id}/confirm
func (h *ChatOrderHandler) ConfirmChatOrder(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		h.logger.Error("Invalid order ID format", "order_id", orderIDStr, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID format"})
		return
	}

	var req struct {
		ChatID string `json:"chat_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for chat order confirmation", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	order, err := h.chatOrderService.ConfirmChatOrder(c.Request.Context(), req.ChatID, orderID)
	if err != nil {
		h.logger.Error("Failed to confirm chat order", "order_id", orderID, "chat_id", req.ChatID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to confirm order", "details": err.Error()})
		return
	}

	h.logger.Info("Chat order confirmed successfully", "order_id", orderID, "chat_id", req.ChatID)
	c.JSON(http.StatusOK, gin.H{"data": order})
}

// CancelChatOrder handles POST /chat/orders/{id}/cancel
func (h *ChatOrderHandler) CancelChatOrder(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		h.logger.Error("Invalid order ID format", "order_id", orderIDStr, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID format"})
		return
	}

	var req struct {
		ChatID string `json:"chat_id" binding:"required"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for chat order cancellation", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err = h.chatOrderService.CancelChatOrder(c.Request.Context(), req.ChatID, orderID, req.Reason)
	if err != nil {
		h.logger.Error("Failed to cancel chat order", "order_id", orderID, "chat_id", req.ChatID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel order", "details": err.Error()})
		return
	}

	h.logger.Info("Chat order cancelled successfully", "order_id", orderID, "chat_id", req.ChatID)
	c.JSON(http.StatusOK, gin.H{"message": "Order cancelled successfully"})
}

// GenerateOrderSummary handles POST /chat/orders/{id}/summary
func (h *ChatOrderHandler) GenerateOrderSummary(c *gin.Context) {
	orderIDStr := c.Param("id")
	_, err := uuid.Parse(orderIDStr)
	if err != nil {
		h.logger.Error("Invalid order ID format", "order_id", orderIDStr, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID format"})
		return
	}

	// ดึงข้อมูล order จาก order service
	// ในกรณีนี้ต้องเข้าถึง OrderService ผ่าน ChatOrderService
	// หรือให้ ChatOrderHandler มี OrderService เป็น dependency เพิ่มเติม
	
	c.JSON(http.StatusNotImplemented, gin.H{"error": "This endpoint is not yet implemented"})
}
