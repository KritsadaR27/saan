package handler

import (
	"net/http"

	"shipping/internal/application"
)

// TrackingHandler handles tracking-related HTTP requests
type TrackingHandler struct {
	trackingUseCase *application.TrackingUseCase
}

// NewTrackingHandler creates a new tracking handler
func NewTrackingHandler(trackingUseCase *application.TrackingUseCase) *TrackingHandler {
	return &TrackingHandler{
		trackingUseCase: trackingUseCase,
	}
}

// GetTracking retrieves tracking information by tracking number
func (h *TrackingHandler) GetTracking(w http.ResponseWriter, r *http.Request) {
	// TODO: Extract tracking number from URL params and implement
	writeErrorResponse(w, r, http.StatusNotImplemented, "NOT_IMPLEMENTED", "GetTracking not implemented", "")
}

// UpdateTracking updates tracking information
func (h *TrackingHandler) UpdateTracking(w http.ResponseWriter, r *http.Request) {
	// TODO: Extract tracking number and implement update
	writeErrorResponse(w, r, http.StatusNotImplemented, "NOT_IMPLEMENTED", "UpdateTracking not implemented", "")
}
