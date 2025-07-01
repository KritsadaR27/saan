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

// CustomerSync handles customer synchronization
type CustomerSync struct {
	client      *connector.Client
	publisher   *events.Publisher
	redis       *redis.Client
	transformer *events.Transformer
}

// NewCustomerSync creates a new customer sync instance
func NewCustomerSync(client *connector.Client, publisher *events.Publisher, redis *redis.Client) *CustomerSync {
	return &CustomerSync{
		client:      client,
		publisher:   publisher,
		redis:       redis,
		transformer: events.NewTransformer(),
	}
}

// Sync performs customer synchronization
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
		"customer_id":          customer.ID,
		"name":                 customer.Name,
		"email":                customer.Email,
		"phone":                customer.Phone,
		"address":              customer.Address,
		"city":                 customer.City,
		"postal_code":          customer.PostalCode,
		"country_code":         customer.CountryCode,
		"note":                 customer.Note,
		"total_orders":         customer.TotalOrders,
		"total_spent":          customer.TotalSpent,
		"average_order_value":  customer.AverageOrderValue,
		"points_balance":       customer.PointsBalance,
		"customer_code":        customer.CustomerCode,
		"tax_number":           customer.TaxNumber,
		"first_visit":          customer.FirstVisit,
		"last_visit":           customer.LastVisit,
		"source":               "loyverse",
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
