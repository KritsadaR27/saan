package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"shipping/internal/application"
	"shipping/internal/domain/entity"
)

// DeliveryHandler handles delivery-related HTTP requests
type DeliveryHandler struct {
	deliveryUseCase *application.DeliveryUsecase
}

// NewDeliveryHandler creates a new delivery handler
func NewDeliveryHandler(deliveryUseCase *application.DeliveryUsecase) *DeliveryHandler {
	return &DeliveryHandler{
		deliveryUseCase: deliveryUseCase,
	}
}

// CreateDelivery creates a new delivery
func (h *DeliveryHandler) CreateDelivery(w http.ResponseWriter, r *http.Request) {
	var req application.CreateDeliveryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequestError(w, r, "Invalid request body")
		return
	}

	response, err := h.deliveryUseCase.CreateDelivery(r.Context(), &req)
	if err != nil {
		writeInternalServerError(w, r, err)
		return
	}

	writeJSONResponse(w, r, http.StatusCreated, response)
}

// GetDelivery retrieves a delivery by ID
func (h *DeliveryHandler) GetDelivery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseUUID(vars["id"])
	if err != nil {
		writeBadRequestError(w, r, "Invalid delivery ID")
		return
	}

	delivery, err := h.deliveryUseCase.GetDelivery(r.Context(), id)
	if err != nil {
		writeNotFoundError(w, r, "Delivery")
		return
	}

	writeJSONResponse(w, r, http.StatusOK, delivery)
}

// GetDeliveries retrieves deliveries with basic filtering
func (h *DeliveryHandler) GetDeliveries(w http.ResponseWriter, r *http.Request) {
	// For now, return a simple message since the use case doesn't have this method
	// TODO: Implement GetDeliveries method in the use case
	writeJSONResponse(w, r, http.StatusOK, map[string]string{
		"message": "Deliveries endpoint - method not implemented yet",
	})
}

// UpdateDelivery updates a delivery (placeholder)
func (h *DeliveryHandler) UpdateDelivery(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement UpdateDelivery method in the use case
	writeJSONResponse(w, r, http.StatusOK, map[string]string{
		"message": "Update delivery endpoint - method not implemented yet",
	})
}

// CancelDelivery cancels a delivery (placeholder)
func (h *DeliveryHandler) CancelDelivery(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement CancelDelivery method in the use case
	writeJSONResponse(w, r, http.StatusOK, map[string]string{
		"message": "Cancel delivery endpoint - method not implemented yet",
	})
}

// AssignDelivery assigns a delivery to a vehicle
func (h *DeliveryHandler) AssignDelivery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deliveryID, err := parseUUID(vars["id"])
	if err != nil {
		writeBadRequestError(w, r, "Invalid delivery ID")
		return
	}

	var req struct {
		VehicleID uuid.UUID `json:"vehicle_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequestError(w, r, "Invalid request body")
		return
	}

	err = h.deliveryUseCase.AssignVehicle(r.Context(), deliveryID, req.VehicleID)
	if err != nil {
		writeInternalServerError(w, r, err)
		return
	}

	writeJSONResponse(w, r, http.StatusOK, map[string]string{"message": "Delivery assigned successfully"})
}

// UpdateStatus updates delivery status
func (h *DeliveryHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deliveryID, err := parseUUID(vars["id"])
	if err != nil {
		writeBadRequestError(w, r, "Invalid delivery ID")
		return
	}

	var req struct {
		Status string     `json:"status"`
		UserID *uuid.UUID `json:"user_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequestError(w, r, "Invalid request body")
		return
	}

	// Convert string status to entity.DeliveryStatus
	var status entity.DeliveryStatus
	switch req.Status {
	case "pending":
		status = entity.DeliveryStatusPending
	case "planned":
		status = entity.DeliveryStatusPlanned
	case "dispatched":
		status = entity.DeliveryStatusDispatched
	case "in_transit":
		status = entity.DeliveryStatusInTransit
	case "delivered":
		status = entity.DeliveryStatusDelivered
	case "failed":
		status = entity.DeliveryStatusFailed
	case "cancelled":
		status = entity.DeliveryStatusCancelled
	default:
		writeBadRequestError(w, r, "Invalid delivery status")
		return
	}

	err = h.deliveryUseCase.UpdateDeliveryStatus(r.Context(), deliveryID, status, req.UserID)
	if err != nil {
		writeInternalServerError(w, r, err)
		return
	}

	writeJSONResponse(w, r, http.StatusOK, map[string]string{"message": "Status updated successfully"})
}

// GetDeliveryByOrder retrieves delivery by order ID (placeholder)
func (h *DeliveryHandler) GetDeliveryByOrder(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement GetDeliveryByOrder method in the use case
	writeJSONResponse(w, r, http.StatusOK, map[string]string{
		"message": "Get delivery by order endpoint - method not implemented yet",
	})
}

// GetByTrackingNumber retrieves delivery by tracking number (placeholder)
func (h *DeliveryHandler) GetByTrackingNumber(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement GetByTrackingNumber method in the use case
	writeJSONResponse(w, r, http.StatusOK, map[string]string{
		"message": "Get delivery by tracking number endpoint - method not implemented yet",
	})
}

// DeleteDelivery deletes a delivery
func (h *DeliveryHandler) DeleteDelivery(w http.ResponseWriter, r *http.Request) {
	// TODO: Extract ID from URL params and implement
	writeErrorResponse(w, r, http.StatusNotImplemented, "NOT_IMPLEMENTED", "DeleteDelivery not implemented", "")
}

// UpdateDeliveryStatus updates delivery status
func (h *DeliveryHandler) UpdateDeliveryStatus(w http.ResponseWriter, r *http.Request) {
	// TODO: Extract ID from URL params and implement
	writeErrorResponse(w, r, http.StatusNotImplemented, "NOT_IMPLEMENTED", "UpdateDeliveryStatus not implemented", "")
}

// GetDeliveryTracking gets delivery tracking information
func (h *DeliveryHandler) GetDeliveryTracking(w http.ResponseWriter, r *http.Request) {
	// TODO: Extract ID from URL params and implement
	writeErrorResponse(w, r, http.StatusNotImplemented, "NOT_IMPLEMENTED", "GetDeliveryTracking not implemented", "")
}

// GetDeliveryByTracking gets delivery by tracking number
func (h *DeliveryHandler) GetDeliveryByTracking(w http.ResponseWriter, r *http.Request) {
	// TODO: Extract tracking number from URL params and implement
	writeErrorResponse(w, r, http.StatusNotImplemented, "NOT_IMPLEMENTED", "GetDeliveryByTracking not implemented", "")
}

// SearchDeliveries searches deliveries based on criteria
func (h *DeliveryHandler) SearchDeliveries(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement search functionality
	writeErrorResponse(w, r, http.StatusNotImplemented, "NOT_IMPLEMENTED", "SearchDeliveries not implemented", "")
}
