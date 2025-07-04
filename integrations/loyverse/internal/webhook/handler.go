package webhook
// integrations/loyverse/internal/webhook/handler.go
package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"integrations/loyverse/internal/events"
)

// Handler handles Loyverse webhooks
type Handler struct {
	secret      string
	processor   *Processor
	publisher   *events.Publisher
	transformer *events.Transformer
}

// NewHandler creates a new webhook handler
func NewHandler(secret string, processor *Processor, publisher *events.Publisher) *Handler {
	return &Handler{
		secret:      secret,
		processor:   processor,
		publisher:   publisher,
		transformer: events.NewTransformer(),
	}
}

// WebhookPayload represents Loyverse webhook payload
type WebhookPayload struct {
	Type      string          `json:"type"`
	CreatedAt time.Time       `json:"created_at"`
	Data      json.RawMessage `json:"data"`
}

// ServeHTTP implements http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading webhook body: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Verify signature if secret is configured
	if h.secret != "" {
		signature := r.Header.Get("X-Loyverse-Signature")
		if !h.verifySignature(body, signature) {
			log.Println("Invalid webhook signature")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// Parse payload
	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("Error parsing webhook payload: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Process webhook asynchronously
	go func() {
		ctx := context.Background()
		if err := h.processWebhook(ctx, payload); err != nil {
			log.Printf("Error processing webhook: %v", err)
		}
	}()

	// Return success immediately
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// verifySignature verifies the webhook signature
func (h *Handler) verifySignature(body []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(h.secret))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// processWebhook processes the webhook payload
func (h *Handler) processWebhook(ctx context.Context, payload WebhookPayload) error {
	log.Printf("Processing webhook: type=%s", payload.Type)

	switch payload.Type {
	case "receipts.update":
		return h.processReceiptUpdate(ctx, payload.Data)
	case "inventory_levels.update":
		return h.processInventoryUpdate(ctx, payload.Data)
	case "items.update":
		return h.processItemUpdate(ctx, payload.Data)
	case "customers.update":
		return h.processCustomerUpdate(ctx, payload.Data)
	case "categories.update":
		return h.processCategoryUpdate(ctx, payload.Data)
	default:
		log.Printf("Unknown webhook type: %s", payload.Type)
		return nil
	}
}

// processReceiptUpdate processes receipt update webhook
func (h *Handler) processReceiptUpdate(ctx context.Context, data json.RawMessage) error {
	var receipts []json.RawMessage
	if err := json.Unmarshal(data, &receipts); err != nil {
		// Try single receipt
		receipts = []json.RawMessage{data}
	}

	var domainEvents []events.DomainEvent
	for _, receipt := range receipts {
		// Store the latest receipt in Redis
		if err := h.processor.StoreLatestReceipt(ctx, receipt); err != nil {
			log.Printf("Error storing latest receipt: %v", err)
		}
		
		event, err := h.transformer.TransformReceipt(receipt)
		if err != nil {
			log.Printf("Error transforming receipt: %v", err)
			continue
		}
		domainEvents = append(domainEvents, event)
	}

	if len(domainEvents) > 0 {
		return h.publisher.PublishBatch(ctx, domainEvents)
	}
	return nil
}

// processInventoryUpdate processes inventory update webhook
func (h *Handler) processInventoryUpdate(ctx context.Context, data json.RawMessage) error {
	var inventoryLevels []json.RawMessage
	if err := json.Unmarshal(data, &inventoryLevels); err != nil {
		// Try single inventory level
		inventoryLevels = []json.RawMessage{data}
	}

	var domainEvents []events.DomainEvent
	for _, inventory := range inventoryLevels {
		event, err := h.transformer.TransformInventory(inventory)
		if err != nil {
			log.Printf("Error transforming inventory: %v", err)
			continue
		}
		domainEvents = append(domainEvents, event)
	}

	if len(domainEvents) > 0 {
		return h.publisher.PublishBatch(ctx, domainEvents)
	}
	return nil
}

// processItemUpdate processes item update webhook
func (h *Handler) processItemUpdate(ctx context.Context, data json.RawMessage) error {
	var items []json.RawMessage
	if err := json.Unmarshal(data, &items); err != nil {
		// Try single item
		items = []json.RawMessage{data}
	}

	var domainEvents []events.DomainEvent
	for _, item := range items {
		event, err := h.transformer.TransformProduct(item)
		if err != nil {
			log.Printf("Error transforming item: %v", err)
			continue
		}
		domainEvents = append(domainEvents, event)
	}

	if len(domainEvents) > 0 {
		return h.publisher.PublishBatch(ctx, domainEvents)
	}
	return nil
}

// processCustomerUpdate processes customer update webhook
func (h *Handler) processCustomerUpdate(ctx context.Context, data json.RawMessage) error {
	// Similar implementation for customer updates
	log.Printf("Processing customer update")
	// TODO: Implement customer transformation
	return nil
}

// processCategoryUpdate processes category update webhook
func (h *Handler) processCategoryUpdate(ctx context.Context, data json.RawMessage) error {
	var categories []json.RawMessage
	if err := json.Unmarshal(data, &categories); err != nil {
		// Try single category
		categories = []json.RawMessage{data}
	}

	// Store categories in Redis
	if err := h.processor.StoreCategories(ctx, categories); err != nil {
		log.Printf("Error storing categories: %v", err)
	}

	var domainEvents []events.DomainEvent
	for _, category := range categories {
		event, err := h.transformer.TransformCategory(category)
		if err != nil {
			log.Printf("Error transforming category: %v", err)
			continue
		}
		domainEvents = append(domainEvents, event)
	}

	if len(domainEvents) > 0 {
		return h.publisher.PublishBatch(ctx, domainEvents)
	}
	return nil
}