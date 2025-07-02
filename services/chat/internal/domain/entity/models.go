package entity

import (
	"time"
	"gorm.io/gorm"
)

// MessageType represents the type of message
type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeImage    MessageType = "image"
	MessageTypeVideo    MessageType = "video"
	MessageTypeAudio    MessageType = "audio"
	MessageTypeFile     MessageType = "file"
	MessageTypeLocation MessageType = "location"
	MessageTypeOrder    MessageType = "order"
)

// MessageDirection indicates whether the message is incoming or outgoing
type MessageDirection string

const (
	MessageDirectionIncoming MessageDirection = "incoming"
	MessageDirectionOutgoing MessageDirection = "outgoing"
)

// Platform represents the messaging platform
type Platform string

const (
	PlatformLINE     Platform = "line"
	PlatformFacebook Platform = "facebook"
	PlatformWhatsApp Platform = "whatsapp"
	PlatformWebChat  Platform = "webchat"
)

// User represents a chat user
type User struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	PlatformID  string    `json:"platform_id" gorm:"uniqueIndex:idx_platform_user"`
	Platform    Platform  `json:"platform" gorm:"uniqueIndex:idx_platform_user"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	Phone       string    `json:"phone"`
	Email       string    `json:"email"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	Conversations []Conversation `json:"conversations" gorm:"foreignKey:UserID"`
	Messages      []Message      `json:"messages" gorm:"foreignKey:UserID"`
}

// Conversation represents a chat conversation
type Conversation struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	UserID      string    `json:"user_id" gorm:"index"`
	Platform    Platform  `json:"platform"`
	Status      string    `json:"status"` // active, archived, blocked
	LastMessage string    `json:"last_message"`
	LastActivity time.Time `json:"last_activity"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	User     User      `json:"user" gorm:"foreignKey:UserID"`
	Messages []Message `json:"messages" gorm:"foreignKey:ConversationID"`
}

// Message represents a chat message
type Message struct {
	ID             string           `json:"id" gorm:"primaryKey"`
	ConversationID string           `json:"conversation_id" gorm:"index"`
	UserID         string           `json:"user_id" gorm:"index"`
	Platform       Platform         `json:"platform"`
	Direction      MessageDirection `json:"direction"`
	Type           MessageType      `json:"type"`
	Content        string           `json:"content"`
	MediaURL       string           `json:"media_url"`
	Metadata       string           `json:"metadata"` // JSON string for additional data
	PlatformMsgID  string           `json:"platform_msg_id" gorm:"uniqueIndex:idx_platform_msg"`
	IsRead         bool             `json:"is_read"`
	Timestamp      time.Time        `json:"timestamp"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
	DeletedAt      gorm.DeletedAt   `json:"deleted_at" gorm:"index"`

	// Relationships
	Conversation Conversation `json:"conversation" gorm:"foreignKey:ConversationID"`
	User         User         `json:"user" gorm:"foreignKey:UserID"`
}

// ChatSession represents an active chat session for real-time messaging
type ChatSession struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	UserID       string    `json:"user_id" gorm:"index"`
	Platform     Platform  `json:"platform"`
	IsActive     bool      `json:"is_active"`
	SocketID     string    `json:"socket_id"`
	LastPing     time.Time `json:"last_ping"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	User User `json:"user" gorm:"foreignKey:UserID"`
}
