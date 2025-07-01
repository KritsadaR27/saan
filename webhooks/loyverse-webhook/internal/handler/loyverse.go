// webhooks/loyverse-webhook/internal/handler/loyverse.go
package handler

import (
	"context"
	"io"
	"log"
	"net/http"

	"webhooks/loyverse-webhook/internal/processor"
	"webhooks/loyverse-webhook/internal/validator"
)

// LoyverseHandler handles HTTP requests for Loyverse webhooks
type LoyverseHandler struct {
	validator *validator.LoyverseValidator
	processor *processor.LoyverseProcessor
}

// NewLoyverseHandler creates a new Loyverse webhook handler
func NewLoyverseHandler(validator *validator.LoyverseValidator, processor *processor.LoyverseProcessor) *LoyverseHandler {
	return &LoyverseHandler{
		validator: validator,
		processor: processor,
	}
}

// HandleWebhook handles incoming Loyverse webhook requests
func (h *LoyverseHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Log request for debugging
	log.Printf("Received Loyverse webhook from %s", r.RemoteAddr)

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading webhook body: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate signature
	signature := r.Header.Get("X-Loyverse-Signature")
	if !h.validator.ValidateSignature(body, signature) {
		log.Printf("Invalid webhook signature from %s", r.RemoteAddr)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Validate payload structure
	payload, err := h.validator.ValidatePayload(body)
	if err != nil {
		log.Printf("Invalid webhook payload: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Process webhook asynchronously
	go func() {
		ctx := context.Background()
		if err := h.processor.ProcessWebhook(ctx, payload); err != nil {
			log.Printf("Error processing webhook: %v", err)
			// Note: We don't return error to client as webhook is already accepted
		}
	}()

	// Return success immediately (async processing pattern)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"accepted","message":"Webhook received and queued for processing"}`))

	log.Printf("Webhook accepted: type=%s", payload.Type)
}
