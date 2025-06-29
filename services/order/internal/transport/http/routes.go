package http

import (
	"github.com/gin-gonic/gin"
	"github.com/saan/order-service/internal/transport/http/middleware"
	"github.com/saan/order-service/pkg/logger"
)

// SetupRoutes sets up all HTTP routes
func SetupRoutes(orderHandler *OrderHandler, chatOrderHandler *ChatOrderHandler, logger logger.Logger) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)
	
	// Create router
	r := gin.New()
	
	// Add middleware
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.LoggingMiddleware(logger))
	r.Use(middleware.ErrorHandlingMiddleware(logger))
	
	// Health check
	r.GET("/health", orderHandler.HealthCheck)
	
	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Order routes
		orders := v1.Group("/orders")
		{
			orders.POST("", orderHandler.CreateOrder)
			orders.GET("", orderHandler.ListOrders)
			orders.GET("/:id", orderHandler.GetOrder)
			orders.PUT("/:id", orderHandler.UpdateOrder)
			orders.DELETE("/:id", orderHandler.DeleteOrder)
			orders.PATCH("/:id/status", orderHandler.UpdateOrderStatus)
			orders.POST("/:id/confirm-with-override", orderHandler.ConfirmOrderWithStockOverride)
			orders.GET("/status/:status", orderHandler.GetOrdersByStatus)
		}
		
		// Customer orders routes
		customers := v1.Group("/customers")
		{
			customers.GET("/:customerId/orders", orderHandler.GetOrdersByCustomer)
		}
		
		// Chat-based order routes
		if chatOrderHandler != nil {
			chat := v1.Group("/chat")
			{
				chat.POST("/orders", chatOrderHandler.CreateOrderFromChat)
				chat.POST("/orders/:id/confirm", chatOrderHandler.ConfirmChatOrder)
				chat.POST("/orders/:id/cancel", chatOrderHandler.CancelChatOrder)
				chat.POST("/orders/:id/summary", chatOrderHandler.GenerateOrderSummary)
			}
		}
	}
	
	return r
}
