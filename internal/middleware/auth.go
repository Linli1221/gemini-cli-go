package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"gemini-cli-go/internal/constants"
	"gemini-cli-go/internal/types"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware handles API key authentication
func AuthMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication if no API key is configured
		if apiKey == "" {
			c.Next()
			return
		}

		// Get authorization header
		authHeader := c.GetHeader(constants.AuthorizationHeader)
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, types.ErrorResponse{
				Error:   "Missing authorization header",
				Message: "Please provide the Authorization header with your API key",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Check if it starts with "Bearer "
		if !strings.HasPrefix(authHeader, constants.BearerPrefix) {
			c.JSON(http.StatusUnauthorized, types.ErrorResponse{
				Error:   "Invalid authorization header format",
				Message: "Authorization header must start with 'Bearer '",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Extract the token
		token := strings.TrimPrefix(authHeader, constants.BearerPrefix)
		if token == "" {
			c.JSON(http.StatusUnauthorized, types.ErrorResponse{
				Error:   "Missing API key",
				Message: "Please provide a valid API key in the Authorization header",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Validate the token
		if token != apiKey {
			c.JSON(http.StatusUnauthorized, types.ErrorResponse{
				Error:   "Invalid API key",
				Message: "The provided API key is invalid",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Token is valid, continue
		c.Next()
	}
}

// OptionalAuthMiddleware handles optional API key authentication
func OptionalAuthMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// If no API key is configured, skip authentication
		if apiKey == "" {
			c.Next()
			return
		}

		// Get authorization header
		authHeader := c.GetHeader(constants.AuthorizationHeader)
		if authHeader == "" {
			// No auth header provided, but auth is optional
			c.Next()
			return
		}

		// If auth header is provided, validate it
		if !strings.HasPrefix(authHeader, constants.BearerPrefix) {
			c.JSON(http.StatusUnauthorized, types.ErrorResponse{
				Error:   "Invalid authorization header format",
				Message: "Authorization header must start with 'Bearer '",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, constants.BearerPrefix)
		if token != apiKey {
			c.JSON(http.StatusUnauthorized, types.ErrorResponse{
				Error:   "Invalid API key",
				Message: "The provided API key is invalid",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Mark as authenticated
		c.Set("authenticated", true)
		c.Next()
	}
}

// IsAuthenticated checks if the request is authenticated
func IsAuthenticated(c *gin.Context) bool {
	authenticated, exists := c.Get("authenticated")
	if !exists {
		return false
	}
	return authenticated.(bool)
}

// RequireAuthMiddleware requires authentication for specific endpoints
func RequireAuthMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if apiKey == "" {
			// No API key configured, allow access
			c.Next()
			return
		}

		// Get authorization header
		authHeader := c.GetHeader(constants.AuthorizationHeader)
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, types.ErrorResponse{
				Error:   "Authentication required",
				Message: "This endpoint requires authentication. Please provide an API key.",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Validate authorization header
		if !strings.HasPrefix(authHeader, constants.BearerPrefix) {
			c.JSON(http.StatusUnauthorized, types.ErrorResponse{
				Error:   "Invalid authorization header format",
				Message: "Authorization header must start with 'Bearer '",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, constants.BearerPrefix)
		if token != apiKey {
			c.JSON(http.StatusUnauthorized, types.ErrorResponse{
				Error:   "Invalid API key",
				Message: "The provided API key is invalid",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// APIKeyValidator validates API key format
type APIKeyValidator struct {
	minLength int
	prefix    string
}

// NewAPIKeyValidator creates a new API key validator
func NewAPIKeyValidator() *APIKeyValidator {
	return &APIKeyValidator{
		minLength: 10,
		prefix:    "sk-",
	}
}

// Validate validates an API key format
func (v *APIKeyValidator) Validate(apiKey string) error {
	if len(apiKey) < v.minLength {
		return fmt.Errorf("API key must be at least %d characters long", v.minLength)
	}

	if v.prefix != "" && !strings.HasPrefix(apiKey, v.prefix) {
		return fmt.Errorf("API key must start with '%s'", v.prefix)
	}

	return nil
}

// AuthContext holds authentication context
type AuthContext struct {
	IsAuthenticated bool   `json:"is_authenticated"`
	APIKeyProvided  bool   `json:"api_key_provided"`
	UserID          string `json:"user_id,omitempty"`
}

// GetAuthContext extracts authentication context from gin context
func GetAuthContext(c *gin.Context) *AuthContext {
	ctx := &AuthContext{}
	
	if auth, exists := c.Get("authenticated"); exists {
		ctx.IsAuthenticated = auth.(bool)
	}

	authHeader := c.GetHeader(constants.AuthorizationHeader)
	ctx.APIKeyProvided = authHeader != ""

	if userID, exists := c.Get("user_id"); exists {
		ctx.UserID = userID.(string)
	}

	return ctx
}

// SetUserID sets the user ID in the context
func SetUserID(c *gin.Context, userID string) {
	c.Set("user_id", userID)
}

// GetUserID gets the user ID from the context
func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		return userID.(string)
	}
	return ""
}