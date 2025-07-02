package websocket

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for thread-safe operations
	mutex sync.RWMutex
}

// Client is a middleman between the websocket connection and the hub
type Client struct {
	hub *Hub

	// The websocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan []byte

	// User ID for identification
	userID string

	// Platform information
	platform string

	// Conversation ID
	conversationID string
}

// Message represents a WebSocket message
type Message struct {
	Type           string                 `json:"type"`
	ConversationID string                 `json:"conversation_id,omitempty"`
	UserID         string                 `json:"user_id,omitempty"`
	Content        string                 `json:"content,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
}

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin in development
		return true
	},
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// Run starts the WebSocket hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			logrus.Infof("Client connected: %s", client.userID)

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				logrus.Infof("Client disconnected: %s", client.userID)
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// BroadcastToConversation sends a message to all clients in a conversation
func (h *Hub) BroadcastToConversation(conversationID string, message Message) {
	data, err := json.Marshal(message)
	if err != nil {
		logrus.Errorf("Error marshaling message: %v", err)
		return
	}

	h.mutex.RLock()
	for client := range h.clients {
		if client.conversationID == conversationID {
			select {
			case client.send <- data:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
	h.mutex.RUnlock()
}

// BroadcastToUser sends a message to a specific user
func (h *Hub) BroadcastToUser(userID string, message Message) {
	data, err := json.Marshal(message)
	if err != nil {
		logrus.Errorf("Error marshaling message: %v", err)
		return
	}

	h.mutex.RLock()
	for client := range h.clients {
		if client.userID == userID {
			select {
			case client.send <- data:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
	h.mutex.RUnlock()
}

// HandleWebSocket handles WebSocket connections
func (h *Hub) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.Errorf("Failed to upgrade connection: %v", err)
		return
	}

	// Get user info from query parameters
	userID := c.Query("user_id")
	platform := c.Query("platform")
	conversationID := c.Query("conversation_id")

	if userID == "" {
		logrus.Error("Missing user_id parameter")
		conn.Close()
		return
	}

	client := &Client{
		hub:            h,
		conn:           conn,
		send:           make(chan []byte, 256),
		userID:         userID,
		platform:       platform,
		conversationID: conversationID,
	}

	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.Errorf("WebSocket error: %v", err)
			}
			break
		}

		// Process received message
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			logrus.Errorf("Error unmarshaling message: %v", err)
			continue
		}

		// Add metadata
		msg.UserID = c.userID
		msg.Timestamp = time.Now()

		// Broadcast the message
		data, _ := json.Marshal(msg)
		c.hub.broadcast <- data
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// GetConnectedUsers returns the list of connected user IDs
func (h *Hub) GetConnectedUsers() []string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	users := make(map[string]bool)
	for client := range h.clients {
		users[client.userID] = true
	}

	result := make([]string, 0, len(users))
	for userID := range users {
		result = append(result, userID)
	}

	return result
}
