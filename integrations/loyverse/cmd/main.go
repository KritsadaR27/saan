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

	"integrations/loyverse/config"
	"integrations/loyverse/internal/connector"
	"integrations/loyverse/internal/events"
	"integrations/loyverse/internal/webhook"
	
	"github.com/go-redis/redis/v8"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Setup context with cancellation
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Printf("Starting Loyverse integration service on port %d", cfg.Port)
	log.Printf("Redis: %s", cfg.RedisAddr)
	log.Printf("Kafka: %s", cfg.KafkaBrokers)

	// Initialize Redis client  
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
		DB:   0,
	})
	defer redisClient.Close()

	// Initialize Kafka event publisher
	publisher := events.NewPublisher(cfg.KafkaBrokers, cfg.KafkaTopic)
	defer publisher.Close()

	// Initialize webhook processor
	processor := webhook.NewProcessor(redisClient)

	// Initialize webhook handler
	webhookHandler := webhook.NewHandler(cfg.WebhookSecret, processor, publisher)

	// Setup HTTP routes
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","service":"loyverse-integration","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	})

	// Webhook endpoint for Loyverse
	http.Handle("/webhook/loyverse", webhookHandler)

	// API endpoint to get latest receipt
	http.HandleFunc("/api/latest-receipt", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := context.Background()
		receiptData, updatedAt, err := processor.GetLatestReceipt(ctx)
		if err != nil {
			log.Printf("Error getting latest receipt: %v", err)
			http.Error(w, "No receipt found", http.StatusNotFound)
			return
		}

		response := map[string]interface{}{
			"receipt":    json.RawMessage(receiptData),
			"updated_at": updatedAt.Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	})

	// API endpoint to get all categories
	http.HandleFunc("/api/categories", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := context.Background()
		categories, err := processor.GetAllCategories(ctx)
		if err != nil {
			log.Printf("Error getting categories: %v", err)
			http.Error(w, "Error retrieving categories", http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"categories": categories,
			"count":      len(categories),
			"timestamp":  time.Now().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	})

	// API endpoint to fetch categories from Loyverse API directly
	http.HandleFunc("/api/categories/fetch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if cfg.LoyverseAPIToken == "" {
			http.Error(w, "Loyverse API token not configured", http.StatusBadRequest)
			return
		}

		// Create Loyverse API client
		loyverseAPI := connector.NewLoyverseAPI(cfg.LoyverseAPIToken)
		
		// Fetch categories from Loyverse API
		categories, err := loyverseAPI.GetCategories()
		if err != nil {
			log.Printf("Error fetching categories from Loyverse API: %v", err)
			http.Error(w, fmt.Sprintf("Failed to fetch categories: %v", err), http.StatusInternalServerError)
			return
		}

		log.Printf("Fetched %d categories from Loyverse API", len(categories))

		// Convert categories to JSON format for storage
		var categoryRawMessages []json.RawMessage
		for _, category := range categories {
			categoryData, err := json.Marshal(category)
			if err != nil {
				log.Printf("Error marshaling category %s: %v", category.ID, err)
				continue
			}
			categoryRawMessages = append(categoryRawMessages, json.RawMessage(categoryData))
		}

		// Store categories in Redis
		ctx := context.Background()
		if err := processor.StoreCategories(ctx, categoryRawMessages); err != nil {
			log.Printf("Error storing categories in Redis: %v", err)
			http.Error(w, "Failed to store categories", http.StatusInternalServerError)
			return
		}

		log.Printf("Stored %d categories in Redis", len(categoryRawMessages))

		// Publish category sync events to Kafka
		for _, category := range categories {
			// Transform to domain event
			categoryData, _ := json.Marshal(category)
			transformedEvent := events.DomainEvent{
				ID:            fmt.Sprintf("category_%s_%d", category.ID, time.Now().Unix()),
				Type:          events.EventCategoryUpdated,
				AggregateID:   category.ID,
				AggregateType: "category",
				Timestamp:     time.Now(),
				Version:       1,
				Data:          json.RawMessage(categoryData),
				Source:        "loyverse-api",
			}

			if err := publisher.Publish(ctx, transformedEvent); err != nil {
				log.Printf("Error publishing category event for %s: %v", category.ID, err)
				// Don't fail the request if Kafka publishing fails
			}
		}

		// Return success response
		response := map[string]interface{}{
			"success":        true,
			"message":        "Categories fetched and stored successfully",
			"categories_count": len(categories),
			"timestamp":      time.Now().Format(time.RFC3339),
			"source":         "loyverse_api",
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	})

	// API endpoint to test various Loyverse endpoints
	http.HandleFunc("/api/test/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if cfg.LoyverseAPIToken == "" {
			http.Error(w, "Loyverse API token not configured", http.StatusBadRequest)
			return
		}

		// Extract endpoint name from URL path
		endpoint := r.URL.Path[len("/api/test/"):]
		if endpoint == "" || endpoint == "help" {
			// Get all available endpoints from client
			client := connector.NewClient(cfg.LoyverseAPIToken)
			allEndpoints := client.GetAvailableEndpoints()
			
			endpoints := make([]string, 0, len(allEndpoints))
			for name := range allEndpoints {
				endpoints = append(endpoints, name)
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"available_endpoints": endpoints,
				"endpoint_urls": allEndpoints,
				"usage": "GET /api/test/{endpoint}",
				"examples": []string{
					"GET /api/test/stores",
					"GET /api/test/categories", 
					"GET /api/test/items",
					"GET /api/test/receipts",
					"GET /api/test/customers",
					"GET /api/test/store_stocks",
				},
				"special_endpoints": map[string]string{
					"test_all": "GET /api/test/test_all - Test all endpoints at once",
					"help": "GET /api/test/help - Show this help",
				},
			})
			return
		}

		// Create Loyverse API client
		client := connector.NewClient(cfg.LoyverseAPIToken)
		ctx := context.Background()

		var data []json.RawMessage
		var err error

		// Route to appropriate method based on endpoint
		switch endpoint {
		case "test_all":
			// Test all endpoints at once
			results := client.TestAllEndpoints(ctx)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(results)
			return
		case "stores":
			data, err = client.GetStores(ctx)
		case "customers":
			data, err = client.GetCustomers(ctx)
		case "employees":
			data, err = client.GetEmployees(ctx)
		case "discounts":
			data, err = client.GetDiscounts(ctx)
		case "modifiers":
			data, err = client.GetModifiers(ctx)
		case "taxes":
			data, err = client.GetTaxes(ctx)
		case "payment_types":
			data, err = client.GetPaymentTypes(ctx)
		case "variants":
			data, err = client.GetVariants(ctx)
		case "suppliers":
			data, err = client.GetSuppliers(ctx)
		case "purchase_orders":
			data, err = client.GetPurchaseOrders(ctx)
		case "pos_devices":
			data, err = client.GetPOSDevices(ctx)
		case "cash_registers":
			data, err = client.GetCashRegisters(ctx)
		case "webhooks":
			data, err = client.GetWebhooks(ctx)
		case "store_stocks":
			data, err = client.GetInventoryLevels(ctx)
		case "categories":
			loyverseAPI := connector.NewLoyverseAPI(cfg.LoyverseAPIToken)
			categories, err := loyverseAPI.GetCategories()
			if err == nil {
				for _, cat := range categories {
					catData, _ := json.Marshal(cat)
					data = append(data, json.RawMessage(catData))
				}
			}
		case "items":
			data, err = client.GetProducts(ctx)
		case "inventory":
			data, err = client.GetInventoryLevels(ctx)
		case "recent_receipts":
			data, err = client.GetRecentReceipts(ctx)
		case "receipts":
			data, err = client.GetReceiptsWithParams(ctx, nil)
		case "account":
			rawData, err := client.GetAccount(ctx)
			if err == nil {
				data = append(data, json.RawMessage(rawData))
			}
		case "settings":
			rawData, err := client.GetSettings(ctx)
			if err == nil {
				data = append(data, json.RawMessage(rawData))
			}
		default:
			// Try as raw endpoint
			rawData, rawErr := client.GetEndpoint(ctx, "/"+endpoint)
			if rawErr != nil {
				http.Error(w, fmt.Sprintf("Unknown endpoint '%s' or API error: %v", endpoint, rawErr), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(rawData)
			return
		}

		if err != nil {
			log.Printf("Error fetching %s: %v", endpoint, err)
			http.Error(w, fmt.Sprintf("Failed to fetch %s: %v", endpoint, err), http.StatusInternalServerError)
			return
		}

		// Return results
		response := map[string]interface{}{
			"endpoint": endpoint,
			"count":    len(data),
			"data":     data,
			"timestamp": time.Now().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	})

	log.Println("Registered endpoints:")
	log.Println("  GET  /health")
	log.Println("  POST /webhook/loyverse")
	log.Println("  GET  /api/latest-receipt")
	log.Println("  GET  /api/categories")
	log.Println("  POST /api/categories/fetch")
	log.Println("  GET  /api/test/{endpoint}")

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Server starting on %s", addr)
	
	server := &http.Server{Addr: addr}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}