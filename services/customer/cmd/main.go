package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/saan-system/services/customer/internal/application"
	"github.com/saan-system/services/customer/internal/infrastructure/cache"
	"github.com/saan-system/services/customer/internal/infrastructure/database"
	"github.com/saan-system/services/customer/internal/infrastructure/events"
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

	// Initialize database connection
	dbURL := getEnv("DATABASE_URL", "postgres://customer:password@localhost:5432/customer_db?sslmode=disable")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
	}

	// Initialize repositories
	customerRepo := database.NewCustomerRepository(db)
	addressRepo := database.NewCustomerAddressRepository(db)
	vipBenefitsRepo := database.NewVIPTierBenefitsRepository(db)
	pointsRepo := database.NewCustomerPointsRepository(db)
	analyticsRepo := database.NewCustomerAnalyticsRepository(db)
	thaiAddressRepo := database.NewThaiAddressRepository(db)
	deliveryRouteRepo := database.NewDeliveryRouteRepository(db)

	// Initialize Redis cache
	redisClient, err := cache.NewRedisCache(getEnv("REDIS_URL", "localhost:6379"))
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	// Initialize Kafka event publisher
	eventPublisher := events.NewKafkaEventPublisher(
		[]string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		getEnv("KAFKA_TOPIC", "customer-events"),
	)

	// Initialize Loyverse client
	loyverseClient := loyverse.NewClient(
		getEnv("LOYVERSE_API_TOKEN", ""),
		getEnv("LOYVERSE_BASE_URL", "https://api.loyverse.com/v1.0"),
	)

	// Create application dependencies
	deps := application.Dependencies{
		CustomerRepo:       customerRepo,
		AddressRepo:        addressRepo,
		VIPBenefitsRepo:    vipBenefitsRepo,
		PointsRepo:         pointsRepo,
		AnalyticsRepo:      analyticsRepo,
		ThaiAddressRepo:    thaiAddressRepo,
		DeliveryRouteRepo:  deliveryRouteRepo,
		CacheRepo:          redisClient,
		EventPublisher:     eventPublisher,
		LoyverseClient:     loyverseClient,
		Logger:             logger,
	}

	// Initialize application services
	app := application.New(deps)

	// Initialize HTTP server
	router := gin.New()

	// Setup routes (middleware is applied inside SetupRoutes)
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
