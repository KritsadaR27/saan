// integrations/loyverse/internal/sync/simplified_product_sync.go
package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"integrations/loyverse/internal/connector"
	"integrations/loyverse/internal/models"
	"integrations/loyverse/internal/redis"
)

// SimplifiedProductSync - Direct sync without Kafka complexity
type SimplifiedProductSync struct {
	loyverseClient    *connector.Client
	inventoryBaseURL  string
	httpClient        *http.Client
	redis             *redis.Client
}

// UpsertProductRequest - Request structure for inventory service
type UpsertProductRequest struct {
	ProductID          string           `json:"product_id"`
	Name               string           `json:"name"`
	Description        string           `json:"description"`
	CategoryID         *string          `json:"category_id"`
	PrimarySupplierID  *string          `json:"primary_supplier_id"`
	SKU                string           `json:"sku"`
	Barcode            string           `json:"barcode"`
	TrackStock         bool             `json:"track_stock"`
	SoldByWeight       bool             `json:"sold_by_weight"`
	IsComposite        bool             `json:"is_composite"`
	UseProduction      bool             `json:"use_production"`
	Variants           []models.Variant `json:"variants"`
	Source             string           `json:"source"`
	UpdatedAt          time.Time        `json:"updated_at"`
}

// UpsertProductResponse - Response from inventory service
type UpsertProductResponse struct {
	Success   bool   `json:"success"`
	ProductID string `json:"product_id"`
	Message   string `json:"message"`
}

// NewSimplifiedProductSync creates a new simplified product sync
func NewSimplifiedProductSync(loyverseClient *connector.Client, inventoryBaseURL string, redis *redis.Client) *SimplifiedProductSync {
	return &SimplifiedProductSync{
		loyverseClient:   loyverseClient,
		inventoryBaseURL: inventoryBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		redis: redis,
	}
}

// Sync performs direct synchronization
func (s *SimplifiedProductSync) Sync(ctx context.Context) error {
	log.Println("🔄 Starting simplified product sync...")
	
	startTime := time.Now()
	
	// 1. Get products from Loyverse
	rawData, err := s.loyverseClient.GetProducts(ctx)
	if err != nil {
		return fmt.Errorf("fetching products from Loyverse: %w", err)
	}
	
	log.Printf("📦 Retrieved %d products from Loyverse", len(rawData))
	
	// 2. Process each product
	successCount := 0
	errorCount := 0
	
	for i, raw := range rawData {
		var item models.Item
		if err := json.Unmarshal(raw, &item); err != nil {
			log.Printf("❌ Error unmarshaling product %d: %v", i, err)
			errorCount++
			continue
		}
		
		// 3. Send directly to Inventory Service
		if err := s.syncProductToInventory(ctx, item); err != nil {
			log.Printf("❌ Failed to sync product %s: %v", item.ID, err)
			errorCount++
			continue
		}
		
		// 4. Cache for quick access (optional)
		s.cacheProduct(ctx, item, raw)
		
		successCount++
		
		// Log progress every 10 products
		if (i+1)%10 == 0 {
			log.Printf("📊 Progress: %d/%d products synced", i+1, len(rawData))
		}
	}
	
	duration := time.Since(startTime)
	
	// 5. Update sync statistics
	s.updateSyncStats(ctx, successCount, errorCount, duration)
	
	log.Printf("✅ Sync completed in %v - Success: %d, Errors: %d", duration, successCount, errorCount)
	
	if errorCount > 0 {
		return fmt.Errorf("sync completed with %d errors out of %d products", errorCount, len(rawData))
	}
	
	return nil
}

// syncProductToInventory sends product data directly to inventory service
func (s *SimplifiedProductSync) syncProductToInventory(ctx context.Context, item models.Item) error {
	// Prepare request payload
	request := UpsertProductRequest{
		ProductID:         item.ID,
		Name:              item.Name,
		Description:       item.Description,
		CategoryID:        item.CategoryID,
		PrimarySupplierID: item.PrimarySupplierID,
		SKU:               item.SKU,
		Barcode:           item.Barcode,
		TrackStock:        item.TrackStock,
		SoldByWeight:      item.SoldByWeight,
		IsComposite:       item.IsComposite,
		UseProduction:     item.UseProduction,
		Variants:          item.Variants,
		Source:            "loyverse",
		UpdatedAt:         item.UpdatedAt,
	}
	
	// Marshal to JSON
	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}
	
	// Create HTTP request
	url := fmt.Sprintf("%s/api/v1/products/upsert", s.inventoryBaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Source", "loyverse-integration")
	
	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("inventory service returned status %d", resp.StatusCode)
	}
	
	// Parse response
	var response UpsertProductResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}
	
	if !response.Success {
		return fmt.Errorf("inventory service failed: %s", response.Message)
	}
	
	return nil
}

// cacheProduct caches product data in Redis for quick access
func (s *SimplifiedProductSync) cacheProduct(ctx context.Context, item models.Item, raw []byte) {
	// Cache product data
	productKey := fmt.Sprintf("loyverse:product:%s", item.ID)
	if err := s.redis.SafeSet(ctx, productKey, string(raw), 24*time.Hour); err != nil {
		log.Printf("⚠️  Warning: Failed to cache product %s: %v", item.ID, err)
	}
	
	// Cache variants
	for _, variant := range item.Variants {
		variantKey := fmt.Sprintf("loyverse:variant:%s", variant.ID)
		variantData, _ := json.Marshal(variant)
		if err := s.redis.SafeSet(ctx, variantKey, string(variantData), 24*time.Hour); err != nil {
			log.Printf("⚠️  Warning: Failed to cache variant %s: %v", variant.ID, err)
		}
	}
}

// updateSyncStats updates sync statistics in Redis
func (s *SimplifiedProductSync) updateSyncStats(ctx context.Context, successCount, errorCount int, duration time.Duration) {
	stats := map[string]interface{}{
		"success_count":    successCount,
		"error_count":      errorCount,
		"duration_seconds": duration.Seconds(),
		"last_sync_at":     time.Now().Format(time.RFC3339),
	}
	
	statsData, _ := json.Marshal(stats)
	if err := s.redis.SafeSet(ctx, "loyverse:sync:stats:products", string(statsData), 7*24*time.Hour); err != nil {
		log.Printf("⚠️  Warning: Failed to update sync stats: %v", err)
	}
}

// GetSyncStats returns the last sync statistics
func (s *SimplifiedProductSync) GetSyncStats(ctx context.Context) (map[string]interface{}, error) {
	data, err := s.redis.SafeGet(ctx, "loyverse:sync:stats:products")
	if err != nil {
		return nil, fmt.Errorf("getting sync stats: %w", err)
	}
	
	var stats map[string]interface{}
	if err := json.Unmarshal([]byte(data), &stats); err != nil {
		return nil, fmt.Errorf("unmarshaling stats: %w", err)
	}
	
	return stats, nil
}
