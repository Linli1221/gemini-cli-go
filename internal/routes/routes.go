package routes

import (
	"time"

	"gemini-cli-go/internal/auth"
	"gemini-cli-go/internal/config"
	"gemini-cli-go/internal/constants"
	"gemini-cli-go/internal/gemini"
	"gemini-cli-go/internal/handlers"
	"gemini-cli-go/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all routes for the application
func SetupRoutes(cfg *config.Config) *gin.Engine {
	// Create Gin engine
	engine := gin.New()

	// Add global middleware
	engine.Use(middleware.LoggingMiddleware())
	engine.Use(middleware.RequestLoggingMiddleware(cfg))
	engine.Use(middleware.CORSMiddleware())
	engine.Use(middleware.SecurityHeadersMiddleware())
	engine.Use(middleware.RequestIDMiddleware())
	engine.Use(middleware.ErrorLoggingMiddleware())

	// Add rate limiting (100 requests per minute)
	engine.Use(middleware.BasicRateLimitMiddleware(100, time.Minute))

	// Create auth manager and Gemini client
	authManager := auth.NewAuthManager(&cfg.Environment)
	geminiClient := gemini.NewClient(&cfg.Environment, authManager)

	// Create handlers
	openaiHandler := handlers.NewOpenAIHandler(&cfg.Environment, authManager, geminiClient)
	debugHandler := handlers.NewDebugHandler(&cfg.Environment, authManager)

	// Root endpoint
	engine.GET(constants.PathRoot, debugHandler.ServiceInfo)

	// Health check endpoint
	engine.GET(constants.PathHealth, debugHandler.Health)

	// V1 API routes
	v1 := engine.Group(constants.PathV1)
	{
		// Apply authentication middleware to all v1 routes
		v1.Use(middleware.AuthMiddleware(cfg.Environment.OpenAIAPIKey))

		// OpenAI-compatible endpoints
		v1.GET(constants.PathModels, openaiHandler.ListModels)
		v1.POST(constants.PathChatCompletions, openaiHandler.ChatCompletions)

		// Debug endpoints
		debug := v1.Group(constants.PathDebug)
		{
			debug.GET("/cache", debugHandler.CacheInfo)
			debug.DELETE("/cache", debugHandler.ClearCache)
			debug.POST("/refresh", debugHandler.RefreshToken)
			debug.GET("/status", debugHandler.SystemStatus)
			debug.GET("/metrics", debugHandler.Metrics)
		}

		// Additional debug endpoints at v1 level for backward compatibility
		v1.POST(constants.PathTokenTest, debugHandler.TokenTest)
		v1.POST(constants.PathTest, debugHandler.FullTest)
	}

	return engine
}

// SetupTestRoutes sets up routes for testing (without rate limiting and with relaxed middleware)
func SetupTestRoutes(cfg *config.Config) *gin.Engine {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create Gin engine
	engine := gin.New()

	// Add minimal middleware for testing
	engine.Use(middleware.CORSMiddleware())
	engine.Use(middleware.RequestIDMiddleware())

	// Create auth manager and Gemini client
	authManager := auth.NewAuthManager(&cfg.Environment)
	geminiClient := gemini.NewClient(&cfg.Environment, authManager)

	// Create handlers
	openaiHandler := handlers.NewOpenAIHandler(&cfg.Environment, authManager, geminiClient)
	debugHandler := handlers.NewDebugHandler(&cfg.Environment, authManager)

	// Root endpoint
	engine.GET(constants.PathRoot, debugHandler.ServiceInfo)

	// Health check endpoint
	engine.GET(constants.PathHealth, debugHandler.Health)

	// V1 API routes
	v1 := engine.Group(constants.PathV1)
	{
		// Optional auth for testing
		v1.Use(middleware.OptionalAuthMiddleware(cfg.Environment.OpenAIAPIKey))

		// OpenAI-compatible endpoints
		v1.GET(constants.PathModels, openaiHandler.ListModels)
		v1.POST(constants.PathChatCompletions, openaiHandler.ChatCompletions)

		// Debug endpoints
		debug := v1.Group(constants.PathDebug)
		{
			debug.GET("/cache", debugHandler.CacheInfo)
			debug.DELETE("/cache", debugHandler.ClearCache)
			debug.POST("/refresh", debugHandler.RefreshToken)
			debug.GET("/status", debugHandler.SystemStatus)
			debug.GET("/metrics", debugHandler.Metrics)
		}

		// Additional debug endpoints
		v1.POST(constants.PathTokenTest, debugHandler.TokenTest)
		v1.POST(constants.PathTest, debugHandler.FullTest)
	}

	return engine
}

// SetupProductionRoutes sets up routes optimized for production
func SetupProductionRoutes(cfg *config.Config) *gin.Engine {
	// Set Gin to release mode
	gin.SetMode(gin.ReleaseMode)

	// Create Gin engine
	engine := gin.New()

	// Add production middleware
	engine.Use(middleware.LoggingMiddleware())
	engine.Use(middleware.CORSMiddleware())
	engine.Use(middleware.SecurityHeadersMiddleware())
	engine.Use(middleware.RequestIDMiddleware())
	engine.Use(middleware.ErrorLoggingMiddleware())

	// Stricter rate limiting for production (50 requests per minute)
	engine.Use(middleware.BasicRateLimitMiddleware(50, time.Minute))

	// Create auth manager and Gemini client
	authManager := auth.NewAuthManager(&cfg.Environment)
	geminiClient := gemini.NewClient(&cfg.Environment, authManager)

	// Create handlers
	openaiHandler := handlers.NewOpenAIHandler(&cfg.Environment, authManager, geminiClient)
	debugHandler := handlers.NewDebugHandler(&cfg.Environment, authManager)

	// Root endpoint
	engine.GET(constants.PathRoot, debugHandler.ServiceInfo)

	// Health check endpoint (no auth required)
	engine.GET(constants.PathHealth, debugHandler.Health)

	// V1 API routes
	v1 := engine.Group(constants.PathV1)
	{
		// Require authentication for all v1 routes in production
		v1.Use(middleware.RequireAuthMiddleware(cfg.Environment.OpenAIAPIKey))

		// OpenAI-compatible endpoints
		v1.GET(constants.PathModels, openaiHandler.ListModels)
		v1.POST(constants.PathChatCompletions, openaiHandler.ChatCompletions)

		// Limited debug endpoints in production
		debug := v1.Group(constants.PathDebug)
		{
			debug.GET("/status", debugHandler.SystemStatus)
			// Cache and metrics endpoints only available in production if explicitly enabled
			if cfg.Environment.LogLevel == constants.LogLevelDebug {
				debug.GET("/cache", debugHandler.CacheInfo)
				debug.GET("/metrics", debugHandler.Metrics)
			}
		}
	}

	return engine
}
