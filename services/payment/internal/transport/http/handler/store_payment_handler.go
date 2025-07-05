package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"payment/internal/application/dto"
	"payment/internal/application/usecase"
)

// StorePaymentHandler handles store-based payment queries (Type 1)
type StorePaymentHandler struct {
	storePaymentUseCase *usecase.StorePaymentUseCase
	logger              *logrus.Logger
}

// NewStorePaymentHandler creates a new store payment handler
func NewStorePaymentHandler(
	storePaymentUseCase *usecase.StorePaymentUseCase,
	logger *logrus.Logger,
) *StorePaymentHandler {
	return &StorePaymentHandler{
		storePaymentUseCase: storePaymentUseCase,
		logger:             logger,
	}
}

// GetStorePayments handles GET /stores/:store_id/payments
func (h *StorePaymentHandler) GetStorePayments(c *gin.Context) {
	storeID := c.Param("store_id")
	if storeID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Store ID is required",
			Code:  "MISSING_STORE_ID",
		})
		return
	}

	// Parse query parameters for filters
	var req dto.GetStorePaymentsRequest
	req.StoreID = storeID

	// Parse filters from query parameters
	if status := c.Query("status"); status != "" {
		// Convert string to PaymentStatus
		// Implementation would include proper validation
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
		if l, err := parseInt(limit); err == nil {
			req.Filters.Limit = l
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := parseInt(offset); err == nil {
			req.Filters.Offset = o
		}
	}

	req.Filters.SortBy = c.DefaultQuery("sort_by", "created_at")
	req.Filters.SortOrder = c.DefaultQuery("sort_order", "DESC")

	// Get store payments
	payments, err := h.storePaymentUseCase.GetStorePayments(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get store payments")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve store payments",
			Code:  "STORE_PAYMENTS_RETRIEVAL_FAILED",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Store payments retrieved successfully",
		Data:    payments,
	})
}

// GetStoreAnalytics handles GET /stores/:store_id/analytics
func (h *StorePaymentHandler) GetStoreAnalytics(c *gin.Context) {
	storeID := c.Param("store_id")
	if storeID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Store ID is required",
			Code:  "MISSING_STORE_ID",
		})
		return
	}

	// Parse date range parameters
	dateFromStr := c.Query("date_from")
	dateToStr := c.Query("date_to")

	if dateFromStr == "" || dateToStr == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Both date_from and date_to are required",
			Code:  "MISSING_DATE_RANGE",
		})
		return
	}

	dateFrom, err := time.Parse("2006-01-02", dateFromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid date_from format (expected YYYY-MM-DD)",
			Code:  "INVALID_DATE_FORMAT",
		})
		return
	}

	dateTo, err := time.Parse("2006-01-02", dateToStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid date_to format (expected YYYY-MM-DD)",
			Code:  "INVALID_DATE_FORMAT",
		})
		return
	}

	req := &dto.GetStoreAnalyticsRequest{
		StoreID:  storeID,
		DateFrom: dateFrom,
		DateTo:   dateTo,
	}

	// Get store analytics
	analytics, err := h.storePaymentUseCase.GetStoreAnalytics(c.Request.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get store analytics")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to retrieve store analytics",
			Code:  "STORE_ANALYTICS_RETRIEVAL_FAILED",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Store analytics retrieved successfully",
		Data:    analytics,
	})
}

// RegisterRoutes registers store payment routes
func (h *StorePaymentHandler) RegisterRoutes(router *gin.RouterGroup) {
	stores := router.Group("/stores")
	{
		stores.GET("/:store_id/payments", h.GetStorePayments)
		stores.GET("/:store_id/analytics", h.GetStoreAnalytics)
	}
}
