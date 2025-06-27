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
	"github.com/saan/order-service/internal/infrastructure/config"
	"github.com/saan/order-service/internal/infrastructure/db"
	"github.com/saan/order-service/internal/infrastructure/repository"
	httpTransport "github.com/saan/order-service/internal/transport/http"
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
	
	// Initialize services
	orderService := application.NewOrderService(orderRepo, orderItemRepo)
	
	// Initialize handlers
	orderHandler := httpTransport.NewOrderHandler(orderService, log)
	
	// Setup routes
	router := httpTransport.SetupRoutes(orderHandler, log)
	
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
