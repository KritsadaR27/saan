package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"order/internal/application"
	"order/internal/infrastructure/cache"
	"order/internal/infrastructure/config"
	"order/internal/infrastructure/database"
	"order/internal/infrastructure/events"
	"order/internal/infrastructure/repository"
	httpTransport "order/internal/transport/http"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	
	// Setup logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	if cfg.Logging.Level == "debug" {
		logger.SetLevel(logrus.DebugLevel)
	}
	if cfg.Server.Environment == "production" {
		logger.SetFormatter(&logrus.JSONFormatter{})
		gin.SetMode(gin.ReleaseMode)
	}
	
	logger.Info("Starting Order Service...")
	
	// Initialize database
	db, err := database.NewConnection(cfg.Database, logger)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	logger.Info("Connected to database successfully")

	// Initialize Redis cache
	var redisCache *cache.RedisClient
	if cfg.Redis.Host != "" {
		redisCache, err = cache.NewRedisClient(cfg.Redis, logger)
		if err != nil {
			logger.Warnf("Failed to initialize Redis cache: %v", err)
			logger.Info("Continuing without Redis cache")
		} else {
			defer redisCache.Close()
			logger.Info("Redis cache initialized successfully")
		}
	}

	// Initialize event publisher
	var eventPublisher events.Publisher
	if len(cfg.Kafka.Brokers) > 0 && cfg.Kafka.Brokers[0] != "" {
		eventPublisher = events.NewKafkaPublisher(cfg.Kafka, logger)
		logger.Info("Kafka event publisher initialized")
	} else {
		eventPublisher = events.NewNoopPublisher(logger)
		logger.Warn("Kafka brokers not configured, using noop event publisher")
	}
	defer eventPublisher.Close()
	
	// Initialize repositories
	orderRepo := repository.NewOrderRepository(db)
	orderItemRepo := repository.NewOrderItemRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	orderEventRepo := repository.NewEventRepository(db)
	
	// Initialize service
	orderService := application.NewService(orderRepo, orderItemRepo, auditRepo, orderEventRepo, eventPublisher, redisCache, logger)
	
	// Setup routes
	router := httpTransport.SetupRoutes(orderService, logger)
	
	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}
	
	// Start server in a goroutine
	go func() {
		logger.Infof("Server starting on %s:%s", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
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
		logger.Errorf("Server forced to shutdown: %v", err)
	} else {
		logger.Info("Server shutdown completed")
	}
}
