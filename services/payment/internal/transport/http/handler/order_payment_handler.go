package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"payment/internal/application/dto"
	"payment/internal/application/usecase"
)

// OrderPaymentHandler handles order-based payment queries (Type 3)
type OrderPaymentHandler struct {
	orderPaymentUseCase *usecase.OrderPaymentUseCase
	logger              *logrus.Logger
}

// NewOrderPaymentHandler creates a new order payment handler
func NewOrderPaymentHandler(
	orderPaymentUseCase *usecase.OrderPaymentUseCase,
	logger *logrus.Logger,
) *OrderPaymentHandler {
	return &OrderPaymentHandler{
		orderPaymentUseCase: orderPaymentUseCase,
		logger:             logger,
	}
}

// GetOrderPayments handles GET /orders/:order_id/payments
func (h *OrderPaymentHandler) GetOrderPayments(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid order ID format",
			Code:  "INVALID_ORDER_ID",
		})
		return
	}

	req := &dto.GetOrderPaymentsRequest{
		OrderID: orderID,
	}

	// Get order payments
	payments, err := h.orderPaymentUseCase.GetOrderPayments(c.Request.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get order payments")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve order payments",
			Code:  "ORDER_PAYMENTS_RETRIEVAL_FAILED",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Order payments retrieved successfully",
		Data:    payments,
	})
}

// GetOrderPaymentSummary handles GET /orders/:order_id/payment-summary
func (h *OrderPaymentHandler) GetOrderPaymentSummary(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid order ID format",
			Code:  "INVALID_ORDER_ID",
		})
		return
	}

	req := &dto.GetOrderPaymentSummaryRequest{
		OrderID: orderID,
	}

	// Get order payment summary
	summary, err := h.orderPaymentUseCase.GetOrderPaymentSummary(c.Request.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get order payment summary")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve order payment summary",
			Code:  "ORDER_PAYMENT_SUMMARY_RETRIEVAL_FAILED",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Order payment summary retrieved successfully",
		Data:    summary,
	})
}

// GetOrderPaymentTimeline handles GET /orders/:order_id/payment-timeline
func (h *OrderPaymentHandler) GetOrderPaymentTimeline(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid order ID format",
			Code:  "INVALID_ORDER_ID",
		})
		return
	}

	// Get order payment timeline
	timeline, err := h.orderPaymentUseCase.GetOrderPaymentTimeline(c.Request.Context(), orderID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get order payment timeline")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve order payment timeline",
			Code:  "ORDER_PAYMENT_TIMELINE_RETRIEVAL_FAILED",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Order payment timeline retrieved successfully",
		Data:    timeline,
	})
}

// ValidateOrderPayment handles POST /orders/:order_id/validate-payment
func (h *OrderPaymentHandler) ValidateOrderPayment(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid order ID format",
			Code:  "INVALID_ORDER_ID",
		})
		return
	}

	var req struct {
		Amount float64 `json:"amount" validate:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind request")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request format",
			Code:  "INVALID_REQUEST",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return
	}

	// Validate order payment
	err = h.orderPaymentUseCase.ValidateOrderPayment(c.Request.Context(), orderID, req.Amount)
	if err != nil {
		h.logger.WithError(err).Warn("Order payment validation failed")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Payment validation failed",
			Code:  "PAYMENT_VALIDATION_FAILED",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Payment validation successful",
		Data: map[string]interface{}{
			"order_id": orderID,
			"amount":   req.Amount,
			"valid":    true,
		},
	})
}

// ProcessOrderPayment handles POST /orders/:order_id/process-payment
func (h *OrderPaymentHandler) ProcessOrderPayment(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid order ID format",
			Code:  "INVALID_ORDER_ID",
		})
		return
	}

	var req struct {
		Amount float64 `json:"amount" validate:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to bind request")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid request format",
			Code:  "INVALID_REQUEST",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return
	}

	// Process order payment
	result, err := h.orderPaymentUseCase.ProcessOrderPayment(c.Request.Context(), orderID, req.Amount)
	if err != nil {
		h.logger.WithError(err).Error("Failed to process order payment")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to process order payment",
			Code:  "ORDER_PAYMENT_PROCESSING_FAILED",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Order payment processed successfully",
		Data:    result,
	})
}

// RegisterRoutes registers order payment routes
func (h *OrderPaymentHandler) RegisterRoutes(router *gin.RouterGroup) {
	orders := router.Group("/orders")
	{
		orders.GET("/:order_id/payments", h.GetOrderPayments)
		orders.GET("/:order_id/payment-summary", h.GetOrderPaymentSummary)
		orders.GET("/:order_id/payment-timeline", h.GetOrderPaymentTimeline)
		orders.POST("/:order_id/validate-payment", h.ValidateOrderPayment)
		orders.POST("/:order_id/process-payment", h.ProcessOrderPayment)
	}
}
