package handlers

import (
	"gemini-cli-go/internal/auth"
	"gemini-cli-go/internal/config"
	"gemini-cli-go/internal/gemini"
	"gemini-cli-go/internal/models"
	"gemini-cli-go/internal/types"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// OpenAIHandler handles OpenAI-compatible requests
type OpenAIHandler struct {
	config       *config.Config
	authManager  *auth.AuthManager
	geminiClient *gemini.Client
}

// NewOpenAIHandler creates a new OpenAI handler
func NewOpenAIHandler(config *config.Config, authManager *auth.AuthManager, geminiClient *gemini.Client) *OpenAIHandler {
	return &OpenAIHandler{
		config:       config,
		authManager:  authManager,
		geminiClient: geminiClient,
	}
}

// ListModels handles GET /v1/models
func (h *OpenAIHandler) ListModels(c *gin.Context) {
	c.JSON(http.StatusOK, models.GetModelList())
}

// ChatCompletions handles POST /v1/chat/completions
func (h *OpenAIHandler) ChatCompletions(c *gin.Context) {
	var req types.ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stream := false
	if req.Stream != nil && *req.Stream {
		stream = true
	}

	if stream {
		h.streamChatCompletions(c, req)
	} else {
		h.getChatCompletion(c, req)
	}
}

func (h *OpenAIHandler) getChatCompletion(c *gin.Context, req types.ChatCompletionRequest) {
	result, err := h.geminiClient.GetCompletion(c.Request.Context(), req.Model, req.Messages, &gemini.StreamOptions{
		EnableFakeThinking: h.config.IsFakeThinkingEnabled(),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := types.ChatCompletionResponse{
		ID:      "chatcmpl-" + uuid.New().String(),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []types.ChatCompletionChoice{
			{
				Index: 0,
				Message: &types.ChatCompletionMessage{
					Role:    "assistant",
					Content: result.Content,
				},
				FinishReason: strPtr("stop"),
			},
		},
		Usage: &types.ChatCompletionUsage{
			PromptTokens:     result.Usage.InputTokens,
			CompletionTokens: result.Usage.OutputTokens,
			TotalTokens:      result.Usage.InputTokens + result.Usage.OutputTokens,
		},
	}

	c.JSON(http.StatusOK, response)
}

func (h *OpenAIHandler) streamChatCompletions(c *gin.Context, req types.ChatCompletionRequest) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	chunkChan, err := h.geminiClient.StreamContent(c.Request.Context(), req.Model, "", req.Messages, &gemini.StreamOptions{
		EnableFakeThinking: h.config.IsFakeThinkingEnabled(),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for chunk := range chunkChan {
		response := types.ChatCompletionResponse{
			ID:      "chatcmpl-" + uuid.New().String(),
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   req.Model,
			Choices: []types.ChatCompletionChoice{
				{
					Index: 0,
					Delta: &types.ChatCompletionDelta{
						Content: chunk.Data.(string),
					},
				},
			},
		}
		c.SSEEvent("data", response)
		c.Writer.Flush()
	}

	c.SSEEvent("data", "[DONE]")
	c.Writer.Flush()
}

func strPtr(s string) *string {
	return &s
}