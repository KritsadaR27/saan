package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"services/inventory/internal/application"
	"services/inventory/internal/config"
	"services/inventory/internal/infrastructure/kafka"
	"services/inventory/internal/infrastructure/postgres"
	"services/inventory/internal/infrastructure/redis"
	"services/inventory/internal/interfaces/http/routes"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
		logger.SetLevel(logrus.InfoLevel)
	} else {
		gin.SetMode(gin.DebugMode)
		logger.SetLevel(logrus.DebugLevel)
	}

	// Initialize infrastructure
	redisClient := redis.NewClient(cfg.RedisAddr, cfg.RedisPassword)
	dbConn := postgres.NewConnection(cfg.DatabaseURL)
	kafkaConsumer := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaConsumerGroup)

	// Initialize application services
	productService := application.NewProductService(dbConn, logger)

	// Register Kafka event handlers
	kafkaConsumer.RegisterHandler("product.updated", func(eventType string, data []byte) error {
		return productService.UpsertProduct(context.Background(), data)
	})

	// Defer cleanup
	defer func() {
		redisClient.Close()
		dbConn.Close()
		kafkaConsumer.Close()
	}()

	// Initialize HTTP router with custom routes
	router := routes.SetupRoutes(redisClient, dbConn, kafkaConsumer, logger)
	
	// Add direct product upsert endpoint (bypassing Kafka)
	router.POST("/api/v1/products/upsert", gin.HandlerFunc(func(c *gin.Context) {
		var productData map[string]interface{}
		if err := c.ShouldBindJSON(&productData); err != nil {
			logger.WithError(err).Error("Failed to decode product request")
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request format",
				"error":   err.Error(),
			})
			return
		}

		// Validate required fields
		productID, hasID := productData["product_id"].(string)
		name, hasName := productData["name"].(string)
		source, hasSource := productData["source"].(string)

		if !hasID || !hasName || !hasSource || productID == "" || name == "" || source == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Missing required fields: product_id, name, source",
			})
			return
		}

		// Log the request
		logger.WithFields(logrus.Fields{
			"product_id": productID,
			"name":       name,
			"source":     source,
		}).Info("✅ Received direct product upsert request")

		// Convert to JSON for service (same format as Kafka events)
		data, err := json.Marshal(productData)
		if err != nil {
			logger.WithError(err).Error("❌ Failed to marshal product data")
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to process product data",
			})
			return
		}

		// Call service method (reuse the Kafka event handler logic)
		if err := productService.UpsertProduct(c.Request.Context(), data); err != nil {
			logger.WithError(err).WithField("product_id", productID).Error("❌ Failed to upsert product")
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to upsert product",
				"error":   err.Error(),
			})
			return
		}

		// Success response
		logger.WithField("product_id", productID).Info("✅ Product upserted successfully")
		c.JSON(http.StatusOK, gin.H{
			"success":    true,
			"product_id": productID,
			"message":    "Product upserted successfully",
		})
	}))

	logger.Info("🔗 Registered direct product upsert route: POST /api/v1/products/upsert")

	// Setup HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start Kafka consumer in background
	go func() {
		logger.Info("Starting Kafka consumer for real-time inventory updates")
		if err := kafkaConsumer.StartConsuming(); err != nil {
			logger.WithError(err).Error("Failed to start Kafka consumer")
		}
	}()

	// Start server in background
	go func() {
		logger.WithField("port", cfg.Port).Info("Starting Inventory Service HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Inventory Service...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Failed to gracefully shutdown server")
	}

	logger.Info("Inventory Service stopped")
}

func UpsertProductHandler(c *gin.Context) {
	var productData map[string]interface{}
	if err := c.ShouldBindJSON(&productData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	router := gin.Default()

	// Middleware to inject dependencies
	router.Use(func(c *gin.Context) {
		c.Set("redisClient", redisClient)
		c.Set("dbConn", dbConn)
		c.Set("kafkaConsumer", kafkaConsumer)
		c.Set("logger", logger)
		c.Next()
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Product upsert endpoint
	router.POST("/products/upsert", UpsertProductHandler)

	return router
}
