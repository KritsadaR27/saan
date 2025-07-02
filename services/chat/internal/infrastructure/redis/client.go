package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// Client wraps Redis client with additional functionality
type Client struct {
	rdb *redis.Client
}

// NewClient creates a new Redis client
func NewClient(addr, password string) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	logrus.Info("Connected to Redis successfully")
	return &Client{rdb: rdb}, nil
}

// Set stores a value with expiration
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, data, expiration).Err()
}

// Get retrieves a value and unmarshals it
func (c *Client) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// Delete removes a key
func (c *Client) Delete(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

// Exists checks if a key exists
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.rdb.Exists(ctx, key).Result()
	return count > 0, err
}

// Publish publishes a message to a channel
func (c *Client) Publish(ctx context.Context, channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return c.rdb.Publish(ctx, channel, data).Err()
}

// Subscribe subscribes to a channel
func (c *Client) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return c.rdb.Subscribe(ctx, channels...)
}

// SetUserSession stores user session data
func (c *Client) SetUserSession(ctx context.Context, userID string, sessionData map[string]interface{}) error {
	key := "user_session:" + userID
	return c.Set(ctx, key, sessionData, 24*time.Hour)
}

// GetUserSession retrieves user session data
func (c *Client) GetUserSession(ctx context.Context, userID string) (map[string]interface{}, error) {
	key := "user_session:" + userID
	var sessionData map[string]interface{}
	err := c.Get(ctx, key, &sessionData)
	if err == redis.Nil {
		return make(map[string]interface{}), nil
	}
	return sessionData, err
}

// SetConversationState stores conversation state
func (c *Client) SetConversationState(ctx context.Context, conversationID string, state string) error {
	key := "conversation_state:" + conversationID
	return c.Set(ctx, key, state, time.Hour)
}

// GetConversationState retrieves conversation state
func (c *Client) GetConversationState(ctx context.Context, conversationID string) (string, error) {
	key := "conversation_state:" + conversationID
	var state string
	err := c.Get(ctx, key, &state)
	if err == redis.Nil {
		return "", nil
	}
	return state, err
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.rdb.Close()
}
