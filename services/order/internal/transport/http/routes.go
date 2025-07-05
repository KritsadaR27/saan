package http

import (
	"github.com/gin-gonic/gin"
	"order/internal/application"
	"github.com/sirupsen/logrus"
)

// SetupRoutes configures the HTTP routes using the new handler
func SetupRoutes(service *application.Service, logger *logrus.Logger) *gin.Engine {
	router := gin.New()
	
	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	
	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})
	
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})
	
	// Create handler
	handler := NewHandler(service, logger)
	
	// API routes
	v1 := router.Group("/api/v1")
	{
		// Order routes
		orders := v1.Group("/orders")
		{
			orders.POST("", handler.CreateOrder)
			orders.GET("", handler.ListOrders)
			orders.GET("/:id", handler.GetOrder)
			orders.PUT("/:id/status", handler.UpdateOrderStatus)
			orders.POST("/:id/cancel", handler.CancelOrder)
		}
		
		// Customer order routes
		customers := v1.Group("/customers")
		{
			customers.GET("/:customer_id/orders", handler.GetOrdersByCustomer)
		}
	}
	
	return router
}
