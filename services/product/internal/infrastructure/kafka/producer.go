package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"product-service/internal/infrastructure/config"

	"github.com/segmentio/kafka-go"
)

// Producer represents a Kafka producer
type Producer struct {
	writers map[string]*kafka.Writer
	config  config.KafkaConfig
}

// NewProducer creates a new Kafka producer
func NewProducer(cfg config.KafkaConfig) (*Producer, error) {
	writers := make(map[string]*kafka.Writer)
	
	// Create writers for each topic
	topics := map[string]string{
		"product.created":    cfg.Topics.ProductCreated,
		"product.updated":    cfg.Topics.ProductUpdated,
		"product.deleted":    cfg.Topics.ProductDeleted,
		"product.synced":     cfg.Topics.ProductSynced,
		"price.changed":      cfg.Topics.PriceChanged,
		"stock.changed":      cfg.Topics.StockChanged,
		"inventory.low":      cfg.Topics.InventoryLow,
		"inventory.alert":    cfg.Topics.InventoryAlert,
	}

	for key, topic := range topics {
		writer := &kafka.Writer{
			Addr:         kafka.TCP(cfg.Brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireOne,
			Async:        true,
		}
		writers[key] = writer
	}

	return &Producer{
		writers: writers,
		config:  cfg,
	}, nil
}

// PublishProductCreated publishes a product created event
func (p *Producer) PublishProductCreated(ctx context.Context, event ProductCreatedEvent) error {
	return p.publishEvent(ctx, "product.created", event)
}

// PublishProductUpdated publishes a product updated event
func (p *Producer) PublishProductUpdated(ctx context.Context, event ProductUpdatedEvent) error {
	return p.publishEvent(ctx, "product.updated", event)
}

// PublishProductDeleted publishes a product deleted event
func (p *Producer) PublishProductDeleted(ctx context.Context, event ProductDeletedEvent) error {
	return p.publishEvent(ctx, "product.deleted", event)
}

// PublishProductSynced publishes a product synced event
func (p *Producer) PublishProductSynced(ctx context.Context, event ProductSyncedEvent) error {
	return p.publishEvent(ctx, "product.synced", event)
}

// PublishPriceChanged publishes a price changed event
func (p *Producer) PublishPriceChanged(ctx context.Context, event PriceChangedEvent) error {
	return p.publishEvent(ctx, "price.changed", event)
}

// PublishStockChanged publishes a stock changed event
func (p *Producer) PublishStockChanged(ctx context.Context, event StockChangedEvent) error {
	return p.publishEvent(ctx, "stock.changed", event)
}

// PublishInventoryLow publishes an inventory low event
func (p *Producer) PublishInventoryLow(ctx context.Context, event InventoryLowEvent) error {
	return p.publishEvent(ctx, "inventory.low", event)
}

// PublishInventoryAlert publishes an inventory alert event
func (p *Producer) PublishInventoryAlert(ctx context.Context, event InventoryAlertEvent) error {
	return p.publishEvent(ctx, "inventory.alert", event)
}

// publishEvent publishes an event to the specified topic
func (p *Producer) publishEvent(ctx context.Context, topic string, event interface{}) error {
	writer, exists := p.writers[topic]
	if !exists {
		return fmt.Errorf("writer not found for topic: %s", topic)
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(fmt.Sprintf("%v", getEventKey(event))),
		Value: data,
		Time:  time.Now(),
	}

	return writer.WriteMessages(ctx, message)
}

// getEventKey extracts the key from an event for partitioning
func getEventKey(event interface{}) string {
	switch e := event.(type) {
	case ProductCreatedEvent:
		return e.ProductID
	case ProductUpdatedEvent:
		return e.ProductID
	case ProductDeletedEvent:
		return e.ProductID
	case ProductSyncedEvent:
		return e.ProductID
	case PriceChangedEvent:
		return e.ProductID
	case StockChangedEvent:
		return e.ProductID
	case InventoryLowEvent:
		return e.ProductID
	case InventoryAlertEvent:
		return e.ProductID
	default:
		return "unknown"
	}
}

// Close closes all writers
func (p *Producer) Close() error {
	for _, writer := range p.writers {
		if err := writer.Close(); err != nil {
			log.Printf("Error closing writer: %v", err)
		}
	}
	return nil
}

// Event types
type ProductCreatedEvent struct {
	ProductID    string    `json:"product_id"`
	Name         string    `json:"name"`
	SKU          string    `json:"sku"`
	CategoryID   *string   `json:"category_id"`
	BasePrice    float64   `json:"base_price"`
	IsActive     bool      `json:"is_active"`
	IsVIPOnly    bool      `json:"is_vip_only"`
	DataSource   string    `json:"data_source"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    *string   `json:"created_by"`
}

type ProductUpdatedEvent struct {
	ProductID    string    `json:"product_id"`
	Name         string    `json:"name"`
	SKU          string    `json:"sku"`
	CategoryID   *string   `json:"category_id"`
	BasePrice    float64   `json:"base_price"`
	IsActive     bool      `json:"is_active"`
	IsVIPOnly    bool      `json:"is_vip_only"`
	DataSource   string    `json:"data_source"`
	UpdatedAt    time.Time `json:"updated_at"`
	UpdatedBy    *string   `json:"updated_by"`
	Version      int       `json:"version"`
	Changes      []string  `json:"changes"`
}

type ProductDeletedEvent struct {
	ProductID  string    `json:"product_id"`
	Name       string    `json:"name"`
	SKU        string    `json:"sku"`
	DeletedAt  time.Time `json:"deleted_at"`
	DeletedBy  *string   `json:"deleted_by"`
	Reason     *string   `json:"reason"`
}

type ProductSyncedEvent struct {
	ProductID      string    `json:"product_id"`
	DataSource     string    `json:"data_source"`
	DataSourceID   string    `json:"data_source_id"`
	SyncedAt       time.Time `json:"synced_at"`
	SyncStatus     string    `json:"sync_status"`
	Changes        []string  `json:"changes"`
	ErrorMessage   *string   `json:"error_message"`
}

type PriceChangedEvent struct {
	ProductID    string    `json:"product_id"`
	PriceType    string    `json:"price_type"`
	OldPrice     float64   `json:"old_price"`
	NewPrice     float64   `json:"new_price"`
	Currency     string    `json:"currency"`
	ValidFrom    *time.Time `json:"valid_from"`
	ValidTo      *time.Time `json:"valid_to"`
	ChangedAt    time.Time `json:"changed_at"`
	ChangedBy    *string   `json:"changed_by"`
	Reason       *string   `json:"reason"`
}

type StockChangedEvent struct {
	ProductID      string    `json:"product_id"`
	LocationID     string    `json:"location_id"`
	OldStock       float64   `json:"old_stock"`
	NewStock       float64   `json:"new_stock"`
	Delta          float64   `json:"delta"`
	ChangeType     string    `json:"change_type"` // "adjustment", "sale", "purchase", "transfer"
	ChangedAt      time.Time `json:"changed_at"`
	ChangedBy      *string   `json:"changed_by"`
	Reason         *string   `json:"reason"`
	ReferenceID    *string   `json:"reference_id"`
}

type InventoryLowEvent struct {
	ProductID         string    `json:"product_id"`
	LocationID        string    `json:"location_id"`
	ProductName       string    `json:"product_name"`
	SKU               string    `json:"sku"`
	CurrentStock      float64   `json:"current_stock"`
	LowStockThreshold float64   `json:"low_stock_threshold"`
	ReorderPoint      *float64  `json:"reorder_point"`
	DetectedAt        time.Time `json:"detected_at"`
	IsUrgent          bool      `json:"is_urgent"`
}

type InventoryAlertEvent struct {
	ProductID    string    `json:"product_id"`
	LocationID   string    `json:"location_id"`
	ProductName  string    `json:"product_name"`
	SKU          string    `json:"sku"`
	AlertType    string    `json:"alert_type"` // "low_stock", "out_of_stock", "overstock"
	CurrentStock float64   `json:"current_stock"`
	Message      string    `json:"message"`
	Severity     string    `json:"severity"` // "info", "warning", "critical"
	DetectedAt   time.Time `json:"detected_at"`
}
