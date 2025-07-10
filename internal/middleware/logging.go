package middleware

import (
	"bytes"
	"fmt"
	"gemini-cli-go/internal/config"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"gemini-cli-go/internal/constants"

	"github.com/gin-gonic/gin"
)

// LoggingMiddleware provides structured logging for HTTP requests
func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC3339),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// RequestLoggingMiddleware logs detailed request information
func RequestLoggingMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.GetLogLevel() != constants.LogLevelDebug {
			c.Next()
			return
		}

		// Create body log writer to capture response
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Read and restore request body for logging
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Log request details
		logRequest(c, path, raw, bodyBytes, blw.body.Bytes(), latency)
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// logRequest logs detailed request information
func logRequest(c *gin.Context, path, raw string, bodyBytes, responseBody []byte, latency time.Duration) {
	clientIP := c.ClientIP()
	method := c.Request.Method
	statusCode := c.Writer.Status()
	userAgent := c.Request.UserAgent()
	
	if raw != "" {
		path = path + "?" + raw
	}

	// Basic request info
	log.Printf("[%s] %s %s - %d - %v - %s",
		method,
		clientIP,
		path,
		statusCode,
		latency,
		userAgent,
	)

	// Log request body for POST/PUT requests (but not for sensitive endpoints)
	if shouldLogBody(method, path) && len(bodyBytes) > 0 {
		bodyStr := string(bodyBytes)
		if len(bodyStr) > 1000 {
			bodyStr = bodyStr[:1000] + "... (truncated)"
		}
		log.Printf("[REQUEST BODY] %s", bodyStr)
	}

	// Log response body
	if len(responseBody) > 0 {
		bodyStr := string(responseBody)
		if len(bodyStr) > 1000 {
			bodyStr = bodyStr[:1000] + "... (truncated)"
		}
		log.Printf("[RESPONSE BODY] %s", bodyStr)
	}

	// Log errors if any
	if len(c.Errors) > 0 {
		log.Printf("[ERRORS] %v", c.Errors.String())
	}
}

// shouldLogBody determines if request body should be logged
func shouldLogBody(method, path string) bool {
	// Only log body for POST and PUT requests
	if method != "POST" && method != "PUT" {
		return false
	}

	// Skip logging for sensitive endpoints
	sensitiveEndpoints := []string{
		"/v1/token-test",
		"/v1/test",
	}

	for _, endpoint := range sensitiveEndpoints {
		if strings.Contains(path, endpoint) {
			return false
		}
	}

	return true
}

// CORSMiddleware handles CORS headers
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", constants.CORSAllowOrigin)
		c.Header("Access-Control-Allow-Methods", constants.CORSAllowMethods)
		c.Header("Access-Control-Allow-Headers", constants.CORSAllowHeaders)
		c.Header("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// ErrorLoggingMiddleware logs errors in detail
func ErrorLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Log any errors that occurred during request processing
		for _, err := range c.Errors {
			log.Printf("[ERROR] %s - %s %s - %v",
				c.ClientIP(),
				c.Request.Method,
				c.Request.URL.Path,
				err.Error(),
			)
		}
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := generateRequestID()
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// GetRequestID gets the request ID from the context
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		return requestID.(string)
	}
	return ""
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// RateLimitInfo holds rate limiting information
type RateLimitInfo struct {
	Requests  int           `json:"requests"`
	Window    time.Duration `json:"window"`
	Remaining int           `json:"remaining"`
	ResetTime time.Time     `json:"reset_time"`
}

// BasicRateLimitMiddleware provides basic rate limiting
func BasicRateLimitMiddleware(requests int, window time.Duration) gin.HandlerFunc {
	clients := make(map[string]*RateLimitInfo)
	var mu sync.RWMutex

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		mu.Lock()
		defer mu.Unlock()

		// Clean up old entries
		for ip, info := range clients {
			if now.After(info.ResetTime) {
				delete(clients, ip)
			}
		}

		// Check current client
		info, exists := clients[clientIP]
		if !exists {
			clients[clientIP] = &RateLimitInfo{
				Requests:  1,
				Window:    window,
				Remaining: requests - 1,
				ResetTime: now.Add(window),
			}
			c.Next()
			return
		}

		// Check if window has passed
		if now.After(info.ResetTime) {
			info.Requests = 1
			info.Remaining = requests - 1
			info.ResetTime = now.Add(window)
			c.Next()
			return
		}

		// Check if limit exceeded
		if info.Requests >= requests {
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", requests))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", info.ResetTime.Unix()))
			
			c.JSON(429, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}

		// Update counter
		info.Requests++
		info.Remaining = requests - info.Requests

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", requests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", info.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", info.ResetTime.Unix()))

		c.Next()
	}
}

// HealthCheckMiddleware handles health check requests
func HealthCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			c.JSON(200, gin.H{
				"status":    constants.MsgHealthOK,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"version":   constants.AppVersion,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}