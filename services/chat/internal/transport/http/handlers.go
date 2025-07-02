package http

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/saan/chat-service/internal/application"
	"github.com/saan/chat-service/internal/config"
	"github.com/saan/chat-service/internal/domain/entity"
	"github.com/saan/chat-service/internal/infrastructure/websocket"
)

// Handlers contains HTTP handlers for the chat service
type Handlers struct {
	chatService *application.ChatService
	wsHub       *websocket.Hub
	config      *config.Config
}

// NewHandlers creates new HTTP handlers
func NewHandlers(chatService *application.ChatService, wsHub *websocket.Hub, config *config.Config) *Handlers {
	return &Handlers{
		chatService: chatService,
		wsHub:       wsHub,
		config:      config,
	}
}

// SetupRoutes configures the HTTP routes
func (h *Handlers) SetupRoutes(router *gin.Engine) {
	// Health check
	router.GET("/health", h.healthCheck)

	// WebSocket endpoint
	router.GET("/ws", h.wsHub.HandleWebSocket)

	// API routes
	api := router.Group("/api/v1")
	{
		// Message routes
		messages := api.Group("/messages")
		{
			messages.POST("/", h.processMessage)
			messages.POST("/send", h.sendMessage)
			messages.GET("/conversation/:id", h.getConversationMessages)
			messages.PUT("/read/:conversation_id", h.markMessagesAsRead)
		}

		// Conversation routes
		conversations := api.Group("/conversations")
		{
			conversations.GET("/user/:user_id", h.getUserConversations)
			conversations.GET("/active", h.getActiveConversations)
			conversations.GET("/:id", h.getConversation)
		}

		// Platform-specific webhook endpoints
		platforms := api.Group("/platforms")
		{
			platforms.POST("/line/webhook", h.handleLineWebhook)
			platforms.POST("/facebook/webhook", h.handleFacebookWebhook)
			platforms.GET("/facebook/webhook", h.verifyFacebookWebhook)
		}

		// Admin routes
		admin := api.Group("/admin")
		admin.Use(h.authMiddleware())
		{
			admin.GET("/stats", h.getStats)
			admin.GET("/users/connected", h.getConnectedUsers)
		}
	}
}

// Health check endpoint
func (h *Handlers) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "chat-service",
		"timestamp": time.Now(),
	})
}

// Process message endpoint
func (h *Handlers) processMessage(c *gin.Context) {
	var req application.ProcessMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.chatService.ProcessMessage(c.Request.Context(), req)
	if err != nil {
		logrus.Errorf("Failed to process message: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process message"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Send message endpoint
func (h *Handlers) sendMessage(c *gin.Context) {
	var req application.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message, err := h.chatService.SendMessage(c.Request.Context(), req)
	if err != nil {
		logrus.Errorf("Failed to send message: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	c.JSON(http.StatusOK, message)
}

// Get conversation messages
func (h *Handlers) getConversationMessages(c *gin.Context) {
	conversationID := c.Param("id")
	
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	messages, err := h.chatService.GetConversationMessages(c.Request.Context(), conversationID, limit, offset)
	if err != nil {
		logrus.Errorf("Failed to get conversation messages: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// Mark messages as read
func (h *Handlers) markMessagesAsRead(c *gin.Context) {
	conversationID := c.Param("conversation_id")
	userID := c.Query("user_id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	err := h.chatService.MarkMessagesAsRead(c.Request.Context(), conversationID, userID)
	if err != nil {
		logrus.Errorf("Failed to mark messages as read: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark messages as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Get user conversations
func (h *Handlers) getUserConversations(c *gin.Context) {
	userID := c.Param("user_id")

	conversations, err := h.chatService.GetUserConversations(c.Request.Context(), userID)
	if err != nil {
		logrus.Errorf("Failed to get user conversations: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get conversations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"conversations": conversations})
}

// Get active conversations
func (h *Handlers) getActiveConversations(c *gin.Context) {
	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	conversations, err := h.chatService.GetActiveConversations(c.Request.Context(), limit, offset)
	if err != nil {
		logrus.Errorf("Failed to get active conversations: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get conversations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"conversations": conversations})
}

// Get single conversation
func (h *Handlers) getConversation(c *gin.Context) {
	// Implementation would go here
	c.JSON(http.StatusOK, gin.H{"message": "Get conversation endpoint"})
}

// LINE webhook handler
func (h *Handlers) handleLineWebhook(c *gin.Context) {
	// Basic LINE webhook processing
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Process LINE events
	if events, ok := payload["events"].([]interface{}); ok {
		for _, event := range events {
			if eventMap, ok := event.(map[string]interface{}); ok {
				h.processLineEvent(c.Request.Context(), eventMap)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Facebook webhook handler
func (h *Handlers) handleFacebookWebhook(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Process Facebook events
	if entry, ok := payload["entry"].([]interface{}); ok {
		for _, e := range entry {
			if entryMap, ok := e.(map[string]interface{}); ok {
				h.processFacebookEvent(c.Request.Context(), entryMap)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Facebook webhook verification
func (h *Handlers) verifyFacebookWebhook(c *gin.Context) {
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	if mode == "subscribe" && token == h.config.FacebookAppSecret {
		c.String(http.StatusOK, challenge)
		return
	}

	c.JSON(http.StatusForbidden, gin.H{"error": "Verification failed"})
}

// Get statistics
func (h *Handlers) getStats(c *gin.Context) {
	connectedUsers := h.wsHub.GetConnectedUsers()
	
	c.JSON(http.StatusOK, gin.H{
		"connected_users": len(connectedUsers),
		"timestamp":       time.Now(),
	})
}

// Get connected users
func (h *Handlers) getConnectedUsers(c *gin.Context) {
	users := h.wsHub.GetConnectedUsers()
	c.JSON(http.StatusOK, gin.H{"users": users})
}

// Authentication middleware
func (h *Handlers) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token != "Bearer "+h.config.AdminToken {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// Helper methods for processing platform events
func (h *Handlers) processLineEvent(ctx context.Context, event map[string]interface{}) {
	// Extract LINE event data and convert to ProcessMessageRequest
	eventType := event["type"].(string)
	if eventType != "message" {
		return
	}

	source := event["source"].(map[string]interface{})
	userID := source["userId"].(string)
	
	message := event["message"].(map[string]interface{})
	messageType := message["type"].(string)
	
	var content string
	var mediaURL string
	
	switch messageType {
	case "text":
		content = message["text"].(string)
	case "image":
		mediaURL = message["originalContentUrl"].(string)
	}

	req := application.ProcessMessageRequest{
		UserID:            userID,
		Platform:          entity.PlatformLINE,
		MessageType:       entity.MessageType(messageType),
		Content:           content,
		MediaURL:          mediaURL,
		PlatformMessageID: message["id"].(string),
		UserInfo:          make(map[string]interface{}),
	}

	_, err := h.chatService.ProcessMessage(ctx, req)
	if err != nil {
		logrus.Errorf("Failed to process LINE message: %v", err)
	}
}

func (h *Handlers) processFacebookEvent(ctx context.Context, entry map[string]interface{}) {
	messaging := entry["messaging"].([]interface{})
	
	for _, msg := range messaging {
		msgMap := msg.(map[string]interface{})
		sender := msgMap["sender"].(map[string]interface{})
		userID := sender["id"].(string)
		
		if messageData, ok := msgMap["message"].(map[string]interface{}); ok {
			content := ""
			if text, exists := messageData["text"].(string); exists {
				content = text
			}
			
			req := application.ProcessMessageRequest{
				UserID:            userID,
				Platform:          entity.PlatformFacebook,
				MessageType:       entity.MessageTypeText,
				Content:           content,
				PlatformMessageID: messageData["mid"].(string),
				UserInfo:          make(map[string]interface{}),
			}

			_, err := h.chatService.ProcessMessage(ctx, req)
			if err != nil {
				logrus.Errorf("Failed to process Facebook message: %v", err)
			}
		}
	}
}
