package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gemini-cli-go/internal/config"
	"gemini-cli-go/internal/constants"
	"gemini-cli-go/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// Set log level based on config
	setupLogging(cfg.GetLogLevel())

	// Print startup information
	printStartupInfo(cfg)

	// Setup routes based on environment
	var engine *gin.Engine
	if isProduction() {
		engine = routes.SetupProductionRoutes(cfg)
	} else {
		engine = routes.SetupRoutes(cfg)
	}

	// Create HTTP server
	server := &http.Server{
		Addr:           cfg.GetAddress(),
		Handler:        engine,
		ReadTimeout:    time.Duration(cfg.GetRequestTimeout()) * time.Second,
		WriteTimeout:   time.Duration(cfg.GetRequestTimeout()) * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 Starting %s on %s", constants.AppName, cfg.GetAddress())
		log.Printf("📖 Documentation: %s", constants.AppRepository)
		
		if cfg.IsAuthRequired() {
			log.Printf("🔐 Authentication: Required (API key configured)")
		} else {
			log.Printf("🔓 Authentication: Not required (no API key configured)")
		}

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Gracefully shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("❌ Server forced to shutdown: %v", err)
	} else {
		log.Println("✅ Server shutdown complete")
	}
}

// setupLogging configures logging based on the specified level
func setupLogging(level string) {
	switch level {
	case constants.LogLevelDebug:
		gin.SetMode(gin.DebugMode)
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	case constants.LogLevelInfo:
		gin.SetMode(gin.ReleaseMode)
		log.SetFlags(log.LstdFlags)
	case constants.LogLevelWarn, constants.LogLevelError:
		gin.SetMode(gin.ReleaseMode)
		log.SetFlags(log.LstdFlags)
	default:
		gin.SetMode(gin.ReleaseMode)
		log.SetFlags(log.LstdFlags)
	}
}

// isProduction checks if the application is running in production mode
func isProduction() bool {
	env := os.Getenv("GIN_MODE")
	return env == "release" || env == "production"
}

// printStartupInfo prints startup information
func printStartupInfo(cfg *config.Config) {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Printf("║  %s%-50s  ║\n", constants.AppName, "")
	fmt.Printf("║  Version: %-47s  ║\n", constants.AppVersion)
	fmt.Printf("║  Description: %-43s  ║\n", constants.AppDescription)
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	fmt.Printf("║  🌐 Server Address: %-37s  ║\n", cfg.GetAddress())
	fmt.Printf("║  🔧 Log Level: %-42s  ║\n", cfg.GetLogLevel())
	
	if cfg.IsAuthRequired() {
		fmt.Printf("║  🔐 Authentication: %-37s  ║\n", "Required")
	} else {
		fmt.Printf("║  🔓 Authentication: %-37s  ║\n", "Not Required")
	}
	
	if cfg.IsFakeThinkingEnabled() {
		fmt.Printf("║  🤔 Fake Thinking: %-38s  ║\n", "Enabled")
	}
	
	if cfg.IsRealThinkingEnabled() {
		fmt.Printf("║  🧠 Real Thinking: %-38s  ║\n", "Enabled")
	}
	
	if cfg.IsStreamThinkingAsContent() {
		fmt.Printf("║  📝 Stream Thinking as Content: %-25s  ║\n", "Enabled")
	}
	
	if cfg.GetGeminiProjectID() != "" {
		fmt.Printf("║  📋 Project ID: %-39s  ║\n", "Configured")
	}

	fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	fmt.Printf("║  📚 API endpoints available at: %-28s  ║\n", cfg.GetAddress())
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

// handlePanic recovers from panics and logs them
func handlePanic() {
	if r := recover(); r != nil {
		log.Printf("❌ Panic recovered: %v", r)
	}
}

// version command handler
func printVersion() {
	fmt.Printf("%s %s\n", constants.AppName, constants.AppVersion)
	fmt.Printf("Description: %s\n", constants.AppDescription)
	fmt.Printf("Repository: %s\n", constants.AppRepository)
	os.Exit(0)
}

// help command handler
func printHelp() {
	fmt.Printf("Usage: %s [options]\n\n", os.Args[0])
	fmt.Println("Options:")
	fmt.Println("  -h, --help     Show this help message")
	fmt.Println("  -v, --version  Show version information")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Printf("  %-25s %s\n", "GCP_SERVICE_ACCOUNT", "OAuth2 credentials JSON (required)")
	fmt.Printf("  %-25s %s\n", "GEMINI_PROJECT_ID", "Google Cloud Project ID (optional)")
	fmt.Printf("  %-25s %s\n", "OPENAI_API_KEY", "API key for authentication (optional)")
	fmt.Printf("  %-25s %s\n", "ENABLE_FAKE_THINKING", "Enable fake thinking output (optional)")
	fmt.Printf("  %-25s %s\n", "ENABLE_REAL_THINKING", "Enable real thinking output (optional)")
	fmt.Printf("  %-25s %s\n", "STREAM_THINKING_AS_CONTENT", "Stream thinking as content (optional)")
	fmt.Printf("  %-25s %s\n", "PORT", "Server port (default: 8080)")
	fmt.Printf("  %-25s %s\n", "LOG_LEVEL", "Log level (debug, info, warn, error)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Start with environment file")
	fmt.Println("  export $(cat .env | xargs) && ./gemini-cli-go")
	fmt.Println()
	fmt.Println("  # Start with inline environment variables")
	fmt.Println("  GCP_SERVICE_ACCOUNT='...' PORT=3000 ./gemini-cli-go")
	fmt.Println()
	fmt.Printf("Documentation: %s\n", constants.AppRepository)
	os.Exit(0)
}

func init() {
	// Handle command line arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h", "--help":
			printHelp()
		case "-v", "--version":
			printVersion()
		}
	}

	// Set up panic recovery
	defer handlePanic()
}