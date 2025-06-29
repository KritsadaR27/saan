package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/saan/order-service/internal/application"
	"github.com/saan/order-service/internal/application/template"
	"github.com/saan/order-service/internal/infrastructure/client"
	"github.com/saan/order-service/internal/infrastructure/config"
	"github.com/saan/order-service/internal/infrastructure/db"
	"github.com/saan/order-service/internal/infrastructure/event"
	"github.com/saan/order-service/internal/infrastructure/repository"
	httpTransport "github.com/saan/order-service/internal/transport/http"
	"github.com/saan/order-service/internal/transport/http/middleware"
	"github.com/saan/order-service/pkg/logger"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	
	// Initialize logger
	log := logger.NewLogger(cfg.Logger.Level, cfg.Logger.Format)
	log.Info("Starting Order Service...")
	
	// Connect to database
	database, err := db.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()
	log.Info("Connected to database successfully")
	
	// Initialize repositories
	orderRepo := repository.NewPostgresOrderRepository(database)
	orderItemRepo := repository.NewPostgresOrderItemRepository(database)
	auditRepo := repository.NewPostgresAuditRepository(database)
	orderEventRepo := repository.NewPostgresEventRepository(database)
	
	// Initialize HTTP clients (following PROJECT_RULES.md service names)
	inventoryClient := client.NewHTTPInventoryClient("http://inventory-service:8082")
	customerClient := client.NewHTTPCustomerClient("http://user-service:8088") 
	notificationClient := client.NewHTTPNotificationClient("http://notification-service:8092")
	
	// Initialize event publisher (Kafka)
	kafkaBrokers := []string{"kafka:9092"} // Use service name as per PROJECT_RULES.md
	eventPublisher, err := event.NewKafkaEventPublisher(kafkaBrokers, "order-events")
	if err != nil {
		log.Fatalf("Failed to create event publisher: %v", err)
	}
	defer eventPublisher.Close()
	
	// Initialize outbox worker
	outboxConfig := event.DefaultOutboxWorkerConfig()
	outboxWorker := event.NewOutboxWorker(orderEventRepo, eventPublisher, outboxConfig, log)
	
	// Start outbox worker
	ctx := context.Background()
	outboxWorker.Start(ctx)
	defer outboxWorker.Stop()
	
	// Initialize services
	orderService := application.NewOrderService(orderRepo, orderItemRepo, auditRepo, orderEventRepo, eventPublisher, log)
	
	// Initialize template selector for chat integration
	templateSelector := template.NewTemplateSelector()
	
	// Initialize chat order service
	chatOrderService := application.NewChatOrderService(
		orderService,
		customerClient,
		inventoryClient,
		notificationClient,
		templateSelector,
		log,
	)
	
	// Initialize handlers
	orderHandler := httpTransport.NewOrderHandler(orderService, log)
	chatOrderHandler := httpTransport.NewChatOrderHandler(chatOrderService, log)
	
	// Initialize auth config
	authConfig := &middleware.AuthConfig{
		AuthServiceURL: "http://user-service:8088", // Following PROJECT_RULES.md service names
		JWTSecret:      cfg.JWT.Secret,
		Logger:         log,
	}
	
	// Setup routes with auth config
	router := httpTransport.SetupRoutes(orderHandler, chatOrderHandler, authConfig, log)
	
	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}
	
	// Start server in a goroutine
	go func() {
		log.Infof("Server starting on %s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	
	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")
	
	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	} else {
		log.Info("Server shutdown completed")
	}
}
