package http

import (
	"net/http"

	"saan/payment/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PaymentHandler struct {
	paymentService domain.PaymentService
	refundService  domain.RefundService
}

func NewRouter(paymentService domain.PaymentService, refundService domain.RefundService) *gin.Engine {
	handler := &PaymentHandler{
		paymentService: paymentService,
		refundService:  refundService,
	}

	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "payment-service",
			"status":  "healthy",
		})
	})

	// API routes
	api := router.Group("/api/payment")
	{
		// Payments
		api.POST("/payments", handler.CreatePayment)
		api.GET("/payments/:id", handler.GetPayment)
		api.POST("/payments/:id/process", handler.ProcessPayment)
		api.POST("/payments/:id/complete", handler.CompletePayment)
		api.POST("/payments/:id/fail", handler.FailPayment)
		api.GET("/orders/:orderId/payments", handler.GetPaymentsByOrder)
		api.GET("/payments/:id/history", handler.GetPaymentHistory)
		
		// Refunds
		api.POST("/refunds", handler.CreateRefund)
		api.GET("/refunds/:id", handler.GetRefund)
		api.POST("/refunds/:id/process", handler.ProcessRefund)
		api.GET("/refunds/by-payment/:paymentId", handler.GetRefundsByPayment)
	}

	return router
}

func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req domain.PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, err := h.paymentService.CreatePayment(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, payment)
}

func (h *PaymentHandler) GetPayment(c *gin.Context) {
	paymentIDStr := c.Param("id")
	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment id"})
		return
	}

	payment, err := h.paymentService.GetPaymentByID(paymentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payment)
}

func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	paymentIDStr := c.Param("id")
	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment id"})
		return
	}

	response, err := h.paymentService.ProcessPayment(paymentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *PaymentHandler) CompletePayment(c *gin.Context) {
	paymentIDStr := c.Param("id")
	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment id"})
		return
	}

	var req struct {
		ExternalTransactionID string `json:"external_transaction_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.paymentService.CompletePayment(paymentID, req.ExternalTransactionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "payment completed"})
}

func (h *PaymentHandler) FailPayment(c *gin.Context) {
	paymentIDStr := c.Param("id")
	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment id"})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.paymentService.FailPayment(paymentID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "payment failed"})
}

func (h *PaymentHandler) GetPaymentsByOrder(c *gin.Context) {
	orderIDStr := c.Param("orderId")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	payments, err := h.paymentService.GetPaymentsByOrderID(orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payments": payments})
}

func (h *PaymentHandler) GetPaymentHistory(c *gin.Context) {
	paymentIDStr := c.Param("id")
	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment id"})
		return
	}

	history, err := h.paymentService.GetPaymentHistory(paymentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"history": history})
}

func (h *PaymentHandler) CreateRefund(c *gin.Context) {
	var req struct {
		PaymentID uuid.UUID `json:"payment_id" binding:"required"`
		Amount    float64   `json:"amount" binding:"required"`
		Reason    string    `json:"reason" binding:"required"`
		CreatedBy uuid.UUID `json:"created_by" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	refund, err := h.refundService.CreateRefund(req.PaymentID, req.Amount, req.Reason, req.CreatedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, refund)
}

func (h *PaymentHandler) GetRefund(c *gin.Context) {
	refundIDStr := c.Param("id")
	refundID, err := uuid.Parse(refundIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid refund id"})
		return
	}

	refund, err := h.refundService.GetRefundByID(refundID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, refund)
}

func (h *PaymentHandler) ProcessRefund(c *gin.Context) {
	refundIDStr := c.Param("id")
	refundID, err := uuid.Parse(refundIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid refund id"})
		return
	}

	if err := h.refundService.ProcessRefund(refundID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "refund processed"})
}

func (h *PaymentHandler) GetRefundsByPayment(c *gin.Context) {
	paymentIDStr := c.Param("paymentId")
	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment id"})
		return
	}

	refunds, err := h.refundService.GetRefundsByPaymentID(paymentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"refunds": refunds})
}
