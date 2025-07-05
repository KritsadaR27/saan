package handler

import (
	"net/http"

	"shipping/internal/application"
)

// CoverageHandler handles coverage area-related HTTP requests
type CoverageHandler struct {
	coverageUseCase *application.CoverageUseCase
}

// NewCoverageHandler creates a new coverage handler
func NewCoverageHandler(coverageUseCase *application.CoverageUseCase) *CoverageHandler {
	return &CoverageHandler{
		coverageUseCase: coverageUseCase,
	}
}

// GetCoverageAreas retrieves all coverage areas
func (h *CoverageHandler) GetCoverageAreas(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement based on available use case methods
	writeErrorResponse(w, r, http.StatusNotImplemented, "NOT_IMPLEMENTED", "GetCoverageAreas not implemented", "")
}

// CheckCoverage checks if an address is covered
func (h *CoverageHandler) CheckCoverage(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement coverage checking
	writeErrorResponse(w, r, http.StatusNotImplemented, "NOT_IMPLEMENTED", "CheckCoverage not implemented", "")
}
