package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// APIResponse represents a standard API response
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// APIError represents an API error
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// PaginationResponse represents paginated response
type PaginatedResponse struct {
	Data       interface{}       `json:"data"`
	Pagination PaginationDetails `json:"pagination"`
}

// PaginationDetails contains pagination metadata
type PaginationDetails struct {
	Page      int   `json:"page"`
	Limit     int   `json:"limit"`
	Total     int64 `json:"total"`
	TotalPages int  `json:"total_pages"`
}

// writeJSONResponse writes a JSON response
func writeJSONResponse(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) {
	requestID := getRequestID(r)
	
	response := APIResponse{
		Success:   statusCode < 400,
		Data:      data,
		RequestID: requestID,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// writeErrorResponse writes an error JSON response
func writeErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, code, message, details string) {
	requestID := getRequestID(r)
	
	response := APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
		RequestID: requestID,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// writePaginatedResponse writes a paginated JSON response
func writePaginatedResponse(w http.ResponseWriter, r *http.Request, data interface{}, page, limit int, total int64) {
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	
	paginatedData := PaginatedResponse{
		Data: data,
		Pagination: PaginationDetails{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	writeJSONResponse(w, r, http.StatusOK, paginatedData)
}

// getRequestID extracts request ID from context
func getRequestID(r *http.Request) string {
	if requestID := r.Context().Value("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return uuid.New().String()
}

// parseUUID parses UUID from string
func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// Common error responses
func writeBadRequestError(w http.ResponseWriter, r *http.Request, message string) {
	writeErrorResponse(w, r, http.StatusBadRequest, "BAD_REQUEST", message, "")
}

func writeNotFoundError(w http.ResponseWriter, r *http.Request, resource string) {
	writeErrorResponse(w, r, http.StatusNotFound, "NOT_FOUND", resource+" not found", "")
}

func writeInternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	writeErrorResponse(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", err.Error())
}

func writeValidationError(w http.ResponseWriter, r *http.Request, message string) {
	writeErrorResponse(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", message, "")
}

func writeConflictError(w http.ResponseWriter, r *http.Request, message string) {
	writeErrorResponse(w, r, http.StatusConflict, "CONFLICT", message, "")
}

func writeUnauthorizedError(w http.ResponseWriter, r *http.Request) {
	writeErrorResponse(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", "")
}

func writeForbiddenError(w http.ResponseWriter, r *http.Request) {
	writeErrorResponse(w, r, http.StatusForbidden, "FORBIDDEN", "Access denied", "")
}
