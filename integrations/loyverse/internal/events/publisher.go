// integrations/loyverse/internal/events/publisher.go
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// EventType represents different types of domain events
type EventType string

const (
	EventProductUpdated    EventType = "product.updated"
	EventInventoryUpdated  EventType = "inventory.updated"
	EventReceiptCreated    EventType = "receipt.created"
	EventCustomerUpdated   EventType = "customer.updated"
	EventEmployeeUpdated   EventType = "employee.updated"
	EventCategoryUpdated   EventType = "category.updated"
	EventSupplierUpdated   EventType = "supplier.updated"
	EventDiscountUpdated   EventType = "discount.updated"
	EventPaymentTypeUpdated EventType = "payment_type.updated"
	EventStoreUpdated      EventType = "store.updated"
)

// DomainEvent represents a domain event
type DomainEvent struct {
	ID            string          `json:"id"`
	Type          EventType       `json:"type"`
	AggregateID   string          `json:"aggregate_id"`
	AggregateType string          `json:"aggregate_type"`
	Timestamp     time.Time       `json:"timestamp"`
	Version       int             `json:"version"`
	Data          json.RawMessage `json:"data"`
	Source        string          `json:"source"`
}

// Publisher publishes events to Kafka
type Publisher struct {
	writer *kafka.Writer
}

// NewPublisher creates a new event publisher
func NewPublisher(brokers []string, topic string) *Publisher {
	return &Publisher{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
			Compression: kafka.Snappy,
		},
	}
}

// Publish publishes an event to Kafka
func (p *Publisher) Publish(ctx context.Context, event DomainEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshaling event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(event.AggregateID),
		Value: data,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(event.Type)},
			{Key: "source", Value: []byte("loyverse")},
		},
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("writing message: %w", err)
	}

	return nil
}

// PublishBatch publishes multiple events
func (p *Publisher) PublishBatch(ctx context.Context, events []DomainEvent) error {
	messages := make([]kafka.Message, 0, len(events))
	
	for _, event := range events {
		data, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("marshaling event: %w", err)
		}

		messages = append(messages, kafka.Message{
			Key:   []byte(event.AggregateID),
			Value: data,
			Headers: []kafka.Header{
				{Key: "event_type", Value: []byte(event.Type)},
				{Key: "source", Value: []byte("loyverse")},
			},
		})
	}

	if err := p.writer.WriteMessages(ctx, messages...); err != nil {
		return fmt.Errorf("writing batch: %w", err)
	}

	return nil
}

// Close closes the publisher
func (p *Publisher) Close() error {
	return p.writer.Close()
}

// Transformer transforms Loyverse data to domain events
type Transformer struct {
	source string
}

// NewTransformer creates a new event transformer
func NewTransformer() *Transformer {
	return &Transformer{
		source: "loyverse-integration",
	}
}

// TransformProduct transforms Loyverse product to domain event
func (t *Transformer) TransformProduct(loyverseProduct json.RawMessage) (DomainEvent, error) {
	var product struct {
		ID          string `json:"id"`
		ItemName    string `json:"item_name"`
		Description string `json:"description"`
		CategoryID  string `json:"category_id"`
		UpdatedAt   string `json:"updated_at"`
	}

	if err := json.Unmarshal(loyverseProduct, &product); err != nil {
		return DomainEvent{}, fmt.Errorf("unmarshaling product: %w", err)
	}

	eventData := map[string]interface{}{
		"product_id":  product.ID,
		"name":        product.ItemName,
		"description": product.Description,
		"category_id": product.CategoryID,
		"source":      "loyverse",
	}

	data, err := json.Marshal(eventData)
	if err != nil {
		return DomainEvent{}, fmt.Errorf("marshaling event data: %w", err)
	}

	return DomainEvent{
		ID:            fmt.Sprintf("loyverse_product_%s_%d", product.ID, time.Now().Unix()),
		Type:          EventProductUpdated,
		AggregateID:   product.ID,
		AggregateType: "product",
		Timestamp:     time.Now(),
		Version:       1,
		Data:          data,
		Source:        t.source,
	}, nil
}

// TransformInventory transforms Loyverse inventory level to domain event
func (t *Transformer) TransformInventory(loyverseInventory json.RawMessage) (DomainEvent, error) {
	var inventory struct {
		VariantID string  `json:"variant_id"`
		StoreID   string  `json:"store_id"`
		InStock   float64 `json:"in_stock"`
		UpdatedAt string  `json:"updated_at"`
	}

	if err := json.Unmarshal(loyverseInventory, &inventory); err != nil {
		return DomainEvent{}, fmt.Errorf("unmarshaling inventory: %w", err)
	}

	eventData := map[string]interface{}{
		"variant_id": inventory.VariantID,
		"store_id":   inventory.StoreID,
		"quantity":   inventory.InStock,
		"source":     "loyverse",
	}

	data, err := json.Marshal(eventData)
	if err != nil {
		return DomainEvent{}, fmt.Errorf("marshaling event data: %w", err)
	}

	return DomainEvent{
		ID:            fmt.Sprintf("loyverse_inventory_%s_%d", inventory.VariantID, time.Now().Unix()),
		Type:          EventInventoryUpdated,
		AggregateID:   inventory.VariantID,
		AggregateType: "inventory",
		Timestamp:     time.Now(),
		Version:       1,
		Data:          data,
		Source:        t.source,
	}, nil
}

// TransformReceipt transforms Loyverse receipt to domain event
func (t *Transformer) TransformReceipt(loyverseReceipt json.RawMessage) (DomainEvent, error) {
	var receipt struct {
		ReceiptNumber string    `json:"receipt_number"`
		CustomerID    string    `json:"customer_id"`
		TotalMoney    float64   `json:"total_money"`
		CreatedAt     time.Time `json:"created_at"`
	}

	if err := json.Unmarshal(loyverseReceipt, &receipt); err != nil {
		return DomainEvent{}, fmt.Errorf("unmarshaling receipt: %w", err)
	}

	eventData := map[string]interface{}{
		"receipt_number": receipt.ReceiptNumber,
		"customer_id":    receipt.CustomerID,
		"total_amount":   receipt.TotalMoney,
		"created_at":     receipt.CreatedAt,
		"source":         "loyverse",
	}

	data, err := json.Marshal(eventData)
	if err != nil {
		return DomainEvent{}, fmt.Errorf("marshaling event data: %w", err)
	}

	return DomainEvent{
		ID:            fmt.Sprintf("loyverse_receipt_%s_%d", receipt.ReceiptNumber, time.Now().Unix()),
		Type:          EventReceiptCreated,
		AggregateID:   receipt.ReceiptNumber,
		AggregateType: "receipt",
		Timestamp:     time.Now(),
		Version:       1,
		Data:          data,
		Source:        t.source,
	}, nil
}

// TransformCategory transforms Loyverse category to domain event
func (t *Transformer) TransformCategory(loyverseCategory json.RawMessage) (DomainEvent, error) {
	var category struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		Color     string    `json:"color"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	if err := json.Unmarshal(loyverseCategory, &category); err != nil {
		return DomainEvent{}, fmt.Errorf("unmarshaling category: %w", err)
	}

	eventData := map[string]interface{}{
		"category_id":   category.ID,
		"name":          category.Name,
		"color":         category.Color,
		"created_at":    category.CreatedAt,
		"updated_at":    category.UpdatedAt,
		"source":        "loyverse",
	}

	data, err := json.Marshal(eventData)
	if err != nil {
		return DomainEvent{}, fmt.Errorf("marshaling event data: %w", err)
	}

	return DomainEvent{
		ID:            fmt.Sprintf("loyverse_category_%s_%d", category.ID, time.Now().Unix()),
		Type:          "category.updated",
		AggregateID:   category.ID,
		AggregateType: "category",
		Timestamp:     time.Now(),
		Version:       1,
		Data:          data,
		Source:        t.source,
	}, nil
}