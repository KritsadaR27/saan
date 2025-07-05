package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"payment/internal/application/dto"
	"payment/internal/application/usecase"
)

// PaymentHandler handles payment-related HTTP requests
type PaymentHandler struct {
	paymentUseCase  *usecase.PaymentUseCase
	logger          *logrus.Logger
}

// NewPaymentHandler creates a new payment handler
func NewPaymentHandler(
	paymentUseCase *usecase.PaymentUseCase,
	logger *logrus.Logger,
) *PaymentHandler {
	return &PaymentHandler{
		paymentUseCase: paymentUseCase,
		logger:        logger,
	}
}

// CreatePayment handles POST /payments
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req dto.CreatePaymentRequest
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

	// Create payment
	payment, err := h.paymentUseCase.CreatePayment(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create payment")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to create payment",
			Code:  "PAYMENT_CREATION_FAILED",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse{
		Message: "Payment created successfully",
		Data:    payment,
	})
}

// GetPayment handles GET /payments/:id
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	idStr := c.Param("id")
	paymentID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid payment ID format",
			Code:  "INVALID_PAYMENT_ID",
		})
		return
	}

	payment, err := h.paymentUseCase.GetPaymentByID(c.Request.Context(), paymentID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get payment")
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Payment not found",
			Code:  "PAYMENT_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Payment retrieved successfully",
		Data:    payment,
	})
}

// UpdatePaymentStatus handles PUT /payments/:id/status
func (h *PaymentHandler) UpdatePaymentStatus(c *gin.Context) {
	idStr := c.Param("id")
	paymentID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid payment ID format",
			Code:  "INVALID_PAYMENT_ID",
		})
		return
	}

	var req dto.UpdatePaymentStatusRequest
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

	// Update payment status
	payment, err := h.paymentUseCase.UpdatePaymentStatus(c.Request.Context(), paymentID, &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update payment status")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to update payment status",
			Code:  "PAYMENT_UPDATE_FAILED",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Payment status updated successfully",
		Data:    payment,
	})
}

// RegisterRoutes registers payment routes
func (h *PaymentHandler) RegisterRoutes(router *gin.RouterGroup) {
	payments := router.Group("/payments")
	{
		payments.POST("", h.CreatePayment)
		payments.GET("/:id", h.GetPayment)
		payments.PUT("/:id/status", h.UpdatePaymentStatus)
	}
}
