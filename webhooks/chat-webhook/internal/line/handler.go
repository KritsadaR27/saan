// webhooks/chat-webhook/internal/line/handler.go
package line

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
)

// Handler handles LINE webhook events
type Handler struct {
	channelSecret      string
	channelAccessToken string
	kafkaWriter        *kafka.Writer
	redisClient        *redis.Client
}

// NewHandler creates a new LINE webhook handler
func NewHandler(channelSecret, channelAccessToken string, kafkaWriter *kafka.Writer, redisClient *redis.Client) *Handler {
	return &Handler{
		channelSecret:      channelSecret,
		channelAccessToken: channelAccessToken,
		kafkaWriter:        kafkaWriter,
		redisClient:        redisClient,
	}
}

// WebhookEvent represents LINE webhook event structure
type WebhookEvent struct {
	ReplyToken string      `json:"replyToken,omitempty"`
	Type       string      `json:"type"`
	Mode       string      `json:"mode,omitempty"`
	Timestamp  int64       `json:"timestamp"`
	Source     EventSource `json:"source"`
	Message    *Message    `json:"message,omitempty"`
	Postback   *Postback   `json:"postback,omitempty"`
}

// EventSource represents the source of LINE event
type EventSource struct {
	Type    string `json:"type"`
	UserID  string `json:"userId,omitempty"`
	GroupID string `json:"groupId,omitempty"`
	RoomID  string `json:"roomId,omitempty"`
}

// Message represents LINE message
type Message struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// Postback represents LINE postback
type Postback struct {
	Data   string `json:"data"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// LineWebhook represents the full LINE webhook payload
type LineWebhook struct {
	Destination string         `json:"destination"`
	Events      []WebhookEvent `json:"events"`
}

// HandleWebhook handles LINE webhook requests
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading LINE webhook body: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Verify signature
	signature := r.Header.Get("X-Line-Signature")
	if !h.verifySignature(body, signature) {
		log.Printf("Invalid LINE webhook signature")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse webhook payload
	var webhook LineWebhook
	if err := json.Unmarshal(body, &webhook); err != nil {
		log.Printf("Error parsing LINE webhook: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Process webhook asynchronously
	go func() {
		ctx := context.Background()
		if err := h.processWebhook(ctx, &webhook); err != nil {
			log.Printf("Error processing LINE webhook: %v", err)
		}
	}()

	// Return success immediately
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// verifySignature verifies LINE webhook signature
func (h *Handler) verifySignature(body []byte, signature string) bool {
	if h.channelSecret == "" {
		log.Println("Warning: No LINE channel secret configured, skipping signature verification")
		return true
	}

	mac := hmac.New(sha256.New, []byte(h.channelSecret))
	mac.Write(body)
	expectedSignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// processWebhook processes the LINE webhook
func (h *Handler) processWebhook(ctx context.Context, webhook *LineWebhook) error {
	log.Printf("Processing LINE webhook with %d events for destination %s", len(webhook.Events), webhook.Destination)

	for _, event := range webhook.Events {
		if err := h.processEvent(ctx, &event, webhook.Destination); err != nil {
			log.Printf("Error processing LINE event: %v", err)
			continue
		}
	}

	return nil
}

// processEvent processes a single LINE event
func (h *Handler) processEvent(ctx context.Context, event *WebhookEvent, destination string) error {
	// Cache the event for debugging
	if err := h.cacheEvent(ctx, event, destination); err != nil {
		log.Printf("Warning: Failed to cache LINE event: %v", err)
	}

	// Create domain event
	domainEvent := h.createChatEvent(event, destination)

	// Publish to Kafka
	return h.publishEvent(ctx, domainEvent)
}

// cacheEvent stores the event in Redis for debugging
func (h *Handler) cacheEvent(ctx context.Context, event *WebhookEvent, destination string) error {
	var userID string
	if event.Source.UserID != "" {
		userID = event.Source.UserID
	} else if event.Source.GroupID != "" {
		userID = event.Source.GroupID
	} else if event.Source.RoomID != "" {
		userID = event.Source.RoomID
	}

	cacheKey := fmt.Sprintf("line:event:%s:%d", userID, event.Timestamp)

	eventData := map[string]interface{}{
		"type":        event.Type,
		"timestamp":   event.Timestamp,
		"source":      event.Source,
		"destination": destination,
		"message":     event.Message,
		"postback":    event.Postback,
		"reply_token": event.ReplyToken,
		"cached_at":   time.Now(),
	}

	data, err := json.Marshal(eventData)
	if err != nil {
		return err
	}

	// Cache for 7 days
	return h.redisClient.Set(ctx, cacheKey, data, 7*24*time.Hour).Err()
}

// createChatEvent creates a domain event for LINE events
func (h *Handler) createChatEvent(event *WebhookEvent, destination string) map[string]interface{} {
	eventType := fmt.Sprintf("line.%s", event.Type)
	
	var aggregateID string
	if event.Source.UserID != "" {
		aggregateID = event.Source.UserID
	} else if event.Source.GroupID != "" {
		aggregateID = event.Source.GroupID
	} else if event.Source.RoomID != "" {
		aggregateID = event.Source.RoomID
	} else {
		aggregateID = fmt.Sprintf("line-unknown-%d", event.Timestamp)
	}

	return map[string]interface{}{
		"id":            fmt.Sprintf("line-%s-%d", aggregateID, event.Timestamp),
		"type":          eventType,
		"aggregate_id":  aggregateID,
		"aggregate_type": fmt.Sprintf("line_%s", event.Source.Type),
		"timestamp":     time.Unix(event.Timestamp/1000, 0),
		"version":       1,
		"data": map[string]interface{}{
			"platform":     "line",
			"destination":  destination,
			"event_type":   event.Type,
			"source":       event.Source,
			"message":      event.Message,
			"postback":     event.Postback,
			"reply_token":  event.ReplyToken,
		},
		"source": "line-webhook",
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
			{Key: "platform", Value: []byte("line")},
			{Key: "source", Value: []byte("line-webhook")},
		},
		Time: time.Now(),
	}

	return h.kafkaWriter.WriteMessages(ctx, message)
}
