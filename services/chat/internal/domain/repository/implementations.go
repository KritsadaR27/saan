package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"github.com/google/uuid"
	"chat/internal/domain/entity"
)

// messageRepository implements MessageRepository
type messageRepository struct {
	db *gorm.DB
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, message *entity.Message) error {
	if message.ID == "" {
		message.ID = uuid.New().String()
	}
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *messageRepository) GetByID(ctx context.Context, id string) (*entity.Message, error) {
	var message entity.Message
	err := r.db.WithContext(ctx).
		Preload("Conversation").
		Preload("User").
		First(&message, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *messageRepository) GetByConversationID(ctx context.Context, conversationID string, limit, offset int) ([]*entity.Message, error) {
	var messages []*entity.Message
	err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Preload("User").
		Find(&messages).Error
	return messages, err
}

func (r *messageRepository) GetRecentMessages(ctx context.Context, conversationID string, since time.Time) ([]*entity.Message, error) {
	var messages []*entity.Message
	err := r.db.WithContext(ctx).
		Where("conversation_id = ? AND timestamp > ?", conversationID, since).
		Order("timestamp ASC").
		Preload("User").
		Find(&messages).Error
	return messages, err
}

func (r *messageRepository) Update(ctx context.Context, message *entity.Message) error {
	return r.db.WithContext(ctx).Save(message).Error
}

func (r *messageRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Message{}, "id = ?", id).Error
}

func (r *messageRepository) MarkAsRead(ctx context.Context, conversationID, userID string) error {
	return r.db.WithContext(ctx).
		Model(&entity.Message{}).
		Where("conversation_id = ? AND user_id != ? AND is_read = false", conversationID, userID).
		Update("is_read", true).Error
}

// conversationRepository implements ConversationRepository
type conversationRepository struct {
	db *gorm.DB
}

// NewConversationRepository creates a new conversation repository
func NewConversationRepository(db *gorm.DB) ConversationRepository {
	return &conversationRepository{db: db}
}

func (r *conversationRepository) Create(ctx context.Context, conversation *entity.Conversation) error {
	if conversation.ID == "" {
		conversation.ID = uuid.New().String()
	}
	return r.db.WithContext(ctx).Create(conversation).Error
}

func (r *conversationRepository) GetByID(ctx context.Context, id string) (*entity.Conversation, error) {
	var conversation entity.Conversation
	err := r.db.WithContext(ctx).
		Preload("User").
		First(&conversation, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

func (r *conversationRepository) GetByUserID(ctx context.Context, userID string) ([]*entity.Conversation, error) {
	var conversations []*entity.Conversation
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("last_activity DESC").
		Preload("User").
		Find(&conversations).Error
	return conversations, err
}

func (r *conversationRepository) GetByUserAndPlatform(ctx context.Context, userID string, platform entity.Platform) (*entity.Conversation, error) {
	var conversation entity.Conversation
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND platform = ?", userID, platform).
		Preload("User").
		First(&conversation).Error
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

func (r *conversationRepository) Update(ctx context.Context, conversation *entity.Conversation) error {
	return r.db.WithContext(ctx).Save(conversation).Error
}

func (r *conversationRepository) UpdateLastActivity(ctx context.Context, id string, lastMessage string) error {
	return r.db.WithContext(ctx).
		Model(&entity.Conversation{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_message":  lastMessage,
			"last_activity": time.Now(),
			"updated_at":    time.Now(),
		}).Error
}

func (r *conversationRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Conversation{}, "id = ?", id).Error
}

func (r *conversationRepository) GetActiveConversations(ctx context.Context, limit, offset int) ([]*entity.Conversation, error) {
	var conversations []*entity.Conversation
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("last_activity DESC").
		Limit(limit).
		Offset(offset).
		Preload("User").
		Find(&conversations).Error
	return conversations, err
}

// userRepository implements UserRepository
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByPlatformID(ctx context.Context, platformID string, platform entity.Platform) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Where("platform_id = ? AND platform = ?", platformID, platform).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.User{}, "id = ?", id).Error
}

func (r *userRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entity.User, error) {
	var users []*entity.User
	searchQuery := "%" + query + "%"
	err := r.db.WithContext(ctx).
		Where("display_name ILIKE ? OR phone ILIKE ? OR email ILIKE ?", searchQuery, searchQuery, searchQuery).
		Limit(limit).
		Offset(offset).
		Find(&users).Error
	return users, err
}
