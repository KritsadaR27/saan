package repository

import (
	"context"
	"time"

	"github.com/saan/chat-service/internal/domain/entity"
)

// MessageRepository defines the interface for message data operations
type MessageRepository interface {
	Create(ctx context.Context, message *entity.Message) error
	GetByID(ctx context.Context, id string) (*entity.Message, error)
	GetByConversationID(ctx context.Context, conversationID string, limit, offset int) ([]*entity.Message, error)
	GetRecentMessages(ctx context.Context, conversationID string, since time.Time) ([]*entity.Message, error)
	Update(ctx context.Context, message *entity.Message) error
	Delete(ctx context.Context, id string) error
	MarkAsRead(ctx context.Context, conversationID, userID string) error
}

// ConversationRepository defines the interface for conversation data operations
type ConversationRepository interface {
	Create(ctx context.Context, conversation *entity.Conversation) error
	GetByID(ctx context.Context, id string) (*entity.Conversation, error)
	GetByUserID(ctx context.Context, userID string) ([]*entity.Conversation, error)
	GetByUserAndPlatform(ctx context.Context, userID string, platform entity.Platform) (*entity.Conversation, error)
	Update(ctx context.Context, conversation *entity.Conversation) error
	UpdateLastActivity(ctx context.Context, id string, lastMessage string) error
	Delete(ctx context.Context, id string) error
	GetActiveConversations(ctx context.Context, limit, offset int) ([]*entity.Conversation, error)
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id string) (*entity.User, error)
	GetByPlatformID(ctx context.Context, platformID string, platform entity.Platform) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, query string, limit, offset int) ([]*entity.User, error)
}

// ChatSessionRepository defines the interface for chat session data operations
type ChatSessionRepository interface {
	Create(ctx context.Context, session *entity.ChatSession) error
	GetByUserID(ctx context.Context, userID string) (*entity.ChatSession, error)
	GetActiveByPlatform(ctx context.Context, platform entity.Platform) ([]*entity.ChatSession, error)
	Update(ctx context.Context, session *entity.ChatSession) error
	Delete(ctx context.Context, id string) error
	UpdatePing(ctx context.Context, id string) error
	CleanupInactiveSessions(ctx context.Context, before time.Time) error
}
