package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"chat/internal/config"
	"chat/internal/domain/entity"
	"chat/internal/domain/repository"
	"chat/internal/infrastructure/kafka"
	"chat/internal/infrastructure/redis"
	"chat/internal/infrastructure/websocket"
)

// ChatService handles chat-related business logic
type ChatService struct {
	messageRepo      repository.MessageRepository
	conversationRepo repository.ConversationRepository
	userRepo         repository.UserRepository
	redisClient      *redis.Client
	kafkaProducer    *kafka.Producer
	wsHub            *websocket.Hub
	config           *config.Config
}

// NewChatService creates a new chat service
func NewChatService(
	messageRepo repository.MessageRepository,
	conversationRepo repository.ConversationRepository,
	userRepo repository.UserRepository,
	redisClient *redis.Client,
	kafkaProducer *kafka.Producer,
	wsHub *websocket.Hub,
	config *config.Config,
) *ChatService {
	return &ChatService{
		messageRepo:      messageRepo,
		conversationRepo: conversationRepo,
		userRepo:         userRepo,
		redisClient:      redisClient,
		kafkaProducer:    kafkaProducer,
		wsHub:            wsHub,
		config:           config,
	}
}

// ProcessMessage processes an incoming message from any platform
func (s *ChatService) ProcessMessage(ctx context.Context, req ProcessMessageRequest) (*ProcessMessageResponse, error) {
	// Get or create user
	user, err := s.getOrCreateUser(ctx, req.UserID, req.Platform, req.UserInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create user: %w", err)
	}

	// Get or create conversation
	conversation, err := s.getOrCreateConversation(ctx, user.ID, req.Platform)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create conversation: %w", err)
	}

	// Create and save message
	message := &entity.Message{
		ID:             uuid.New().String(),
		ConversationID: conversation.ID,
		UserID:         user.ID,
		Platform:       req.Platform,
		Direction:      entity.MessageDirectionIncoming,
		Type:           req.MessageType,
		Content:        req.Content,
		MediaURL:       req.MediaURL,
		Metadata:       req.Metadata,
		PlatformMsgID:  req.PlatformMessageID,
		IsRead:         false,
		Timestamp:      time.Now(),
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	// Update conversation last activity
	if err := s.conversationRepo.UpdateLastActivity(ctx, conversation.ID, message.Content); err != nil {
		logrus.Errorf("Failed to update conversation last activity: %v", err)
	}

	// Process message content for AI/order intent
	response := s.processMessageContent(ctx, message, conversation, user)

	// Publish message event to Kafka
	s.publishMessageEvent(ctx, message)

	// Send real-time notification via WebSocket
	s.sendWebSocketNotification(conversation.ID, message)

	// Send auto-response if generated
	if response.AutoResponse != "" {
		autoResponseMsg, err := s.sendAutoResponse(ctx, conversation.ID, user.ID, req.Platform, response.AutoResponse)
		if err != nil {
			logrus.Errorf("Failed to send auto response: %v", err)
		} else {
			response.ResponseMessage = autoResponseMsg
		}
	}

	response.Message = message
	response.Conversation = conversation
	response.User = user

	return response, nil
}

// SendMessage sends a message to a platform
func (s *ChatService) SendMessage(ctx context.Context, req SendMessageRequest) (*entity.Message, error) {
	// Get conversation
	conversation, err := s.conversationRepo.GetByID(ctx, req.ConversationID)
	if err != nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}

	// Create outgoing message
	message := &entity.Message{
		ID:             uuid.New().String(),
		ConversationID: req.ConversationID,
		UserID:         req.UserID,
		Platform:       conversation.Platform,
		Direction:      entity.MessageDirectionOutgoing,
		Type:           req.MessageType,
		Content:        req.Content,
		MediaURL:       req.MediaURL,
		Metadata:       req.Metadata,
		IsRead:         true, // Outgoing messages are marked as read
		Timestamp:      time.Now(),
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	// Update conversation last activity
	if err := s.conversationRepo.UpdateLastActivity(ctx, conversation.ID, message.Content); err != nil {
		logrus.Errorf("Failed to update conversation last activity: %v", err)
	}

	// Publish message event
	s.publishMessageEvent(ctx, message)

	// Send real-time notification via WebSocket
	s.sendWebSocketNotification(conversation.ID, message)

	return message, nil
}

// GetConversationMessages retrieves messages for a conversation
func (s *ChatService) GetConversationMessages(ctx context.Context, conversationID string, limit, offset int) ([]*entity.Message, error) {
	return s.messageRepo.GetByConversationID(ctx, conversationID, limit, offset)
}

// GetUserConversations retrieves conversations for a user
func (s *ChatService) GetUserConversations(ctx context.Context, userID string) ([]*entity.Conversation, error) {
	return s.conversationRepo.GetByUserID(ctx, userID)
}

// GetActiveConversations retrieves active conversations
func (s *ChatService) GetActiveConversations(ctx context.Context, limit, offset int) ([]*entity.Conversation, error) {
	return s.conversationRepo.GetActiveConversations(ctx, limit, offset)
}

// MarkMessagesAsRead marks messages in a conversation as read
func (s *ChatService) MarkMessagesAsRead(ctx context.Context, conversationID, userID string) error {
	return s.messageRepo.MarkAsRead(ctx, conversationID, userID)
}

// Helper methods

func (s *ChatService) getOrCreateUser(ctx context.Context, platformID string, platform entity.Platform, userInfo map[string]interface{}) (*entity.User, error) {
	// Try to get existing user
	user, err := s.userRepo.GetByPlatformID(ctx, platformID, platform)
	if err == nil {
		return user, nil
	}

	// Create new user
	displayName := platformID
	if name, ok := userInfo["display_name"].(string); ok && name != "" {
		displayName = name
	}

	avatarURL := ""
	if avatar, ok := userInfo["avatar_url"].(string); ok {
		avatarURL = avatar
	}

	user = &entity.User{
		ID:          uuid.New().String(),
		PlatformID:  platformID,
		Platform:    platform,
		DisplayName: displayName,
		AvatarURL:   avatarURL,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *ChatService) getOrCreateConversation(ctx context.Context, userID string, platform entity.Platform) (*entity.Conversation, error) {
	// Try to get existing conversation
	conversation, err := s.conversationRepo.GetByUserAndPlatform(ctx, userID, platform)
	if err == nil {
		return conversation, nil
	}

	// Create new conversation
	conversation = &entity.Conversation{
		ID:           uuid.New().String(),
		UserID:       userID,
		Platform:     platform,
		Status:       "active",
		LastActivity: time.Now(),
	}

	if err := s.conversationRepo.Create(ctx, conversation); err != nil {
		return nil, err
	}

	return conversation, nil
}

func (s *ChatService) processMessageContent(ctx context.Context, message *entity.Message, conversation *entity.Conversation, user *entity.User) *ProcessMessageResponse {
	response := &ProcessMessageResponse{}

	// Simple keyword-based processing (can be enhanced with AI later)
	content := strings.ToLower(message.Content)

	// Check for order intent
	if s.containsOrderKeywords(content) {
		response.Intent = "place_order"
		response.AutoResponse = "สวัสดีครับ! เมนูอะไรดีครับวันนี้? พิมพ์ 'เมนู' เพื่อดูรายการอาหารทั้งหมด"
		
		// Publish order intent event
		s.publishOrderIntentEvent(ctx, message, "place_order")
	} else if s.containsMenuKeywords(content) {
		response.Intent = "check_menu"
		response.AutoResponse = "🍜 เมนูแนะนำวันนี้:\n1. ข้าวมันไก่ - 50 บาท\n2. ก๋วยเตี๋ยวหมู - 45 บาท\n3. ผัดไทย - 60 บาท\n\nพิมพ์หมายเลขเพื่อสั่งได้เลยครับ!"
	} else if strings.Contains(content, "สวัสดี") || strings.Contains(content, "hello") {
		response.Intent = "greeting"
		response.AutoResponse = fmt.Sprintf("สวัสดีครับคุณ %s! ยินดีต้อนรับสู่ร้านอาหารของเรา 🍽️ มีอะไรให้ช่วยไหมครับ?", user.DisplayName)
	} else {
		response.Intent = "general"
		response.AutoResponse = "ขอบคุณสำหรับข้อความครับ เรากำลังดำเนินการตอบกลับให้คุณในไม่ช้า"
	}

	return response
}

func (s *ChatService) containsOrderKeywords(content string) bool {
	orderKeywords := []string{"สั่ง", "ขอ", "เอา", "order", "want", "สั่งอาหาร"}
	for _, keyword := range orderKeywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}
	return false
}

func (s *ChatService) containsMenuKeywords(content string) bool {
	menuKeywords := []string{"เมนู", "menu", "รายการ", "อาหาร", "ขายอะไร"}
	for _, keyword := range menuKeywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}
	return false
}

func (s *ChatService) sendAutoResponse(ctx context.Context, conversationID, userID string, platform entity.Platform, content string) (*entity.Message, error) {
	message := &entity.Message{
		ID:             uuid.New().String(),
		ConversationID: conversationID,
		UserID:         userID,
		Platform:       platform,
		Direction:      entity.MessageDirectionOutgoing,
		Type:           entity.MessageTypeText,
		Content:        content,
		IsRead:         true,
		Timestamp:      time.Now(),
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, err
	}

	// Send via WebSocket
	s.sendWebSocketNotification(conversationID, message)

	// Publish to Kafka
	s.publishMessageEvent(ctx, message)

	return message, nil
}

func (s *ChatService) publishMessageEvent(ctx context.Context, message *entity.Message) {
	event := kafka.ChatMessageEvent{
		MessageID:      message.ID,
		ConversationID: message.ConversationID,
		UserID:         message.UserID,
		Platform:       string(message.Platform),
		Direction:      string(message.Direction),
		Type:           string(message.Type),
		Content:        message.Content,
		MediaURL:       message.MediaURL,
		Timestamp:      message.Timestamp,
	}

	if err := s.kafkaProducer.PublishChatMessage(ctx, event); err != nil {
		logrus.Errorf("Failed to publish chat message event: %v", err)
	}
}

func (s *ChatService) publishOrderIntentEvent(ctx context.Context, message *entity.Message, intent string) {
	event := kafka.OrderIntentEvent{
		ConversationID: message.ConversationID,
		UserID:         message.UserID,
		Platform:       string(message.Platform),
		Intent:         intent,
		Timestamp:      time.Now(),
	}

	if err := s.kafkaProducer.PublishOrderIntent(ctx, event); err != nil {
		logrus.Errorf("Failed to publish order intent event: %v", err)
	}
}

func (s *ChatService) sendWebSocketNotification(conversationID string, message *entity.Message) {
	wsMessage := websocket.Message{
		Type:           "new_message",
		ConversationID: conversationID,
		UserID:         message.UserID,
		Content:        message.Content,
		Timestamp:      message.Timestamp,
		Metadata: map[string]interface{}{
			"message_id": message.ID,
			"direction":  message.Direction,
			"type":       message.Type,
		},
	}

	s.wsHub.BroadcastToConversation(conversationID, wsMessage)
}

// Request/Response types
type ProcessMessageRequest struct {
	UserID            string                 `json:"user_id"`
	Platform          entity.Platform        `json:"platform"`
	MessageType       entity.MessageType     `json:"message_type"`
	Content           string                 `json:"content"`
	MediaURL          string                 `json:"media_url"`
	Metadata          string                 `json:"metadata"`
	PlatformMessageID string                 `json:"platform_message_id"`
	UserInfo          map[string]interface{} `json:"user_info"`
}

type ProcessMessageResponse struct {
	Message         *entity.Message      `json:"message"`
	Conversation    *entity.Conversation `json:"conversation"`
	User            *entity.User         `json:"user"`
	ResponseMessage *entity.Message      `json:"response_message,omitempty"`
	Intent          string               `json:"intent"`
	AutoResponse    string               `json:"auto_response"`
}

type SendMessageRequest struct {
	ConversationID string             `json:"conversation_id"`
	UserID         string             `json:"user_id"`
	MessageType    entity.MessageType `json:"message_type"`
	Content        string             `json:"content"`
	MediaURL       string             `json:"media_url"`
	Metadata       string             `json:"metadata"`
}
