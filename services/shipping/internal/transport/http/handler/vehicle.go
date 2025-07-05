package handler

import (
	"net/http"

	"shipping/internal/application"
)

// VehicleHandler handles vehicle-related HTTP requests
type VehicleHandler struct {
	vehicleUseCase *application.VehicleUseCase
}

// NewVehicleHandler creates a new vehicle handler
func NewVehicleHandler(vehicleUseCase *application.VehicleUseCase) *VehicleHandler {
	return &VehicleHandler{
		vehicleUseCase: vehicleUseCase,
	}
}

// GetVehicles retrieves available vehicles
func (h *VehicleHandler) GetVehicles(w http.ResponseWriter, r *http.Request) {
	vehicles, err := h.vehicleUseCase.GetAvailableVehicles(r.Context())
	if err != nil {
		writeInternalServerError(w, r, err)
		return
	}

	writeJSONResponse(w, r, http.StatusOK, vehicles)
}

// GetVehicle retrieves a vehicle by ID
func (h *VehicleHandler) GetVehicle(w http.ResponseWriter, r *http.Request) {
	// TODO: Extract ID from URL params
	writeErrorResponse(w, r, http.StatusNotImplemented, "NOT_IMPLEMENTED", "GetVehicle not implemented", "")
}

// CreateVehicle creates a new vehicle
func (h *VehicleHandler) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement vehicle creation
	writeErrorResponse(w, r, http.StatusNotImplemented, "NOT_IMPLEMENTED", "CreateVehicle not implemented", "")
}

// UpdateVehicle updates a vehicle
func (h *VehicleHandler) UpdateVehicle(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement vehicle update
	writeErrorResponse(w, r, http.StatusNotImplemented, "NOT_IMPLEMENTED", "UpdateVehicle not implemented", "")
}

// DeleteVehicle deletes a vehicle
func (h *VehicleHandler) DeleteVehicle(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement vehicle deletion
	writeErrorResponse(w, r, http.StatusNotImplemented, "NOT_IMPLEMENTED", "DeleteVehicle not implemented", "")
}
