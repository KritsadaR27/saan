package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/saan/order-service/internal/transport/http/middleware"
	"github.com/saan/order-service/pkg/logger"
)

// TestRBACMiddleware tests Role-Based Access Control middleware
func TestRBACMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.NewLogger("debug", "json")

	// Create test auth config
	authConfig := &middleware.AuthConfig{
		AuthServiceURL: "http://user-service:8088",
		JWTSecret:      "test-secret",
		Logger:         log,
	}

	tests := []struct {
		name             string
		token            string
		requiredRoles    []middleware.Role
		expectedStatus   int
		expectedMessage  string
	}{
		{
			name:           "Missing Authorization Header",
			token:          "",
			requiredRoles:  []middleware.Role{middleware.RoleSales},
			expectedStatus: http.StatusUnauthorized,
			expectedMessage: "Authorization header required",
		},
		{
			name:           "Invalid Authorization Header Format",
			token:          "InvalidToken",
			requiredRoles:  []middleware.Role{middleware.RoleSales},
			expectedStatus: http.StatusUnauthorized,
			expectedMessage: "Invalid authorization header format",
		},
		{
			name:           "Valid Bearer Token Format",
			token:          "Bearer valid-jwt-token",
			requiredRoles:  []middleware.Role{middleware.RoleSales},
			expectedStatus: http.StatusOK, // Will fail at auth service verification in real scenario
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			
			// Add RBAC middleware
			router.Use(middleware.RequireRole(authConfig, tt.requiredRoles...))
			
			// Add test endpoint
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}
			
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
			
			if tt.expectedMessage != "" {
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedMessage)
			}
		})
	}
}

// TestRolePermissions tests role permission mapping
func TestRolePermissions(t *testing.T) {
	tests := []struct {
		role                middleware.Role
		expectedPermissions []string
	}{
		{
			role: middleware.RoleSales,
			expectedPermissions: []string{
				"orders:create",
				"orders:view",
				"customers:view",
			},
		},
		{
			role: middleware.RoleManager,
			expectedPermissions: []string{
				"orders:create",
				"orders:view", 
				"orders:update",
				"orders:confirm",
				"orders:cancel",
				"orders:override_stock",
				"customers:view",
				"customers:update",
			},
		},
		{
			role: middleware.RoleAdmin,
			expectedPermissions: []string{
				"orders:*",
				"customers:*",
				"inventory:*",
				"reports:*",
				"admin:*",
			},
		},
		{
			role: middleware.RoleAIAssistant,
			expectedPermissions: []string{
				"orders:create_draft",
				"orders:view",
				"customers:view",
				"inventory:view",
			},
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			permissions := middleware.GetRolePermissions(tt.role)
			assert.Equal(t, tt.expectedPermissions, permissions)
		})
	}
}

// TestAdminEndpointsProtection tests that admin endpoints are properly protected
func TestAdminEndpointsProtection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.NewLogger("debug", "json")

	authConfig := &middleware.AuthConfig{
		AuthServiceURL: "http://user-service:8088",
		JWTSecret:      "test-secret",
		Logger:         log,
	}

	router := gin.New()
	
	// Add admin routes with RBAC protection
	admin := router.Group("/admin")
	admin.Use(middleware.RequireRole(authConfig, middleware.RoleAdmin))
	{
		admin.POST("/orders", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "order created"})
		})
		admin.POST("/orders/bulk-status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "bulk update success"})
		})
		admin.GET("/orders/export", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "export ready"})
		})
	}

	adminEndpoints := []string{
		"/admin/orders",
		"/admin/orders/bulk-status", 
		"/admin/orders/export",
	}

	for _, endpoint := range adminEndpoints {
		t.Run("Unauthorized access to "+endpoint, func(t *testing.T) {
			var req *http.Request
			if endpoint == "/admin/orders" || endpoint == "/admin/orders/bulk-status" {
				req = httptest.NewRequest("POST", endpoint, bytes.NewBufferString("{}"))
			} else {
				req = httptest.NewRequest("GET", endpoint, nil)
			}
			
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			// Should be unauthorized without proper token
			assert.Equal(t, http.StatusUnauthorized, recorder.Code)
		})
	}
}
