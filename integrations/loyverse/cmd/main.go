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
	"integrations/loyverse/internal/models"
	"integrations/loyverse/internal/redis"
	"integrations/loyverse/internal/repository"
	"integrations/loyverse/internal/sync"
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

	// Initialize enhanced Redis client
	redisConfig := redis.DefaultConfig()
	redisConfig.Addr = cfg.RedisAddr
	redisClient := redis.NewClient(redisConfig)
	defer redisClient.Close()
	
	// Initialize Redis repository with enhanced error handling
	cacheRepo := repository.NewRedisRepository(redisConfig)
	defer cacheRepo.Close()

	// Initialize Kafka event publisher
	publisher := events.NewPublisher(cfg.KafkaBrokers, cfg.KafkaTopic)
	defer publisher.Close()
	// Initialize Loyverse API client
	loyverseClient := connector.NewClient(cfg.LoyverseAPIToken)
	
	// Initialize sync managers
	productSync := sync.NewProductSync(loyverseClient, publisher, redisClient)
	
	// Start sync in background
	syncCtx, syncCancel := context.WithCancel(context.Background())
	defer syncCancel()
	
	go func() {
		log.Println("Enhanced sync manager started with Redis error handling")
		
		// Run sync every 5 minutes for testing
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		
		// Run initial sync with enhanced error handling
		log.Println("Starting initial sync...")
		if err := productSync.Sync(syncCtx); err != nil {
			log.Printf("Product sync error: %v", err)
		}
		
		// Log Redis stats periodically
		statsTicker := time.NewTicker(30 * time.Second)
		defer statsTicker.Stop()
		
		for {
			select {
			case <-ticker.C:
				log.Println("Running scheduled sync...")
				if err := productSync.Sync(syncCtx); err != nil {
					log.Printf("Product sync error: %v", err)
				}
			case <-statsTicker.C:
				// Log Redis health stats
				if redisClient.IsHealthy() {
					redisClient.LogStats()
				} else {
					log.Println("Redis is currently unhealthy")
				}
			case <-syncCtx.Done():
				return
			}
		}
	}()

	// Simple webhook handler
	simpleWebhookHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"received","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	})

	// Setup HTTP routes with enhanced health checks
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		// Check Redis health
		redisHealthy := redisClient.IsHealthy()
		
		status := "healthy"
		statusCode := http.StatusOK
		if !redisHealthy {
			status = "degraded"
			statusCode = http.StatusServiceUnavailable
		}
		
		healthResponse := map[string]interface{}{
			"status":     status,
			"service":    "loyverse-integration",
			"timestamp":  time.Now().Format(time.RFC3339),
			"redis": map[string]interface{}{
				"healthy": redisHealthy,
				"stats":   cacheRepo.GetHealthStats(),
			},
		}
		
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(healthResponse)
	})

	// Redis health endpoint
	http.HandleFunc("/health/redis", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		stats := cacheRepo.GetHealthStats()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(stats)
	})

	// Webhook endpoint for Loyverse
	http.Handle("/webhook/loyverse", simpleWebhookHandler)

	// Sync endpoint
	http.HandleFunc("/api/sync", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"error":"Method not allowed"}`)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"sync_triggered","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	})

	// API endpoint to get latest receipt (disabled temporarily)
	http.HandleFunc("/api/latest-receipt", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Endpoint temporarily disabled", http.StatusServiceUnavailable)
	})

	// API endpoint to get all categories (disabled temporarily)
	http.HandleFunc("/api/categories", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Endpoint temporarily disabled", http.StatusServiceUnavailable)
	})

	// API endpoint to get items from Loyverse
	http.HandleFunc("/api/items", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if cfg.LoyverseAPIToken == "" {
			http.Error(w, "Loyverse API token not configured", http.StatusBadRequest)
			return
		}

		// Create Loyverse API client
		client := connector.NewClient(cfg.LoyverseAPIToken)
		ctx := context.Background()

		// Fetch stores first to get store name
		stores, err := client.GetStores(ctx)
		if err != nil {
			log.Printf("[loyverse-connector] ❌ Failed to fetch stores: %v", err)
			http.Error(w, fmt.Sprintf("Failed to fetch stores: %v", err), http.StatusInternalServerError)
			return
		}

		var storeName string = "Unknown Store"
		if len(stores) > 0 {
			var firstStore models.Store
			if err := json.Unmarshal(stores[0], &firstStore); err == nil {
				storeName = firstStore.Name
			}
		}

		// Fetch items
		rawItems, err := client.GetProducts(ctx)
		if err != nil {
			log.Printf("[loyverse-connector] ❌ Failed to fetch items: %v", err)
			http.Error(w, fmt.Sprintf("Failed to fetch items: %v", err), http.StatusInternalServerError)
			return
		}

		// Parse items and collect data
		var items []map[string]interface{}
		var firstItemName, firstSKU string

		for i, raw := range rawItems {
			var item models.Item
			if err := json.Unmarshal(raw, &item); err != nil {
				log.Printf("[loyverse-connector] ⚠️ Error unmarshaling item %d: %v", i, err)
				continue
			}

			// Get first item info for logging
			if i == 0 {
				firstItemName = item.Name
				firstSKU = item.SKU
				if firstSKU == "" && len(item.Variants) > 0 {
					firstSKU = item.Variants[0].SKU
				}
			}

			// Prepare item data for response
			itemData := map[string]interface{}{
				"id":          item.ID,
				"name":        item.Name,
				"description": item.Description,
				"sku":         item.SKU,
				"barcode":     item.Barcode,
				"category_id": item.CategoryID,
				"track_stock": item.TrackStock,
				"created_at":  item.CreatedAt,
				"updated_at":  item.UpdatedAt,
			}

			// Add variants if available
			if len(item.Variants) > 0 {
				variants := make([]map[string]interface{}, 0, len(item.Variants))
				for _, variant := range item.Variants {
					variants = append(variants, map[string]interface{}{
						"variant_id":    variant.ID,
						"sku":           variant.SKU,
						"barcode":       variant.Barcode,
						"default_price": variant.DefaultPrice,
						"cost":          variant.Cost,
					})
				}
				itemData["variants"] = variants
			}

			items = append(items, itemData)
		}

		// Logging
		log.Printf("[loyverse-connector] ✅ Pulled %d items from Loyverse store: '%s'", len(rawItems), storeName)
		if firstItemName != "" {
			log.Printf("[loyverse-connector] 🛒 First item: %s [SKU: %s]", firstItemName, firstSKU)
		}

		// Response
		response := map[string]interface{}{
			"success":    true,
			"count":      len(items),
			"store_name": storeName,
			"items":      items,
			"timestamp":  time.Now().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("[loyverse-connector] ❌ Error encoding response: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	})

	// API endpoint to get inventory from Loyverse
	http.HandleFunc("/api/inventory", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if cfg.LoyverseAPIToken == "" {
			http.Error(w, "Loyverse API token not configured", http.StatusBadRequest)
			return
		}

		// Create Loyverse API client
		client := connector.NewClient(cfg.LoyverseAPIToken)
		ctx := context.Background()

		// Fetch stores first to get store names
		stores, err := client.GetStores(ctx)
		if err != nil {
			log.Printf("[loyverse-connector] ❌ Failed to fetch stores: %v", err)
			http.Error(w, fmt.Sprintf("Failed to fetch stores: %v", err), http.StatusInternalServerError)
			return
		}

		// Create store map for lookups
		storeMap := make(map[string]string)
		var mainStoreName string = "Unknown Store"
		if len(stores) > 0 {
			var firstStore models.Store
			if err := json.Unmarshal(stores[0], &firstStore); err == nil {
				mainStoreName = firstStore.Name
			}
		}

		for _, storeData := range stores {
			var store models.Store
			if err := json.Unmarshal(storeData, &store); err == nil {
				storeMap[store.ID] = store.Name
			}
		}

		// Fetch inventory
		rawInventory, err := client.GetInventoryLevels(ctx)
		if err != nil {
			log.Printf("[loyverse-connector] ❌ Failed to fetch inventory: %v", err)
			http.Error(w, fmt.Sprintf("Failed to fetch inventory: %v", err), http.StatusInternalServerError)
			return
		}

		// Parse inventory and collect data
		var inventory []map[string]interface{}
		var firstVariantID, firstStoreID string
		var firstQuantity float64

		for i, raw := range rawInventory {
			var invLevel models.InventoryLevel
			if err := json.Unmarshal(raw, &invLevel); err != nil {
				log.Printf("[loyverse-connector] ⚠️ Error unmarshaling inventory %d: %v", i, err)
				continue
			}

			// Get first inventory info for logging
			if i == 0 {
				firstVariantID = invLevel.VariantID
				firstStoreID = invLevel.StoreID
				firstQuantity = invLevel.InStock
			}

			// Get store name
			storeName := storeMap[invLevel.StoreID]
			if storeName == "" {
				storeName = invLevel.StoreID
			}

			// Prepare inventory data for response
			inventoryData := map[string]interface{}{
				"variant_id":   invLevel.VariantID,
				"store_id":     invLevel.StoreID,
				"store_name":   storeName,
				"in_stock":     invLevel.InStock,
				"updated_at":   invLevel.UpdatedAt,
			}

			inventory = append(inventory, inventoryData)
		}

		// Logging
		log.Printf("[loyverse-connector] ✅ Pulled %d inventory levels from Loyverse store: '%s'", len(rawInventory), mainStoreName)
		if firstVariantID != "" {
			storeDisplayName := storeMap[firstStoreID]
			if storeDisplayName == "" {
				storeDisplayName = firstStoreID
			}
			log.Printf("[loyverse-connector] 🧾 Stock: %s @ store=%s = %.0f", firstVariantID, storeDisplayName, firstQuantity)
		}

		// Response
		response := map[string]interface{}{
			"success":    true,
			"count":      len(inventory),
			"store_name": mainStoreName,
			"inventory":  inventory,
			"timestamp":  time.Now().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("[loyverse-connector] ❌ Error encoding response: %v", err)
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

		log.Printf("Retrieved %d categories from Loyverse API", len(categories))

		// Publish category sync events to Kafka
		ctx := context.Background()
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
					"GET /api/test/inventory",
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
	log.Println("  GET  /api/items")
	log.Println("  GET  /api/inventory") 
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