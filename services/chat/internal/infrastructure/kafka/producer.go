package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

// Producer wraps Kafka producer
type Producer struct {
	writer *kafka.Writer
}

// NewProducer creates a new Kafka producer
func NewProducer(brokers []string) (*Producer, error) {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		BatchSize:    100,
	}

	logrus.Info("Kafka producer created successfully")
	return &Producer{writer: writer}, nil
}

// PublishMessage publishes a message to a Kafka topic
func (p *Producer) PublishMessage(ctx context.Context, topic string, key string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: data,
		Time:  time.Now(),
	})
}

// PublishChatMessage publishes a chat message event
func (p *Producer) PublishChatMessage(ctx context.Context, message ChatMessageEvent) error {
	return p.PublishMessage(ctx, "chat-messages", message.MessageID, message)
}

// PublishOrderIntent publishes an order intent event
func (p *Producer) PublishOrderIntent(ctx context.Context, orderIntent OrderIntentEvent) error {
	return p.PublishMessage(ctx, "order-intents", orderIntent.ConversationID, orderIntent)
}

// Close closes the producer
func (p *Producer) Close() error {
	return p.writer.Close()
}

// Event types for Kafka messages
type ChatMessageEvent struct {
	MessageID      string                 `json:"message_id"`
	ConversationID string                 `json:"conversation_id"`
	UserID         string                 `json:"user_id"`
	Platform       string                 `json:"platform"`
	Direction      string                 `json:"direction"`
	Type           string                 `json:"type"`
	Content        string                 `json:"content"`
	MediaURL       string                 `json:"media_url"`
	Metadata       map[string]interface{} `json:"metadata"`
	Timestamp      time.Time              `json:"timestamp"`
}

type OrderIntentEvent struct {
	ConversationID string                 `json:"conversation_id"`
	UserID         string                 `json:"user_id"`
	Platform       string                 `json:"platform"`
	Intent         string                 `json:"intent"` // "place_order", "check_menu", "check_status"
	Products       []string               `json:"products"`
	Quantity       map[string]int         `json:"quantity"`
	Metadata       map[string]interface{} `json:"metadata"`
	Timestamp      time.Time              `json:"timestamp"`
}

type UserActivityEvent struct {
	UserID         string                 `json:"user_id"`
	ConversationID string                 `json:"conversation_id"`
	Platform       string                 `json:"platform"`
	Activity       string                 `json:"activity"` // "joined", "left", "typing", "idle"
	Metadata       map[string]interface{} `json:"metadata"`
	Timestamp      time.Time              `json:"timestamp"`
}
