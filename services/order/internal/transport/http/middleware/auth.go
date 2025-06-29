package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/saan/order-service/pkg/logger"
)

// Role represents user roles in the system
type Role string

const (
	RoleSales      Role = "sales"
	RoleManager    Role = "manager"
	RoleAdmin      Role = "admin"
	RoleAIAssistant Role = "ai_assistant"
)

// User represents authenticated user information
type User struct {
	ID       string   `json:"id"`
	Email    string   `json:"email"`
	Name     string   `json:"name"`
	Role     Role     `json:"role"`
	Permissions []string `json:"permissions"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	AuthServiceURL string
	JWTSecret     string
	Logger        logger.Logger
}

// RequireRole creates middleware that requires specific roles
func RequireRole(config *AuthConfig, allowedRoles ...Role) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 1. Extract JWT token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			config.Logger.Warn("Missing Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Expected format: "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			config.Logger.Warn("Invalid Authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := tokenParts[1]

		// 2. Verify token with Auth Service
		user, err := verifyTokenWithAuthService(c.Request.Context(), config, token)
		if err != nil {
			config.Logger.Error("Token verification failed", "error", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// 3. Check role permissions
		if !hasRequiredRole(user.Role, allowedRoles) {
			config.Logger.Warn("Insufficient permissions", 
				"user_id", user.ID, 
				"user_role", user.Role, 
				"required_roles", allowedRoles)
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
				"required_roles": allowedRoles,
				"user_role": user.Role,
			})
			c.Abort()
			return
		}

		// 4. Set user context for downstream handlers
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("user_role", user.Role)

		config.Logger.Info("User authenticated successfully", 
			"user_id", user.ID, 
			"role", user.Role, 
			"endpoint", c.Request.URL.Path)

		c.Next()
	})
}

// RequirePermission creates middleware that requires specific permissions
func RequirePermission(config *AuthConfig, requiredPermissions ...string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Extract and verify token (same as above)
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			config.Logger.Warn("Missing Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			config.Logger.Warn("Invalid Authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := tokenParts[1]
		user, err := verifyTokenWithAuthService(c.Request.Context(), config, token)
		if err != nil {
			config.Logger.Error("Token verification failed", "error", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Check specific permissions
		if !hasRequiredPermissions(user.Permissions, requiredPermissions) {
			config.Logger.Warn("Insufficient permissions", 
				"user_id", user.ID, 
				"user_permissions", user.Permissions, 
				"required_permissions", requiredPermissions)
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
				"required_permissions": requiredPermissions,
				"user_permissions": user.Permissions,
			})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("user_role", user.Role)

		c.Next()
	})
}

// OptionalAuth provides optional authentication (sets user if token is valid, but doesn't require it)
func OptionalAuth(config *AuthConfig) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth header, continue without user context
			c.Next()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			// Invalid format, continue without user context
			c.Next()
			return
		}

		token := tokenParts[1]
		user, err := verifyTokenWithAuthService(c.Request.Context(), config, token)
		if err != nil {
			// Invalid token, continue without user context
			config.Logger.Warn("Optional auth token verification failed", "error", err)
			c.Next()
			return
		}

		// Set user context if token is valid
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("user_role", user.Role)

		c.Next()
	})
}

// verifyTokenWithAuthService verifies JWT token with the Auth Service
func verifyTokenWithAuthService(ctx context.Context, config *AuthConfig, token string) (*User, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Prepare request to Auth Service
	authServiceURL := config.AuthServiceURL
	if authServiceURL == "" {
		// Use service name as per PROJECT_RULES.md
		authServiceURL = "http://user-service:8088"
	}

	url := fmt.Sprintf("%s/api/auth/verify", authServiceURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set Authorization header with the token
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Make request to Auth Service
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token with auth service: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("invalid or expired token")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth service returned status %d", resp.StatusCode)
	}

	// Parse response
	var authResponse struct {
		User User `json:"user"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return nil, fmt.Errorf("failed to decode auth response: %w", err)
	}

	// Ensure user has default permissions based on role
	if len(authResponse.User.Permissions) == 0 {
		authResponse.User.Permissions = GetRolePermissions(authResponse.User.Role)
	}

	return &authResponse.User, nil
}

// hasRequiredRole checks if user role is in allowed roles
func hasRequiredRole(userRole Role, allowedRoles []Role) bool {
	for _, role := range allowedRoles {
		if userRole == role {
			return true
		}
	}
	return false
}

// hasRequiredPermissions checks if user has all required permissions
func hasRequiredPermissions(userPermissions []string, requiredPermissions []string) bool {
	userPermSet := make(map[string]bool)
	for _, perm := range userPermissions {
		userPermSet[perm] = true
	}

	for _, requiredPerm := range requiredPermissions {
		if !userPermSet[requiredPerm] {
			return false
		}
	}
	return true
}

// GetRolePermissions returns default permissions for each role
func GetRolePermissions(role Role) []string {
	switch role {
	case RoleSales:
		return []string{
			"orders:create",
			"orders:view",
			"customers:view",
		}
	case RoleManager:
		return []string{
			"orders:create",
			"orders:view",
			"orders:update",
			"orders:confirm",
			"orders:cancel",
			"orders:override_stock",
			"customers:view",
			"customers:update",
		}
	case RoleAdmin:
		return []string{
			"orders:*",
			"customers:*",
			"inventory:*",
			"reports:*",
			"admin:*",
		}
	case RoleAIAssistant:
		return []string{
			"orders:create_draft",
			"orders:view",
			"customers:view",
			"inventory:view",
		}
	default:
		return []string{}
	}
}
