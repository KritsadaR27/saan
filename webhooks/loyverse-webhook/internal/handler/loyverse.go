// webhooks/loyverse-webhook/internal/handler/loyverse.go
package handler

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

	"github.com/segmentio/kafka-go"
	"webhooks/loyverse-webhook/internal/processor"
)

// Handler handles Loyverse webhooks
type Handler struct {
	secret      string
	processor   *processor.Processor
	kafkaWriter *kafka.Writer
}

// NewHandler creates a new webhook handler
func NewHandler(secret string, processor *processor.Processor, kafkaWriter *kafka.Writer) *Handler {
	return &Handler{
		secret:      secret,
		processor:   processor,
		kafkaWriter: kafkaWriter,
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

	// Generate event ID for deduplication
	eventID := h.generateEventID(payload)
	
	// Process with deduplication
	if err := h.processor.ProcessEvent(ctx, eventID, payload.Type, payload.Data); err != nil {
		return err
	}

	// Publish to Kafka
	return h.publishToKafka(ctx, payload)
}

// generateEventID generates a unique event ID for deduplication
func (h *Handler) generateEventID(payload WebhookPayload) string {
	hasher := sha256.New()
	hasher.Write([]byte(payload.Type))
	hasher.Write([]byte(payload.CreatedAt.Format(time.RFC3339)))
	hasher.Write(payload.Data)
	return hex.EncodeToString(hasher.Sum(nil))
}

// publishToKafka publishes the webhook event to Kafka
func (h *Handler) publishToKafka(ctx context.Context, payload WebhookPayload) error {
	message := kafka.Message{
		Key:   []byte(payload.Type),
		Value: payload.Data,
		Headers: []kafka.Header{
			{Key: "source", Value: []byte("loyverse")},
			{Key: "type", Value: []byte(payload.Type)},
			{Key: "timestamp", Value: []byte(payload.CreatedAt.Format(time.RFC3339))},
		},
	}

	return h.kafkaWriter.WriteMessages(ctx, message)
}
