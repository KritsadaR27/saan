// webhooks/chat-webhook/internal/router/chat.go
package router

import (
	"github.com/gorilla/mux"
	"webhooks/chat-webhook/internal/facebook"
	"webhooks/chat-webhook/internal/line"
)

// ChatRouter handles routing for chat webhook endpoints
type ChatRouter struct {
	facebookHandler *facebook.Handler
	lineHandler     *line.Handler
}

// NewChatRouter creates a new chat router
func NewChatRouter(facebookHandler *facebook.Handler, lineHandler *line.Handler) *ChatRouter {
	return &ChatRouter{
		facebookHandler: facebookHandler,
		lineHandler:     lineHandler,
	}
}

// RegisterRoutes registers all chat webhook routes
func (cr *ChatRouter) RegisterRoutes(router *mux.Router) {
	// Facebook webhook routes
	router.HandleFunc("/webhook/facebook", cr.facebookHandler.HandleWebhook).Methods("GET", "POST")
	
	// LINE webhook routes  
	router.HandleFunc("/webhook/line", cr.lineHandler.HandleWebhook).Methods("POST")
}
