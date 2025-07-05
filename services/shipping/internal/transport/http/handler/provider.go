package handler

import (
	"net/http"

	"shipping/internal/application"
)

// ProviderHandler handles provider-related HTTP requests
type ProviderHandler struct {
	providerUseCase *application.ProviderUseCase
}

// NewProviderHandler creates a new provider handler
func NewProviderHandler(providerUseCase *application.ProviderUseCase) *ProviderHandler {
	return &ProviderHandler{
		providerUseCase: providerUseCase,
	}
}

// GetProviders retrieves all providers
func (h *ProviderHandler) GetProviders(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement based on available use case methods
	writeErrorResponse(w, r, http.StatusNotImplemented, "NOT_IMPLEMENTED", "GetProviders not implemented", "")
}

// GetProvider retrieves a provider by ID
func (h *ProviderHandler) GetProvider(w http.ResponseWriter, r *http.Request) {
	// TODO: Extract ID from URL params
	writeErrorResponse(w, r, http.StatusNotImplemented, "NOT_IMPLEMENTED", "GetProvider not implemented", "")
}

// CreateProvider creates a new provider
func (h *ProviderHandler) CreateProvider(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement provider creation
	writeErrorResponse(w, r, http.StatusNotImplemented, "NOT_IMPLEMENTED", "CreateProvider not implemented", "")
}

// UpdateProvider updates a provider
func (h *ProviderHandler) UpdateProvider(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement provider update
	writeErrorResponse(w, r, http.StatusNotImplemented, "NOT_IMPLEMENTED", "UpdateProvider not implemented", "")
}

// DeleteProvider deletes a provider
func (h *ProviderHandler) DeleteProvider(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement provider deletion
	writeErrorResponse(w, r, http.StatusNotImplemented, "NOT_IMPLEMENTED", "DeleteProvider not implemented", "")
}
