// webhooks/loyverse-webhook/internal/validator/loyverse.go
package validator

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"time"
)

// LoyverseValidator handles Loyverse webhook signature validation
type LoyverseValidator struct {
	secret string
}

// NewLoyverseValidator creates a new Loyverse webhook validator
func NewLoyverseValidator(secret string) *LoyverseValidator {
	return &LoyverseValidator{
		secret: secret,
	}
}

// WebhookPayload represents Loyverse webhook payload structure
type WebhookPayload struct {
	Type      string          `json:"type"`
	CreatedAt time.Time       `json:"created_at"`
	Data      json.RawMessage `json:"data"`
}

// ValidateSignature verifies the webhook signature using HMAC-SHA256
func (v *LoyverseValidator) ValidateSignature(body []byte, signature string) bool {
	if v.secret == "" {
		log.Println("Warning: No webhook secret configured, skipping signature validation")
		return true
	}

	mac := hmac.New(sha256.New, []byte(v.secret))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// ValidatePayload validates the webhook payload structure
func (v *LoyverseValidator) ValidatePayload(body []byte) (*WebhookPayload, error) {
	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	// Validate required fields
	if payload.Type == "" {
		return nil, &ValidationError{Message: "missing webhook type"}
	}

	if payload.Data == nil {
		return nil, &ValidationError{Message: "missing webhook data"}
	}

	return &payload, nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return "validation error: " + e.Message
}
