package http

import (
	"time"

	"github.com/gin-gonic/gin"

	"customer/internal/application"
	"customer/internal/transport/http/handler"
	"customer/internal/transport/http/middleware"
)

// SetupRoutes sets up HTTP routes
func SetupRoutes(router *gin.Engine, app *application.Application) {
	// Initialize handlers
	customerHandler := handler.NewCustomerHandler(
		app.CustomerUsecase,
		app.AddressUsecase,
		app.PointsUsecase,
	)
	addressHandler := handler.NewAddressHandler(app.AddressUsecase)
	pointsHandler := handler.NewPointsHandler(app.PointsUsecase)

	// Apply global middleware
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())
	router.Use(middleware.Timeout(30 * time.Second))

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
			customers.POST("/:id/addresses", addressHandler.AddCustomerAddress)
			customers.PUT("/:id/addresses/:address_id", addressHandler.UpdateCustomerAddress)
			customers.DELETE("/:id/addresses/:address_id", addressHandler.DeleteCustomerAddress)
			customers.POST("/:id/addresses/:address_id/default", addressHandler.SetDefaultAddress)

			// Customer points routes
			customers.GET("/:id/points", pointsHandler.GetPointsBalance)
			customers.POST("/:id/points/earn", pointsHandler.EarnPoints)
			customers.POST("/:id/points/redeem", pointsHandler.RedeemPoints)
			customers.GET("/:id/points/history", pointsHandler.GetPointsHistory)
			customers.GET("/:id/points/stats", pointsHandler.GetPointsStats)

			// Loyverse sync
			customers.POST("/:id/sync/loyverse", customerHandler.SyncWithLoyverse)
		}

		// Thai address routes
		addresses := v1.Group("/addresses")
		{
			// Address suggestions endpoint (ตาม SAAN_FLOW.MD)
			addresses.GET("/suggest", addressHandler.GetAddressSuggestions)
			addresses.GET("/thai/search", addressHandler.SearchThaiAddresses)
			addresses.GET("/thai/postal/:postal_code", addressHandler.GetThaiAddressByPostalCode)
		}

		// VIP tier routes
		vip := v1.Group("/vip")
		{
			vip.GET("/tiers", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"tiers": []gin.H{
						{"level": 1, "name": "Bronze", "min_spent": 0},
						{"level": 2, "name": "Silver", "min_spent": 10000},
						{"level": 3, "name": "Gold", "min_spent": 50000},
						{"level": 4, "name": "Platinum", "min_spent": 100000},
						{"level": 5, "name": "Diamond", "min_spent": 250000},
					},
				})
			})
		}
	}
}
