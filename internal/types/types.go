package types

import "time"

// Environment represents the application environment configuration
type Environment struct {
	GCPServiceAccount       string `json:"gcp_service_account"`
	GeminiProjectID         string `json:"gemini_project_id"`
	OpenAIAPIKey            string `json:"openai_api_key"`
	EnableFakeThinking      string `json:"enable_fake_thinking"`
	EnableRealThinking      string `json:"enable_real_thinking"`
	StreamThinkingAsContent string `json:"stream_thinking_as_content"`
	Port                    string `json:"port"`
	LogLevel                string `json:"log_level"`
	TokenCacheExpiry        int    `json:"token_cache_expiry"`
	RequestTimeout          int    `json:"request_timeout"`
	SkipTLSVerify           string `json:"skip_tls_verify"`
}

// OAuth2Credentials represents OAuth2 credentials from Gemini CLI
type OAuth2Credentials struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
	IDToken      string `json:"id_token"`
	ExpiryDate   int64  `json:"expiry_date"`
}

// ModelInfo represents information about a Gemini model
type ModelInfo struct {
	MaxTokens           int     `json:"max_tokens"`
	ContextWindow       int     `json:"context_window"`
	SupportsImages      bool    `json:"supports_images"`
	SupportsPromptCache bool    `json:"supports_prompt_cache"`
	InputPrice          float64 `json:"input_price"`
	OutputPrice         float64 `json:"output_price"`
	Description         string  `json:"description"`
	Thinking            bool    `json:"thinking"`
}

// ChatCompletionRequest represents an OpenAI chat completion request
type ChatCompletionRequest struct {
	Model          string        `json:"model"`
	Messages       []ChatMessage `json:"messages"`
	Stream         *bool         `json:"stream,omitempty"`
	ThinkingBudget *int          `json:"thinking_budget,omitempty"`
	Temperature    *float64      `json:"temperature,omitempty"`
	MaxTokens      *int          `json:"max_tokens,omitempty"`
	TopP           *float64      `json:"top_p,omitempty"`
	Stop           []string      `json:"stop,omitempty"`
}

// ChatMessage represents a message in a chat conversation
type ChatMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

// MessageContent represents structured content within a message
type MessageContent struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

// ImageURL represents image URL information
type ImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

// ChatCompletionResponse represents an OpenAI chat completion response
type ChatCompletionResponse struct {
	ID      string                   `json:"id"`
	Object  string                   `json:"object"`
	Created int64                    `json:"created"`
	Model   string                   `json:"model"`
	Choices []ChatCompletionChoice   `json:"choices"`
	Usage   *ChatCompletionUsage     `json:"usage,omitempty"`
}

// ChatCompletionChoice represents a choice in a chat completion response
type ChatCompletionChoice struct {
	Index        int                    `json:"index"`
	Message      *ChatCompletionMessage `json:"message,omitempty"`
	Delta        *ChatCompletionDelta   `json:"delta,omitempty"`
	FinishReason *string                `json:"finish_reason"`
}

// ChatCompletionMessage represents a message in a chat completion response
type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionDelta represents a delta in a streaming chat completion response
type ChatCompletionDelta struct {
	Role      string `json:"role,omitempty"`
	Content   string `json:"content,omitempty"`
	Reasoning string `json:"reasoning,omitempty"`
}

// ChatCompletionUsage represents usage information in a chat completion response
type ChatCompletionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// UsageData represents token usage information
type UsageData struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ReasoningData represents reasoning information
type ReasoningData struct {
	Reasoning string `json:"reasoning"`
}

// StreamChunk represents a chunk of streaming data
type StreamChunk struct {
	Type StreamChunkType `json:"type"`
	Data interface{}     `json:"data"`
}

// StreamChunkType represents the type of stream chunk
type StreamChunkType string

const (
	StreamChunkTypeText          StreamChunkType = "text"
	StreamChunkTypeUsage         StreamChunkType = "usage"
	StreamChunkTypeReasoning     StreamChunkType = "reasoning"
	StreamChunkTypeThinkingContent StreamChunkType = "thinking_content"
	StreamChunkTypeRealThinking  StreamChunkType = "real_thinking"
)

// TokenRefreshResponse represents a token refresh response
type TokenRefreshResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// CachedTokenData represents cached token data
type CachedTokenData struct {
	AccessToken string    `json:"access_token"`
	ExpiryDate  time.Time `json:"expiry_date"`
	CachedAt    time.Time `json:"cached_at"`
}

// TokenCacheInfo represents token cache information
type TokenCacheInfo struct {
	Cached               bool      `json:"cached"`
	CachedAt             time.Time `json:"cached_at,omitempty"`
	ExpiresAt            time.Time `json:"expires_at,omitempty"`
	TimeUntilExpirySeconds int     `json:"time_until_expiry_seconds,omitempty"`
	IsExpired            bool      `json:"is_expired,omitempty"`
	Message              string    `json:"message,omitempty"`
	Error                string    `json:"error,omitempty"`
}

// GeminiResponse represents a response from the Gemini API
type GeminiResponse struct {
	Response *struct {
		Candidates    []GeminiCandidate    `json:"candidates"`
		UsageMetadata *GeminiUsageMetadata `json:"usageMetadata"`
	} `json:"response"`
}

// GeminiCandidate represents a candidate in a Gemini response
type GeminiCandidate struct {
	Content *struct {
		Parts []GeminiPart `json:"parts"`
	} `json:"content"`
}

// GeminiPart represents a part in a Gemini candidate
type GeminiPart struct {
	Text       string `json:"text,omitempty"`
	Thought    bool   `json:"thought,omitempty"`
	InlineData *struct {
		MimeType string `json:"mimeType"`
		Data     string `json:"data"`
	} `json:"inlineData,omitempty"`
	FileData *struct {
		MimeType string `json:"mimeType"`
		FileURI  string `json:"fileUri"`
	} `json:"fileData,omitempty"`
}

// GeminiUsageMetadata represents usage metadata in a Gemini response
type GeminiUsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
}

// GeminiFormattedMessage represents a message formatted for Gemini API
type GeminiFormattedMessage struct {
	Role  string       `json:"role"`
	Parts []GeminiPart `json:"parts"`
}

// ProjectDiscoveryResponse represents a project discovery response
type ProjectDiscoveryResponse struct {
	CloudAICompanionProject string `json:"cloudaicompanionProject"`
}

// ErrorResponse represents a generic error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

// ModelListResponse represents a model list response
type ModelListResponse struct {
	Object string      `json:"object"`
	Data   []ModelItem `json:"data"`
}

// ModelItem represents a model item in a model list response
type ModelItem struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version,omitempty"`
}

// ServiceInfo represents service information
type ServiceInfo struct {
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Version        string                 `json:"version"`
	Authentication map[string]interface{} `json:"authentication"`
	Endpoints      map[string]interface{} `json:"endpoints"`
	Documentation  string                 `json:"documentation"`
}

// ServiceInfoResponse represents the response for the service info endpoint
type ServiceInfoResponse struct {
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Version        string                 `json:"version"`
	Authentication AuthInfo               `json:"authentication"`
	Endpoints      EndpointInfo           `json:"endpoints"`
	Documentation  string                 `json:"documentation"`
}

// AuthInfo represents authentication information
type AuthInfo struct {
	Required bool   `json:"required"`
	Type     string `json:"type"`
}

// EndpointInfo represents endpoint information
type EndpointInfo struct {
	ChatCompletions string      `json:"chat_completions"`
	Models          string      `json:"models"`
	Debug           DebugRoutes `json:"debug"`
}

// DebugRoutes represents debug routes
type DebugRoutes struct {
	Cache     string `json:"cache"`
	TokenTest string `json:"token_test"`
	FullTest  string `json:"full_test"`
}