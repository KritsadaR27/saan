// integrations/loyverse/internal/sync/product_sync.go
package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"integrations/loyverse/internal/connector"
	"integrations/loyverse/internal/events"
	"integrations/loyverse/internal/models"
)

type ProductSync struct {
	client      *connector.Client
	publisher   *events.Publisher
	redis       *redis.Client
	transformer *events.Transformer
}

func NewProductSync(client *connector.Client, publisher *events.Publisher, redis *redis.Client) *ProductSync {
	return &ProductSync{
		client:      client,
		publisher:   publisher,
		redis:       redis,
		transformer: events.NewTransformer(),
	}
}

func (s *ProductSync) Sync(ctx context.Context) error {
	log.Println("Starting product sync...")
	
	// Get last sync time
	lastSyncKey := "loyverse:sync:product:last"
	lastSync, _ := s.redis.Get(ctx, lastSyncKey).Result()
	
	// Fetch products
	rawData, err := s.client.GetProducts(ctx)
	if err != nil {
		return fmt.Errorf("fetching products: %w", err)
	}
	
	var domainEvents []events.DomainEvent
	processedCount := 0
	
	for _, raw := range rawData {
		var item models.Item
		if err := json.Unmarshal(raw, &item); err != nil {
			log.Printf("Error unmarshaling item: %v", err)
			continue
		}
		
		// Skip if not updated since last sync
		if lastSync != "" {
			lastSyncTime, _ := time.Parse(time.RFC3339, lastSync)
			if item.UpdatedAt.Before(lastSyncTime) {
				continue
			}
		}
		
		// Create domain event
		event := s.createProductEvent(item)
		domainEvents = append(domainEvents, event)
		processedCount++
		
		// Cache product data
		cacheKey := fmt.Sprintf("loyverse:product:%s", item.ID)
		s.redis.Set(ctx, cacheKey, raw, 24*time.Hour)
		
		// Cache variants
		for _, variant := range item.Variants {
			variantKey := fmt.Sprintf("loyverse:variant:%s", variant.ID)
			variantData, _ := json.Marshal(variant)
			s.redis.Set(ctx, variantKey, variantData, 24*time.Hour)
		}
	}
	
	// Update product count
	countKey := "loyverse:sync:count:products"
	s.redis.Set(ctx, countKey, processedCount, 0)
	
	// Publish events
	if len(domainEvents) > 0 {
		if err := s.publisher.PublishBatch(ctx, domainEvents); err != nil {
			return fmt.Errorf("publishing events: %w", err)
		}
	}
	
	// Update last sync time
	s.redis.Set(ctx, lastSyncKey, time.Now().Format(time.RFC3339), 0)
	
	log.Printf("Product sync completed. Processed %d products", processedCount)
	return nil
}

func (s *ProductSync) createProductEvent(item models.Item) events.DomainEvent {
	eventData := map[string]interface{}{
		"product_id":          item.ID,
		"name":                item.Name,
		"description":         item.Description,
		"category_id":         item.CategoryID,
		"primary_supplier_id": item.PrimarySupplierID,
		"sku":                 item.SKU,
		"barcode":             item.Barcode,
		"track_stock":         item.TrackStock,
		"sold_by_weight":      item.SoldByWeight,
		"is_composite":        item.IsComposite,
		"use_production":      item.UseProduction,
		"variants":            item.Variants,
		"source":              "loyverse",
	}
	
	data, _ := json.Marshal(eventData)
	
	return events.DomainEvent{
		ID:            fmt.Sprintf("product-%s-%d", item.ID, time.Now().Unix()),
		Type:          events.EventProductUpdated,
		AggregateID:   item.ID,
		AggregateType: "product",
		Timestamp:     time.Now(),
		Version:       1,
		Data:          data,
		Source:        "loyverse-integration",
	}
}

// integrations/loyverse/internal/sync/inventory_sync.go
package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"integrations/loyverse/internal/connector"
	"integrations/loyverse/internal/events"
	"integrations/loyverse/internal/models"
)

type InventorySync struct {
	client      *connector.Client
	publisher   *events.Publisher
	redis       *redis.Client
	transformer *events.Transformer
}

func NewInventorySync(client *connector.Client, publisher *events.Publisher, redis *redis.Client) *InventorySync {
	return &InventorySync{
		client:      client,
		publisher:   publisher,
		redis:       redis,
		transformer: events.NewTransformer(),
	}
}

func (s *InventorySync) Sync(ctx context.Context) error {
	log.Println("Starting inventory sync...")
	
	// Fetch inventory levels
	rawData, err := s.client.GetInventoryLevels(ctx)
	if err != nil {
		return fmt.Errorf("fetching inventory levels: %w", err)
	}
	
	var domainEvents []events.DomainEvent
	processedCount := 0
	
	// Group by store for caching
	storeInventory := make(map[string][]models.InventoryLevel)
	
	for _, raw := range rawData {
		var level models.InventoryLevel
		if err := json.Unmarshal(raw, &level); err != nil {
			log.Printf("Error unmarshaling inventory level: %v", err)
			continue
		}
		
		// Create domain event
		event, err := s.transformer.TransformInventory(raw)
		if err != nil {
			log.Printf("Error transforming inventory: %v", err)
			continue
		}
		
		domainEvents = append(domainEvents, event)
		processedCount++
		
		// Cache inventory level
		cacheKey := fmt.Sprintf("loyverse:inventory:%s:%s", level.StoreID, level.VariantID)
		s.redis.Set(ctx, cacheKey, raw, 1*time.Hour) // Shorter TTL for inventory
		
		// Add to store grouping
		storeInventory[level.StoreID] = append(storeInventory[level.StoreID], level)
	}
	
	// Cache store inventory summaries
	for storeID, levels := range storeInventory {
		summaryKey := fmt.Sprintf("loyverse:inventory:store:%s", storeID)
		summaryData, _ := json.Marshal(levels)
		s.redis.Set(ctx, summaryKey, summaryData, 1*time.Hour)
	}
	
	// Update inventory count
	countKey := "loyverse:sync:count:inventory"
	s.redis.Set(ctx, countKey, processedCount, 0)
	
	// Publish events
	if len(domainEvents) > 0 {
		if err := s.publisher.PublishBatch(ctx, domainEvents); err != nil {
			return fmt.Errorf("publishing events: %w", err)
		}
	}
	
	log.Printf("Inventory sync completed. Processed %d inventory levels", processedCount)
	return nil
}

// integrations/loyverse/internal/sync/receipt_sync.go
package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"integrations/loyverse/internal/connector"
	"integrations/loyverse/internal/events"
	"integrations/loyverse/internal/models"
)

type ReceiptSync struct {
	client      *connector.Client
	publisher   *events.Publisher
	redis       *redis.Client
	transformer *events.Transformer
}

func NewReceiptSync(client *connector.Client, publisher *events.Publisher, redis *redis.Client) *ReceiptSync {
	return &ReceiptSync{
		client:      client,
		publisher:   publisher,
		redis:       redis,
		transformer: events.NewTransformer(),
	}
}

func (s *ReceiptSync) Sync(ctx context.Context) error {
	log.Println("Starting receipt sync...")
	
	// Get last sync time
	lastSyncKey := "loyverse:sync:receipt:last"
	lastSync, _ := s.redis.Get(ctx, lastSyncKey).Result()
	
	var since time.Time
	if lastSync != "" {
		since, _ = time.Parse(time.RFC3339, lastSync)
	} else {
		// Default to 24 hours ago if no last sync
		since = time.Now().Add(-24 * time.Hour)
	}
	
	// Fetch receipts since last sync
	rawData, err := s.client.GetReceipts(ctx, since)
	if err != nil {
		return fmt.Errorf("fetching receipts: %w", err)
	}
	
	var domainEvents []events.DomainEvent
	processedCount := 0
	totalAmount := 0.0
	
	for _, raw := range rawData {
		var receipt models.Receipt
		if err := json.Unmarshal(raw, &receipt); err != nil {
			log.Printf("Error unmarshaling receipt: %v", err)
			continue
		}
		
		// Skip cancelled receipts in sync (but still process them)
		if receipt.CancelledAt != nil {
			log.Printf("Skipping cancelled receipt: %s", receipt.Number)
			continue
		}
		
		// Create domain event
		event, err := s.transformer.TransformReceipt(raw)
		if err != nil {
			log.Printf("Error transforming receipt: %v", err)
			continue
		}
		
		domainEvents = append(domainEvents, event)
		processedCount++
		totalAmount += receipt.TotalMoney
		
		// Cache receipt
		cacheKey := fmt.Sprintf("loyverse:receipt:%s", receipt.Number)
		s.redis.Set(ctx, cacheKey, raw, 7*24*time.Hour)
		
		// Update daily revenue
		dayKey := fmt.Sprintf("loyverse:revenue:%s:%s", 
			receipt.StoreID, 
			receipt.ReceiptDate.Format("2006-01-02"))
		s.redis.IncrByFloat(ctx, dayKey, receipt.TotalMoney)
		s.redis.Expire(ctx, dayKey, 30*24*time.Hour)
	}
	
	// Update receipt stats
	statsKey := "loyverse:sync:stats:receipts"
	stats := map[string]interface{}{
		"last_sync":     time.Now().Format(time.RFC3339),
		"processed":     processedCount,
		"total_amount":  totalAmount,
	}
	statsData, _ := json.Marshal(stats)
	s.redis.Set(ctx, statsKey, statsData, 0)
	
	// Publish events
	if len(domainEvents) > 0 {
		if err := s.publisher.PublishBatch(ctx, domainEvents); err != nil {
			return fmt.Errorf("publishing events: %w", err)
		}
	}
	
	// Update last sync time
	s.redis.Set(ctx, lastSyncKey, time.Now().Format(time.RFC3339), 0)
	
	log.Printf("Receipt sync completed. Processed %d receipts, total amount: %.2f", processedCount, totalAmount)
	return nil
}

// integrations/loyverse/internal/sync/customer_sync.go
package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"integrations/loyverse/internal/connector"
	"integrations/loyverse/internal/events"
	"integrations/loyverse/internal/models"
)

type CustomerSync struct {
	client      *connector.Client
	publisher   *events.Publisher
	redis       *redis.Client
	transformer *events.Transformer
}

func NewCustomerSync(client *connector.Client, publisher *events.Publisher, redis *redis.Client) *CustomerSync {
	return &CustomerSync{
		client:      client,
		publisher:   publisher,
		redis:       redis,
		transformer: events.NewTransformer(),
	}
}

func (s *CustomerSync) Sync(ctx context.Context) error {
	log.Println("Starting customer sync...")
	
	// Get last sync time
	lastSyncKey := "loyverse:sync:customer:last"
	lastSync, _ := s.redis.Get(ctx, lastSyncKey).Result()
	
	// Fetch customers
	rawData, err := s.client.GetCustomers(ctx)
	if err != nil {
		return fmt.Errorf("fetching customers: %w", err)
	}
	
	var domainEvents []events.DomainEvent
	processedCount := 0
	
	for _, raw := range rawData {
		var customer models.Customer
		if err := json.Unmarshal(raw, &customer); err != nil {
			log.Printf("Error unmarshaling customer: %v", err)
			continue
		}
		
		// Skip if not updated since last sync
		if lastSync != "" {
			lastSyncTime, _ := time.Parse(time.RFC3339, lastSync)
			if customer.UpdatedAt.Before(lastSyncTime) {
				continue
			}
		}
		
		// Create domain event
		event := s.createCustomerEvent(customer)
		domainEvents = append(domainEvents, event)
		processedCount++
		
		// Cache customer data
		cacheKey := fmt.Sprintf("loyverse:customer:%s", customer.ID)
		s.redis.Set(ctx, cacheKey, raw, 24*time.Hour)
		
		// Index by phone number for quick lookup
		if customer.Phone != "" {
			phoneKey := fmt.Sprintf("loyverse:customer:phone:%s", customer.Phone)
			s.redis.Set(ctx, phoneKey, customer.ID, 24*time.Hour)
		}
	}
	
	// Update customer count
	countKey := "loyverse:sync:count:customers"
	s.redis.Set(ctx, countKey, processedCount, 0)
	
	// Publish events
	if len(domainEvents) > 0 {
		if err := s.publisher.PublishBatch(ctx, domainEvents); err != nil {
			return fmt.Errorf("publishing events: %w", err)
		}
	}
	
	// Update last sync time
	s.redis.Set(ctx, lastSyncKey, time.Now().Format(time.RFC3339), 0)
	
	log.Printf("Customer sync completed. Processed %d customers", processedCount)
	return nil
}

func (s *CustomerSync) createCustomerEvent(customer models.Customer) events.DomainEvent {
	eventData := map[string]interface{}{
		"customer_id":   customer.ID,
		"name":          customer.Name,
		"email":         customer.Email,
		"phone":         customer.Phone,
		"address":       customer.Address,
		"city":          customer.City,
		"postal_code":   customer.PostalCode,
		"country_code":  customer.CountryCode,
		"points_balance": customer.PointsBalance,
		"total_spent":   customer.TotalSpent,
		"total_orders":  customer.TotalOrders,
		"source":        "loyverse",
	}
	
	data, _ := json.Marshal(eventData)
	
	return events.DomainEvent{
		ID:            fmt.Sprintf("customer-%s-%d", customer.ID, time.Now().Unix()),
		Type:          events.EventCustomerUpdated,
		AggregateID:   customer.ID,
		AggregateType: "customer",
		Timestamp:     time.Now(),
		Version:       1,
		Data:          data,
		Source:        "loyverse-integration",
	}
}