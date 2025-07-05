package routes

import (
	"net/http"
	"inventory/internal/infrastructure/cache"
	"inventory/internal/infrastructure/database"
	"inventory/internal/infrastructure/events"
	"inventory/internal/interfaces/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func SetupRoutes(
	redisClient *cache.RedisClient,
	dbConn *database.Connection,
	kafkaConsumer *events.Consumer,
	logger *logrus.Logger,
) *gin.Engine {
	// Initialize Gin router
	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.RequestID())

	// Basic health check endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "inventory"})
	})

	router.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready", "service": "inventory"})
	})

	// Basic API routes (simplified for now)
	api := router.Group("/api/v1")
	{
		// Basic inventory endpoints
		inventory := api.Group("/inventory")
		{
			inventory.GET("/status", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Inventory service running with new infrastructure"})
			})
		}
	}

	return router
}
