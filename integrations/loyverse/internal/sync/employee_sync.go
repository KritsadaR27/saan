// integrations/loyverse/internal/sync/employee_sync.go
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

type EmployeeSync struct {
	client      *connector.Client
	publisher   *events.Publisher
	redis       *redis.Client
	transformer *events.Transformer
}

func NewEmployeeSync(client *connector.Client, publisher *events.Publisher, redis *redis.Client) *EmployeeSync {
	return &EmployeeSync{
		client:      client,
		publisher:   publisher,
		redis:       redis,
		transformer: events.NewTransformer(),
	}
}

func (s *EmployeeSync) Sync(ctx context.Context) error {
	log.Println("Starting employee sync...")
	
	// Get last sync time
	lastSyncKey := "loyverse:sync:employee:last"
	lastSync, _ := s.redis.Get(ctx, lastSyncKey).Result()
	
	// Fetch employees
	rawData, err := s.client.GetWithPagination(ctx, "/employees", 250)
	if err != nil {
		return fmt.Errorf("fetching employees: %w", err)
	}
	
	var domainEvents []events.DomainEvent
	processedCount := 0
	
	for _, raw := range rawData {
		var employee models.Employee
		if err := json.Unmarshal(raw, &employee); err != nil {
			log.Printf("Error unmarshaling employee: %v", err)
			continue
		}
		
		// Skip if not updated since last sync
		if lastSync != "" {
			lastSyncTime, _ := time.Parse(time.RFC3339, lastSync)
			if employee.UpdatedAt.Before(lastSyncTime) {
				continue
			}
		}
		
		// Create domain event
		event := s.createEmployeeEvent(employee)
		domainEvents = append(domainEvents, event)
		processedCount++
		
		// Cache employee data
		cacheKey := fmt.Sprintf("loyverse:employee:%s", employee.ID)
		s.redis.Set(ctx, cacheKey, raw, 24*time.Hour)
	}
	
	// Publish events
	if len(domainEvents) > 0 {
		if err := s.publisher.PublishBatch(ctx, domainEvents); err != nil {
			return fmt.Errorf("publishing events: %w", err)
		}
	}
	
	// Update last sync time
	s.redis.Set(ctx, lastSyncKey, time.Now().Format(time.RFC3339), 0)
	
	log.Printf("Employee sync completed. Processed %d employees", processedCount)
	return nil
}

func (s *EmployeeSync) createEmployeeEvent(employee models.Employee) events.DomainEvent {
	eventData := map[string]interface{}{
		"employee_id": employee.ID,
		"name":        employee.Name,
		"email":       employee.Email,
		"phone":       employee.Phone,
		"store_id":    employee.StoreID,
		"roles":       employee.Roles,
		"is_owner":    employee.IsOwner,
		"source":      "loyverse",
	}
	
	data, _ := json.Marshal(eventData)
	
	return events.DomainEvent{
		ID:            fmt.Sprintf("emp-%s-%d", employee.ID, time.Now().Unix()),
		Type:          events.EventEmployeeUpdated,
		AggregateID:   employee.ID,
		AggregateType: "employee",
		Timestamp:     time.Now(),
		Version:       1,
		Data:          data,
		Source:        "loyverse-integration",
	}
}

// integrations/loyverse/internal/sync/discount_sync.go
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

type DiscountSync struct {
	client      *connector.Client
	publisher   *events.Publisher
	redis       *redis.Client
	transformer *events.Transformer
}

func NewDiscountSync(client *connector.Client, publisher *events.Publisher, redis *redis.Client) *DiscountSync {
	return &DiscountSync{
		client:      client,
		publisher:   publisher,
		redis:       redis,
		transformer: events.NewTransformer(),
	}
}

func (s *DiscountSync) Sync(ctx context.Context) error {
	log.Println("Starting discount sync...")
	
	// Get last sync time
	lastSyncKey := "loyverse:sync:discount:last"
	lastSync, _ := s.redis.Get(ctx, lastSyncKey).Result()
	
	// Fetch discounts
	rawData, err := s.client.GetWithPagination(ctx, "/discounts", 250)
	if err != nil {
		return fmt.Errorf("fetching discounts: %w", err)
	}
	
	var domainEvents []events.DomainEvent
	processedCount := 0
	
	for _, raw := range rawData {
		var discount models.Discount
		if err := json.Unmarshal(raw, &discount); err != nil {
			log.Printf("Error unmarshaling discount: %v", err)
			continue
		}
		
		// Skip if not updated since last sync
		if lastSync != "" {
			lastSyncTime, _ := time.Parse(time.RFC3339, lastSync)
			if discount.UpdatedAt.Before(lastSyncTime) {
				continue
			}
		}
		
		// Create domain event
		event := s.createDiscountEvent(discount)
		domainEvents = append(domainEvents, event)
		processedCount++
		
		// Cache discount data
		cacheKey := fmt.Sprintf("loyverse:discount:%s", discount.ID)
		s.redis.Set(ctx, cacheKey, raw, 24*time.Hour)
	}
	
	// Publish events
	if len(domainEvents) > 0 {
		if err := s.publisher.PublishBatch(ctx, domainEvents); err != nil {
			return fmt.Errorf("publishing events: %w", err)
		}
	}
	
	// Update last sync time
	s.redis.Set(ctx, lastSyncKey, time.Now().Format(time.RFC3339), 0)
	
	log.Printf("Discount sync completed. Processed %d discounts", processedCount)
	return nil
}

func (s *DiscountSync) createDiscountEvent(discount models.Discount) events.DomainEvent {
	eventData := map[string]interface{}{
		"discount_id":        discount.ID,
		"name":               discount.Name,
		"type":               discount.Type,
		"value":              discount.Value,
		"is_restricted":      discount.IsRestricted,
		"minimum_amount":     discount.MinimumAmount,
		"valid_since":        discount.ValidSince,
		"valid_until":        discount.ValidUntil,
		"store_ids":          discount.StoreIDs,
		"source":             "loyverse",
	}
	
	data, _ := json.Marshal(eventData)
	
	return events.DomainEvent{
		ID:            fmt.Sprintf("disc-%s-%d", discount.ID, time.Now().Unix()),
		Type:          events.EventDiscountUpdated,
		AggregateID:   discount.ID,
		AggregateType: "discount",
		Timestamp:     time.Now(),
		Version:       1,
		Data:          data,
		Source:        "loyverse-integration",
	}
}

// integrations/loyverse/internal/sync/category_sync.go
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

type CategorySync struct {
	client      *connector.Client
	publisher   *events.Publisher
	redis       *redis.Client
	transformer *events.Transformer
}

func NewCategorySync(client *connector.Client, publisher *events.Publisher, redis *redis.Client) *CategorySync {
	return &CategorySync{
		client:      client,
		publisher:   publisher,
		redis:       redis,
		transformer: events.NewTransformer(),
	}
}

func (s *CategorySync) Sync(ctx context.Context) error {
	log.Println("Starting category sync...")
	
	// Fetch all categories (usually not too many)
	rawData, err := s.client.GetWithPagination(ctx, "/categories", 250)
	if err != nil {
		return fmt.Errorf("fetching categories: %w", err)
	}
	
	var domainEvents []events.DomainEvent
	processedCount := 0
	
	// Build category tree
	categoryMap := make(map[string]*models.Category)
	var rootCategories []*models.Category
	
	// First pass: unmarshal all categories
	for _, raw := range rawData {
		var category models.Category
		if err := json.Unmarshal(raw, &category); err != nil {
			log.Printf("Error unmarshaling category: %v", err)
			continue
		}
		
		categoryMap[category.ID] = &category
		
		if category.ParentID == nil {
			rootCategories = append(rootCategories, &category)
		}
	}
	
	// Second pass: create events with hierarchy info
	for _, category := range categoryMap {
		event := s.createCategoryEvent(*category, s.getCategoryPath(category, categoryMap))
		domainEvents = append(domainEvents, event)
		processedCount++
		
		// Cache category data
		cacheData, _ := json.Marshal(category)
		cacheKey := fmt.Sprintf("loyverse:category:%s", category.ID)
		s.redis.Set(ctx, cacheKey, cacheData, 24*time.Hour)
	}
	
	// Publish events
	if len(domainEvents) > 0 {
		if err := s.publisher.PublishBatch(ctx, domainEvents); err != nil {
			return fmt.Errorf("publishing events: %w", err)
		}
	}
	
	// Cache category tree
	treeData, _ := json.Marshal(rootCategories)
	s.redis.Set(ctx, "loyverse:categories:tree", treeData, 24*time.Hour)
	
	log.Printf("Category sync completed. Processed %d categories", processedCount)
	return nil
}

func (s *CategorySync) getCategoryPath(category *models.Category, categoryMap map[string]*models.Category) []string {
	path := []string{category.Name}
	current := category
	
	for current.ParentID != nil {
		parent, exists := categoryMap[*current.ParentID]
		if !exists {
			break
		}
		path = append([]string{parent.Name}, path...)
		current = parent
	}
	
	return path
}

func (s *CategorySync) createCategoryEvent(category models.Category, path []string) events.DomainEvent {
	eventData := map[string]interface{}{
		"category_id": category.ID,
		"name":        category.Name,
		"color":       category.Color,
		"parent_id":   category.ParentID,
		"path":        path,
		"source":      "loyverse",
	}
	
	data, _ := json.Marshal(eventData)
	
	return events.DomainEvent{
		ID:            fmt.Sprintf("cat-%s-%d", category.ID, time.Now().Unix()),
		Type:          events.EventCategoryUpdated,
		AggregateID:   category.ID,
		AggregateType: "category",
		Timestamp:     time.Now(),
		Version:       1,
		Data:          data,
		Source:        "loyverse-integration",
	}
}

// integrations/loyverse/internal/sync/store_sync.go
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

type StoreSync struct {
	client      *connector.Client
	publisher   *events.Publisher
	redis       *redis.Client
	transformer *events.Transformer
}

func NewStoreSync(client *connector.Client, publisher *events.Publisher, redis *redis.Client) *StoreSync {
	return &StoreSync{
		client:      client,
		publisher:   publisher,
		redis:       redis,
		transformer: events.NewTransformer(),
	}
}

func (s *StoreSync) Sync(ctx context.Context) error {
	log.Println("Starting store sync...")
	
	// Fetch stores
	rawData, err := s.client.GetWithPagination(ctx, "/stores", 250)
	if err != nil {
		return fmt.Errorf("fetching stores: %w", err)
	}
	
	var domainEvents []events.DomainEvent
	processedCount := 0
	activeStores := make([]string, 0)
	
	for _, raw := range rawData {
		var store models.Store
		if err := json.Unmarshal(raw, &store); err != nil {
			log.Printf("Error unmarshaling store: %v", err)
			continue
		}
		
		// Track active stores
		if store.DeletedAt == nil {
			activeStores = append(activeStores, store.ID)
		}
		
		// Create domain event
		event := s.createStoreEvent(store)
		domainEvents = append(domainEvents, event)
		processedCount++
		
		// Cache store data
		cacheKey := fmt.Sprintf("loyverse:store:%s", store.ID)
		s.redis.Set(ctx, cacheKey, raw, 24*time.Hour)
	}
	
	// Cache active stores list
	activeStoresData, _ := json.Marshal(activeStores)
	s.redis.Set(ctx, "loyverse:stores:active", activeStoresData, 24*time.Hour)
	
	// Publish events
	if len(domainEvents) > 0 {
		if err := s.publisher.PublishBatch(ctx, domainEvents); err != nil {
			return fmt.Errorf("publishing events: %w", err)
		}
	}
	
	log.Printf("Store sync completed. Processed %d stores (%d active)", processedCount, len(activeStores))
	return nil
}

func (s *StoreSync) createStoreEvent(store models.Store) events.DomainEvent {
	eventData := map[string]interface{}{
		"store_id":     store.ID,
		"name":         store.Name,
		"address":      store.Address,
		"phone":        store.Phone,
		"email":        store.Email,
		"tax_number":   store.TaxNumber,
		"is_active":    store.DeletedAt == nil,
		"source":       "loyverse",
	}
	
	data, _ := json.Marshal(eventData)
	
	return events.DomainEvent{
		ID:            fmt.Sprintf("store-%s-%d", store.ID, time.Now().Unix()),
		Type:          events.EventStoreUpdated,
		AggregateID:   store.ID,
		AggregateType: "store",
		Timestamp:     time.Now(),
		Version:       1,
		Data:          data,
		Source:        "loyverse-integration",
	}
}