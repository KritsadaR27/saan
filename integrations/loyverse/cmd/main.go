// integrations/loyverse/cmd/main.go
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

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"

	"integrations/loyverse/config"
	"integrations/loyverse/internal/connector"
	"integrations/loyverse/internal/events"
	"integrations/loyverse/internal/sync"
	"integrations/loyverse/internal/webhook"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer redisClient.Close()

	// Test Redis connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Initialize Loyverse API client
	loyverseClient := connector.NewClient(cfg.LoyverseAPIToken)

	// Initialize Kafka publisher
	publisher := events.NewPublisher(cfg.KafkaBrokers, cfg.KafkaTopic)
	defer publisher.Close()

	// Initialize sync services
	productSync := sync.NewProductSync(loyverseClient, publisher, redisClient)
	inventorySync := sync.NewInventorySync(loyverseClient, publisher, redisClient)
	receiptSync := sync.NewReceiptSync(loyverseClient, publisher, redisClient)
	customerSync := sync.NewCustomerSync(loyverseClient, publisher, redisClient)

	// Initialize sync manager
	syncManager, err := sync.NewManager(
		productSync,
		inventorySync,
		receiptSync,
		customerSync,
		redisClient,
		sync.Config{
			ProductSyncInterval:   cfg.ProductSyncInterval,
			InventorySyncInterval: cfg.InventorySyncInterval,
			ReceiptSyncInterval:   cfg.ReceiptSyncInterval,
			CustomerSyncInterval:  cfg.CustomerSyncInterval,
			TimeZone:              cfg.TimeZone,
		},
	)
	if err != nil {
		log.Fatalf("Failed to create sync manager: %v", err)
	}

	// Start sync manager
	if err := syncManager.Start(ctx); err != nil {
		log.Fatalf("Failed to start sync manager: %v", err)
	}
	defer syncManager.Stop()

	// Initialize webhook handler
	webhookProcessor := webhook.NewProcessor(redisClient)
	webhookHandler := webhook.NewHandler(cfg.WebhookSecret, webhookProcessor, publisher)

	// Setup HTTP server
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","service":"loyverse-integration","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	}).Methods("GET")

	// Webhook endpoint
	router.Handle("/webhook/loyverse", webhookHandler).Methods("POST")

	// Admin endpoints
	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(authMiddleware(cfg.AdminToken))

	// Manual sync triggers
	adminRouter.HandleFunc("/sync/{type}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		syncType := vars["type"]

		if err := syncManager.TriggerSync(ctx, syncType); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message":"Sync triggered for %s"}`, syncType)
	}).Methods("POST")

	// Sync status endpoint
	adminRouter.HandleFunc("/sync/status", func(w http.ResponseWriter, r *http.Request) {
		status, err := syncManager.GetSyncStatus(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(status); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}).Methods("GET")

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting Loyverse integration service on port %d", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}

// authMiddleware checks for admin token
func authMiddleware(adminToken string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("X-Admin-Token")
			if token != adminToken {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
