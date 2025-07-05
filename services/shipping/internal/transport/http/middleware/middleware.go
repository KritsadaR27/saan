package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
)

// CORS adds CORS headers to responses
func CORS() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
			w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// JSON sets JSON content type header
func JSON() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			next.ServeHTTP(w, r)
		})
	}
}

// RequestID adds a unique request ID to each request
func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			w.Header().Set("X-Request-ID", requestID)
			ctx := context.WithValue(r.Context(), "request_id", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Logger logs HTTP requests
func Logger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := r.Context().Value("request_id")

			// Create a custom ResponseWriter to capture status code
			wrapper := &responseWrapper{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapper, r)

			duration := time.Since(start)
			log.Printf("[%s] %s %s %d %v", 
				requestID, 
				r.Method, 
				r.URL.Path, 
				wrapper.statusCode, 
				duration)
		})
	}
}

// Recovery recovers from panics and returns a 500 error
func Recovery() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					requestID := r.Context().Value("request_id")
					log.Printf("[%s] PANIC: %v\n%s", requestID, err, debug.Stack())
					
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, `{"error": "Internal server error", "request_id": "%s"}`, requestID)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// responseWrapper wraps http.ResponseWriter to capture status code
type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Auth middleware for authentication (placeholder implementation)
func Auth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: Implement JWT token validation
			// For now, just pass through all requests
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimit middleware for rate limiting (placeholder implementation)
func RateLimit() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: Implement rate limiting logic
			// For now, just pass through all requests
			next.ServeHTTP(w, r)
		})
	}
}
