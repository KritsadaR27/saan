package routes

import (
	"services/inventory/internal/infrastructure/kafka"
	"services/inventory/internal/infrastructure/postgres"
	"services/inventory/internal/infrastructure/redis"
	"services/inventory/internal/interfaces/http/handlers"
	"services/inventory/internal/interfaces/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func SetupRoutes(
	redisClient *redis.Client,
	dbConn *postgres.Connection,
	kafkaConsumer *kafka.Consumer,
	logger *logrus.Logger,
) *gin.Engine {
	// Initialize Gin router
	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.RequestID())

	// Initialize handlers
	inventoryHandler := handlers.NewInventoryHandler(redisClient, dbConn.DB, logger)
	analyticsHandler := handlers.NewAnalyticsHandler(redisClient, dbConn.DB, logger)
	healthHandler := handlers.NewHealthHandler(redisClient, dbConn, logger)

	// Health check endpoint
	router.GET("/health", healthHandler.HealthCheck)
	router.GET("/ready", healthHandler.ReadinessCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Inventory routes
		inventory := v1.Group("/inventory")
		{
			// Product endpoints
			inventory.GET("/products", inventoryHandler.GetAllProducts)
			inventory.GET("/products/:id", inventoryHandler.GetProduct)
			inventory.GET("/products/:id/stock", inventoryHandler.GetProductStock)
			inventory.GET("/search", inventoryHandler.SearchProducts)

			// Store endpoints
			inventory.GET("/stores", inventoryHandler.GetAllStores)
			inventory.GET("/stores/:id/stock", inventoryHandler.GetStoreStock)

			// Category endpoints
			inventory.GET("/categories", inventoryHandler.GetAllCategories)

			// Stock operations
			inventory.GET("/stock/low", inventoryHandler.GetLowStockItems)
			inventory.GET("/alerts", inventoryHandler.GetInventoryAlerts)
		}

		// Analytics routes
		analytics := v1.Group("/analytics")
		{
			analytics.GET("/dashboard", analyticsHandler.GetDashboard)
			analytics.GET("/performance/products", analyticsHandler.GetProductPerformance)
			analytics.GET("/performance/categories", analyticsHandler.GetCategoryPerformance)
			analytics.GET("/trends/daily", analyticsHandler.GetDailyTrends)
			analytics.GET("/trends/weekly", analyticsHandler.GetWeeklyTrends)
			analytics.GET("/suggestions/reorder", analyticsHandler.GetReorderSuggestions)
		}

		// Admin routes (with authentication middleware)
		admin := v1.Group("/admin")
		admin.Use(middleware.AdminAuth())
		{
			admin.POST("/sync/trigger", inventoryHandler.TriggerSync)
			admin.POST("/cache/refresh", inventoryHandler.RefreshCache)
			admin.GET("/stats", analyticsHandler.GetSystemStats)
		}
	}

	// WebSocket endpoint for real-time updates
	router.GET("/ws/inventory", inventoryHandler.WebSocketHandler)

	return router
}
