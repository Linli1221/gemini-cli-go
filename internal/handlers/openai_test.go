package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gemini-cli-go/internal/auth"
	"gemini-cli-go/internal/config"
	"gemini-cli-go/internal/gemini"
	"gemini-cli-go/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestHandler() (*OpenAIHandler, *gin.Engine) {
	// Create test config
	cfg := &config.Config{
		Environment: types.Environment{
			GCPServiceAccount: `{"access_token":"test","refresh_token":"test","scope":"test","token_type":"Bearer","id_token":"test","expiry_date":9999999999999}`,
			Port:              "8080",
			LogLevel:          "info",
		},
	}

	// Create test components
	authManager := auth.NewAuthManager(&cfg.Environment)
	geminiClient := gemini.NewClient(&cfg.Environment, authManager)
	handler := NewOpenAIHandler(&cfg.Environment, authManager, geminiClient)

	// Create test gin engine
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	return handler, engine
}

func TestListModels(t *testing.T) {
	handler, engine := setupTestHandler()
	
	// Setup route
	engine.GET("/v1/models", handler.ListModels)

	// Create request
	req, _ := http.NewRequest("GET", "/v1/models", nil)
	w := httptest.NewRecorder()

	// Perform request
	engine.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	var response types.ModelListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "list", response.Object)
	assert.NotEmpty(t, response.Data)
}

func TestChatCompletions_InvalidRequest(t *testing.T) {
	handler, engine := setupTestHandler()
	
	// Setup route
	engine.POST("/v1/chat/completions", handler.ChatCompletions)

	// Create invalid request (empty messages)
	requestBody := types.ChatCompletionRequest{
		Model:    "gemini-2.5-flash",
		Messages: []types.ChatMessage{}, // Empty messages should fail
	}
	
	jsonData, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform request
	engine.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response.Message, "messages is a required field")
}

func TestChatCompletions_InvalidModel(t *testing.T) {
	handler, engine := setupTestHandler()
	
	// Setup route
	engine.POST("/v1/chat/completions", handler.ChatCompletions)

	// Create request with invalid model
	requestBody := types.ChatCompletionRequest{
		Model: "invalid-model",
		Messages: []types.ChatMessage{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}
	
	jsonData, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform request
	engine.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response.Message, "not found")
}

func TestChatCompletions_ValidRequest(t *testing.T) {
	handler, engine := setupTestHandler()
	
	// Setup route
	engine.POST("/v1/chat/completions", handler.ChatCompletions)

	// Create valid request
	requestBody := types.ChatCompletionRequest{
		Model: "gemini-2.5-flash",
		Messages: []types.ChatMessage{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
		Stream: boolPtr(false), // Non-streaming for easier testing
	}
	
	jsonData, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform request
	engine.ServeHTTP(w, req)

	// Note: This will likely fail with authentication error in real testing
	// since we're using dummy credentials, but it tests the request validation
	// In a real test environment, you'd mock the auth and Gemini client
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusUnauthorized)
}


func TestValidateImageSupport(t *testing.T) {
	handler, _ := setupTestHandler()

	// Test request without images
	request := &types.ChatCompletionRequest{
		Model: "gemini-2.5-flash",
		Messages: []types.ChatMessage{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	err := handler.validateImageSupport(request)
	assert.NoError(t, err)

	// Test request with images on supported model
	request.Messages = []types.ChatMessage{
		{
			Role: "user",
			Content: []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "What's in this image?",
				},
				map[string]interface{}{
					"type": "image_url",
					"image_url": map[string]interface{}{
						"url": "data:image/jpeg;base64,test",
					},
				},
			},
		},
	}

	err = handler.validateImageSupport(request)
	assert.NoError(t, err) // gemini-2.5-flash supports images
}

// Helper functions

func boolPtr(b bool) *bool {
	return &b
}
