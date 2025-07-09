package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gemini-cli-go/internal/auth"
	"gemini-cli-go/internal/constants"
	"gemini-cli-go/internal/gemini"
	"gemini-cli-go/internal/models"
	"gemini-cli-go/internal/stream"
	"gemini-cli-go/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// OpenAIHandler handles OpenAI-compatible API requests
type OpenAIHandler struct {
	authManager   *auth.AuthManager
	geminiClient  *gemini.Client
	config        *types.Environment
}

// NewOpenAIHandler creates a new OpenAI handler
func NewOpenAIHandler(config *types.Environment, authManager *auth.AuthManager, geminiClient *gemini.Client) *OpenAIHandler {
	return &OpenAIHandler{
		authManager:  authManager,
		geminiClient: geminiClient,
		config:       config,
	}
}

// ListModels handles GET /v1/models
func (h *OpenAIHandler) ListModels(c *gin.Context) {
	modelIDs := models.GetAllModelIDs()
	modelData := make([]types.ModelItem, 0, len(modelIDs))

	for _, modelID := range modelIDs {
		modelData = append(modelData, types.ModelItem{
			ID:      modelID,
			Object:  constants.OpenAIModelObject,
			Created: time.Now().Unix(),
			OwnedBy: constants.OpenAIModelOwner,
		})
	}

	response := types.ModelListResponse{
		Object: constants.OpenAIModelListObject,
		Data:   modelData,
	}

	c.JSON(http.StatusOK, response)
}

// ChatCompletions handles POST /v1/chat/completions
func (h *OpenAIHandler) ChatCompletions(c *gin.Context) {
	var request types.ChatCompletionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   constants.ErrInvalidRequest,
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Validate request
	if err := h.validateChatCompletionRequest(&request); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   constants.ErrInvalidRequest,
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Set defaults
	h.setRequestDefaults(&request)

	// Check if model supports images
	if err := h.validateImageSupport(&request); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   constants.ErrModelNotSupported,
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Prepare stream options
	streamOptions := h.createStreamOptions(&request)

	// The geminiClient will handle system prompts internally.

	// Test authentication
	if err := h.authManager.TestAuthentication(); err != nil {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error:   constants.ErrAuthenticationFailed,
			Message: err.Error(),
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Handle streaming vs non-streaming
	if request.Stream != nil && *request.Stream {
		h.handleStreamingResponse(c, &request, streamOptions)
	} else {
		h.handleNonStreamingResponse(c, &request, streamOptions)
	}
}

// validateChatCompletionRequest validates the chat completion request
func (h *OpenAIHandler) validateChatCompletionRequest(request *types.ChatCompletionRequest) error {
	if len(request.Messages) == 0 {
		return fmt.Errorf(constants.ErrMissingMessages)
	}

	if !models.IsValidModel(request.Model) {
		availableModels := models.GetAllModelIDs()
		return fmt.Errorf("model '%s' not found. Available models: %v", request.Model, availableModels)
	}

	return nil
}

// setRequestDefaults sets default values for the request
func (h *OpenAIHandler) setRequestDefaults(request *types.ChatCompletionRequest) {
	if request.Model == "" {
		request.Model = models.GetDefaultModel()
	}

	if request.Stream == nil {
		defaultStream := false
		request.Stream = &defaultStream
	}

	if request.Temperature == nil {
		defaultTemp := constants.DefaultTemperature
		request.Temperature = &defaultTemp
	}
}

// validateImageSupport validates image support for the request
func (h *OpenAIHandler) validateImageSupport(request *types.ChatCompletionRequest) error {
	hasImages := false
	for _, msg := range request.Messages {
		if contentArray, ok := msg.Content.([]interface{}); ok {
			for _, content := range contentArray {
				if contentMap, ok := content.(map[string]interface{}); ok {
					if contentType, ok := contentMap["type"].(string); ok && contentType == "image_url" {
						hasImages = true
						break
					}
				}
			}
		}
		if hasImages {
			break
		}
	}

	if hasImages && !models.SupportsImages(request.Model) {
		return fmt.Errorf("model '%s' does not support image inputs. Please use a vision-capable model", request.Model)
	}

	return nil
}

// createStreamOptions creates stream options from the request
func (h *OpenAIHandler) createStreamOptions(request *types.ChatCompletionRequest) *gemini.StreamOptions {
	options := &gemini.StreamOptions{
		Temperature:             request.Temperature,
		MaxTokens:               request.MaxTokens,
		TopP:                    request.TopP,
		ThinkingBudget:          request.ThinkingBudget,
		EnableRealThinking:      h.config.EnableRealThinking == "true",
		EnableFakeThinking:      h.config.EnableFakeThinking == "true",
		StreamThinkingAsContent: h.config.StreamThinkingAsContent == "true",
		IncludeReasoning:        h.config.EnableRealThinking == "true",
	}

	return options
}


// handleStreamingResponse handles streaming responses
func (h *OpenAIHandler) handleStreamingResponse(c *gin.Context, request *types.ChatCompletionRequest, options *gemini.StreamOptions) {
	// Set streaming headers
	c.Header("Content-Type", constants.ContentTypeSSE)
	c.Header("Cache-Control", constants.CacheControlNoCache)
	c.Header("Connection", constants.ConnectionKeepAlive)
	c.Status(http.StatusOK)

	// Create stream writer
	writer := &GinResponseWriter{Context: c}
	streamWriter := stream.NewStreamWriter(writer, request.Model)
	streamWriter.WriteHeaders()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(h.config.RequestTimeout)*time.Second)
	defer cancel()

	// Get stream from Gemini client
	chunkChan, err := h.geminiClient.StreamContent(ctx, request.Model, "", request.Messages, options)
	if err != nil {
		errorChunk := types.StreamChunk{
			Type: types.StreamChunkTypeText,
			Data: fmt.Sprintf("Error: %v", err),
		}
		streamWriter.WriteChunk(errorChunk)
		streamWriter.WriteFinalChunk()
		return
	}

	// Stream chunks to client
	for chunk := range chunkChan {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := streamWriter.WriteChunk(chunk); err != nil {
			return
		}
	}

	// Write final chunk
	streamWriter.WriteFinalChunk()
}

// handleNonStreamingResponse handles non-streaming responses
func (h *OpenAIHandler) handleNonStreamingResponse(c *gin.Context, request *types.ChatCompletionRequest, options *gemini.StreamOptions) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(h.config.RequestTimeout)*time.Second)
	defer cancel()

	// Get completion from Gemini client
	result, err := h.geminiClient.GetCompletion(ctx, request.Model, request.Messages, options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "Completion failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Create response
	response := types.ChatCompletionResponse{
		ID:      constants.ChatCompletionIDPrefix + strings.ReplaceAll(uuid.New().String(), "-", ""),
		Object:  constants.OpenAICompletionObject,
		Created: time.Now().Unix(),
		Model:   request.Model,
		Choices: []types.ChatCompletionChoice{
			{
				Index: 0,
				Message: &types.ChatCompletionMessage{
					Role:    "assistant",
					Content: result.Content,
				},
				FinishReason: stringPtr("stop"),
			},
		},
	}

	// Add usage information if available
	if result.Usage != nil {
		response.Usage = &types.ChatCompletionUsage{
			PromptTokens:     result.Usage.InputTokens,
			CompletionTokens: result.Usage.OutputTokens,
			TotalTokens:      result.Usage.InputTokens + result.Usage.OutputTokens,
		}
	}

	c.JSON(http.StatusOK, response)
}

// GinResponseWriter adapts gin.Context to stream.ResponseWriter
type GinResponseWriter struct {
	Context *gin.Context
}

func (w *GinResponseWriter) Write(data []byte) (int, error) {
	return w.Context.Writer.Write(data)
}

func (w *GinResponseWriter) Flush() {
	w.Context.Writer.Flush()
}

func (w *GinResponseWriter) Header() map[string]string {
	headers := make(map[string]string)
	for key, values := range w.Context.Writer.Header() {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	return headers
}

func (w *GinResponseWriter) SetHeader(key, value string) {
	w.Context.Header(key, value)
}

func (w *GinResponseWriter) WriteHeader(statusCode int) {
	w.Context.Status(statusCode)
}

// Helper functions

func stringPtr(s string) *string {
	return &s
}