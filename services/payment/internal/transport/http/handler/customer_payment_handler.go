package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"payment/internal/application/dto"
	"payment/internal/application/usecase"
)

// CustomerPaymentHandler handles customer-based payment queries (Type 2)
type CustomerPaymentHandler struct {
	customerPaymentUseCase *usecase.CustomerPaymentUseCase
	logger                 *logrus.Logger
}

// NewCustomerPaymentHandler creates a new customer payment handler
func NewCustomerPaymentHandler(
	customerPaymentUseCase *usecase.CustomerPaymentUseCase,
	logger *logrus.Logger,
) *CustomerPaymentHandler {
	return &CustomerPaymentHandler{
		customerPaymentUseCase: customerPaymentUseCase,
		logger:                logger,
	}
}

// GetCustomerPayments handles GET /customers/:customer_id/payments
func (h *CustomerPaymentHandler) GetCustomerPayments(c *gin.Context) {
	customerIDStr := c.Param("customer_id")
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid customer ID format",
			Code:  "INVALID_CUSTOMER_ID",
		})
		return
	}

	// Parse query parameters for filters
	var req dto.GetCustomerPaymentsRequest
	req.CustomerID = customerID

	// Parse filters from query parameters
	if status := c.Query("status"); status != "" {
		req.Filters.Status = parsePaymentStatus(status)
	}

	if method := c.Query("payment_method"); method != "" {
		req.Filters.PaymentMethod = parsePaymentMethod(method)
	}

	if channel := c.Query("payment_channel"); channel != "" {
		req.Filters.PaymentChannel = parsePaymentChannel(channel)
	}

	if timing := c.Query("payment_timing"); timing != "" {
		req.Filters.PaymentTiming = parsePaymentTiming(timing)
	}

	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if t, err := time.Parse("2006-01-02", dateFrom); err == nil {
			req.Filters.DateFrom = &t
		}
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		if t, err := time.Parse("2006-01-02", dateTo); err == nil {
			req.Filters.DateTo = &t
		}
	}

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			req.Filters.Limit = l
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			req.Filters.Offset = o
		}
	}

	req.Filters.SortBy = c.DefaultQuery("sort_by", "created_at")
	req.Filters.SortOrder = c.DefaultQuery("sort_order", "DESC")

	// Get customer payments
	payments, err := h.customerPaymentUseCase.GetCustomerPayments(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get customer payments")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve customer payments",
			Code:  "CUSTOMER_PAYMENTS_RETRIEVAL_FAILED",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Customer payments retrieved successfully",
		Data:    payments,
	})
}

// GetCustomerPaymentHistory handles GET /customers/:customer_id/payment-history
func (h *CustomerPaymentHandler) GetCustomerPaymentHistory(c *gin.Context) {
	customerIDStr := c.Param("customer_id")
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid customer ID format",
			Code:  "INVALID_CUSTOMER_ID",
		})
		return
	}

	// Parse limit parameter
	limit := 20 // Default limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	req := &dto.GetCustomerPaymentHistoryRequest{
		CustomerID: customerID,
		Limit:      limit,
	}

	// Get customer payment history
	history, err := h.customerPaymentUseCase.GetCustomerPaymentHistory(c.Request.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get customer payment history")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve customer payment history",
			Code:  "CUSTOMER_PAYMENT_HISTORY_RETRIEVAL_FAILED",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Customer payment history retrieved successfully",
		Data:    history,
	})
}

// GetCustomerPaymentStats handles GET /customers/:customer_id/payment-stats
func (h *CustomerPaymentHandler) GetCustomerPaymentStats(c *gin.Context) {
	customerIDStr := c.Param("customer_id")
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid customer ID format",
			Code:  "INVALID_CUSTOMER_ID",
		})
		return
	}

	// Get customer payment stats
	stats, err := h.customerPaymentUseCase.GetCustomerPaymentStats(c.Request.Context(), customerID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get customer payment stats")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve customer payment statistics",
			Code:  "CUSTOMER_PAYMENT_STATS_RETRIEVAL_FAILED",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Customer payment statistics retrieved successfully",
		Data:    stats,
	})
}

// RegisterRoutes registers customer payment routes
func (h *CustomerPaymentHandler) RegisterRoutes(router *gin.RouterGroup) {
	customers := router.Group("/customers")
	{
		customers.GET("/:customer_id/payments", h.GetCustomerPayments)
		customers.GET("/:customer_id/payment-history", h.GetCustomerPaymentHistory)
		customers.GET("/:customer_id/payment-stats", h.GetCustomerPaymentStats)
	}
}


