package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// Defer cleanup
	defer func() {
		redisClient.Close()
		dbConn.Close()
		kafkaConsumer.Close()
	}()

	// Initialize HTTP router
	router := routes.SetupRoutes(redisClient, dbConn, kafkaConsumer, logger)

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
