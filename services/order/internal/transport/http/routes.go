package http

import (
	"github.com/gin-gonic/gin"
	"github.com/saan/order-service/internal/transport/http/middleware"
	"github.com/saan/order-service/pkg/logger"
)

// SetupRoutes sets up all HTTP routes
func SetupRoutes(orderHandler *OrderHandler, chatOrderHandler *ChatOrderHandler, authConfig *middleware.AuthConfig, logger logger.Logger) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)
	
	// Create router
	r := gin.New()
	
	// Add middleware
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.LoggingMiddleware(logger))
	r.Use(middleware.ErrorHandlingMiddleware(logger))
	
	// Health check (no auth required)
	r.GET("/health", orderHandler.HealthCheck)
	
	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Public order routes (with optional auth for user context)
		orders := v1.Group("/orders")
		orders.Use(middleware.OptionalAuth(authConfig))
		{
			orders.POST("", middleware.RequireRole(authConfig, middleware.RoleSales, middleware.RoleManager, middleware.RoleAdmin), orderHandler.CreateOrder)
			orders.GET("", middleware.RequireRole(authConfig, middleware.RoleSales, middleware.RoleManager, middleware.RoleAdmin), orderHandler.ListOrders)
			orders.GET("/:id", middleware.RequireRole(authConfig, middleware.RoleSales, middleware.RoleManager, middleware.RoleAdmin), orderHandler.GetOrder)
			orders.PUT("/:id", middleware.RequireRole(authConfig, middleware.RoleManager, middleware.RoleAdmin), orderHandler.UpdateOrder)
			orders.DELETE("/:id", middleware.RequireRole(authConfig, middleware.RoleAdmin), orderHandler.DeleteOrder)
			orders.PATCH("/:id/status", middleware.RequireRole(authConfig, middleware.RoleManager, middleware.RoleAdmin), orderHandler.UpdateOrderStatus)
			orders.POST("/:id/confirm-with-override", middleware.RequirePermission(authConfig, "orders:override_stock"), orderHandler.ConfirmOrderWithStockOverride)
			orders.GET("/status/:status", middleware.RequireRole(authConfig, middleware.RoleSales, middleware.RoleManager, middleware.RoleAdmin), orderHandler.GetOrdersByStatus)
		}
		
		// Customer orders routes (require at least sales role)
		customers := v1.Group("/customers")
		customers.Use(middleware.RequireRole(authConfig, middleware.RoleSales, middleware.RoleManager, middleware.RoleAdmin))
		{
			customers.GET("/:customerId/orders", orderHandler.GetOrdersByCustomer)
		}
		
		// Chat-based order routes (AI assistant access)
		if chatOrderHandler != nil {
			chat := v1.Group("/chat")
			chat.Use(middleware.RequireRole(authConfig, middleware.RoleAIAssistant, middleware.RoleManager, middleware.RoleAdmin))
			{
				chat.POST("/orders", chatOrderHandler.CreateOrderFromChat)
				chat.POST("/orders/:id/confirm", chatOrderHandler.ConfirmChatOrder)
				chat.POST("/orders/:id/cancel", chatOrderHandler.CancelChatOrder)
				chat.POST("/orders/:id/summary", chatOrderHandler.GenerateOrderSummary)
			}
		}
		
		// Admin routes (require admin role or specific permissions)
		admin := v1.Group("/admin")
		admin.Use(middleware.RequireRole(authConfig, middleware.RoleAdmin))
		{
			orders := admin.Group("/orders")
			{
				orders.POST("", middleware.RequirePermission(authConfig, "admin:create_order"), orderHandler.CreateOrderForCustomer)
				orders.POST("/:id/link-chat", middleware.RequirePermission(authConfig, "admin:link_chat"), orderHandler.LinkOrderToChat)
				orders.POST("/bulk-status", middleware.RequirePermission(authConfig, "admin:bulk_update"), orderHandler.BulkUpdateOrderStatus)
				orders.GET("/export", middleware.RequirePermission(authConfig, "admin:export"), orderHandler.ExportOrders)
			}
		}
	}
	
	return r
}
