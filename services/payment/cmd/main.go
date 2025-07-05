package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	_ "github.com/lib/pq"

	"payment/internal/application/usecase"
	"payment/internal/infrastructure/config"
	repoImpl "payment/internal/infrastructure/repository"
	"payment/internal/transport/http/handler"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	db, err := initDatabase(cfg, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize database")
	}
	defer db.Close()

	// Initialize Redis
	redisClient, err := initRedis(cfg, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize Redis")
	}
	defer redisClient.Close()

	// Initialize repositories
	paymentRepo := repoImpl.NewPostgresPaymentRepository(db)
	
	// Initialize use cases
	paymentUseCase := usecase.NewPaymentUseCase(
		paymentRepo,
		nil, // loyverseStoreRepo - to be implemented
		nil, // deliveryContextRepo - to be implemented
		nil, // eventRepo - to be implemented
		logger,
	)

	storePaymentUseCase := usecase.NewStorePaymentUseCase(
		paymentRepo,
		nil, // loyverseStoreRepo - to be implemented
	)

	customerPaymentUseCase := usecase.NewCustomerPaymentUseCase(
		paymentRepo,
		nil, // deliveryContextRepo - to be implemented
	)

	orderPaymentUseCase := usecase.NewOrderPaymentUseCase(
		paymentRepo,
		nil, // deliveryContextRepo - to be implemented
	)

	// Initialize HTTP handlers
	paymentHandler := handler.NewPaymentHandler(paymentUseCase, logger)
	storePaymentHandler := handler.NewStorePaymentHandler(storePaymentUseCase, logger)
	customerPaymentHandler := handler.NewCustomerPaymentHandler(customerPaymentUseCase, logger)
	orderPaymentHandler := handler.NewOrderPaymentHandler(orderPaymentUseCase, logger)

	// Setup HTTP server
	server := setupHTTPServer(cfg, logger, 
		paymentHandler, 
		storePaymentHandler, 
		customerPaymentHandler, 
		orderPaymentHandler,
	)

	// Start server
	go func() {
		logger.WithField("port", cfg.Server.Port).Info("Starting Payment Service")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Payment Service...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	}

	logger.Info("Payment Service stopped")
}

func initDatabase(cfg *config.Config, logger *logrus.Logger) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)  // Default max open connections
	db.SetMaxIdleConns(10)  // Default max idle connections
	db.SetConnMaxLifetime(time.Hour) // Default connection lifetime

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established")
	return db, nil
}

func initRedis(cfg *config.Config, logger *logrus.Logger) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis connection established")
	return client, nil
}

func setupHTTPServer(
	cfg *config.Config,
	logger *logrus.Logger,
	paymentHandler *handler.PaymentHandler,
	storePaymentHandler *handler.StorePaymentHandler,
	customerPaymentHandler *handler.CustomerPaymentHandler,
	orderPaymentHandler *handler.OrderPaymentHandler,
) *http.Server {
	// Set Gin mode
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(loggingMiddleware(logger))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "payment-service",
			"version":   "1.0.0",
			"timestamp": time.Now().UTC(),
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Register all handlers
		paymentHandler.RegisterRoutes(api)
		storePaymentHandler.RegisterRoutes(api)
		customerPaymentHandler.RegisterRoutes(api)
		orderPaymentHandler.RegisterRoutes(api)
	}

	return &http.Server{
		Addr:           cfg.Server.Host + ":" + cfg.Server.Port,
		Handler:        router,
		ReadTimeout:    30 * time.Second,  // Default read timeout
		WriteTimeout:   30 * time.Second,  // Default write timeout
		IdleTimeout:    60 * time.Second,  // Default idle timeout
		MaxHeaderBytes: 1 << 20,           // Default max header bytes (1MB)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func loggingMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return gin.LoggerWithWriter(logger.Writer())
}
