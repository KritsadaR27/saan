package http

import (
	"github.com/gin-gonic/gin"

	"github.com/saan-system/services/customer/internal/application"
)

// SetupRoutes sets up HTTP routes
func SetupRoutes(router *gin.Engine, app *application.Application) {
	// Initialize handlers
	customerHandler := NewCustomerHandler(app)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "service": "customer-service"})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Customer routes
		customers := v1.Group("/customers")
		{
			customers.POST("/", customerHandler.CreateCustomer)
			customers.GET("/", customerHandler.ListCustomers)
			customers.GET("/search/email", customerHandler.GetCustomerByEmail)
			customers.GET("/search/phone", customerHandler.GetCustomerByPhone)
			customers.GET("/:id", customerHandler.GetCustomer)
			customers.PUT("/:id", customerHandler.UpdateCustomer)
			customers.DELETE("/:id", customerHandler.DeleteCustomer)

			// Customer address routes
			customers.POST("/:id/addresses", customerHandler.AddCustomerAddress)
			customers.PUT("/:id/addresses/:address_id", customerHandler.UpdateCustomerAddress)
			customers.DELETE("/:id/addresses/:address_id", customerHandler.DeleteCustomerAddress)
			customers.POST("/:id/addresses/:address_id/default", customerHandler.SetDefaultAddress)

			// Loyverse sync
			customers.POST("/:id/sync/loyverse", customerHandler.SyncWithLoyverse)
		}

		// Thai address routes
		addresses := v1.Group("/addresses")
		{
			// Address suggestions endpoint (ตาม SAAN_FLOW.MD)
			addresses.GET("/suggest", customerHandler.GetAddressSuggestions)
			addresses.GET("/thai/search", customerHandler.SearchThaiAddresses)
			addresses.GET("/thai/postal/:postal_code", customerHandler.GetThaiAddressByPostalCode)
		}
	}
}
