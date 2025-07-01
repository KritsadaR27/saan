// webhooks/chat-webhook/internal/facebook/handler.go
package facebook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
)

// Handler handles Facebook Messenger webhooks
type Handler struct {
	verifyToken     string
	pageAccessToken string
	appSecret       string
	kafkaWriter     *kafka.Writer
	redisClient     *redis.Client
}

// NewHandler creates a new Facebook webhook handler
func NewHandler(verifyToken, pageAccessToken, appSecret string, kafkaWriter *kafka.Writer, redisClient *redis.Client) *Handler {
	return &Handler{
		verifyToken:     verifyToken,
		pageAccessToken: pageAccessToken,
		appSecret:       appSecret,
		kafkaWriter:     kafkaWriter,
		redisClient:     redisClient,
	}
}

// WebhookEntry represents Facebook webhook entry
type WebhookEntry struct {
	ID      string      `json:"id"`
	Time    int64       `json:"time"`
	Changes []Change    `json:"changes,omitempty"`
	Messaging []Messaging `json:"messaging,omitempty"`
}

// Change represents Facebook webhook change
type Change struct {
	Value interface{} `json:"value"`
	Field string      `json:"field"`
}

// Messaging represents Facebook messaging event
type Messaging struct {
	Sender    User    `json:"sender"`
	Recipient User    `json:"recipient"`
	Timestamp int64   `json:"timestamp"`
	Message   *Message `json:"message,omitempty"`
	Postback  *Postback `json:"postback,omitempty"`
}

// User represents Facebook user
type User struct {
	ID string `json:"id"`
}

// Message represents Facebook message
type Message struct {
	MID  string `json:"mid"`
	Text string `json:"text"`
}

// Postback represents Facebook postback
type Postback struct {
	Title   string `json:"title"`
	Payload string `json:"payload"`
}

// FacebookWebhook represents the full webhook payload
type FacebookWebhook struct {
	Object string         `json:"object"`
	Entry  []WebhookEntry `json:"entry"`
}

// HandleWebhook handles Facebook webhook verification and message processing
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleVerification(w, r)
	case http.MethodPost:
		h.handleMessage(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleVerification handles Facebook webhook verification
func (h *Handler) handleVerification(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("hub.mode")
	token := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")

	if mode == "subscribe" && token == h.verifyToken {
		log.Println("Facebook webhook verified successfully")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(challenge))
	} else {
		log.Printf("Failed to verify Facebook webhook: mode=%s, token=%s", mode, token)
		http.Error(w, "Verification failed", http.StatusForbidden)
	}
}

// handleMessage handles incoming Facebook messages
func (h *Handler) handleMessage(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading Facebook webhook body: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Verify signature
	signature := r.Header.Get("X-Hub-Signature-256")
	if !h.verifySignature(body, signature) {
		log.Printf("Invalid Facebook webhook signature")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse webhook payload
	var webhook FacebookWebhook
	if err := json.Unmarshal(body, &webhook); err != nil {
		log.Printf("Error parsing Facebook webhook: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Process webhook asynchronously
	go func() {
		ctx := context.Background()
		if err := h.processWebhook(ctx, &webhook); err != nil {
			log.Printf("Error processing Facebook webhook: %v", err)
		}
	}()

	// Return success immediately
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// verifySignature verifies Facebook webhook signature
func (h *Handler) verifySignature(body []byte, signature string) bool {
	if h.appSecret == "" {
		log.Println("Warning: No Facebook app secret configured, skipping signature verification")
		return true
	}

	// Remove "sha256=" prefix if present
	if len(signature) > 7 && signature[:7] == "sha256=" {
		signature = signature[7:]
	}

	mac := hmac.New(sha256.New, []byte(h.appSecret))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// processWebhook processes the Facebook webhook
func (h *Handler) processWebhook(ctx context.Context, webhook *FacebookWebhook) error {
	log.Printf("Processing Facebook webhook with %d entries", len(webhook.Entry))

	for _, entry := range webhook.Entry {
		// Process messaging events
		for _, messaging := range entry.Messaging {
			if err := h.processMessaging(ctx, &messaging, entry.ID); err != nil {
				log.Printf("Error processing messaging event: %v", err)
				continue
			}
		}

		// Process page changes
		for _, change := range entry.Changes {
			if err := h.processChange(ctx, &change, entry.ID); err != nil {
				log.Printf("Error processing change event: %v", err)
				continue
			}
		}
	}

	return nil
}

// processMessaging processes a messaging event
func (h *Handler) processMessaging(ctx context.Context, messaging *Messaging, pageID string) error {
	// Cache the message for debugging
	if err := h.cacheMessage(ctx, messaging, pageID); err != nil {
		log.Printf("Warning: Failed to cache Facebook message: %v", err)
	}

	// Create domain event
	event := h.createChatEvent(messaging, pageID)

	// Publish to Kafka
	return h.publishEvent(ctx, event)
}

// processChange processes a page change event
func (h *Handler) processChange(ctx context.Context, change *Change, pageID string) error {
	// For now, just log the change
	log.Printf("Facebook page change: field=%s, page=%s", change.Field, pageID)

	// Create domain event for page changes
	event := h.createPageEvent(change, pageID)

	// Publish to Kafka
	return h.publishEvent(ctx, event)
}

// cacheMessage stores the message in Redis for debugging
func (h *Handler) cacheMessage(ctx context.Context, messaging *Messaging, pageID string) error {
	cacheKey := fmt.Sprintf("facebook:message:%s:%d", messaging.Sender.ID, messaging.Timestamp)

	messageData := map[string]interface{}{
		"sender_id":    messaging.Sender.ID,
		"recipient_id": messaging.Recipient.ID,
		"timestamp":    messaging.Timestamp,
		"page_id":      pageID,
		"message":      messaging.Message,
		"postback":     messaging.Postback,
		"cached_at":    time.Now(),
	}

	data, err := json.Marshal(messageData)
	if err != nil {
		return err
	}

	// Cache for 7 days
	return h.redisClient.Set(ctx, cacheKey, data, 7*24*time.Hour).Err()
}

// createChatEvent creates a domain event for chat messages
func (h *Handler) createChatEvent(messaging *Messaging, pageID string) map[string]interface{} {
	eventType := "facebook.message.received"
	if messaging.Postback != nil {
		eventType = "facebook.postback.received"
	}

	return map[string]interface{}{
		"id":            fmt.Sprintf("fb-%s-%d", messaging.Sender.ID, messaging.Timestamp),
		"type":          eventType,
		"aggregate_id":  messaging.Sender.ID,
		"aggregate_type": "facebook_user",
		"timestamp":     time.Unix(messaging.Timestamp/1000, 0),
		"version":       1,
		"data": map[string]interface{}{
			"platform":     "facebook",
			"page_id":      pageID,
			"sender_id":    messaging.Sender.ID,
			"recipient_id": messaging.Recipient.ID,
			"message":      messaging.Message,
			"postback":     messaging.Postback,
		},
		"source": "facebook-webhook",
	}
}

// createPageEvent creates a domain event for page changes
func (h *Handler) createPageEvent(change *Change, pageID string) map[string]interface{} {
	return map[string]interface{}{
		"id":            fmt.Sprintf("fb-page-%s-%d", pageID, time.Now().Unix()),
		"type":          fmt.Sprintf("facebook.page.%s", change.Field),
		"aggregate_id":  pageID,
		"aggregate_type": "facebook_page",
		"timestamp":     time.Now(),
		"version":       1,
		"data": map[string]interface{}{
			"platform": "facebook",
			"page_id":  pageID,
			"field":    change.Field,
			"value":    change.Value,
		},
		"source": "facebook-webhook",
	}
}

// publishEvent publishes an event to Kafka
func (h *Handler) publishEvent(ctx context.Context, event map[string]interface{}) error {
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	aggregateID, _ := event["aggregate_id"].(string)
	eventType, _ := event["type"].(string)

	message := kafka.Message{
		Key:   []byte(aggregateID),
		Value: eventData,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte(eventType)},
			{Key: "platform", Value: []byte("facebook")},
			{Key: "source", Value: []byte("facebook-webhook")},
		},
		Time: time.Now(),
	}

	return h.kafkaWriter.WriteMessages(ctx, message)
}
