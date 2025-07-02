package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/saan-system/services/customer/internal/application"
	"github.com/saan-system/services/customer/internal/infrastructure/database"
	"github.com/saan-system/services/customer/internal/infrastructure/cache"
	"github.com/saan-system/services/customer/internal/infrastructure/messaging"
	"github.com/saan-system/services/customer/internal/infrastructure/loyverse"
	httphandler "github.com/saan-system/services/customer/internal/transport/http"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	// Initialize database
	db, err := database.New()
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	// Initialize Redis cache
	redisClient, err := cache.New()
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// Initialize Kafka
	kafkaProducer, err := messaging.NewProducer()
	if err != nil {
		logger.Fatal("Failed to initialize Kafka producer", zap.Error(err))
	}
	defer kafkaProducer.Close()

	// Initialize Loyverse client
	loyverseClient := loyverse.NewClient()

	// Initialize application services
	app := application.New(
		db,
		redisClient,
		kafkaProducer,
		loyverseClient,
		logger,
	)

	// Initialize HTTP server
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Setup routes
	httphandler.SetupRoutes(router, app)

	// Configure server
	port := getEnv("PORT", "8110")
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting Customer Service", zap.String("port", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
