package handlers

import (
	"net/http"
	"time"

	"gemini-cli-go/internal/auth"
	"gemini-cli-go/internal/config"
	"gemini-cli-go/internal/constants"
	"gemini-cli-go/internal/types"

	"github.com/gin-gonic/gin"
)

// DebugHandler handles debug-related requests
type DebugHandler struct {
	authManager *auth.AuthManager
	config      *config.Config
}

// NewDebugHandler creates a new debug handler
func NewDebugHandler(config *config.Config, authManager *auth.AuthManager) *DebugHandler {
	return &DebugHandler{
		authManager: authManager,
		config:      config,
	}
}

// CacheInfo handles GET /v1/debug/cache
func (h *DebugHandler) CacheInfo(c *gin.Context) {
	cacheInfo := h.authManager.GetCachedTokenInfo()
	
	response := map[string]interface{}{
		"cache_info": cacheInfo,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

// TokenTest handles POST /v1/token-test
func (h *DebugHandler) TokenTest(c *gin.Context) {
	err := h.authManager.TestAuthentication()
	
	if err != nil {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error:   constants.ErrAuthenticationFailed,
			Message: err.Error(),
			Code:    http.StatusUnauthorized,
		})
		return
	}

	response := map[string]interface{}{
		"status":        constants.MsgAuthenticationOK,
		"authenticated": true,
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
		"token_info":    h.authManager.GetCachedTokenInfo(),
	}

	c.JSON(http.StatusOK, response)
}

// FullTest handles POST /v1/test
func (h *DebugHandler) FullTest(c *gin.Context) {
	// Test authentication
	authErr := h.authManager.TestAuthentication()
	
	// Get authentication status
	authStatus := h.authManager.GetAuthenticationStatus()
	
	// Build response
	response := map[string]interface{}{
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"authentication": authStatus,
		"configuration": map[string]interface{}{
			"fake_thinking_enabled":       h.config.IsFakeThinkingEnabled(),
			"real_thinking_enabled":       h.config.IsRealThinkingEnabled(),
			"stream_thinking_as_content":  h.config.IsStreamThinkingAsContent(),
			"api_key_required":           h.config.IsAuthRequired(),
			"gemini_project_id_set":      h.config.GetGeminiProjectID() != "",
		},
		"cache_info": h.authManager.GetCachedTokenInfo(),
	}

	// Set status based on authentication result
	if authErr != nil {
		response["status"] = "failed"
		response["error"] = authErr.Error()
		c.JSON(http.StatusUnauthorized, response)
	} else {
		response["status"] = "success"
		c.JSON(http.StatusOK, response)
	}
}

// Health handles GET /health
func (h *DebugHandler) Health(c *gin.Context) {
	response := types.HealthResponse{
		Status:    constants.MsgHealthOK,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   constants.AppVersion,
	}

	c.JSON(http.StatusOK, response)
}

// ServiceInfo handles GET /
func (h *DebugHandler) ServiceInfo(c *gin.Context) {
	requiresAuth := h.config.IsAuthRequired()

	authType := "None"
	if requiresAuth {
		authType = "Bearer token in Authorization header"
	}

	response := types.ServiceInfoResponse{
		Name:        constants.AppName,
		Description: constants.AppDescription,
		Version:     constants.AppVersion,
		Authentication: types.AuthInfo{
			Required: requiresAuth,
			Type:     authType,
		},
		Endpoints: types.EndpointInfo{
			ChatCompletions: constants.PathV1 + constants.PathChatCompletions,
			Models:          constants.PathV1 + constants.PathModels,
			Debug: types.DebugRoutes{
				Cache:     constants.PathV1 + constants.PathDebugCache,
				TokenTest: constants.PathV1 + constants.PathTokenTest,
				FullTest:  constants.PathV1 + constants.PathTest,
			},
		},
		Documentation: constants.AppRepository,
	}

	c.JSON(http.StatusOK, response)
}

// ClearCache handles DELETE /v1/debug/cache
func (h *DebugHandler) ClearCache(c *gin.Context) {
	h.authManager.ClearTokenCache()
	
	response := map[string]interface{}{
		"status":    "cache cleared",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"message":   "Token cache has been cleared successfully",
	}

	c.JSON(http.StatusOK, response)
}

// RefreshToken handles POST /v1/debug/refresh
func (h *DebugHandler) RefreshToken(c *gin.Context) {
	// Clear existing cache
	h.authManager.ClearTokenCache()
	
	// Force re-initialization (which will refresh the token)
	err := h.authManager.InitializeAuth()
	if err != nil {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error:   constants.ErrTokenRefreshFailed,
			Message: err.Error(),
			Code:    http.StatusUnauthorized,
		})
		return
	}

	response := map[string]interface{}{
		"status":      constants.MsgTokenRefreshed,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"token_info":  h.authManager.GetCachedTokenInfo(),
	}

	c.JSON(http.StatusOK, response)
}

// SystemStatus handles GET /v1/debug/status
func (h *DebugHandler) SystemStatus(c *gin.Context) {
	// Check authentication
	authErr := h.authManager.TestAuthentication()
	isAuthenticated := authErr == nil

	// Get cache info
	cacheInfo := h.authManager.GetCachedTokenInfo()

	// Build system status
	response := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   constants.AppVersion,
		"system": map[string]interface{}{
			"authenticated":    isAuthenticated,
			"cache_healthy":    cacheInfo.Cached && !cacheInfo.IsExpired,
			"config_valid":     h.validateConfig(),
		},
		"authentication": map[string]interface{}{
			"status":     isAuthenticated,
			"error":      getErrorString(authErr),
			"cache_info": cacheInfo,
		},
		"configuration": map[string]interface{}{
			"fake_thinking":              h.config.IsFakeThinkingEnabled(),
			"real_thinking":              h.config.IsRealThinkingEnabled(),
			"stream_thinking_as_content": h.config.IsStreamThinkingAsContent(),
			"api_key_required":          h.config.IsAuthRequired(),
			"project_id_configured":     h.config.GetGeminiProjectID() != "",
			"gcp_credentials_provided":  h.config.GetGCPServiceAccount() != "",
		},
	}

	// Set appropriate status code
	if isAuthenticated {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusServiceUnavailable, response)
	}
}

// validateConfig validates the current configuration
func (h *DebugHandler) validateConfig() bool {
	// Check if required fields are set
	if h.config.GetGCPServiceAccount() == "" {
		return false
	}

	// Check if port is valid
	if h.config.GetAddress() == ":" {
		return false
	}

	return true
}

// getErrorString safely converts error to string
func getErrorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// Metrics handles GET /v1/debug/metrics (simple metrics)
func (h *DebugHandler) Metrics(c *gin.Context) {
	// Simple metrics - in a real implementation, you'd use prometheus or similar
	response := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"metrics": map[string]interface{}{
			"requests_total":     "not_implemented", // Would track total requests
			"errors_total":       "not_implemented", // Would track total errors
			"response_time_avg":  "not_implemented", // Would track average response time
			"cache_hits":         "not_implemented", // Would track cache hits
			"auth_failures":      "not_implemented", // Would track auth failures
		},
		"health": map[string]interface{}{
			"uptime":        "not_implemented", // Would track uptime
			"memory_usage":  "not_implemented", // Would track memory usage
			"cpu_usage":     "not_implemented", // Would track CPU usage
		},
	}

	c.JSON(http.StatusOK, response)
}
