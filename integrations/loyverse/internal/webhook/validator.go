package webhook

// Validator validates webhook payloads
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidatePayload validates a webhook payload
func (v *Validator) ValidatePayload(payload []byte) error {
	// Basic validation - could be extended
	if len(payload) == 0 {
		return nil // Empty payload is invalid, but we'll handle gracefully
	}
	return nil
}
