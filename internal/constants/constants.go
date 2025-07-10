package constants

import "time"

// API endpoints and configuration
const (
	// Google Code Assist API Constants
	CodeAssistEndpoint   = "https://cloudcode-pa.googleapis.com"
	CodeAssistAPIVersion = "v1internal"

	// OAuth2 Configuration
	OAuthClientID     = "681255809395-oo8ft2oprdrnp9e3aqf6av3hmdib135j.apps.googleusercontent.com"
	OAuthClientSecret = "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"
	OAuthRefreshURL   = "https://oauth2.googleapis.com/token"

	// OpenAI API Constants
	OpenAIChatCompletionObject = "chat.completion.chunk"
	OpenAIModelOwner           = "google-gemini-cli"
	OpenAICompletionObject     = "chat.completion"
	OpenAIModelObject          = "model"
	OpenAIModelListObject      = "list"

	// Default values
	DefaultModel       = "gemini-2.5-flash"
	DefaultPort        = "8080"
	DefaultLogLevel    = "info"
	DefaultTemperature = 0.7

	// Token management
	TokenBufferTime         = 5 * time.Minute
	DefaultTokenCacheExpiry = 3600 // seconds
	DefaultRequestTimeout   = 30   // seconds

	// Thinking budget constants
	DefaultThinkingBudget  = -1 // -1 means dynamic allocation by Gemini
	DisabledThinkingBudget = 0  // 0 disables thinking entirely

	// Streaming constants
	ReasoningChunkDelay      = 100 * time.Millisecond
	ThinkingContentChunkSize = 15
	StreamingChunkDelay      = 50 * time.Millisecond
	SSEDataPrefix            = "data: "
	SSEDoneMessage           = "data: [DONE]"
	SSENewLine               = "\n"

	// HTTP headers
	ContentTypeJSON           = "application/json"
	ContentTypeSSE            = "text/event-stream"
	ContentTypeFormURLEncoded = "application/x-www-form-urlencoded"
	AuthorizationHeader       = "Authorization"
	BearerPrefix              = "Bearer "

	// CORS headers
	CORSAllowOrigin  = "*"
	CORSAllowMethods = "GET, POST, OPTIONS"
	CORSAllowHeaders = "Content-Type, Authorization"

	// Cache control
	CacheControlNoCache = "no-cache"
	ConnectionKeepAlive = "keep-alive"

	// Model context windows and limits
	MaxTokensDefault     = 65536
	ContextWindowDefault = 1048576

	// Environment variable names
	EnvGCPServiceAccount       = "GCP_SERVICE_ACCOUNT"
	EnvGeminiProjectID         = "GEMINI_PROJECT_ID"
	EnvOpenAIAPIKey            = "OPENAI_API_KEY"
	EnvEnableFakeThinking      = "ENABLE_FAKE_THINKING"
	EnvEnableRealThinking      = "ENABLE_REAL_THINKING"
	EnvStreamThinkingAsContent = "STREAM_THINKING_AS_CONTENT"
	EnvPort                    = "PORT"
	EnvLogLevel                = "LOG_LEVEL"
	EnvTokenCacheExpiry        = "TOKEN_CACHE_EXPIRY"
	EnvRequestTimeout          = "REQUEST_TIMEOUT"
	EnvSkipTLSVerify           = "SKIP_TLS_VERIFY"

	// API paths
	PathV1              = "/v1"
	PathModels          = "/models"
	PathChatCompletions = "/chat/completions"
	PathDebug           = "/debug"
	PathDebugCache      = "/debug/cache"
	PathTokenTest       = "/token-test"
	PathTest            = "/test"
	PathHealth          = "/health"
	PathRoot            = "/"

	// Error messages
	ErrMissingGCPServiceAccount = "GCP_SERVICE_ACCOUNT environment variable not set"
	ErrInvalidOAuth2Credentials = "Invalid OAuth2 credentials format"
	ErrAuthenticationFailed     = "Authentication failed"
	ErrTokenRefreshFailed       = "Token refresh failed"
	ErrProjectDiscoveryFailed   = "Project ID discovery failed"
	ErrModelNotFound            = "Model not found"
	ErrModelNotSupported        = "Model not supported"
	ErrInvalidImageFormat       = "Invalid image format"
	ErrMissingMessages          = "messages is a required field"
	ErrStreamRequestFailed      = "Stream request failed"
	ErrInvalidRequest           = "Invalid request"

	// Success messages
	MsgHealthOK          = "OK"
	MsgAuthenticationOK  = "Authentication successful"
	MsgTokenRefreshed    = "Token refreshed successfully"
	MsgProjectDiscovered = "Project ID discovered successfully"

	// Thinking tags
	ThinkingOpenTag  = "<thinking>\n"
	ThinkingCloseTag = "\n</thinking>\n\n"

	// Request/Response IDs
	ChatCompletionIDPrefix = "chatcmpl-"

	// Retry settings
	MaxRetries = 3
	RetryDelay = 1 * time.Second
)

// Static reasoning messages for thinking models
var ReasoningMessages = []string{
	"🔍 **Analyzing the request: \"{requestPreview}\"**\n\n",
	"🤔 Let me think about this step by step... ",
	"💭 I need to consider the context and provide a comprehensive response. ",
	"🎯 Based on my understanding, I should address the key points while being accurate and helpful. ",
	"✨ Let me formulate a clear and structured answer.\n\n",
}

// Supported image MIME types
var SupportedImageMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// HTTP status codes
const (
	StatusOK                  = 200
	StatusCreated             = 201
	StatusNoContent           = 204
	StatusBadRequest          = 400
	StatusUnauthorized        = 401
	StatusForbidden           = 403
	StatusNotFound            = 404
	StatusMethodNotAllowed    = 405
	StatusConflict            = 409
	StatusUnprocessableEntity = 422
	StatusInternalServerError = 500
	StatusBadGateway          = 502
	StatusServiceUnavailable  = 503
	StatusGatewayTimeout      = 504
)

// Log levels
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

// Application metadata
const (
	AppName        = "Gemini CLI OpenAI Go"
	AppDescription = "OpenAI-compatible API for Google Gemini models via OAuth"
	AppVersion     = "1.0.0"
	AppRepository  = "https://github.com/Linli1221/gemini-cli-go"
)
