package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/saan/order-service/internal/application"
	"github.com/saan/order-service/internal/application/dto"
	"github.com/saan/order-service/pkg/logger"
)

// StatsHandler handles HTTP requests for order statistics
type StatsHandler struct {
	statsService *application.OrderStatsService
	validator    *validator.Validate
	logger       logger.Logger
}

// NewStatsHandler creates a new statistics handler
func NewStatsHandler(statsService *application.OrderStatsService, logger logger.Logger) *StatsHandler {
	return &StatsHandler{
		statsService: statsService,
		validator:    validator.New(),
		logger:       logger,
	}
}

// GetDailyStats handles GET /stats/daily
func (h *StatsHandler) GetDailyStats(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		h.logger.Error("Invalid date format", "date", dateStr, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}

	stats, err := h.statsService.GetDailyStats(c.Request.Context(), date)
	if err != nil {
		h.logger.Error("Failed to get daily stats", "date", dateStr, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get daily statistics"})
		return
	}

	h.logger.Info("Daily stats retrieved successfully", "date", dateStr)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetMonthlyStats handles GET /stats/monthly
func (h *StatsHandler) GetMonthlyStats(c *gin.Context) {
	yearStr := c.Query("year")
	monthStr := c.Query("month")

	// Default to current year and month if not provided
	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		} else {
			h.logger.Error("Invalid year format", "year", yearStr)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year format"})
			return
		}
	}

	if monthStr != "" {
		if m, err := strconv.Atoi(monthStr); err == nil && m >= 1 && m <= 12 {
			month = m
		} else {
			h.logger.Error("Invalid month format", "month", monthStr)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid month format. Use 1-12"})
			return
		}
	}

	stats, err := h.statsService.GetMonthlyStats(c.Request.Context(), year, month)
	if err != nil {
		h.logger.Error("Failed to get monthly stats", "year", year, "month", month, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get monthly statistics"})
		return
	}

	h.logger.Info("Monthly stats retrieved successfully", "year", year, "month", month)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetTopProducts handles GET /stats/top-products
func (h *StatsHandler) GetTopProducts(c *gin.Context) {
	var req dto.TopProductsRequest
	
	// Set defaults
	req.Limit = 10
	req.SortBy = "order_count"

	// Parse query parameters
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.Limit = limit
		}
	}

	if sortBy := c.Query("sort_by"); sortBy != "" {
		switch sortBy {
		case "order_count", "revenue", "quantity":
			req.SortBy = sortBy
		default:
			h.logger.Warn("Invalid sort_by parameter, using default", "sort_by", sortBy)
		}
	}

	// Parse date range if provided
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			req.StartDate = &startDate
		} else {
			h.logger.Error("Invalid start_date format", "start_date", startDateStr)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD"})
			return
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			req.EndDate = &endDate
		} else {
			h.logger.Error("Invalid end_date format", "end_date", endDateStr)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD"})
			return
		}
	}

	products, err := h.statsService.GetTopProducts(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to get top products", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get top products"})
		return
	}

	h.logger.Info("Top products retrieved successfully", "count", len(products), "limit", req.Limit)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    products,
		"meta": gin.H{
			"count":    len(products),
			"limit":    req.Limit,
			"sort_by":  req.SortBy,
		},
	})
}

// GetCustomerStats handles GET /stats/customer/:customer_id
func (h *StatsHandler) GetCustomerStats(c *gin.Context) {
	customerIDStr := c.Param("customer_id")
	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		h.logger.Error("Invalid customer ID format", "customer_id", customerIDStr, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID format"})
		return
	}

	var req dto.CustomerStatsRequest
	req.CustomerID = customerID

	// Parse query parameters
	if includeFavStr := c.Query("include_favorite_products"); includeFavStr == "true" {
		req.IncludeFavoriteProducts = true
		
		if limitStr := c.Query("favorite_products_limit"); limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
				req.FavoriteProductsLimit = limit
			} else {
				req.FavoriteProductsLimit = 5 // default
			}
		} else {
			req.FavoriteProductsLimit = 5 // default
		}
	}

	stats, err := h.statsService.GetCustomerStats(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to get customer stats", "customer_id", customerID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get customer statistics"})
		return
	}

	h.logger.Info("Customer stats retrieved successfully", "customer_id", customerID)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetOverallStats handles GET /stats/overview
func (h *StatsHandler) GetOverallStats(c *gin.Context) {
	// Get stats for the current month
	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	monthlyStats, err := h.statsService.GetMonthlyStats(c.Request.Context(), year, month)
	if err != nil {
		h.logger.Error("Failed to get monthly stats for overview", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get overview statistics"})
		return
	}

	// Get today's stats
	todayStats, err := h.statsService.GetDailyStats(c.Request.Context(), now)
	if err != nil {
		h.logger.Error("Failed to get daily stats for overview", "error", err)
		// Continue without today's stats
		todayStats = &dto.DailyStats{
			Date: now,
		}
	}

	// Get top products for this month
	topProductsReq := &dto.TopProductsRequest{
		Limit:  5,
		SortBy: "revenue",
	}
	
	startOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)
	topProductsReq.StartDate = &startOfMonth
	topProductsReq.EndDate = &endOfMonth

	topProducts, err := h.statsService.GetTopProducts(c.Request.Context(), topProductsReq)
	if err != nil {
		h.logger.Error("Failed to get top products for overview", "error", err)
		// Continue without top products
		topProducts = []dto.ProductStats{}
	}

	overview := gin.H{
		"current_month": monthlyStats,
		"today":        todayStats,
		"top_products": topProducts,
		"generated_at": time.Now(),
	}

	h.logger.Info("Overview stats retrieved successfully")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    overview,
	})
}
