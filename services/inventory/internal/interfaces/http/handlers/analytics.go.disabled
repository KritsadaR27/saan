package handlers

import (
	"database/sql"
	"net/http"

	"services/inventory/internal/infrastructure/redis"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AnalyticsHandler struct {
	redisClient *redis.Client
	db          *sql.DB
	logger      *logrus.Logger
}

func NewAnalyticsHandler(redisClient *redis.Client, db *sql.DB, logger *logrus.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{
		redisClient: redisClient,
		db:          db,
		logger:      logger,
	}
}

// GetDashboard returns analytics dashboard data
func (h *AnalyticsHandler) GetDashboard(c *gin.Context) {
	// Mock dashboard data for now
	dashboard := gin.H{
		"summary": gin.H{
			"total_products":    150,
			"total_stores":      3,
			"low_stock_items":   12,
			"total_stock_value": 125000.50,
		},
		"stock_status": gin.H{
			"in_stock":     128,
			"low_stock":    12,
			"out_of_stock": 10,
		},
		"recent_movements": []gin.H{
			{
				"type":        "SALE",
				"product":     "หมูสามชั้น",
				"quantity":    2.5,
				"store":       "สาขา 1",
				"timestamp":   "2024-01-15T10:30:00Z",
			},
			{
				"type":        "PURCHASE",
				"product":     "ไก่ทั้งตัว",
				"quantity":    10,
				"store":       "สาขา 2",
				"timestamp":   "2024-01-15T09:15:00Z",
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dashboard,
	})
}

// GetProductPerformance returns product performance metrics
func (h *AnalyticsHandler) GetProductPerformance(c *gin.Context) {
	// Mock product performance data
	performance := []gin.H{
		{
			"product_id":    "1",
			"product_name":  "หมูสามชั้น",
			"category":      "เนื้อสัตว์",
			"total_sold":    45.5,
			"revenue":       9100.0,
			"current_stock": 12.5,
			"turnover_rate": 3.64,
			"rank":          1,
		},
		{
			"product_id":    "2",
			"product_name":  "ไก่ทั้งตัว",
			"category":      "เนื้อสัตว์",
			"total_sold":    30,
			"revenue":       4500.0,
			"current_stock": 8,
			"turnover_rate": 3.75,
			"rank":          2,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"performance": performance,
			"count":       len(performance),
		},
	})
}

// GetCategoryPerformance returns category performance metrics
func (h *AnalyticsHandler) GetCategoryPerformance(c *gin.Context) {
	// Mock category performance data
	performance := []gin.H{
		{
			"category_id":   "1",
			"category_name": "เนื้อสัตว์",
			"product_count": 25,
			"total_sold":    150.5,
			"revenue":       45000.0,
			"avg_turnover":  3.2,
			"rank":          1,
		},
		{
			"category_id":   "2", 
			"category_name": "ผักสด",
			"product_count": 40,
			"total_sold":    220.0,
			"revenue":       22000.0,
			"avg_turnover":  2.8,
			"rank":          2,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"performance": performance,
			"count":       len(performance),
		},
	})
}

// GetDailyTrends returns daily movement trends
func (h *AnalyticsHandler) GetDailyTrends(c *gin.Context) {
	// Mock daily trends data
	trends := []gin.H{
		{
			"date":               "2024-01-15",
			"total_sales":        245.5,
			"total_purchases":    150.0,
			"total_adjustments":  5.0,
			"net_movement":       -90.5,
			"value_moved":        12500.0,
		},
		{
			"date":               "2024-01-14",
			"total_sales":        198.0,
			"total_purchases":    200.0,
			"total_adjustments":  -2.0,
			"net_movement":       0.0,
			"value_moved":        11800.0,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"trends": trends,
			"count":  len(trends),
		},
	})
}

// GetWeeklyTrends returns weekly trend analysis
func (h *AnalyticsHandler) GetWeeklyTrends(c *gin.Context) {
	// Mock weekly trends data
	trends := []gin.H{
		{
			"week_start":   "2024-01-08",
			"week_end":     "2024-01-14",
			"total_sales":  1250.5,
			"revenue":      87500.0,
			"growth_rate":  12.5,
			"top_category": "เนื้อสัตว์",
		},
		{
			"week_start":   "2024-01-01",
			"week_end":     "2024-01-07",
			"total_sales":  1112.0,
			"revenue":      77800.0,
			"growth_rate":  8.2,
			"top_category": "เนื้อสัตว์",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"trends": trends,
			"count":  len(trends),
		},
	})
}

// GetReorderSuggestions returns intelligent reorder suggestions
func (h *AnalyticsHandler) GetReorderSuggestions(c *gin.Context) {
	// Mock reorder suggestions
	suggestions := []gin.H{
		{
			"product_id":       "1",
			"product_name":     "หมูสามชั้น",
			"store_id":         "1",
			"store_name":       "สาขา 1",
			"current_stock":    5.5,
			"suggested_qty":    25.0,
			"reason_code":      "LOW_STOCK",
			"confidence":       0.85,
			"estimated_cost":   5000.0,
			"estimated_demand": 20.0,
			"lead_time_days":   3,
			"priority":         1,
		},
		{
			"product_id":       "5",
			"product_name":     "ผักกาดขาว",
			"store_id":         "2",
			"store_name":       "สาขา 2",
			"current_stock":    8.0,
			"suggested_qty":    30.0,
			"reason_code":      "SEASONAL",
			"confidence":       0.72,
			"estimated_cost":   900.0,
			"estimated_demand": 25.0,
			"lead_time_days":   1,
			"priority":         2,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"suggestions": suggestions,
			"count":       len(suggestions),
		},
	})
}

// GetSystemStats returns system statistics (admin only)
func (h *AnalyticsHandler) GetSystemStats(c *gin.Context) {
	stats := gin.H{
		"cache_status": gin.H{
			"redis_connected":   true,
			"total_keys":        1250,
			"cache_hit_rate":    0.89,
			"last_sync":         "2024-01-15T11:30:00Z",
		},
		"database_status": gin.H{
			"db_connected":      true,
			"active_connections": 5,
			"max_connections":   100,
		},
		"service_health": gin.H{
			"uptime_seconds":    86400,
			"total_requests":    15420,
			"error_rate":        0.02,
			"avg_response_time": "45ms",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}
