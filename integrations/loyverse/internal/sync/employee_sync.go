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
	rawData, err := s.client.GetEmployees(ctx)
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