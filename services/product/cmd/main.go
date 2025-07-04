package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"product-service/internal/application"
	"product-service/internal/infrastructure/cache"
	"product-service/internal/infrastructure/config"
	"product-service/internal/infrastructure/database"
	"product-service/internal/infrastructure/events"
	"product-service/internal/infrastructure/loyverse"
	"product-service/internal/transport/http/handler"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Setup logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	if cfg.Environment == "production" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

	// Initialize database
	db, err := database.Initialize(cfg.Database)
	if err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close(db)

	// Initialize Redis cache
	redisCache, err := cache.NewRedisCache(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password, cfg.Redis.Database, logger)
	if err != nil {
		logger.Fatalf("Failed to initialize Redis cache: %v", err)
	}

	// Initialize Kafka event publisher
	var eventPublisher events.Publisher
	if len(cfg.Kafka.Brokers) > 0 {
		eventPublisher = events.NewKafkaPublisher(cfg.Kafka.Brokers, logger)
		logger.Info("Kafka event publisher initialized")
	} else {
		logger.Warn("Kafka brokers not configured, events will not be published")
		eventPublisher = events.NewNoOpPublisher() // Create a no-op publisher for development
	}

	// Initialize repositories
	productRepo := database.NewProductRepository(db)
	categoryRepo := database.NewCategoryRepository(db) // Add this for sync functionality
	// TODO: Add other repositories when implementations are ready
	// priceRepo := database.NewPriceRepository(db)
	// inventoryRepo := database.NewInventoryRepository(db)

	// Initialize use cases
	// For most operations, use direct database access (following PROJECT_RULES.md)
	productUsecase := application.NewProductUsecase(productRepo, redisCache, logger)
	// TODO: Uncomment when repository implementations are ready
	// categoryUsecase := application.NewCategoryUsecase(categoryRepo, logger)
	// pricingUsecase := application.NewPricingUsecase(priceRepo, productRepo, logger)
	// inventoryUsecase := application.NewInventoryUsecase(inventoryRepo, productRepo, logger)
	
	// Initialize sync usecase for Loyverse integration
	syncUsecase := application.NewSyncUsecase(productRepo, categoryRepo, eventPublisher, logger)

	// Initialize Loyverse integration
	var loyverseSyncService *loyverse.SyncService
	if cfg.External.LoyverseAPIKey != "" {
		loyverseClient := loyverse.NewClient(cfg.External.LoyverseAPIKey, logger)
		loyverseSyncService = loyverse.NewSyncService(loyverseClient, syncUsecase, eventPublisher, logger)
		logger.Info("Loyverse integration initialized")
	} else {
		logger.Warn("Loyverse API key not configured, sync functionality will be disabled")
	}

	// Initialize handlers
	productHandler := handler.NewProductHandler(productUsecase, logger)
	syncHandler := handler.NewSyncHandler(syncUsecase, loyverseSyncService, logger)
	// TODO: Add other handlers when ready
	// categoryHandler := handler.NewCategoryHandler(categoryUsecase, logger)
	// pricingHandler := handler.NewPricingHandler(pricingUsecase, logger)
	// inventoryHandler := handler.NewInventoryHandler(inventoryUsecase, logger)

	// Setup router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check endpoints following PROJECT_RULES.md standards
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "product",
			"version":   "1.0.0",
			"timestamp": time.Now().Unix(),
		})
	})
	
	router.GET("/ready", func(c *gin.Context) {
		// TODO: Add readiness checks for database, redis, etc.
		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
			"checks": gin.H{
				"database": "ok",
				"redis":    "ok",
				"kafka":    "ok",
			},
		})
	})
	
	router.GET("/metrics", func(c *gin.Context) {
		// TODO: Add Prometheus metrics
		c.JSON(http.StatusOK, gin.H{
			"metrics": "TODO: Implement Prometheus metrics",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		products := v1.Group("/products")
		{
			products.POST("", productHandler.CreateProduct)
			products.GET("", productHandler.GetProducts)
			products.GET("/:id", productHandler.GetProduct)
			products.PUT("/:id", productHandler.UpdateProduct)
			products.DELETE("/:id", productHandler.DeleteProduct)
		}

		sync := v1.Group("/sync")
		{
			sync.POST("/loyverse", syncHandler.SyncFromLoyverse)
			sync.POST("/loyverse/products/:loyverse_id", syncHandler.SyncProductFromLoyverse)
			sync.GET("/status/:sync_id", syncHandler.GetSyncStatus)
			sync.GET("/last", syncHandler.GetLastSyncTime)
		}
	}

	// Start server
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	logger.WithFields(logrus.Fields{
		"port":        cfg.Port,
		"environment": cfg.Environment,
		"service":     "product",
	}).Info("Product Service started")

	// Log service configuration for debugging
	logger.WithFields(logrus.Fields{
		"database_host": cfg.Database.Host,
		"redis_host":    cfg.Redis.Host,
		"kafka_brokers": cfg.Kafka.Brokers,
	}).Debug("Service configuration loaded")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Product Service...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown:", err)
	}

	logger.Info("Product Service stopped")
}
