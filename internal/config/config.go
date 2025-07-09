package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gemini-cli-go/internal/constants"
	"gemini-cli-go/internal/types"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	Environment types.Environment
}

// New creates a new configuration instance
func New() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	config := &Config{
		Environment: types.Environment{
			GCPServiceAccount:       getEnv(constants.EnvGCPServiceAccount, ""),
			GeminiProjectID:         getEnv(constants.EnvGeminiProjectID, ""),
			OpenAIAPIKey:            getEnv(constants.EnvOpenAIAPIKey, ""),
			EnableFakeThinking:      getEnv(constants.EnvEnableFakeThinking, "false"),
			EnableRealThinking:      getEnv(constants.EnvEnableRealThinking, "false"),
			StreamThinkingAsContent: getEnv(constants.EnvStreamThinkingAsContent, "false"),
			Port:                    getEnv(constants.EnvPort, constants.DefaultPort),
			LogLevel:                getEnv(constants.EnvLogLevel, constants.DefaultLogLevel),
			TokenCacheExpiry:        getEnvAsInt(constants.EnvTokenCacheExpiry, constants.DefaultTokenCacheExpiry),
			RequestTimeout:          getEnvAsInt(constants.EnvRequestTimeout, constants.DefaultRequestTimeout),
		},
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Environment.GCPServiceAccount == "" {
		return fmt.Errorf(constants.ErrMissingGCPServiceAccount)
	}

	// Validate log level
	validLogLevels := []string{
		constants.LogLevelDebug,
		constants.LogLevelInfo,
		constants.LogLevelWarn,
		constants.LogLevelError,
	}
	
	if !contains(validLogLevels, c.Environment.LogLevel) {
		c.Environment.LogLevel = constants.DefaultLogLevel
	}

	// Validate port
	if c.Environment.Port == "" {
		c.Environment.Port = constants.DefaultPort
	}

	return nil
}

// GetAddress returns the server address
func (c *Config) GetAddress() string {
	return ":" + c.Environment.Port
}

// IsAuthRequired returns true if API key authentication is required
func (c *Config) IsAuthRequired() bool {
	return c.Environment.OpenAIAPIKey != ""
}

// IsFakeThinkingEnabled returns true if fake thinking is enabled
func (c *Config) IsFakeThinkingEnabled() bool {
	return strings.ToLower(c.Environment.EnableFakeThinking) == "true"
}

// IsRealThinkingEnabled returns true if real thinking is enabled
func (c *Config) IsRealThinkingEnabled() bool {
	return strings.ToLower(c.Environment.EnableRealThinking) == "true"
}

// IsStreamThinkingAsContent returns true if thinking should be streamed as content
func (c *Config) IsStreamThinkingAsContent() bool {
	return strings.ToLower(c.Environment.StreamThinkingAsContent) == "true"
}

// GetTokenCacheExpiry returns the token cache expiry duration
func (c *Config) GetTokenCacheExpiry() int {
	return c.Environment.TokenCacheExpiry
}

// GetRequestTimeout returns the request timeout duration
func (c *Config) GetRequestTimeout() int {
	return c.Environment.RequestTimeout
}

// GetLogLevel returns the configured log level
func (c *Config) GetLogLevel() string {
	return c.Environment.LogLevel
}

// GetGCPServiceAccount returns the GCP service account credentials
func (c *Config) GetGCPServiceAccount() string {
	return c.Environment.GCPServiceAccount
}

// GetGeminiProjectID returns the Gemini project ID
func (c *Config) GetGeminiProjectID() string {
	return c.Environment.GeminiProjectID
}

// GetOpenAIAPIKey returns the OpenAI API key
func (c *Config) GetOpenAIAPIKey() string {
	return c.Environment.OpenAIAPIKey
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}